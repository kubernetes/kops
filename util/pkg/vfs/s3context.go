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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/glog"
	"os"
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
		config = config.WithCredentialsChainVerboseErrors(true)

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
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}
	s3Client, err := s.getClient(awsRegion)
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
