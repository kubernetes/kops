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

type KubectlSpec struct {
	TopLevelCommandGroups []TopLevelCommands `yaml:",omitempty"`
}

type TopLevelCommands struct {
	Group    string            `yaml:",omitempty"`
	Commands []TopLevelCommand `yaml:",omitempty"`
}
type TopLevelCommand struct {
	MainCommand *Command `yaml:",omitempty"`
	SubCommands Commands `yaml:",omitempty"`
}

type Options []*Option
type Option struct {
	Name         string `yaml:",omitempty"`
	Shorthand    string `yaml:",omitempty"`
	DefaultValue string `yaml:"default_value,omitempty"`
	Usage        string `yaml:",omitempty"`
}

type Commands []*Command
type Command struct {
	Name             string   `yaml:",omitempty"`
	Path             string   `yaml:",omitempty"`
	Synopsis         string   `yaml:",omitempty"`
	Description      string   `yaml:",omitempty"`
	Options          Options  `yaml:",omitempty"`
	InheritedOptions Options  `yaml:"inherited_options,omitempty"`
	Example          string   `yaml:",omitempty"`
	SeeAlso          []string `yaml:"see_also,omitempty"`
	Usage            string   `yaml:",omitempty"`
}

type Manifest struct {
	Docs     []Doc    `json:"docs,omitempty"`
	Title     string `json:"title,omitempty"`
	Copyright string `json:"copyright,omitempty"`
}

type Doc struct {
	Filename string `json:"filename,omitempty"`
}

type ToC struct {
	Categories []Category `yaml:",omitempty"`
}

type Category struct {
	Name     string `yaml:",omitempty"`
	Commands []string `yaml:",omitempty"`
	Include string `yaml:",omitempty"`
}

