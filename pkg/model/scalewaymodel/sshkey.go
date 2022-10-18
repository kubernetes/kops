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

var _ fi.ModelBuilder = &SSHKeyModelBuilder{}

func (b *SSHKeyModelBuilder) Build(c *fi.ModelBuilderContext) error {
	name, err := b.SSHKeyName()
	if err != nil {
		return fmt.Errorf("error building ssh key task: %w", err)
	}
	sshKeyResource := fi.Resource(fi.NewStringResource(string(b.SSHPublicKeys[0])))

	t := &scalewaytasks.SSHKey{
		Name:      fi.String(name),
		Lifecycle: b.Lifecycle,
		PublicKey: &sshKeyResource,
	}
	c.AddTask(t)

	return nil
}
