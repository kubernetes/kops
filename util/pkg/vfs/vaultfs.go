/*
Copyright 2020 The Kubernetes Authors.

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

package vfs

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"k8s.io/klog/v2"
)

type VaultPath struct {
	vaultClient *vault.Client
	scheme      string
	mountPoint  string
	path        string
}

var _ Path = &VaultPath{}

func newVaultPath(client *vault.Client, scheme string, path string) (*VaultPath, error) {
	if scheme != "https://" && scheme != "http://" {
		return nil, fmt.Errorf("scheme must be http:// or https://")
	}

	path = strings.TrimPrefix(path, "/")
	dirs := strings.SplitN(path, "/", 2)
	if len(dirs) != 2 {
		return nil, fmt.Errorf("vault path must have both a mount point and a path. Got: %q", path)
	}

	if client == nil {
		return nil, fmt.Errorf("vault path needs to have a vault client")

	}

	return &VaultPath{
		vaultClient: client,
		scheme:      scheme,
		mountPoint:  dirs[0],
		path:        dirs[1],
	}, nil
}

func (p *VaultPath) WriteFile(data io.ReadSeeker, acl ACL) error {
	klog.V(4).Infof("Writing file %q", p)

	file, err := encodeData(data)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"file": file,
		},
	}

	_, err = p.vaultClient.Logical().Write(p.dataPath(), payload)
	return err
}

func (p *VaultPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	file, _ := p.ReadFile()
	if file == nil {
		return p.WriteFile(data, acl)
	} else {
		return fmt.Errorf("file already exists: %v", p.path)
	}
}

func (p *VaultPath) ReadFile() ([]byte, error) {
	secret, err := p.vaultClient.Logical().Read(p.dataPath())
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data["data"] == nil {
		return nil, os.ErrNotExist
	}

	data := secret.Data["data"].(map[string]interface{})

	encodedString := data["file"].(string)
	return base64.StdEncoding.DecodeString(encodedString)
}

func (p *VaultPath) Remove() error {
	klog.V(8).Infof("removing file %s", p)
	_, err := p.vaultClient.Logical().Delete(p.dataPath())
	return err
}

func (p *VaultPath) RemoveAllVersions() error {
	klog.V(8).Infof("removing all versions of file %s", p)

	data := map[string][]string{
		"versions": {"1"},
	}
	_, err := p.vaultClient.Logical().DeleteWithData(p.dataPath(), data)
	if err != nil {
		return err
	}
	err = p.destroy()
	if err != nil {
		return err
	}
	err = p.deleteMetadata()
	return err
}

func (p *VaultPath) Base() string {
	return path.Base(p.path)
}

func (p *VaultPath) Path() string {
	query := ""
	if p.scheme == "http://" {
		query = "?tls=false"
	}

	return fmt.Sprintf("vault://%s/%s/%s%s", strings.TrimPrefix(p.vaultClient.Address(), p.scheme), p.mountPoint, p.path, query)
}

func (p *VaultPath) ReadDir() ([]Path, error) {
	secret, err := p.vaultClient.Logical().List(p.metadataPath())
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, os.ErrNotExist
	}
	data := secret.Data["keys"].([]interface{})
	paths := make([]Path, 0)
	for _, key := range data {
		path := p.Join(key.(string))
		paths = append(paths, path)
	}
	return paths, nil
}

func (p *VaultPath) ReadTree() ([]Path, error) {
	content, err := p.ReadDir()
	files := make([]Path, 0)
	if os.IsNotExist(err) {
		return files, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading path %v, %v", p, err)
	}
	for _, path := range content {
		if IsDirectory(path) {
			subTree, err := path.ReadTree()
			if err != nil {
				return nil, err
			}
			files = append(files, subTree...)
		} else {
			files = append(files, path)
		}
	}
	return files, nil
}

func (p *VaultPath) Join(relativePath ...string) Path {
	args := []string{p.fullPath()}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	path, _ := newVaultPath(p.vaultClient, p.scheme, joined)
	return path
}

func (p *VaultPath) WriteTo(out io.Writer) (int64, error) {
	panic("writeTo not implemented")
}

func encodeData(data io.ReadSeeker) (string, error) {
	pr, pw := io.Pipe()
	encoder := base64.NewEncoder(base64.StdEncoding, pw)

	go func() {
		_, err := io.Copy(encoder, data)
		encoder.Close()

		if err != nil {
			pw.CloseWithError(err)
		} else {
			pw.Close()
		}
	}()

	out, err := ioutil.ReadAll(pr)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (p *VaultPath) dataPath() string {
	return p.mountPoint + "/data/" + p.path
}

func (p *VaultPath) metadataPath() string {
	return p.mountPoint + "/metadata/" + p.path
}

func (p *VaultPath) fullPath() string {
	return p.mountPoint + "/" + p.path
}

func (p *VaultPath) String() string {
	return p.Path()
}

func (p VaultPath) destroy() error {

	data := map[string][]string{
		"versions": {"1"},
	}
	r := p.vaultClient.NewRequest("PUT", "/v1/"+p.mountPoint+"/destroy/"+p.path)

	r.SetJSONBody(data)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	resp, err := p.vaultClient.RawRequestWithContext(ctx, r)
	if resp != nil {
		defer resp.Body.Close()
	}

	return err

}

func (p VaultPath) deleteMetadata() error {

	r := p.vaultClient.NewRequest("DELETE", "/v1/"+p.mountPoint+"/metadata/"+p.path)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	resp, err := p.vaultClient.RawRequestWithContext(ctx, r)
	if resp != nil {
		defer resp.Body.Close()
	}

	return err

}

func (p VaultPath) SetClientToken(token string) {
	p.vaultClient.SetToken(token)
}
