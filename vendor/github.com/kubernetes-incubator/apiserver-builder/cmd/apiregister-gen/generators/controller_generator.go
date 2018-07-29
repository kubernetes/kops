/*
Copyright 2017 The Kubernetes Authors.

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
	"io"
	"strings"
	"text/template"

	"github.com/markbates/inflect"
	"k8s.io/gengo/generator"
)

type controllerGenerator struct {
	generator.DefaultGen
	controller Controller
}

var _ generator.Generator = &controllerGenerator{}

func CreateControllerGenerator(controller Controller, filename string) generator.Generator {
	return &controllerGenerator{
		generator.DefaultGen{OptionalName: filename},
		controller,
	}
}

func (d *controllerGenerator) Imports(c *generator.Context) []string {
	im := []string{
		"github.com/golang/glog",
		"github.com/kubernetes-incubator/apiserver-builder/pkg/controller",
		"k8s.io/apimachinery/pkg/api/errors",
		"k8s.io/client-go/rest",
		"k8s.io/client-go/tools/cache",
		"k8s.io/client-go/util/workqueue",
		d.controller.Repo + "/pkg/controller/sharedinformers",
	}

	return im
}

func (d *controllerGenerator) Finalize(context *generator.Context, w io.Writer) error {
	temp := template.Must(template.New("controller-template").Funcs(
		template.FuncMap{
			"title":  strings.Title,
			"plural": inflect.NewDefaultRuleset().Pluralize,
		},
	).Parse(ControllerAPITemplate))
	return temp.Execute(w, d.controller)
}

var ControllerAPITemplate = `
// {{.Target.Kind}}Controller implements the controller.{{.Target.Kind}}Controller interface
type {{.Target.Kind}}Controller struct {
	queue *controller.QueueWorker

	// Handles messages
	controller *{{.Target.Kind}}ControllerImpl

	Name string

	BeforeReconcile func(key string)
	AfterReconcile  func(key string, err error)

	Informers *sharedinformers.SharedInformers
}

// NewController returns a new {{.Target.Kind}}Controller for responding to {{.Target.Kind}} events
func New{{.Target.Kind}}Controller(config *rest.Config, si *sharedinformers.SharedInformers) *{{.Target.Kind}}Controller {
	q := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "{{.Target.Kind}}")

	queue := &controller.QueueWorker{q, 10, "{{.Target.Kind}}", nil}
	c := &{{.Target.Kind}}Controller{queue, nil, "{{.Target.Kind}}", nil, nil, si}

	// For non-generated code to add events
	uc := &{{.Target.Kind}}ControllerImpl{}
	var ci sharedinformers.Controller = uc

    // Call the Init method that is implemented.
    // Support multiple Init methods for backwards compatibility
	if i, ok := ci.(sharedinformers.LegacyControllerInit); ok {
        i.Init(config, si, c.LookupAndReconcile)
    } else if i, ok := ci.(sharedinformers.ControllerInit); ok {
        i.Init(&sharedinformers.ControllerInitArgumentsImpl{si, config, c.LookupAndReconcile})
    }

	c.controller = uc

	queue.Reconcile = c.reconcile
	if c.Informers.WorkerQueues == nil {
		c.Informers.WorkerQueues = map[string]*controller.QueueWorker{}
	}
	c.Informers.WorkerQueues["{{.Target.Kind}}"] = queue
	si.Factory.{{title .Target.Group}}().{{title .Target.Version}}().{{plural .Target.Kind }}().Informer().
        AddEventHandler(&controller.QueueingEventHandler{q, nil, false})
	return c
}

func (c *{{.Target.Kind}}Controller) GetName() string {
	return c.Name
}

func (c *{{.Target.Kind}}Controller) LookupAndReconcile(key string) (err error) {
	return c.reconcile(key)
}

func (c *{{.Target.Kind}}Controller) reconcile(key string) (err error) {
	var namespace, name string

	if c.BeforeReconcile != nil {
		c.BeforeReconcile(key)
	}
	if c.AfterReconcile != nil {
		// Wrap in a function so err is evaluated after it is set
		defer func() { c.AfterReconcile(key, err) }()
	}

	namespace, name, err = cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return
	}

	u, err := c.controller.Get(namespace, name)
	if errors.IsNotFound(err) {
		glog.Infof("Not doing work for {{.Target.Kind}} %v because it has been deleted", key)
		// Set error so it is picked up by AfterReconcile and the return function
		err = nil
		return
	}
	if err != nil {
		glog.Errorf("Unable to retrieve {{.Target.Kind}} %v from store: %v", key, err)
		return
	}

	// Set error so it is picked up by AfterReconcile and the return function
	err = c.controller.Reconcile(u)

	return
}

func (c *{{.Target.Kind}}Controller) Run(stopCh <-chan struct{}) {
	for _, q := range c.Informers.WorkerQueues {
		q.Run(stopCh)
	}
    controller.GetDefaults(c.controller).Run(stopCh)
}
`

type allControllerGenerator struct {
	generator.DefaultGen
	Controllers []Controller
}

var _ generator.Generator = &allControllerGenerator{}

func CreateAllControllerGenerator(controllers []Controller, filename string) generator.Generator {
	return &allControllerGenerator{
		generator.DefaultGen{OptionalName: filename},
		controllers,
	}
}

func (d *allControllerGenerator) Imports(c *generator.Context) []string {
	if len(d.Controllers) == 0 {
		return []string{}
	}

	repo := d.Controllers[0].Repo
	im := []string{
		"k8s.io/client-go/rest",
		"github.com/kubernetes-incubator/apiserver-builder/pkg/controller",
		repo + "/pkg/controller/sharedinformers",
	}

	// Import package for each controller
	repos := map[string]string{}
	for _, c := range d.Controllers {
		repos[c.Pkg.Path] = ""
	}
	for k, _ := range repos {
		im = append(im, k)
	}

	return im
}

func (d *allControllerGenerator) Finalize(context *generator.Context, w io.Writer) error {
	temp := template.Must(template.New("all-controller-template").Funcs(
		template.FuncMap{
			"title":  strings.Title,
			"plural": inflect.NewDefaultRuleset().Pluralize,
		},
	).Parse(AllControllerAPITemplate))
	return temp.Execute(w, d)
}

var AllControllerAPITemplate = `

func GetAllControllers(config *rest.Config) ([]controller.Controller, chan struct{}) {
	shutdown := make(chan struct{})
	si := sharedinformers.NewSharedInformers(config, shutdown)
	return []controller.Controller{
		{{ range $c := .Controllers -}}
		{{ $c.Pkg.Name }}.New{{ $c.Target.Kind }}Controller(config, si),
		{{ end -}}
	}, shutdown
}

`

type informersGenerator struct {
	generator.DefaultGen
	Controllers []Controller
}

var _ generator.Generator = &informersGenerator{}

func CreateInformersGenerator(controllers []Controller, filename string) generator.Generator {
	return &informersGenerator{
		generator.DefaultGen{OptionalName: filename},
		controllers,
	}
}

func (d *informersGenerator) Imports(c *generator.Context) []string {
	if len(d.Controllers) == 0 {
		return []string{}
	}

	repo := d.Controllers[0].Repo
	return []string{
		"time",
		"github.com/kubernetes-incubator/apiserver-builder/pkg/controller",
		"k8s.io/client-go/rest",
		repo + "/pkg/client/clientset_generated/clientset",
		repo + "/pkg/client/informers_generated/externalversions",
		"k8s.io/client-go/tools/cache",
	}
}

func (d *informersGenerator) Finalize(context *generator.Context, w io.Writer) error {
	temp := template.Must(template.New("informersGenerator-template").Funcs(
		template.FuncMap{
			"title":  strings.Title,
			"plural": inflect.NewDefaultRuleset().Pluralize,
		},
	).Parse(InformersTemplate))
	return temp.Execute(w, d.Controllers)
}

var InformersTemplate = `
// SharedInformers wraps all informers used by controllers so that
// they are shared across controller implementations
type SharedInformers struct {
	controller.SharedInformersDefaults
	Factory           externalversions.SharedInformerFactory
}

// newSharedInformers returns a set of started informers
func NewSharedInformers(config *rest.Config, shutdown <-chan struct{}) *SharedInformers {
	si := &SharedInformers{
		controller.SharedInformersDefaults{},
		externalversions.NewSharedInformerFactory(clientset.NewForConfigOrDie(config), 10*time.Minute),
	}
    if si.SetupKubernetesTypes() {
        si.InitKubernetesInformers(config)
    }
	si.Init()
	si.startInformers(shutdown)
	si.StartAdditionalInformers(shutdown)
	return si
}

// startInformers starts all of the informers
func (si *SharedInformers) startInformers(shutdown <-chan struct{}) {
	{{ range $c := . -}}
	go si.Factory.{{title $c.Target.Group}}().{{title $c.Target.Version}}().{{plural $c.Target.Kind}}().Informer().Run(shutdown)
	{{ end -}}
}

// ControllerInitArguments are arguments provided to the Init function for a new controller.
type ControllerInitArguments interface {
    // GetSharedInformers returns the SharedInformers that can be used to access
    // informers and listers for watching and indexing Kubernetes Resources
    GetSharedInformers() *SharedInformers

    // GetRestConfig returns the Config to create new client-go clients
    GetRestConfig() *rest.Config

    // Watch uses resourceInformer to watch a resource.  When create, update, or deletes
    // to the resource type are encountered, watch uses watchResourceToReconcileResourceKey
    // to lookup the key for the resource reconciled by the controller (maybe a different type
    // than the watched resource), and enqueue it to be reconciled.
    // watchName: name of the informer.  may appear in logs
    // resourceInformer: gotten from the SharedInformer.  controls which resource type is watched
    // getReconcileKey: takes an instance of the watched resource and returns
    //                                      a key for the reconciled resource type to enqueue.
	Watch(watchName string, resourceInformer cache.SharedIndexInformer,
            getReconcileKey func(interface{}) (string, error))
}

type ControllerInitArgumentsImpl struct {
    Si *SharedInformers
    Rc *rest.Config
    Rk func(key string) error
}

func (c ControllerInitArgumentsImpl) GetSharedInformers() *SharedInformers {
  return c.Si
}

func (c ControllerInitArgumentsImpl) GetRestConfig() *rest.Config {
  return c.Rc
}

// Watch uses resourceInformer to watch a resource.  When create, update, or deletes
// to the resource type are encountered, watch uses watchResourceToReconcileResourceKey
// to lookup the key for the resource reconciled by the controller (maybe a different type
// than the watched resource), and enqueue it to be reconciled.
// watchName: name of the informer.  may appear in logs
// resourceInformer: gotten from the SharedInformer.  controls which resource type is watched
// getReconcileKey: takes an instance of the watched resource and returns
//                                      a key for the reconciled resource type to enqueue.
func (c ControllerInitArgumentsImpl) Watch(
    watchName string, resourceInformer cache.SharedIndexInformer,
    getReconcileKey func(interface{}) (string, error)) {
    c.Si.Watch(watchName, resourceInformer, getReconcileKey, c.Rk)
}

type Controller interface {}

// LegacyControllerInit old controllers may implement this, and we keep
// it for backwards compatibility.
type LegacyControllerInit interface {
    Init(config *rest.Config, si *SharedInformers, r func(key string) error)
}

// ControllerInit new controllers should implement this.  It is more flexible in
// allowing additional options to be passed in
type ControllerInit interface {
    Init(args ControllerInitArguments)
}
`
