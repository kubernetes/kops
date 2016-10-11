package kutil

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"os"
	"os/exec"
	"strings"
)

type Kubectl struct {
	KubectlPath string
}

//func (k *Kubectl) GetCurrentContext() (string, error) {
//	s, err := k.execKubectl("config", "current-context")
//	if err != nil {
//		return "", err
//	}
//	s = strings.TrimSpace(s)
//	return s, nil
//}

func (k *Kubectl) GetCurrentContext() (string, error) {
	pathOptions := clientcmd.NewDefaultPathOptions()

	config, err := pathOptions.GetStartingConfig()
	if err != nil {
		return "", err
	}

	return config.CurrentContext, nil

	//s, err := k.execKubectl("config", "current-context")
	//if err != nil {
	//	return "", err
	//}
	//s = strings.TrimSpace(s)
	//return s, nil
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

// Apply calls kubectl apply to apply the manifest.
// We will likely in future change this to create things directly (or more likely embed this logic into kubectl itself)
func (k *Kubectl) Apply(context string, data []byte) error {
	localManifestFile, err := ioutil.TempFile("", "manifest")
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}

	defer func() {
		if err := os.Remove(localManifestFile.Name()); err != nil {
			glog.Warningf("error deleting temp file %q: %v", localManifestFile.Name(), err)
		}
	}()

	if err := ioutil.WriteFile(localManifestFile.Name(), data, 0600); err != nil {
		return fmt.Errorf("error writing temp file: %v", err)
	}

	_, err = execKubectl("apply", "--context", context, "-f", localManifestFile.Name())
	return err
}

func execKubectl(args ...string) (string, error) {
	kubectlPath := "kubectl" // Assume in PATH
	cmd := exec.Command(kubectlPath, args...)
	env := os.Environ()
	cmd.Env = env

	human := strings.Join(cmd.Args, " ")
	glog.V(2).Infof("Running command: %s", human)
	output, err := cmd.CombinedOutput()
	if err != nil {
		glog.Infof("error running %s", human)
		glog.Info(string(output))
		return string(output), fmt.Errorf("error running kubectl")
	}

	return string(output), err
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
		glog.Infof("error running %s:", human)
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
	Server                   string `json:"server"`
	CertificateAuthorityData []byte `json:"certificate-authority-data,omitempty"`
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
	ClientCertificateData []byte `json:"client-certificate-data,omitempty"`
	ClientKeyData         []byte `json:"client-key-data,omitempty"`
	Password              string `json:"password,omitempty"`
	Username              string `json:"username,omitempty"`
	Token                 string `json:"token,omitempty"`
}
