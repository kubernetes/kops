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

package api

import (
	"fmt"
)

type Config struct {
	ApiGroups           []ApiGroup          `yaml:"api_groups,omitempty"`
	ExampleLocation     string              `yaml:"example_location,omitempty"`
	OperationCategories []OperationCategory `yaml:"operation_categories,omitempty"`
	ResourceCategories  []ResourceCategory  `yaml:"resource_categories,omitempty"`

	// Used to map the group as the resource sees it to the group
	// as the operation sees it
	GroupMap            map[string]string

	Definitions Definitions
	Operations  Operations
}

// InlineDefinition defines a definition that should be inlined when displaying a Concept instead of appearing the in "Definitions"
type InlineDefinition struct {
	// Name is the name of the definition category
	Name string `yaml:",omitempty"`
	// Match the regular expression of defintion names that match this group where '${resource}' matches the resource name.
	// e.g. if Match == "${resource}Spec" then DeploymentSpec would be inlined into the "Deployment" Concept
	Match string `yaml:",omitempty"`
}

func (c Config) GetTopLevelConcepts() []string {
	s := []string{}
	for _, c := range c.ResourceCategories {
		for _, r := range c.Resources {
			s = append(s, r.Name)
		}
	}
	return s
}

/////////////////////////////////////////////////////
// Resources
/////////////////////////////////////////////////////
type Resources []*Resource

// ResourceCategory defines a category of Concepts
type ResourceCategory struct {
	// Name is the display name of this group
	Name string `yaml:",omitempty"`
	// Include is the name of the _resource.md file to include in the index.html.md
	Include string `yaml:",omitempty"`
	// Resources are the collection of Resources in this group
	Resources Resources `yaml:",omitempty"`
	// LinkToMd is the relative path to the md file containing the contents that clicking on this should link to
	LinkToMd string `yaml:"link_to_md,omitempty"`
}

type Resource struct {
	// Name is the display name of this Resource
	Name    string `yaml:",omitempty"`
	Version string `yaml:",omitempty"`
	Group   string `yaml:",omitempty"`
	// InlineDefinition is a list of definitions to show along side this resource when displaying it
	InlineDefinition []string `yaml:inline_definition",omitempty"`
	// DescriptionWarning is a warning message to show along side this resource when displaying it
	DescriptionWarning string `yaml:"description_warning,omitempty"`
	// DescriptionNote is a note message to show along side this resource when displaying it
	DescriptionNote string `yaml:"description_note,omitempty"`
	// ConceptGuide is a link to the concept guide for this resource if it exists
	ConceptGuide string `yaml:"concept_guide,omitempty"`
	// RelatedTasks is as list of tasks related to this concept
	RelatedTasks []string `yaml:"related_tasks,omitempty"`
	// IncludeDescription is the path to an md file to incline into the description
	IncludeDescription string `yaml:"include_description,omitempty"`
	// LinkToMd is the relative path to the md file containing the contents that clicking on this should link to
	LinkToMd string `yaml:"link_to_md,omitempty"`

	// Definition of the object
	Definition *Definition
}

type ExampleConfig struct {
	Name         string `yaml:",omitempty"`
	Namespace    string `yaml:",omitempty"`
	Request      string `yaml:",omitempty"`
	Response     string `yaml:",omitempty"`
	RequestNote  string `yaml:",omitempty"`
	ResponseNote string `yaml:",omitempty"`
}

type SampleConfig struct {
	Note   string `yaml:",omitempty"`
	Sample string `yaml:",omitempty"`
}

type ResourceVisitor func(resource *Resource, d *Definition)

// For each resource in the ToC, look up its definition and visit it.
func (c *Config) VisitResourcesInToc(definitions Definitions, fn ResourceVisitor) {
	missing := false
	for _, cat := range c.ResourceCategories {
		for _, resource := range cat.Resources {
			if definition, found := definitions.GetByVersionKind(resource.Group, resource.Version, resource.Name); found {
				fn(resource, definition)
			} else {
				fmt.Printf("Could not find definition for resource appearing in TOC: %s %s %s.\n", resource.Group, resource.Version, resource.Name)
				missing = true
			}
		}
	}
	if missing {
		fmt.Printf("All known definitions: %v\n", definitions.GetAllDefinitions())
	}
}
