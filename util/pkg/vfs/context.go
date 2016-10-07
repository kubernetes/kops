package vfs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// VFSContext is a 'context' for VFS, that is normally a singleton
// but allows us to configure S3 credentials, for example
type VFSContext struct {
	s3Context *S3Context
}

var Context = VFSContext{
	s3Context: NewS3Context(),
}

// ReadLocation reads a file from a vfs URL
// It supports additional schemes which don't (yet) have full VFS implementations:
//   metadata: reads from instance metadata on GCE/AWS
//   http / https: reads from HTTP
func (c *VFSContext) ReadFile(location string) ([]byte, error) {
	if strings.Contains(location, "://") {
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

	return nil, fmt.Errorf("unknown / unhandled path type: %q", p)
}

// readHttpLocation reads an http (or https) url.
// It returns the contents, or an error on any non-200 response.  On a 404, it will return os.ErrNotExist
func (c *VFSContext) readHttpLocation(httpURL string, httpHeaders map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", httpURL, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range httpHeaders {
		req.Header.Add(k, v)
	}
	response, err := http.DefaultClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching %q: %v", httpURL, err)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response for %q: %v", httpURL, err)
	}
	if response.StatusCode == 404 {
		return nil, os.ErrNotExist
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response code %q for %q: %v", response.Status, httpURL, string(body))
	}
	return body, nil
}

func (c *VFSContext) buildS3Path(p string) (*S3Path, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid s3 path: %q", err)
	}

	bucket := strings.TrimSuffix(u.Host, "/")

	s3path := NewS3Path(c.s3Context, bucket, u.Path)
	return s3path, nil
}
