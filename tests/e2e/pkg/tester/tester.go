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
	"fmt"
	"os"

	"k8s.io/klog/v2"

	"sigs.k8s.io/kubetest2/pkg/testers/ginkgo"
)

// Tester wraps kubetest2's ginkgo tester with additional functionality
type Tester struct {
	Ginkgo *ginkgo.Tester
}

// Test runs the test
func (t *Tester) Test() error {
	if err := t.pretestSetup(); err != nil {
		return err
	}
	return t.Ginkgo.Test()
}

func (t *Tester) pretestSetup() error {
	kubectlPath, err := t.AcquireKubectl()
	if err != nil {
		return fmt.Errorf("failed to get kubectl package from published releases: %s", err)
	}
	existingPath := os.Getenv("PATH")
	os.Setenv("PATH", fmt.Sprintf("%v:%v", kubectlPath, existingPath))
	return nil
}

func (t *Tester) Execute() error {
	return t.Ginkgo.Execute()
}

func NewDefaultTester() *Tester {
	return &Tester{
		Ginkgo: ginkgo.NewDefaultTester(),
	}
}

func Main() {
	t := NewDefaultTester()
	if err := t.Execute(); err != nil {
		klog.Fatalf("failed to run ginkgo tester: %v", err)
	}
}
