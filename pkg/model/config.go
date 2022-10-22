/*
Copyright 2021 The Kubernetes Authors.

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

package model

import (
	"fmt"

	kopsbase "k8s.io/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

// ConfigBuilder populates the config store.
type ConfigBuilder struct {
	*KopsModelContext

	Lifecycle fi.Lifecycle
}

func (b *ConfigBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	c.AddTask(&fitasks.ManagedFile{
		Name:      fi.PtrTo(registry.PathKopsVersionUpdated),
		Lifecycle: b.Lifecycle,
		Base:      fi.PtrTo(b.Cluster.Spec.ConfigBase),
		Location:  fi.PtrTo(registry.PathKopsVersionUpdated),
		Contents:  fi.NewStringResource(kopsbase.Version),
	})

	versionedYaml, err := kopscodecs.ToVersionedYamlWithVersion(b.Cluster, v1alpha2.SchemeGroupVersion)
	if err != nil {
		return fmt.Errorf("serializing completed cluster spec: %w", err)
	}
	c.AddTask(&fitasks.ManagedFile{
		Name:      fi.PtrTo(registry.PathClusterCompleted),
		Lifecycle: b.Lifecycle,
		Base:      fi.PtrTo(b.Cluster.Spec.ConfigBase),
		Location:  fi.PtrTo(registry.PathClusterCompleted),
		Contents:  fi.NewBytesResource(versionedYaml),
	})

	return nil
}
