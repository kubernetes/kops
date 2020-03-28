/*
Copyright 2020 The Kubernetes Authors.

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

package terraform

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

// Version represents which terraform version is targeted
type Version string

// Version011 represents terraform versions before 0.12
const Version011 Version = "0.11"

// Version012 represents terraform versions 0.12 and above
const Version012 Version = "0.12"

type TerraformTarget struct {
	Cloud   fi.Cloud
	Region  string
	Project string

	ClusterName string
	Version     Version

	outDir string

	// mutex protects the following items (resources & files)
	mutex sync.Mutex
	// resources is a list of TF items that should be created
	resources []*terraformResource
	// outputs is a list of our TF output variables
	outputs map[string]*terraformOutputVariable
	// files is a map of TF resource files that should be created
	files map[string][]byte
	// extra config to add to the provider block
	clusterSpecTarget *kops.TargetSpec
}

func NewTerraformTarget(cloud fi.Cloud, region, project string, outDir string, version Version, clusterSpecTarget *kops.TargetSpec) *TerraformTarget {
	return &TerraformTarget{
		Cloud:   cloud,
		Region:  region,
		Project: project,
		Version: version,

		outDir:            outDir,
		files:             make(map[string][]byte),
		outputs:           make(map[string]*terraformOutputVariable),
		clusterSpecTarget: clusterSpecTarget,
	}
}

var _ fi.Target = &TerraformTarget{}

type terraformResource struct {
	ResourceType string
	ResourceName string
	Item         interface{}
}

type byTypeAndName []*terraformResource

func (a byTypeAndName) Len() int { return len(a) }
func (a byTypeAndName) Less(i, j int) bool {
	return a[i].ResourceType+a[i].ResourceName < a[j].ResourceType+a[j].ResourceName
}
func (a byTypeAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type terraformOutputVariable struct {
	Key        string
	Value      *Literal
	ValueArray []*Literal
}

// A TF name can't have dots in it (if we want to refer to it from a literal),
// so we replace them
func tfSanitize(name string) string {
	if _, err := strconv.Atoi(string(name[0])); err == nil {
		panic(fmt.Sprintf("Terraform resource names cannot start with a digit. This is a bug in Kops, please report this in a GitHub Issue. Name: %v", name))
	}
	return strings.NewReplacer(".", "-", "/", "--", ":", "_").Replace(name)
}

func (t *TerraformTarget) AddFile(resourceType string, resourceName string, key string, r fi.Resource) (*Literal, error) {
	id := resourceType + "_" + resourceName + "_" + key

	d, err := fi.ResourceAsBytes(r)
	if err != nil {
		return nil, fmt.Errorf("error rending resource %s %v", id, err)
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	p := path.Join("data", id)
	t.files[p] = d

	modulePath := path.Join("${path.module}", p)
	l := LiteralFileExpression(modulePath)
	return l, nil
}

func (t *TerraformTarget) ProcessDeletions() bool {
	// Terraform tracks & performs deletions itself
	return false
}

func (t *TerraformTarget) RenderResource(resourceType string, resourceName string, e interface{}) error {
	res := &terraformResource{
		ResourceType: resourceType,
		ResourceName: resourceName,
		Item:         e,
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.resources = append(t.resources, res)

	return nil
}

func (t *TerraformTarget) AddOutputVariable(key string, literal *Literal) error {
	v := &terraformOutputVariable{
		Key:   key,
		Value: literal,
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.outputs[key] != nil {
		return fmt.Errorf("duplicate variable: %q", key)
	}
	t.outputs[key] = v

	return nil
}

func (t *TerraformTarget) AddOutputVariableArray(key string, literal *Literal) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.outputs[key] == nil {
		v := &terraformOutputVariable{
			Key: key,
		}
		t.outputs[key] = v
	}
	if t.outputs[key].Value != nil {
		return fmt.Errorf("variable %q is both an array and a scalar", key)
	}

	t.outputs[key].ValueArray = append(t.outputs[key].ValueArray, literal)

	return nil
}

// tfGetProviderExtraConfig is a helper function to get extra config with safety checks on the pointers.
func tfGetProviderExtraConfig(c *kops.TargetSpec) map[string]string {
	if c != nil &&
		c.Terraform != nil &&
		c.Terraform.ProviderExtraConfig != nil {
		return *c.Terraform.ProviderExtraConfig
	}
	return nil
}

func (t *TerraformTarget) Finish(taskMap map[string]fi.Task) error {
	var err error
	switch t.Version {
	case Version011:
		err = t.finish011(taskMap)
	case Version012:
		err = t.finish012(taskMap)
	default:
		err = fmt.Errorf("unrecognized terraform version %v", t.Version)
	}
	if err != nil {
		return err
	}

	for relativePath, contents := range t.files {
		p := path.Join(t.outDir, relativePath)

		err = os.MkdirAll(path.Dir(p), os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("error creating output directory %q: %v", path.Dir(p), err)
		}

		err = ioutil.WriteFile(p, contents, os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("error writing terraform data to output file %q: %v", p, err)
		}
	}
	klog.Infof("Terraform output is in %s", t.outDir)

	return nil
}
