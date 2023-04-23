package ceph

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	vm "github.com/crazytaxii/volume-qos-controller/pkg/qos-controller/volume-manager"

	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	DefaultCSIDriver     = "rbd.csi.ceph.com"
	tmpKeyFileLocation   = "/tmp"
	tmpKeyFileNamePrefix = "keyfile-"

	RBDQoSLimitIOPSKey      = "conf_rbd_qos_iops_limit"
	RBDQoSLimitReadIOPSKey  = "conf_rbd_qos_read_iops_limit"
	RBDQoSLimitWriteIOPSKey = "conf_rbd_qos_write_iops_limit"

	RBDQoSBurstIOPSKey      = "conf_rbd_qos_iops_burst"
	RBDQoSBurstReadIOPSKey  = "conf_rbd_qos_read_iops_burst"
	RBDQoSBurstWriteIOPSKey = "conf_rbd_qos_write_iops_burst"

	RBDQoSLimitBPSKey      = "conf_rbd_qos_bps_limit"
	RBDQoSLimitReadBPSKey  = "conf_rbd_qos_read_bps_limit"
	RBDQoSLimitWriteBPSKey = "conf_rbd_qos_write_bps_limit"

	RBDQoSBurstBPSKey      = "conf_rbd_qos_bps_burst"
	RBDQoSBurstReadBPSKey  = "conf_rbd_qos_read_bps_burst"
	RBDQoSBurstWriteBPSKey = "conf_rbd_qos_write_bps_burst"
)

var (
	QoSKeyMap = map[string]string{
		vm.QoSLimitIOPSKey:      RBDQoSLimitIOPSKey,
		vm.QoSLimitReadIOPSKey:  RBDQoSLimitReadIOPSKey,
		vm.QoSLimitWriteIOPSKey: RBDQoSLimitWriteIOPSKey,

		vm.QoSBurstIOPSKey:      RBDQoSBurstIOPSKey,
		vm.QoSBurstReadIOPSKey:  RBDQoSBurstReadIOPSKey,
		vm.QoSBurstWriteIOPSKey: RBDQoSBurstWriteIOPSKey,

		vm.QoSLimitBPSKey:      RBDQoSLimitBPSKey,
		vm.QoSLimitReadBPSKey:  RBDQoSLimitReadBPSKey,
		vm.QoSLimitWriteBPSKey: RBDQoSLimitWriteBPSKey,

		vm.QoSBurstBPSKey:      RBDQoSBurstBPSKey,
		vm.QoSBurstReadBPSKey:  RBDQoSBurstReadBPSKey,
		vm.QoSBurstWriteBPSKey: RBDQoSBurstWriteBPSKey,
	}

	RBDQoSKeyMap = map[string]struct{}{
		RBDQoSLimitIOPSKey:      {},
		RBDQoSLimitReadIOPSKey:  {},
		RBDQoSLimitWriteIOPSKey: {},

		RBDQoSBurstIOPSKey:      {},
		RBDQoSBurstReadIOPSKey:  {},
		RBDQoSBurstWriteIOPSKey: {},

		RBDQoSLimitBPSKey:      {},
		RBDQoSLimitReadBPSKey:  {},
		RBDQoSLimitWriteBPSKey: {},

		RBDQoSBurstBPSKey:      {},
		RBDQoSBurstReadBPSKey:  {},
		RBDQoSBurstWriteBPSKey: {},
	}
)

type (
	RBDQoSRules      map[string]string
	RBDManagerConfig struct {
		vm.CommonConfig `mapstructure:",squash"`
		Monitors        string `json:"monitors" yaml:"monitors"`
		User            string `json:"user" yaml:"user"`
		Key             string `json:"key" yaml:"key"`
	}
	CephRBDManager struct {
		conn *rados.Conn
		*RBDManagerConfig
	}
)

func DefaultCephRBDConfig() *RBDManagerConfig {
	return &RBDManagerConfig{
		CommonConfig: vm.CommonConfig{Provisioner: DefaultCSIDriver},
	}
}

func NewCephRBDManager(cfg *RBDManagerConfig) (*CephRBDManager, error) {
	// TODO: validate Ceph user & key & monitors

	if err := os.MkdirAll(tmpKeyFileLocation, 0644); err != nil {
		return nil, fmt.Errorf("creating a temporary keyfile directory failed: %w", err)
	}

	// Write user key to a temporary file.
	tmpfile, err := ioutil.TempFile(tmpKeyFileLocation, tmpKeyFileNamePrefix)
	if err != nil {
		return nil, fmt.Errorf("creating a temporary keyfile failed: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	// Write the Ceph user key to the temporary file.
	if _, err := tmpfile.WriteString(cfg.Key); err != nil {
		return nil, fmt.Errorf("writing key to temporary keyfile failed: %w", err)
	}
	defer tmpfile.Close()

	conn, err := rados.NewConnWithUser(cfg.User)
	if err != nil {
		return nil, fmt.Errorf("creating a new Ceph connection failed: %w", err)
	}

	args := []string{"-m", cfg.Monitors, "--keyfile=" + tmpfile.Name()}
	if err := conn.ParseCmdLineArgs(args); err != nil {
		return nil, fmt.Errorf("parsing cmdline args (%v) failed: %w", args, err)
	}

	return &CephRBDManager{
		conn:             conn,
		RBDManagerConfig: cfg,
	}, nil
}

// Connect connects to the Ceph cluster.
func (m *CephRBDManager) Connect() error {
	if err := m.conn.Connect(); err != nil {
		return fmt.Errorf("connecting Ceph cluster failed: %w", err)
	}
	klog.V(4).Info("Connected to Ceph cluster")
	return nil
}

// Close closes the connection to the Ceph cluster.
func (m *CephRBDManager) Close() {
	m.conn.Shutdown()
	klog.V(4).Info("Disconnected to Ceph cluster")
}

// getIOCtx creates and returns a new IOContext for the pool of PV.
func (m *CephRBDManager) getIOCtx(pv *corev1.PersistentVolume) (*rados.IOContext, error) {
	pool, ok := pv.Spec.CSI.VolumeAttributes["pool"]
	if !ok {
		return nil, fmt.Errorf("invalid PV missing pool in volumeAttributes")
	}
	return m.conn.OpenIOContext(pool)
}

// SetQoS configures the QoS settings for the RBD image of the PV.
func (m *CephRBDManager) SetQoS(pv *corev1.PersistentVolume, settings vm.QoSSettings) (err error) {
	if pv == nil {
		return fmt.Errorf("PV is nil")
	}

	name, ok := pv.Spec.CSI.VolumeAttributes["imageName"]
	if !ok {
		return fmt.Errorf("invalid PV %s missing imageName in volumeAttributes", pv.Name)
	}

	ioctx, err := m.getIOCtx(pv)
	if err != nil {
		return fmt.Errorf("failed to open IOContext for PV %s: %w", pv.Name, err)
	}
	defer ioctx.Destroy()

	img, err := rbd.OpenImage(ioctx, name, rbd.NoSnapshot)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %w", name, err)
	}
	defer img.Close()

	klog.V(4).Infof("Opened the RBD image %s of PV %s", name, pv.Name)

	// Get the metadata of rbd image.
	meta, err := img.ListMetadata()
	if err != nil {
		return fmt.Errorf("failed to list metadata of PV %s: %w", pv.Name, err)
	}
	cur := getQoSRulesFromMeta(meta) // the existing QoS rules
	spec := rbdQoSRules(settings)    // the expected QoS rules

	// Get the rules to be set.
	set := calSet(cur, spec)
	for k, v := range set {
		// Add or update QoS rule equals to set metadata for RBD image.
		if cerr := img.SetMetadata(k, v); cerr != nil {
			err = fmt.Errorf("failed to set metadata %s=%s for PV %s: %w", k, v, pv.Name, cerr)
			if isInvalidArgErr(cerr) {
				// If the error is caused by invalid argument, return ErrInvalidArgs.
				return vm.ErrInvalidArgs{Err: err}
			}
			return
		}
		klog.Infof("set metadata for PV %s: %s=%s", pv.Name, k, v)
	}

	// Get the rules to be removed.
	remove := calRemove(cur, spec)
	for k, v := range remove {
		// Remove QoS rule equals to remove metadata for RBD image.
		if err := img.RemoveMetadata(k); err != nil {
			return fmt.Errorf("failed to remove metadata %s=%s for PV %s: %w", k, v, pv.Name, err)
		}
		klog.Infof("remove metadata for PV %s: %s=%s", pv.Name, k, v)
	}

	return
}

func (m *CephRBDManager) Validate(settings vm.QoSSettings) error {
	for k, v := range settings {
		if !isQoSValueValid(v) {
			return fmt.Errorf("invalid value %q for QoS key %q", v, k)
		}
	}
	return nil
}

// isInvalidArgErr checks if the error is caused by invalid argument
func isInvalidArgErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Invalid argument")
}
