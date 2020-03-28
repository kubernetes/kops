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
	"encoding/json"
	"fmt"
	"os"
	"path"

	hcl_parser "github.com/hashicorp/hcl/json/parser"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
)

func (t *TerraformTarget) finish011(taskMap map[string]fi.Task) error {
	resourcesByType := make(map[string]map[string]interface{})

	for _, res := range t.resources {
		resources := resourcesByType[res.ResourceType]
		if resources == nil {
			resources = make(map[string]interface{})
			resourcesByType[res.ResourceType] = resources
		}

		tfName := tfSanitize(res.ResourceName)

		if resources[tfName] != nil {
			return fmt.Errorf("duplicate resource found: %s.%s", res.ResourceType, tfName)
		}

		resources[tfName] = res.Item
	}

	providersByName := make(map[string]map[string]interface{})
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		providerGoogle := make(map[string]interface{})
		providerGoogle["project"] = t.Project
		providerGoogle["region"] = t.Region
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			providerGoogle[k] = v
		}
		providersByName["google"] = providerGoogle
		providerGoogle["version"] = ">= 3.0.0"
	} else if t.Cloud.ProviderID() == kops.CloudProviderAWS {
		providerAWS := make(map[string]interface{})
		providerAWS["region"] = t.Region
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			providerAWS[k] = v
		}
		providersByName["aws"] = providerAWS
	} else if t.Cloud.ProviderID() == kops.CloudProviderVSphere {
		providerVSphere := make(map[string]interface{})
		providerVSphere["region"] = t.Region
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			providerVSphere[k] = v
		}
		providersByName["vsphere"] = providerVSphere
	}

	outputVariables := make(map[string]interface{})
	for _, v := range t.outputs {
		tfName := tfSanitize(v.Key)

		if outputVariables[tfName] != nil {
			return fmt.Errorf("duplicate variable found: %s", tfName)
		}

		tfVar := make(map[string]interface{})
		if v.Value != nil {
			tfVar["value"] = v.Value
		} else {
			SortLiterals(v.ValueArray)
			deduped, err := DedupLiterals(v.ValueArray)
			if err != nil {
				return err
			}
			tfVar["value"] = deduped
		}
		outputVariables[tfName] = tfVar
	}

	localVariables := make(map[string]interface{})
	for _, v := range t.outputs {
		tfName := tfSanitize(v.Key)

		if localVariables[tfName] != nil {
			return fmt.Errorf("duplicate variable found: %s", tfName)
		}

		if v.Value != nil {
			localVariables[tfName] = v.Value
		} else {
			SortLiterals(v.ValueArray)
			deduped, err := DedupLiterals(v.ValueArray)
			if err != nil {
				return err
			}
			localVariables[tfName] = deduped
		}
	}

	// See https://github.com/kubernetes/kops/pull/2424 for why we require 0.9.3
	terraformConfiguration := make(map[string]interface{})
	if featureflag.TerraformJSON.Enabled() {
		terraformConfiguration["required_version"] = ">= 0.12.0"
	} else {
		terraformConfiguration["required_version"] = ">= 0.9.3"
	}

	data := make(map[string]interface{})
	data["terraform"] = terraformConfiguration
	data["resource"] = resourcesByType
	if len(providersByName) != 0 {
		data["provider"] = providersByName
	}
	if len(outputVariables) != 0 {
		data["output"] = outputVariables
	}
	if len(localVariables) != 0 {
		data["locals"] = localVariables
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling terraform data to json: %v", err)
	}

	if featureflag.TerraformJSON.Enabled() {
		t.files["kubernetes.tf.json"] = jsonBytes
		p := path.Join(t.outDir, "kubernetes.tf")
		if _, err := os.Stat(p); err == nil {
			return fmt.Errorf("Error generating kubernetes.tf.json: If you are upgrading from terraform 0.11 or earlier please read the release notes. Also, the kubernetes.tf file is already present. Please move the file away since it will be replaced by the kubernetes.tf.json file. ")
		}
	} else {
		f, err := hcl_parser.Parse(jsonBytes)
		if err != nil {
			return fmt.Errorf("error parsing terraform json: %v", err)
		}

		b, err := hclPrint(f)
		if err != nil {
			return fmt.Errorf("error writing terraform data to output: %v", err)
		}

		t.files["kubernetes.tf"] = b
	}
	return nil
}
