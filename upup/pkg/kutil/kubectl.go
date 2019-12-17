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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"k8s.io/kops/pkg/kubeconfig"
)

type Kubectl struct {
	KubectlPath string
}

//func (k *Kubectl) GetCurrentContext() (string, error) {
//	s, err := k.execKubectl("config", "current-context")
//	if err != nil {
//		return "", err
//	}
//	s = strings.TrimSpace(s)
//	return s, nil
//}

func (k *Kubectl) GetCurrentContext() (string, error) {
	pathOptions := clientcmd.NewDefaultPathOptions()

	config, err := pathOptions.GetStartingConfig()
	if err != nil {
		return "", err
	}

	return config.CurrentContext, nil

	//s, err := k.execKubectl("config", "current-context")
	//if err != nil {
	//	return "", err
	//}
	//s = strings.TrimSpace(s)
	//return s, nil
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

// Apply calls kubectl apply to apply the manifest.
// We will likely in future change this to create things directly (or more likely embed this logic into kubectl itself)
func (k *Kubectl) Apply(context string, data []byte) error {
	localManifestFile, err := ioutil.TempFile("", "manifest")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}

	defer func() {
		if err := os.Remove(localManifestFile.Name()); err != nil {
			klog.Warningf("error deleting temp file %q: %v", localManifestFile.Name(), err)
		}
	}()

	if err := ioutil.WriteFile(localManifestFile.Name(), data, 0600); err != nil {
		return fmt.Errorf("error writing temp file: %v", err)
	}

	_, _, err = k.execKubectl("apply", "--context", context, "-f", localManifestFile.Name())
	return err
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
