package vfs

import (
	"strings"
	"io/ioutil"
	"net/url"
	"fmt"
	"net/http"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/glog"
)

// VFSContext is a 'context' for VFS, that is normally a singleton
// but allows us to configure S3 credentials, for example
type VFSContext struct {

}

var Context VFSContext

// ReadLocation reads a file from a vfs URL
// It supports additional schemes which don't (yet) have full VFS implementations:
//   metadata: reads from instance metadata on GCE/AWS
//   http / https: reads from HTTP
func (c*VFSContext) ReadFile(location string) ([]byte, error) {
	if strings.Contains(location, "://") {
		// Handle our special case schemas
		u, err := url.Parse(location)
		if err != nil {
			return nil, fmt.Errorf("error parsing location %q - not a valid URI")
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

func (c*VFSContext) BuildVfsPath(p string) (Path, error) {
	if !strings.Contains(p, "://") {
		return NewFSPath(p), nil
	}

	if strings.HasPrefix(p, "s3://") {
		return c.buildS3Path(p)
	}

	return nil, fmt.Errorf("unknown / unhandled path type: %q", p)
}

func (c*VFSContext) readHttpLocation(httpURL string, httpHeaders map[string]string) ([]byte, error) {
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
	return body, nil
}

func (c*VFSContext) buildS3Path(p string) (*S3Path, error) {
	u, err := url.Parse(p)
	if err != nil {
		return nil, fmt.Errorf("invalid s3 path: %q", err)
	}

	bucket := strings.TrimSuffix(u.Host, "/")

	var region string
	{
		// Probe to find correct region for bucket
		// TODO: Caching (both of the client & of the bucket location)
		config := aws.NewConfig().WithRegion("us-east-1")
		session := session.New()
		s3Client := s3.New(session, config)

		request := &s3.GetBucketLocationInput{}
		request.Bucket = aws.String(bucket)

		response, err := s3Client.GetBucketLocation(request)
		if err != nil {
			// TODO: Auto-create bucket?
			return nil, fmt.Errorf("error getting location for S3 bucket %q: %v", bucket, err)
		}
		if response.LocationConstraint == nil {
			// US Classic does not return a region
			region = "us-east-1"
		} else {
			region = *response.LocationConstraint
			// Another special case: "EU" can mean eu-west-1
			if region == "EU" {
				region = "eu-west-1"
			}
		}
		glog.V(2).Infof("Found bucket %q in region %q", bucket, region)
	}

	// TODO: Caching (of the S3 client)
	config := aws.NewConfig().WithRegion(region)
	session := session.New()
	s3Client := s3.New(session, config)

	s3path := NewS3Path(s3Client, bucket, u.Path)
	return s3path, nil
}
