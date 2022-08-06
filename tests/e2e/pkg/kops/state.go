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

package kops

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	api "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

// GetCluster will retrieve the specified Cluster from the state store.
func GetCluster(kopsBinary, clusterName string, env []string) (*api.Cluster, error) {
	args := []string{
		kopsBinary, "get", "cluster", clusterName, "-ojson", "--full",
	}
	c := exec.Command(args[0], args[1:]...)
	c.SetEnv(env...)
	var stdout bytes.Buffer
	c.SetStdout(&stdout)
	var stderr bytes.Buffer
	c.SetStderr(&stderr)
	if err := c.Run(); err != nil {
		klog.Warningf("failed to run %s; stderr=%s", strings.Join(args, " "), stderr.String())
		return nil, fmt.Errorf("error querying cluster from %s: %w", strings.Join(args, " "), err)
	}

	var jsonBytes []byte

	cluster := &api.Cluster{}
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//") {
			continue
		}
		jsonBytes = append(jsonBytes, scanner.Bytes()...)
	}
	if err := json.Unmarshal(jsonBytes, cluster); err != nil {
		return nil, fmt.Errorf("error parsing cluster json: %w", err)
	}
	return cluster, nil
}

// GetInstanceGroups will retrieve the instance groups for the specified Cluster from the state store.
func GetInstanceGroups(kopsBinary, clusterName string, env []string) ([]*api.InstanceGroup, error) {
	args := []string{
		kopsBinary, "get", "instancegroups", "--name", clusterName, "-ojson",
	}
	c := exec.Command(args[0], args[1:]...)
	c.SetEnv(env...)
	var stdout bytes.Buffer
	c.SetStdout(&stdout)
	var stderr bytes.Buffer
	c.SetStderr(&stderr)
	if err := c.Run(); err != nil {
		klog.Warningf("failed to run %s; stderr=%s", strings.Join(args, " "), stderr.String())
		return nil, fmt.Errorf("error querying instance groups from %s: %w", strings.Join(args, " "), err)
	}

	var igs []*api.InstanceGroup
	if err := json.Unmarshal(stdout.Bytes(), &igs); err != nil {
		return nil, fmt.Errorf("error parsing instance groups json: %w", err)
	}
	return igs, nil
}
