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
	"k8s.io/kops/upup/pkg/fi"
)

// AWSEBSCSIDriverBuilder writes AWSEBSCSIDriver's assets
type AWSEBSCSIDriverBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &AWSEBSCSIDriverBuilder{}

// Build is responsible for configuring the EBS CSI Driver stuff
func (b *AWSEBSCSIDriverBuilder) Build(c *fi.ModelBuilderContext) error {
	csi := b.Cluster.Spec.CloudConfig.AWSEBSCSIDriver

	if csi == nil || !fi.BoolValue(csi.Enabled) {
		return nil
	}

	// Pulling CSI driver image
	image := "k8s.gcr.io/provider-aws/aws-ebs-csi-driver:" + *csi.Version
	b.WarmPullImage(c, image)

	// Pulling CSI sidecars images
	sidecars := []string{
		"csi-node-driver-registrar:v2.1.0",
		"livenessprobe:v2.2.0",
	}
	for _, s := range sidecars {
		image = "k8s.gcr.io/sig-storage/" + s
		b.WarmPullImage(c, image)
	}

	return nil

}
