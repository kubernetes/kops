package terraform

import (
	"encoding/json"
	"fmt"
	"io"
	"k8s.io/kube-deploy/upup/pkg/fi"
)

type TerraformTarget struct {
	Region    string
	Project   string
	resources []*terraformResource

	out io.Writer
}

func NewTerraformTarget(region, project string, out io.Writer) *TerraformTarget {
	return &TerraformTarget{
		Region:  region,
		Project: project,
		out:     out,
	}
}

var _ fi.Target = &TerraformTarget{}

type terraformResource struct {
	ResourceType string
	ResourceName string
	Item         interface{}
}

func (t *TerraformTarget) RenderResource(resourceType string, resourceName string, e interface{}) error {
	res := &terraformResource{
		ResourceType: resourceType,
		ResourceName: resourceName,
		Item:         e,
	}

	t.resources = append(t.resources, res)

	return nil
}

func (t *TerraformTarget) Finish(taskMap map[string]fi.Task) error {
	resourcesByType := make(map[string]map[string]interface{})

	for _, res := range t.resources {
		resources := resourcesByType[res.ResourceType]
		if resources == nil {
			resources = make(map[string]interface{})
			resourcesByType[res.ResourceType] = resources
		}

		if resources[res.ResourceName] != nil {
			return fmt.Errorf("duplicate resource found: %s.%s", res.ResourceType, res.ResourceName)
		}

		resources[res.ResourceName] = res.Item
	}

	providersByName := make(map[string]map[string]interface{})
	providerGoogle := make(map[string]interface{})
	providerGoogle["project"] = t.Project
	providerGoogle["region"] = t.Region
	providersByName["google"] = providerGoogle

	data := make(map[string]interface{})
	data["resource"] = resourcesByType
	data["provider"] = providersByName

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling terraform data to json: %v", err)
	}

	_, err = t.out.Write(jsonBytes)
	if err != nil {
		return fmt.Errorf("error writing terraform data to output: %v", err)
	}
	return nil
}
