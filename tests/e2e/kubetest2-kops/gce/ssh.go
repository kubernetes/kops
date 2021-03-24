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

package gce

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

func SetupSSH(project string) (string, string, error) {
	dir, err := os.MkdirTemp("", "kops-ssh")
	if err != nil {
		return "", "", err
	}

	privateKey := filepath.Join(dir, "key")
	configArgs := []string{
		"gcloud",
		"compute",
		fmt.Sprintf("--project=%v", project),
		"config-ssh",
		fmt.Sprintf("--ssh-key-file=%v", privateKey),
	}
	klog.Info(strings.Join(configArgs, " "))
	cmd := exec.Command(configArgs[0], configArgs[1:]...)

	exec.InheritOutput(cmd)
	err = cmd.Run()
	if err != nil {
		return "", "", err
	}

	return privateKey, fmt.Sprintf("%v.pub", privateKey), nil
}
