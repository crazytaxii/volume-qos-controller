package volumemanager

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPVCQoSSettings(t *testing.T) {
	type args struct {
		pvc *corev1.PersistentVolumeClaim
	}
	tests := []struct {
		name string
		args args
		want QoSSettings
	}{
		{
			name: "none",
			args: args{
				pvc: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"pv.kubernetes.io/bind-completed":      "yes",
							"pv.kubernetes.io/bound-by-controller": "yes",
						},
					},
				},
			},
			want: QoSSettings{},
		},
		{
			name: "mixed",
			args: args{
				pvc: &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"pv.kubernetes.io/bind-completed":      "yes",
							"pv.kubernetes.io/bound-by-controller": "yes",
							QoSLimitIOPSKey:                        "1",
							QoSLimitReadIOPSKey:                    "1",
							QoSLimitWriteIOPSKey:                   "1",
						},
					},
				},
			},
			want: QoSSettings{
				QoSLimitIOPSKey:      "1",
				QoSLimitReadIOPSKey:  "1",
				QoSLimitWriteIOPSKey: "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPVCQoSSettings(tt.args.pvc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPVCQoSSettings() = %v, want %v", got, tt.want)
			}
		})
	}
}
