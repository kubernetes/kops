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
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
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

func Test_BuildPod_LinodeUsesNativeBackupStoreAndVolumeProvider(t *testing.T) {
	featureflag.ParseFlags("-ImageDigest")

	kopsModelContext, err := LoadKopsModelContext("tests/minimal")
	if err != nil {
		t.Fatalf("error loading model: %v", err)
	}

	kopsModelContext.Cluster.Spec.CloudProvider = kops.CloudProviderSpec{
		Linode: &kops.LinodeSpec{},
	}
	kopsModelContext.Cluster.Spec.EtcdClusters[0].Backups = &kops.EtcdBackupSpec{
		BackupStore: "linode://kops-test/minimal.example.com/backups/etcd/main",
	}

	builder := EtcdManagerBuilder{
		KopsModelContext: kopsModelContext,
		AssetBuilder:     assets.NewAssetBuilder(vfs.Context, kopsModelContext.Cluster.Spec.Assets, false),
	}

	pod, err := builder.buildPod(kopsModelContext.Cluster.Spec.EtcdClusters[0], "master-us-test-1a")
	if err != nil {
		t.Fatalf("buildPod returned error: %v", err)
	}

	if len(pod.Spec.Containers) != 1 {
		t.Fatalf("expected exactly one container, got %d", len(pod.Spec.Containers))
	}

	command := strings.Join(pod.Spec.Containers[0].Command, " ")
	wantArgs := []string{
		"--backup-store=linode://kops-test/minimal.example.com/backups/etcd/main",
		"--volume-provider=linode",
		"--volume-name-tag=kops.k8s.io/instance-group:master-us-test-1a",
		"--volume-tag=kops.k8s.io/cluster:minimal.example.com",
		"--volume-tag=kops.k8s.io/etcd:main",
		"--volume-tag=kops.k8s.io/instance-role:ControlPlane",
	}

	for _, want := range wantArgs {
		if !strings.Contains(command, want) {
			t.Fatalf("expected etcd-manager command to contain %q, got: %s", want, command)
		}
	}

	for _, notWant := range []string{"--data-dir=/var/lib/etcd-manager/main", "--backup-store=s3://", "--backup-store=do://"} {
		if strings.Contains(command, notWant) {
			t.Fatalf("expected etcd-manager command to not contain %q, got: %s", notWant, command)
		}
	}

	envByName := map[string]string{}
	for _, envVar := range pod.Spec.Containers[0].Env {
		envByName[envVar.Name] = envVar.Value
	}

	if envByName["AWS_REQUEST_CHECKSUM_CALCULATION"] != "when_required" {
		t.Fatalf("expected AWS_REQUEST_CHECKSUM_CALCULATION=when_required, got %q", envByName["AWS_REQUEST_CHECKSUM_CALCULATION"])
	}
	if envByName["AWS_RESPONSE_CHECKSUM_VALIDATION"] != "when_required" {
		t.Fatalf("expected AWS_RESPONSE_CHECKSUM_VALIDATION=when_required, got %q", envByName["AWS_RESPONSE_CHECKSUM_VALIDATION"])
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
