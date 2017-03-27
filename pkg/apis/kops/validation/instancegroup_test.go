package validation

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"strings"
	"testing"
)

func TestDefaultTaintsEnforcedBefore160(t *testing.T) {
	type param struct {
		ver       string
		role      kops.InstanceGroupRole
		taints    []string
		shouldErr bool
	}

	params := []param{
		{"1.5.0", kops.InstanceGroupRoleNode, []string{kops.TaintNoScheduleMaster15}, true},
		{"1.5.1", kops.InstanceGroupRoleNode, nil, false},
		{"1.5.2", kops.InstanceGroupRoleNode, []string{}, false},
		{"1.6.0", kops.InstanceGroupRoleNode, []string{kops.TaintNoScheduleMaster15}, false},
		{"1.6.1", kops.InstanceGroupRoleNode, []string{"Foo"}, false},
	}

	for _, p := range params {
		cluster := &kops.Cluster{Spec: kops.ClusterSpec{KubernetesVersion: p.ver}}
		ig := &kops.InstanceGroup{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: kops.InstanceGroupSpec{
				Taints: p.taints,
				Role:   p.role,
			},
		}

		err := CrossValidateInstanceGroup(ig, cluster, false)
		if p.shouldErr {
			if err == nil {
				t.Fatal("Expected error building kubelet config, received nil.")
			} else if !strings.Contains(err.Error(), "User-specified taints are not supported before kubernetes version 1.6.0") {
				t.Fatalf("Received an unexpected error validating taints: '%s'", err.Error())
			}
		} else {
			if err != nil {
				t.Fatalf("Received an unexpected error validating taints: '%s', params: '%v'", err.Error(), p)
			}
		}
	}
}
