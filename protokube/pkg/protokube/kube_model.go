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

package protokube

//
//import (
//	"fmt"
//	"github.com/golang/glog"
//)
//
//// ApplyModel applies the configuration as specified in the model
//func (k *KubeBoot) ApplyModel() error {
//	etcdClusters, err := k.BuildEtcdClusters(modelDir)
//	if err != nil {
//		return fmt.Errorf("error building etcd models: %v", err)
//	}
//
//	for _, etcdCluster := range etcdClusters {
//		glog.Infof("configuring etcd cluster %s", etcdCluster.ClusterName)
//		err := etcdCluster.configure(k)
//		if err != nil {
//			return fmt.Errorf("error applying etcd model: %v", err)
//		}
//	}
//
//	return nil
//}
