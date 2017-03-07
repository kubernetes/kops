/*
Copyright 2016 The Kubernetes Authors.

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

package tasks

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/federation/targets/kubernetes"
	"k8s.io/kops/upup/pkg/fi"
)

//go:generate fitask -type=KubernetesResource
type KubernetesResource struct {
	Name *string

	Manifest *fi.ResourceHolder
}

//var _ fi.HasCheckExisting = &KubernetesResource{}
//
//// It's important always to check for the existing key, so we don't regenerate keys e.g. on terraform
//func (e *KubernetesResource) CheckExisting(c *fi.Context) bool {
//	return true
//}

func (e *KubernetesResource) Find(c *fi.Context) (*KubernetesResource, error) {
	// We always apply...
	// TODO: parse the existing kubectl apply annotations
	return nil, nil
}

func (e *KubernetesResource) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *KubernetesResource) CheckChanges(a, e, changes *KubernetesResource) error {
	return nil
}

func (_ *KubernetesResource) Render(c *fi.Context, a, e, changes *KubernetesResource) error {
	name := fi.StringValue(e.Name)
	if name == "" {
		return field.Required(field.NewPath("Name"), "")
	}

	target, ok := c.Target.(*kubernetes.KubernetesTarget)
	if !ok {
		return fmt.Errorf("Expected KubernetesTarget, got %T", c.Target)
	}

	manifestData, err := e.Manifest.AsBytes()
	if err != nil {
		return fmt.Errorf("error rending manifest template: %v", err)
	}

	err = target.Apply(manifestData)
	if err != nil {
		return fmt.Errorf("error applying manifest %q: %v", name, err)
	}

	return nil
}
