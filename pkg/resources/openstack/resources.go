/*
Copyright 2018 The Kubernetes Authors.

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

package openstack

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

type openstackListFn func() ([]*resources.Resource, error)

const (
	typeSSHKey = "SSHKey"
)

func ListResources(cloud openstack.OpenstackCloud, clusterName string) (map[string]*resources.Resource, error) {
	resources := make(map[string]*resources.Resource)

	os := &clusterDiscoveryOS{
		cloud:       cloud,
		osCloud:     cloud,
		clusterName: clusterName,
	}

	listFunctions := []openstackListFn{
		os.ListKeypairs,
	}
	for _, fn := range listFunctions {
		resourceTrackers, err := fn()
		if err != nil {
			return nil, err
		}
		for _, t := range resourceTrackers {
			resources[t.Type+":"+t.ID] = t
		}
	}

	return resources, nil
}

type clusterDiscoveryOS struct {
	cloud       fi.Cloud
	osCloud     openstack.OpenstackCloud
	clusterName string

	//instanceTemplates []*compute.InstanceTemplate
	zones []string
}

func openstackKeyPairName(org string) string {
	name := strings.Replace(org, ".", "-", -1)
	name = strings.Replace(name, ":", "_", -1)
	return name
}

func (os *clusterDiscoveryOS) ListKeypairs() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	ks, err := os.osCloud.ListKeypairs()
	if err != nil {
		return resourceTrackers, err
	}

	for _, key := range ks {
		prefix := "kubernetes-" + openstackKeyPairName(os.clusterName)
		if strings.HasPrefix(key.Name, prefix) {
			resourceTracker := &resources.Resource{
				Name:    key.Name,
				ID:      key.Name,
				Type:    typeSSHKey,
				Deleter: DeleteSSHKey,
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}
	return resourceTrackers, nil
}

func DeleteSSHKey(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(openstack.OpenstackCloud)
	name := r.Name

	glog.V(2).Infof("Deleting Openstack Keypair %q", name)

	err := c.DeleteKeyPair(r.Name)
	if err != nil {
		return fmt.Errorf("error deleting KeyPair %q: %v", name, err)
	}
	return nil
}
