package cloudup

import (
	"fmt"

	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
)

type BootstrapChannelBuilder struct {
	cluster *kops.Cluster
}

var _ TaskBuilder = &BootstrapChannelBuilder{}

func (b *BootstrapChannelBuilder) BuildTasks(l *Loader) error {
	addons, manifests := b.buildManifest()
	addonsYAML, err := utils.YamlMarshal(addons)
	if err != nil {
		return fmt.Errorf("error serializing addons yaml: %v", err)
	}

	name := b.cluster.Name + "-addons-bootstrap"

	l.tasks[name] = &fitasks.ManagedFile{
		Name:     fi.String(name),
		Location: fi.String("addons/bootstrap-channel.yaml"),
		Contents: fi.WrapResource(fi.NewBytesResource(addonsYAML)),
	}

	for key, resource := range manifests {
		name := b.cluster.Name + "-addons-" + key
		l.tasks[name] = &fitasks.ManagedFile{
			Name:     fi.String(name),
			Location: fi.String(resource),
			Contents: &fi.ResourceHolder{Name: resource},
		}
	}

	return nil
}

func (b *BootstrapChannelBuilder) buildManifest() (*channelsapi.Addons, map[string]string) {
	manifests := make(map[string]string)

	addons := &channelsapi.Addons{}
	addons.Kind = "Addons"
	addons.Name = "bootstrap"

	addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
		Name:     fi.String("kube-dns"),
		Version:  fi.String("1.4.0"),
		Selector: map[string]string{"k8s-addon": "kube-dns.addons.k8s.io"},
		Manifest: fi.String("kube-dns/v1.4.0.yaml"),
	})

	addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
		Name:     fi.String("core"),
		Version:  fi.String("1.4.0"),
		Selector: map[string]string{"k8s-addon": "core.addons.k8s.io"},
		Manifest: fi.String("core/v1.4.0.yaml"),
	})

	addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
		Name:     fi.String("dns-controller"),
		Version:  fi.String("1.4.1"),
		Selector: map[string]string{"k8s-addon": "dns-controller.addons.k8s.io"},
		Manifest: fi.String("dns-controller/v1.4.1.yaml"),
	})

	if b.cluster.Spec.Networking.Kopeio != nil {
		key := "networking.kope.io"
		version := "1.0.20161116"

		// TODO: Create configuration object for cni providers (maybe create it but orphan it)?
		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})

		manifests[key] = "addons/" + location
	}

	if b.cluster.Spec.Networking.Weave != nil {
		key := "networking.weave"
		version := "1.8.0.20161116"

		// TODO: Create configuration object for cni providers (maybe create it but orphan it)?
		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})

		manifests[key] = "addons/" + location
	}

	return addons, manifests
}
