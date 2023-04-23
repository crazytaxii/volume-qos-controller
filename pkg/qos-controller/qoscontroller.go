package qoscontroller

import (
	"fmt"
	goruntime "runtime"
	"time"

	vm "github.com/crazytaxii/volume-qos-controller/pkg/qos-controller/volume-manager"
	"github.com/crazytaxii/volume-qos-controller/pkg/qos-controller/volume-manager/ceph"

	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	DefaultResyncPeriod = 30 * time.Minute

	controllerAgentName = "volume-qos-controller"

	AnnStorageProvisioner = "volume.kubernetes.io/storage-provisioner"
)

type (
	ControllerConfig struct {
		ResyncPeriod time.Duration          `json:"resync_period,omitempty" yaml:"resyncPeriod,omitempty"`
		Workers      int                    `json:"workers,omitempty" yaml:"workers,omitempty"` // the number of threadiness
		CephRBD      *ceph.RBDManagerConfig `json:"ceph_rbd" yaml:"cephRBD"`
	}
	VolumeQoSController struct {
		// kubeClient is a standard kubernetes clientset
		kubeClient kubernetes.Interface

		kubeInformerFactory kubeinformers.SharedInformerFactory

		pvcInformer coreinformers.PersistentVolumeClaimInformer
		pvcLister   corelisters.PersistentVolumeClaimLister

		pvInformer coreinformers.PersistentVolumeInformer
		pvLister   corelisters.PersistentVolumeLister

		workqueue workqueue.RateLimitingInterface

		// recorder is an event recorder for recording Event resources to the Kubernetes API.
		recorder record.EventRecorder

		volManagers map[string]vm.VolumeManager

		*ControllerConfig
	}
)

func DefaultControllerConfig() *ControllerConfig {
	return &ControllerConfig{
		ResyncPeriod: DefaultResyncPeriod,
		Workers:      goruntime.NumCPU() / 2,
		CephRBD:      ceph.DefaultCephRBDConfig(),
	}
}

func (cc *ControllerConfig) InitVolumeManagers() (managers map[string]vm.VolumeManager, err error) {
	managers = make(map[string]vm.VolumeManager)

	// Ceph RBD
	if rbd := cc.CephRBD; rbd != nil && rbd.HasProvisioner() {
		if managers[rbd.Provisioner], err = ceph.NewCephRBDManager(rbd); err != nil {
			return nil, fmt.Errorf("error initing a Ceph RBD volume manager: %v", err)
		}
	}
	return
}

func (cc *ControllerConfig) AddControllerConfigFlags(fs *pflag.FlagSet) {
	fs.DurationVarP(&cc.ResyncPeriod, "resync-period", "", cc.ResyncPeriod, "the resync interval duration for the QoS controller")
	fs.IntVarP(&cc.Workers, "workers", "", cc.Workers, "the number of threadiness")
}

func NewQosController(kubeClient kubernetes.Interface, cfg *ControllerConfig) (*VolumeQoSController, error) {
	// create event broadcaster
	utilruntime.Must(scheme.AddToScheme(scheme.Scheme))
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events(corev1.NamespaceAll)})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, cfg.ResyncPeriod)
	pvcInformer := kubeInformerFactory.Core().V1().PersistentVolumeClaims()
	pvInformer := kubeInformerFactory.Core().V1().PersistentVolumes()

	c := &VolumeQoSController{
		kubeInformerFactory: kubeInformerFactory,
		pvcInformer:         pvcInformer,
		pvcLister:           pvcInformer.Lister(),
		pvInformer:          pvInformer,
		pvLister:            pvInformer.Lister(),
		workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "VolumeQoS"),
		recorder:            recorder,
		ControllerConfig:    cfg,
	}

	// init volume managers
	var err error
	c.volManagers, err = cfg.InitVolumeManagers()
	if err != nil {
		return nil, err
	}

	klog.V(4).Infof("Setting up event handlers")
	pvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueuePVC,
		UpdateFunc: func(old, new interface{}) {
			oldPVC := old.(*corev1.PersistentVolumeClaim)
			newPVC := new.(*corev1.PersistentVolumeClaim)
			if oldPVC.ResourceVersion == newPVC.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			c.enqueuePVC(new)
		},
		DeleteFunc: c.enqueuePVC,
	})

	return c, nil
}

func (c *VolumeQoSController) Run(stopCh <-chan struct{}) error {
	// Don't let panics crash the process.
	defer utilruntime.HandleCrash()
	// Make sure the work queue is shutdown which will trigger workers to end.
	defer c.workqueue.ShutDown()

	klog.Info("Starting volume QoS controller")
	for _, manager := range c.volManagers {
		if err := manager.Connect(); err != nil {
			return err
		}
	}

	c.kubeInformerFactory.Start(stopCh)

	klog.Info("Waiting for informer caches to sync")
	if !cache.WaitForCacheSync(stopCh, c.pvcInformer.Informer().HasSynced,
		c.pvInformer.Informer().HasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Infof("Starting %d workers", c.Workers)
	for i := 0; i < c.Workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down volume QoS controller")

	for _, manager := range c.volManagers {
		manager.Close()
	}

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *VolumeQoSController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *VolumeQoSController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	if err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		key, ok := obj.(string)
		if !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the PVC to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing %q: %v, requeuing", key, err)
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		return nil
	}(obj); err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// enqueuePVC takes a PVC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than PVC.
func (c *VolumeQoSController) enqueuePVC(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (c *VolumeQoSController) syncHandler(key string) (err error) {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	pvc, err := c.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	if err != nil {
		// The PVC resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("pvc '%s' in work queue no longer exists", key))
			return nil
		}

		return
	}

	klog.V(4).Infof("Processing PVC %s", key)

	if pvc.Status.Phase != corev1.ClaimBound || !pvc.DeletionTimestamp.IsZero() {
		// It's unnecessary to process PVCs that are unbound or being deleted.
		return
	}

	provisioner, ok := pvc.Annotations[AnnStorageProvisioner]
	if !ok {
		klog.Warningf("Skip processing PVC %s: missing storage provisioner annotation", key)
		return
	}
	manager, ok := c.volManagers[provisioner]
	if !ok {
		klog.Warningf("Skip processing PVC %s: CSI driver %s is not supported", key, provisioner)
		return
	}

	// Get the PV bound to the PVC.
	pv, err := c.pvLister.Get(pvc.Spec.VolumeName)
	if err != nil {
		return err
	}

	klog.V(4).Infof("Processing PV %s bound to PVC %s", pv.Name, key)

	// Get the QoS settings from annotations of the PVC.
	qosSettings := vm.GetPVCQoSSettings(pvc)
	// Validate the value of QoS settings.
	if err := manager.Validate(qosSettings); err != nil {
		klog.Warningf("Failed to validate the QoS setting of PVC %s: %v", key, err)
		c.recorder.Event(pvc, corev1.EventTypeWarning, "InvalidQoSAnnotation", err.Error())
		return nil
	}

	if err = manager.SetQoS(pv, qosSettings); err != nil {
		c.recorder.Event(pvc, corev1.EventTypeWarning, "SettingQoSFailed", err.Error())
		if _, ok := err.(vm.ErrInvalidArgs); ok {
			klog.Error(err.Error())
			// invalid arguments should not be retried.
			return nil
		}
	}

	return
}
