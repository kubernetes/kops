package vfs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/glog"
	"sync"
)

type S3Context struct {
	mutex           sync.Mutex
	clients         map[string]*s3.S3
	bucketLocations map[string]string
}

func NewS3Context() *S3Context {
	return &S3Context{
		clients:         make(map[string]*s3.S3),
		bucketLocations: make(map[string]string),
	}
}

func (s *S3Context) getClient(region string) (*s3.S3, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s3Client := s.clients[region]
	if s3Client == nil {
		config := aws.NewConfig().WithRegion(region)
		session := session.New()
		s3Client = s3.New(session, config)
	}

	s.clients[region] = s3Client

	return s3Client, nil
}

func (s *S3Context) getRegionForBucket(bucket string) (string, error) {
	region := func() string {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		return s.bucketLocations[bucket]
	}()

	if region != "" {
		return region, nil
	}

	// Probe to find correct region for bucket
	s3Client, err := s.getClient("us-east-1")
	if err != nil {
		return "", fmt.Errorf("unable to get client for querying s3 bucket location: %v", err)
	}

	request := &s3.GetBucketLocationInput{}
	request.Bucket = aws.String(bucket)

	glog.V(2).Infof("Querying S3 for bucket location for %q", bucket)
	response, err := s3Client.GetBucketLocation(request)
	if err != nil {
		// TODO: Auto-create bucket?
		return "", fmt.Errorf("error getting location for S3 bucket %q: %v", bucket, err)
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

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.bucketLocations[bucket] = region

	return region, nil
}
