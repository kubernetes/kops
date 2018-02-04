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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api-gcp/util"
	apiutil "k8s.io/kube-deploy/cluster-api/util"
)

const (
	MasterIPAttempts       = 40
	SleepSecondsPerAttempt = 5
	RetryAttempts          = 30
	DeleteAttempts         = 150
	DeleteSleepSeconds     = 5
)

func (d *deployer) createCluster(c *clusterv1.Cluster, machines []*clusterv1.Machine, vmCreated *bool) error {
	if c.GetName() == "" {
		return fmt.Errorf("cluster name must be specified for cluster creation")
	}
	master := util.GetMaster(machines)
	if master == nil {
		return fmt.Errorf("master spec must be provided for cluster creation")
	}

	if master.GetName() == "" && master.GetGenerateName() == "" {
		return fmt.Errorf("master name must be specified for cluster creation")
	}

	if master.GetName() == "" {
		master.Name = master.GetGenerateName() + c.GetName()
	}

	glog.Infof("Starting cluster creation %s", c.GetName())

	glog.Infof("Starting master creation %s", master.GetName())

	if err := d.actuator.Create(c, master); err != nil {
		return err
	}

	*vmCreated = true
	glog.Infof("Created master %s", master.GetName())

	masterIP, err := d.getMasterIP(master)
	if err != nil {
		return fmt.Errorf("unable to get master IP: %v", err)
	}

	c.Status.APIEndpoints = append(c.Status.APIEndpoints,
		clusterv1.APIEndpoint{
			Host: masterIP,
			Port: 443,
		})

	if err := d.copyKubeConfig(master); err != nil {
		return fmt.Errorf("unable to write kubeconfig: %v", err)
	}

	glog.Info("Waiting for apiserver to become healthy...")
	if err := d.waitForApiserver(masterIP, 1*time.Minute); err != nil {
		return fmt.Errorf("apiserver never came up: %v", err)
	}

	if err := d.initApiClient(); err != nil {
		return err
	}
	glog.Info("Starting the machine controller...")
	if err := d.actuator.CreateMachineController(c, machines); err != nil {
		return fmt.Errorf("can't create machine controller: %v", err)
	}

	if err := d.createClusterCRD(); err != nil {
		return err
	}

	if _, err := d.client.Clusters().Create(c); err != nil {
		return err
	}

	if err := d.createMachineCRD(); err != nil {
		return err
	}

	if err := d.createMachines(machines); err != nil {
		return err
	}
	return nil
}

func (d *deployer) createClusterCRD() error {
	cs, err := util.NewClientSet(d.configPath)
	if err != nil {
		return err
	}

	success := false
	for i := 0; i <= RetryAttempts; i++ {
		if _, err = clusterv1.CreateClustersCRD(cs); err != nil {
			glog.Info("Failure creating Clusters CRD (will retry).")
			time.Sleep(SleepSecondsPerAttempt * time.Second)
			continue
		}
		success = true
		glog.Info("Clusters CRD created succuessfully!")
		break
	}

	if !success {
		return fmt.Errorf("error creating Clusters CRD: %v", err)
	}
	return nil
}

func (d *deployer) createMachineCRD() error {
	cs, err := util.NewClientSet(d.configPath)
	if err != nil {
		return err
	}

	success := false
	for i := 0; i <= RetryAttempts; i++ {
		if _, err = clusterv1.CreateMachinesCRD(cs); err != nil {
			glog.Info("Failure creating Machines CRD (will retry).")
			time.Sleep(time.Duration(SleepSecondsPerAttempt) * time.Second)
			continue
		}
		success = true
		glog.Info("Machines CRD created successfully!")
		break
	}

	if !success {
		return fmt.Errorf("error creating Machines CRD: %v", err)
	}
	return nil
}

func (d *deployer) createMachines(machines []*clusterv1.Machine) error {
	for _, machine := range machines {
		m, err := d.client.Machines().Create(machine)
		if err != nil {
			return err
		}
		glog.Infof("Added machine [%s]", m.Name)
	}
	return nil
}

func (d *deployer) createMachine(m *clusterv1.Machine) error {
	return d.createMachines([]*clusterv1.Machine{m})
}

func (d *deployer) deleteAllMachines() error {
	machines, err := d.client.Machines().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, m := range machines.Items {
		if err := d.delete(m.Name); err != nil {
			return err
		}
		glog.Infof("Deleted machine object %s", m.Name)
	}
	return nil
}

func (d *deployer) delete(name string) error {
	// TODO  https://github.com/kubernetes/kube-deploy/issues/390
	return d.client.Machines().Delete(name, &metav1.DeleteOptions{})
}

func (d *deployer) listMachines() ([]*clusterv1.Machine, error) {
	machines, err := d.client.Machines().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return util.MachineP(machines.Items), nil
}

func (d *deployer) getCluster() (*clusterv1.Cluster, error) {
	clusters, err := d.client.Clusters().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if len(clusters.Items) != 1 {
		return nil, fmt.Errorf("cluster object count != 1")
	}
	return &clusters.Items[0], nil
}

func (d *deployer) getMasterIP(master *clusterv1.Machine) (string, error) {
	for i := 0; i < MasterIPAttempts; i++ {
		ip, err := d.actuator.GetIP(master)
		if err != nil || ip == "" {
			glog.Info("Hanging for master IP...")
			time.Sleep(time.Duration(SleepSecondsPerAttempt) * time.Second)
			continue
		}
		return ip, nil
	}
	return "", fmt.Errorf("unable to find Master IP after defined wait")
}

func (d *deployer) copyKubeConfig(master *clusterv1.Machine) error {
	for i := 0; i <= RetryAttempts; i++ {
		var config string
		var err error
		if config, err = d.actuator.GetKubeConfig(master); err != nil || config == "" {
			glog.Infof("Waiting for Kubernetes to come up...")
			time.Sleep(time.Duration(SleepSecondsPerAttempt) * time.Second)
			continue
		}

		return d.writeConfigToDisk(config)
	}
	return fmt.Errorf("timedout writing kubeconfig")
}

func (d *deployer) initApiClient() error {
	c, err := apiutil.NewApiClient(d.configPath)
	if err != nil {
		return err
	}
	d.client = c
	return nil

}
func (d *deployer) writeConfigToDisk(config string) error {
	file, err := os.Create(d.configPath)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(config); err != nil {
		return err
	}
	defer file.Close()

	file.Sync() // flush
	glog.Infof("wrote kubeconfig to [%s]", d.configPath)
	return nil
}

// Make sure you successfully call setMasterIp first.
func (d *deployer) waitForApiserver(master string, timeout time.Duration) error {
	endpoint := fmt.Sprintf("https://%s/healthz", master)

	// Skip certificate validation since we're only looking for signs of
	// health, and we're not going to have the CA in our default chain.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	startTime := time.Now()

	var err error
	var resp *http.Response
	for time.Now().Sub(startTime) < timeout {
		resp, err = client.Get(endpoint)
		if err == nil && resp.StatusCode == 200 {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return err
}
