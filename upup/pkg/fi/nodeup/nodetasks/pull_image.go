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

package nodetasks

import (
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
)

// PullImageTask is responsible for pulling a docker image
type PullImageTask struct {
	Name    string
	Runtime string
}

var (
	_ fi.NodeupTask            = &PullImageTask{}
	_ fi.NodeupHasDependencies = &PullImageTask{}
)

func (t *PullImageTask) GetDependencies(tasks map[string]fi.NodeupTask) []fi.NodeupTask {
	// ImagePullTask depends on the container runtime service to ensure we
	// sideload images after the container runtime is completely updated and
	// configured.
	var deps []fi.NodeupTask
	for _, v := range tasks {
		if svc, ok := v.(*Service); ok && svc.Name == containerdService {
			deps = append(deps, v)
		}
		if svc, ok := v.(*Service); ok && svc.Name == dockerService {
			deps = append(deps, v)
		}
	}
	return deps
}

func (t *PullImageTask) GetName() *string {
	if t.Name == "" {
		return nil
	}
	return &t.Name
}

func (e *PullImageTask) Run(c *fi.NodeupContext) error {
	runtime := e.Runtime
	if runtime != "docker" && runtime != "containerd" {
		return fmt.Errorf("no runtime specified")
	}

	// Pull the container image
	var args []string
	switch runtime {
	case "docker":
		args = []string{"docker", "pull", e.Name}
	case "containerd":
		args = []string{"ctr", "--namespace", "k8s.io", "images", "pull", e.Name}
	default:
		return fmt.Errorf("unknown container runtime: %s", runtime)
	}
	human := strings.Join(args, " ")

	klog.Infof("running command %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pulling docker image with '%s': %v: %s", human, err, string(output))
	}

	return nil
}
