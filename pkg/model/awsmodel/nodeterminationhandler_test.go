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
	"sort"
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

func TestNodeTerminationHandlerBuilder(t *testing.T) {
	grid := []struct {
		name           string
		nth            *kops.NodeTerminationHandlerSpec
		karpenter      *kops.KarpenterConfig
		expectedRules  []string
		expectedHooks  []string
		expectedQueues []string
	}{
		{
			name: "nth queue mode",
			nth:  &kops.NodeTerminationHandlerSpec{Enabled: new(true)},
			expectedRules: []string{
				"testcluster.test.com-ASGLifecycle",
				"testcluster.test.com-InstanceScheduledChange",
				"testcluster.test.com-InstanceStateChange",
				"testcluster.test.com-SpotInterruption",
			},
			expectedHooks:  []string{"nodes-NTHLifecycleHook"},
			expectedQueues: []string{"testcluster-test-com-nth"},
		},
		{
			name: "nth queue mode with rebalance draining",
			nth: &kops.NodeTerminationHandlerSpec{
				Enabled:                 new(true),
				EnableRebalanceDraining: new(true),
			},
			expectedRules: []string{
				"testcluster.test.com-ASGLifecycle",
				"testcluster.test.com-InstanceScheduledChange",
				"testcluster.test.com-InstanceStateChange",
				"testcluster.test.com-RebalanceRecommendation",
				"testcluster.test.com-SpotInterruption",
			},
			expectedHooks:  []string{"nodes-NTHLifecycleHook"},
			expectedQueues: []string{"testcluster-test-com-nth"},
		},
		{
			name:      "karpenter without nth spec",
			karpenter: &kops.KarpenterConfig{Enabled: true},
			expectedRules: []string{
				"testcluster.test.com-InstanceScheduledChange",
				"testcluster.test.com-InstanceStateChange",
				"testcluster.test.com-RebalanceRecommendation",
				"testcluster.test.com-SpotInterruption",
			},
			expectedQueues: []string{"testcluster-test-com-nth"},
		},
		{
			name:      "karpenter with disabled nth spec",
			nth:       &kops.NodeTerminationHandlerSpec{Enabled: new(false)},
			karpenter: &kops.KarpenterConfig{Enabled: true},
			expectedRules: []string{
				"testcluster.test.com-InstanceScheduledChange",
				"testcluster.test.com-InstanceStateChange",
				"testcluster.test.com-RebalanceRecommendation",
				"testcluster.test.com-SpotInterruption",
			},
			expectedQueues: []string{"testcluster-test-com-nth"},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			cluster := buildMinimalCluster()
			cluster.Spec.CloudProvider.AWS.NodeTerminationHandler = g.nth
			cluster.Spec.Karpenter = g.karpenter

			ig := buildNodeInstanceGroup("subnet-us-test-1a")
			ig.Spec.Manager = kops.InstanceManagerCloudGroup
			igs := []*kops.InstanceGroup{ig}

			b := &NodeTerminationHandlerBuilder{
				AWSModelContext: &AWSModelContext{
					KopsModelContext: &model.KopsModelContext{
						IAMModelContext:   iam.IAMModelContext{Cluster: cluster, AWSPartition: "aws"},
						AllInstanceGroups: igs,
						InstanceGroups:    igs,
						Region:            "us-test-1",
					},
				},
				Lifecycle: fi.LifecycleSync,
			}

			c := &fi.CloudupModelBuilderContext{
				Tasks: make(map[string]fi.CloudupTask),
			}
			if err := b.Build(c); err != nil {
				t.Fatalf("error from Build: %v", err)
			}

			var rules, targets, hooks, queues []string
			for _, task := range c.Tasks {
				switch task := task.(type) {
				case *awstasks.EventBridgeRule:
					rules = append(rules, fi.ValueOf(task.Name))
				case *awstasks.EventBridgeTarget:
					targets = append(targets, strings.TrimSuffix(fi.ValueOf(task.Name), "-Target"))
				case *awstasks.AutoscalingLifecycleHook:
					hooks = append(hooks, fi.ValueOf(task.Name))
				case *awstasks.SQS:
					queues = append(queues, fi.ValueOf(task.Name))
				}
			}

			assertNames(t, "rules", rules, g.expectedRules)
			assertNames(t, "targets", targets, g.expectedRules)
			assertNames(t, "lifecycle hooks", hooks, g.expectedHooks)
			assertNames(t, "queues", queues, g.expectedQueues)
		})
	}
}

func assertNames(t *testing.T, kind string, actual, expected []string) {
	t.Helper()
	sort.Strings(actual)
	if strings.Join(actual, ",") != strings.Join(expected, ",") {
		t.Errorf("unexpected %s: got %v, want %v", kind, actual, expected)
	}
}
