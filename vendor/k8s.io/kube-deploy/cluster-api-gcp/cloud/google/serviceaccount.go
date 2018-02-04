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

package google

import (
	"fmt"
	"os/exec"

	"github.com/golang/glog"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api-gcp/util"
)

const (
	ServiceAccountPrefix    = "k8s-machine-controller-"
	ServiceAccount          = "service-account"
	MachineControllerSecret = "machine-controller-credential"
)

// Creates a GCP service account for the machine controller, granted the
// permissions to manage compute instances, and stores its credentials as a
// Kubernetes secret.
func (gce *GCEClient) CreateMachineControllerServiceAccount(cluster *clusterv1.Cluster, initialMachines []*clusterv1.Machine) error {

	if len(initialMachines) == 0 {
		return fmt.Errorf("machine count is zero, cannot create service a/c")
	}

	// TODO: use real go bindings
	// Figure out what projects the service account needs permission to.
	projects, err := gce.getProjects(initialMachines)
	if err != nil {
		return err
	}

	// The service account needs to be created in a single project, so just
	// use the first one, but grant permission to all projects in the list.
	project := projects[0]
	accountId := ServiceAccountPrefix + util.RandomString(5)

	err = run("gcloud", "--project", project, "iam", "service-accounts", "create", "--display-name=k8s machines controller", accountId)
	if err != nil {
		return fmt.Errorf("couldn't create service account: %v", err)
	}

	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountId, project)
	localFile := accountId + "-key.json"

	for _, project := range projects {
		err = run("gcloud", "projects", "add-iam-policy-binding", project, "--member=serviceAccount:"+email, "--role=roles/compute.instanceAdmin.v1")
		if err != nil {
			return fmt.Errorf("couldn't grant permissions to service account: %v", err)
		}
	}

	err = run("gcloud", "--project", project, "iam", "service-accounts", "keys", "create", localFile, "--iam-account", email)
	if err != nil {
		return fmt.Errorf("couldn't create service account key: %v", err)
	}

	err = run("kubectl", "create", "secret", "generic", "-n", "kube-system", MachineControllerSecret, "--from-file=service-account.json="+localFile)
	if err != nil {
		return fmt.Errorf("couldn't import service account key as credential: %v", err)
	}
	if err := run("rm", localFile); err != nil {
		glog.Error(err)
	}

	if cluster.ObjectMeta.Annotations == nil {
		cluster.ObjectMeta.Annotations = make(map[string]string)
	}
	cluster.ObjectMeta.Annotations[ServiceAccount] = email
	return nil
}

func (gce *GCEClient) DeleteMachineControllerServiceAccount(cluster *clusterv1.Cluster, machines []*clusterv1.Machine) error {
	if len(machines) == 0 {
		glog.Info("machine count is zero, cannot determine project for service a/c deletion")
		return nil
	}

	projects, err := gce.getProjects(machines)
	if err != nil {
		return err
	}
	project := projects[0]
	var email string
	if cluster.ObjectMeta.Annotations != nil {
		email = cluster.ObjectMeta.Annotations[ServiceAccount]
	}

	if email == "" {
		glog.Info("No service a/c found in cluster.")
		return nil
	}

	err = run("gcloud", "projects", "remove-iam-policy-binding", project, "--member=serviceAccount:"+email, "--role=roles/compute.instanceAdmin.v1")

	if err != nil {
		return fmt.Errorf("couldn't remove permissions to service account: %v", err)
	}

	err = run("gcloud", "--project", project, "iam", "service-accounts", "delete", email)
	if err != nil {
		return fmt.Errorf("couldn't delete service account: %v", err)
	}
	return nil
}

func (gce *GCEClient) getProjects(machines []*clusterv1.Machine) ([]string, error) {
	// Figure out what projects the service account needs permission to.
	var projects []string
	for _, machine := range machines {
		config, err := gce.providerconfig(machine.Spec.ProviderConfig)
		if err != nil {
			return nil, err
		}

		projects = append(projects, config.Project)
	}
	return projects, nil
}

func run(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	if out, err := c.CombinedOutput(); err != nil {
		return fmt.Errorf("error: %v, output: %s", err, string(out))
	}
	return nil
}
