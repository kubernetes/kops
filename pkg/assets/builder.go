package assets

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/pkg/kubemanifest"
	"os"
	"strings"
)

// AssetBuilder discovers and remaps assets
type AssetBuilder struct {
}

func NewAssetBuilder() *AssetBuilder {
	return &AssetBuilder{}
}

func (a *AssetBuilder) RemapManifest(data []byte) ([]byte, error) {
	manifests, err := kubemanifest.LoadManifestsFrom(data)
	if err != nil {
		return nil, err
	}

	for _, manifest := range manifests {
		err := manifest.RemapImages(a.remapImage)
		if err != nil {
			return nil, fmt.Errorf("error remapping images: %v", err)
		}
		y, err := manifest.ToYAML()
		if err != nil {
			return nil, fmt.Errorf("error re-marshalling manifest: %v", err)
		}

		glog.Infof("manifest: %v", string(y))
	}

	return data, nil
}

func (a *AssetBuilder) remapImage(image string) (string, error) {
	if strings.HasPrefix(image, "kope/dns-controller:") {
		// To use user-defined DNS Controller:
		// 1. DOCKER_REGISTRY=[your docker hub repo] make dns-controller-push
		// 2. export DNSCONTROLLER_IMAGE=[your docker hub repo]
		// 3. make kops and create/apply cluster
		override := os.Getenv("DNSCONTROLLER_IMAGE")
		if override != "" {
			return override, nil
		}
	}

	return image, nil
}
