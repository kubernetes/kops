/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaymodel

import (
	"fmt"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scalewaytasks"
)

// SSHKeyModelBuilder configures SSH objects
type SSHKeyModelBuilder struct {
	*ScwModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &SSHKeyModelBuilder{}

func (b *SSHKeyModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	name, err := b.SSHKeyName()
	if err != nil {
		return fmt.Errorf("error building ssh key task: %w", err)
	}
	sshKeyResource := fi.Resource(fi.NewStringResource(string(b.SSHPublicKeys[0])))

	t := &scalewaytasks.SSHKey{
		Name:      fi.PtrTo(name),
		Lifecycle: b.Lifecycle,
		PublicKey: &sshKeyResource,
	}
	c.AddTask(t)

	return nil
}
