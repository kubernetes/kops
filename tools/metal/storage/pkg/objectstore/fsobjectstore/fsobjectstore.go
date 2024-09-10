/*
Copyright 2024 The Kubernetes Authors.

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

package fsobjectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kubernetes/kops/tools/metal/dhcp/pkg/objectstore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

type FilesystemObjectStore struct {
	mutex   sync.Mutex
	basedir string
}

func New(basedir string) *FilesystemObjectStore {
	return &FilesystemObjectStore{
		basedir: basedir,
	}
}

var _ objectstore.ObjectStore = &FilesystemObjectStore{}

func (m *FilesystemObjectStore) ListBuckets(ctx context.Context) ([]objectstore.BucketInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	entries, err := os.ReadDir(m.basedir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", m.basedir, err)
	}
	buckets := make([]objectstore.BucketInfo, 0, len(entries))
	for _, entry := range entries {
		bucketInfo, err := m.buildBucketInfo(ctx, entry.Name())
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, *bucketInfo)
	}
	return buckets, nil
}

func (m *FilesystemObjectStore) buildBucketInfo(ctx context.Context, bucketName string) (*objectstore.BucketInfo, error) {
	p := filepath.Join(m.basedir, bucketName)
	stat, err := os.Stat(p)
	if err != nil {
		return nil, fmt.Errorf("getting info for directory %q: %w", p, err)
	}
	sysStat := stat.Sys().(*syscall.Stat_t)

	bucketInfo := &objectstore.BucketInfo{
		Name:         stat.Name(),
		CreationDate: time.Unix(sysStat.Ctim.Sec, sysStat.Ctim.Nsec),
		Owner:        getOwnerID(ctx),
	}
	return bucketInfo, nil
}

func (m *FilesystemObjectStore) GetBucket(ctx context.Context, bucketName string) (objectstore.Bucket, *objectstore.BucketInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bucketInfo, err := m.buildBucketInfo(ctx, bucketName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	bucketDir := filepath.Join(m.basedir, bucketName)
	return &FilesystemBucket{basedir: bucketDir, bucketName: bucketName}, bucketInfo, nil
}

// getOwnerID returns the owner ID for the given context.
// This is a fake implementation for testing purposes.
func getOwnerID(ctx context.Context) string {
	return "fake-owner"
}

func (m *FilesystemObjectStore) CreateBucket(ctx context.Context, bucketName string) (*objectstore.BucketInfo, error) {
	log := klog.FromContext(ctx)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	bucketInfo, err := m.buildBucketInfo(ctx, bucketName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// OK
		} else {
			return nil, err
		}
	}
	if bucketInfo != nil {
		// return nil, status.Errorf(codes.AlreadyExists, "bucket %q already exists", bucketName)
		err := status.Errorf(codes.AlreadyExists, "bucket %q already exists", bucketName)
		code := status.Code(err)
		log.Error(err, "failed to create bucket", "code", code)
		return nil, err
	}

	p := filepath.Join(m.basedir, bucketName)
	if err := os.Mkdir(p, 0700); err != nil {
		return nil, fmt.Errorf("creating directory for bucket %q: %w", p, err)
	}

	bucketInfo, err = m.buildBucketInfo(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	return bucketInfo, nil
}

type FilesystemBucket struct {
	basedir    string
	bucketName string
	mutex      sync.Mutex
}

var _ objectstore.Bucket = &FilesystemBucket{}

func (m *FilesystemBucket) ListObjects(ctx context.Context) ([]objectstore.ObjectInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	prefix := m.basedir
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	var objects []objectstore.ObjectInfo
	if err := filepath.WalkDir(m.basedir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasPrefix(path, prefix) {
			return fmt.Errorf("unexpected path walking %q, expected prefix %q: %q", m.basedir, prefix, path)
		}
		key := strings.TrimPrefix(path, prefix)
		objectInfo, err := m.buildObjectInfo(ctx, key)
		if err != nil {
			return err
		}
		objects = append(objects, *objectInfo)
		return nil
	}); err != nil {
		return nil, err
	}

	return objects, nil
}

func (m *FilesystemBucket) GetObject(ctx context.Context, key string) (objectstore.Object, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	p := m.pathForKey(key)

	_, err := m.buildObjectInfo(ctx, key)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	return &FilesystemObject{bucket: m, path: p, key: key}, nil
}

func (m *FilesystemBucket) buildObjectInfo(ctx context.Context, key string) (*objectstore.ObjectInfo, error) {
	p := filepath.Join(m.basedir, key)
	stat, err := os.Stat(p)
	if err != nil {
		return nil, fmt.Errorf("getting info for file %q: %w", p, err)
	}
	objectInfo := &objectstore.ObjectInfo{
		Key:          key,
		LastModified: stat.ModTime(),
		Size:         stat.Size(),
	}
	return objectInfo, nil
}

func (m *FilesystemBucket) pathForKey(key string) string {
	p := filepath.Join(m.basedir, key)
	return p
}

func (m *FilesystemBucket) PutObject(ctx context.Context, key string, r io.Reader) (*objectstore.ObjectInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	p := filepath.Join(m.basedir, key)

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("making directories %q: %w", dir, err)
	}
	f, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("creating file %q: %w", p, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return nil, fmt.Errorf("writing data: %w", err)
	}

	objectInfo, err := m.buildObjectInfo(ctx, key)
	if err != nil {
		return nil, err
	}

	return objectInfo, nil
}

type FilesystemObject struct {
	bucket *FilesystemBucket
	key    string
	path   string
}

var _ objectstore.Object = &FilesystemObject{}

func (o *FilesystemObject) WriteTo(r *http.Request, w http.ResponseWriter) error {
	f, err := os.Open(o.path)
	if err != nil {
		return fmt.Errorf("opening file %q: %w", o.path, err)
	}
	defer f.Close()

	stat, err := os.Stat(o.path)
	if err != nil {
		return fmt.Errorf("getting stat for file %q: %w", o.path, err)
	}
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		if _, err := io.Copy(w, f); err != nil {
			return fmt.Errorf("reading file %q: %w", o.path, err)
		}
	}
	return nil
}
