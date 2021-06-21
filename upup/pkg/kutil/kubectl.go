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

package kutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/kubeconfig"
)

type Kubectl struct {
	KubectlPath string
}

func (k *Kubectl) GetConfig(minify bool) (*kubeconfig.KubectlConfig, error) {
	output := "json"
	// TODO: --context doesn't seem to work
	args := []string{"config", "view"}

	if minify {
		args = append(args, "--minify")
	}

	if output != "" {
		args = append(args, "--output", output)
	}

	configString, _, err := k.execKubectl(args...)
	if err != nil {
		return nil, err
	}
	configString = strings.TrimSpace(configString)

	klog.V(8).Infof("config = %q", configString)

	config := &kubeconfig.KubectlConfig{}
	err = json.Unmarshal([]byte(configString), config)
	if err != nil {
		return nil, fmt.Errorf("cannot parse current config from kubectl: %v", err)
	}

	return config, nil
}

func (k *Kubectl) execKubectl(args ...string) (string, string, error) {
	kubectlPath := k.KubectlPath
	if kubectlPath == "" {
		kubectlPath = "kubectl" // Assume in PATH
	}
	cmd := exec.Command(kubectlPath, args...)
	env := os.Environ()
	cmd.Env = env
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	human := strings.Join(cmd.Args, " ")
	klog.V(2).Infof("Running command: %s", human)
	err := cmd.Run()
	if err != nil {
		klog.Infof("error running %s", human)
		klog.Info(stdout.String())
		klog.Info(stderr.String())
		return stdout.String(), stderr.String(), fmt.Errorf("error running kubectl: %v", err)
	}

	return stdout.String(), stderr.String(), err
}
