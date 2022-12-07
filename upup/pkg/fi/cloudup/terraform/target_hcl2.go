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
	"bytes"
	"fmt"
	"sort"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

func (t *TerraformTarget) finishHCL2() error {
	buf := &bytes.Buffer{}

	outputs, err := t.GetOutputs()
	if err != nil {
		return err
	}
	writeLocalsOutputs(buf, outputs)

	providerName := string(t.Cloud.ProviderID())
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		providerName = "google"
	}
	if t.Cloud.ProviderID() == kops.CloudProviderHetzner {
		providerName = "hcloud"
	}
	providerBody := map[string]string{}
	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		providerBody["project"] = t.Project
	}
	if t.Cloud.ProviderID() != kops.CloudProviderHetzner {
		providerBody["region"] = t.Cloud.Region()
	}
	for k, v := range tfGetProviderExtraConfig(t.clusterSpecTarget) {
		providerBody[k] = v
	}
	mapToElement(providerBody).
		ToObject().
		Write(buf, 0, fmt.Sprintf("provider %q", providerName))
	buf.WriteString("\n")

	// Add the second provider definition for managed files
	if t.filesProvider != nil {
		providerBody := map[string]string{}
		providerBody["alias"] = "files"
		for k, v := range t.filesProvider.Arguments {
			providerBody[k] = v
		}
		for k, v := range tfGetFilesProviderExtraConfig(t.clusterSpecTarget) {
			providerBody[k] = v
		}
		mapToElement(providerBody).
			ToObject().
			Write(buf, 0, fmt.Sprintf("provider %q", t.filesProvider.Name))
		buf.WriteString("\n")
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
			element := toElement(resources[resourceName])
			element.Write(buf, 0, fmt.Sprintf("resource %q %q", resourceType, resourceName))
			buf.WriteString("\n")
		}
	}

	dataSourcesByType, err := t.GetDataSourcesByType()
	if err != nil {
		return err
	}

	dataSourceTypes := make([]string, 0, len(dataSourcesByType))
	for dataSourceType := range dataSourcesByType {
		dataSourceTypes = append(dataSourceTypes, dataSourceType)
	}
	sort.Strings(dataSourceTypes)
	for _, dataSourceType := range dataSourceTypes {
		dataSources := dataSourcesByType[dataSourceType]
		dataSourceNames := make([]string, 0, len(dataSources))
		for dataSourceName := range dataSources {
			dataSourceNames = append(dataSourceNames, dataSourceName)
		}
		sort.Strings(dataSourceNames)
		for _, dataSourceName := range dataSourceNames {
			element := toElement(dataSources[dataSourceName])
			element.Write(buf, 0, fmt.Sprintf("data %q %q", dataSourceType, dataSourceName))
			buf.WriteString("\n")
		}
	}

	buf.WriteString("terraform {\n")
	buf.WriteString("  required_version = \">= 0.15.0\"\n")
	buf.WriteString("  required_providers {\n")

	if t.Cloud.ProviderID() == kops.CloudProviderGCE {
		mapToElement(map[string]*terraformWriter.Literal{
			"source":  terraformWriter.LiteralFromStringValue("hashicorp/google"),
			"version": terraformWriter.LiteralFromStringValue(">= 2.19.0"),
		}).Write(buf, 4, "google")
	} else if t.Cloud.ProviderID() == kops.CloudProviderHetzner {
		mapToElement(map[string]*terraformWriter.Literal{
			"source":  terraformWriter.LiteralFromStringValue("hetznercloud/hcloud"),
			"version": terraformWriter.LiteralFromStringValue(">= 1.35.1"),
		}).Write(buf, 4, "hcloud")
	} else if t.Cloud.ProviderID() == kops.CloudProviderAWS {
		configurationAlias := terraformWriter.LiteralTokens("aws", "files")
		mapToElement(map[string]*terraformWriter.Literal{
			"source":                terraformWriter.LiteralFromStringValue("hashicorp/aws"),
			"version":               terraformWriter.LiteralFromStringValue(">= 4.0.0"),
			"configuration_aliases": terraformWriter.LiteralListExpression(configurationAlias),
		}).Write(buf, 4, "aws")
		if featureflag.Spotinst.Enabled() {
			mapToElement(map[string]*terraformWriter.Literal{
				"source":  terraformWriter.LiteralFromStringValue("spotinst/spotinst"),
				"version": terraformWriter.LiteralFromStringValue(">= 1.33.0"),
			}).Write(buf, 4, "spotinst")
		}
	}

	buf.WriteString("  }\n")
	buf.WriteString("}\n")

	t.Files["kubernetes.tf"] = buf.Bytes()

	return nil
}

type output struct {
	Value *terraformWriter.Literal
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
func writeLocalsOutputs(buf *bytes.Buffer, outputs map[string]terraformWriter.OutputValue) {
	if len(outputs) == 0 {
		return
	}

	outputNames := make([]string, 0, len(outputs))
	locals := make(map[string]*terraformWriter.Literal, len(outputs))
	for k, v := range outputs {
		if _, ok := locals[k]; ok {
			panic(fmt.Sprintf("duplicate variable found: %s", k))
		}
		if v.Value != nil {
			locals[k] = v.Value
		} else {
			locals[k] = terraformWriter.LiteralListExpression(v.ValueArray...)
		}
		outputNames = append(outputNames, k)
	}
	sort.Strings(outputNames)

	mapToElement(locals).
		ToObject().
		Write(buf, 0, "locals")
	buf.WriteString("\n")

	for _, tfName := range outputNames {
		toElement(&output{Value: locals[tfName]}).Write(buf, 0, fmt.Sprintf("output %q", tfName))
		buf.WriteString("\n")
	}
	return
}
