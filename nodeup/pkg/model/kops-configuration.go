/*
Copyright 2023 The Kubernetes Authors.

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
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// KopsConfigurationBuilder configures the kops-configuration service
type KopsConfigurationBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &KopsConfigurationBuilder{}

// Build is responsible for disabling the kops-cofiguration service
func (b *KopsConfigurationBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	serviceName := "kops-configuration.service"

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Run kOps bootstrap (nodeup)")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/kops-configuration")
	manifest.Set("Service", "EnvironmentFile", "/etc/environment")
	manifest.Set("Service", "ExecStart", "/opt/kops/bin/nodeup --conf=/opt/kops/conf/kube_env.yaml --v=8")
	manifest.Set("Service", "Type", "oneshot")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", serviceName, manifestString)

	service := &nodetasks.Service{
		Name:        serviceName,
		Definition:  fi.PtrTo(manifestString),
		Enabled:     fi.PtrTo(false),
		ManageState: fi.PtrTo(false),
	}

	service.InitDefaults()

	c.AddTask(service)

	return nil
}
