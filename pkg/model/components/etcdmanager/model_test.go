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

package etcdmanager

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

func Test_RunEtcdManagerBuilder(t *testing.T) {
	featureflag.ParseFlags("-ImageDigest")
	tests := []string{
		"tests/minimal",
		"tests/interval",
		"tests/proxy",
		"tests/overwrite_settings",
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

			builder := EtcdManagerBuilder{
				KopsModelContext: kopsModelContext,
				AssetBuilder:     assets.NewAssetBuilder(vfs.Context, kopsModelContext.Cluster.Spec.Assets, false),
			}

			if err := builder.Build(context); err != nil {
				t.Fatalf("error from Build: %v", err)
				return
			}

			testutils.ValidateTasks(t, filepath.Join(basedir, "tasks.yaml"), context)
		})
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

	if len(spec.InstanceGroups) == 0 {
		return nil, fmt.Errorf("no instance groups found in %s", basedir)
	}

	kopsContext := &model.KopsModelContext{
		IAMModelContext:   iam.IAMModelContext{Cluster: spec.Cluster},
		AllInstanceGroups: spec.InstanceGroups,
		InstanceGroups:    spec.InstanceGroups,
	}

	return kopsContext, nil
}

func Test_resolveAzureBackupStore(t *testing.T) {
	tests := []struct {
		name            string
		configStoreBase string
		backupStore     string
		wantURL         string
		wantAccount     string
		wantErr         bool
	}{
		{
			name:            "non-azure backup store passes through",
			configStoreBase: "memfs://tests/cluster",
			backupStore:     "memfs://tests/cluster/backups/etcd/main",
			wantURL:         "memfs://tests/cluster/backups/etcd/main",
			wantAccount:     "",
		},
		{
			name:            "non-azure backup store with azure config base passes through",
			configStoreBase: "azureblob://kopsstate/state/cluster",
			backupStore:     "s3://my-bucket/cluster/backups/etcd/main",
			wantURL:         "s3://my-bucket/cluster/backups/etcd/main",
			wantAccount:     "",
		},
		{
			name:            "azureblob with multi-segment key",
			configStoreBase: "azureblob://kopsstate/state/cluster",
			backupStore:     "azureblob://kopsstate/state/cluster.example.com/backups/etcd/main",
			wantURL:         "azureblob://state/cluster.example.com/backups/etcd/main",
			wantAccount:     "kopsstate",
		},
		{
			name:            "azureblob with empty key",
			configStoreBase: "azureblob://kopsstate/state",
			backupStore:     "azureblob://kopsstate/state",
			wantURL:         "azureblob://state",
			wantAccount:     "kopsstate",
		},
		{
			name:            "account taken from configStore.base, not backup store",
			configStoreBase: "azureblob://canonicalacct/state/cluster",
			backupStore:     "azureblob://canonicalacct/backups/etcd/main",
			wantURL:         "azureblob://backups/etcd/main",
			wantAccount:     "canonicalacct",
		},
		{
			name:            "azureblob backup store with non-azure configStore.base is rejected",
			configStoreBase: "s3://my-bucket/state",
			backupStore:     "azureblob://kopsstate/state/cluster/backups/etcd/main",
			wantErr:         true,
		},
		{
			name:            "azureblob missing container is rejected",
			configStoreBase: "azureblob://kopsstate/state",
			backupStore:     "azureblob://kopsstate",
			wantErr:         true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotURL, gotAccount, err := resolveAzureBackupStore(tc.configStoreBase, tc.backupStore)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got URL=%q account=%q", gotURL, gotAccount)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotURL != tc.wantURL {
				t.Errorf("URL: got %q, want %q", gotURL, tc.wantURL)
			}
			if gotAccount != tc.wantAccount {
				t.Errorf("Account: got %q, want %q", gotAccount, tc.wantAccount)
			}
		})
	}
}
