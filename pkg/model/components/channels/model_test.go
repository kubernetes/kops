/*
Copyright 2026 The Kubernetes Authors.

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
	"fmt"
	"path/filepath"
	"testing"

	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

func Test_RunChannelsBuilder(t *testing.T) {
	// Stable golden image strings.
	featureflag.ParseFlags("-ImageDigest")
	// channelList builds VFS paths from configStore.base — fixtures use memfs://.
	vfs.Context.ResetMemfsContext(true)
	tests := []string{
		"tests/minimal",
		"tests/container_registry",
	}
	for _, basedir := range tests {
		t.Run(fmt.Sprintf("basedir=%s", basedir), func(t *testing.T) {
			context := &fi.CloudupModelBuilderContext{
				Tasks: make(map[string]fi.CloudupTask),
			}
			kopsModelContext, err := loadKopsModelContext(basedir)
			if err != nil {
				t.Fatalf("error loading model %q: %v", basedir, err)
			}

			builder := ChannelsBuilder{
				KopsModelContext: kopsModelContext,
				AssetBuilder:     assets.NewAssetBuilder(vfs.Context, kopsModelContext.Cluster.Spec.Assets, false),
			}

			if err := builder.Build(context); err != nil {
				t.Fatalf("error from Build: %v", err)
			}

			testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
		})
	}
}

func loadKopsModelContext(basedir string) (*model.KopsModelContext, error) {
	spec, err := testutils.LoadModel(basedir)
	if err != nil {
		return nil, err
	}
	if spec.Cluster == nil {
		return nil, fmt.Errorf("no cluster found in %s", basedir)
	}
	return &model.KopsModelContext{
		IAMModelContext:   iam.IAMModelContext{Cluster: spec.Cluster},
		AllInstanceGroups: spec.InstanceGroups,
		InstanceGroups:    spec.InstanceGroups,
	}, nil
}
