package kutil

import (
	"encoding/json"
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

func (k *Kubectl) GetConfig(minify bool) (*KubectlConfig, error) {
	output := "json"
	// TODO: --context doesn't seem to work
	args := []string{"config", "view"}

	if minify {
		args = append(args, "--minify")
	}

	if output != "" {
		args = append(args, "--output", output)
	}

	configString, err := k.execKubectl(args...)
	if err != nil {
		return nil, err
	}
	configString = strings.TrimSpace(configString)

	glog.V(8).Infof("config = %q", configString)

	config := &KubectlConfig{}
	err = json.Unmarshal([]byte(configString), config)
	if err != nil {
		return nil, fmt.Errorf("cannot parse current config from kubectl: %v", err)
	}

	return config, nil
}

func (k *Kubectl) execKubectl(args ...string) (string, error) {
	kubectlPath := k.KubectlPath
	if kubectlPath == "" {
		kubectlPath = "kubectl" // Assume in PATH
	}
	cmd := exec.Command(kubectlPath, args...)
	env := os.Environ()
	cmd.Env = env

	human := strings.Join(cmd.Args, " ")
	glog.V(2).Infof("Running command: %s", human)
	output, err := cmd.CombinedOutput()
	if err != nil {
		glog.Info("error running %s:", human)
		glog.Info(string(output))
		return string(output), fmt.Errorf("error running kubectl")
	}

	return string(output), err
}

type KubectlConfig struct {
	Kind           string                    `json:"kind"`
	ApiVersion     string                    `json:"apiVersion"`
	CurrentContext string                    `json:"current-context"`
	Clusters       []*KubectlClusterWithName `json:"clusters"`
	Contexts       []*KubectlContextWithName `json:"contexts"`
	Users          []*KubectlUserWithName    `json:"users"`
}

type KubectlClusterWithName struct {
	Name    string         `json:"name"`
	Cluster KubectlCluster `json:"cluster"`
}

type KubectlCluster struct {
	Server string `json:"server"`
}

type KubectlContextWithName struct {
	Name    string         `json:"name"`
	Context KubectlContext `json:"context"`
}

type KubectlContext struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

type KubectlUserWithName struct {
	Name string      `json:"name"`
	User KubectlUser `json:"user"`
}

type KubectlUser struct {
	ClientCertificateData string `json:"client-certificate-data"`
	ClientKeyData         string `json:"client-key-data"`
	Password              string `json:"password"`
	Username              string `json:"username"`
}
