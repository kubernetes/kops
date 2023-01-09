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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// AWSEBSCSIDriverOptionsBuilder adds options for the AWS EBS CSI driver to the model
type AWSEBSCSIDriverOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &AWSEBSCSIDriverOptionsBuilder{}

func (b *AWSEBSCSIDriverOptionsBuilder) BuildOptions(o interface{}) error {
	aws := o.(*kops.ClusterSpec).CloudProvider.AWS
	if aws == nil {
		return nil
	}

	if aws.EBSCSIDriver == nil {
		aws.EBSCSIDriver = &kops.EBSCSIDriverSpec{
			Enabled: fi.PtrTo(true),
		}
	}
	c := aws.EBSCSIDriver

	if !fi.ValueOf(c.Enabled) {
		return nil
	}

	if c.Version == nil {
		version := "v1.14.1"
		c.Version = &version
	}

	return nil
}
