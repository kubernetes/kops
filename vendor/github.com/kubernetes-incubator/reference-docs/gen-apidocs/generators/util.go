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
	"fmt"
	"github.com/kubernetes-incubator/reference-docs/gen-apidocs/generators/api"
	"strings"
)

const JsonOutputFile = "manifest.json"

func PrintInfo(config *api.Config) {
	definitions := config.Definitions

	hasOrphaned := false
	for name, d := range definitions.GetAllDefinitions() {
		if !d.FoundInField && !d.FoundInOperation {
			if !strings.Contains(name, "meta.v1.APIVersions") && !strings.Contains(name, "meta.v1.Patch") {
				hasOrphaned = true
			}
		}
	}
	if hasOrphaned {
		fmt.Printf("----------------------------------\n")
		fmt.Printf("Orphaned Definitions:\n")
		for name, d := range definitions.GetAllDefinitions() {
			if !d.FoundInField && !d.FoundInOperation {
				if !strings.Contains(name, "meta.v1.APIVersions") && !strings.Contains(name, "meta.v1.Patch") {
					fmt.Printf("[%s]\n", name)
				}
			}
		}
		if !*api.AllowErrors {
			panic("Orphaned definitions found.")
		}
	}

	missingFromToc := false
	for _, d := range definitions.GetAllDefinitions() {
		if !d.InToc && len(d.OperationCategories) > 0 && !d.IsOldVersion && !d.IsInlined {
			missingFromToc = true
		}
	}

	if missingFromToc {
		fmt.Printf("----------------------------------\n")
		fmt.Printf("Definitions with Operations Missing from Toc (Excluding old version):\n")
		for name, d := range definitions.GetAllDefinitions() {
			if !d.InToc && len(d.OperationCategories) > 0 && !d.IsOldVersion && !d.IsInlined {
				fmt.Printf("[%s]\n", name)
				for _, oc := range d.OperationCategories {
					for _, o := range oc.Operations {
						fmt.Printf("\t [%s]\n", o.ID)
					}
				}
			}
		}
	}

	//fmt.Printf("Old definitions:\n")
	//for name, d := range definitions.GetAllDefinitions() {
	//	if !d.InToc && len(d.OperationCategories) > 0 && d.IsOldVersion && !d.IsInlined {
	//		fmt.Printf("[%s]\n", name)
	//		for _, oc := range d.OperationCategories {
	//			for _, o := range oc.Operations {
	//				fmt.Printf("\t [%s]\n", o.ID)
	//			}
	//		}
	//	}
	//}
}

func PrintDebug(config *api.Config) {
	operations := config.Operations
	definitions := config.Definitions

	fmt.Printf("----------------------------------\n")
	fmt.Printf("Operations with no Defintions:\n")
	for _, o := range operations {
		if o.Definition == nil {
			fmt.Printf("%s\n", o.ID)
		}
	}

	fmt.Printf("----------------------------------\n")
	fmt.Printf("\n\nDefinitions in Toc:\n")
	for name, d := range definitions.GetAllDefinitions() {
		if d.InToc {
			fmt.Printf("\n\n%s \n\tFields:\n", name)
			for _, f := range d.Fields {
				if f.Definition != nil {
					fmt.Printf("\t\t%s:[%s](%s) - (%s)\n", f.Name, f.Type, f.Definition.Name, f.Description)
				} else {
					fmt.Printf("\t\t%s:%s - (%s)\n", f.Name, f.Type, f.Description)
				}
			}
			for _, oc := range d.OperationCategories {
				fmt.Printf("\t Operation Category [%s]\n", oc.Name)
				for _, o := range oc.Operations {
					fmt.Printf("\t\t %s (%s) Type: %s\n", o.Type.Name, o.Path, o.Definition.Name)
					if len(o.PathParams) > 0 {
						fmt.Printf("\t\t\t Path Params\n")
						for _, p := range o.PathParams {
							fmt.Printf("\t\t\t %s:%s - (%s)\n", p.Name, p.Type, p.Description)
						}
					}
					if len(o.QueryParams) > 0 {
						fmt.Printf("\t\t\t Query Params\n")
						for _, p := range o.QueryParams {
							fmt.Printf("\t\t\t %s:%s - (%s)\n", p.Name, p.Type, p.Description)
						}
					}
					fmt.Printf("\t\t\t Responses\n")
					for _, hr := range o.HttpResponses {
						fmt.Printf("\t\t\t %s:%s - (%s)\n", hr.Code, hr.Type, hr.Definition.Name)
					}
				}
			}
		}
	}

	fmt.Printf("----------------------------------\n")
	fmt.Printf("\n\nOther Definitions:\n")
	for name, d := range definitions.GetAllDefinitions() {
		if !d.InToc && d.FoundInField {
			fmt.Printf("\n\n%s \n\tFields:\n", name)
			for _, f := range d.Fields {
				if f.Definition != nil {
					fmt.Printf("\t\t%s:[%s](%s) - (%s)\n", f.Name, f.Type, f.Definition.Name, f.Description)
				} else {
					fmt.Printf("\t\t%s:%s - (%s)\n", f.Name, f.Type, f.Description)
				}
			}
			for _, oc := range d.OperationCategories {
				fmt.Printf("\t Operation Category [%s]\n", oc.Name)
				for _, o := range oc.Operations {
					fmt.Printf("\t\t %s (%s) Type: %s\n", o.Type.Name, o.Path, o.Definition.Name)
				}
			}
		}
	}
}
