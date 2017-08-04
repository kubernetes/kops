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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	yaml "gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	toolbox_templating_long = templates.LongDesc(i18n.T(`
	Generate cluster.yaml from values input yaml file and apply template.
	`))

	toolbox_templating_example = templates.Examples(i18n.T(`
	# generate cluster.yaml from template and input values

	kops toolbox template \
		--values values.yaml \
		--template cluster.tmpl.yaml \
		--output cluster.yaml
	`))

	toolbox_templating_short = i18n.T(`Generate cluster.yaml from template`)
)

type ToolboxTemplateOption struct {
	ClusterName  string
	ValuesFile   string
	TemplateFile string
	OutPutFile   string
}

func NewCmdToolboxTemplate(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxTemplateOption{}

	cmd := &cobra.Command{
		Use:     "template",
		Short:   toolbox_templating_short,
		Long:    toolbox_templating_long,
		Example: toolbox_templating_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := rootCommand.ProcessArgs(args)
			if err != nil {
				exitWithError(err)
			}

			options.ClusterName = rootCommand.ClusterName()

			err = RunToolBoxTemplate(f, out, options)
			if err != nil {
				exitWithError(err)
			}
		},
	}

	cmd.Flags().StringVar(&options.ValuesFile, "values", options.ValuesFile, "Path to values yaml file, default: values.yaml")
	cmd.Flags().StringVar(&options.TemplateFile, "template", options.TemplateFile, "Path to template file, default: cluster.tmpl.yaml")
	cmd.Flags().StringVar(&options.OutPutFile, "output", options.OutPutFile, "Path to output file, default: cluster.yaml")

	return cmd
}

func RunToolBoxTemplate(f *util.Factory, out io.Writer, options *ToolboxTemplateOption) error {
	if options.ValuesFile == "" {
		options.ValuesFile = "values.yaml"
	}
	if options.TemplateFile == "" {
		options.TemplateFile = "cluster.tmpl.yaml"
	}
	if options.OutPutFile == "" {
		options.OutPutFile = "cluster.yaml"
	}

	options.ValuesFile = utils.ExpandPath(options.ValuesFile)
	options.TemplateFile = utils.ExpandPath(options.TemplateFile)
	options.OutPutFile = utils.ExpandPath(options.OutPutFile)

	err := ExecTemplate(options, out)
	if err != nil {
		exitWithError(err)
	}

	return nil
}

func ExecTemplate(options *ToolboxTemplateOption, out io.Writer) error {
	valuesByteArr, err := ioutil.ReadFile(options.ValuesFile)
	if err != nil {
		return fmt.Errorf("failed to read values file: %v :%v", options.ValuesFile, err)
	}

	tmpl, err := template.ParseFiles(options.TemplateFile)
	if err != nil {
		return fmt.Errorf("failed to read template file: %v :%v", options.TemplateFile, err)
	}

	var values map[string]interface{}
	err = yaml.Unmarshal(valuesByteArr, &values)
	if err != nil {
		return fmt.Errorf("failed to unmarshal valuesfile: %v :%v", options.ValuesFile, err)
	}

	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)

	err = tmpl.Execute(writer, values)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	err = writer.Flush()
	if err != nil {
		exitWithError(err)
	}

	err = ioutil.WriteFile(options.OutPutFile, buff.Bytes(), 0644)
	if err != nil {
		exitWithError(err)
	}
	return nil
}
