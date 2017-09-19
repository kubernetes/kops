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

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/util/templater"
	"k8s.io/kops/upup/pkg/fi/utils"

	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	toolboxTemplatingLong = templates.LongDesc(i18n.T(`
	Generate cluster.yaml from values input yaml file and apply template.
	`))

	toolboxTemplatingExample = templates.Examples(i18n.T(`
	# generate cluster.yaml from template and input values

	kops toolbox template \
		--values values.yaml --values=another.yaml \
		--snippets file_or_directory --snippets=another.dir \
		--template file_or_directory --template=directory  \
		--output cluster.yaml
	`))

	toolboxTemplatingShort = i18n.T(`Generate cluster.yaml from template`)
)

// the options for the command
type toolboxTemplateOption struct {
	clusterName  string
	configPath   []string
	outputPath   string
	snippetsPath []string
	templatePath []string
}

// NewCmdToolboxTemplate returns a new templating command
func NewCmdToolboxTemplate(f *util.Factory, out io.Writer) *cobra.Command {
	options := &toolboxTemplateOption{}

	cmd := &cobra.Command{
		Use:     "template",
		Short:   toolboxTemplatingShort,
		Long:    toolboxTemplatingLong,
		Example: toolboxTemplatingExample,
		Run: func(cmd *cobra.Command, args []string) {
			if err := rootCommand.ProcessArgs(args); err != nil {
				exitWithError(err)
			}
			options.clusterName = rootCommand.ClusterName()

			if err := runToolBoxTemplate(f, out, options); err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringSliceVar(&options.configPath, "values", options.configPath, "Path to a configuration file containing values to include in template")
	cmd.Flags().StringSliceVar(&options.templatePath, "template", options.templatePath, "Path to template file or directory of templates to render")
	cmd.Flags().StringSliceVar(&options.snippetsPath, "snippets", options.snippetsPath, "Path to directory containing snippets used for templating")
	cmd.Flags().StringVar(&options.outputPath, "output", options.outputPath, "Path to output file, otherwise defaults to stdout")

	return cmd
}

// runToolBoxTemplate is the action for the command
func runToolBoxTemplate(f *util.Factory, out io.Writer, options *toolboxTemplateOption) error {
	// @step: read in the configuration if any
	context := make(map[string]interface{}, 0)
	for _, x := range options.configPath {
		list, err := expandFiles(utils.ExpandPath(x))
		if err != nil {
			return err
		}
		for _, j := range list {
			content, err := ioutil.ReadFile(j)
			if err != nil {
				return fmt.Errorf("unable to configuration file: %s, error: %s", j, err)
			}

			ctx := make(map[string]interface{}, 0)
			if err := utils.YamlUnmarshal(content, &ctx); err != nil {
				return fmt.Errorf("unable decode the configuration file: %s, error: %v", j, err)
			}
			// @step: merge the maps together
			for k, v := range ctx {
				context[k] = v
			}
		}
	}
	context["clusterName"] = options.clusterName

	// @step: expand the list of templates into a list of files to render
	var templates []string
	for _, x := range options.templatePath {
		list, err := expandFiles(utils.ExpandPath(x))
		if err != nil {
			return fmt.Errorf("unable to expand the template: %s, error: %s", x, err)
		}
		templates = append(templates, list...)
	}

	snippets := make(map[string]string, 0)
	for _, x := range options.snippetsPath {
		list, err := expandFiles(utils.ExpandPath(x))
		if err != nil {
			return fmt.Errorf("unable to expand the snippets: %s, error: %s", x, err)
		}

		for _, j := range list {
			content, err := ioutil.ReadFile(j)
			if err != nil {
				return fmt.Errorf("unable to read snippet: %s, error: %s", j, err)
			}
			snippets[j] = string(content)
		}
	}

	// @step: get the output io.Writer
	writer := out
	if options.outputPath != "" {
		w, err := os.OpenFile(utils.ExpandPath(options.outputPath), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0660)
		if err != nil {
			return fmt.Errorf("unable to open file: %s, error: %v", options.outputPath, err)
		}
		writer = w
	}

	// @step: render each of the template and write to location
	r := templater.NewTemplater()
	size := len(templates) - 1
	for i, x := range templates {
		content, err := ioutil.ReadFile(x)
		if err != nil {
			return fmt.Errorf("unable to read template: %s, error: %s", x, err)
		}

		rendered, err := r.Render(string(content), context, snippets)
		if err != nil {
			return fmt.Errorf("unable to render template: %s, error: %s", x, err)
		}
		io.WriteString(writer, rendered)

		// @check if we should need to add document separator
		if i < size {
			io.WriteString(writer, "---\n")
		}
	}

	return nil
}

// expandFiles is responsible for resolving any references to directories
func expandFiles(path string) ([]string, error) {
	// @check if the the path is a directory, if not we can return straight away
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	// @check if no a directory and return as is
	if !stat.IsDir() {
		return []string{path}, nil
	}
	// @step: iterate the directory and get all the files
	var list []string
	if err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		list = append(list, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return list, nil
}
