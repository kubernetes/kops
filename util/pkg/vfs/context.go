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
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// VFSContext is a 'context' for VFS, that is normally a singleton
// but allows us to configure S3 credentials, for example
type VFSContext struct {
	s3Context    *S3Context
	memfsContext *MemFSContext
	// mutex guards gcsClient
	mutex sync.Mutex
	// The google cloud storage client, if initialized
	gcsClient *storage.Service
}

var Context = VFSContext{
	s3Context: NewS3Context(),
}

// ReadLocation reads a file from a vfs URL
// It supports additional schemes which don't (yet) have full VFS implementations:
//   metadata: reads from instance metadata on GCE/AWS
//   http / https: reads from HTTP
func (c *VFSContext) ReadFile(location string) ([]byte, error) {
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
				httpURL := "http://169.254.169.254/computeMetadata/v1/instance/attributes/" + u.Path
				httpHeaders := make(map[string]string)
				httpHeaders["Metadata-Flavor"] = "Google"
				return c.readHttpLocation(httpURL, httpHeaders)
			case "aws":
				httpURL := "http://169.254.169.254/latest/" + u.Path
				return c.readHttpLocation(httpURL, nil)

			default:
				return nil, fmt.Errorf("unknown metadata type: %q in %q", u.Host, location)
			}

		case "http", "https":
			return c.readHttpLocation(location, nil)
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

	if strings.HasPrefix(p, "s3://") {
		return c.buildS3Path(p)
	}

	if strings.HasPrefix(p, "memfs://") {
		return c.buildMemFSPath(p)
	}

	if strings.HasPrefix(p, "gs://") {
		return c.buildGCSPath(p)
	}

	return nil, fmt.Errorf("unknown / unhandled path type: %q", p)
}

// readHttpLocation reads an http (or https) url.
// It returns the contents, or an error on any non-200 response.  On a 404, it will return os.ErrNotExist
// It will retry a few times on a 500 class error
func (c *VFSContext) readHttpLocation(httpURL string, httpHeaders map[string]string) ([]byte, error) {
	// Exponential backoff, starting with 500 milliseconds, doubling each time, 5 steps
	backoff := wait.Backoff{
		Duration: 500 * time.Millisecond,
		Factor:   2,
		Steps:    5,
	}

	var body []byte

	done, err := RetryWithBackoff(backoff, func() (bool, error) {
		glog.V(4).Infof("Performing HTTP request: GET %s", httpURL)
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
			glog.V(2).Infof("retrying after error %v", err)
		}

		if noMoreRetries {
			glog.V(2).Infof("hit maximum retries %d with error %v", i, err)
			return done, err
		}
	}
}

func (c *VFSContext) buildS3Path(p string) (*S3Path, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid s3 path: %q", err)
	}
	if u.Scheme != "s3" {
		return nil, fmt.Errorf("invalid s3 path: %q", err)
	}

	bucket := strings.TrimSuffix(u.Host, "/")
	if bucket == "" {
		return nil, fmt.Errorf("invalid s3 path: %q", err)
	}

	s3path := newS3Path(c.s3Context, bucket, u.Path)
	return s3path, nil
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
		return nil, fmt.Errorf("invalid google cloud storage path: %q", err)
	}

	if u.Scheme != "gs" {
		return nil, fmt.Errorf("invalid google cloud storage path: %q", err)
	}

	bucket := strings.TrimSuffix(u.Host, "/")

	gcsClient, err := c.getGCSClient()
	if err != nil {
		return nil, err
	}

	gcsPath := NewGSPath(gcsClient, bucket, u.Path)
	return gcsPath, nil
}

func (c *VFSContext) getGCSClient() (*storage.Service, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.gcsClient != nil {
		return c.gcsClient, nil
	}

	// TODO: Should we fall back to read-only?
	scope := storage.DevstorageReadWriteScope

	httpClient, err := google.DefaultClient(context.Background(), scope)
	if err != nil {
		return nil, fmt.Errorf("error building GCS HTTP client: %v", err)
	}

	gcsClient, err := storage.New(httpClient)
	if err != nil {
		return nil, fmt.Errorf("error building GCS client: %v", err)
	}

	c.gcsClient = gcsClient
	return gcsClient, nil
}
