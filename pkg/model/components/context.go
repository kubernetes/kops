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
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/k8sversion"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/util/pkg/vfs"

	"github.com/blang/semver"
	"k8s.io/klog"
)

// OptionsContext is the context object for options builders
type OptionsContext struct {
	ClusterName string

	KubernetesVersion semver.Version

	AssetBuilder *assets.AssetBuilder
}

func (c *OptionsContext) IsKubernetesGTE(version string) bool {
	return util.IsKubernetesGTE(version, c.KubernetesVersion)
}

func (c *OptionsContext) IsKubernetesLT(version string) bool {
	return !c.IsKubernetesGTE(version)
}

// Architecture returns the architecture we are using
// We currently only support amd64, and we probably need to pass the InstanceGroup in
// But we can start collecting the architectural dependencies
func (c *OptionsContext) Architecture() string {
	return "amd64"
}

// KubernetesVersion parses the semver version of kubernetes, from the cluster spec
// Deprecated: prefer using OptionsContext.KubernetesVersion
func KubernetesVersion(clusterSpec *kops.ClusterSpec) (*semver.Version, error) {
	kubernetesVersion := clusterSpec.KubernetesVersion

	if kubernetesVersion == "" {
		return nil, fmt.Errorf("KubernetesVersion is required")
	}

	sv, err := util.ParseKubernetesVersion(kubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to determine kubernetes version from %q", kubernetesVersion)
	}

	return sv, nil
}

// UsesKubenet returns true if our networking is derived from kubenet
func UsesKubenet(clusterSpec *kops.ClusterSpec) (bool, error) {
	networking := clusterSpec.Networking
	if networking == nil || networking.Classic != nil {
		return false, nil
	} else if networking.Kubenet != nil {
		return true, nil
	} else if networking.GCE != nil {
		// GCE IP Alias networking is based on kubenet
		return true, nil
	} else if networking.External != nil {
		// external is based on kubenet
		return true, nil
	} else if networking.CNI != nil || networking.Weave != nil || networking.Flannel != nil || networking.Calico != nil || networking.Canal != nil || networking.Kuberouter != nil || networking.Romana != nil || networking.AmazonVPC != nil || networking.Cilium != nil || networking.LyftVPC != nil {
		return false, nil
	} else if networking.Kopeio != nil {
		// Kopeio is based on kubenet / external
		return true, nil
	} else {
		return false, fmt.Errorf("no networking mode set")
	}
}

func WellKnownServiceIP(clusterSpec *kops.ClusterSpec, id int) (net.IP, error) {
	_, cidr, err := net.ParseCIDR(clusterSpec.ServiceClusterIPRange)
	if err != nil {
		return nil, fmt.Errorf("error parsing ServiceClusterIPRange %q: %v", clusterSpec.ServiceClusterIPRange, err)
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

	return nil, fmt.Errorf("unexpected IP address type for ServiceClusterIPRange: %s", clusterSpec.ServiceClusterIPRange)
}

func IsBaseURL(kubernetesVersion string) bool {
	return strings.HasPrefix(kubernetesVersion, "http:") || strings.HasPrefix(kubernetesVersion, "https:") || strings.HasPrefix(kubernetesVersion, "memfs:")
}

// Image returns the docker image name for the specified component
func Image(component string, architecture string, clusterSpec *kops.ClusterSpec, assetsBuilder *assets.AssetBuilder) (string, error) {
	if assetsBuilder == nil {
		return "", fmt.Errorf("unable to parse assets as assetBuilder is not defined")
	}
	// TODO remove this, as it is an addon now
	if component == "kube-dns" {
		// TODO: Once we are shipping different versions, start to use them
		return "k8s.gcr.io/kubedns-amd64:1.3", nil
	}

	kubernetesVersion, err := k8sversion.Parse(clusterSpec.KubernetesVersion)
	if err != nil {
		return "", err
	}

	imageName := component

	if !IsBaseURL(clusterSpec.KubernetesVersion) {
		image := "k8s.gcr.io/" + imageName + ":" + "v" + kubernetesVersion.String()

		image, err := assetsBuilder.RemapImage(image)
		if err != nil {
			return "", fmt.Errorf("unable to remap container %q: %v", image, err)
		}
		return image, nil
	}

	// The simple name is valid when pulling (before 1.16 it was
	// only amd64, as of 1.16 it is a manifest list).  But if we
	// are loading from a tarfile then the image is tagged with
	// the architecture suffix.
	//
	// i.e. k8s.gcr.io/kube-apiserver:v1.16.0 is a manifest list
	// and we _can_ also pull
	// k8s.gcr.io/kube-apiserver-amd64:v1.16.0 directly.  But if
	// we load https://.../v1.16.0/amd64/kube-apiserver.tar then
	// the image inside that tar file is named
	// "k8s.gcr.io/kube-apiserver-amd64:v1.16.0"
	//
	// But ... this is only the case from 1.16 on...
	if kubernetesVersion.IsGTE("1.16") {
		imageName += "-" + architecture
	}

	baseURL := clusterSpec.KubernetesVersion
	baseURL = strings.TrimSuffix(baseURL, "/")

	// TODO path.Join here?
	tagURL := baseURL + "/bin/linux/" + architecture + "/" + component + ".docker_tag"
	klog.V(2).Infof("Downloading docker tag for %s from: %s", component, tagURL)

	b, err := vfs.Context.ReadFile(tagURL)
	if err != nil {
		return "", fmt.Errorf("error reading tag file %q: %v", tagURL, err)
	}
	tag := strings.TrimSpace(string(b))
	klog.V(2).Infof("Found tag %q for %q", tag, component)

	image := "k8s.gcr.io/" + imageName + ":" + tag

	// When we're using a docker load-ed image, we are likely a CI build.
	// But the k8s.gcr.io prefix is an alias, and we only double-tagged from 1.10 onwards.
	// For versions prior to 1.10, remap k8s.gcr.io to the old name.
	// This also means that we won't start using the aliased names on existing clusters,
	// which could otherwise be surprising to users.
	if !kubernetesVersion.IsGTE("1.10") {
		image = "gcr.io/google_containers/" + strings.TrimPrefix(image, "k8s.gcr.io/")
	}

	return image, nil
}

func GCETagForRole(clusterName string, role kops.InstanceGroupRole) string {
	return gce.SafeClusterName(clusterName) + "-" + gce.GceLabelNameRolePrefix + strings.ToLower(string(role))
}
