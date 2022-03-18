/*
Copyright 2021 The Kubernetes Authors.

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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// NodeProblemDetectorOptionsBuilder adds options for the node problem detector to the model.
type NodeProblemDetectorOptionsBuilder struct {
	*OptionsContext
}

var _ loader.OptionsBuilder = &NodeProblemDetectorOptionsBuilder{}

func (b *NodeProblemDetectorOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	if clusterSpec.NodeProblemDetector == nil {
		return nil
	}
	npd := clusterSpec.NodeProblemDetector

	if npd.Enabled == nil {
		npd.Enabled = fi.Bool(false)
	}

	if npd.CPURequest == nil {
		defaultCPURequest := resource.MustParse("10m")
		npd.CPURequest = &defaultCPURequest
	}

	if npd.MemoryRequest == nil {
		defaultMemoryRequest := resource.MustParse("80Mi")
		npd.MemoryRequest = &defaultMemoryRequest
	}

	if npd.CPULimit == nil {
		defaultCPULimit := resource.MustParse("10m")
		npd.CPULimit = &defaultCPULimit
	}

	if npd.MemoryLimit == nil {
		defaultMemoryLimit := resource.MustParse("80Mi")
		npd.MemoryLimit = &defaultMemoryLimit
	}

	if npd.Image == nil {
		npd.Image = fi.String("registry.k8s.io/node-problem-detector/node-problem-detector:v0.8.8")
	}

	return nil
}
