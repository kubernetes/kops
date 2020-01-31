/*
Copyright 2020 The Kubernetes Authors.

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
	"strconv"

	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// KopsControllerBuilder installs the kops-controller dependencies
type KopsControllerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KopsControllerBuilder{}

// Build is responsible for building the manifest for the kube-scheduler
func (b *KopsControllerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	// We run kops-controller under an unprivileged user (10011), and then grant specific permissions
	c.AddTask(&nodetasks.UserTask{
		Name:  "kops-controller",
		UID:   wellknownusers.KopsController,
		Shell: "/sbin/nologin",
	})

	c.AddTask(&nodetasks.File{
		Path:        "/var/log/kops-controller.log",
		Contents:    fi.NewStringResource(""),
		Type:        nodetasks.FileType_File,
		Mode:        s("0600"),
		Owner:       s(strconv.Itoa(wellknownusers.KopsController)),
		IfNotExists: true,
	})

	return nil
}
