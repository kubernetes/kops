package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"os"
	"reflect"
	"strings"
	"text/template"
)

const maxIterations = 10

type OptionsLoader struct {
	config    interface{}
	templates []*template.Template
}

func NewOptionsLoader(config interface{}) *OptionsLoader {
	l := &OptionsLoader{}
	l.config = config
	return l
}

func (l *OptionsLoader) AddTemplate(t *template.Template) {
	l.templates = append(l.templates, t)
}
func copyStruct(dest, src interface{}) {
	vDest := reflect.ValueOf(dest).Elem()
	vSrc := reflect.ValueOf(src).Elem()

	for i := 0; i < vSrc.NumField(); i++ {
		fv := vSrc.Field(i)
		vDest.Field(i).Set(fv)
	}
}

func (l *OptionsLoader) iterate(inConfig interface{}) (interface{}, error) {
	t := reflect.TypeOf(inConfig).Elem()

	options := reflect.New(t).Interface()
	copyStruct(options, inConfig)
	for _, t := range l.templates {
		glog.V(2).Infof("executing template %s", t.Name())

		var buffer bytes.Buffer
		err := t.ExecuteTemplate(&buffer, t.Name(), inConfig)
		if err != nil {
			return nil, fmt.Errorf("error executing template %q: %v", t.Name(), err)
		}

		yamlBytes := buffer.Bytes()

		jsonBytes, err := utils.YamlToJson(yamlBytes)
		if err != nil {
			// TODO: It would be nice if yaml returned us the line number here
			glog.Infof("error parsing yaml.  yaml follows:")
			for i, line := range strings.Split(string(yamlBytes), "\n") {
				fmt.Fprintf(os.Stderr, "%3d: %s\n", i, line)
			}
			return nil, fmt.Errorf("error parsing yaml %q: %v", t.Name(), err)
		}

		err = json.Unmarshal(jsonBytes, options)
		if err != nil {
			return nil, fmt.Errorf("error parsing yaml (converted to JSON) %q: %v", t.Name(), err)
		}
	}

	return options, nil
}

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

func (l *OptionsLoader) HandleOptions(i *TreeWalkItem) error {
	contents, err := i.ReadString()
	if err != nil {
		return err
	}

	t := template.New(i.RelativePath)
	_, err = t.Parse(contents)
	if err != nil {
		return fmt.Errorf("error parsing options template %q: %v", i.Path, err)
	}

	t.Option("missingkey=zero")

	l.AddTemplate(t)
	return nil
}
