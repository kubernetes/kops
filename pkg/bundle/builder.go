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

package bundle

import (
	"archive/tar"
	"fmt"
	"path"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
)

// Builder builds a bundle
type Builder struct {
	Clientset simple.Clientset
}

type Data struct {
	Files []*DataFile
}

type DataFile struct {
	Header tar.Header
	Data   []byte
}

func (b *Builder) Build(cluster *kops.Cluster, ig *kops.InstanceGroup) (*Data, error) {
	klog.Infof("building bundle for %q", ig.Name)
	keyStore, err := b.Clientset.KeyStore(cluster)
	if err != nil {
		return nil, err
	}

	fullCluster := &kops.Cluster{}
	{
		configBase, err := b.Clientset.ConfigBaseFor(cluster)
		if err != nil {
			return nil, fmt.Errorf("error building ConfigBase for cluster: %v", err)
		}

		p := configBase.Join(registry.PathClusterCompleted)

		b, err := p.ReadFile()
		if err != nil {
			return nil, fmt.Errorf("error loading Cluster %q: %v", p, err)
		}

		err = utils.YamlUnmarshal(b, fullCluster)
		if err != nil {
			return nil, fmt.Errorf("error parsing Cluster %q: %v", p, err)
		}
	}

	klog.Infof("fullCluster %v", fullCluster)

	fullCluster.Spec.ConfigBase = "/etc/kubernetes/bootstrap"
	fullCluster.Spec.ConfigStore = "/etc/kubernetes/bootstrap"
	fullCluster.Spec.KeyStore = "/etc/kubernetes/bootstrap/pki"
	fullCluster.Spec.SecretStore = "/etc/kubernetes/bootstrap/secrets"

	var files []*DataFile

	{
		data, err := utils.YamlMarshal(fullCluster)
		if err != nil {
			return nil, fmt.Errorf("error marshaling configuration: %v", err)
		}

		file := &DataFile{}
		file.Header.Name = "cluster.spec"
		file.Header.Size = int64(len(data))
		file.Header.Mode = 0644
		file.Data = data
		files = append(files, file)
	}

	{
		data, err := kopscodecs.ToVersionedYaml(ig)
		if err != nil {
			return nil, fmt.Errorf("error encoding instancegroup: %v", err)
		}

		file := &DataFile{}
		file.Header.Name = "instancegroup/" + ig.Name
		file.Header.Size = int64(len(data))
		file.Header.Mode = 0644
		file.Data = data
		files = append(files, file)
	}

	if pkiFiles, err := b.buildPKIFiles(cluster, ig, keyStore); err != nil {
		return nil, err
	} else {
		klog.Infof("pki files %v", pkiFiles)
		files = append(files, pkiFiles...)
	}

	copyManifest := make(map[string]string)

	{
		phase := cloudup.PhaseCluster
		assetBuilder := assets.NewAssetBuilder(cluster, string(phase))

		applyCmd := &cloudup.ApplyClusterCmd{
			Cluster:        cluster,
			Clientset:      b.Clientset,
			InstanceGroups: []*kops.InstanceGroup{ig},
			Phase:          phase,
		}

		if err := applyCmd.AddFileAssets(assetBuilder); err != nil {
			return nil, fmt.Errorf("error adding assets: %v", err)
		}

		nodeupConfig, err := applyCmd.BuildNodeUpConfig(assetBuilder, ig)
		if err != nil {
			return nil, fmt.Errorf("error building nodeup config: %v", err)
		}

		if !ig.IsMaster() {
			nodeupConfig.ProtokubeImage = nil
			nodeupConfig.Channels = nil
		}

		nodeupConfig.ConfigBase = fi.String("/etc/kubernetes/bootstrap")

		{
			var localChannels []string
			for _, channel := range nodeupConfig.Channels {
				base := path.Base(channel)
				localChannel := "file://" + path.Join("/rootfs/etc/kubernetes/bootstrap/addons/", base)
				localChannels = append(localChannels, localChannel)
				copyManifest[channel] = "addons/" + base
			}
			nodeupConfig.Channels = localChannels
		}

		// Temp hack:
		if ig.IsMaster() {
			if nodeupConfig.ProtokubeImage == nil {
				nodeupConfig.ProtokubeImage = &nodeup.Image{}
			}
			nodeupConfig.ProtokubeImage.Name = "justinsb/protokube:latest"
		}

		bootstrapScript := model.BootstrapScript{}

		{
			asset, err := cloudup.NodeUpAsset(assetBuilder)
			if err != nil {
				return nil, err
			}

			bootstrapScript.NodeUpSource = strings.Join(asset.Locations, ",")
			bootstrapScript.NodeUpSourceHash = asset.Hash.Hex()
		}

		bootstrapScript.NodeUpConfigBuilder = func(ig *kops.InstanceGroup) (*nodeup.Config, error) {
			return nodeupConfig, err
		}

		script, err := bootstrapScript.ResourceNodeUp(ig, cluster)
		if err != nil {
			return nil, fmt.Errorf("error building bootstrap script: %v", err)
		}

		scriptBytes, err := script.AsBytes()
		if err != nil {
			return nil, fmt.Errorf("error building bootstrap script: %v", err)
		}

		file := &DataFile{}
		file.Header.Name = "bootstrap.sh"
		file.Header.Size = int64(len(scriptBytes))
		file.Header.Mode = 0755
		file.Data = scriptBytes
		files = append(files, file)
	}

	klog.Infof("copyManifest %v", copyManifest)

	for src, dest := range copyManifest {
		data, err := vfs.Context.ReadFile(src)
		if err != nil {
			return nil, fmt.Errorf("error reading file %q: %v", src, err)
		}

		file := &DataFile{}
		file.Header.Name = dest
		file.Header.Size = int64(len(data))
		file.Header.Mode = 0644
		file.Data = data
		files = append(files, file)
	}

	return &Data{
		Files: files,
	}, nil
	//bundlePath := "output.tar.gz"
	//if err := writeToTar(files, bundlePath); err != nil {
	//	return err
	//}
}

func (b *Builder) buildPKIFiles(cluster *kops.Cluster, ig *kops.InstanceGroup, keyStore fi.CAStore) ([]*DataFile, error) {
	var files []*DataFile

	certs := []string{fi.CertificateId_CA, "kubelet"}
	keys := []string{"kubelet"}

	// Used by kube-proxy to auth to API
	certs = append(certs, "kube-proxy")
	keys = append(keys, "kube-proxy")

	if ig.IsMaster() {
		// Used by e.g. protokube
		certs = append(certs, "kops")
		keys = append(keys, "kops")

		// Used by apiserver-aggregator
		certs = append(certs, "apiserver-aggregator")
		keys = append(keys, "apiserver-aggregator")
		certs = append(certs, "apiserver-aggregator-ca")

		certs = append(certs, "apiserver-proxy-client")
		keys = append(keys, "apiserver-proxy-client")

		// Used by k-c-m, for example
		//certs = append(certs, "ca")
		keys = append(keys, "ca")

		// Used by kube-controller-manager to auth to API
		certs = append(certs, "kube-controller-manager")
		keys = append(keys, "kube-controller-manager")

		// Used by kube-scheduler to auth to API
		certs = append(certs, "kube-scheduler")
		keys = append(keys, "kube-scheduler")

		// key for the apiserver
		certs = append(certs, "master")
		keys = append(keys, "master")

		// We store kubecfg on the master
		certs = append(certs, "kubecfg")
		keys = append(keys, "kubecfg")
	}

	for _, name := range certs {
		certPool, err := keyStore.FindCertificateKeyset(name)
		if err != nil {
			return nil, fmt.Errorf("error querying certificate %q: %v", name, err)
		}
		if certPool == nil {
			return nil, fmt.Errorf("certificate %q not found", name)
		}

		data, err := fi.SerializeKeyset(certPool)
		if err != nil {
			return nil, fmt.Errorf("error serializing certificate %q: %v", name, err)
		}

		file := &DataFile{}
		file.Header.Name = "pki/issued/" + name + "/keyset.yaml"
		file.Header.Size = int64(len(data))
		file.Header.Mode = 0644
		file.Data = data
		files = append(files, file)
	}

	for _, name := range keys {
		key, err := keyStore.FindPrivateKeyset(name)
		if err != nil {
			return nil, fmt.Errorf("error querying private key %q: %v", name, err)
		}
		if key == nil {
			return nil, fmt.Errorf("private key %q not found", name)
		}

		data, err := fi.SerializeKeyset(key)
		if err != nil {
			return nil, fmt.Errorf("error serializing private key %q: %v", name, err)
		}

		file := &DataFile{}
		file.Header.Name = "pki/private/" + name + "/keyset.yaml"
		file.Header.Size = int64(len(data))
		file.Header.Mode = 0644
		file.Data = data
		files = append(files, file)
	}

	return files, nil
}
