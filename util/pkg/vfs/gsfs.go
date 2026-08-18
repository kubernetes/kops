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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
	"k8s.io/kops/util/pkg/hashing"
)

// GSPath is a vfs path for Google Cloud Storage
type GSPath struct {
	// vfsContext holds the VFS context for this path,
	// in particular the client / credentials we should use.
	vfsContext *VFSContext

	bucket  string
	key     string
	md5Hash string
}

var (
	_ Path          = &GSPath{}
	_ TerraformPath = &GSPath{}
	_ HasHash       = &GSPath{}
)

// gcsReadBackoff is the backoff strategy for GCS read retries
var gcsReadBackoff = wait.Backoff{
	Duration: time.Second,
	Factor:   1.5,
	Jitter:   0.1,
	Steps:    4,
}

// GSAcl is an ACL implementation for objects on Google Cloud Storage
type GSAcl struct {
	Acl []storage.ACLRule
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

func NewGSPath(c *VFSContext, bucket string, key string) *GSPath {
	bucket = strings.TrimSuffix(bucket, "/")
	key = strings.TrimPrefix(key, "/")

	return &GSPath{
		vfsContext: c,
		bucket:     bucket,
		key:        key,
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

// Client returns the storage.Client bound to this path
func (p *GSPath) Client(ctx context.Context) (*storage.Client, error) {
	return p.getStorageClient(ctx)
}

func (p *GSPath) String() string {
	return p.Path()
}

func (p *GSPath) Remove(ctx context.Context) error {
	done, err := RetryWithBackoff(gcsWriteBackoff, func() (bool, error) {
		client, err := p.getStorageClient(ctx)
		if err != nil {
			return false, err
		}
		if err := client.Bucket(p.bucket).Object(p.key).Delete(ctx); err != nil {
			// TODO: Check for not-exists, return os.NotExist
			return false, fmt.Errorf("error deleting %s: %w", p, err)
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

func (p *GSPath) RemoveAll(ctx context.Context) error {
	tree, err := p.ReadTree(ctx)
	if err != nil {
		return err
	}

	for _, objectPath := range tree {
		err := objectPath.Remove(ctx)
		if err != nil {
			return fmt.Errorf("error removing file %s: %w", objectPath, err)
		}
	}

	return nil
}

func (p *GSPath) RemoveAllVersions(ctx context.Context) error {
	return p.Remove(ctx)
}

func (p *GSPath) Join(relativePath ...string) Path {
	args := []string{p.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &GSPath{
		vfsContext: p.vfsContext,
		bucket:     p.bucket,
		key:        joined,
	}
}

func (p *GSPath) WriteFile(ctx context.Context, data io.ReadSeeker, acl ACL) error {
	md5Hash, err := hashing.HashAlgorithmMD5.Hash(data)
	if err != nil {
		return err
	}

	done, err := RetryWithBackoff(gcsWriteBackoff, func() (bool, error) {
		var objectACL []storage.ACLRule
		if acl != nil {
			gsACL, ok := acl.(*GSAcl)
			if !ok {
				return true, fmt.Errorf("write to %s with ACL of unexpected type %T", p, acl)
			}
			objectACL = gsACL.Acl
			klog.V(4).Infof("Writing file %q with ACL %v", p, gsACL)
		} else {
			klog.V(4).Infof("Writing file %q", p)
		}

		if _, err := data.Seek(0, 0); err != nil {
			return false, fmt.Errorf("error seeking to start of data stream for write to %s: %v", p, err)
		}

		client, err := p.getStorageClient(ctx)
		if err != nil {
			return false, err
		}

		w := client.Bucket(p.bucket).Object(p.key).NewWriter(ctx)
		// The upload is rejected if the data does not match this MD5 hash
		w.MD5 = md5Hash.HashValue
		w.ACL = objectACL
		if _, err := io.Copy(w, data); err != nil {
			w.Close()
			return false, fmt.Errorf("error writing %s: %v", p, err)
		}
		if err := w.Close(); err != nil {
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

func (p *GSPath) CreateFile(ctx context.Context, data io.ReadSeeker, acl ACL) error {
	createFileLockGCS.Lock()
	defer createFileLockGCS.Unlock()

	// Check if exists
	_, err := p.ReadFile(ctx)
	if err == nil {
		return os.ErrExist
	}

	if !os.IsNotExist(err) {
		return err
	}

	return p.WriteFile(ctx, data, acl)
}

// ReadFile implements Path::ReadFile
func (p *GSPath) ReadFile(ctx context.Context) ([]byte, error) {
	var b bytes.Buffer
	done, err := RetryWithBackoff(gcsReadBackoff, func() (bool, error) {
		b.Reset()
		_, err := p.WriteToWithContext(ctx, &b)
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
	ctx := context.TODO()
	return p.WriteToWithContext(ctx, out)
}

// WriteToWithContext implements io.WriterTo, but adds a context
func (p *GSPath) WriteToWithContext(ctx context.Context, out io.Writer) (int64, error) {
	klog.V(4).Infof("Reading file %q", p)

	client, err := p.getStorageClient(ctx)
	if err != nil {
		return 0, err
	}

	r, err := client.Bucket(p.bucket).Object(p.key).NewReader(ctx)
	if err != nil {
		if isGCSNotFound(err) {
			return 0, os.ErrNotExist
		}
		return 0, fmt.Errorf("error reading %s: %v", p, err)
	}
	defer r.Close()

	return io.Copy(out, r)
}

// ReadDir implements Path::ReadDir
func (p *GSPath) ReadDir() ([]Path, error) {
	ctx := context.TODO()

	var ret []Path
	done, err := RetryWithBackoff(gcsReadBackoff, func() (bool, error) {
		prefix := p.key
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		client, err := p.getStorageClient(ctx)
		if err != nil {
			return false, err
		}

		var paths []Path
		it := client.Bucket(p.bucket).Objects(ctx, &storage.Query{Delimiter: "/", Prefix: prefix})
		for {
			o, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				if isGCSNotFound(err) {
					return true, os.ErrNotExist
				}
				return false, fmt.Errorf("error listing %s: %v", p, err)
			}
			if o.Prefix != "" {
				// Synthetic directory entry, not an object
				continue
			}
			child := &GSPath{
				vfsContext: p.vfsContext,
				bucket:     p.bucket,
				key:        o.Name,
				md5Hash:    base64.StdEncoding.EncodeToString(o.MD5),
			}
			paths = append(paths, child)
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
func (p *GSPath) ReadTree(ctx context.Context) ([]Path, error) {
	var ret []Path
	done, err := RetryWithBackoff(gcsReadBackoff, func() (bool, error) {
		// No delimiter for recursive search
		prefix := p.key
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		client, err := p.getStorageClient(ctx)
		if err != nil {
			return false, err
		}

		var paths []Path
		it := client.Bucket(p.bucket).Objects(ctx, &storage.Query{Prefix: prefix})
		for {
			o, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				if isGCSNotFound(err) {
					return true, os.ErrNotExist
				}
				return false, fmt.Errorf("error listing tree %s: %v", p, err)
			}
			child := &GSPath{
				vfsContext: p.vfsContext,
				bucket:     p.bucket,
				key:        o.Name,
				md5Hash:    base64.StdEncoding.EncodeToString(o.MD5),
			}
			paths = append(paths, child)
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

func (p *GSPath) GetHTTPsUrl() (string, error) {
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", p.bucket, p.key)
	return strings.TrimSuffix(url, "/"), nil
}

func (p *GSPath) IsBucketPublic(ctx context.Context) (bool, error) {
	client, err := p.Client(ctx)
	if err != nil {
		return false, err
	}

	attrs, err := client.Bucket(p.bucket).Attrs(ctx)
	if err != nil {
		return false, err
	}

	// Check bucket has uniform bucket-level IAM
	if !attrs.UniformBucketLevelAccess.Enabled {
		return false, nil
	}

	// Check `allUsers` IAM has `roles/storage.objectViewer` permission
	policy, err := client.Bucket(p.bucket).IAM().Policy(ctx)
	if err != nil {
		return false, err
	}

	for _, member := range policy.Members("roles/storage.objectViewer") {
		if member == "allUsers" {
			return true, nil
		}
	}

	return false, nil
}

type terraformGSObject struct {
	Bucket   string                   `json:"bucket" cty:"bucket"`
	Name     string                   `json:"name" cty:"name"`
	Source   *terraformWriter.Literal `json:"source" cty:"source"`
	Provider *terraformWriter.Literal `json:"provider,omitempty" cty:"provider"`
}

type terraformGSObjectAccessControl struct {
	Bucket     string                   `json:"bucket" cty:"bucket"`
	Object     *terraformWriter.Literal `json:"object" cty:"object"`
	RoleEntity []string                 `json:"role_entity" cty:"role_entity"`
	Provider   *terraformWriter.Literal `json:"provider,omitempty" cty:"provider"`
}

func (p *GSPath) RenderTerraform(w *terraformWriter.TerraformWriter, name string, data io.Reader, acl ACL) error {
	bytes, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("reading data: %v", err)
	}

	tfProviderArguments := map[string]string{} // GCS doesn't need the project and region specified
	w.EnsureTerraformProvider("google", tfProviderArguments)

	content, err := w.AddFilePath("google_storage_bucket_object", name, "content", bytes, false)
	if err != nil {
		return fmt.Errorf("rendering GCS file: %v", err)
	}

	tf := &terraformGSObject{
		Bucket:   p.Bucket(),
		Name:     p.Object(),
		Source:   content,
		Provider: terraformWriter.LiteralTokens("google", "files"),
	}
	err = w.RenderResource("google_storage_bucket_object", name, tf)
	if err != nil {
		return err
	}

	// file ACLs can be empty on GCP, because of a bucket-only ACL.
	if acl != nil {
		tfACL := &terraformGSObjectAccessControl{
			Bucket:     p.Bucket(),
			Object:     p.TerraformLink(name),
			RoleEntity: make([]string, 0),
			Provider:   terraformWriter.LiteralTokens("google", "files"),
		}
		for _, re := range acl.(*GSAcl).Acl {
			// https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/storage_object_acl#role_entity
			tfACL.RoleEntity = append(tfACL.RoleEntity, fmt.Sprintf("%v:%v", re.Role, re.Entity))
		}
		return w.RenderResource("google_storage_object_acl", name, tfACL)
	}
	return nil
}

func (s *GSPath) TerraformLink(name string) *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("google_storage_bucket_object", name, "output_name")
}

func isGCSNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, storage.ErrObjectNotExist) || errors.Is(err, storage.ErrBucketNotExist) {
		return true
	}
	var ae *googleapi.Error
	return errors.As(err, &ae) && ae.Code == http.StatusNotFound
}

func (p *GSPath) getStorageClient(ctx context.Context) (*storage.Client, error) {
	return p.vfsContext.getGCSClient(ctx)
}
