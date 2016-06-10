package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/template"
)

const maxIterations = 10

type OptionsTemplate struct {
	Name     string
	Tags     []string
	Template *template.Template
}

type OptionsLoader struct {
	config    interface{}
	templates OptionsTemplateList

	TemplateFunctions template.FuncMap
}

type OptionsTemplateList []*OptionsTemplate

func (a OptionsTemplateList) Len() int {
	return len(a)
}
func (a OptionsTemplateList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a OptionsTemplateList) Less(i, j int) bool {
	l := a[i]
	r := a[j]

	// First ordering criteria: Execute things with fewer tags first (more generic)
	if len(l.Tags) != len(r.Tags) {
		return len(l.Tags) < len(r.Tags)
	}

	// TODO: lexicographic sort on tags, for full determinism?

	// Final ordering criteria: order by name
	return l.Name < r.Name
}

func NewOptionsLoader(config interface{}) *OptionsLoader {
	l := &OptionsLoader{}
	l.config = config
	l.TemplateFunctions = make(template.FuncMap)
	return l
}

func (l *OptionsLoader) AddTemplate(t *OptionsTemplate) {
	l.templates = append(l.templates, t)
}

// copyFromStruct merges src into dest
// It uses a JSON marshal & unmarshal, so only fields that are JSON-visible will be copied
func copyFromStruct(dest, src interface{}) {
	// Not the most efficient approach, but simple & relatively well defined
	j, err := json.Marshal(src)
	if err != nil {
		glog.Fatalf("error marshalling config: %v", err)
	}
	err = json.Unmarshal(j, dest)
	if err != nil {
		glog.Fatalf("error unmarshalling config: %v", err)
	}
}

// iterate performs a single iteration of all the templates, executing each template in order
func (l *OptionsLoader) iterate(inConfig interface{}) (interface{}, error) {
	sort.Sort(l.templates)

	t := reflect.TypeOf(inConfig).Elem()

	options := reflect.New(t).Interface()

	// Copy the provided values before applying rules; they act as defaults (and overrides below)
	copyFromStruct(options, inConfig)

	for _, t := range l.templates {
		glog.V(2).Infof("executing template %s (tags=%s)", t.Name, t.Tags)

		var buffer bytes.Buffer
		err := t.Template.ExecuteTemplate(&buffer, t.Name, inConfig)
		if err != nil {
			return nil, fmt.Errorf("error executing template %q: %v", t.Name, err)
		}

		yamlBytes := buffer.Bytes()

		jsonBytes, err := utils.YamlToJson(yamlBytes)
		if err != nil {
			// TODO: It would be nice if yaml returned us the line number here
			glog.Infof("error parsing yaml.  yaml follows:")
			for i, line := range strings.Split(string(yamlBytes), "\n") {
				fmt.Fprintf(os.Stderr, "%3d: %s\n", i, line)
			}
			return nil, fmt.Errorf("error parsing yaml %q: %v", t.Name, err)
		}

		err = json.Unmarshal(jsonBytes, options)
		if err != nil {
			return nil, fmt.Errorf("error parsing yaml (converted to JSON) %q: %v", t.Name, err)
		}
	}

	// Also copy the provided values after applying rules; they act as overrides now
	copyFromStruct(options, inConfig)

	return options, nil
}

// Build executes the options configuration templates, until they converge
// It bails out after maxIterations
func (l *OptionsLoader) Build() (interface{}, error) {
	options := l.config
	iteration := 0
	for {
		nextOptions, err := l.iterate(options)
		if err != nil {
			return nil, err
		}

		if reflect.DeepEqual(options, nextOptions) {
			return options, nil
		}

		iteration++
		if iteration > maxIterations {
			return nil, fmt.Errorf("options did not converge after %d iterations", maxIterations)
		}

		options = nextOptions
	}
}

// HandleOptions is the file handler for options files
// It builds a template with the file, and adds it to the list of options templates
func (l *OptionsLoader) HandleOptions(i *TreeWalkItem) error {
	contents, err := i.ReadString()
	if err != nil {
		return err
	}

	t := template.New(i.RelativePath)
	t.Funcs(l.TemplateFunctions)

	_, err = t.Parse(contents)
	if err != nil {
		return fmt.Errorf("error parsing options template %q: %v", i.Path, err)
	}

	t.Option("missingkey=zero")

	l.AddTemplate(&OptionsTemplate{
		Name:     i.RelativePath,
		Tags:     i.Tags,
		Template: t,
	})
	return nil
}
