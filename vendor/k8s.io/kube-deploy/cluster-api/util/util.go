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

package util

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api/client"
)

const (
	TypeMaster = "Master"
)

func Contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func IsMaster(machine *clusterv1.Machine) bool {
	return Contains(TypeMaster, machine.Spec.Roles)
}

func IsNodeReady(node *v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			return condition.Status == v1.ConditionTrue
		}
	}

	return false
}

func ExecCommand(name string, args ...string) string {
	cmdOut, _ := exec.Command(name, args...).Output()
	return string(cmdOut)
}

func Copy(m *clusterv1.Machine) *clusterv1.Machine {
	ret := &clusterv1.Machine{}
	ret.APIVersion = m.APIVersion
	ret.Kind = m.Kind
	ret.ClusterName = m.ClusterName
	ret.GenerateName = m.GenerateName
	ret.Name = m.Name
	ret.Namespace = m.Namespace
	m.Spec.DeepCopyInto(&ret.Spec)
	return ret
}

func Home() string {
	home := os.Getenv("HOME")
	if strings.Contains(home, "root") {
		return "/root"
	}

	usr, err := user.Current()
	if err != nil {
		glog.Warningf("unable to find user: %v", err)
		return ""
	}
	return usr.HomeDir
}

func GetDefaultKubeConfigPath() string {
	localDir := fmt.Sprintf("%s/.kube", Home())
	if _, err := os.Stat(localDir); os.IsNotExist(err) {
		if err := os.Mkdir(localDir, 0777); err != nil {
			glog.Fatal(err)
		}
	}
	return fmt.Sprintf("%s/config", localDir)
}

func NewApiClient(configPath string) (*client.ClusterAPIV1Alpha1Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	c, err := client.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return c, nil
}