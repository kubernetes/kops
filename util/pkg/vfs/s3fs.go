/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kops/util/pkg/hashing"
	"os"
	"path"
	"strings"
	"sync"
)

type S3Path struct {
	s3Context *S3Context
	bucket    string
	region    string
	key       string
	etag      *string
}

var _ Path = &S3Path{}
var _ HasHash = &S3Path{}

func newS3Path(s3Context *S3Context, bucket string, key string) *S3Path {
	bucket = strings.TrimSuffix(bucket, "/")
	key = strings.TrimPrefix(key, "/")

	return &S3Path{
		s3Context: s3Context,
		bucket:    bucket,
		key:       key,
	}
}

func (p *S3Path) Path() string {
	return "s3://" + p.bucket + "/" + p.key
}

func (p *S3Path) Bucket() string {
	return p.bucket
}

func (p *S3Path) Key() string {
	return p.key
}

func (p *S3Path) String() string {
	return p.Path()
}

func (p *S3Path) Remove() error {
	client, err := p.client()
	if err != nil {
		return err
	}

	request := &s3.DeleteObjectInput{}
	request.Bucket = aws.String(p.bucket)
	request.Key = aws.String(p.key)

	_, err = client.DeleteObject(request)
	if err != nil {
		// TODO: Check for not-exists, return os.NotExist

		return fmt.Errorf("error deleting %s: %v", p, err)
	}

	return nil
}

func (p *S3Path) Join(relativePath ...string) Path {
	args := []string{p.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &S3Path{
		s3Context: p.s3Context,
		bucket:    p.bucket,
		key:       joined,
	}
}

func (p *S3Path) WriteFile(data []byte) error {
	client, err := p.client()
	if err != nil {
		return err
	}

	glog.V(4).Infof("Writing file %q", p)

	// We always use server-side-encryption; it doesn't really cost us anything
	sse := "AES256"

	request := &s3.PutObjectInput{}
	request.Body = bytes.NewReader(data)
	request.Bucket = aws.String(p.bucket)
	request.Key = aws.String(p.key)
	request.ServerSideEncryption = aws.String(sse)

	acl := os.Getenv("KOPS_STATE_S3_ACL")
	acl = strings.TrimSpace(acl)
	if acl != "" {
		glog.Infof("Using KOPS_STATE_S3_ACL=%s", acl)
		request.ACL = aws.String(acl)
	}

	// We don't need Content-MD5: https://github.com/aws/aws-sdk-go/issues/208

	glog.V(8).Infof("Calling S3 PutObject Bucket=%q Key=%q SSE=%q ACL=%q BodyLen=%d", p.bucket, p.key, sse, acl, len(data))

	_, err = client.PutObject(request)
	if err != nil {
		if acl != "" {
			return fmt.Errorf("error writing %s (with ACL=%q): %v", p, acl, err)
		} else {
			return fmt.Errorf("error writing %s: %v", p, err)
		}
	}

	return nil
}

// To prevent concurrent creates on the same file while maintaining atomicity of writes,
// we take a process-wide lock during the operation.
// Not a great approach, but fine for a single process (with low concurrency)
// TODO: should we enable versioning?
var createFileLockS3 sync.Mutex

func (p *S3Path) CreateFile(data []byte) error {
	createFileLockS3.Lock()
	defer createFileLockS3.Unlock()

	// Check if exists
	_, err := p.ReadFile()
	if err == nil {
		return os.ErrExist
	}

	if !os.IsNotExist(err) {
		return err
	}

	return p.WriteFile(data)
}

func (p *S3Path) ReadFile() ([]byte, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	glog.V(4).Infof("Reading file %q", p)

	request := &s3.GetObjectInput{}
	request.Bucket = aws.String(p.bucket)
	request.Key = aws.String(p.key)

	response, err := client.GetObject(request)
	if err != nil {
		if AWSErrorCode(err) == "NoSuchKey" {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("error fetching %s: %v", p, err)
	}
	defer response.Body.Close()

	d, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", p, err)
	}
	return d, nil
}

func (p *S3Path) ReadDir() ([]Path, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	prefix := p.key
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	request := &s3.ListObjectsInput{}
	request.Bucket = aws.String(p.bucket)
	request.Prefix = aws.String(prefix)
	request.Delimiter = aws.String("/")

	glog.V(4).Infof("Listing objects in S3 bucket %q with prefix %q", p.bucket, prefix)
	var paths []Path
	err = client.ListObjectsPages(request, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range page.Contents {
			key := aws.StringValue(o.Key)
			if key == prefix {
				// We have reports (#548 and #520) of the directory being returned as a file
				// And this will indeed happen if the directory has been created as a file,
				// which seems to happen if you use some external tools to manipulate the S3 bucket.
				// We need to tolerate that, so skip the parent directory.
				glog.V(4).Infof("Skipping read of directory: %q", key)
				continue
			}
			child := &S3Path{
				s3Context: p.s3Context,
				bucket:    p.bucket,
				key:       key,
				etag:      o.ETag,
			}
			paths = append(paths, child)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing %s: %v", p, err)
	}
	glog.V(8).Infof("Listed files in %v: %v", p, paths)
	return paths, nil
}

func (p *S3Path) ReadTree() ([]Path, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	request := &s3.ListObjectsInput{}
	request.Bucket = aws.String(p.bucket)
	prefix := p.key
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	request.Prefix = aws.String(prefix)
	// No delimiter for recursive search

	var paths []Path
	err = client.ListObjectsPages(request, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range page.Contents {
			key := aws.StringValue(o.Key)
			child := &S3Path{
				s3Context: p.s3Context,
				bucket:    p.bucket,
				key:       key,
				etag:      o.ETag,
			}
			paths = append(paths, child)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing %s: %v", p, err)
	}
	return paths, nil
}

func (p *S3Path) client() (*s3.S3, error) {
	var err error
	if p.region == "" {
		p.region, err = p.s3Context.getRegionForBucket(p.bucket)
		if err != nil {
			return nil, err
		}
	}

	client, err := p.s3Context.getClient(p.region)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (p *S3Path) Base() string {
	return path.Base(p.key)
}

func (p *S3Path) PreferredHash() (*hashing.Hash, error) {
	return p.Hash(hashing.HashAlgorithmMD5)
}

func (p *S3Path) Hash(a hashing.HashAlgorithm) (*hashing.Hash, error) {
	if a != hashing.HashAlgorithmMD5 {
		return nil, nil
	}

	if p.etag == nil {
		return nil, nil
	}

	md5 := strings.Trim(*p.etag, "\"")

	md5Bytes, err := hex.DecodeString(md5)
	if err != nil {
		return nil, fmt.Errorf("Etag was not a valid MD5 sum: %q", *p.etag)
	}

	return &hashing.Hash{Algorithm: hashing.HashAlgorithmMD5, HashValue: md5Bytes}, nil
}

// AWSErrorCode returns the aws error code, if it is an awserr.Error, otherwise ""
func AWSErrorCode(err error) string {
	if awsError, ok := err.(awserr.Error); ok {
		return awsError.Code()
	}
	return ""
}
