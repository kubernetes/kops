package cloudup

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/golang/glog"
	"io"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/fitasks"
	"k8s.io/kube-deploy/upup/pkg/fi/loader"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"
)

type deferredType int

const (
	KEY_NAME = "name"
	KEY_TYPE = "_type"
)

const (
	deferredUnit deferredType = iota
	deferredResource
)

type Loader struct {
	StateDir      string
	OptionsLoader *loader.OptionsLoader
	NodeModelDir  string

	Tags              map[string]struct{}
	TemplateFunctions template.FuncMap

	typeMap map[string]reflect.Type

	templates []*template.Template
	config    interface{}

	Resources map[string]fi.Resource
	deferred  []*deferredBinding

	tasks map[string]fi.Task

	unmarshaller utils.Unmarshaller
}

type templateResource struct {
	key      string
	loader   *Loader
	template string
	args     []string
}

var _ fi.Resource = &templateResource{}
var _ fi.TemplateResource = &templateResource{}

func (a *templateResource) Open() (io.ReadSeeker, error) {
	var err error
	result, err := a.loader.executeTemplate(a.key, a.template, a.args)
	if err != nil {
		return nil, fmt.Errorf("error executing resource template %q: %v", a.key, err)
	}
	reader := bytes.NewReader([]byte(result))
	return reader, nil
}

func (a *templateResource) Curry(args []string) fi.TemplateResource {
	curried := &templateResource{}
	*curried = *a
	curried.args = append(curried.args, args...)
	return curried
}

type deferredBinding struct {
	name         string
	dest         utils.Settable
	src          string
	deferredType deferredType
}

func (l *Loader) Init() {
	l.tasks = make(map[string]fi.Task)
	l.typeMap = make(map[string]reflect.Type)
	l.unmarshaller.SpecialCases = l.unmarshalSpecialCases
	l.Resources = make(map[string]fi.Resource)
	l.TemplateFunctions = make(template.FuncMap)
}

func (l *Loader) AddTypes(types map[string]interface{}) {
	for key, proto := range types {
		_, exists := l.typeMap[key]
		if exists {
			glog.Fatalf("duplicate type key: %q", key)
		}

		t := reflect.TypeOf(proto)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		l.typeMap[key] = t
	}
}

func (l *Loader) executeTemplate(key string, d string, args []string) (string, error) {
	t := template.New(key)

	funcMap := make(template.FuncMap)
	funcMap["Base64Encode"] = func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	funcMap["Args"] = func() []string {
		return args
	}
	funcMap["BuildNodeConfig"] = func(target string, configResourceName string, args []string) (string, error) {
		return l.buildNodeConfig(target, configResourceName, args)
	}
	funcMap["RenderResource"] = func(resourceName string, args []string) (string, error) {
		return l.renderResource(resourceName, args)
	}
	for k, fn := range l.TemplateFunctions {
		funcMap[k] = fn
	}
	t.Funcs(funcMap)

	context := l.config

	_, err := t.Parse(d)
	if err != nil {
		return "", fmt.Errorf("error parsing template %q: %v", key, err)
	}

	t.Option("missingkey=zero")

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, key, context)
	if err != nil {
		return "", fmt.Errorf("error executing template %q: %v", key, err)
	}

	return buffer.String(), nil
}

func ignoreHandler(i *loader.TreeWalkItem) error {
	return nil
}

func (l *Loader) Build(baseDir string) (map[string]fi.Task, error) {
	// First pass: load options
	tw := &loader.TreeWalker{
		DefaultHandler: ignoreHandler,
		Contexts: map[string]loader.Handler{
			"resources": ignoreHandler,
			"pki":       ignoreHandler,
		},
		Extensions: map[string]loader.Handler{
			".options": l.OptionsLoader.HandleOptions,
		},
		Tags: l.Tags,
	}
	err := tw.Walk(baseDir)
	if err != nil {
		return nil, err
	}

	l.config, err = l.OptionsLoader.Build()
	if err != nil {
		return nil, err
	}
	glog.Infof("options: %s", fi.DebugAsJsonStringIndent(l.config))

	// Second pass: load everything else
	tw = &loader.TreeWalker{
		DefaultHandler: l.objectHandler,
		Contexts: map[string]loader.Handler{
			"resources": l.resourceHandler,
			"pki":       l.pkiHandler,
		},
		Extensions: map[string]loader.Handler{
			".options": ignoreHandler,
		},
		Tags: l.Tags,
	}

	err = tw.Walk(baseDir)
	if err != nil {
		return nil, err
	}

	err = l.processDeferrals()
	if err != nil {
		return nil, err
	}
	return l.tasks, nil
}

func (l *Loader) processDeferrals() error {
	if len(l.deferred) != 0 {
		unitMap := make(map[string]fi.Task)

		for k, o := range l.tasks {
			if unit, ok := o.(fi.Task); ok {
				unitMap[k] = unit
			}
		}

		for _, d := range l.deferred {
			src := d.src

			switch d.deferredType {
			case deferredUnit:
				unit, found := unitMap[src]
				if !found {
					glog.Infof("Known targets:")
					for k := range unitMap {
						glog.Infof("  %s", k)
					}
					return fmt.Errorf("cannot resolve link at %q to %q", d.name, d.src)
				}

				d.dest.Set(reflect.ValueOf(unit))

			case deferredResource:
				// Resources can contain template 'arguments', separated by spaces
				// <resourcename> <arg1> <arg2>
				tokens := strings.Split(src, " ")
				match := tokens[0]
				args := tokens[1:]

				match = strings.TrimPrefix(match, "resources/")
				found := l.Resources[match]

				if found == nil {
					glog.Infof("Known resources:")
					for k := range l.Resources {
						glog.Infof("  %s", k)
					}
					return fmt.Errorf("cannot resolve resource link %q (at %q)", d.src, d.name)
				}

				err := l.populateResource(d.name, d.dest, found, args)
				if err != nil {
					return fmt.Errorf("error setting resource value: %v", err)
				}

			default:
				panic("unhandled deferred type")
			}
		}
	}

	return nil
}

func (l *Loader) resourceHandler(i *loader.TreeWalkItem) error {
	contents, err := i.ReadBytes()
	if err != nil {
		return err
	}

	var a fi.Resource
	key := i.RelativePath
	if strings.HasSuffix(key, ".template") {
		key = strings.TrimSuffix(key, ".template")
		glog.V(2).Infof("loading (templated) resource %q", key)

		a = &templateResource{
			template: string(contents),
			loader:   l,
			key:      key,
		}
	} else {
		glog.V(2).Infof("loading resource %q", key)
		a = fi.NewBytesResource(contents)

	}

	l.Resources[key] = a
	return nil
}

func (l *Loader) pkiHandler(i *loader.TreeWalkItem) error {
	contents, err := i.ReadString()
	if err != nil {
		return err
	}

	key := i.RelativePath

	contents, err = l.executeTemplate(key, contents, nil)
	if err != nil {
		return err
	}

	task, err := fitasks.NewPKIKeyPairTask(key, contents, "")
	if err != nil {
		return err
	}
	l.tasks["pki/"+i.RelativePath] = task
	return nil
}

func (l *Loader) objectHandler(i *loader.TreeWalkItem) error {
	contents, err := i.ReadString()
	if err != nil {
		return err
	}

	data, err := l.executeTemplate(i.RelativePath, contents, nil)
	if err != nil {
		return err
	}

	objects, err := l.loadYamlObjects(i.RelativePath, data)
	if err != nil {
		return err
	}

	for k, v := range objects {
		_, found := l.tasks[k]
		if found {
			return fmt.Errorf("found duplicate object: %q", k)
		}
		l.tasks[k] = v.(fi.Task)
	}
	return nil
}

func (l *Loader) loadYamlObjects(key string, data string) (map[string]interface{}, error) {
	var o map[string]interface{}
	err := utils.YamlUnmarshal([]byte(data), &o)
	if err != nil {
		// TODO: It would be nice if yaml returned us the line number here
		glog.Infof("error parsing yaml.  yaml follows:")
		for i, line := range strings.Split(string(data), "\n") {
			fmt.Fprintf(os.Stderr, "%3d: %s\n", i, line)
		}
		return nil, fmt.Errorf("error parsing yaml %q: %v", key, err)
	}

	return l.loadObjectMap(key, o)
}

func (l *Loader) loadObjectMap(key string, data map[string]interface{}) (map[string]interface{}, error) {
	loaded := make(map[string]interface{})

	for k, v := range data {
		typeId := ""
		name := ""

		// If the name & type are not specified in the values,
		// we infer them from the key (first component -> typeid, last component -> name)
		if vMap, ok := v.(map[string]interface{}); ok {
			if s, ok := vMap[KEY_TYPE]; ok {
				typeId = s.(string)
			}
			if s, ok := vMap[KEY_NAME]; ok {
				name = s.(string)
			}
		}

		inferredName := false

		if name == "" {
			lastSlash := strings.LastIndex(k, "/")
			name = k[lastSlash+1:]
			inferredName = true
		}

		if typeId == "" {
			firstSlash := strings.Index(k, "/")
			if firstSlash != -1 {
				typeId = k[:firstSlash]
			}

			if typeId == "" {
				return nil, fmt.Errorf("cannot determine type for %q", k)
			}
		}

		t, found := l.typeMap[typeId]
		if !found {
			return nil, fmt.Errorf("unknown type %q (in %q)", typeId, key)
		}

		o := reflect.New(t)
		err := l.unmarshaller.UnmarshalStruct(key+":"+k, o, v)
		if err != nil {
			return nil, err
		}
		//glog.Infof("Built %s:%s => %v", key, k, o.Interface())

		if inferredName {
			nameField := o.Elem().FieldByName("Name")
			if nameField.IsValid() {
				err := l.unmarshaller.UnmarshalSettable(k+":Name", utils.Settable{Value: nameField}, name)
				if err != nil {
					return nil, err
				}
			}
		}
		loaded[k] = o.Interface()
	}
	return loaded, nil
}

func (l *Loader) unmarshalSpecialCases(name string, dest utils.Settable, src interface{}) (bool, error) {
	if dest.Type().Kind() == reflect.Slice {
		switch src := src.(type) {
		case []interface{}:
			destValueArray := reflect.MakeSlice(dest.Type(), len(src), len(src))
			for i, srcElem := range src {
				done, err := l.unmarshalSpecialCases(fmt.Sprintf("%s[%d]", name, i),
					utils.Settable{Value: destValueArray.Index(i)},
					srcElem)
				if err != nil {
					return false, err
				}
				if !done {
					return false, nil
				}
			}
			dest.Set(destValueArray)
			return true, nil
		default:
			return false, fmt.Errorf("unhandled conversion for %q: %T -> %s", name, src, dest.Value.Type().Name())
		}
	}

	if dest.Type().Kind() == reflect.Ptr || dest.Type().Kind() == reflect.Interface {
		resourceType := reflect.TypeOf((*fi.Resource)(nil)).Elem()
		if dest.Value.Type().AssignableTo(resourceType) {
			d := &deferredBinding{
				name:         name,
				dest:         dest,
				deferredType: deferredResource,
			}
			switch src := src.(type) {
			case string:
				d.src = src
			default:
				return false, fmt.Errorf("unhandled conversion for %q: %T -> %s", name, src, dest.Value.Type().Name())
			}
			l.deferred = append(l.deferred, d)
			return true, nil
		}

		taskType := reflect.TypeOf((*fi.Task)(nil)).Elem()
		if dest.Value.Type().AssignableTo(taskType) {
			d := &deferredBinding{
				name:         name,
				dest:         dest,
				deferredType: deferredUnit,
			}
			switch src := src.(type) {
			case string:
				d.src = src
			default:
				return false, fmt.Errorf("unhandled conversion for %q: %T -> %s", name, src, dest.Value.Type().Name())
			}
			l.deferred = append(l.deferred, d)
			return true, nil
		}
	}

	return false, nil
}

func (l *Loader) populateResource(name string, dest utils.Settable, src interface{}, args []string) error {
	if src == nil {
		return nil
	}

	destTypeName := utils.BuildTypeName(dest.Type())

	switch destTypeName {
	case "Resource":
		{
			switch src := src.(type) {
			case []byte:
				if len(args) != 0 {
					return fmt.Errorf("cannot have arguments with static resources")
				}
				dest.Set(reflect.ValueOf(fi.NewBytesResource(src)))

			default:
				if resource, ok := src.(fi.Resource); ok {
					if len(args) != 0 {
						templateResource, ok := resource.(fi.TemplateResource)
						if !ok {
							return fmt.Errorf("cannot have arguments with resources of type %T", resource)
						}
						resource = templateResource.Curry(args)
					}
					dest.Set(reflect.ValueOf(resource))
				} else {
					return fmt.Errorf("unhandled conversion for %q: %T -> %s", name, src, destTypeName)
				}
			}
			return nil
		}

	default:
		return fmt.Errorf("unhandled destination type for %q: %s", name, destTypeName)
	}

}

func (l *Loader) buildNodeConfig(target string, configResourceName string, args []string) (string, error) {
	assetDir := path.Join(l.StateDir, "node/assets")

	confData, err := l.renderResource(configResourceName, args)
	if err != nil {
		return "", err
	}

	config := &nodeup.NodeConfig{}
	err = utils.YamlUnmarshal([]byte(confData), config)
	if err != nil {
		return "", fmt.Errorf("error parsing configuration %q: %v", configResourceName, err)
	}

	cmd := &nodeup.NodeUpCommand{
		Config:         config,
		ConfigLocation: "",
		ModelDir:       l.NodeModelDir,
		Target:         target,
		AssetDir:       assetDir,
	}

	var buff bytes.Buffer
	err = cmd.Run(&buff)
	if err != nil {
		return "", fmt.Errorf("error building node configuration: %v", err)
	}

	return buff.String(), nil
}

func (l *Loader) renderResource(resourceName string, args []string) (string, error) {
	resourceKey := strings.TrimSuffix(resourceName, ".template")
	resourceKey = strings.TrimPrefix(resourceKey, "resources/")
	configResource := l.Resources[resourceKey]
	if configResource == nil {
		return "", fmt.Errorf("cannot find resource %q", resourceName)
	}

	if tr, ok := configResource.(fi.TemplateResource); ok {
		configResource = tr.Curry(args)
	} else if len(args) != 0 {
		return "", fmt.Errorf("args passed when building node config, but config was not a template %q", resourceName)
	}

	data, err := fi.ResourceAsBytes(configResource)
	if err != nil {
		return "", fmt.Errorf("error reading resource %q: %v", resourceName, err)
	}

	return string(data), nil
}
