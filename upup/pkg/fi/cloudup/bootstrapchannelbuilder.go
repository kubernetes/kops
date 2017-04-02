/*
Copyright 2016 The Kubernetes Authors.

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

package cloudup

import (
	"fmt"

	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
	kube_api "k8s.io/kubernetes/pkg/api"
	kube_api_ext "k8s.io/kubernetes/pkg/apis/extensions"
)

type BootstrapChannelBuilder struct {
	cluster *kops.Cluster
}

var _ fi.ModelBuilder = &BootstrapChannelBuilder{}

func (b *BootstrapChannelBuilder) Build(c *fi.ModelBuilderContext) error {
	addons, manifests, err := b.buildManifest()
	if err != nil {
		return err
	}

	addonsYAML, err := utils.YamlMarshal(addons)

	if err != nil {
		return fmt.Errorf("error serializing addons yaml: %v", err)
	}

	name := b.cluster.ObjectMeta.Name + "-addons-bootstrap"

	tasks := c.Tasks

	tasks[name] = &fitasks.ManagedFile{
		Name:     fi.String(name),
		Location: fi.String("addons/bootstrap-channel.yaml"),
		Contents: fi.WrapResource(fi.NewBytesResource(addonsYAML)),
	}

	for key, manifest := range manifests {
		data := func(addons *channelsapi.Addons, key string) string {
			for _, addon := range addons.Spec.Addons {
				if *addon.Name == key {
					if addon.Yamldata != nil {
						return *addon.Yamldata
					}
				}
			}
			return ""
		}

		d := data(addons, key)
		name := b.cluster.ObjectMeta.Name + "-addons-" + key
		tasks[name] = &fitasks.ManagedFile{
			Name:     fi.String(name),
			Location: fi.String(manifest),
			Contents: &fi.ResourceHolder{
				Name:     manifest,
				Resource: fi.NewStringResource(d),
			},
		}
	}

	return nil
}

func (b *BootstrapChannelBuilder) buildManifest() (*channelsapi.Addons, map[string]string, error) {
	manifests := make(map[string]string)

	addons := &channelsapi.Addons{}
	addons.Kind = "Addons"
	addons.ObjectMeta.Name = "bootstrap"

	kv, err := util.ParseKubernetesVersion(b.cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to determine kubernetes version from %q", b.cluster.Spec.KubernetesVersion)
	}

	{
		key := "core.addons.k8s.io"
		version := "1.4.0"

		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
		manifests[key] = "addons/" + location
	}

	{
		key := "kube-dns.addons.k8s.io"

		var version string
		var location string
		switch {
		case kv.Major == 1 && kv.Minor <= 5:
			version = "1.5.1"
			location = key + "/k8s-1.5.yaml"
		default:
			version = "1.6.0"
			location = key + "/k8s-1.6.yaml"
		}

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
		manifests[key] = "addons/" + location
	}

	{
		key := "limit-range.addons.k8s.io"
		version := "1.5.0"

		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
		manifests[key] = "addons/" + location
	}

	{
		key := "dns-controller.addons.k8s.io"

		var version string
		var location string
		switch {
		case kv.Major == 1 && kv.Minor <= 5:
			// This is awkward... we would like to do version 1.6.0,
			// but if we do then we won't get the new manifest when we upgrade to 1.6.0
			version = "1.5.3"
			location = key + "/k8s-1.5.yaml"
		default:
			version = "1.6.0"
			location = key + "/k8s-1.6.yaml"
		}

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
		manifests[key] = "addons/" + location
	}

	{
		key := "storage-aws.addons.k8s.io"
		version := "1.6.0"

		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
		manifests[key] = "addons/" + location
	}

	// The role.kubernetes.io/networking is used to label anything related to a networking addin,
	// so that if we switch networking plugins (e.g. calico -> weave or vice-versa), we'll replace the
	// old networking plugin, and there won't be old pods "floating around".

	// This means whenever we create or update a networking plugin, we should be sure that:
	// 1. the selector is role.kubernetes.io/networking=1
	// 2. every object in the manifest is labeleled with role.kubernetes.io/networking=1

	// TODO: Some way to test/enforce this?

	// TODO: Create "empty" configurations for others, so we can delete e.g. the kopeio configuration
	// if we switch to kubenet?

	if b.cluster.Spec.Networking.Kopeio != nil {
		key := "networking.kope.io"
		version := "1.0.20161116"

		// TODO: Create configuration object for cni providers (maybe create it but orphan it)?
		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"role.kubernetes.io/networking": "1"},
			Manifest: fi.String(location),
		})

		manifests[key] = "addons/" + location
	}

	if b.cluster.Spec.Networking.Weave != nil {
		key := "networking.weave"
		version := "1.9.4"

		// TODO: Create configuration object for cni providers (maybe create it but orphan it)?

		location := key + "/v" + version + ".yaml"

		fmt.Printf("DEBUG location %v\n\n", location)
		fmt.Printf("DEBUG key %v\n\n", key)

		if b.cluster.Spec.Networking.Weave.Encrypt {
			fmt.Printf("DEBUG encrypted %v\n\n", b.cluster.Spec.Networking.Weave.Encrypt)
			secret, err := fi.CreateSecret()
			secData := make(map[string][]byte)
			secData["weave-pass"] = []byte(secret.Data)
			seConfig := kube_api.Secret{
				Data: secData,
				Type: kube_api.SecretTypeOpaque,
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "weave-pass",
					Namespace: "kube-system"}}
			gv := v1.SchemeGroupVersion
			info, _ := runtime.SerializerInfoForMediaType(kube_api.Codecs.SupportedMediaTypes(), "application/yaml")

			// FIXME require split objects in yaml
			encoder := kube_api.Codecs.EncoderForVersion(info.Serializer, gv)
			secretData, err := runtime.Encode(encoder, &seConfig)
			if err != nil {
				fmt.Errorf("error marshaling secret yaml: %s", err)
				panic(err)
			}

			// FIXME: Remove dump for debuug
			fmt.Printf("--- t dump:\n%s\n\n", string(secretData))
			prefix := "addons/"
			weaveLoc := prefix + key + "/secret.yaml"
			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key + "secret"),
				Version:  fi.String("0.0.1"),
				Selector: map[string]string{"role.kubernetes.io/networking": "1"},
				Manifest: fi.String(key + "/secret.yaml"),
				Yamldata: fi.String(string(secretData)),
			})

			manifests[key+"secret"] = weaveLoc

			// read weave yaml
			weave_file := "upup/models/cloudup/resources/addons/" + location
			weavesource, err := ioutil.ReadFile(weave_file)
			if err != nil {
				panic(err)
			}
			gv = v1beta1.SchemeGroupVersion
			encoder = kube_api.Codecs.EncoderForVersion(info.Serializer, gv)
			weaveconfigObj, err := runtime.Decode(kube_api.Codecs.UniversalDecoder(), weavesource)
			weaveconfig := weaveconfigObj.(*kube_api_ext.DaemonSet)
			if err != nil {
				panic(err)
			}
			// edit weave yaml
			newenv := []kube_api.EnvVar{kube_api.EnvVar{
				Name: "WEAVE_PASSWORD",
				ValueFrom: &kube_api.EnvVarSource{
					SecretKeyRef: &kube_api.SecretKeySelector{
						LocalObjectReference: kube_api.LocalObjectReference{
							Name: "weave-pass"},
						Key: "weave-pass"}}}}

			// assign to all container new env variable
			containers := make([]kube_api.Container, len(weaveconfig.Spec.Template.Spec.Containers))
			for i, cont := range weaveconfig.Spec.Template.Spec.Containers {
				cont.Env = newenv
				containers[i] = cont
			}
			weaveconfig.Spec.Template.Spec.Containers = containers
			weaveData, err := runtime.Encode(encoder, weaveconfig)
			if err != nil {
				panic(err)
			}
			fmt.Printf("--- t dump:\n%s\n\n", string(weaveData))
			newLocation := prefix + key + "/weave.yaml"
			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Version:  fi.String(version),
				Selector: map[string]string{"role.kubernetes.io/networking": "1"},
				Manifest: fi.String(key + "/weave.yaml"),
				Yamldata: fi.String(string(weaveData)),
			})

			manifests[key] = newLocation
		} else {

			fmt.Printf("DEBUG NOT!!! encrypted %v\n\n", b.cluster.Spec.Networking.Weave.Encrypt)
			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Version:  fi.String(version),
				Selector: map[string]string{"role.kubernetes.io/networking": "1"},
				Manifest: fi.String(location),
			})

			manifests[key] = "addons/" + location
		}
	}

	if b.cluster.Spec.Networking.Flannel != nil {
		key := "networking.flannel"
		version := "0.7.0"

		// TODO: Create configuration object for cni providers (maybe create it but orphan it)?
		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"role.kubernetes.io/networking": "1"},
			Manifest: fi.String(location),
		})

		manifests[key] = "addons/" + location
	}

	if b.cluster.Spec.Networking.Calico != nil {
		key := "networking.projectcalico.org"
		version := "2.0.2"

		// TODO: Create configuration object for cni providers (maybe create it but orphan it)?
		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"role.kubernetes.io/networking": "1"},
			Manifest: fi.String(location),
		})

		manifests[key] = "addons/" + location
	}

	if b.cluster.Spec.Networking.Canal != nil {
		key := "networking.projectcalico.org.canal"
		version := "1.0"

		// TODO: Create configuration object for cni providers (maybe create it but orphan it)?
		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Version:  fi.String(version),
			Selector: map[string]string{"role.kubernetes.io/networking": "1"},
			Manifest: fi.String(location),
		})

		manifests[key] = "addons/" + location
	}

	return addons, manifests, nil
}
