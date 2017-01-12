package model

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// SSHKeyModelBuilder configures SSH objects
type SSHKeyModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &SSHKeyModelBuilder{}

func (b *SSHKeyModelBuilder) Build(c *fi.ModelBuilderContext) error {
	name, err := b.SSHKeyName()
	if err != nil {
		return err
	}
	t := &awstasks.SSHKey{
		Name:      s(name),
		PublicKey: fi.WrapResource(fi.NewStringResource(string(b.SSHPublicKeys[0]))),
	}
	c.AddTask(t)

	return nil
}
