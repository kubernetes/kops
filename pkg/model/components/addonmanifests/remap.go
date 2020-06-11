package addonmanifests

import (
	"fmt"

	"k8s.io/klog"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components/addonmanifests/kopscontroller"
)

func RemapAddonManifest(name string, context *model.KopsModelContext, assetBuilder *assets.AssetBuilder, manifest []byte) ([]byte, error) {
	{
		objects, err := kubemanifest.LoadObjectsFrom(manifest)
		if err != nil {
			return nil, err
		}

		if name == "kops-controller.addons.k8s.io" {
			if err := kopscontroller.Remap(context, objects); err != nil {
				return nil, err
			}
		}

		b, err := kubemanifest.ToYAML(objects)
		if err != nil {
			return nil, err
		}

		if name == "kops-controller.addons.k8s.io" {
			klog.Infof("remapped %s", string(b))
		}
		manifest = b
	}

	{
		remapped, err := assetBuilder.RemapManifest(manifest)
		if err != nil {
			klog.Infof("invalid manifest: %s", string(manifest))
			return nil, fmt.Errorf("error remapping manifest %s: %v", manifest, err)
		}
		manifest = remapped
	}

	return manifest, nil
}
