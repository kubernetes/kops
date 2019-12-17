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

package assettasks

import (
	"fmt"
	"os/exec"

	"k8s.io/klog"
)

// dockerCLI encapsulates access to docker via the CLI
type dockerCLI struct {
}

// newDockerCLI builds a dockerCLI object, for talking to docker via the CLI
func newDockerCLI() (*dockerCLI, error) {
	return &dockerCLI{}, nil
}

// pullImage does a `docker pull`, shelling out to the CLI
func (d *dockerCLI) pullImage(name string) error {
	klog.V(4).Infof("docker pull for image %q", name)

	cmd := exec.Command("docker", "pull", name)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error pulling image %q: %v", name, err)
	}

	return nil
}

// pushImage does a docker push, shelling out to the CLI
func (d *dockerCLI) pushImage(name string) error {
	klog.V(4).Infof("docker push for image %q", name)

	cmd := exec.Command("docker", "push", name)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error pushing image %q: %v", name, err)
	}

	return nil
}
