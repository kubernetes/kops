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

package cloud

import (
	"fmt"

	"github.com/golang/glog"

	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/client"
	"k8s.io/kube-deploy/cluster-api-gcp/cloud/google"
)

// An actuator that just logs instead of doing anything.
type loggingMachineActuator struct{}

const config = `
apiVersion: v1
kind: config
preferences: {}
`

func NewMachineActuator(cloud string, kubeadmToken string, machineClient client.MachinesInterface) (MachineActuator, error) {
	switch cloud {
	case "google":
		return google.NewMachineActuator(kubeadmToken, machineClient)
	case "test", "aws", "azure":
		return &loggingMachineActuator{}, nil
	default:
		return nil, fmt.Errorf("Not recognized cloud provider: %s\n", cloud)
	}
}

func (a loggingMachineActuator) Create(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	glog.Infof("actuator received create: %s\n", machine.ObjectMeta.Name)
	return nil
}

func (a loggingMachineActuator) Delete(machine *clusterv1.Machine) error {
	glog.Infof("actuator received delete: %s\n", machine.ObjectMeta.Name)
	return nil
}

func (a loggingMachineActuator) Update(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	glog.Infof("actuator received update: %s\n", machine.ObjectMeta.Name)
	return nil
}

func (a loggingMachineActuator) Exists(machine *clusterv1.Machine) (bool, error) {
	glog.Infof("actuator received exists %s\n", machine.ObjectMeta.Name)
	return false, nil
}

func (a loggingMachineActuator) GetIP(machine *clusterv1.Machine) (string, error) {
	glog.Infof("actuator received GetIP: %s\n", machine.ObjectMeta.Name)
	return "0.0.0.0", nil
}

func (a loggingMachineActuator) GetKubeConfig(master *clusterv1.Machine) (string, error) {
	glog.Infof("actuator received GetKubeConfig: %s\n", master.ObjectMeta.Name)
	return config, nil
}

func (a loggingMachineActuator) CreateMachineController(cluster *clusterv1.Cluster, machines []*clusterv1.Machine) error {
	glog.Infof("actuator received CreateMachineController: %q\n", machines)
	return nil
}

func (a loggingMachineActuator) PostDelete(cluster *clusterv1.Cluster, machines []*clusterv1.Machine) error {
	glog.Infof("actuator received PostDelete: %q\n", machines)
	return nil
}
