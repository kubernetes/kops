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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

// ImageRegistryTransform modifies all Pods to use registry for the image source and adds the imagePullSecret
func ImageRegistryTransform(registry, imagePullSecret string) ObjectTransform {
	return func(c context.Context, o DeclarativeObject, m *manifest.Objects) error {
		return applyImageRegistry(c, o, m, registry, imagePullSecret)
	}
}

func applyImageRegistry(ctx context.Context, operatorObject DeclarativeObject, manifest *manifest.Objects, registry, secret string) error {
	log := log.Log
	if registry == "" && secret == "" {
		return nil
	}
	for _, manifestItem := range manifest.Items {
		if manifestItem.Kind == "Deployment" || manifestItem.Kind == "DaemonSet" ||
			manifestItem.Kind == "StatefulSet" || manifestItem.Kind == "Job" ||
			manifestItem.Kind == "CronJob" {
			if registry != "" {
				log.WithValues("manifest", manifestItem).WithValues("registry", registry).V(1).Info("applying image registory to manifest")
				if err := manifestItem.MutateContainers(applyPrivateRegistryToContainer(registry)); err != nil {
					return fmt.Errorf("error applying private registry: %v", err)
				}
			}
			if secret != "" {
				log.WithValues("manifest", manifestItem).WithValues("secret", secret).V(1).Info("applying image pull secret to manifest")
				if err := manifestItem.MutatePodSpec(applyImagePullSecret(secret)); err != nil {
					return fmt.Errorf("error applying image pull secret: %v", err)
				}
			}
		}
	}
	return nil
}

func applyImagePullSecret(secret string) func(map[string]interface{}) error {
	return func(podSpec map[string]interface{}) error {
		imagePullSecret := map[string]interface{}{"name": secret}
		if err := unstructured.SetNestedSlice(podSpec, []interface{}{imagePullSecret}, "imagePullSecrets"); err != nil {
			return fmt.Errorf("error applying pull image secret: %v", err)
		}
		return nil
	}
}

func applyPrivateRegistryToContainer(registry string) func(map[string]interface{}) error {
	return func(container map[string]interface{}) error {
		image, _, err := unstructured.NestedString(container, "image")
		if err != nil {
			return fmt.Errorf("error reading container image: %v", err)
		}
		container["image"] = applyPrivateRegistryToImage(registry, image)
		return nil
	}
}

func applyPrivateRegistryToImage(registry, image string) string {
	parts := strings.SplitN(image, "/", 3)
	imageName := parts[len(parts)-1]
	return registry + "/" + imageName
}
