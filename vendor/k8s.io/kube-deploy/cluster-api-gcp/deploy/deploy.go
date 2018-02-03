/*
Copyright 2017 The Kubernetes Authors.

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

package deploy

import (
	"fmt"

	"github.com/golang/glog"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/client"
	"k8s.io/kube-deploy/cluster-api-gcp/cloud"
	"k8s.io/kube-deploy/cluster-api-gcp/util"
	apiutil "k8s.io/kube-deploy/cluster-api/util"
)

type deployer struct {
	token      string
	configPath string
	actuator   cloud.MachineActuator
	client     *client.ClusterAPIV1Alpha1Client
}

//it takes path for kubeconfig file.
func NewDeployer(provider string, configPath string) *deployer {
	token := util.RandomToken()
	if configPath == "" {
		configPath = apiutil.GetDefaultKubeConfigPath()
	}
	a, err := cloud.NewMachineActuator(provider, token, nil)
	if err != nil {
		glog.Exit(err)
	}
	return &deployer{
		token:      token,
		actuator:   a,
		configPath: configPath,
	}
}

func (d *deployer) CreateCluster(c *clusterv1.Cluster, machines []*clusterv1.Machine) error {
	vmCreated := false
	if err := d.createCluster(c, machines, &vmCreated); err != nil {
		if vmCreated {
			d.deleteMasterVM(machines)
		}
		d.actuator.PostDelete(c, machines)
		return err
	}

	glog.Infof("The [%s] cluster has been created successfully!", c.Name)
	glog.Info("You can now `kubectl get nodes`")
	return nil
}

func (d *deployer) AddNodes(machines []*clusterv1.Machine) error {
	if err := d.createMachines(machines); err != nil {
		return err
	}
	return nil
}

func (d *deployer) DeleteCluster() error {
	if err := d.initApiClient(); err != nil {
		return err
	}

	machines, err := d.listMachines()
	if err != nil {
		return err
	}

	cluster, err := d.getCluster()
	if err != nil {
		return err
	}

	glog.Info("Deleting machine objects")
	if err := d.deleteAllMachines(); err != nil {
		return err
	}

	if err := d.deleteMasterVM(machines); err != nil {
		glog.Errorf("Error deleting master vm", err)
	}

	glog.Info("Running post delete operations")
	if err := d.actuator.PostDelete(cluster, machines); err != nil {
		return err
	}
	glog.Infof("Deletion complete")
	return nil
}

func (d *deployer) deleteMasterVM(machines []*clusterv1.Machine) error {
	master := util.GetMaster(machines)
	if master == nil {
		return fmt.Errorf("error deleting master vm, no master found")
	}

	glog.Infof("Deleting master vm %s", master.Name)
	if err := d.actuator.Delete(master); err != nil {
		return err
	}
	return nil
}
