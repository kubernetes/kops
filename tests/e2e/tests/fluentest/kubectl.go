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

package fluentest

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func KubectlCurrentContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running command %s: %w", strings.Join(cmd.Args, " "), err)
	}

	context := stdout.String()
	context = strings.TrimSpace(context)
	if context == "" {
		return "", fmt.Errorf("kubectl current-context was empty")
	}
	return context, nil
}

func RESTConfigFromKubeconfig() (*rest.Config, error) {
	home := homedir.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("failed to get homedir for default kube config location")
	}

	kubeconfigPath := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %q: %w", kubeconfigPath, err)
	}

	return config, nil
}
