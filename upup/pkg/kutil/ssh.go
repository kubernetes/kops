/*
Copyright 2019 The Kubernetes Authors.

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

package kutil

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
	"k8s.io/kops/util/pkg/vfs"
)

type NodeSSH struct {
	Hostname  string
	SSHConfig ssh.ClientConfig
	sshClient *ssh.Client
}

func (m *NodeSSH) Root() (*vfs.SSHPath, error) {
	client, err := m.GetSSHClient()
	if err != nil {
		return nil, err
	}
	sudo := true
	return vfs.NewSSHPath(client, m.Hostname, "/", sudo), nil
}

func AddSSHIdentity(sshConfig *ssh.ClientConfig, p string) error {
	a, err := parsePrivateKeyFile(p)
	if err != nil {
		return err
	}
	sshConfig.Auth = append(sshConfig.Auth, a)
	return nil
}

func (m *NodeSSH) dial() (*ssh.Client, error) {
	users := []string{"admin", "ubuntu"}
	if m.SSHConfig.User != "" {
		users = []string{m.SSHConfig.User}
	}

	var lastError error
	for _, user := range users {
		m.SSHConfig.User = user
		sshClient, err := ssh.Dial("tcp", m.Hostname+":22", &m.SSHConfig)
		if err == nil {
			return sshClient, err
		}
		lastError = err
	}

	return nil, fmt.Errorf("error connecting to SSH on server %q: %v", m.Hostname, lastError)
}

func (m *NodeSSH) GetSSHClient() (*ssh.Client, error) {
	if m.sshClient == nil {
		sshClient, err := m.dial()
		if err != nil {
			return nil, err
		}
		m.sshClient = sshClient
	}
	return m.sshClient, nil
}

//func (m *NodeSSH) ReadFile(remotePath string) ([]byte, error) {
//	b, err := m.exec("sudo cat " + remotePath)
//	if err != nil {
//		return nil, fmt.Errorf("error reading remote file %q: %v", remotePath, err)
//	}
//	return b, nil
//}

//func (m *NodeSSH) exec(cmd string) ([]byte, error) {
//	client, err := m.GetSSHClient()
//	if err != nil {
//		return nil, err
//	}
//
//	session, err := client.NewSession()
//	if err != nil {
//		return nil, fmt.Errorf("error creating SSH session: %v", err)
//	}
//	defer session.Close()
//
//	b, err := session.Output(cmd)
//	if err != nil {
//		return nil, fmt.Errorf("error executing command %q: %v", cmd, err)
//	}
//	return b, nil
//}

func parsePrivateKeyFile(p string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("error reading SSH key file %q: %v", p, err)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("error parsing key file %q: %v", p, err)
	}
	return ssh.PublicKeys(key), nil
}
