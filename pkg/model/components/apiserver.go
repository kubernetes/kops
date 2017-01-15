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
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeAPIServerOptionsBuilder adds options for the apiserver to the model
type KubeAPIServerOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeAPIServerOptionsBuilder{}

func (b *KubeAPIServerOptionsBuilder) BuildOptions(o interface{}) error {
	options := o.(*kops.ClusterSpec)
	if options.KubeAPIServer == nil {
		options.KubeAPIServer = &kops.KubeAPIServerConfig{}
	}

	if options.KubeAPIServer.APIServerCount == nil {
		count := b.buildAPIServerCount()
		if count == 0 {
			return fmt.Errorf("no instance groups found")
		}
		options.KubeAPIServer.APIServerCount = fi.Int32(int32(count))
	}

	if options.KubeAPIServer.StorageBackend == nil {
		// For the moment, we continue to use etcd2
		options.KubeAPIServer.StorageBackend = fi.String("etcd2")
	}

	return nil
}

func (b *KubeAPIServerOptionsBuilder) buildAPIServerCount() int {
	// The --apiserver-count flag is (generally agreed) to be something we need to get rid of in k8s

	// We should do something like this:

	//count := 0
	//for _, ig := range b.InstanceGroups {
	//	if !ig.IsMaster() {
	//		continue
	//	}
	//	size := fi.IntValue(ig.Spec.MaxSize)
	//	if size == 0 {
	//		size = fi.IntValue(ig.Spec.MinSize)
	//	}
	//	count += size
	//}

	// But if we do, we end up with a weird dependency on InstanceGroups.  We actually could tolerate
	// that in kops, but we don't really want to.

	// So instead, we assume that the etcd cluster size is the API Server Count.
	// We can re-examine this when we allow separate etcd clusters - at which time hopefully
	// the flag won't exist

	counts := make(map[string]int)
	for _, etcdCluster := range b.Context.Cluster.Spec.EtcdClusters {
		counts[etcdCluster.Name] = len(etcdCluster.Members)
	}

	count := counts["main"]

	return count
}
