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

package declarative

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/kubectlcmd"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

var _ reconcile.Reconciler = &Reconciler{}

type Reconciler struct {
	prototype DeclarativeObject
	client    client.Client
	config    *rest.Config
	kubectl   kubectlClient

	mgr manager.Manager

	options reconcilerParams
}

type kubectlClient interface {
	Apply(ctx context.Context, namespace string, manifest string, args ...string) error
}

type DeclarativeObject interface {
	runtime.Object
	metav1.Object
}

// For mocking
var kubectl = kubectlcmd.New()

func (r *Reconciler) Init(mgr manager.Manager, prototype DeclarativeObject, opts ...reconcilerOption) error {
	r.prototype = prototype
	r.kubectl = kubectl

	r.client = mgr.GetClient()
	r.config = mgr.GetConfig()
	r.mgr = mgr

	r.applyOptions(opts...)

	return r.validateOptions()
}

// +rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	log := log.Log

	// Fetch the object
	instance := r.prototype.DeepCopyObject().(DeclarativeObject)
	if err := r.client.Get(ctx, request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "error reading object")
		return reconcile.Result{}, err
	}

	if r.options.status != nil {
		if err := r.options.status.Preflight(ctx, instance); err != nil {
			log.Error(err, "preflight check failed, not reconciling")
			return reconcile.Result{}, err
		}
	}

	return r.reconcileExists(ctx, request.NamespacedName, instance)
}

func (r *Reconciler) reconcileExists(ctx context.Context, name types.NamespacedName, instance DeclarativeObject) (reconcile.Result, error) {
	log := log.Log
	log.WithValues("object", name.String()).Info("reconciling")

	objects, err := r.BuildDeploymentObjects(ctx, name, instance)
	if err != nil {
		log.Error(err, "building deployment objects")
		return reconcile.Result{}, fmt.Errorf("error building deployment objects: %v", err)
	}
	log.WithValues("objects", fmt.Sprintf("%d", len(objects.Items))).Info("built deployment objects")

	defer func() {
		if r.options.status != nil {
			if err := r.options.status.Reconciled(ctx, instance, objects); err != nil {
				log.Error(err, "failed to reconcile status")
			}
		}
	}()

	err = r.injectOwnerRef(ctx, instance, objects)
	if err != nil {
		return reconcile.Result{}, err
	}

	m, err := objects.JSONManifest()
	if err != nil {
		log.Error(err, "creating final manifest")
	}

	extraArgs := []string{"--force"}

	if r.options.prune {
		var labels []string
		for k, v := range r.options.labelMaker(ctx, instance) {
			labels = append(labels, fmt.Sprintf("%s=%s", k, v))
		}

		extraArgs = append(extraArgs, "--prune", "--selector", strings.Join(labels, ","))
	}

	ns := ""
	if !r.options.preserveNamespace {
		ns = name.Namespace
	}

	if err := r.kubectl.Apply(ctx, ns, m, extraArgs...); err != nil {
		log.Error(err, "applying manifest")
		return reconcile.Result{}, fmt.Errorf("error applying manifest: %v", err)
	}

	if r.options.sink != nil {
		if err := r.options.sink.Notify(ctx, instance, objects); err != nil {
			log.Error(err, "notifying sink")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// BuildDeploymentObjects performs all manifest operations to build a final set of objects for deployment
func (r *Reconciler) BuildDeploymentObjects(ctx context.Context, name types.NamespacedName, instance DeclarativeObject) (*manifest.Objects, error) {
	log := log.Log

	// 1. Load the manifest
	manifestStr, err := r.loadRawManifest(ctx, instance)
	if err != nil {
		log.Error(err, "error loading raw manifest")
		return nil, err
	}

	// 2. Perform raw string operations
	for _, t := range r.options.rawManifestOperations {
		transformed, err := t(ctx, instance, manifestStr)
		if err != nil {
			log.Error(err, "error performing raw manifest operations")
			return nil, err
		}
		manifestStr = transformed
	}

	// 3. Parse manifest into objects
	objects, err := manifest.ParseObjects(ctx, manifestStr)
	if err != nil {
		log.Error(err, "error parsing manifest")
		return nil, err
	}

	// 4. Perform object transformations
	transforms := r.options.objectTransformations
	if r.options.labelMaker != nil {
		transforms = append(transforms, AddLabels(r.options.labelMaker(ctx, instance)))
	}
	// TODO(jrjohnson): apply namespace here
	for _, t := range transforms {
		err := t(ctx, instance, objects)
		if err != nil {
			return nil, err
		}
	}

	// 5. Sort objects to work around dependent objects in the same manifest (eg: service-account, deployment)
	objects.Sort(DefaultObjectOrder(ctx))

	return objects, nil
}

// loadRawManifest loads the raw manifest YAML from the repository
func (r *Reconciler) loadRawManifest(ctx context.Context, o DeclarativeObject) (string, error) {
	s, err := r.options.manifestController.ResolveManifest(ctx, o)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (r *Reconciler) applyOptions(opts ...reconcilerOption) {
	params := reconcilerParams{}

	opts = append(Options.Begin, opts...)
	opts = append(opts, Options.End...)

	for _, opt := range opts {
		params = opt(params)
	}

	// Default the manifest controller if not set
	if params.manifestController == nil && DefaultManifestLoader != nil {
		params.manifestController = DefaultManifestLoader()
	}

	r.options = params
}

// Validate compatibility of selected options
func (r *Reconciler) validateOptions() error {
	var errs []string

	if r.options.prune && r.options.labelMaker == nil {
		errs = append(errs, "WithApplyPrune must be used with the WithLabels option")
	}

	if r.options.manifestController == nil {
		errs = append(errs, "ManifestController must be set either by configuring DefaultManifestLoader or specifying the WithManifestController option")
	}

	if len(errs) != 0 {
		return fmt.Errorf(strings.Join(errs, ","))
	}

	return nil
}

func (r *Reconciler) injectOwnerRef(ctx context.Context, instance DeclarativeObject, objects *manifest.Objects) error {
	if r.options.ownerFn == nil {
		return nil
	}

	log := log.Log
	log.WithValues("object", fmt.Sprintf("%s/%s", instance.GetName(), instance.GetNamespace())).Info("injecting owner references")

	for _, o := range objects.Items {
		owner, err := r.options.ownerFn(ctx, instance, *o, *objects)
		if err != nil {
			log.WithValues("object", o).Error(err, "resolving owner ref", o)
			return err
		}
		if owner == nil {
			log.WithValues("object", o).Info("no owner resolved")
			continue
		}
		if owner.GetName() == "" {
			log.WithValues("object", o).Info("has no name")
			continue
		}
		if string(owner.GetUID()) == "" {
			log.WithValues("object", o).Info("has no UID")
			continue
		}

		gvk, err := apiutil.GVKForObject(owner, r.mgr.GetScheme())
		if gvk.Group == "" || gvk.Version == "" {
			log.WithValues("object", o).WithValues("GroupVersionKind", gvk).Info("is not valid")
			continue
		}

		// TODO, error/skip if:
		// - owner is namespaced and o is not
		// - owner is in a different namespace than o

		ownerRefs := []interface{}{
			map[string]interface{}{
				"apiVersion":         gvk.Group + "/" + gvk.Version,
				"blockOwnerDeletion": true,
				"controller":         true,
				"kind":               gvk.Kind,
				"name":               owner.GetName(),
				"uid":                string(owner.GetUID()),
			},
		}
		if err := o.SetNestedField(ownerRefs, "metadata", "ownerReferences"); err != nil {
			return err
		}
	}
	return nil
}

// SetSink provides a Sink that will be notified for all deployments
func (r *Reconciler) SetSink(sink Sink) {
	r.options.sink = sink
}
