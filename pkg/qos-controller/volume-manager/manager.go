package volumemanager

import (
	corev1 "k8s.io/api/core/v1"
)

type ErrInvalidArgs struct {
	Err error
}

func (e ErrInvalidArgs) Error() string {
	return e.Err.Error()
}

type VolumeManager interface {
	Connect() error
	Close()
	SetQoS(pv *corev1.PersistentVolume, settings QoSSettings) error
	Validate(settings QoSSettings) error
}

type CommonConfig struct {
	Provisioner string `json:"provisioner" yaml:"provisioner"`
}

func (cc CommonConfig) HasProvisioner() bool {
	return cc.Provisioner != ""
}
