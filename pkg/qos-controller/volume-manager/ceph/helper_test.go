package ceph

import (
	"reflect"
	"testing"

	vm "github.com/crazytaxii/volume-qos-controller/pkg/qos-controller/volume-manager"
)

func Test_getQoSRulesFromMeta(t *testing.T) {
	type args struct {
		meta map[string]string
	}
	tests := []struct {
		name string
		args args
		want RBDQoSRules
	}{
		{
			name: "none",
			args: args{
				meta: map[string]string{
					"foo": "bar",
				},
			},
			want: RBDQoSRules{},
		},
		{
			name: "all",
			args: args{
				meta: map[string]string{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{
				RBDQoSLimitIOPSKey:      "1",
				RBDQoSLimitReadIOPSKey:  "1",
				RBDQoSLimitWriteIOPSKey: "1",
			},
		},
		{
			name: "mixed",
			args: args{
				meta: map[string]string{
					"foo":                   "bar",
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{
				RBDQoSLimitIOPSKey:      "1",
				RBDQoSLimitReadIOPSKey:  "1",
				RBDQoSLimitWriteIOPSKey: "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getQoSRulesFromMeta(tt.args.meta); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getQoSRulesFromMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rbdQoSRules(t *testing.T) {
	type args struct {
		settings vm.QoSSettings
	}
	tests := []struct {
		name string
		args args
		want RBDQoSRules
	}{
		{
			name: "convert",
			args: args{
				settings: vm.QoSSettings{
					vm.QoSLimitIOPSKey:      "1",
					vm.QoSLimitReadIOPSKey:  "1",
					vm.QoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{
				RBDQoSLimitIOPSKey:      "1",
				RBDQoSLimitReadIOPSKey:  "1",
				RBDQoSLimitWriteIOPSKey: "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rbdQoSRules(tt.args.settings); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rbdQoSRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calSet(t *testing.T) {
	type args struct {
		cur  RBDQoSRules
		spec RBDQoSRules
	}
	tests := []struct {
		name string
		args args
		want RBDQoSRules
	}{
		{
			name: "add all",
			args: args{
				cur: RBDQoSRules{},
				spec: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{
				RBDQoSLimitIOPSKey:      "1",
				RBDQoSLimitReadIOPSKey:  "1",
				RBDQoSLimitWriteIOPSKey: "1",
			},
		},
		{
			name: "add inexisted",
			args: args{
				cur: RBDQoSRules{
					RBDQoSLimitIOPSKey:     "1",
					RBDQoSLimitReadIOPSKey: "1",
				},
				spec: RBDQoSRules{
					RBDQoSLimitIOPSKey:     "1",
					RBDQoSLimitReadIOPSKey: "1",

					RBDQoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{
				RBDQoSLimitWriteIOPSKey: "1",
			},
		},
		{
			name: "add none 1",
			args: args{
				cur: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
				spec: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{},
		},
		{
			name: "add none 2",
			args: args{
				cur: RBDQoSRules{
					RBDQoSBurstIOPSKey: "1",

					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
				spec: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calSet(tt.args.cur, tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calRemove(t *testing.T) {
	type args struct {
		cur  RBDQoSRules
		spec RBDQoSRules
	}
	tests := []struct {
		name string
		args args
		want RBDQoSRules
	}{
		{
			name: "remove all",
			args: args{
				cur: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
				spec: RBDQoSRules{},
			},
			want: RBDQoSRules{
				RBDQoSLimitIOPSKey:      "1",
				RBDQoSLimitReadIOPSKey:  "1",
				RBDQoSLimitWriteIOPSKey: "1",
			},
		},
		{
			name: "remove one",
			args: args{
				cur: RBDQoSRules{
					RBDQoSLimitIOPSKey: "1",

					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
				spec: RBDQoSRules{
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{
				RBDQoSLimitIOPSKey: "1",
			},
		},
		{
			name: "remove none 1",
			args: args{
				cur: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
				spec: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
			},
			want: RBDQoSRules{},
		},
		{
			name: "remove none 2",
			args: args{
				cur: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",
				},
				spec: RBDQoSRules{
					RBDQoSLimitIOPSKey:      "1",
					RBDQoSLimitReadIOPSKey:  "1",
					RBDQoSLimitWriteIOPSKey: "1",

					RBDQoSBurstIOPSKey: "1",
				},
			},
			want: RBDQoSRules{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calRemove(tt.args.cur, tt.args.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calRemove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isQoSValueValid(t *testing.T) {
	type args struct {
		unit string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test 1",
			args: args{
				unit: "10",
			},
			want: true,
		},
		{
			name: "test 2",
			args: args{
				unit: "01",
			},
			want: false,
		},
		{
			name: "test 3",
			args: args{
				unit: "999",
			},
			want: true,
		},
		{
			name: "test 4",
			args: args{
				unit: "1M",
			},
			want: true,
		},
		{
			name: "test 5",
			args: args{
				unit: "1G",
			},
			want: true,
		},
		{
			name: "test 6",
			args: args{
				unit: "100M",
			},
			want: true,
		},
		{
			name: "test 7",
			args: args{
				unit: "100G",
			},
			want: true,
		},
		{
			name: "test 8",
			args: args{
				unit: "100K",
			},
			want: false,
		},
		{
			name: "test 9",
			args: args{
				unit: "100MM",
			},
			want: false,
		},
		{
			name: "test 10",
			args: args{
				unit: "100MG",
			},
			want: false,
		},
		{
			name: "test 11",
			args: args{
				unit: "100Mi",
			},
			want: false,
		},
		{
			name: "test 12",
			args: args{
				unit: "100m",
			},
			want: false,
		},
		{
			name: "test 13",
			args: args{
				unit: "1",
			},
			want: true,
		},
		{
			name: "test 14",
			args: args{
				unit: "100T",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isQoSValueValid(tt.args.unit); got != tt.want {
				t.Errorf("isValidBPS() = %v, want %v", got, tt.want)
			}
		})
	}
}
