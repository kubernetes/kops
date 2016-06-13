package kutil

import (
	"fmt"
	"github.com/golang/glog"
	"os"
	"os/exec"
	"strings"
)

type Kubectl struct {
	KubectlPath string
}

func (k *Kubectl) GetCurrentContext() (string, error) {
	s, err := k.execKubectl("config", "current-context")
	if err != nil {
		return "", err
	}
	s = strings.TrimSpace(s)
	return s, nil
}

func (k *Kubectl) GetConfig(minify bool, output string) (string, error) {
	// TODO: --context doesn't seem to work
	args := []string{"config", "view"}

	if minify {
		args = append(args, "--minify")
	}

	if output != "" {
		args = append(args, "--output", output)
	}

	s, err := k.execKubectl(args...)
	if err != nil {
		return "", err
	}
	s = strings.TrimSpace(s)
	return s, nil
}

func (k *Kubectl) execKubectl(args ...string) (string, error) {
	kubectlPath := k.KubectlPath
	if kubectlPath == "" {
		kubectlPath = "kubectl" // Assume in PATH
	}
	cmd := exec.Command(kubectlPath, args...)
	env := os.Environ()
	cmd.Env = env

	human := cmd.Path + strings.Join(cmd.Args, " ")
	glog.V(2).Infof("Running command: %s", human)
	output, err := cmd.CombinedOutput()
	if err != nil {
		glog.Info("error running %s:", human)
		glog.Info(string(output))
		return string(output), fmt.Errorf("error running kubectl")
	}

	return string(output), err
}
