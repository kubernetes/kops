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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeDnsOptionsBuilder adds options for kube-dns
type KubeDnsOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeDnsOptionsBuilder{}

// BuildOptions fills in the kubedns model
func (b *KubeDnsOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)

	if clusterSpec.KubeDNS == nil {
		clusterSpec.KubeDNS = &kops.KubeDNSConfig{}
	}

	clusterSpec.KubeDNS.Replicas = 2

	if clusterSpec.KubeDNS.CacheMaxSize == 0 {
		clusterSpec.KubeDNS.CacheMaxSize = 1000
	}

	if clusterSpec.KubeDNS.CacheMaxConcurrent == 0 {
		clusterSpec.KubeDNS.CacheMaxConcurrent = 150
	}

	if clusterSpec.KubeDNS.ServerIP == "" {
		ip, err := WellKnownServiceIP(clusterSpec, 10)
		if err != nil {
			return err
		}
		clusterSpec.KubeDNS.ServerIP = ip.String()
	}

	if clusterSpec.KubeDNS.Domain == "" {
		clusterSpec.KubeDNS.Domain = clusterSpec.ClusterDNSDomain
	}
	
	if clusterSpec.KubeDNS.MemoryRequest != "" {
		MemoryRequest, err := resource.ParseQuantity(clusterSpec.KubeDNS.MemoryRequest)
		if err != nil {
			return fmt.Errorf("Error parsing MemoryRequest=%q", clusterSpec.KubeDNS.MemoryRequest)
		}
		resourceLimits["cpu"] = MemoryRequest
	}else{
		clusterSpec.KubeDNS.MemoryRequest="70m"
	}
	
	if clusterSpec.KubeDNS.CPURequest != "" {
		CPURequest, err := resource.ParseQuantity(clusterSpec.KubeDNS.CPURequest)
		if err != nil {
			return fmt.Errorf("Error parsing CPURequest=%q", clusterSpec.KubeDNS.CPURequest)
		}
		resourceLimits["cpu"] = CPURequest
	}else{
		clusterSpec.KubeDNS.CPURequest="100m"
	}
	
	if clusterSpec.KubeDNS.MemoryLimit != "" {
		MemoryLimit, err := resource.ParseQuantity(clusterSpec.KubeDNS.MemoryLimit)
		if err != nil {
			return fmt.Errorf("Error parsing MemoryLimit=%q", clusterSpec.KubeDNS.MemoryLimit)
		}
		resourceLimits["cpu"] = MemoryLimit
	}else{
		clusterSpec.KubeDNS.MemoryLimit="170m"
	}

	return nil
}
