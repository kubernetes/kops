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

package testobjectstore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kubernetes/kops/tools/metal/dhcp/pkg/objectstore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

type TestObjectStore struct {
	mutex   sync.Mutex
	buckets map[string]*TestBucket
}

func New() *TestObjectStore {
	return &TestObjectStore{
		buckets: make(map[string]*TestBucket),
	}
}

var _ objectstore.ObjectStore = &TestObjectStore{}

func (m *TestObjectStore) ListBuckets(ctx context.Context) ([]objectstore.BucketInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	buckets := make([]objectstore.BucketInfo, 0, len(m.buckets))
	for _, bucket := range m.buckets {
		buckets = append(buckets, bucket.info)
	}
	return buckets, nil
}

func (m *TestObjectStore) GetBucket(ctx context.Context, bucketName string) (objectstore.Bucket, *objectstore.BucketInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bucket := m.buckets[bucketName]
	if bucket == nil {
		return nil, nil, nil
	}
	return bucket, &bucket.info, nil
}

// getOwnerID returns the owner ID for the given context.
// This is a fake implementation for testing purposes.
func getOwnerID(ctx context.Context) string {
	return "fake-owner"
}

func (m *TestObjectStore) CreateBucket(ctx context.Context, bucketName string) (*objectstore.BucketInfo, error) {
	log := klog.FromContext(ctx)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	bucket := m.buckets[bucketName]
	if bucket != nil {
		// return nil, status.Errorf(codes.AlreadyExists, "bucket %q already exists", bucketName)
		err := status.Errorf(codes.AlreadyExists, "bucket %q already exists", bucketName)
		code := status.Code(err)
		log.Error(err, "failed to create bucket", "code", code)
		return nil, err
	}

	bucket = &TestBucket{
		info: objectstore.BucketInfo{
			Name:         bucketName,
			CreationDate: time.Now().UTC(),
			Owner:        getOwnerID(ctx),
		},
		objects: make(map[string]*TestObject),
	}
	m.buckets[bucketName] = bucket
	return &bucket.info, nil
}

type TestBucket struct {
	mutex   sync.Mutex
	info    objectstore.BucketInfo
	objects map[string]*TestObject
}

var _ objectstore.Bucket = &TestBucket{}

func (m *TestBucket) ListObjects(ctx context.Context) ([]objectstore.ObjectInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	objects := make([]objectstore.ObjectInfo, 0, len(m.objects))
	for _, obj := range m.objects {
		objects = append(objects, obj.info)
	}
	return objects, nil
}

func (m *TestBucket) GetObject(ctx context.Context, key string) (objectstore.Object, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	obj := m.objects[key]
	if obj == nil {
		return nil, nil
	}
	return obj, nil
}

func (m *TestBucket) PutObject(ctx context.Context, key string, r io.Reader) (*objectstore.ObjectInfo, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading data: %w", err)
	}

	info := objectstore.ObjectInfo{
		Key:          key,
		LastModified: time.Now().UTC(),
		Size:         int64(len(b)),
	}

	m.objects[key] = &TestObject{
		data: b,
		info: info,
	}
	return &info, nil
}

type TestObject struct {
	data []byte
	info objectstore.ObjectInfo
}

var _ objectstore.Object = &TestObject{}

func (o *TestObject) WriteTo(r *http.Request, w http.ResponseWriter) error {
	w.Header().Set("Content-Length", strconv.Itoa(len(o.data)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, err := w.Write(o.data)

		return err
	}
	return nil
}
