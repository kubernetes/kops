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
	"context"
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
)

// PullImageTask is responsible for pulling a docker image
type PullImageTask struct {
	Name string

	// auth returns credentials for pulling the image. It is unexported so that
	// task reflection ignores it.
	auth func(ctx context.Context) (user string, password string, err error)
}

// SetAuth injects the credentials for pulling the image.
func (t *PullImageTask) SetAuth(auth func(ctx context.Context) (string, string, error)) {
	t.auth = auth
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
	// Pull the container image
	args := []string{"ctr", "--namespace", "k8s.io", "images", "pull"}
	human := strings.Join(append(append([]string{}, args...), e.Name), " ")

	if e.auth != nil {
		user, password, err := e.auth(context.TODO())
		if err != nil {
			return fmt.Errorf("getting credentials for image %q: %w", e.Name, err)
		}
		args = append(args, "--user", user+":"+password)
	}
	args = append(args, e.Name)

	klog.Infof("running command %s", human)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pulling docker image with '%s': %v: %s", human, err, string(output))
	}

	return nil
}
