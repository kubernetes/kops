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

package generators

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/kubernetes-incubator/reference-docs/gen-apidocs/generators/api"
)

func WriteTemplates(config *api.Config) {
	if _, err := os.Stat(*api.ConfigDir + "/includes"); os.IsNotExist(err) {
		os.Mkdir(*api.ConfigDir+"/includes", os.FileMode(0700))
	}

	// Write the index file importing each of the top level concept files
	WriteIndexFile(config)

	//// Write each concept file imported by the index file
	WriteConceptFiles(config)

	//// Write each definition file imported by the index file
	WriteDefinitionFiles(config)
}

func getStaticIncludesDir() string {
	return filepath.Join(*api.ConfigDir, "static_includes")
}

func WriteStaticFile(title, location string) {
	staticFileTemplate, err := template.New("static-file-template").Parse(DefaultHeader)
	if err != nil {
		panic(fmt.Errorf("Could not parse %v %s", err, DefaultHeader))
	}

	f := filepath.Join(getStaticIncludesDir(), location)
	_, err = os.Stat(f)
	if err == nil {
		// Don't create the file if it exists
		return
	}

	if !os.IsNotExist(err) {
		panic(fmt.Sprintf("Could not stat file %s %v", f, err))
	}
	fmt.Printf("Creating %s file\n", f)
	file, err := os.Create(f)
	if err != nil {
		panic(err)
	}
	file.Close()

	file, err = os.OpenFile(f, os.O_WRONLY, 0)
	if err != nil {
		fmt.Println(err)
	}
	err = staticFileTemplate.Execute(file, title)
	if err != nil {
		fmt.Println(err)
	}
	file.Close()
}

func WriteIndexFile(config *api.Config) {
	includes := []string{}

	manifest := Manifest{}

	manifest.Copyright = "<a href=\"https://github.com/kubernetes/kubernetes\">Copyright 2016 The Kubernetes Authors.</a>"

	WriteStaticFile("Overview", "_overview.md")
	WriteStaticFile("Old Versions", "_oldversions.md")
	WriteStaticFile("Definitions", "_definitions.md")
	for _, c := range config.ResourceCategories {
		name := "_" + c.Include + ".md"
		WriteStaticFile(c.Include, name)
	}

	if !*api.BuildOps {
		manifest.Title = "Kubernetes Resource Reference Docs"
	} else {
		manifest.Title = "Kubernetes API Reference Docs"
		manifest.Docs = append(manifest.Docs, Doc{"_overview.md"})
	}

	// Copy over the includes
	err := filepath.Walk(getStaticIncludesDir(), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			to := filepath.Join(*api.ConfigDir, "includes", filepath.Base(path))
			return os.Link(path, to)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Failed to copy includes %v.\n", err)
		return
	}

	// Add Toc Imports
	for _, c := range config.ResourceCategories {
		includes = append(includes, c.Include)
		name := "_" + c.Include + ".md"
		manifest.Docs = append(manifest.Docs, Doc{name})
		WriteStaticFile(c.Include, name)

		for _, r := range c.Resources {
			if r.Definition == nil {
				fmt.Printf("Warning: Missing definition for item in ToC %s\n", r.Name)
				continue
			}

			includes = append(includes, GetConceptImport(r.Definition))
			manifest.Docs = append(manifest.Docs, Doc{"_" + GetConceptImport(r.Definition) + ".md"})
		}
	}

	// Add other definition imports
	definitions := api.SortDefinitionsByName{}
	for _, definition := range config.Definitions.GetAllDefinitions() {

		// Don't add definitions for top level resources in the toc or inlined resources
		if definition.InToc || definition.IsInlined || definition.IsOldVersion {
			continue
		}
		definitions = append(definitions, definition)
	}
	sort.Sort(definitions)
	manifest.Docs = append(manifest.Docs, Doc{"_definitions.md"})
	includes = append(includes, "definitions")
	for _, d := range definitions {
		//definitions[i] = GetDefinitionImport(name)
		manifest.Docs = append(manifest.Docs, Doc{"_" + GetDefinitionImport(d) + ".md"})
		includes = append(includes, GetDefinitionImport(d))
	}

	// Add definitions for older version of objects
	definitions = api.SortDefinitionsByName{}
	for _, definition := range config.Definitions.GetAllDefinitions() {
		// Don't add definitions for top level resources in the toc or inlined resources
		if definition.IsOldVersion {
			definitions = append(definitions, definition)
		}
	}
	sort.Sort(definitions)
	manifest.Docs = append(manifest.Docs, Doc{"_oldversions.md"})
	includes = append(includes, "oldversions")
	for _, d := range definitions {
		// Skip Inlined definitions
		if d.IsInlined {
			continue
		}
		manifest.Docs = append(manifest.Docs, Doc{"_" + GetConceptImport(d) + ".md"})
		includes = append(includes, GetConceptImport(d))
	}

	// Write out the json manifest
	jsonbytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fmt.Printf("Could not Marshal manfiest %+v due to error: %v.\n", manifest, err)
	} else {
		jsonfile, err := os.Create(*api.ConfigDir + "/" + JsonOutputFile)
		if err != nil {
			fmt.Printf("Could not create file %s due to error: %v.\n", JsonOutputFile, err)
		} else {
			defer jsonfile.Close()
			_, err := jsonfile.Write(jsonbytes)
			if err != nil {
				fmt.Printf("Failed to write bytes %s to file %s: %v.\n", jsonbytes, JsonOutputFile, err)
			}
		}
	}
}

const DefaultHeader = `
# <strong>{{.}}</strong>

------------

`

func WriteConceptFiles(config *api.Config) {
	// Setup the template to be instantiated
	t, err := template.New("concept.template").Parse(ConceptTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v", err)
		os.Exit(1)
	}

	// Write concepts for old versions
	for _, d := range config.Definitions.GetAllDefinitions() {
		if !d.IsOldVersion {
			continue
		}
		r := &api.Resource{Definition: d, Name: d.Name}
		WriteTemplate(t, r, GetConceptFilePath(d))
	}
	// Write concepts for items in the Toc
	for _, rc := range config.ResourceCategories {
		for _, r := range rc.Resources {
			WriteTemplate(t, r, GetConceptFilePath(r.Definition))
		}
	}
}

func WriteDefinitionFiles(config *api.Config) {
	// Setup the template to be instantiated
	t, err := template.New("definition.template").Parse(DefinitionTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v", err)
		os.Exit(1)
	}

	for _, definition := range config.Definitions.GetAllDefinitions() {
		// Skip things already present in concept docs
		if definition.InToc || definition.IsInlined || definition.IsOldVersion {
			continue
		}
		WriteTemplate(t, definition, GetDefinitionFilePath(definition))
	}
}

func WriteTemplate(t *template.Template, data interface{}, path string) {
	conceptFile, err := os.Create(path)
	defer conceptFile.Close()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
	err = t.Execute(conceptFile, data)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
}

func getLink(s string) string {
	return "#" + strings.ToLower(strings.Replace(s, " ", "-", -1))
}

func getImport(s string) string {
	return "generated_" + strings.ToLower(strings.Replace(s, ".", "_", 50))
}

func toFileName(s string) string {
	return fmt.Sprintf("%s/includes/_%s.md", *api.ConfigDir, s)
}

func GetDefinitionImport(d *api.Definition) string {
	return fmt.Sprintf("%s_%s_%s_definition", getImport(d.Name), d.Version, d.Group)
}

func GetDefinitionFilePath(d *api.Definition) string {
	return toFileName(GetDefinitionImport(d))
}

// GetConceptImport returns the name to import in the index.html.md file
func GetConceptImport(d *api.Definition) string {
	return fmt.Sprintf("%s_%s_%s_concept", getImport(d.Name), d.Version, d.Group)
}

// GetConceptFilePath returns the filepath to write when instantiating a concept template
func GetConceptFilePath(d *api.Definition) string {
	return toFileName(GetConceptImport(d))
}

type Manifest struct {
	ExampleTabs     []ExampleTab    `json:"example_tabs,omitempty"`
	TableOfContents TableOfContents `json:"table_of_contents,omitempty"`
	Docs            []Doc           `json:"docs,omitempty"`
	Title           string          `json:"title,omitempty"`
	Copyright       string          `json:"copyright,omitempty"`
}

type TableOfContents struct {
	Items []TableOfContentsItem `json:"body_md_files,omitempty"`
}

type TableOfContentsItem struct {
	DisplayName string                `json:"display_name,omitempty"`
	Type        string                `json:"type,omitempty"`
	Link        string                `json:"link,omitempty"`
	Items       []TableOfContentsItem `json:"items,omitempty"`
}

type Doc struct {
	Filename string `json:"filename,omitempty"`
}

type ExampleTab struct {
	DisplayName string `json:"display_name,omitempty"`
	SyntaxType  string `json:"syntax_type,omitempty"`
	HoverText   string `json:"hover_text,omitempty"`
}
