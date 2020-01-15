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

package model

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// SSHKeyModelBuilder configures SSH objects
type SSHKeyModelBuilder struct {
	*KopsModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &SSHKeyModelBuilder{}

func (b *SSHKeyModelBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.UseSSHKey() {
		return nil
	}

	name, err := b.SSHKeyName()
	if err != nil {
		return err
	}
	t := &awstasks.SSHKey{
		Name:      s(name),
		Lifecycle: b.Lifecycle,
	}
	if len(b.SSHPublicKeys) >= 1 {
		t.PublicKey = fi.WrapResource(fi.NewStringResource(string(b.SSHPublicKeys[0])))
	}
	c.AddTask(t)

	return nil
}
