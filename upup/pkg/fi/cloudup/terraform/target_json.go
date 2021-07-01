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
)

func (t *TerraformTarget) finishJSON() error {
	resourcesByType, err := t.GetResourcesByType()
	if err != nil {
		return err
	}

	providersByName := make(map[string]map[string]interface{})
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		providerGoogle := make(map[string]interface{})
		providerGoogle["project"] = t.Project
		providerGoogle["region"] = t.Cloud.Region()
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			providerGoogle[k] = v
		}
		providersByName["google"] = providerGoogle
	} else if t.Cloud.ProviderID() == kops.CloudProviderAWS {
		providerAWS := make(map[string]interface{})
		providerAWS["region"] = t.Cloud.Region()
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			providerAWS[k] = v
		}
		providersByName["aws"] = providerAWS
	}

	outputs, err := t.GetOutputs()
	if err != nil {
		return err
	}
	outputVariables := make(map[string]interface{})
	localVariables := make(map[string]interface{})
	for tfName, v := range outputs {
		var tfVar interface{}
		if v.Value != nil {
			tfVar = v.Value
		} else {
			tfVar = v.ValueArray
		}
		outputVariables[tfName] = map[string]interface{}{"value": tfVar}
		localVariables[tfName] = tfVar
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
		requiredProviderGoogle["version"] = ">= 2.19.0"
		for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
			requiredProviderGoogle[k] = v
		}
		requiredProvidersByName["google"] = requiredProviderGoogle
	} else if t.Cloud.ProviderID() == kops.CloudProviderAWS {
		requiredProviderAWS := make(map[string]interface{})
		requiredProviderAWS["source"] = "hashicorp/aws"
		requiredProviderAWS["version"] = ">= 2.46.0"
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

	t.Files["kubernetes.tf.json"] = jsonBytes
	return nil
}
