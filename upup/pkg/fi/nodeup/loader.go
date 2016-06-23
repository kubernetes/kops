package nodeup

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/loader"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/nodetasks"
	"os"
	"strings"
	"text/template"
)

type Loader struct {
	templates     []*template.Template
	optionsLoader *loader.OptionsLoader
	config        *NodeConfig

	assets *fi.AssetStore
	tasks  map[string]fi.Task

	tags map[string]struct{}

	TemplateFunctions template.FuncMap
}

func NewLoader(config *NodeConfig, assets *fi.AssetStore) *Loader {
	l := &Loader{}
	l.assets = assets
	l.tasks = make(map[string]fi.Task)
	l.optionsLoader = loader.NewOptionsLoader(config)
	l.config = config
	l.TemplateFunctions = make(template.FuncMap)

	return l
}

func (l *Loader) executeTemplate(key string, d string) (string, error) {
	t := template.New(key)

	funcMap := make(template.FuncMap)
	funcMap["BuildFlags"] = buildFlags
	funcMap["Base64Encode"] = func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	funcMap["HasTag"] = func(tag string) bool {
		_, found := l.tags[tag]
		return found
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
	tags := make(map[string]struct{})
	for _, tag := range l.config.Tags {
		tags[tag] = struct{}{}
	}

	l.tags = tags

	// First pass: load options
	tw := &loader.TreeWalker{
		DefaultHandler: ignoreHandler,
		Contexts: map[string]loader.Handler{
			"options":  l.optionsLoader.HandleOptions,
			"files":    ignoreHandler,
			"disks":    ignoreHandler,
			"packages": ignoreHandler,
			"services": ignoreHandler,
			"users":    ignoreHandler,
		},
		Tags: tags,
	}

	err := tw.Walk(baseDir)
	if err != nil {
		return nil, err
	}

	config, err := l.optionsLoader.Build()
	if err != nil {
		return nil, err
	}
	l.config = config.(*NodeConfig)
	glog.V(4).Infof("options: %s", fi.DebugAsJsonStringIndent(l.config))

	// Second pass: load everything else
	tw = &loader.TreeWalker{
		DefaultHandler: l.handleFile,
		Contexts: map[string]loader.Handler{
			"options":  ignoreHandler,
			"files":    l.handleFile,
			"disks":    l.newTaskHandler("disk/", nodetasks.NewMountDiskTask),
			"packages": l.newTaskHandler("package/", nodetasks.NewPackage),
			"services": l.newTaskHandler("service/", nodetasks.NewService),
			"users":    l.newTaskHandler("user/", nodetasks.NewUserTask),
		},
		Tags: tags,
	}

	err = tw.Walk(baseDir)
	if err != nil {
		return nil, err
	}

	// If there is a package task, we need an update packages task
	for _, t := range l.tasks {
		if _, ok := t.(*nodetasks.Package); ok {
			l.tasks["UpdatePackages"] = &nodetasks.UpdatePackages{}
		}
	}

	return l.tasks, nil
}

type TaskBuilder func(name string, contents string, meta string) (fi.Task, error)

func (r *Loader) newTaskHandler(prefix string, builder TaskBuilder) loader.Handler {
	return func(i *loader.TreeWalkItem) error {
		contents, err := i.ReadString()
		if err != nil {
			return err
		}
		task, err := builder(i.Name, contents, i.Meta)
		if err != nil {
			return fmt.Errorf("error building %s for %q: %v", i.Name, i.Path, err)
		}
		key := prefix + i.RelativePath

		if task != nil {
			r.tasks[key] = task
		}
		return nil
	}
}

func (r *Loader) handleFile(i *loader.TreeWalkItem) error {
	var task *nodetasks.File
	defaultFileType := nodetasks.FileType_File

	var err error
	if strings.HasSuffix(i.RelativePath, ".template") {
		contents, err := i.ReadString()
		if err != nil {
			return err
		}

		// TODO: Use template resource here to defer execution?
		destPath := "/" + strings.TrimSuffix(i.RelativePath, ".template")
		name := strings.TrimSuffix(i.Name, ".template")
		expanded, err := r.executeTemplate(name, contents)
		if err != nil {
			return fmt.Errorf("error executing template %q: %v", i.RelativePath, err)
		}

		task, err = nodetasks.NewFileTask(name, fi.NewStringResource(expanded), destPath, i.Meta)
	} else if strings.HasSuffix(i.RelativePath, ".asset") {
		contents, err := i.ReadBytes()
		if err != nil {
			return err
		}

		destPath := "/" + strings.TrimSuffix(i.RelativePath, ".asset")
		name := strings.TrimSuffix(i.Name, ".asset")

		def := &nodetasks.AssetDefinition{}
		err = json.Unmarshal(contents, def)
		if err != nil {
			return fmt.Errorf("error parsing json for asset %q: %v", name, err)
		}

		asset, err := r.assets.Find(name, def.AssetPath)
		if err != nil {
			return fmt.Errorf("error trying to locate asset %q: %v", name, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", name)
		}

		task, err = nodetasks.NewFileTask(i.Name, asset, destPath, i.Meta)
	} else {
		stat, err := os.Stat(i.Path)
		if err != nil {
			return fmt.Errorf("error doing stat on %q: %v", i.Path, err)
		}
		var contents fi.Resource
		if stat.IsDir() {
			defaultFileType = nodetasks.FileType_Directory
		} else {
			contents = fi.NewFileResource(i.Path)
		}
		task, err = nodetasks.NewFileTask(i.Name, contents, "/"+i.RelativePath, i.Meta)
	}

	if task.Type == "" {
		task.Type = defaultFileType
	}

	if err != nil {
		return fmt.Errorf("error building task %q: %v", i.RelativePath, err)
	}
	glog.V(2).Infof("path %q -> task %v", i.Path, task)

	if task != nil {
		key := "file/" + i.RelativePath
		r.tasks[key] = task
	}
	return nil
}
