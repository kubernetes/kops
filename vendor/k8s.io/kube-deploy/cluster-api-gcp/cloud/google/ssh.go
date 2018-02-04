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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/glog"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
)

const (
	MachineControllerSshKeySecret = "machine-controller-sshkeys"
	// Arbitrary name used for SSH.
	SshUser                = "clusterapi"
	SshKeyFile             = "clusterapi-key"
	SshKeyFilePublic       = SshKeyFile + ".pub"
	SshKeyFilePublicGcloud = SshKeyFilePublic + ".gcloud"
)

func createSshKeyPairs() error {
	err := run("ssh-keygen", "-t", "rsa", "-f", SshKeyFile, "-C", SshUser, "-N", "")
	if err != nil {
		return fmt.Errorf("couldn't generate RSA keys: %v", err)
	}

	// Prepare a gce format public key file
	outfile, err := os.Create(SshKeyFilePublicGcloud)
	if err != nil {
		return err
	}
	defer outfile.Close()

	b, err := ioutil.ReadFile(SshKeyFilePublic)
	if err == nil {
		outfile.WriteString(SshUser + ":" + string(b))
	}

	return err
}

func cleanupSshKeyPairs() {
	os.Remove(SshKeyFile)
	os.Remove(SshKeyFilePublic)
	os.Remove(SshKeyFilePublicGcloud)
}

// It creates secret to store private key.
func (gce *GCEClient) setupSSHAccess(m *clusterv1.Machine) error {
	// Create public/private key pairs
	err := createSshKeyPairs()
	if err != nil {
		return err
	}

	config, err := gce.providerconfig(m.Spec.ProviderConfig)
	if err != nil {
		return err
	}

	err = run("gcloud", "compute", "instances", "add-metadata", m.Name,
		"--metadata-from-file", "ssh-keys="+SshKeyFile+".pub.gcloud",
		"--project", config.Project, "--zone", config.Zone)
	if err != nil {
		return err
	}

	// Create secrets so that machine controller container can load them.
	err = run("kubectl", "create", "secret", "generic", "-n", "kube-system", MachineControllerSshKeySecret, "--from-file=private="+SshKeyFile, "--from-literal=user="+SshUser)
	if err != nil {
		return fmt.Errorf("couldn't create service account key as credential: %v", err)
	}

	cleanupSshKeyPairs()

	return err
}

func (gce *GCEClient) remoteSshCommand(m *clusterv1.Machine, cmd string) (string, error) {
	glog.Infof("Remote SSH execution '%s' on %s", cmd, m.ObjectMeta.Name)

	publicIP, err := gce.GetIP(m)
	if err != nil {
		return "", err
	}

	command := fmt.Sprintf("echo STARTFILE; %s", cmd)
	c := exec.Command("ssh", "-i", gce.sshCreds.privateKeyPath, gce.sshCreds.user+"@"+publicIP, command)
	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error: %v, output: %s", err, string(out))
	}
	result := strings.TrimSpace(string(out))
	parts := strings.Split(result, "STARTFILE")
	if len(parts) != 2 {
		return "", nil
	}
	// TODO: Check error.
	return strings.TrimSpace(parts[1]), nil
}
