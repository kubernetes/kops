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
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/denverdino/aliyungo/oss"
	"github.com/gophercloud/gophercloud"
	"google.golang.org/api/option"
	storage "google.golang.org/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	vault "github.com/hashicorp/vault/api"
)

// VFSContext is a 'context' for VFS, that is normally a singleton
// but allows us to configure S3 credentials, for example
type VFSContext struct {
	s3Context    *S3Context
	k8sContext   *KubernetesContext
	memfsContext *MemFSContext
	// mutex guards gcsClient
	mutex sync.Mutex
	// The google cloud storage client, if initialized
	gcsClient *storage.Service
	// swiftClient is the openstack swift client
	swiftClient *gophercloud.ServiceClient
	// ossClient is the Aliyun Open Source Storage client
	ossClient *oss.Client

	vaultClient *vault.Client
}

var Context = VFSContext{
	s3Context:  NewS3Context(),
	k8sContext: NewKubernetesContext(),
}

type vfsOptions struct {
	backoff wait.Backoff
}

type VFSOption func(options *vfsOptions)

// WithBackoff specifies a custom VFS backoff policy
func WithBackoff(backoff wait.Backoff) VFSOption {
	return func(options *vfsOptions) {
		options.backoff = backoff
	}
}

// ReadFile reads a file from a vfs URL
// It supports additional schemes which don't (yet) have full VFS implementations:
//   metadata: reads from instance metadata on GCE/AWS
//   http / https: reads from HTTP
func (c *VFSContext) ReadFile(location string, options ...VFSOption) ([]byte, error) {
	ctx := context.TODO()

	var opts vfsOptions
	// Exponential backoff, starting with 500 milliseconds, doubling each time, 5 steps
	opts.backoff = wait.Backoff{
		Duration: 500 * time.Millisecond,
		Factor:   2,
		Steps:    5,
	}

	for _, option := range options {
		option(&opts)
	}

	if strings.Contains(location, "://") && !strings.HasPrefix(location, "file://") {
		// Handle our special case schemas
		u, err := url.Parse(location)
		if err != nil {
			return nil, fmt.Errorf("error parsing location %q - not a valid URI", location)
		}

		switch u.Scheme {
		case "metadata":
			switch u.Host {
			case "gce":
				httpURL := "http://169.254.169.254/computeMetadata/v1/" + u.Path
				httpHeaders := make(map[string]string)
				httpHeaders["Metadata-Flavor"] = "Google"
				return c.readHTTPLocation(httpURL, httpHeaders, opts)
			case "aws":
				return c.readAWSMetadata(ctx, u.Path)
			case "digitalocean":
				httpURL := "http://169.254.169.254/metadata/v1" + u.Path
				return c.readHTTPLocation(httpURL, nil, opts)
			case "alicloud":
				httpURL := "http://100.100.100.200/latest/meta-data/" + u.Path
				return c.readHTTPLocation(httpURL, nil, opts)
			case "openstack":
				httpURL := "http://169.254.169.254/latest/meta-data/" + u.Path
				return c.readHTTPLocation(httpURL, nil, opts)
			default:
				return nil, fmt.Errorf("unknown metadata type: %q in %q", u.Host, location)
			}

		case "http", "https":
			return c.readHTTPLocation(location, nil, opts)
		}
	}

	location = strings.TrimPrefix(location, "file://")

	p, err := c.BuildVfsPath(location)
	if err != nil {
		return nil, err
	}
	return p.ReadFile()
}

func (c *VFSContext) BuildVfsPath(p string) (Path, error) {
	if !strings.Contains(p, "://") {
		return NewFSPath(p), nil
	}

	if strings.HasPrefix(p, "file://") {
		f := strings.TrimPrefix(p, "file://")
		return NewFSPath(f), nil
	}

	if strings.HasPrefix(p, "s3://") {
		return c.buildS3Path(p)
	}

	if strings.HasPrefix(p, "do://") {
		return c.buildDOPath(p)
	}

	if strings.HasPrefix(p, "memfs://") {
		return c.buildMemFSPath(p)
	}

	if strings.HasPrefix(p, "gs://") {
		return c.buildGCSPath(p)
	}

	if strings.HasPrefix(p, "k8s://") {
		return c.buildKubernetesPath(p)
	}

	if strings.HasPrefix(p, "swift://") {
		return c.buildOpenstackSwiftPath(p)
	}

	if strings.HasPrefix(p, "oss://") {
		return c.buildOSSPath(p)
	}

	if strings.HasPrefix(p, "vault://") {
		return c.buildVaultPath(p)
	}

	return nil, fmt.Errorf("unknown / unhandled path type: %q", p)
}

// readAWSMetadata reads the specified path from the AWS EC2 metadata service
func (c *VFSContext) readAWSMetadata(ctx context.Context, path string) ([]byte, error) {
	awsSession, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("error building AWS session: %v", err)
	}
	client := ec2metadata.New(awsSession)
	if strings.HasPrefix(path, "/meta-data/") {
		s, err := client.GetMetadataWithContext(ctx, strings.TrimPrefix(path, "/meta-data/"))
		if err != nil {
			return nil, fmt.Errorf("error reading from AWS metadata service: %v", err)
		}
		return []byte(s), nil
	}
	// There are others (e.g. user-data), but as we don't use them yet let's not expose them
	return nil, fmt.Errorf("unhandled aws metadata path %q", path)
}

// readHTTPLocation reads an http (or https) url.
// It returns the contents, or an error on any non-200 response.  On a 404, it will return os.ErrNotExist
// It will retry a few times on a 500 class error
func (c *VFSContext) readHTTPLocation(httpURL string, httpHeaders map[string]string, opts vfsOptions) ([]byte, error) {
	var body []byte

	done, err := RetryWithBackoff(opts.backoff, func() (bool, error) {
		klog.V(4).Infof("Performing HTTP request: GET %s", httpURL)
		req, err := http.NewRequest("GET", httpURL, nil)
		if err != nil {
			return false, err
		}
		for k, v := range httpHeaders {
			req.Header.Add(k, v)
		}
		response, err := http.DefaultClient.Do(req)
		if response != nil {
			defer response.Body.Close()
		}
		if err != nil {
			return false, fmt.Errorf("error fetching %q: %v", httpURL, err)
		}
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return false, fmt.Errorf("error reading response for %q: %v", httpURL, err)
		}
		if response.StatusCode == 404 {
			// We retry on 404s in case of eventual consistency
			return false, os.ErrNotExist
		}
		if response.StatusCode >= 500 && response.StatusCode <= 599 {
			// Retry on 5XX errors
			return false, fmt.Errorf("unexpected response code %q for %q: %v", response.Status, httpURL, string(body))
		}

		if response.StatusCode == 200 {
			return true, nil
		}

		// Don't retry on other errors
		return true, fmt.Errorf("unexpected response code %q for %q: %v", response.Status, httpURL, string(body))
	})
	if err != nil {
		return nil, err
	} else if done {
		return body, nil
	} else {
		// Shouldn't happen - we always return a non-nil error with false
		return nil, wait.ErrWaitTimeout
	}
}

// RetryWithBackoff runs until a condition function returns true, or until Steps attempts have been taken
// As compared to wait.ExponentialBackoff, this function returns the results from the function on the final attempt
func RetryWithBackoff(backoff wait.Backoff, condition func() (bool, error)) (bool, error) {
	duration := backoff.Duration
	i := 0
	for {
		if i != 0 {
			adjusted := duration
			if backoff.Jitter > 0.0 {
				adjusted = wait.Jitter(duration, backoff.Jitter)
			}
			time.Sleep(adjusted)
			duration = time.Duration(float64(duration) * backoff.Factor)
		}

		i++

		done, err := condition()
		if done {
			return done, err
		}
		noMoreRetries := i >= backoff.Steps
		if !noMoreRetries && err != nil {
			klog.V(2).Infof("retrying after error %v", err)
		}

		if noMoreRetries {
			klog.V(2).Infof("hit maximum retries %d with error %v", i, err)
			return done, err
		}
	}
}

func (c *VFSContext) buildS3Path(p string) (*S3Path, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid s3 path: %q", p)
	}
	if u.Scheme != "s3" {
		return nil, fmt.Errorf("invalid s3 path: %q", p)
	}

	bucket := strings.TrimSuffix(u.Host, "/")
	if bucket == "" {
		return nil, fmt.Errorf("invalid s3 path: %q", p)
	}

	s3path := newS3Path(c.s3Context, u.Scheme, bucket, u.Path, true)
	return s3path, nil
}

func (c *VFSContext) buildDOPath(p string) (*S3Path, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid spaces path: %q", p)
	}
	if u.Scheme != "do" {
		return nil, fmt.Errorf("invalid spaces path: %q", p)
	}

	bucket := strings.TrimSuffix(u.Host, "/")
	if bucket == "" {
		return nil, fmt.Errorf("invalid spaces path: %q", p)
	}

	s3path := newS3Path(c.s3Context, u.Scheme, bucket, u.Path, false)
	return s3path, nil
}

func (c *VFSContext) buildKubernetesPath(p string) (*KubernetesPath, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid kubernetes vfs path: %q", p)
	}
	if u.Scheme != "k8s" {
		return nil, fmt.Errorf("invalid kubernetes vfs path: %q", p)
	}

	bucket := strings.TrimSuffix(u.Host, "/")
	if bucket == "" {
		return nil, fmt.Errorf("invalid kubernetes vfs path: %q", p)
	}

	k8sPath := newKubernetesPath(c.k8sContext, bucket, u.Path)
	return k8sPath, nil
}

func (c *VFSContext) buildMemFSPath(p string) (*MemFSPath, error) {
	if !strings.HasPrefix(p, "memfs://") {
		return nil, fmt.Errorf("memfs path not recognized: %q", p)
	}
	location := strings.TrimPrefix(p, "memfs://")
	if c.memfsContext == nil {
		// We only initialize this in unit tests etc
		return nil, fmt.Errorf("memfs context not initialized")
	}
	fspath := NewMemFSPath(c.memfsContext, location)
	return fspath, nil
}

func (c *VFSContext) ResetMemfsContext(clusterReadable bool) {
	c.memfsContext = NewMemFSContext()
	if clusterReadable {
		c.memfsContext.MarkClusterReadable()
	}
}

func (c *VFSContext) buildGCSPath(p string) (*GSPath, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid google cloud storage path: %q", p)
	}

	if u.Scheme != "gs" {
		return nil, fmt.Errorf("invalid google cloud storage path: %q", p)
	}

	bucket := strings.TrimSuffix(u.Host, "/")

	gcsClient, err := c.getGCSClient()
	if err != nil {
		return nil, err
	}

	gcsPath := NewGSPath(gcsClient, bucket, u.Path)
	return gcsPath, nil
}

// getGCSClient returns the google storage.Service client, caching it for future calls
func (c *VFSContext) getGCSClient() (*storage.Service, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.gcsClient != nil {
		return c.gcsClient, nil
	}

	// TODO: Should we fall back to read-only?
	scope := storage.DevstorageReadWriteScope

	ctx := context.Background()
	gcsClient, err := storage.NewService(ctx, option.WithScopes(scope))
	if err != nil {
		return nil, fmt.Errorf("error building GCS client: %v", err)
	}

	c.gcsClient = gcsClient
	return gcsClient, nil
}

func (c *VFSContext) buildOpenstackSwiftPath(p string) (*SwiftPath, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid openstack cloud storage path: %q", p)
	}

	if u.Scheme != "swift" {
		return nil, fmt.Errorf("invalid openstack cloud storage path: %q", p)
	}

	bucket := strings.TrimSuffix(u.Host, "/")
	if bucket == "" {
		return nil, fmt.Errorf("invalid swift path: %q", p)
	}

	if c.swiftClient == nil {
		swiftClient, err := NewSwiftClient()
		if err != nil {
			return nil, err
		}
		c.swiftClient = swiftClient
	}

	return NewSwiftPath(c.swiftClient, bucket, u.Path)
}

func (c *VFSContext) buildOSSPath(p string) (*OSSPath, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid aliyun oss path: %q", p)
	}

	if u.Scheme != "oss" {
		return nil, fmt.Errorf("invalid aliyun oss path: %q", p)
	}

	bucket := strings.TrimSuffix(u.Host, "/")
	if bucket == "" {
		return nil, fmt.Errorf("invalid aliyun oss path: %q", p)
	}

	if c.ossClient == nil {
		ossClient, err := NewAliOSSClient()
		if err != nil {
			return nil, err
		}
		c.ossClient = ossClient
	}

	return NewOSSPath(c.ossClient, bucket, u.Path)
}

func (c *VFSContext) buildVaultPath(p string) (*VaultPath, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid vault url: %q", p)
	}

	var scheme string

	if u.Scheme != "vault" {
		return nil, fmt.Errorf("invalid vault url: %q", p)
	}

	queryValues := u.Query()

	scheme = "https://"
	if queryValues.Get("tls") == "false" {
		scheme = "http://"
	}

	if c.vaultClient == nil {

		vaultClient, err := newVaultClient(scheme, u.Hostname(), u.Port())
		if err != nil {
			return nil, err
		}

		c.vaultClient = vaultClient
	}

	return newVaultPath(c.vaultClient, scheme, u.Path)
}
