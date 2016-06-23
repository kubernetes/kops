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
	"k8s.io/kube-deploy/upup/pkg/fi/hashing"
	"os"
	"path"
	"strings"
	"sync"
)

type S3Path struct {
	client *s3.S3
	bucket string
	key    string
	etag   *string
}

var _ Path = &S3Path{}
var _ HasHash = &S3Path{}

func NewS3Path(client *s3.S3, bucket string, key string) *S3Path {
	bucket = strings.TrimSuffix(bucket, "/")
	key = strings.TrimPrefix(key, "/")

	return &S3Path{
		client: client,
		bucket: bucket,
		key:    key,
	}
}

func (p *S3Path) Path() string {
	return "s3://" + p.bucket + "/" + p.key
}

func (p *S3Path) Bucket() string {
	return p.bucket
}

func (p *S3Path) String() string {
	return p.Path()
}

func (p *S3Path) Remove() error {
	request := &s3.DeleteObjectInput{}
	request.Bucket = aws.String(p.bucket)
	request.Key = aws.String(p.key)

	_, err := p.client.DeleteObject(request)
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
		client: p.client,
		bucket: p.bucket,
		key:    joined,
	}
}

func (p *S3Path) WriteFile(data []byte) error {
	glog.V(4).Infof("Writing file %q", p)

	request := &s3.PutObjectInput{}
	request.Body = bytes.NewReader(data)
	request.Bucket = aws.String(p.bucket)
	request.Key = aws.String(p.key)

	// We don't need Content-MD5: https://github.com/aws/aws-sdk-go/issues/208

	_, err := p.client.PutObject(request)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", p, err)
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
	glog.V(4).Infof("Reading file %q", p)

	request := &s3.GetObjectInput{}
	request.Bucket = aws.String(p.bucket)
	request.Key = aws.String(p.key)

	response, err := p.client.GetObject(request)
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
	prefix := p.key
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	request := &s3.ListObjectsInput{}
	request.Bucket = aws.String(p.bucket)
	request.Prefix = aws.String(prefix)
	request.Delimiter = aws.String("/")

	var paths []Path
	err := p.client.ListObjectsPages(request, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range page.Contents {
			key := aws.StringValue(o.Key)
			child := &S3Path{
				client: p.client,
				bucket: p.bucket,
				key:    key,
				etag:   o.ETag,
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
	request := &s3.ListObjectsInput{}
	request.Bucket = aws.String(p.bucket)
	request.Prefix = aws.String(p.key)
	// No delimiter for recursive search

	var paths []Path
	err := p.client.ListObjectsPages(request, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range page.Contents {
			key := aws.StringValue(o.Key)
			child := &S3Path{
				client: p.client,
				bucket: p.bucket,
				key:    key,
				etag:   o.ETag,
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
