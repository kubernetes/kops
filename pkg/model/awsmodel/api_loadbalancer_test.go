/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awsmodel

import (
	"testing"

	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// TestAPILoadBalancerHealthCheck verifies that the API target group requests /readyz over HTTPS,
// except when kube-apiserver only accepts TLS 1.3, which the NLB health checker does not support.
func TestAPILoadBalancerHealthCheck(t *testing.T) {
	grid := []struct {
		tlsMinVersion    string
		expectedProtocol elbv2types.ProtocolEnum
		expectedPath     *string
	}{
		{tlsMinVersion: "", expectedProtocol: elbv2types.ProtocolEnumHttps, expectedPath: new("/readyz")},
		{tlsMinVersion: "VersionTLS12", expectedProtocol: elbv2types.ProtocolEnumHttps, expectedPath: new("/readyz")},
		{tlsMinVersion: "VersionTLS13", expectedProtocol: elbv2types.ProtocolEnumTcp, expectedPath: nil},
	}

	for _, g := range grid {
		t.Run("tlsMinVersion="+g.tlsMinVersion, func(t *testing.T) {
			cluster := buildMinimalCluster()
			cluster.Spec.API = kops.APISpec{
				LoadBalancer: &kops.LoadBalancerAccessSpec{
					Class: kops.LoadBalancerClassNetwork,
					Type:  kops.LoadBalancerTypePublic,
				},
			}
			cluster.Spec.KubeAPIServer = &kops.KubeAPIServerConfig{
				TLSMinVersion: g.tlsMinVersion,
			}

			igs := []*kops.InstanceGroup{
				{
					ObjectMeta: v1.ObjectMeta{Name: "control-plane"},
					Spec: kops.InstanceGroupSpec{
						Role:    kops.InstanceGroupRoleControlPlane,
						Subnets: []string{cluster.Spec.Networking.Subnets[0].Name},
					},
				},
			}

			b := APILoadBalancerBuilder{
				AWSModelContext: &AWSModelContext{
					KopsModelContext: &model.KopsModelContext{
						IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
						AllInstanceGroups: igs,
						InstanceGroups:    igs,
					},
				},
				Lifecycle:         fi.LifecycleSync,
				SecurityLifecycle: fi.LifecycleSync,
			}

			c := &fi.CloudupModelBuilderContext{
				Tasks: make(map[string]fi.CloudupTask),
			}
			if err := b.Build(c); err != nil {
				t.Fatalf("unexpected error from Build: %v", err)
			}

			var apiTargetGroup *awstasks.TargetGroup
			for _, task := range c.Tasks {
				tg, ok := task.(*awstasks.TargetGroup)
				if !ok || tg.Protocol != elbv2types.ProtocolEnumTcp || fi.ValueOf(tg.Port) != 443 {
					continue
				}
				if apiTargetGroup != nil {
					t.Fatalf("found more than one API target group")
				}
				apiTargetGroup = tg
			}
			if apiTargetGroup == nil {
				t.Fatalf("API target group not found")
			}

			if apiTargetGroup.HealthCheckProtocol != g.expectedProtocol {
				t.Errorf("unexpected health check protocol: got %q, want %q", apiTargetGroup.HealthCheckProtocol, g.expectedProtocol)
			}
			if fi.ValueOf(apiTargetGroup.HealthCheckPath) != fi.ValueOf(g.expectedPath) {
				t.Errorf("unexpected health check path: got %q, want %q", fi.ValueOf(apiTargetGroup.HealthCheckPath), fi.ValueOf(g.expectedPath))
			}
		})
	}
}
