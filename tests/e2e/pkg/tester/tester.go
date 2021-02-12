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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	ps "github.com/mitchellh/go-ps"
	"github.com/octago/sflags/gen/gpflag"
	"k8s.io/klog/v2"

	"sigs.k8s.io/kubetest2/pkg/testers/ginkgo"
)

// Tester wraps kubetest2's ginkgo tester with additional functionality
type Tester struct {
	*ginkgo.Tester
	Timeout time.Duration `desc:"Terminate the testing after the specified amount of time."`
}

func (t *Tester) pretestSetup() error {
	klog.Infof("Setting timeout value of %v", t.Timeout)
	kubectlPath, err := t.AcquireKubectl()
	if err != nil {
		return fmt.Errorf("failed to get kubectl package from published releases: %s", err)
	}

	existingPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%v:%v", filepath.Dir(kubectlPath), existingPath)
	klog.Info("Setting PATH=", newPath)
	return os.Setenv("PATH", newPath)
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

	execCh := make(chan error, 1)
	go func() {
		execCh <- t.Execute()
	}()
	select {
	case err := <-execCh:
		if err != nil {
			return err
		}
	case <-time.After(t.Timeout):
		signalChildren()
		return errors.New("ginkgo timeout")
	}
	return nil
}

func NewDefaultTester() *Tester {
	return &Tester{
		Tester: ginkgo.NewDefaultTester(),
		// TODO: Set back to 0 before merging this PR
		Timeout: time.Duration(5 * time.Minute),
	}
}

func Main() {
	t := NewDefaultTester()
	if err := t.execute(); err != nil {
		klog.Fatalf("failed to run ginkgo tester: %v", err)
	}
}

// A temporary hack to send SIGINT to the e2e.test executable to dump its stack trace
// Only signals direct children and not descendents.
func signalChildren() {
	pid := os.Getpid()
	allProcesses, err := ps.Processes()
	if err != nil {
		klog.Warning("Failed to list all processes: %v", err)
		return
	}
	for _, p := range allProcesses {
		if p.PPid() != pid {
			continue
		}
		if err := syscall.Kill(p.Pid(), syscall.SIGINT); err != nil {
			klog.Warningf("Failed to issue SIGINT to process %+v", p)
		}
	}
	return
}
