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

package vfs

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/googleapi"
	storage "google.golang.org/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/hashing"
)

// GSPath is a vfs path for Google Cloud Storage
type GSPath struct {
	client  *storage.Service
	bucket  string
	key     string
	md5Hash string
}

var _ Path = &GSPath{}
var _ HasHash = &GSPath{}

// gcsReadBackoff is the backoff strategy for GCS read retries
var gcsReadBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    4,
}

// GSAcl is an ACL implementation for objects on Google Cloud Storage
type GSAcl struct {
	Acl []*storage.ObjectAccessControl
}

func (a *GSAcl) String() string {
	var s []string
	for _, acl := range a.Acl {
		s = append(s, fmt.Sprintf("%+v", acl))
	}

	return "{" + strings.Join(s, ", ") + "}"
}

var _ ACL = &GSAcl{}

// gcsWriteBackoff is the backoff strategy for GCS write retries
var gcsWriteBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    5,
}

func NewGSPath(client *storage.Service, bucket string, key string) *GSPath {
	bucket = strings.TrimSuffix(bucket, "/")
	key = strings.TrimPrefix(key, "/")

	return &GSPath{
		client: client,
		bucket: bucket,
		key:    key,
	}
}

func (p *GSPath) Path() string {
	return "gs://" + p.bucket + "/" + p.key
}

func (p *GSPath) Bucket() string {
	return p.bucket
}

func (p *GSPath) Object() string {
	return p.key
}

// Client returns the storage.Service bound to this path
func (p *GSPath) Client() *storage.Service {
	return p.client
}

func (p *GSPath) String() string {
	return p.Path()
}

func (p *GSPath) Remove() error {
	done, err := RetryWithBackoff(gcsWriteBackoff, func() (bool, error) {
		err := p.client.Objects.Delete(p.bucket, p.key).Do()
		if err != nil {
			// TODO: Check for not-exists, return os.NotExist

			return false, fmt.Errorf("error deleting %s: %v", p, err)
		}

		return true, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		// Shouldn't happen - we always return a non-nil error with false
		return wait.ErrWaitTimeout
	}
}

func (p *GSPath) RemoveAllVersions() error {
	return p.Remove()
}

func (p *GSPath) Join(relativePath ...string) Path {
	args := []string{p.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &GSPath{
		client: p.client,
		bucket: p.bucket,
		key:    joined,
	}
}

func (p *GSPath) WriteFile(data io.ReadSeeker, acl ACL) error {
	md5Hash, err := hashing.HashAlgorithmMD5.Hash(data)
	if err != nil {
		return err
	}

	done, err := RetryWithBackoff(gcsWriteBackoff, func() (bool, error) {
		obj := &storage.Object{
			Name:    p.key,
			Md5Hash: base64.StdEncoding.EncodeToString(md5Hash.HashValue),
		}

		if acl != nil {
			gsACL, ok := acl.(*GSAcl)
			if !ok {
				return true, fmt.Errorf("write to %s with ACL of unexpected type %T", p, acl)
			}
			obj.Acl = gsACL.Acl
			klog.V(4).Infof("Writing file %q with ACL %v", p, gsACL)
		} else {
			klog.V(4).Infof("Writing file %q", p)
		}

		if _, err := data.Seek(0, 0); err != nil {
			return false, fmt.Errorf("error seeking to start of data stream for write to %s: %v", p, err)
		}

		_, err = p.client.Objects.Insert(p.bucket, obj).Media(data).Do()
		if err != nil {
			return false, fmt.Errorf("error writing %s: %v", p, err)
		}

		return true, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		// Shouldn't happen - we always return a non-nil error with false
		return wait.ErrWaitTimeout
	}
}

// To prevent concurrent creates on the same file while maintaining atomicity of writes,
// we take a process-wide lock during the operation.
// Not a great approach, but fine for a single process (with low concurrency)
// TODO: should we enable versioning?
var createFileLockGCS sync.Mutex

func (p *GSPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	createFileLockGCS.Lock()
	defer createFileLockGCS.Unlock()

	// Check if exists
	_, err := p.ReadFile()
	if err == nil {
		return os.ErrExist
	}

	if !os.IsNotExist(err) {
		return err
	}

	return p.WriteFile(data, acl)
}

// ReadFile implements Path::ReadFile
func (p *GSPath) ReadFile() ([]byte, error) {
	var b bytes.Buffer
	done, err := RetryWithBackoff(gcsReadBackoff, func() (bool, error) {
		b.Reset()
		_, err := p.WriteTo(&b)
		if err != nil {
			if os.IsNotExist(err) {
				// Not recoverable
				return true, err
			}
			return false, err
		}
		// Success!
		return true, nil
	})
	if err != nil {
		return nil, err
	} else if done {
		return b.Bytes(), nil
	} else {
		// Shouldn't happen - we always return a non-nil error with false
		return nil, wait.ErrWaitTimeout
	}
}

// WriteTo implements io.WriterTo::WriteTo
func (p *GSPath) WriteTo(out io.Writer) (int64, error) {
	klog.V(4).Infof("Reading file %q", p)

	response, err := p.client.Objects.Get(p.bucket, p.key).Download()
	if err != nil {
		if isGCSNotFound(err) {
			return 0, os.ErrNotExist
		}
		return 0, fmt.Errorf("error reading %s: %v", p, err)
	}
	if response == nil {
		return 0, fmt.Errorf("no response returned from reading %s", p)
	}
	defer response.Body.Close()

	return io.Copy(out, response.Body)
}

// ReadDir implements Path::ReadDir
func (p *GSPath) ReadDir() ([]Path, error) {
	var ret []Path
	done, err := RetryWithBackoff(gcsReadBackoff, func() (bool, error) {
		prefix := p.key
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		ctx := context.Background()
		var paths []Path
		err := p.client.Objects.List(p.bucket).Delimiter("/").Prefix(prefix).Pages(ctx, func(page *storage.Objects) error {
			for _, o := range page.Items {
				child := &GSPath{
					client:  p.client,
					bucket:  p.bucket,
					key:     o.Name,
					md5Hash: o.Md5Hash,
				}
				paths = append(paths, child)
			}
			return nil
		})
		if err != nil {
			if isGCSNotFound(err) {
				return true, os.ErrNotExist
			}
			return false, fmt.Errorf("error listing %s: %v", p, err)
		}
		klog.V(8).Infof("Listed files in %v: %v", p, paths)
		ret = paths
		return true, nil
	})
	if err != nil {
		return nil, err
	} else if done {
		return ret, nil
	} else {
		// Shouldn't happen - we always return a non-nil error with false
		return nil, wait.ErrWaitTimeout
	}
}

// ReadTree implements Path::ReadTree
func (p *GSPath) ReadTree() ([]Path, error) {
	var ret []Path
	done, err := RetryWithBackoff(gcsReadBackoff, func() (bool, error) {
		// No delimiter for recursive search
		prefix := p.key
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		ctx := context.Background()
		var paths []Path
		err := p.client.Objects.List(p.bucket).Prefix(prefix).Pages(ctx, func(page *storage.Objects) error {
			for _, o := range page.Items {
				key := o.Name
				child := &GSPath{
					client:  p.client,
					bucket:  p.bucket,
					key:     key,
					md5Hash: o.Md5Hash,
				}
				paths = append(paths, child)
			}
			return nil
		})
		if err != nil {
			if isGCSNotFound(err) {
				return true, os.ErrNotExist
			}
			return false, fmt.Errorf("error listing tree %s: %v", p, err)
		}
		ret = paths
		return true, nil
	})
	if err != nil {
		return nil, err
	} else if done {
		return ret, nil
	} else {
		// Shouldn't happen - we always return a non-nil error with false
		return nil, wait.ErrWaitTimeout
	}
}

func (p *GSPath) Base() string {
	return path.Base(p.key)
}

func (p *GSPath) PreferredHash() (*hashing.Hash, error) {
	return p.Hash(hashing.HashAlgorithmMD5)
}

func (p *GSPath) Hash(a hashing.HashAlgorithm) (*hashing.Hash, error) {
	if a != hashing.HashAlgorithmMD5 {
		return nil, nil
	}

	md5 := p.md5Hash
	if md5 == "" {
		return nil, nil
	}

	md5Bytes, err := base64.StdEncoding.DecodeString(md5)
	if err != nil {
		return nil, fmt.Errorf("Etag was not a valid MD5 sum: %q", md5)
	}

	return &hashing.Hash{Algorithm: hashing.HashAlgorithmMD5, HashValue: md5Bytes}, nil
}

func isGCSNotFound(err error) bool {
	if err == nil {
		return false
	}
	ae, ok := err.(*googleapi.Error)
	return ok && ae.Code == http.StatusNotFound
}
