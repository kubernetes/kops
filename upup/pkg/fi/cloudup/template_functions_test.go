package cloudup

import (
	"k8s.io/kops/pkg/featureflag"
	"path/filepath"
	"io/ioutil"
	"bytes"
	"text/template"
	"k8s.io/kops/pkg/apis/kops"
	"reflect"
	"testing"
	"fmt"
	"strings"
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


func Test_executeTemplate(t *testing.T ) {
	tests := []struct{
			desc string
			cluster *kops.Cluster
			templateFilename string
			expectedManifestPath string
	}{
		{
			desc: "test cloud controller template",
			cluster:  &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: string(kops.CloudProviderOpenstack),
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterName: "k8s",
					Image: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:1.13",
				},
		},
	},
	templateFilename: "../../../models/cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.13.yaml.template",
	expectedManifestPath: "./tests/manifests/k8s-1.13.yaml",
},
{
		desc: "test cloud controller template",
		cluster:  &kops.Cluster{Spec: kops.ClusterSpec{
			CloudProvider: string(kops.CloudProviderOpenstack),
			ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
				ClusterName: "k8s",
				Image: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:1.13",
				},
			},
		},
		templateFilename: "../../../models/cloudup/resources/addons/openstack.addons.k8s.io/k8s-1.11.yaml.template",
		expectedManifestPath: "./tests/manifests/k8s-1.11.yaml",
		},
}

	for _, testCase := range tests {
			t.Run(testCase.desc, func(t *testing.T) {
				featureflag.EnableExternalCloudController = featureflag.New("TotalyNotEnableExternalCloudController", featureflag.Bool(true))

				fmt.Println("EnableExternalCloudController:", featureflag.EnableExternalCloudController.Enabled())

				templateFileAbsolutePath, filePathError :=  filepath.Abs(testCase.templateFilename)
				if filePathError != nil {
					t.Fatalf("error getting path to template: %v",  filePathError)
				}

				tpl := template.New(filepath.Base(templateFileAbsolutePath))

				funcMap := make(template.FuncMap)
				templateFunctions :=  TemplateFunctions{cluster: testCase.cluster}
				templateFunctions.AddTo(funcMap, nil)
				//funcMap["Args"] = func() []string {
				//	return args
				//}
				//funcMap["RenderResource"] = func(resourceName string, args []string) (string, error) {
				//	return l.renderResource(resourceName, args)
				//}
				// for k, fn := range l.TemplateFunctions {
				// 	funcMap[k] = fn
				// }
				// templateFunctions :=  TemplateFunctions{cluster: testCase.cluster}

				fmt.Println("funcMap",funcMap)
				fmt.Println("tpl:",tpl)
				tpl.Funcs(funcMap)

				tpl.Option("missingkey=zero")
				_, err := tpl.ParseFiles(templateFileAbsolutePath)
				if err != nil {
					t.Fatalf("error parsing template %q: %v",  "template", err)
				}
				var buffer bytes.Buffer
				err = tpl.Execute(&buffer, testCase.cluster.Spec)
				if err != nil {
					 t.Fatalf("error executing template %q: %v", "template", err)
				}
				actualManifest := buffer.Bytes()
				expectedFileAbsolutePath, _ := filepath.Abs(testCase.expectedManifestPath)
				expectedManifest, _ := ioutil.ReadFile( expectedFileAbsolutePath )

				actualString := strings.TrimSpace(string(actualManifest))
				expectedString := strings.TrimSpace(string(expectedManifest))
				if !reflect.DeepEqual(actualString,expectedString,) {
					t.Fatalf("Manifests differs: %+v instead of %+v", actualString, expectedString )
				}
			})
	}
}
