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

package tester

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/octago/sflags/gen/gpflag"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kubetest2/pkg/testers/ginkgo"
)

// Tester wraps kubetest2's ginkgo tester with additional functionality
type Tester struct {
	*ginkgo.Tester
}

func (t *Tester) pretestSetup() error {
	kubectlPath, err := t.AcquireKubectl()
	if err != nil {
		return fmt.Errorf("failed to get kubectl package from published releases: %s", err)
	}

	existingPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%v:%v", filepath.Dir(kubectlPath), existingPath)
	klog.Info("Setting PATH=", newPath)
	return os.Setenv("PATH", newPath)
}

// The --host argument was required in the kubernetes e2e tests, until https://github.com/kubernetes/kubernetes/pull/87030
// We can likely drop this when we drop support / testing for k8s 1.17
func (t *Tester) addHostArgument() error {
	args := []string{
		"kubectl", "config", "view", "--minify", "-o", "jsonpath='{.clusters[0].cluster.server}'",
	}
	c := exec.Command(args[0], args[1:]...)
	var stdout bytes.Buffer
	c.Stdout = &stdout
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		klog.Warningf("failed to run %s; stderr=%s", strings.Join(args, " "), stderr.String())
		return fmt.Errorf("error querying current config from kubectl: %w", err)
	}

	server := strings.TrimSpace(stdout.String())
	if server == "" {
		return fmt.Errorf("kubeconfig did not contain server")
	}

	klog.Info("Adding --host=%s", server)
	t.TestArgs += " --host=" + server
	return nil
}

func (t *Tester) execute() error {
	fs, err := gpflag.Parse(t)
	if err != nil {
		return fmt.Errorf("failed to initialize tester: %v", err)
	}

	help := fs.BoolP("help", "h", false, "")
	if err := fs.Parse(os.Args); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	if *help {
		fs.SetOutput(os.Stdout)
		fs.PrintDefaults()
		return nil
	}

	if err := t.pretestSetup(); err != nil {
		return err
	}

	if err := t.addHostArgument(); err != nil {
		return err
	}

	return t.Execute()
}

func NewDefaultTester() *Tester {
	return &Tester{
		ginkgo.NewDefaultTester(),
	}
}

func Main() {
	t := NewDefaultTester()
	if err := t.execute(); err != nil {
		klog.Fatalf("failed to run ginkgo tester: %v", err)
	}
}
