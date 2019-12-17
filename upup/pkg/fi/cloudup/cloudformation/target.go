/*
Copyright 2017 The Kubernetes Authors.

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

package cloudformation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
)

type CloudformationTarget struct {
	Cloud   fi.Cloud
	Region  string
	Project string

	outDir string

	// mutex protects the following items (resources & files)
	mutex     sync.Mutex
	resources map[string]*cloudformationResource
}

func NewCloudformationTarget(cloud fi.Cloud, region, project string, outDir string) *CloudformationTarget {
	return &CloudformationTarget{
		Cloud:     cloud,
		Region:    region,
		Project:   project,
		outDir:    outDir,
		resources: make(map[string]*cloudformationResource),
	}
}

var _ fi.Target = &CloudformationTarget{}

type cloudformationResource struct {
	Type       string
	Properties interface{}
}

// A cloudformation resource name must be alphanumeric
func sanitizeCloudformationResourceName(name string) string {
	name = strings.Replace(name, ".", "", -1)
	name = strings.Replace(name, "-", "", -1)
	name = strings.Replace(name, ":", "", -1)
	name = strings.Replace(name, "/", "", -1)
	return name
}

func (t *CloudformationTarget) ProcessDeletions() bool {
	// Terraform tracks & performs deletions itself
	return false
}

func (t *CloudformationTarget) RenderResource(resourceType string, resourceName string, e interface{}) error {
	res := &cloudformationResource{
		Type:       resourceType,
		Properties: e,
	}

	name := resourceType + "::" + resourceName
	name = sanitizeCloudformationResourceName(name)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.resources[name] != nil {
		return fmt.Errorf("resource %q already exists in cloudformation", name)
	}
	t.resources[name] = res

	return nil
}

func (t *CloudformationTarget) Find(ref *Literal) (interface{}, bool) {
	key := ref.extractRef()
	if key == "" {
		klog.Warningf("Unable to extract ref from %v", ref)
		return nil, false
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	r := t.resources[key]
	if r == nil {
		return nil, false
	}
	return r.Properties, true
}

func (t *CloudformationTarget) Finish(taskMap map[string]fi.Task) error {
	//resourcesByType := make(map[string]map[string]interface{})
	//
	//for _, res := range t.resources {
	//	resources := resourcesByType[res.ResourceType]
	//	if resources == nil {
	//		resources = make(map[string]interface{})
	//		resourcesByType[res.ResourceType] = resources
	//	}
	//
	//	tfName := tfSanitize(res.ResourceName)
	//
	//	if resources[tfName] != nil {
	//		return fmt.Errorf("duplicate resource found: %s.%s", res.ResourceType, tfName)
	//	}
	//
	//	resources[tfName] = res.Item
	//}

	//providersByName := make(map[string]map[string]interface{})
	//if t.Cloud.ProviderID() == kops.CloudProviderGCE {
	//	providerGoogle := make(map[string]interface{})
	//	providerGoogle["project"] = t.Project
	//	providerGoogle["region"] = t.Region
	//	providersByName["google"] = providerGoogle
	//} else if t.Cloud.ProviderID() == kops.CloudProviderAWS {
	//	providerAWS := make(map[string]interface{})
	//	providerAWS["region"] = t.Region
	//	providersByName["aws"] = providerAWS
	//}

	data := make(map[string]interface{})
	data["Resources"] = t.resources
	//if len(providersByName) != 0 {
	//	data["provider"] = providersByName
	//}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling cloudformation data to json: %v", err)
	}

	files := make(map[string][]byte)
	files["kubernetes.json"] = jsonBytes

	for relativePath, contents := range files {
		p := path.Join(t.outDir, relativePath)

		err = os.MkdirAll(path.Dir(p), os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("error creating output directory %q: %v", path.Dir(p), err)
		}

		err = ioutil.WriteFile(p, contents, os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("error writing cloudformation data to output file %q: %v", p, err)
		}
	}

	klog.Infof("Cloudformation output is in %s", t.outDir)

	return nil
}
