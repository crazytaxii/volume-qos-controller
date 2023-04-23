package volumemanager

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IOPS: number of I/Os per second (any type of I/O)
// read IOPS: number of read I/Os per second
// write IOPS: number of write I/Os per second
// bps: bytes per second (any type of I/O)
// read bps: bytes per second read
// write bps: bytes per second written

const (
	QoSPrefix = "pv.kubernetes.io/"

	// iops limit
	QoSLimitIOPSKey      = QoSPrefix + "qos-iops-limit"
	QoSLimitReadIOPSKey  = QoSPrefix + "qos-read-iops-limit"
	QoSLimitWriteIOPSKey = QoSPrefix + "qos-write-iops-limit"

	// iops burst
	QoSBurstIOPSKey      = QoSPrefix + "qos-iops-burst"
	QoSBurstReadIOPSKey  = QoSPrefix + "qos-read-iops-burst"
	QoSBurstWriteIOPSKey = QoSPrefix + "qos-write-iops-burst"

	// bps limit
	QoSLimitBPSKey      = QoSPrefix + "qos-bps-limit"
	QoSLimitReadBPSKey  = QoSPrefix + "qos-read-bps-limit"
	QoSLimitWriteBPSKey = QoSPrefix + "qos-write-bps-limit"

	// bps burst
	QoSBurstBPSKey      = QoSPrefix + "qos-bps-burst"
	QoSBurstReadBPSKey  = QoSPrefix + "qos-read-bps-burst"
	QoSBurstWriteBPSKey = QoSPrefix + "qos-write-bps-burst"
)

type QoSSettings map[string]string

// GetPVCQoSSettings extracts the QoS settings from the PVC annotations.
func GetPVCQoSSettings(pvc *corev1.PersistentVolumeClaim) QoSSettings {
	settings := make(QoSSettings)

	for _, key := range []string{
		QoSLimitIOPSKey,
		QoSLimitReadIOPSKey,
		QoSLimitWriteIOPSKey,

		QoSBurstIOPSKey,
		QoSBurstReadIOPSKey,
		QoSBurstWriteIOPSKey,

		QoSLimitBPSKey,
		QoSLimitReadBPSKey,
		QoSLimitWriteBPSKey,

		QoSBurstBPSKey,
		QoSBurstReadBPSKey,
		QoSBurstWriteBPSKey,
	} {
		if metav1.HasAnnotation(pvc.ObjectMeta, key) {
			settings[key] = pvc.Annotations[key]
		}
	}

	return settings
}
