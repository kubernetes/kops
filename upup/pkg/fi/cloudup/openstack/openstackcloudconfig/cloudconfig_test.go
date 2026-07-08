/*
Copyright 2019 The Kubernetes Authors.

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

package openstackcloudconfig

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func Test_OpenstackCloud_MakeCloud(t *testing.T) {
	baseCloudConfigWithBlockStorage := []string{
		"auth-url=\"\"",
		"username=\"\"",
		"password=\"\"",
		"region=\"\"",
		"tenant-id=\"\"",
		"tenant-name=\"\"",
		"domain-name=\"\"",
		"domain-id=\"\"",
		"application-credential-id=\"\"",
		"application-credential-secret=\"\"",
		"",
		"[BlockStorage]",
		"bs-version=",
		"ignore-volume-az=false",
	}

	tests := []struct {
		desc                string
		cluster             *kops.Cluster
		expectedCloudConfig []string
	}{
		{
			desc: "Ignore volume microversion is set to false when not configured",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							BlockStorage: &kops.OpenstackBlockStorageConfig{},
						},
					},
				},
			},
			expectedCloudConfig: append(baseCloudConfigWithBlockStorage,
				"ignore-volume-microversion=false",
				"",
			),
		},
		{
			desc: "Ignore volume microversion is set to configured value",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							BlockStorage: &kops.OpenstackBlockStorageConfig{
								IgnoreVolumeMicroVersion: new(true),
							},
						},
					},
				},
			},
			expectedCloudConfig: append(baseCloudConfigWithBlockStorage,
				"ignore-volume-microversion=true",
				"",
			),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.desc, func(t *testing.T) {
			actualCloudConfig := MakeCloudConfig(testCase.cluster.Spec.CloudProvider.Openstack)

			if !reflect.DeepEqual(actualCloudConfig, testCase.expectedCloudConfig) {
				t.Errorf("Ingress status differ: expected\n%+#v\n\tgot:\n%+#v\n", testCase.expectedCloudConfig, actualCloudConfig)
			}
		})
	}
}
