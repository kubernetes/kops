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

package kubescheduler

import (
	"fmt"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_RunKubeSchedulerBuilder(t *testing.T) {
	tests := []string{
		"tests/minimal",
		"tests/kubeschedulerconfig",
		"tests/mixing",
	}
	for _, basedir := range tests {
		basedir := basedir

		t.Run(fmt.Sprintf("basedir=%s", basedir), func(t *testing.T) {
			context := &fi.CloudupModelBuilderContext{
				Tasks: make(map[string]fi.CloudupTask),
			}
			kopsModelContext, err := LoadKopsModelContext(basedir)
			if err != nil {
				t.Fatalf("error loading model %q: %v", basedir, err)
				return
			}

			builder := KubeSchedulerBuilder{
				KopsModelContext: kopsModelContext,
				AssetBuilder:     assets.NewAssetBuilder(vfs.Context, kopsModelContext.Cluster.Spec.Assets, false),
			}

			if err := builder.Build(context); err != nil {
				t.Fatalf("error from Build: %v", err)
				return
			}

			testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
			testutils.ValidateStaticFiles(t, basedir, builder.AssetBuilder)
			testutils.ValidateCompletedCluster(t, filepath.Join(basedir, "completed-cluster.yaml"), builder.Cluster)
		})
	}
}

func Test_MapToUnstructured_WithQpsAndBurst(t *testing.T) {
	qps := resource.MustParse("500")

	kubeScheduler := &kops.KubeSchedulerConfig{
		Qps:   &qps,
		Burst: 500,
	}

	target := &unstructured.Unstructured{}
	target.SetKind("KubeSchedulerConfiguration")
	target.SetAPIVersion("kubescheduler.config.k8s.io/v1")

	err := MapToUnstructured(kubeScheduler, target)
	if err != nil {
		t.Fatalf("MapToUnstructured failed: %v", err)
	}

	qpsVal, found, err := unstructured.NestedFloat64(target.Object, "clientConnection", "qps")
	if err != nil {
		t.Fatalf("error getting qps: %v", err)
	}
	if !found {
		t.Error("qps not found in target")
	}
	if qpsVal != 500.0 {
		t.Errorf("expected qps=500.0, got %v", qpsVal)
	}

	burst, found, err := unstructured.NestedFieldNoCopy(target.Object, "clientConnection", "burst")
	if err != nil {
		t.Fatalf("error getting burst: %v", err)
	}
	if !found {
		t.Error("burst not found in target")
	}

	burstVal, ok := burst.(int32)
	if !ok {
		t.Errorf("expected burst to be int32, got %T: %v", burst, burst)
	}
	if burstVal != 500 {
		t.Errorf("expected burst=500, got %v", burstVal)
	}
}

func LoadKopsModelContext(basedir string) (*model.KopsModelContext, error) {
	spec, err := testutils.LoadModel(basedir)
	if err != nil {
		return nil, err
	}

	if spec.Cluster == nil {
		return nil, fmt.Errorf("no cluster found in %s", basedir)
	}

	kopsContext := &model.KopsModelContext{
		IAMModelContext:   iam.IAMModelContext{Cluster: spec.Cluster},
		AllInstanceGroups: spec.InstanceGroups,
		InstanceGroups:    spec.InstanceGroups,
	}

	for _, u := range spec.AdditionalObjects {
		kopsContext.AdditionalObjects = append(kopsContext.AdditionalObjects, kubemanifest.NewObject(u.Object))
	}

	return kopsContext, nil
}
