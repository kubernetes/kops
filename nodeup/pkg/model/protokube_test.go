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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
)

func TestProtokubeBuilder(t *testing.T) {
	basedir := "tests/protokube/docker"

	context := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}
	nodeUpModelContext, err := BuildNodeupModelContext(basedir)
	if err != nil {
		t.Fatalf("error loading model %q: %v", basedir, err)
		return
	}

	cluster := nodeupModelContext.Cluster
	if cluster.Spec.MasterKubelet == nil {
		cluster.Spec.MasterKubelet = &kops.KubeletConfigSpec{}
	}
	if cluster.Spec.MasterKubelet == nil {
		cluster.Spec.MasterKubelet = &kops.KubeletConfigSpec{}
	}
	cluster.Spec.Kubelet.HostnameOverride = "example-hostname"

	nodeUpModelContext.IsMaster = true

	nodeUpModelContext.NodeupConfig = &nodeup.Config{}

	// These trigger use of etcd-manager
	nodeUpModelContext.NodeupConfig.EtcdManifests = []string{
		"memfs://clusters.example.com/minimal.example.com/manifests/etcd/main.yaml",
		"memfs://clusters.example.com/minimal.example.com/manifests/etcd/events.yaml",
	}

	nodeUpModelContext.NodeupConfig.ProtokubeImage = &nodeup.Image{}
	nodeUpModelContext.NodeupConfig.ProtokubeImage.Name = "protokube:test"

	builder := &ProtokubeBuilder{NodeupModelContext: nodeUpModelContext}

	if task, err := builder.buildSystemdService(); err != nil {
		t.Fatalf("error from buildSystemdService: %v", err)
	} else {
		context.AddTask(task)
	}

	testutils.ValidateTasks(t, basedir, context)
}
