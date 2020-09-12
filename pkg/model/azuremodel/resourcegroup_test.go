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

package azuremodel

import (
	"testing"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

func TestResourceGroupModelBuilder_Build(t *testing.T) {
	b := ResourceGroupModelBuilder{
		AzureModelContext: newTestAzureModelContext(),
	}
	c := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}
	err := b.Build(c)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
	if len(c.Tasks) != 1 {
		t.Errorf("unexpected number of tasks: %s", c.Tasks)
	}
	var task fi.Task
	for _, t := range c.Tasks {
		task = t
		break
	}
	rg, ok := task.(*azuretasks.ResourceGroup)
	if !ok {
		t.Fatalf("unexpected type of task: %T", t)
	}
	if !*rg.Shared {
		t.Errorf("unexpected shared")
	}
}
