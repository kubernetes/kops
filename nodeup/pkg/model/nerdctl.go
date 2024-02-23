package model

import (
	"fmt"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

type NerdctlBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &NerdctlBuilder{}

func (b *NerdctlBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.skipInstall() {
		klog.Info("SkipInstall is set to true; won't install nerdctl")
		return nil
	}

	assetName := "nerdctl"
	assetPath := ""
	asset, err := b.Assets.Find(assetName, assetPath)
	if err != nil {
		return fmt.Errorf("unable to locate asset %q", assetName)
	}

	c.AddTask(&nodetasks.File{
		Path:     b.nerdctlPath(),
		Contents: asset,
		Type:     nodetasks.FileType_File,
		Mode:     s("0755"),
	})

	return nil
}

func (b *NerdctlBuilder) binaryPath() string {
	path := "/usr/local/bin"
	if b.Distribution == distributions.DistributionFlatcar {
		path = "/opt/kops/bin"
	}
	if b.Distribution == distributions.DistributionContainerOS {
		path = "/home/kubernetes/bin"
	}
	return path

}

func (b *NerdctlBuilder) nerdctlPath() string {
	return b.binaryPath() + "/nerdctl"
}

func (b *NerdctlBuilder) skipInstall() bool {
	d := b.NodeupConfig.ContainerdConfig

	if d == nil {
		return false
	}

	return d.SkipInstall
}
