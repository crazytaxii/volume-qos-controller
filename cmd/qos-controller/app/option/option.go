package option

import (
	"github.com/crazytaxii/volume-qos-controller/cmd/qos-controller/app/config"
	qc "github.com/crazytaxii/volume-qos-controller/pkg/qos-controller"

	"github.com/spf13/cobra"
)

type Options struct {
	ConfigFile string
	*config.LeaderElection
	*qc.ControllerConfig
}

func NewOptions() *Options {
	return &Options{
		LeaderElection:   config.DefaultLeaderElection(),
		ControllerConfig: qc.DefaultControllerConfig(),
	}
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ConfigFile, "config-file", "", "", "config file path will be read in")

	cmd.Flags().BoolVar(&o.LeaderElect, "leader-elect", o.LeaderElect, ""+
		"Start a leader election client and gain leadership before "+
		"executing the main loop. Enable this when running replicated "+
		"components for high availability.")
	cmd.Flags().DurationVar(&o.LeaseDuration, "leader-elect-lease-duration", o.LeaseDuration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	cmd.Flags().DurationVar(&o.RenewDeadline, "leader-elect-renew-deadline", o.RenewDeadline, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	cmd.Flags().DurationVar(&o.RetryPeriod, "leader-elect-retry-period", o.RetryPeriod, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
	cmd.Flags().StringVar(&o.ResourceLock, "leader-elect-resource-lock", o.ResourceLock, ""+
		"The type of resource object that is used for locking during "+
		"leader election. Supported options are 'endpoints', 'configmaps', "+
		"'leases', 'endpointsleases' and 'configmapsleases'.")
	cmd.Flags().StringVar(&o.ResourceName, "leader-elect-resource-name", o.ResourceName, ""+
		"The name of resource object that is used for locking during "+
		"leader election.")
	cmd.Flags().StringVar(&o.ResourceNamespace, "leader-elect-resource-namespace", o.ResourceNamespace, ""+
		"The namespace of resource object that is used for locking during "+
		"leader election.")

	o.AddControllerConfigFlags(cmd.Flags())
}

func (o *Options) Config() (*config.Config, error) {
	if o.ConfigFile != "" {
		return config.LoadConfigFile(o.ConfigFile)
	}
	return &config.Config{
		LeaderElection:   o.LeaderElection,
		ControllerConfig: o.ControllerConfig,
	}, nil
}
