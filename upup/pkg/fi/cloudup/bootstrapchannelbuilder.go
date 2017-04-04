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
	"bytes"
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
		managedfile := &fitasks.ManagedFile{
			Name:     fi.String(name),
			Location: fi.String(manifest),
			Contents: &fi.ResourceHolder{
				Name: manifest,
			},
		}
		if d != "" {
			managedfile.Contents.Resource = fi.NewStringResource(d)
		}
		tasks[name] = managedfile
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
		var version string
		switch {
		case kv.Major == 1 && kv.Minor <= 5:
			version = "1.9.3"
		default:
			version = "1.9.4"
		}

		// TODO: Create configuration object for cni providers (maybe create it but orphan it)?

		location := key + "/v" + version + ".yaml"

		seConfig, _ := BuildSecret()
		if b.cluster.Spec.Networking.Weave.Encrypt {
			info, _ := runtime.SerializerInfoForMediaType(kube_api.Codecs.SupportedMediaTypes(), "application/yaml")

			encoder := kube_api.Codecs.EncoderForVersion(info.Serializer, v1.SchemeGroupVersion)
			secretData, err := runtime.Encode(encoder, &seConfig)
			if err != nil {
				return nil, nil, fmt.Errorf("error marshaling secret yaml: %s", err)
			}

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
			encoder = kube_api.Codecs.EncoderForVersion(info.Serializer, v1beta1.SchemeGroupVersion)
			delimiter := []byte("\n---\n")
			sections := bytes.Split(weavesource, delimiter)
			var newSections []byte
			for _, section := range sections {
				obj, err := runtime.Decode(kube_api.Codecs.UniversalDecoder(), section)
				if err != nil {
					return nil, nil, fmt.Errorf("error parsing file %s obj %s: %v", weavesource, string(section), err)
				}
				switch v := obj.(type) {
				case *kube_api_ext.DaemonSet:
					weaveconfig := BuildWeaveDaemonSet(obj)
					weaveData, err := runtime.Encode(encoder, &weaveconfig)
					if err != nil {
						return nil, nil, fmt.Errorf("error encode file %s obj %v: %v", weavesource, weaveconfig, err)
					}
					newSections = append(newSections[:], weaveData[:]...)
				default:
					fmt.Printf("not changed %s,\n%v", v, string(section))
					newSections = append(newSections[:], section[:]...)
				}
				newSections = append(newSections[:], delimiter[:]...)
			}

			newLocation := prefix + key + "/weave.yaml"
			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Version:  fi.String(version),
				Selector: map[string]string{"role.kubernetes.io/networking": "1"},
				Manifest: fi.String(key + "/weave.yaml"),
				Yamldata: fi.String(string(newSections)),
			})

			manifests[key] = newLocation
		} else {
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

func BuildSecret() (kube_api.Secret, error) {
	secret, err := fi.CreateSecret()
	if err != nil {
		return kube_api.Secret{}, fmt.Errorf("error create secret: %s", err)
	}
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
	return seConfig, err
}

func BuildNewEnv() []kube_api.EnvVar {
	newenv := []kube_api.EnvVar{{
		Name: "WEAVE_PASSWORD",
		ValueFrom: &kube_api.EnvVarSource{
			SecretKeyRef: &kube_api.SecretKeySelector{
				LocalObjectReference: kube_api.LocalObjectReference{
					Name: "weave-pass"},
				Key: "weave-pass"}}}}
	return newenv
}

// edit weave yaml
func BuildWeaveDaemonSet(obj runtime.Object) kube_api_ext.DaemonSet {
	// assign to all container new env variable
	weaveConfig := obj.(*kube_api_ext.DaemonSet)
	containers := make([]kube_api.Container, len(weaveConfig.Spec.Template.Spec.Containers))
	newenv := BuildNewEnv()
	for i, cont := range weaveConfig.Spec.Template.Spec.Containers {
		cont.Env = newenv
		containers[i] = cont
	}
	weaveConfig.Spec.Template.Spec.Containers = containers
	return *weaveConfig
}
