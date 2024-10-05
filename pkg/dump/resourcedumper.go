/*
Copyright 2023 The Kubernetes Authors.

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

package dump

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"slices"

	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	resourceDumpConcurrency = 20
)

var (
	ignoredResources = map[string]struct{}{
		"componentstatuses":      {},
		"podtemplates":           {},
		"replicationcontrollers": {},
		"controllerrevisions":    {},
	}
)

type gvrNamespace struct {
	namespace string
	gvr       schema.GroupVersionResource
}

func (d *gvrNamespace) String() string {
	var gr string
	if d.gvr.Group == "" {
		gr = d.gvr.Resource
	} else {
		gr = fmt.Sprintf("%v.%v", d.gvr.Group, d.gvr.Resource)
	}
	return path.Join(d.namespace, gr)
}

type resourceDumper struct {
	k8sConfig     *rest.Config
	dynamicClient *dynamic.DynamicClient
	output        string
	artifactsDir  string
}

type resourceDumpResult struct {
	err error
}

func NewResourceDumper(k8sConfig *rest.Config, output, artifactsDir string) (*resourceDumper, error) {
	k8sConfig.QPS = 50
	k8sConfig.Burst = 100
	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("creating dynamic client: %w", err)
	}
	return &resourceDumper{
		k8sConfig:     k8sConfig,
		dynamicClient: dynamicClient,
		output:        output,
		artifactsDir:  artifactsDir,
	}, nil
}

func (d *resourceDumper) DumpResources(ctx context.Context) error {
	klog.Info("Dumping k8s resources")
	clientSet, err := kubernetes.NewForConfig(d.k8sConfig)
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}

	namespaces, err := clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing namespaces: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(d.k8sConfig)
	if err != nil {
		return fmt.Errorf("creating discovery client: %w", err)
	}

	resourceLists, err := discoveryClient.ServerPreferredResources()
	var discoveryErr *discovery.ErrGroupDiscoveryFailed
	if errors.As(err, &discoveryErr) {
		klog.Warningf("using incomplete list of API groups: %v", discoveryErr)
	} else if err != nil {
		return fmt.Errorf("listing server preferred resources: %w", err)
	}

	gvrNamespaces, err := getGVRNamespaces(resourceLists, namespaces.Items)
	if err != nil {
		return fmt.Errorf("getting GVR namespaces: %w", err)
	}

	jobs := make(chan gvrNamespace, len(gvrNamespaces))
	results := make(chan resourceDumpResult, len(gvrNamespaces))

	for i := 0; i < resourceDumpConcurrency; i++ {
		go d.dumpGVRNamespaces(ctx, jobs, results)
	}

	var dumpErr error

	for _, gvrn := range gvrNamespaces {
		jobs <- gvrn
	}
	close(jobs)

	for i := 0; i < len(gvrNamespaces); i++ {
		result := <-results
		if result.err != nil {
			errors.Join(dumpErr, result.err)
		}
	}
	close(results)
	return dumpErr
}

func getGVRNamespaces(resourceLists []*metav1.APIResourceList, namespaces []v1.Namespace) ([]gvrNamespace, error) {
	gvrNamespaces := make([]gvrNamespace, 0)
	for _, resourceList := range resourceLists {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return nil, err
		}
		for _, apiResource := range resourceList.APIResources {
			if _, ok := ignoredResources[apiResource.Name]; ok || !slices.Contains(apiResource.Verbs, "list") {
				continue
			}
			if apiResource.Namespaced {
				for _, ns := range namespaces {
					gvrNamespaces = append(gvrNamespaces, gvrNamespace{
						gvr: schema.GroupVersionResource{
							Group:    gv.Group,
							Version:  gv.Version,
							Resource: apiResource.Name,
						},
						namespace: ns.Name,
					})
				}
			} else {
				gvrNamespaces = append(gvrNamespaces, gvrNamespace{
					gvr: schema.GroupVersionResource{
						Group:    gv.Group,
						Version:  gv.Version,
						Resource: apiResource.Name,
					},
				})
			}
		}
	}
	return gvrNamespaces, nil
}

func (d *resourceDumper) dumpGVRNamespaces(ctx context.Context, jobs chan gvrNamespace, results chan resourceDumpResult) {
	for job := range jobs {
		var lister dynamic.ResourceInterface
		if job.namespace != "" {
			lister = d.dynamicClient.Resource(job.gvr).Namespace(job.namespace)
		} else {
			lister = d.dynamicClient.Resource(job.gvr)
		}
		resourceList, err := lister.List(ctx, metav1.ListOptions{})
		if err != nil {
			var statusErr *k8sErrors.StatusError
			if errors.As(err, &statusErr) && statusErr.ErrStatus.Code >= 400 && statusErr.ErrStatus.Code < 500 {
				continue
			}
			results <- resourceDumpResult{
				err: fmt.Errorf("listing resources for %v: %w", job, err),
			}
			continue
		}
		resPath := path.Join(d.artifactsDir, "cluster-info", fmt.Sprintf("%v.%v", job.String(), d.output))
		err = os.MkdirAll(path.Dir(resPath), 0755)
		if err != nil {
			results <- resourceDumpResult{
				err: fmt.Errorf("creating directory %q: %w", resPath, err),
			}
			continue
		}
		resFile, err := os.Create(resPath)
		if err != nil {
			results <- resourceDumpResult{
				err: fmt.Errorf("creating file %q: %w", resPath, err),
			}
			continue
		}

		err = resourceList.EachListItem(func(obj runtime.Object) error {
			o, err := meta.Accessor(obj)
			if err != nil {
				return err
			}
			o.SetManagedFields(nil)
			if err := maskObject(obj); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			results <- resourceDumpResult{
				err: fmt.Errorf("creating accessor for %v: %w", job, err),
			}
			continue
		}
		contents, err := resourceList.MarshalJSON()
		if err != nil {
			results <- resourceDumpResult{
				err: fmt.Errorf("marshaling to json for %v: %w", job, err),
			}
			continue
		}

		switch d.output {
		case "yaml":
			contents, err = yaml.JSONToYAML(contents)
			if err != nil {
				results <- resourceDumpResult{
					err: fmt.Errorf("marshaling to yaml for %v: %w", job, err),
				}
				continue
			}
		}
		_, err = resFile.Write(contents)
		if err != nil {
			results <- resourceDumpResult{
				err: fmt.Errorf("encoding resources for %v: %w", job, err),
			}
			continue
		}
		results <- resourceDumpResult{}
	}
}

func maskObject(obj runtime.Object) error {
	switch obj.GetObjectKind().GroupVersionKind() {
	case schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"}:
		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		data, ok, err := unstructured.NestedMap(unstructuredObj, "data")
		if err != nil {
			return fmt.Errorf("getting data from secret: %w", err)
		}
		if ok {
			for k := range data {
				data[k] = "REDACTED"
			}
			unstructured.SetNestedMap(unstructuredObj, data, "data")
		}

	}
	return nil
}
