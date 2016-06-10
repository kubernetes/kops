package vfs

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type S3Path struct {
	client *s3.S3
	bucket string
	key    string
}

var _ Path = &S3Path{}

func NewS3Path(client *s3.S3, bucket string, key string) *S3Path {
	return &S3Path{
		client: client,
		bucket: bucket,
		key:    key,
	}
}

func (p *S3Path) String() string {
	return "s3://" + p.bucket + "/" + p.key
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
	request := &s3.PutObjectInput{}
	request.Body = bytes.NewReader(data)
	request.Bucket = aws.String(p.bucket)
	request.Key = aws.String(p.key)

	// We don't need Content-MD5: https://github.com/aws/aws-sdk-go/issues/208

	_, err := p.client.PutObject(request)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", p, err)
	}

	return err
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
	request := &s3.GetObjectInput{}
	request.Bucket = aws.String(p.bucket)
	request.Key = aws.String(p.key)

	response, err := p.client.GetObject(request)
	if err != nil {
		// TODO: If not found, return os.NotExist
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
	request := &s3.ListObjectsInput{}
	request.Bucket = aws.String(p.bucket)
	request.Prefix = aws.String(p.key)
	request.Delimiter = aws.String("/")

	var paths []Path
	err := p.client.ListObjectsPages(request, func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, o := range page.Contents {
			key := aws.StringValue(o.Key)
			child := &S3Path{
				client: p.client,
				bucket: p.bucket,
				key:    key,
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
