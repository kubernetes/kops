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
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
)

type SpecBuilder struct {
	OptionsLoader *loader.OptionsLoader

	Tags sets.String
}

func (l *SpecBuilder) BuildCompleteSpec(clusterSpec *api.ClusterSpec, modelStore vfs.Path, models []string) (*api.ClusterSpec, error) {
	// First pass over models: load options
	tw := &loader.TreeWalker{
		DefaultHandler: ignoreHandler,
		Contexts: map[string]loader.Handler{
			"resources": ignoreHandler,
		},
		Extensions: map[string]loader.Handler{
			".options": l.OptionsLoader.HandleOptions,
		},
		Tags: l.Tags,
	}
	for _, model := range models {
		modelDir := modelStore.Join(model)
		err := tw.Walk(modelDir)
		if err != nil {
			return nil, err
		}
	}

	loaded, err := l.OptionsLoader.Build(clusterSpec)
	if err != nil {
		return nil, err
	}
	completed := &api.ClusterSpec{}
	*completed = *(loaded.(*api.ClusterSpec))

	// Master kubelet config = (base kubelet config + master kubelet config)
	masterKubelet := &api.KubeletConfigSpec{}
	utils.JsonMergeStruct(masterKubelet, completed.Kubelet)
	utils.JsonMergeStruct(masterKubelet, completed.MasterKubelet)
	completed.MasterKubelet = masterKubelet

	glog.V(1).Infof("options: %s", fi.DebugAsJsonStringIndent(completed))
	return completed, nil
}
