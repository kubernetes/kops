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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/hashing"
)

// AzureBlobPath is a path in the VFS space backed by Azure Blob.
type AzureBlobPath struct {
	vfsContext *VFSContext
	container  string
	key        string
	md5Hash    string
}

var (
	_ Path    = &AzureBlobPath{}
	_ HasHash = &AzureBlobPath{}
)

// NewAzureBlobPath returns a new AzureBlobPath.
func NewAzureBlobPath(vfsContext *VFSContext, container string, key string) *AzureBlobPath {
	return &AzureBlobPath{
		vfsContext: vfsContext,
		container:  strings.TrimSuffix(container, "/"),
		key:        strings.TrimPrefix(key, "/"),
	}
}

// Base returns the base name (last element).
func (p *AzureBlobPath) Base() string {
	return path.Base(p.key)
}

// PreferredHash returns the hash of the file contents, with the preferred hash algorithm.
func (p *AzureBlobPath) PreferredHash() (*hashing.Hash, error) {
	return p.Hash(hashing.HashAlgorithmMD5)
}

// Hash gets the hash, or nil if the hash cannot be (easily) computed.
func (p *AzureBlobPath) Hash(a hashing.HashAlgorithm) (*hashing.Hash, error) {
	if a != hashing.HashAlgorithmMD5 {
		return nil, nil
	}

	if p.md5Hash == "" {
		return nil, nil
	}

	md5Bytes, err := base64.StdEncoding.DecodeString(p.md5Hash)
	if err != nil {
		return nil, fmt.Errorf("not valid MD5 sum: %q", p.md5Hash)
	}

	return &hashing.Hash{Algorithm: hashing.HashAlgorithmMD5, HashValue: md5Bytes}, nil
}

// Path returns a string representing the full path.
func (p *AzureBlobPath) Path() string {
	return fmt.Sprintf("azureblob://%s/%s", p.container, p.key)
}

// Join returns a new path that joins the current path and given relative paths.
func (p *AzureBlobPath) Join(relativePath ...string) Path {
	args := []string{p.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &AzureBlobPath{
		vfsContext: p.vfsContext,
		container:  p.container,
		key:        joined,
	}
}

// ReadFile returns the content of the blob.
func (p *AzureBlobPath) ReadFile(ctx context.Context) ([]byte, error) {
	klog.V(8).Infof("Reading file: %s - %s", p.container, p.key)

	client, err := p.getClient(ctx)
	if err != nil {
		return nil, err
	}

	get, err := client.DownloadStream(ctx, p.container, p.key, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.ContainerNotFound) || bloberror.HasCode(err, bloberror.BlobNotFound) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}

	b := &bytes.Buffer{}
	retryReader := get.NewRetryReader(ctx, &azblob.RetryReaderOptions{})
	_, err = b.ReadFrom(retryReader)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// WriteTo writes the content of the blob to the writer.
func (p *AzureBlobPath) WriteTo(w io.Writer) (int64, error) {
	klog.V(8).Infof("Writing to: %s - %s", p.container, p.key)

	ctx := context.TODO()

	b, err := p.ReadFile(ctx)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(b)
	if err != nil {
		return 0, err
	}

	return int64(n), err
}

// createFileLockAzureBLob prevents concurrent creates on the same
// file while maintaining atomicity of writes.
var createFileLockAzureBlob sync.Mutex

// CreateFile writes the file contents only if the file does not already exist.
func (p *AzureBlobPath) CreateFile(ctx context.Context, data io.ReadSeeker, acl ACL) error {
	klog.V(8).Infof("Creating file: %s - %s", p.container, p.key)

	createFileLockAzureBlob.Lock()
	defer createFileLockAzureBlob.Unlock()

	// Check if the blob exists.
	_, err := p.ReadFile(ctx)
	if err == nil {
		return os.ErrExist
	}
	if !os.IsNotExist(err) {
		return err
	}
	return p.WriteFile(ctx, data, acl)
}

// WriteFile writes the blob to the reader.
func (p *AzureBlobPath) WriteFile(ctx context.Context, data io.ReadSeeker, acl ACL) error {
	klog.V(8).Infof("Writing file: %s - %s", p.container, p.key)

	client, err := p.getClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.CreateContainer(ctx, p.container, nil)
	if err != nil && !bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
		return err
	}

	_, err = client.UploadStream(ctx, p.container, p.key, data, nil)
	return err
}

// Remove deletes the blob.
func (p *AzureBlobPath) Remove(ctx context.Context) error {
	klog.V(8).Infof("Removing file: %q - %q", p.container, p.key)

	client, err := p.getClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteBlob(ctx, p.container, p.key, nil)
	if err != nil {
		return err
	}

	return nil
}

func (p *AzureBlobPath) RemoveAll(ctx context.Context) error {
	klog.V(8).Infof("Removing ALL files: %s - %s", p.container, p.key)

	tree, err := p.ReadTree(ctx)
	if err != nil {
		return err
	}

	for _, blobPath := range tree {
		err := blobPath.Remove(ctx)
		if err != nil {
			return fmt.Errorf("removing file %s: %w", blobPath, err)
		}
	}

	return nil
}

func (p *AzureBlobPath) RemoveAllVersions(ctx context.Context) error {
	klog.V(8).Infof("Removing ALL file versions: %s - %s", p.container, p.key)

	tree, err := p.ReadTree(ctx)
	if err != nil {
		return err
	}

	for _, blobPath := range tree {
		err := blobPath.Remove(ctx)
		if err != nil {
			return fmt.Errorf("removing file %s: %w", blobPath, err)
		}
	}

	return nil
}

// ReadDir lists the blobs under the current Path.
func (p *AzureBlobPath) ReadDir() ([]Path, error) {
	klog.V(8).Infof("Reading dir: %s - %s", p.container, p.key)

	ctx := context.TODO()

	tree, err := p.ReadTree(ctx)
	if err != nil {
		return nil, err
	}

	var paths []Path
	for _, blob := range tree {
		if p.Join(blob.Base()).Path() == blob.Path() {
			klog.V(8).Infof("Found file: %q", blob.Path())
			paths = append(paths, blob)
		}
	}

	return paths, nil
}

// ReadTree lists all blobs (recursively) in the subtree rooted at the current Path.
func (p *AzureBlobPath) ReadTree(ctx context.Context) ([]Path, error) {
	klog.V(8).Infof("Reading tree: %s - %s", p.container, p.key)

	client, err := p.getClient(ctx)
	if err != nil {
		return nil, err
	}

	var paths []Path
	pager := client.NewListBlobsFlatPager(p.container, &azblob.ListBlobsFlatOptions{
		Prefix: to.Ptr(p.key),
	})
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, blob := range resp.Segment.BlobItems {
			paths = append(paths, &AzureBlobPath{
				vfsContext: p.vfsContext,
				container:  p.container,
				key:        *blob.Name,
			})
		}
	}

	return paths, nil
}

// getClient returns the client for azure blob storage.
func (p *AzureBlobPath) getClient(ctx context.Context) (*azblob.Client, error) {
	return p.vfsContext.getAzureBlobClient(ctx)
}
