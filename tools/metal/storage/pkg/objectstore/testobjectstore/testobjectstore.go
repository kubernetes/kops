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
	"net/http"
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

func (m *TestObjectStore) ListBuckets(ctx context.Context) []objectstore.BucketInfo {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	buckets := make([]objectstore.BucketInfo, 0, len(m.buckets))
	for _, bucket := range m.buckets {
		buckets = append(buckets, bucket.info)
	}
	return buckets
}

func (m *TestObjectStore) GetBucket(ctx context.Context, bucketName string) (objectstore.Bucket, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	bucket := m.buckets[bucketName]
	if bucket == nil {
		return nil, nil
	}
	return bucket, nil
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
			CreationDate: time.Now(),
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

type TestObject struct {
	data []byte
	info objectstore.ObjectInfo
}

var _ objectstore.Object = &TestObject{}

func (o *TestObject) WriteTo(w http.ResponseWriter) error {
	_, err := w.Write(o.data)

	return err
}
