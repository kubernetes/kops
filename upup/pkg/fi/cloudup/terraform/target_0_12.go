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
	"sort"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func (t *TerraformTarget) finish012(taskMap map[string]fi.Task) error {
	resourcesByType := make(map[string]map[string]interface{})

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	writeLocalsOutputs(rootBody, t.outputs)

	providerName := string(t.Cloud.ProviderID())
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		providerName = "google"
	}
	providerBlock := rootBody.AppendNewBlock("provider", []string{providerName})
	providerBody := providerBlock.Body()
	providerBody.SetAttributeValue("region", cty.StringVal(t.Region))
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		providerBody.SetAttributeValue("version", cty.StringVal(">= 3.0.0"))
	}
	for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
		providerBody.SetAttributeValue(k, cty.StringVal(v))
	}
	rootBody.AppendNewline()

	sort.Sort(byTypeAndName(t.resources))
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

		resBlock := rootBody.AppendNewBlock("resource", []string{res.ResourceType, tfName})
		resBody := resBlock.Body()
		resType, err := gocty.ImpliedType(res.Item)
		if err != nil {
			return err
		}
		resVal, err := gocty.ToCtyValue(res.Item, resType)
		if err != nil {
			return err
		}
		if resVal.IsNull() {
			continue
		}
		resVal.ForEachElement(func(key cty.Value, value cty.Value) bool {
			writeValue(resBody, key.AsString(), value)
			return false
		})
		rootBody.AppendNewline()
	}

	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	terraformBody := terraformBlock.Body()
	terraformBody.SetAttributeValue("required_version", cty.StringVal(">= 0.12.0"))

	bytes := hclwrite.Format(f.Bytes())
	t.files["kubernetes.tf"] = bytes

	return nil
}

// writeLocalsOutputs creates the locals block and output blocks for all output variables
// Example:
// locals {
//   key1 = "value1"
//   key2 = "value2"
// }
// output "key1" {
//   value = "value1"
// }
// output "key2" {
//   value = "value2"
// }
func writeLocalsOutputs(body *hclwrite.Body, outputs map[string]*terraformOutputVariable) error {
	if len(outputs) == 0 {
		return nil
	}

	localsBlock := body.AppendNewBlock("locals", []string{})
	body.AppendNewline()
	// each output is added to a single locals block and its own output block
	localsBody := localsBlock.Body()
	existingOutputVars := make(map[string]bool)

	outputNames := make([]string, 0, len(outputs))
	for k := range outputs {
		outputNames = append(outputNames, k)
	}
	sort.Strings(outputNames)

	for _, n := range outputNames {
		v := outputs[n]
		tfName := tfSanitize(v.Key)
		outputBlock := body.AppendNewBlock("output", []string{tfName})
		outputBody := outputBlock.Body()
		if v.Value != nil {
			writeLiteral(outputBody, "value", v.Value)
			writeLiteral(localsBody, tfName, v.Value)
		} else {
			SortLiterals(v.ValueArray)
			writeLiteralList(outputBody, "value", v.ValueArray)
			writeLiteralList(localsBody, tfName, v.ValueArray)
		}

		if existingOutputVars[tfName] {
			return fmt.Errorf("duplicate variable found: %s", tfName)
		}
		existingOutputVars[tfName] = true
		body.AppendNewline()
	}
	return nil
}
