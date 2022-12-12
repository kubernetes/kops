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

package terraform

import (
	"context"
	"fmt"
	"os"
	"path"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
	"k8s.io/kops/util/pkg/vfs"
)

type TerraformTarget struct {
	terraformWriter.TerraformWriter
	Cloud   fi.Cloud
	Project string

	ClusterName string

	outDir string
	// extra config to add to the provider block
	clusterSpecTarget *kops.TargetSpec
	filesProvider     *vfs.TerraformProvider
}

func NewTerraformTarget(cloud fi.Cloud, project string, filesProvider *vfs.TerraformProvider, outDir string, clusterSpecTarget *kops.TargetSpec) *TerraformTarget {
	target := TerraformTarget{
		Cloud:   cloud,
		Project: project,

		outDir:            outDir,
		clusterSpecTarget: clusterSpecTarget,
		filesProvider:     filesProvider,
	}
	target.InitTerraformWriter()
	return &target
}

var _ fi.Target = &TerraformTarget{}

func (t *TerraformTarget) AddFileResource(ctx context.Context, resourceType string, resourceName string, key string, r fi.Resource, base64 bool) (*terraformWriter.Literal, error) {
	d, err := fi.ResourceAsBytes(ctx, r)
	if err != nil {
		id := resourceType + "_" + resourceName + "_" + key
		return nil, fmt.Errorf("error rending resource %s %v", id, err)
	}

	return t.AddFileBytes(resourceType, resourceName, key, d, base64)
}

func (t *TerraformTarget) ProcessDeletions() bool {
	// Terraform tracks & performs deletions itself
	return false
}

// tfGetProviderExtraConfig is a helper function to get extra config with safety checks on the pointers.
func tfGetProviderExtraConfig(c *kops.TargetSpec) map[string]string {
	if c != nil &&
		c.Terraform != nil &&
		c.Terraform.FilesProviderExtraConfig != nil {
		return *c.Terraform.ProviderExtraConfig
	}
	return nil
}

// tfGetFilesProviderExtraConfig is a helper function to get extra config with safety checks on the pointers.
func tfGetFilesProviderExtraConfig(c *kops.TargetSpec) map[string]string {
	if c != nil &&
		c.Terraform != nil &&
		c.Terraform.FilesProviderExtraConfig != nil {
		return *c.Terraform.FilesProviderExtraConfig
	}
	return nil
}

func (t *TerraformTarget) Finish(ctx context.Context, taskMap map[string]fi.Task) error {
	if err := t.finishHCL2(); err != nil {
		return err
	}

	for relativePath, contents := range t.Files {
		p := path.Join(t.outDir, relativePath)

		err := os.MkdirAll(path.Dir(p), os.FileMode(0o755))
		if err != nil {
			return fmt.Errorf("error creating output directory %q: %v", path.Dir(p), err)
		}

		err = os.WriteFile(p, contents, os.FileMode(0o644))
		if err != nil {
			return fmt.Errorf("error writing terraform data to output file %q: %v", p, err)
		}
	}
	klog.Infof("Terraform output is in %s", t.outDir)

	return nil
}
