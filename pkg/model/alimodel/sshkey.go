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

package alimodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/alitasks"
)

// SSHKeyModelBuilder configures SSH objects
type SSHKeyModelBuilder struct {
	*ALIModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &SSHKeyModelBuilder{}

func (b *SSHKeyModelBuilder) Build(c *fi.ModelBuilderContext) error {
	name := b.GetNameForSSHKey()
	t := &alitasks.SSHKey{
		Name:      s(name),
		Lifecycle: b.Lifecycle,
		PublicKey: fi.WrapResource(fi.NewStringResource(string(b.SSHPublicKeys[0]))),
	}
	c.AddTask(t)

	return nil
}
