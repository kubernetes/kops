/*
Copyright 2019 The Kubernetes Authors.

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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/yaml"

	helmvalues "helm.sh/helm/v3/pkg/cli/values"

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/try"
	"k8s.io/kops/pkg/util/templater"
	"k8s.io/kops/upup/pkg/fi/utils"
)

var (
	toolboxTemplatingLong = templates.LongDesc(i18n.T(`
	Generate cluster.yaml from values input yaml file and apply template.
	`))

	toolboxTemplatingExample = templates.Examples(i18n.T(`
	# Generate cluster.yaml from template and input values
	kops toolbox template \
		--values values.yaml --values=another.yaml \
		--set var=value --set-string othervar=true \
		--snippets file_or_directory --snippets=another.dir \
		--template file_or_directory --template=directory  \
		--output cluster.yaml
	`))

	toolboxTemplatingShort = i18n.T(`Generate cluster.yaml from template`)
)

type ToolboxTemplateOptions struct {
	ClusterName   string
	configPath    []string
	configValue   string
	failOnMissing bool
	formatYAML    bool
	outputPath    string
	snippetsPath  []string
	templatePath  []string
	values        []string
	stringValues  []string
	channel       string
}

// NewCmdToolboxTemplate returns a new templating command.
func NewCmdToolboxTemplate(f commandutils.Factory, out io.Writer) *cobra.Command {
	options := &ToolboxTemplateOptions{
		channel: kopsapi.DefaultChannel,
	}

	cmd := &cobra.Command{
		Use:               "template [CLUSTER]",
		Short:             toolboxTemplatingShort,
		Long:              toolboxTemplatingLong,
		Example:           toolboxTemplatingExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunToolBoxTemplate(f, out, options)
		},
	}

	cmd.Flags().StringSliceVar(&options.configPath, "values", options.configPath, "Path to a configuration file containing values to include in template")
	cmd.RegisterFlagCompletionFunc("values", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"yaml", "json"}, cobra.ShellCompDirectiveFilterFileExt
	})
	cmd.Flags().StringArrayVar(&options.values, "set", options.values, "Set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.RegisterFlagCompletionFunc("set", cobra.NoFileCompletions)
	cmd.Flags().StringArrayVar(&options.stringValues, "set-string", options.stringValues, "Set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.RegisterFlagCompletionFunc("set", cobra.NoFileCompletions)
	cmd.Flags().StringSliceVar(&options.templatePath, "template", options.templatePath, "Path to template file or directory of templates to render")
	cmd.Flags().StringSliceVar(&options.snippetsPath, "snippets", options.snippetsPath, "Path to directory containing snippets used for templating")
	cmd.MarkFlagDirname("snippets")
	cmd.Flags().StringVar(&options.channel, "channel", options.channel, "Channel to use for the channel* functions")
	cmd.RegisterFlagCompletionFunc("channel", completeChannel)
	cmd.Flags().StringVar(&options.outputPath, "out", options.outputPath, "Path to output file. Defaults to stdout")
	cmd.Flags().StringVar(&options.configValue, "config-value", "", "Show the value of a specific configuration value")
	cmd.RegisterFlagCompletionFunc("config-value", cobra.NoFileCompletions)
	cmd.Flags().BoolVar(&options.failOnMissing, "fail-on-missing", true, "Fail on referencing unset variables in templates")
	cmd.Flags().BoolVar(&options.formatYAML, "format-yaml", false, "Attempt to format the generated yaml content before output")
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		// Old flag name for outputPath
		if name == "output" {
			name = "out"
		}
		return pflag.NormalizedName(name)
	})

	return cmd
}

// RunToolBoxTemplate is the action for the command
func RunToolBoxTemplate(f commandutils.Factory, out io.Writer, options *ToolboxTemplateOptions) error {
	// @step: read in the configuration if any
	context, err := newTemplateContext(options.configPath, options.values, options.stringValues)
	if err != nil {
		return err
	}

	// @step: set clusterName from template's values or cli flag
	value, ok := context["clusterName"].(string)
	if ok {
		options.ClusterName = value
	} else {
		context["clusterName"] = options.ClusterName
	}

	// @check if we are just rendering the config value
	if options.configValue != "" {
		v, found := context[options.configValue]
		switch found {
		case true:
			fmt.Fprintf(out, "%v\n", v)
		default:
			fmt.Fprintf(out, "null\n")
		}
		return nil
	}

	// @step: expand the list of templates into a list of files to render
	var templates []string
	for _, x := range options.templatePath {
		list, err := expandFiles(utils.ExpandPath(x))
		if err != nil {
			return fmt.Errorf("unable to expand the template: %s, error: %s", x, err)
		}
		templates = append(templates, list...)
	}

	snippets := make(map[string]string)
	for _, x := range options.snippetsPath {
		list, err := expandFiles(utils.ExpandPath(x))
		if err != nil {
			return fmt.Errorf("unable to expand the snippets: %s, error: %s", x, err)
		}

		for _, j := range list {
			content, err := os.ReadFile(j)
			if err != nil {
				return fmt.Errorf("unable to read snippet: %s, error: %s", j, err)
			}
			snippets[path.Base(j)] = string(content)
		}
	}

	channel, err := kopsapi.LoadChannel(options.channel)
	if err != nil {
		return fmt.Errorf("error loading channel %q: %v", options.channel, err)
	}

	// @step: render each of the templates, splitting on the documents
	r := templater.NewTemplater(channel)
	var documents []string
	for _, x := range templates {
		content, err := os.ReadFile(x)
		if err != nil {
			return fmt.Errorf("unable to read template: %s, error: %s", x, err)
		}

		rendered, err := r.Render(string(content), context, snippets, options.failOnMissing)
		if err != nil {
			return fmt.Errorf("unable to render template: %s, error: %s", x, err)
		}
		// @check if the content is zero ignore it
		if len(rendered) <= 0 {
			continue
		}

		if !options.formatYAML {
			documents = append(documents, strings.Split(rendered, "---\n")...)
			continue
		}

		for _, x := range strings.Split(rendered, "---\n") {
			var data map[string]interface{}
			if err := yaml.Unmarshal([]byte(x), &data); err != nil {
				return fmt.Errorf("unable to unmarshall content from template: %s, error: %s", x, err)
			}
			if len(data) <= 0 {
				continue
			}
			formatted, err := yaml.Marshal(&data)
			if err != nil {
				return fmt.Errorf("unable to marhshal formatted content to yaml: %s", err)
			}
			documents = append(documents, string(formatted))
		}
	}
	// join in harmony all the YAML documents back together
	content := strings.Join(documents, "---\n")

	iowriter := out
	// @check if we are writing to a file rather than stdout
	if options.outputPath != "" {
		w, err := os.OpenFile(utils.ExpandPath(options.outputPath), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0o660)
		if err != nil {
			return fmt.Errorf("unable to open file: %s, error: %v", options.outputPath, err)
		}
		defer try.CloseFile(w)
		iowriter = w
	}

	if _, err := iowriter.Write([]byte(content)); err != nil {
		return fmt.Errorf("unable to write template: %s", err)
	}

	return nil
}

// newTemplateContext is responsible for loading the --values and build a context for the template
func newTemplateContext(files []string, values []string, stringValues []string) (map[string]interface{}, error) {
	context := make(map[string]interface{})

	for _, x := range files {
		list, err := expandFiles(utils.ExpandPath(x))
		if err != nil {
			return nil, err
		}
		for _, j := range list {
			content, err := os.ReadFile(j)
			if err != nil {
				return nil, fmt.Errorf("unable to configuration file: %s, error: %s", j, err)
			}

			ctx := make(map[string]interface{})
			if err := utils.YamlUnmarshal(content, &ctx); err != nil {
				return nil, fmt.Errorf("unable decode the configuration file: %s, error: %v", j, err)
			}

			valueOpts := &helmvalues.Options{
				Values:       values,
				ValueFiles:   files,
				StringValues: stringValues,
			}

			context, err = valueOpts.MergeValues(nil)
			if err != nil {
				return nil, err
			}
		}
	}

	// User specified a value via --set
	for _, value := range values {
		if err := strvals.ParseInto(value, context); err != nil {
			return nil, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	// User specified a value via --set-string
	for _, value := range stringValues {
		if err := strvals.ParseIntoString(value, context); err != nil {
			return nil, fmt.Errorf("failed parsing --set-string data: %s", err)
		}
	}

	return context, nil
}

// expandFiles is responsible for resolving any references to directories
func expandFiles(path string) ([]string, error) {
	// @check if the path is a directory, if not we can return straight away
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	// @check if not a directory and return as is
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
