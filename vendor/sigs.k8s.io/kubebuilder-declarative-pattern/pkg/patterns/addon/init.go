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

package addon

import (
	"context"
	"flag"
	"sync"

	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/addon/pkg/loaders"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

var (
	privateRegistry = flag.String("private-registry", "", "private image registry, if set overwrites the image repo on all pods")
	imagePullSecret = flag.String("image-pull-secret", "", "secret used accessing private image registry, if set imagePullSecret annotation is added to all pods")
)

var initOnce sync.Once

// Init should be called at the beginning of the main function for all addon operator controllers
//
// This function configures the environment and declarative library
// with defaults specific to addons.
func Init() {
	initOnce.Do(func() {
		if declarative.DefaultManifestLoader == nil {
			declarative.DefaultManifestLoader = func() declarative.ManifestController {
				return loaders.NewManifestLoader()
			}
		}

		declarative.Options.Begin = append(declarative.Options.Begin, declarative.WithObjectTransform(func(ctx context.Context, obj declarative.DeclarativeObject, m *manifest.Objects) error {
			if *privateRegistry != "" || *imagePullSecret != "" {
				return declarative.ImageRegistryTransform(*privateRegistry, *imagePullSecret)(ctx, obj, m)
			}
			return nil
		}))
	})
}
