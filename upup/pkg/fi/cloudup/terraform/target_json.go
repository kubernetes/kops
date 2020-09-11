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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func (t *TerraformTarget) finishJSON(taskMap map[string]fi.Task) error {
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
	} else if t.Cloud.ProviderID() == kops.CloudProviderAWS {
		providerAWS := make(map[string]interface{})
		providerAWS["region"] = t.Region
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			providerAWS[k] = v
		}
		providersByName["aws"] = providerAWS
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

	data := make(map[string]interface{})
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

	terraformConfiguration := make(map[string]interface{})
	terraformConfiguration["required_version"] = ">= 0.12.26"

	requiredProvidersByName := make(map[string]interface{})
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		requiredProviderGoogle := make(map[string]interface{})
		requiredProviderGoogle["source"] = "hashicorp/google"
		requiredProviderGoogle["version"] = ">= 3.44.0"
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			requiredProviderGoogle[k] = v
		}
		requiredProvidersByName["google"] = requiredProviderGoogle
	} else if t.Cloud.ProviderID() == kops.CloudProviderAWS {
		requiredProviderAWS := make(map[string]interface{})
		requiredProviderAWS["source"] = "hashicorp/aws"
		requiredProviderAWS["version"] = ">= 3.12.0"
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			requiredProviderAWS[k] = v
		}
		requiredProvidersByName["aws"] = requiredProviderAWS
	}

	if len(requiredProvidersByName) != 0 {
		terraformConfiguration["required_providers"] = requiredProvidersByName
	}

	data["terraform"] = terraformConfiguration

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling terraform data to json: %v", err)
	}

	t.files["kubernetes.tf.json"] = jsonBytes
	return nil
}
