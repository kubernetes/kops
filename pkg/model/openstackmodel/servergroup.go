/*
Copyright 2018 The Kubernetes Authors.

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

package openstackmodel

import (
	"fmt"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

// ServerGroupModelBuilder configures server group objects
type ServerGroupModelBuilder struct {
	*OpenstackModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &ServerGroupModelBuilder{}

func (b *ServerGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	clusterName := b.ClusterName()

	for _, ig := range b.InstanceGroups {
		t := &openstacktasks.ServerGroup{
			Name:      s(fmt.Sprintf("%s-%s", clusterName, ig.Spec.Role)),
			Policies:  []string{"anti-affinity"},
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	return nil
}
