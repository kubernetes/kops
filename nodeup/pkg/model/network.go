/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"path/filepath"
)

// NetworkBuilder writes CNI assets
type NetworkBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &NetworkBuilder{}

func (b *NetworkBuilder) Build(c *fi.ModelBuilderContext) error {
	var assetNames []string

	networking := b.Cluster.Spec.Networking
	if networking == nil || networking.Classic != nil {
	} else if networking.Kubenet != nil {
		assetNames = append(assetNames, "bridge", "host-local", "loopback")
	} else if networking.External != nil {
		// external is based on kubenet
		assetNames = append(assetNames, "bridge", "host-local", "loopback")
	} else if networking.CNI != nil || networking.Weave != nil || networking.Flannel != nil || networking.Calico != nil || networking.Canal != nil || networking.Kuberouter != nil || networking.Romana != nil {
		assetNames = append(assetNames, "bridge", "host-local", "loopback", "ptp")
		// Do we need tuning?

		// TODO: Only when using flannel ?
		assetNames = append(assetNames, "flannel")
	} else if networking.Kopeio != nil {
		// TODO combine with External
		// Kopeio is based on kubenet / external
		assetNames = append(assetNames, "bridge", "host-local", "loopback")
	} else {
		return fmt.Errorf("no networking mode set")
	}

	for _, assetName := range assetNames {
		if err := b.addCNIBinAsset(c, assetName); err != nil {
			return err
		}
	}

	return nil
}

func (b *NetworkBuilder) addCNIBinAsset(c *fi.ModelBuilderContext, assetName string) error {
	assetPath := ""
	asset, err := b.Assets.Find(assetName, assetPath)
	if err != nil {
		return fmt.Errorf("error trying to locate asset %q: %v", assetName, err)
	}
	if asset == nil {
		return fmt.Errorf("unable to locate asset %q", assetName)
	}

	t := &nodetasks.File{
		Path:     filepath.Join(b.CNIBinDir(), assetName),
		Contents: asset,
		Type:     nodetasks.FileType_File,
		Mode:     s("0755"),
	}
	c.AddTask(t)

	return nil
}
