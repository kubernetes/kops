/*
Copyright 2026 The Kubernetes Authors.

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

package builder

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/azure"
	"k8s.io/kops/util/pkg/vfs"
)

// azureBlobStore is an Azure Blob Storage container that the staged artifacts are uploaded to
// through the kops VFS, which authenticates like kops does, and that serves them over https://.
type azureBlobStore struct {
	account   string
	container string
	prefix    string
}

// azureBlobStore returns the azureBlobStore of the azureblob:// stage location.
func (b *BuildOptions) azureBlobStore() (*azureBlobStore, error) {
	account, container, prefix, err := azure.SplitBlobURL(b.StageLocation)
	if err != nil {
		return nil, err
	}
	return &azureBlobStore{account: account, container: container, prefix: prefix}, nil
}

func (s *azureBlobStore) stagePrefix() string {
	return s.prefix
}

// sync uploads the files below dir to the blobs served under urlPath.
func (s *azureBlobStore) sync(dir, urlPath string) error {
	return filepath.WalkDir(dir, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, file)
		if err != nil {
			return err
		}
		return s.copy(file, path.Join(urlPath, filepath.ToSlash(rel)))
	})
}

// copy uploads file to the blob served at urlPath.
func (s *azureBlobStore) copy(file, urlPath string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	p, err := vfs.Context.BuildVfsPath(s.uri(urlPath))
	if err != nil {
		return err
	}
	klog.Infof("Uploading %s to %s", file, p)
	if err := p.WriteFile(context.Background(), f, nil); err != nil {
		return fmt.Errorf("uploading %s to %s: %w", file, p, err)
	}
	return nil
}

// uri returns the azureblob:// URI of the blob served at urlPath.
func (s *azureBlobStore) uri(urlPath string) string {
	return "azureblob://" + s.account + "/" + s.container + "/" + strings.Trim(urlPath, "/")
}

func (s *azureBlobStore) objectURL(_ context.Context, urlPath string) (string, error) {
	return azure.PublicObjectURL(s.account, s.container, urlPath), nil
}
