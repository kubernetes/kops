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

package channels

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/kops/pkg/applylib/applyset"
	"k8s.io/kops/pkg/kubemanifest"
)

type ClientApplier struct {
	Client     dynamic.Interface
	RESTMapper *restmapper.DeferredDiscoveryRESTMapper
}

// Apply applies the manifest to the cluster.
func (p *ClientApplier) Apply(ctx context.Context, manifest []byte) error {
	objects, err := kubemanifest.LoadObjectsFrom(manifest)
	if err != nil {
		return fmt.Errorf("failed to parse objects: %w", err)
	}

	// TODO: Cache applyset for more efficient applying
	patchOptions := metav1.PatchOptions{
		FieldManager: "kops",
	}

	// We force to overcome errors like: Apply failed with 1 conflict: conflict with "kubectl-client-side-apply" using apps/v1: .spec.template.spec.containers[name="foo"].image
	// TODO: How to handle this better?   In a controller we don't have a choice and have to force eventually.
	// But we could do something like try first without forcing, log the conflict if there is one, and then force.
	// This would mean that if there was a loop we could log/detect it.
	// We could even do things like back-off on the force apply.
	force := true
	patchOptions.Force = &force

	s, err := applyset.New(applyset.Options{
		RESTMapper:   p.RESTMapper,
		Client:       p.Client,
		PatchOptions: patchOptions,
	})
	if err != nil {
		return err
	}

	var applyableObjects []applyset.ApplyableObject
	for _, object := range objects {
		applyableObjects = append(applyableObjects, object)
	}
	if err := s.SetDesiredObjects(applyableObjects); err != nil {
		return err
	}

	results, err := s.ApplyOnce(ctx)
	if err != nil {
		return fmt.Errorf("failed to apply objects: %w", err)
	}

	// TODO: Implement pruning

	if !results.AllApplied() {
		return fmt.Errorf("not all objects were applied")
	}

	// TODO: Check object health status
	if !results.AllHealthy() {
		return fmt.Errorf("not all objects were healthy")
	}

	return nil
}
