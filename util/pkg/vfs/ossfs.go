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
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/denverdino/aliyungo/oss"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/hashing"
)

// OSSPath is a vfs path for Aliyun Open Storage Service
type OSSPath struct {
	client *oss.Client
	bucket string
	hash   string
	key    string
}

var _ Path = &OSSPath{}
var _ HasHash = &OSSPath{}

// ossReadBackoff is the backoff strategy for Aliyun OSS read retries.
var ossReadBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    4,
}

// ossWriteBackoff is the backoff strategy for Aliyun OSS write retries
var ossWriteBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    5,
}

type listOption struct {
	prefix string
	delim  string
	marker string
	max    int
}

// WriteTo implements io.WriteTo
func (p *OSSPath) WriteTo(out io.Writer) (int64, error) {
	klog.V(4).Infof("Reading file %q", p)

	b := p.client.Bucket(p.bucket)
	headers := http.Header{}

	response, err := b.GetResponseWithHeaders(p.key, headers)
	if err != nil {
		if isOSSNotFound(err) {
			return 0, os.ErrNotExist
		}
		return 0, fmt.Errorf("error fetching %s: %v", p, err)
	}
	defer response.Body.Close()

	n, err := io.Copy(out, response.Body)
	if err != nil {
		return n, fmt.Errorf("error reading %s: %v", p, err)
	}
	return n, nil
}

func (p *OSSPath) Join(relativePath ...string) Path {
	args := []string{p.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &OSSPath{
		client: p.client,
		bucket: p.bucket,
		key:    joined,
	}
}

func (p *OSSPath) ReadFile() ([]byte, error) {
	var b bytes.Buffer
	done, err := RetryWithBackoff(ossReadBackoff, func() (bool, error) {
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

func (p *OSSPath) WriteFile(data io.ReadSeeker, acl ACL) error {
	b := p.client.Bucket(p.bucket)

	done, err := RetryWithBackoff(ossWriteBackoff, func() (bool, error) {
		klog.V(4).Infof("Writing file %q", p)

		var perm oss.ACL
		var ok bool
		if acl != nil {
			perm, ok = acl.(oss.ACL)
			if !ok {
				return true, fmt.Errorf("write to %s with ACL of unexpected type %T", p, acl)
			}
		} else {
			// Private currently is the default ACL
			perm = oss.Private
		}

		if _, err := data.Seek(0, 0); err != nil {
			return false, fmt.Errorf("error seeking to start of data stream for write to %s: %v", p, err)
		}

		bytes, err := ioutil.ReadAll(data)
		if err != nil {
			return false, fmt.Errorf("error reading from data stream: %v", err)
		}

		contType := "application/octet-stream"
		err = b.Put(p.key, bytes, contType, perm, oss.Options{})
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
var createFileLockOSS sync.Mutex

func (p *OSSPath) CreateFile(data io.ReadSeeker, acl ACL) error {
	createFileLockOSS.Lock()
	defer createFileLockOSS.Unlock()

	// Check if exists
	b := p.client.Bucket(p.bucket)
	exist, err := b.Exists(p.key)
	if err != nil {
		return err
	}
	if exist {
		return os.ErrExist
	}

	return p.WriteFile(data, acl)
}

func (p *OSSPath) Remove() error {
	b := p.client.Bucket(p.bucket)

	done, err := RetryWithBackoff(ossWriteBackoff, func() (bool, error) {
		klog.V(8).Infof("removing file %s", p)

		err := b.Del(p.key)
		if err != nil {
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

func (p *OSSPath) RemoveAllVersions() error {
	return p.Remove()
}

func (p *OSSPath) Base() string {
	return path.Base(p.key)
}

func (p *OSSPath) String() string {
	return p.Path()
}

func (p *OSSPath) Path() string {
	return "oss://" + p.bucket + "/" + p.key
}

func (p *OSSPath) Bucket() string {
	return p.bucket
}

func (p *OSSPath) Key() string {
	return p.key
}

func (p *OSSPath) ReadDir() ([]Path, error) {
	prefix := p.key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	// OSS can return at most 1000 paths(keys + common prefixes) at a time
	opt := listOption{
		prefix: prefix,
		delim:  "/",
		marker: "",
		max:    1000,
	}

	return p.listPath(opt)
}

func (p *OSSPath) ReadTree() ([]Path, error) {
	prefix := p.key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	// OSS can return at most 1000 paths(keys + common prefixes) at a time
	opt := listOption{
		prefix: prefix,
		// No delimiter for recursive search
		delim:  "",
		marker: "",
		max:    1000,
	}

	return p.listPath(opt)
}

func (p *OSSPath) PreferredHash() (*hashing.Hash, error) {
	return p.Hash(hashing.HashAlgorithmMD5)
}

func (p *OSSPath) Hash(a hashing.HashAlgorithm) (*hashing.Hash, error) {
	if a != hashing.HashAlgorithmMD5 {
		return nil, nil
	}

	md5 := p.hash
	if md5 == "" {
		return nil, nil
	}

	md5Bytes, err := hex.DecodeString(md5)
	if err != nil {
		return nil, fmt.Errorf("Etag was not a valid MD5 sum: %q", md5)
	}

	return &hashing.Hash{Algorithm: hashing.HashAlgorithmMD5, HashValue: md5Bytes}, nil
}

func (p *OSSPath) listPath(opt listOption) ([]Path, error) {
	var ret []Path
	b := p.client.Bucket(p.bucket)

	done, err := RetryWithBackoff(ossReadBackoff, func() (bool, error) {

		var paths []Path
		for {
			// OSS can return at most 1000 paths(keys + common prefixes) at a time
			resp, err := b.List(opt.prefix, opt.delim, opt.marker, opt.max)
			if err != nil {
				if isOSSNotFound(err) {
					return true, os.ErrNotExist
				}
				return false, fmt.Errorf("error listing %s: %v", p, err)
			}

			if len(resp.Contents) != 0 || len(resp.CommonPrefixes) != 0 {
				// Contents represent files
				for _, k := range resp.Contents {
					child := &OSSPath{
						client: p.client,
						bucket: p.bucket,
						key:    k.Key,
					}
					paths = append(paths, child)
				}
				if len(resp.Contents) != 0 {
					// start with the last key in next iteration of listing.
					opt.marker = resp.Contents[len(resp.Contents)-1].Key
				}

				// CommonPrefixes represent directories
				for _, d := range resp.CommonPrefixes {
					child := &OSSPath{
						client: p.client,
						bucket: p.bucket,
						key:    d,
					}
					paths = append(paths, child)
				}
				if len(resp.CommonPrefixes) != 0 {
					lastComPref := resp.CommonPrefixes[len(resp.CommonPrefixes)-1]
					if strings.Compare(lastComPref, opt.marker) == 1 {
						opt.marker = lastComPref
					}
				}
			} else {
				// no more files or directories
				break
			}
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

func isOSSNotFound(err error) bool {
	if err == nil {
		return false
	}
	ossErr, ok := err.(*oss.Error)
	return ok && ossErr.StatusCode == 404
}
