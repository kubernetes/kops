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

package components

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	kopsmodel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/k8sversion"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"

	"github.com/blang/semver/v4"
	"k8s.io/klog/v2"
)

// OptionsContext is the context object for options builders
type OptionsContext struct {
	ClusterName string

	// Deprecated: Prefer using NodeKubernetesVersion() and ControlPlaneKubernetesVersion()
	KubernetesVersion semver.Version

	AssetBuilder *assets.AssetBuilder

	nodeKubernetesVersion         kopsmodel.KubernetesVersion
	controlPlaneKubernetesVersion kopsmodel.KubernetesVersion
}

func NewOptionsContext(cluster *kops.Cluster, assetBuilder *assets.AssetBuilder, maxKubeletSupportedVersion string) (*OptionsContext, error) {
	optionsContext := &OptionsContext{
		ClusterName:  cluster.ObjectMeta.Name,
		AssetBuilder: assetBuilder,
	}

	sv, err := util.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to determine kubernetes version from %q", cluster.Spec.KubernetesVersion)
	}
	optionsContext.KubernetesVersion = *sv

	controlPlaneKubernetesVersion, err := kopsmodel.ParseKubernetesVersion(cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to determine kubernetes version from %q: %w", cluster.Spec.KubernetesVersion, err)
	}
	nodeKubernetesVersion := controlPlaneKubernetesVersion
	if maxKubeletSupportedVersion != "" {
		nodeKubernetesVersion, err = kopsmodel.ParseKubernetesVersion(maxKubeletSupportedVersion)
		if err != nil {
			return nil, fmt.Errorf("unable to determine kubernetes version from %q: %w", maxKubeletSupportedVersion, err)
		}
	}

	optionsContext.nodeKubernetesVersion = *nodeKubernetesVersion
	optionsContext.controlPlaneKubernetesVersion = *controlPlaneKubernetesVersion

	return optionsContext, nil
}

func (c *OptionsContext) NodeKubernetesVersion() kopsmodel.KubernetesVersion {
	return c.nodeKubernetesVersion
}

func (c *OptionsContext) ControlPlaneKubernetesVersion() kopsmodel.KubernetesVersion {
	return c.controlPlaneKubernetesVersion
}

// Deprecated: prefer using NodeKubernetesVersion() and ControlPlaneKubernetesVersion()
func (c *OptionsContext) IsKubernetesGTE(version string) bool {
	return util.IsKubernetesGTE(version, c.KubernetesVersion)
}

// Deprecated: prefer using NodeKubernetesVersion() and ControlPlaneKubernetesVersion()
func (c *OptionsContext) IsKubernetesLT(version string) bool {
	return !c.IsKubernetesGTE(version)
}

// UsesCNI returns true if the networking provider is a CNI plugin
func UsesCNI(networking *kops.NetworkingSpec) bool {
	// Kubenet and CNI are the only kubelet networking plugins right now.
	return !networking.UsesKubenet()
}

func WellKnownServiceIP(networkingSpec *kops.NetworkingSpec, id int) (net.IP, error) {
	_, cidr, err := net.ParseCIDR(networkingSpec.ServiceClusterIPRange)
	if err != nil {
		return nil, fmt.Errorf("error parsing ServiceClusterIPRange %q: %v", networkingSpec.ServiceClusterIPRange, err)
	}

	ip4 := cidr.IP.To4()
	if ip4 != nil {
		n := binary.BigEndian.Uint32(ip4)
		n += uint32(id)
		serviceIP := make(net.IP, len(ip4))
		binary.BigEndian.PutUint32(serviceIP, n)
		return serviceIP, nil
	}

	ip6 := cidr.IP.To16()
	if ip6 != nil {
		baseIPInt := big.NewInt(0)
		baseIPInt.SetBytes(ip6)
		serviceIPInt := big.NewInt(0)
		serviceIPInt.Add(big.NewInt(int64(id)), baseIPInt)
		serviceIP := make(net.IP, len(ip6))
		serviceIPBytes := serviceIPInt.Bytes()
		for i := range serviceIPBytes {
			serviceIP[len(serviceIP)-len(serviceIPBytes)+i] = serviceIPBytes[i]
		}
		return serviceIP, nil
	}

	return nil, fmt.Errorf("unexpected IP address type for ServiceClusterIPRange: %s", networkingSpec.ServiceClusterIPRange)
}

// Image returns the docker image name for the specified component
func Image(component string, clusterSpec *kops.ClusterSpec, assetsBuilder *assets.AssetBuilder) (string, error) {
	if assetsBuilder == nil {
		return "", fmt.Errorf("unable to parse assets as assetBuilder is not defined")
	}

	kubernetesVersion, err := k8sversion.Parse(clusterSpec.KubernetesVersion)
	if err != nil {
		return "", err
	}

	imageName := component

	if !kopsmodel.IsBaseURL(clusterSpec.KubernetesVersion) {
		image := "registry.k8s.io/" + imageName + ":" + "v" + kubernetesVersion.String()

		return assetsBuilder.RemapImage(image), nil
	}

	// The simple name is valid when pulling.  But if we
	// are loading from a tarfile then the image is tagged with
	// the architecture suffix.
	//
	// i.e. registry.k8s.io/kube-apiserver:v1.20.0 is a manifest list
	// and we _can_ also pull
	// registry.k8s.io/kube-apiserver-amd64:v1.20.0 directly.  But if
	// we load https://.../v1.20.0/amd64/kube-apiserver.tar then
	// the image inside that tar file is named
	// "registry.k8s.io/kube-apiserver-amd64:v1.20.0"
	imageName += "-amd64"

	baseURL := clusterSpec.KubernetesVersion
	baseURL = strings.TrimSuffix(baseURL, "/")

	tagURL := baseURL + "/bin/linux/amd64/" + component + ".docker_tag"
	klog.V(2).Infof("Downloading docker tag for %s from: %s", component, tagURL)

	b, err := vfs.Context.ReadFile(tagURL)
	if err != nil {
		return "", fmt.Errorf("error reading tag file %q: %v", tagURL, err)
	}
	tag := strings.TrimSpace(string(b))
	klog.V(2).Infof("Found tag %q for %q", tag, component)

	image := "registry.k8s.io/" + imageName + ":" + tag

	return image, nil
}

// IsCertManagerEnabled returns true if the cluster has the capability to handle cert-manager PKI
func IsCertManagerEnabled(cluster *kops.Cluster) bool {
	return cluster.Spec.CertManager != nil && fi.ValueOf(cluster.Spec.CertManager.Enabled)
}
