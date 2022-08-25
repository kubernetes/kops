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
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

func (t *TerraformTarget) finishHCL2() error {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	outputs, err := t.GetOutputs()
	if err != nil {
		return err
	}
	writeLocalsOutputs(rootBody, outputs)

	providerName := string(t.Cloud.ProviderID())
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		providerName = "google"
	}
	if t.Cloud.ProviderID() == kops.CloudProviderHetzner {
		providerName = "hcloud"
	}
	providerBlock := rootBody.AppendNewBlock("provider", []string{providerName})
	providerBody := providerBlock.Body()
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		providerBody.SetAttributeValue("project", cty.StringVal(t.Project))
	}
	if t.Cloud.ProviderID() != kops.CloudProviderHetzner {
		providerBody.SetAttributeValue("region", cty.StringVal(t.Cloud.Region()))
	}
	for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
		providerBody.SetAttributeValue(k, cty.StringVal(v))
	}
	rootBody.AppendNewline()

	if err := t.writeFilesProvider(rootBody); err != nil {
		return err
	}

	resourcesByType, err := t.GetResourcesByType()
	if err != nil {
		return err
	}

	resourceTypes := make([]string, 0, len(resourcesByType))
	for resourceType := range resourcesByType {
		resourceTypes = append(resourceTypes, resourceType)
	}
	sort.Strings(resourceTypes)
	for _, resourceType := range resourceTypes {
		resources := resourcesByType[resourceType]
		resourceNames := make([]string, 0, len(resources))
		for resourceName := range resources {
			resourceNames = append(resourceNames, resourceName)
		}
		sort.Strings(resourceNames)
		for _, resourceName := range resourceNames {
			item := resources[resourceName]

			resBlock := rootBody.AppendNewBlock("resource", []string{resourceType, resourceName})
			resBody := resBlock.Body()
			resType, err := gocty.ImpliedType(item)
			if err != nil {
				return err
			}
			resVal, err := gocty.ToCtyValue(item, resType)
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
	}

	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	terraformBody := terraformBlock.Body()
	terraformBody.SetAttributeValue("required_version", cty.StringVal(">= 0.15.0"))

	requiredProvidersBlock := terraformBody.AppendNewBlock("required_providers", []string{})
	requiredProvidersBody := requiredProvidersBlock.Body()

	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		writeMap(requiredProvidersBody, "google", map[string]cty.Value{
			"source":  cty.StringVal("hashicorp/google"),
			"version": cty.StringVal(">= 2.19.0"),
		})
	} else if t.Cloud.ProviderID() == kops.CloudProviderHetzner {
		writeMap(requiredProvidersBody, "hcloud", map[string]cty.Value{
			"source":  cty.StringVal("hetznercloud/hcloud"),
			"version": cty.StringVal(">= 1.35.1"),
		})
	} else if t.Cloud.ProviderID() == kops.CloudProviderAWS {
		configurationAliases := []*terraformWriter.Literal{terraformWriter.LiteralTokens("aws", "files")}
		aliasesType, err := gocty.ImpliedType(configurationAliases)
		if err != nil {
			return err
		}
		aliasesVal, err := gocty.ToCtyValue(configurationAliases, aliasesType)
		if err != nil {
			return err
		}
		writeMap(requiredProvidersBody, "aws", map[string]cty.Value{
			"source":                cty.StringVal("hashicorp/aws"),
			"version":               cty.StringVal(">= 4.0.0"),
			"configuration_aliases": aliasesVal,
		})
		if featureflag.Spotinst.Enabled() {
			writeMap(requiredProvidersBody, "spotinst", map[string]cty.Value{
				"source":  cty.StringVal("spotinst/spotinst"),
				"version": cty.StringVal(">= 1.33.0"),
			})
		}
	}

	bytes := hclwrite.Format(f.Bytes())
	t.Files["kubernetes.tf"] = bytes

	return nil
}

// writeLocalsOutputs creates the locals block and output blocks for all output variables
// Example:
//
//	locals {
//	  key1 = "value1"
//	  key2 = "value2"
//	}
//
//	output "key1" {
//	  value = "value1"
//	}
//
//	output "key2" {
//	  value = "value2"
//	}
func writeLocalsOutputs(body *hclwrite.Body, outputs map[string]terraformWriter.OutputValue) error {
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

	for _, tfName := range outputNames {
		v := outputs[tfName]
		outputBlock := body.AppendNewBlock("output", []string{tfName})
		outputBody := outputBlock.Body()
		if v.Value != nil {
			writeLiteral(outputBody, "value", v.Value)
			writeLiteral(localsBody, tfName, v.Value)
		} else {
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

// writeFilesProvider adds the second provider definition for managed files
func (t *TerraformTarget) writeFilesProvider(body *hclwrite.Body) error {
	if t.filesProvider == nil {
		return nil
	}
	providerBlock := body.AppendNewBlock("provider", []string{t.filesProvider.Name})
	providerBody := providerBlock.Body()
	providerBody.SetAttributeValue("alias", cty.StringVal("files"))
	for k, v := range t.filesProvider.Arguments {
		providerBody.SetAttributeValue(k, cty.StringVal(v))
	}
	for k, v := range tfGetFilesProviderExtraConfig(t.clusterSpecTarget) {
		providerBody.SetAttributeValue(k, cty.StringVal(v))
	}
	body.AppendNewline()
	return nil
}
