package cloudup

import (
	"k8s.io/kops/pkg/apis/kops"
	"reflect"
	"testing"
	"fmt"
)

func Test_TemplateFunctions_CloudControllerConfigArgv(t *testing.T) {
	tests := []struct {
		desc         string
		cluster      *kops.Cluster
		expectedArgv []string
		expectedError error
	}{
		{
			desc: "Default Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},

			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
			},
		},
		{
			desc: "Log Level Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: string(kops.CloudProviderOpenstack),
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						LogLevel: 3,
					},
				},
			},
			expectedArgv: []string{
				"--v=3",
				"--cloud-provider=openstack",
			},
		},
		{
			desc: "ExternalCloudControllerManager CloudProvider Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						CloudProvider: string(kops.CloudProviderOpenstack),
						LogLevel: 3,
					},
				},
			},
			expectedArgv: []string{
				"--v=3",
				"--cloud-provider=openstack",
			},
		},
		{
			desc: "No CloudProvider Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						LogLevel: 3,
					},
				},
			},
			expectedError: fmt.Errorf("Cloud Provider is not set"),
		},
		{
			desc: "Default Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterName: "k8s",
				},


			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--cluster-name=k8s",
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.desc, func(t *testing.T) {
			tf := &TemplateFunctions{
				cluster: testCase.cluster,
			}
			actual, error := tf.CloudControllerConfigArgv()
			if !reflect.DeepEqual(error, testCase.expectedError) {
				t.Errorf("Error differs: %+v instead of %+v", error, testCase.expectedError)
			}
			if !reflect.DeepEqual(actual, testCase.expectedArgv) {
				t.Errorf("Argv differs: %+v instead of %+v", actual, testCase.expectedArgv)
			}
		})
	}
}
