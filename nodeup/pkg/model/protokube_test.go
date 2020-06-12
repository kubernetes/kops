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

package model

import (
	"testing"

	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/upup/pkg/fi"
)

func TestProtokubeBuilder_Docker(t *testing.T) {
	RunGoldenTest(t, "tests/protokube/docker", "protokube", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		builder := ProtokubeBuilder{NodeupModelContext: nodeupModelContext}
		populateImage(nodeupModelContext)
		return builder.Build(target)
	})
}

func TestProtokubeBuilder_containerd(t *testing.T) {
	RunGoldenTest(t, "tests/protokube/containerd", "protokube", func(nodeupModelContext *NodeupModelContext, target *fi.ModelBuilderContext) error {
		builder := ProtokubeBuilder{NodeupModelContext: nodeupModelContext}
		populateImage(nodeupModelContext)
		return builder.Build(target)
	})
}

func populateImage(ctx *NodeupModelContext) {
	if ctx.NodeupConfig == nil {
		ctx.NodeupConfig = &nodeup.Config{}
	}
	ctx.NodeupConfig.ProtokubeImage = &nodeup.Image{
		Name: "protokube image name",
	}
}
