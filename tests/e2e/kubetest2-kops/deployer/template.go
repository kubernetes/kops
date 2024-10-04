/*
Copyright 2021 The Kubernetes Authors.

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

package deployer

import (
	"fmt"
	"os"
	"path"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/exec"
	"sigs.k8s.io/yaml"
)

// renderTemplate will render the manifest template with the provided values,
// setting the deployer's manifestPath
func (d *deployer) renderTemplate(values map[string]interface{}) error {
	dir, err := os.MkdirTemp("", "kops-template")
	if err != nil {
		return err
	}

	valuesBytes, err := yaml.Marshal(values)
	if err != nil {
		return err
	}
	valuesPath := path.Join(dir, "values.yaml")
	err = os.WriteFile(valuesPath, valuesBytes, 0o644)
	if err != nil {
		return err
	}

	manifestPath := path.Join(dir, "manifest.yaml")
	d.manifestPath = manifestPath

	args := []string{
		d.KopsBinaryPath, "toolbox", "template",
		"--template", d.TemplatePath,
		"--output", manifestPath,
		"--values", valuesPath,
		"--name", d.ClusterName,
	}
	klog.Info(strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.SetEnv(d.env()...)

	exec.InheritOutput(cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *deployer) templateValues(zones []string, publicIP string) (map[string]interface{}, error) {
	publicKey, err := os.ReadFile(d.SSHPublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error reading public key file: %q", err)
	}
	return map[string]interface{}{
		"cloudProvider":     d.CloudProvider,
		"clusterName":       d.ClusterName,
		"kubernetesVersion": d.KubernetesVersion,
		"publicIP":          publicIP,
		"stateStore":        d.stateStore(),
		"discoveryStore":    d.discoveryStore(),
		"zones":             zones,
		"sshPublicKey":      string(publicKey),
	}, nil
}
