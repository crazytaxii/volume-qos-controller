package config

import (
	"path/filepath"
	"strings"
	"time"

	qc "github.com/crazytaxii/volume-qos-controller/pkg/qos-controller"

	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	defaultKubeConfigDir     = ".kube/config"
	DefaultLeaseDuration     = 30 * time.Second
	DefaultRenewDeadline     = 15 * time.Second
	DefaultRetryPeriod       = 5 * time.Second
	DefaultResourceName      = "qos-controller-leader-lock"
	DefaultResourceLock      = "endpointsleases"
	DefaultResourceNamespace = "default"
)

type (
	LeaderElection struct {
		// leaderElect enables a leader election client to gain leadership
		// before executing the main loop. Enable this when running replicated
		// components for high availability.
		LeaderElect bool `json:"leader_elect" yaml:"leaderElect"`
		// leaseDuration is the duration that non-leader candidates will wait
		// after observing a leadership renewal until attempting to acquire
		// leadership of a led but unrenewed leader slot. This is effectively the
		// maximum duration that a leader can be stopped before it is replaced
		// by another candidate. This is only applicable if leader election is
		// enabled.
		LeaseDuration time.Duration `json:"lease_duration" yaml:"leaseDuration"`
		// renewDeadline is the interval between attempts by the acting master to
		// renew a leadership slot before it stops leading. This must be less
		// than or equal to the lease duration. This is only applicable if leader
		// election is enabled.
		RenewDeadline time.Duration `json:"renew_deadline" yaml:"renewDeadline"`
		// retryPeriod is the duration the clients should wait between attempting
		// acquisition and renewal of a leadership. This is only applicable if
		// leader election is enabled.
		RetryPeriod time.Duration `json:"retry_period" yaml:"retryPeriod"`
		// resourceLock indicates the resource object type that will be used to lock
		// during leader election cycles.
		ResourceLock string `json:"resource_lock" yaml:"resourceLock"`
		// resourceName indicates the name of resource object that will be used to lock
		// during leader election cycles.
		ResourceName string `json:"resource_name" yaml:"resourceName"`
		// resourceNamespace indicates the namespace of resource object that will be used to lock
		// during leader election cycles.
		ResourceNamespace string `json:"resource_namespace" yaml:"resourceNamespace"`
	}
	Config struct {
		*LeaderElection      `json:"leader_election" yaml:"leaderElection"`
		*qc.ControllerConfig `json:"controller_config" yaml:"controllerConfig"`
	}
)

func DefaultLeaderElection() *LeaderElection {
	return &LeaderElection{
		LeaseDuration:     DefaultLeaseDuration,
		RenewDeadline:     DefaultRenewDeadline,
		RetryPeriod:       DefaultRetryPeriod,
		ResourceLock:      DefaultResourceLock,
		ResourceName:      DefaultResourceName,
		ResourceNamespace: DefaultResourceNamespace,
	}
}

func NewDefaultConfig() *Config {
	return &Config{
		LeaderElection:   DefaultLeaderElection(),
		ControllerConfig: qc.DefaultControllerConfig(),
	}
}

// LoadConfigFile reads in and parses the config file.
func LoadConfigFile(path string) (*Config, error) {
	cfgFile, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	cfgDir := filepath.Dir(cfgFile)
	name := filepath.Base(cfgFile)
	ext := filepath.Ext(name)

	viper.SetConfigType(strings.TrimPrefix(ext, "."))
	viper.SetConfigName(strings.TrimSuffix(name, ext))
	viper.AddConfigPath(cfgDir)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := NewDefaultConfig()
	err = viper.Unmarshal(cfg)
	return cfg, err
}

// BuildKubeConfig builds a kube client config from the Pod environment or default kube config file.
func BuildKubeConfig() (*rest.Config, error) {
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, nil
	}
	return clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), defaultKubeConfigDir))
}
