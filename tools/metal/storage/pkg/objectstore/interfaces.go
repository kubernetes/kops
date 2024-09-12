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

package objectstore

import (
	"context"
	"io"
	"net/http"
	"time"
)

type ObjectStore interface {
	ListBuckets(ctx context.Context) ([]BucketInfo, error)

	// GetBucket returns the bucket with the given name.
	// If the bucket does not exist, it returns (nil, nil).
	GetBucket(ctx context.Context, name string) (Bucket, *BucketInfo, error)

	// CreateBucket creates the bucket with the given name.
	// If the bucket already exist, it returns codes.AlreadyExists.
	CreateBucket(ctx context.Context, name string) (*BucketInfo, error)
}

type BucketInfo struct {
	Name         string
	CreationDate time.Time

	// Owner is the AWS ID of the bucket owner.
	Owner string
}

type Bucket interface {
	// GetObject returns the object with the given key.
	// If the object does not exist, it returns (nil, nil).
	GetObject(ctx context.Context, key string) (Object, error)

	// PutObject creates the object with the given key.
	PutObject(ctx context.Context, key string, r io.Reader) (*ObjectInfo, error)

	// ListObjects returns the list of objects in the bucket.
	ListObjects(ctx context.Context) ([]ObjectInfo, error)
}

type ObjectInfo struct {
	Key          string
	LastModified time.Time
	Size         int64
}

type Object interface {
	WriteTo(req *http.Request, w http.ResponseWriter) error
}
