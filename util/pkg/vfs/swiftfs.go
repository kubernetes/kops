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

package vfs

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	swiftcontainer "github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	swiftobject "github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"
	"github.com/gophercloud/gophercloud/pagination"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	osauth "k8s.io/kops/upup/pkg/fi/cloudup/openstackauth"

	"k8s.io/kops/util/pkg/hashing"
)

func NewSwiftClient() (*gophercloud.ServiceClient, error) {
	InsecureSkipVerify := true
	provider, err := osauth.NewOpenStackProvider(InsecureSkipVerify)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate to OpenStack: %v", err)
	}

	client, err := openstack.NewObjectStorageV1(provider, osauth.GetOpenStackEndpointOpts())
	if err != nil {
		return nil, fmt.Errorf("failed to get Swift client: %v", err)
	}

	return client, nil
}

// SwiftPath is a vfs path for Openstack Cloud Storage.
type SwiftPath struct {
	client *gophercloud.ServiceClient
	bucket string
	key    string
	hash   string
}

var _ Path = &SwiftPath{}
var _ HasHash = &SwiftPath{}

// swiftReadBackoff is the backoff strategy for Swift read retries.
var swiftReadBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    4,
}

// swiftWriteBackoff is the backoff strategy for Swift write retries.
var swiftWriteBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    5,
}

func NewSwiftPath(client *gophercloud.ServiceClient, bucket string, key string) (*SwiftPath, error) {
	bucket = strings.TrimSuffix(bucket, "/")
	key = strings.TrimPrefix(key, "/")

	return &SwiftPath{
		client: client,
		bucket: bucket,
		key:    key,
	}, nil
}

func (p *SwiftPath) Path() string {
	return "swift://" + p.bucket + "/" + p.key
}

func (p *SwiftPath) Bucket() string {
	return p.bucket
}

func (p *SwiftPath) String() string {
	return p.Path()
}

func (p *SwiftPath) Remove() error {
	done, err := RetryWithBackoff(swiftWriteBackoff, func() (bool, error) {
		opt := swiftobject.DeleteOpts{}
		_, err := swiftobject.Delete(p.client, p.bucket, p.key, opt).Extract()
		if err != nil {
			if isSwiftNotFound(err) {
				return true, os.ErrNotExist
			}
			return false, fmt.Errorf("error deleting %s: %v", p, err)
		}

		return true, nil
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		return wait.ErrWaitTimeout
	}
}

func (p *SwiftPath) RemoveAllVersions() error {
	return p.Remove()
}

func (p *SwiftPath) Join(relativePath ...string) Path {
	args := []string{p.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &SwiftPath{
		client: p.client,
		bucket: p.bucket,
		key:    joined,
	}
}

func (p *SwiftPath) WriteFile(data io.ReadSeeker, acl ACL) error {
	done, err := RetryWithBackoff(swiftWriteBackoff, func() (bool, error) {
		klog.V(4).Infof("Writing file %q", p)
		if _, err := data.Seek(0, 0); err != nil {
			return false, fmt.Errorf("error seeking to start of data stream for %s: %v", p, err)
		}

		createOpts := swiftobject.CreateOpts{Content: data}
		_, err := swiftobject.Create(p.client, p.bucket, p.key, createOpts).Extract()
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
		// Shouldn't happen - we always return a non-nil error with false.
		return wait.ErrWaitTimeout
	}
}

// To prevent concurrent creates on the same file while maintaining atomicity of writes,
// we take a process-wide lock during the operation.
// Not a great approach, but fine for a single process (with low concurrency).
// TODO: should we enable versioning?
var createFileLockSwift sync.Mutex

func (p *SwiftPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	createFileLockSwift.Lock()
	defer createFileLockSwift.Unlock()

	// Check if exists.
	_, err := RetryWithBackoff(swiftReadBackoff, func() (bool, error) {
		klog.V(4).Infof("Getting file %q", p)

		_, err := swiftobject.Get(p.client, p.bucket, p.key, swiftobject.GetOpts{}).Extract()
		if err == nil {
			return true, nil
		} else if isSwiftNotFound(err) {
			return true, os.ErrNotExist
		} else {
			return false, fmt.Errorf("error getting %s: %v", p, err)
		}
	})
	if err == nil {
		return os.ErrExist
	} else if !os.IsNotExist(err) {
		return err
	}

	err = p.createBucket()
	if err != nil {
		return err
	}

	return p.WriteFile(data, acl)
}

func (p *SwiftPath) createBucket() error {
	done, err := RetryWithBackoff(swiftWriteBackoff, func() (bool, error) {
		_, err := swiftcontainer.Get(p.client, p.bucket, swiftcontainer.GetOpts{}).Extract()
		if err == nil {
			return true, nil
		}
		if isSwiftNotFound(err) {
			createOpts := swiftcontainer.CreateOpts{}
			_, err = swiftcontainer.Create(p.client, p.bucket, createOpts).Extract()
			return err == nil, err
		}
		return false, err
	})
	if err != nil {
		return err
	} else if done {
		return nil
	} else {
		// Shouldn't happen - we always return a non-nil error with false.
		return wait.ErrWaitTimeout
	}
}

// ReadFile implements Path::ReadFile
func (p *SwiftPath) ReadFile() ([]byte, error) {
	var b bytes.Buffer
	done, err := RetryWithBackoff(swiftReadBackoff, func() (bool, error) {
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

// WriteTo implements io.WriterTo
func (p *SwiftPath) WriteTo(out io.Writer) (int64, error) {
	klog.V(4).Infof("Reading file %q", p)

	opt := swiftobject.DownloadOpts{}
	result := swiftobject.Download(p.client, p.bucket, p.key, opt)
	if result.Err != nil {
		if isSwiftNotFound(result.Err) {
			return 0, os.ErrNotExist
		}
		return 0, fmt.Errorf("error reading %s: %v", p, result.Err)
	}
	defer result.Body.Close()

	return io.Copy(out, result.Body)
}

func (p *SwiftPath) readPath(opt swiftobject.ListOpts) ([]Path, error) {
	var ret []Path
	done, err := RetryWithBackoff(swiftReadBackoff, func() (bool, error) {
		var paths []Path
		pager := swiftobject.List(p.client, p.bucket, opt)
		err := pager.EachPage(func(page pagination.Page) (bool, error) {
			objects, err1 := swiftobject.ExtractInfo(page)
			if err1 != nil {
				return false, err1
			}
			for _, o := range objects {
				child := &SwiftPath{
					client: p.client,
					bucket: p.bucket,
					key:    o.Name,
					hash:   o.Hash,
				}
				paths = append(paths, child)
			}

			return true, nil
		})
		if err != nil {
			if isSwiftNotFound(err) {
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
		return nil, wait.ErrWaitTimeout
	}
}

// ReadDir implements Path::ReadDir.
func (p *SwiftPath) ReadDir() ([]Path, error) {
	prefix := p.key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	opt := swiftobject.ListOpts{
		Full: true,
		Path: prefix,
	}
	return p.readPath(opt)
}

// ReadTree implements Path::ReadTree.
func (p *SwiftPath) ReadTree() ([]Path, error) {
	prefix := p.key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	opt := swiftobject.ListOpts{
		Full:   true,
		Prefix: prefix,
	}
	return p.readPath(opt)
}

func (p *SwiftPath) Base() string {
	return path.Base(p.key)
}

func (p *SwiftPath) PreferredHash() (*hashing.Hash, error) {
	return p.Hash(hashing.HashAlgorithmMD5)
}

func (p *SwiftPath) Hash(a hashing.HashAlgorithm) (*hashing.Hash, error) {
	if a != hashing.HashAlgorithmMD5 {
		return nil, nil
	}

	md5 := p.hash
	if md5 == "" {
		return nil, nil
	}

	md5Bytes, err := hex.DecodeString(md5)
	if err != nil {
		return nil, fmt.Errorf("etag was not a valid MD5 sum: %q", md5)
	}

	return &hashing.Hash{Algorithm: hashing.HashAlgorithmMD5, HashValue: md5Bytes}, nil
}

func isSwiftNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(gophercloud.ErrDefault404)
	return ok
}
