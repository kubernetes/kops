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

package do

import (
	"strings"

	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/upup/pkg/fi"
)

const TagKubernetesClusterIndex = "k8s-index"
const TagNameEtcdClusterPrefix = "etcdCluster-"
const TagNameRolePrefix = "k8s.io/role/"
const TagKubernetesClusterNamePrefix = "KubernetesCluster"
const TagKubernetesClusterMasterPrefix = "KubernetesCluster-Master"
const TagKubernetesClusterInstanceGroupPrefix = "kops-instancegroup"

func SafeClusterName(clusterName string) string {
	// DO does not support . in tags / names
	safeClusterName := strings.Replace(clusterName, ".", "-", -1)
	return safeClusterName
}

func NewDOCloud(region string) (fi.Cloud, error) {
	return digitalocean.NewCloud(region)
}
