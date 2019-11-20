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

package cloudup

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"text/template"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/controller/nodeipam/ipam"
)

func Test_TemplateFunctions_CloudControllerConfigArgv(t *testing.T) {
	tests := []struct {
		desc          string
		cluster       *kops.Cluster
		expectedArgv  []string
		expectedError error
		helperString string
	}{
		{
			desc: "Default Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider:                  string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
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
				"--use-service-account-credentials=true",
			},
		},
		{
			desc: "ExternalCloudControllerManager CloudProvider Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						CloudProvider: string(kops.CloudProviderOpenstack),
						LogLevel:      3,
					},
				},
			},
			expectedArgv: []string{
				"--v=3",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
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
			desc: "k8s cluster name",
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
				"--use-service-account-credentials=true",
			},
		},
		{
			desc: "Default Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider:                  string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					Master: "127.0.0.1",
				},
			}},
			expectedArgv: []string{
				"--master=127.0.0.1",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
			},
		},
		{
			desc: "Cluster-cidr Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider:                  string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterCIDR: "10.0.0.0/24",
				},
			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--cluster-cidr=10.0.0.0/24",
				"--use-service-account-credentials=true",
			},
		},
		{
			desc: "AllocateNodeCIDRs Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider:                  string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					AllocateNodeCIDRs: fi.Bool(true),
				},
			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--allocate-node-cidrs=true",
				"--use-service-account-credentials=true",

			},
		},
		{
			desc: "ConfigureCloudRoutes Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider:                  string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ConfigureCloudRoutes: fi.Bool(true),
				},
			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--configure-cloud-routes=true",
				"--use-service-account-credentials=true",
			},
		},
		{
			desc: "CIDRAllocatorType Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider:                  string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					CIDRAllocatorType: fi.String(string(ipam.RangeAllocatorType)),
				},
			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--cidr-allocator-type=RangeAllocator",
				"--use-service-account-credentials=true",
			},
		},
		{
			desc: "CIDRAllocatorType Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider:                  string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					UseServiceAccountCredentials: fi.Bool(false),
				},
			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=false",
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

func Test_executeTemplate(t *testing.T) {
	tests := []struct {
		desc                 string
		cluster              *kops.Cluster
		templateFilename     string
		expectedManifestPath string
	}{
		{
			desc: "test cloud controller template",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterName: "k8s",
					Image:       "docker.io/k8scloudprovider/openstack-cloud-controller-manager:1.13",
				},
			},
			},
			templateFilename:     "../../../models/cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.13.yaml.template",
			expectedManifestPath: "./tests/manifests/k8s-1.13.yaml",
		},
		{
			desc: "test cloud controller template",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterName: "k8s",
					Image:       "docker.io/k8scloudprovider/openstack-cloud-controller-manager:1.13",
				},
			},
			},
			templateFilename:     "../../../models/cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.11.yaml.template",
			expectedManifestPath: "./tests/manifests/k8s-1.11.yaml",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.desc, func(t *testing.T) {
			featureflag.EnableExternalCloudController = featureflag.New("TotalyNotEnableExternalCloudController", featureflag.Bool(true))
			templateFileAbsolutePath, filePathError := filepath.Abs(testCase.templateFilename)
			if filePathError != nil {
				t.Fatalf("error getting path to template: %v", filePathError)
			}

			tpl := template.New(filepath.Base(templateFileAbsolutePath))

			funcMap := make(template.FuncMap)
			templateFunctions := TemplateFunctions{cluster: testCase.cluster}
			templateFunctions.AddTo(funcMap, nil)

			tpl.Funcs(funcMap)

			tpl.Option("missingkey=zero")
			_, err := tpl.ParseFiles(templateFileAbsolutePath)
			if err != nil {
				t.Fatalf("error parsing template %q: %v", "template", err)
			}
			var buffer bytes.Buffer
			err = tpl.Execute(&buffer, testCase.cluster.Spec)
			if err != nil {
				t.Fatalf("error executing template %q: %v", "template", err)
			}
			actualManifest := buffer.Bytes()
			expectedFileAbsolutePath, _ := filepath.Abs(testCase.expectedManifestPath)
			expectedManifest, _ := ioutil.ReadFile(expectedFileAbsolutePath)

			actualString := strings.TrimSpace(string(actualManifest))
			expectedString := strings.TrimSpace(string(expectedManifest))
			if !reflect.DeepEqual(actualString, expectedString) {
				t.Fatalf("Manifests differs: %+v instead of %+v", actualString, expectedString)
			}
		})
	}
}
