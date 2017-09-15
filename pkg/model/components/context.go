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
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/util/pkg/vfs"

	"github.com/blang/semver"
	"github.com/golang/glog"
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
	} else if networking.External != nil {
		// external is based on kubenet
		return true, nil
	} else if networking.CNI != nil || networking.Weave != nil || networking.Flannel != nil || networking.Calico != nil || networking.Canal != nil || networking.Kuberouter != nil || networking.Romana != nil {
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

	return nil, fmt.Errorf("Unexpected IP address type for ServiceClusterIPRange: %s", clusterSpec.ServiceClusterIPRange)
}

func IsBaseURL(kubernetesVersion string) bool {
	return strings.HasPrefix(kubernetesVersion, "http:") || strings.HasPrefix(kubernetesVersion, "https:")
}

// Image returns the docker image name for the specified component
func Image(component string, clusterSpec *kops.ClusterSpec, assetsBuilder *assets.AssetBuilder) (string, error) {
	if assetsBuilder == nil {
		return "", fmt.Errorf("unable to parse assets as assetBuilder is not defined")
	}
	// TODO remove this, as it is an addon now
	if component == "kube-dns" {
		// TODO: Once we are shipping different versions, start to use them
		return "gcr.io/google_containers/kubedns-amd64:1.3", nil
	}

	if !IsBaseURL(clusterSpec.KubernetesVersion) {
		image := "gcr.io/google_containers/" + component + ":" + "v" + clusterSpec.KubernetesVersion

		image, err := assetsBuilder.RemapImage(image)
		if err != nil {
			return "", fmt.Errorf("unable to remap container %q: %v", image, err)
		}
		return image, nil
	}

	baseURL := clusterSpec.KubernetesVersion
	baseURL = strings.TrimSuffix(baseURL, "/")

	tagURL := baseURL + "/bin/linux/amd64/" + component + ".docker_tag"
	glog.V(2).Infof("Downloading docker tag for %s from: %s", component, tagURL)

	b, err := vfs.Context.ReadFile(tagURL)
	if err != nil {
		return "", fmt.Errorf("error reading tag file %q: %v", tagURL, err)
	}
	tag := strings.TrimSpace(string(b))
	glog.V(2).Infof("Found tag %q for %q", tag, component)

	return "gcr.io/google_containers/" + component + ":" + tag, nil
}

func GCETagForRole(clusterName string, role kops.InstanceGroupRole) string {
	return gce.SafeClusterName(clusterName) + "-" + gce.GceLabelNameRolePrefix + strings.ToLower(string(role))
}
