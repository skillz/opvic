package agent

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func Test_getFeilds(t *testing.T) {
	type args struct {
		jsonPath string
		resource interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Test_getFeilds",
			args: args{
				jsonPath: ".status.nodeInfo.kubeletVersion",
				resource: &corev1.Node{
					Status: corev1.NodeStatus{
						NodeInfo: corev1.NodeSystemInfo{
							KubeletVersion: "v1.18.0",
						},
					},
				},
			},
			want:    []string{"v1.18.0"},
			wantErr: false,
		},
		{
			name: "Test_getFeilds_invalid_jsonpath",
			args: args{
				jsonPath: "invalid_jsonpath",
				resource: &corev1.Node{
					Status: corev1.NodeStatus{
						NodeInfo: corev1.NodeSystemInfo{
							KubeletVersion: "v1.18.0",
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFeilds(tt.args.jsonPath, tt.args.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFeilds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFeilds() = %v, want %v", got, tt.want)
			}
		})
	}
}
