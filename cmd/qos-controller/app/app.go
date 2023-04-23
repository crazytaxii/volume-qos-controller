package app

import (
	"context"
	"fmt"
	"os"

	"github.com/crazytaxii/volume-qos-controller/cmd/qos-controller/app/config"
	"github.com/crazytaxii/volume-qos-controller/cmd/qos-controller/app/option"
	qc "github.com/crazytaxii/volume-qos-controller/pkg/qos-controller"
	"github.com/crazytaxii/volume-qos-controller/pkg/signals"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
)

func NewCommand(version string) *cobra.Command {
	opts := option.NewOptions()
	cmd := &cobra.Command{
		Use:  "qos-controller",
		Long: "qos-controller is a controller for setting volume(PV) QoS",
	}

	runCmd := &cobra.Command{
		Use:     "run",
		Short:   "Launch a volume QoS controller daemon",
		Long:    "run subcommand will launch a volume QoS controller daemon",
		Example: "qos-controller run --config=/path/to/config.yaml",
		Run: func(_ *cobra.Command, _ []string) {
			if err := run(signals.SetupSignalHandler(), opts); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
	}
	opts.AddFlags(runCmd)

	verCmd := &cobra.Command{
		Use:     "version",
		Short:   "Print version and exit",
		Long:    "version subcommand will print version and exit",
		Example: "qos-controller version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("version:", version)
		},
	}

	cmd.AddCommand(runCmd, verCmd)
	return cmd
}

func run(ctx context.Context, opts *option.Options) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	kubeConfig, err := config.BuildKubeConfig()
	if err != nil {
		return err
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	ctrl, err := qc.NewQosController(kubeClient, cfg.ControllerConfig)
	if err != nil {
		return err
	}

	if cfg.LeaderElect {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}

		id := fmt.Sprintf("%s-%s", hostname, string(uuid.NewUUID()))
		lock := &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      cfg.ResourceName,
				Namespace: cfg.ResourceNamespace,
			},
			Client: kubeClient.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: id,
			},
		}

		c, cancel := context.WithCancel(ctx)
		defer cancel()
		// start the leader election loop
		leaderelection.RunOrDie(c, leaderelection.LeaderElectionConfig{
			Lock:            lock,
			ReleaseOnCancel: true,
			LeaseDuration:   cfg.LeaseDuration,
			RenewDeadline:   cfg.RenewDeadline,
			RetryPeriod:     cfg.RetryPeriod,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					if err := ctrl.Run(ctx.Done()); err != nil {
						klog.Errorf("error running QoS controller: %v", err)
					}
					cancel()
				},
				OnStoppedLeading: func() {
					klog.Errorf("leader lost: %s", id)
				},
				OnNewLeader: func(identity string) {
					// we're notified when new leader elected
					if identity == id {
						// I just got the lock
						return
					}
					klog.Infof("New leader elected: %s", identity)
				},
			},
		})

		return nil
	}

	if err := ctrl.Run(ctx.Done()); err != nil {
		return fmt.Errorf("error running QoS controller: %v", err)
	}
	return nil
}
