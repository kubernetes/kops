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
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"html"

	"github.com/go-openapi/loads"
)

var AllowErrors = flag.Bool("allow-errors", false, "If true, don't fail on errors.")
var ConfigDir = flag.String("config-dir", "", "Directory contain api files.")
var UseTags = flag.Bool("use-tags", false, "If true, use the openapi tags instead of the config yaml.")
var MungeGroups = flag.Bool("munge-groups", true, "If true, munge the group names for the operations to match.")

func (config *Config) genConfigFromTags(specs []*loads.Document) {
	log.Printf("Using openapi extension tags to configure.")

	config.ExampleLocation = "examples"
	// build the apis from the groups that are observed
	groupsMap := map[ApiGroup]DefinitionList{}
	for _, definition := range config.Definitions.GetAllDefinitions() {
		if strings.HasSuffix(definition.Name, "List") {
			continue
		}
		if strings.HasSuffix(definition.Name, "Status") {
			continue
		}
		if strings.HasPrefix(definition.Description(), "Deprecated. Please use") {
			// Don't look at deprecated types
			continue
		}
		config.initDefExample(definition) // Init the example yaml
		g := definition.Group
		groupsMap[g] = append(groupsMap[g], definition)
	}
	groupsList := ApiGroups{}
	for g := range groupsMap {
		groupsList = append(groupsList, g)
	}
	sort.Sort(groupsList)
	for _, g := range groupsList {
		groupName := strings.Title(string(g))
		config.ApiGroups = append(config.ApiGroups, ApiGroup(groupName))
		rc := ResourceCategory{}
		rc.Include = string(g)
		rc.Name = groupName
		defList := groupsMap[g]
		sort.Sort(defList)
		for _, d := range defList {
			r := &Resource{}
			r.Name = d.Name
			r.Group = string(d.Group)
			r.Version = string(d.Version)
			r.Definition = d
			rc.Resources = append(rc.Resources, r)
		}
		config.ResourceCategories = append(config.ResourceCategories, rc)
	}
}

func NewConfig() *Config {
	config := loadYamlConfig()
	specs := LoadOpenApiSpec()

	// Initialize all of the operations
	config.Definitions = GetDefinitions(specs)

	if *UseTags {
		// Initialize the config and ToC from the tags on definitions
		config.genConfigFromTags(specs)
	} else {
		// Initialization for ToC resources only
		vistToc := func(resource *Resource, definition *Definition) {
			definition.InToc = true // Mark as in Toc
			resource.Definition = definition
			config.initDefExample(definition) // Init the example yaml
		}
		config.VisitResourcesInToc(config.Definitions, vistToc)
	}

	// Get the map of operations appearing in the open-api spec keyed by id
	config.InitOperations(specs)


	// In the descriptions, replace unicode escape sequences with HTML entities.
	config.createDescriptionsWithEntities()

	config.CleanUp()

	// Prune anything that shouldn't be in the ToC
	if *UseTags {
		categories := []ResourceCategory{}
		for _, c := range config.ResourceCategories {
			resources := Resources{}
			for _, r := range c.Resources {
				if d, f := config.Definitions.GetByVersionKind(r.Group, r.Version, r.Name); f {
					if d.InToc {
						resources = append(resources, r)
					}
				}
			}
			c.Resources = resources
			if len(resources) > 0 {
				categories = append(categories, c)
			}
		}
		config.ResourceCategories = categories
	}

	return config
}

func verifyBlacklisted(operation Operation) {
	switch {
	case strings.Contains(operation.ID, "connectCoreV1Patch"):
	//case strings.Contains(operation.ID, "NamespacedScheduledJob"):
	//case strings.Contains(operation.ID, "ScheduledJobForAllNamespaces"):
	//case strings.Contains(operation.ID, "ScheduledJobListForAllNamespaces"):
	case strings.Contains(operation.ID, "V1beta1NamespacedReplicationControllerDummyScale"):
	case strings.Contains(operation.ID, "NamespacedPodAttach"):
	case strings.Contains(operation.ID, "NamespacedPodWithPath"):
	case strings.Contains(operation.ID, "proxyCoreV1"):
	//case strings.Contains(operation.ID, "NamespacedScaleScale"):
	//case strings.Contains(operation.ID, "NamespacedBindingBinding"):
	case strings.Contains(operation.ID, "NamespacedPodExe"):
	case strings.Contains(operation.ID, "logFileHandler"):
	case strings.Contains(operation.ID, "logFileListHandler"):
	case strings.Contains(operation.ID, "replaceCoreV1NamespaceFinalize"):
	//case strings.Contains(operation.ID, "NamespacedEvictionEviction"):
	case strings.Contains(operation.ID, "getCodeVersion"):
	case strings.Contains(operation.ID, "V1beta1CertificateSigningRequestApproval"):
	default:
		//panic(fmt.Sprintf("No Definition found for %s [%s].  \n", operation.ID, operation.Path))
		fmt.Printf("No Definition found for %s [%s].  \n", operation.ID, operation.Path)
	}
}

// /apis/<group>/<version>/namespaces/{namespace}/<resources>/{name}/<subresource>
var matchNamespaced = regexp.MustCompile(
	`^/apis/([A-Za-z0-9\.]+)/([A-Za-z0-9]+)/namespaces/\{namespace\}/([A-Za-z0-9\.]+)/\{name\}/([A-Za-z0-9\.]+)$`)
var matchUnnamespaced = regexp.MustCompile(
	`^/apis/([A-Za-z0-9\.]+)/([A-Za-z0-9]+)/([A-Za-z0-9\.]+)/\{name\}/([A-Za-z0-9\.]+)$`)

func GetMethod(o *Operation) string {
	switch o.HttpMethod {
	case "GET":
		return "List"
	case "POST":
		return "Create"
	case "PATCH":
		return "Patch"
	case "DELETE":
		return "Delete"
	case "PUT":
		return "Update"
	}
	return ""
}

func GetGroupVersionKindSub(o *Operation) (string, string, string, string) {
	if matchNamespaced.MatchString(o.Path) {
		m := matchNamespaced.FindStringSubmatch(o.Path)
		//fmt.Printf("Match %s\n", o.Path)
		group := m[1]
		group = strings.Split(group, ".")[0]
		version := m[2]
		resource := m[3]
		subresource := m[4]
		return group, version, resource, subresource

	} else if matchUnnamespaced.MatchString(o.Path) {
		m := matchUnnamespaced.FindStringSubmatch(o.Path)
		//fmt.Printf("Match %s\n", o.Path)
		group := m[1]
		version := m[2]
		resource := m[3]
		subresource := m[4]
		return group, version, resource, subresource
	}
	return "", "", "", ""
}

func GetResourceName(d *Definition) string {
	if len(d.Resource) > 0 {
		return d.Resource
	}
	resource := strings.ToLower(d.Name)
	if strings.HasSuffix(resource, "y") {
		return strings.TrimSuffix(resource, "y") + "ies"
	}
	return resource + "s"
}

func (config *Config) initOperationsFromTags(specs []*loads.Document) {
	if *UseTags {
		ops := map[string]map[string][]*Operation{}
		defs := map[string]*Definition{}
		for _, c := range config.Definitions.ByGroupVersionKind {
			defs[fmt.Sprintf("%s.%s.%s", c.Group, c.Version, GetResourceName(c))] = c
		}

		VisitOperations(specs, func(operation Operation) {
			if o, found := config.Operations[operation.ID]; found && o.Definition != nil {
				return
			}
			op := operation
			o := &op
			config.Operations[operation.ID] = o
			group, version, resource, sub := GetGroupVersionKindSub(o)
			if sub == "status" {
				return
			}
			if len(group) == 0 {
				return
			}
			key := fmt.Sprintf("%s.%s.%s", group, version, resource)
			o.Definition = defs[key]

			// Index by group and subresource
			if _, f := ops[key]; !f {
				ops[key] = map[string][]*Operation{}
			}
			ops[key][sub] = append(ops[key][sub], o)
		})

		for key, subMap := range ops {
			def := defs[key]
			if def == nil {
				panic(fmt.Errorf("Unable to locate resource %s in resource map\n%v\n", key, defs))
			}
			subs := []string{}
			for s := range subMap {
				subs = append(subs, s)
			}
			sort.Strings(subs)
			for _, s := range subs {
				cat := &OperationCategory{}
				cat.Name = strings.Title(s) + " Operations"
				oplist := subMap[s]
				for _, op := range oplist {
					ot := OperationType{}
					ot.Name = GetMethod(op) + " " + strings.Title(s)
					op.Type = ot
					cat.Operations = append(cat.Operations, op)
				}
				def.OperationCategories = append(def.OperationCategories, cat)
			}
		}
	}
}

// GetOperations returns all Operations found in the Documents
func (config *Config) InitOperations(specs []*loads.Document) {
	o := Operations{}

	config.GroupMap = map[string]string{}
	VisitOperations(specs, func(operation Operation) {
		o[operation.ID] = &operation

		// Build a map of the group names to the group name appearing in operation ids
		// This is necessary because the group will appear without the domain
		// in the resource, but with the domain in the operationID, and we
		// will be unable to match the operationID to the resource because they
		// don't agree on the name of the group.
		// TODO: Fix this by getting the group-version-kind in the resource
		if *MungeGroups {
			if v, f := operation.op.Extensions[typeKey]; f {
				gvk := v.(map[string]interface{})
				group, ok := gvk["group"].(string)
				if !ok {
					log.Fatalf("group not type string %v", v)
				}
				groupId := ""
				for _, s := range strings.Split(group, ".") {
					groupId = groupId + strings.Title(s)
				}
				config.GroupMap[strings.Title(strings.Split(group, ".")[0])] = groupId
			}
		}
	})
	config.Operations = o

	config.mapOperationsToDefinitions()
	config.initOperationsFromTags(specs)

	VisitOperations(specs, func(operation Operation) {
		if o, found := config.Operations[operation.ID]; !found || o.Definition == nil {
			verifyBlacklisted(operation)
		}
	})
	config.Definitions.initializeOperationParameters(config.Operations)

	// Clear the operations.  We still have to calculate the operations because that is how we determine
	// the API Group for each definition.
	if !*BuildOps {
		config.Operations = Operations{}
		config.OperationCategories = []OperationCategory{}
		for _, d := range config.Definitions.GetAllDefinitions() {
			d.OperationCategories = []*OperationCategory{}
		}
	}
}

// CleanUp sorts and dedups fields
func (c *Config) CleanUp() {
	for _, d := range c.Definitions.GetAllDefinitions() {
		sort.Sort(d.AppearsIn)
		sort.Sort(d.Fields)
		dedup := SortDefinitionsByName{}
		var last *Definition
		for _, i := range d.AppearsIn {
			if last != nil &&
				i.Name == last.Name &&
				i.Group.String() == last.Group.String() &&
				i.Version.String() == last.Version.String() {
				continue
			}
			last = i
			dedup = append(dedup, i)
		}
		d.AppearsIn = dedup
	}
}

// loadYamlConfig reads the config yaml file into a struct
func loadYamlConfig() *Config {
	f := filepath.Join(*ConfigDir, "config.yaml")

	config := &Config{}
	contents, err := ioutil.ReadFile(f)
	if err != nil {
		if !*UseTags {
			fmt.Printf("Failed to read yaml file %s: %v", f, err)
			os.Exit(2)
		}
	} else {
		err = yaml.Unmarshal(contents, config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	writeCategory := OperationCategory{
		Name: "Write Operations",
		OperationTypes: []OperationType{
			{
				Name:  "Create",
				Match: "create${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Create Eviction",
				Match: "create${group}${version}(Namespaced)?${resource}Eviction",
			},
			{
				Name:  "Patch",
				Match: "patch${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Replace",
				Match: "replace${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Delete",
				Match: "delete${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Delete Collection",
				Match: "delete${group}${version}Collection(Namespaced)?${resource}",
			},
		},
	}

	readCategory := OperationCategory{
		Name: "Read Operations",
		OperationTypes: []OperationType{
			{
				Name:  "Read",
				Match: "read${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "List",
				Match: "list${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "List All Namespaces",
				Match: "list${group}${version}(Namespaced)?${resource}ForAllNamespaces",
			},
			{
				Name:  "Watch",
				Match: "watch${group}${version}(Namespaced)?${resource}",
			},
			{
				Name:  "Watch List",
				Match: "watch${group}${version}(Namespaced)?${resource}List",
			},
			{
				Name:  "Watch List All Namespaces",
				Match: "watch${group}${version}(Namespaced)?${resource}ListForAllNamespaces",
			},
		},
	}

	statusCategory := OperationCategory{
		Name: "Status Operations",
		OperationTypes: []OperationType{
			{
				Name:  "Patch Status",
				Match: "patch${group}${version}(Namespaced)?${resource}Status",
			},
			{
				Name:  "Read Status",
				Match: "read${group}${version}(Namespaced)?${resource}Status",
			},
			{
				Name:  "Replace Status",
				Match: "replace${group}${version}(Namespaced)?${resource}Status",
			},
		},
	}

	config.OperationCategories = append([]OperationCategory{writeCategory, readCategory, statusCategory}, config.OperationCategories...)

	return config
}

// initOpExample reads the example config for each operation and sets it
func (config *Config) initOpExample(o *Operation) {
	path := o.Type.Name + ".yaml"
	path = filepath.Join(*ConfigDir, config.ExampleLocation, o.Definition.Name, path)
	path = strings.Replace(path, " ", "_", -1)
	path = strings.ToLower(path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(content, &o.ExampleConfig)
	if err != nil {
		panic(fmt.Sprintf("Could not Unmarshal ExampleConfig yaml: %s\n", content))
	}
}

func (config *Config) GetDefExampleFile(d *Definition) string {
	return strings.Replace(strings.ToLower(filepath.Join(*ConfigDir, config.ExampleLocation, d.Name, d.Name+".yaml")), " ", "_", -1)
}

func (config *Config) initDefExample(d *Definition) {
	content, err := ioutil.ReadFile(config.GetDefExampleFile(d))
	if err != nil || len(content) <= 0 {
		//fmt.Printf("Missing example: %s %v\n", d.Name, err)
		return
	}
	err = yaml.Unmarshal(content, &d.Sample)
	if err != nil {
		panic(fmt.Sprintf("Could not Unmarshal SampleConfig yaml: %s\n", content))
	}
}

func (config *Config) getOperationId(match string, group string, version ApiVersion, kind string) string {
	// Lookup the name of the group as the operation expects it (different than the resource)
	if g, f := config.GroupMap[group]; f {
		group = g
	}

	// Substitute the api definition group-version-kind into the operation template and look for a match
	v, k := doScaleIdHack(string(version), kind, match)
	match = strings.Replace(match, "${group}", string(group), -1)
	match = strings.Replace(match, "${version}", v, -1)
	match = strings.Replace(match, "${resource}", k, -1)
	return match
}

func (config *Config) setOperation(match, namespaceRep string,
	ot *OperationType, oc *OperationCategory, definition *Definition) {

	key := strings.Replace(match, "(Namespaced)?", namespaceRep, -1)
	if o, found := config.Operations[key]; found {
		// Each operation should have exactly 1 definition
		if o.Definition != nil {
			panic(fmt.Sprintf(
				"Found multiple matching defintions [%s/%s/%s, %s/%s/%s] for operation key: %s",
				definition.Group, definition.Version, definition.Name, o.Definition.Group, o.Definition.Version, o.Definition.Name, key))
		}
		o.Type = *ot
		o.Definition = definition
		oc.Operations = append(oc.Operations, o)
		config.initOpExample(o)

		// When using tags for the configuration, everything with an operation goes in the ToC
		if *UseTags && !o.Definition.IsOldVersion {
			o.Definition.InToc = true
		}
	}
}

// mapOperationsToDefinitions adds operations to the definitions they operate
// This is done by - for each definition - look at all potentially matching operations from operation categories
func (config *Config) mapOperationsToDefinitions() {
	// Look for matching operations for each definition
	for _, definition := range config.Definitions.GetAllDefinitions() {
		// Inlined definitions don't have operations
		if definition.IsInlined {
			continue
		}

		// Iterate through categories
		for i := range config.OperationCategories {
			oc := config.OperationCategories[i]

			// Iterate through possible operation matches
			for j := range oc.OperationTypes {
				// Iterate through possible api groups since we don't know the api group of the definition
				ot := oc.OperationTypes[j]

				operationId := config.getOperationId(ot.Match, definition.GetOperationGroupName(), definition.Version, definition.Name)
				// Look for a matching operation and set on the definition if found
				config.setOperation(operationId, "Namespaced", &ot, &oc, definition)
				config.setOperation(operationId, "", &ot, &oc, definition)
			}

			// If we found operations for this category, add the category to the definition
			if len(oc.Operations) > 0 {
				definition.OperationCategories = append(definition.OperationCategories, &oc)
			}
		}
	}
}

func doScaleIdHack(version, name, match string) (string, string) {
	// Hack to get around ids
	// if strings.HasSuffix(match, "${resource}Scale") && name != "Scale" {
	//
	//	fmt.Println()
	//	fmt.Println("doScaleIdHack: ", version, name,  match)

		// Scale names don't generate properly
	//	name = strings.ToLower(name) + "s"
	//	out := []rune(name)
	//	out[0] = unicode.ToUpper(out[0])
	//	name = string(out)
	// }
	out := []rune(version)
	out[0] = unicode.ToUpper(out[0])
	version = string(out)

	return version, name
}

func (config *Config) createDescriptionsWithEntities () {

	// The OpenAPI spec has escape sequences like \u003c. When the spec is unmarshaled,
	// the escape sequences get converted to ordinary characters. For example,
	// \u003c gets converted to a regular < character. But we can't use  regular <
	// and > characters in our HTML document. This function replaces these regular
	// characters with HTML entities: <, >, &, ', and ".

	for _, definition := range config.Definitions.GetAllDefinitions() {
		d := definition.Description()
		d = html.EscapeString(d)
		definition.DescriptionWithEntities = d

		for _,field := range definition.Fields {
			d := field.Description
			d = html.EscapeString(d)
			field.DescriptionWithEntities = d
		}
	}

	for _, operation := range config.Operations {

		for _, field := range operation.BodyParams {
			d := field.Description
			d = html.EscapeString(d)
			field.DescriptionWithEntities = d
		}

		for _, field := range operation.QueryParams {
			d := field.Description
			d = html.EscapeString(d)
			field.DescriptionWithEntities = d
		}

		for _, field := range operation.PathParams {
			d := field.Description
			d = html.EscapeString(d)
			field.DescriptionWithEntities = d
		}

		for _, resp := range operation.HttpResponses {
			d := resp.Description
			d = html.EscapeString(d)
			resp.DescriptionWithEntities = d
		}
	}
}
