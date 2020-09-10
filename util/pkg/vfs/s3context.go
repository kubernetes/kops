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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"k8s.io/klog/v2"
)

var (
	// matches all regional naming conventions of S3:
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
	// TODO: perhaps make region regex more specific, i.e. (us|eu|ap|cn|ca|sa), to prevent matching bucket names that match region format?
	//       but that will mean updating this list when AWS introduces new regions
	s3UrlRegexp = regexp.MustCompile(`(s3([-.](?P<region>\w{2}-\w+-\d{1})|[-.](?P<bucket>[\w.\-\_]+)|)?|(?P<bucket>[\w.\-\_]+).s3.(?P<region>\w{2}-\w+-\d{1})).amazonaws.com(.cn)?(?P<path>.*)?`)
)

type S3BucketDetails struct {
	// context is the S3Context we are associated with
	context *S3Context

	// region is the region we have determined for the bucket
	region string

	// name is the name of the bucket
	name string

	// mutex protects applyServerSideEncryptionByDefault
	mutex sync.Mutex

	// applyServerSideEncryptionByDefault caches information on whether server-side encryption is enabled on the bucket
	applyServerSideEncryptionByDefault *bool
}

type S3Context struct {
	mutex         sync.Mutex
	clients       map[string]*s3.S3
	bucketDetails map[string]*S3BucketDetails
}

func NewS3Context() *S3Context {
	return &S3Context{
		clients:       make(map[string]*s3.S3),
		bucketDetails: make(map[string]*S3BucketDetails),
	}
}

func (s *S3Context) getClient(region string) (*s3.S3, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s3Client := s.clients[region]
	if s3Client == nil {
		var config *aws.Config
		var err error
		endpoint := os.Getenv("S3_ENDPOINT")
		if endpoint == "" {
			config = aws.NewConfig().WithRegion(region)
			config = config.WithCredentialsChainVerboseErrors(true)
		} else {
			// Use customized S3 storage
			klog.Infof("Found S3_ENDPOINT=%q, using as non-AWS S3 backend", endpoint)
			config, err = getCustomS3Config(endpoint, region)
			if err != nil {
				return nil, err
			}
		}

		sess, err := session.NewSession(config)
		if err != nil {
			return nil, fmt.Errorf("error starting new AWS session: %v", err)
		}
		s3Client = s3.New(sess, config)
		s.clients[region] = s3Client
	}

	return s3Client, nil
}

func getCustomS3Config(endpoint string, region string) (*aws.Config, error) {
	accessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	if accessKeyID == "" {
		return nil, fmt.Errorf("S3_ACCESS_KEY_ID cannot be empty when S3_ENDPOINT is not empty")
	}
	secretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		return nil, fmt.Errorf("S3_SECRET_ACCESS_KEY cannot be empty when S3_ENDPOINT is not empty")
	}

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
	}
	s3Config = s3Config.WithCredentialsChainVerboseErrors(true)

	return s3Config, nil
}

func (s *S3Context) getDetailsForBucket(bucket string) (*S3BucketDetails, error) {
	s.mutex.Lock()
	bucketDetails := s.bucketDetails[bucket]
	s.mutex.Unlock()

	if bucketDetails != nil && bucketDetails.region != "" {
		return bucketDetails, nil
	}

	bucketDetails = &S3BucketDetails{
		context: s,
		region:  "",
		name:    bucket,
	}

	// Probe to find correct region for bucket
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint != "" {
		// If customized S3 storage is set, return user-defined region
		bucketDetails.region = os.Getenv("S3_REGION")
		if bucketDetails.region == "" {
			bucketDetails.region = "us-east-1"
		}
		return bucketDetails, nil
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		isEC2, err := isRunningOnEC2()
		if isEC2 || err != nil {
			region, err := getRegionFromMetadata()
			if err != nil {
				klog.V(2).Infof("unable to get region from metadata:%v", err)
			} else {
				awsRegion = region
				klog.V(2).Infof("got region from metadata: %q", awsRegion)
			}
		}
	}

	if awsRegion == "" {
		awsRegion = "us-east-1"
		klog.V(2).Infof("defaulting region to %q", awsRegion)
	}

	if err := validateRegion(awsRegion); err != nil {
		return bucketDetails, err
	}

	request := &s3.GetBucketLocationInput{
		Bucket: &bucket,
	}
	var response *s3.GetBucketLocationOutput

	s3Client, err := s.getClient(awsRegion)
	if err != nil {
		return bucketDetails, fmt.Errorf("error connecting to S3: %s", err)
	}
	// Attempt one GetBucketLocation call the "normal" way (i.e. as the bucket owner)
	response, err = s3Client.GetBucketLocation(request)

	// and fallback to brute-forcing if it fails
	if err != nil {
		klog.V(2).Infof("unable to get bucket location from region %q; scanning all regions: %v", awsRegion, err)
		response, err = bruteforceBucketLocation(&awsRegion, request)
	}

	if err != nil {
		return bucketDetails, err
	}

	if response.LocationConstraint == nil {
		// US Classic does not return a region
		bucketDetails.region = "us-east-1"
	} else {
		bucketDetails.region = *response.LocationConstraint
		// Another special case: "EU" can mean eu-west-1
		if bucketDetails.region == "EU" {
			bucketDetails.region = "eu-west-1"
		}
	}

	klog.V(2).Infof("found bucket in region %q", bucketDetails.region)

	s.mutex.Lock()
	s.bucketDetails[bucket] = bucketDetails
	s.mutex.Unlock()

	return bucketDetails, nil
}

func (b *S3BucketDetails) hasServerSideEncryptionByDefault() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.applyServerSideEncryptionByDefault != nil {
		return *b.applyServerSideEncryptionByDefault
	}

	applyServerSideEncryptionByDefault := false

	// We only make one attempt to find the SSE policy (even if there's an error)
	b.applyServerSideEncryptionByDefault = &applyServerSideEncryptionByDefault

	client, err := b.context.getClient(b.region)
	if err != nil {
		klog.Warningf("Unable to read bucket encryption policy for %q in region %q: will encrypt using AES256", b.name, b.region)
		return false
	}

	klog.V(4).Infof("Checking default bucket encryption for %q", b.name)

	request := &s3.GetBucketEncryptionInput{}
	request.Bucket = aws.String(b.name)

	klog.V(8).Infof("Calling S3 GetBucketEncryption Bucket=%q", b.name)

	result, err := client.GetBucketEncryption(request)
	if err != nil {
		// the following cases might lead to the operation failing:
		// 1. A deny policy on s3:GetEncryptionConfiguration
		// 2. No default encryption policy set
		klog.V(8).Infof("Unable to read bucket encryption policy for %q: will encrypt using AES256", b.name)
		return false
	}

	// currently, only one element is in the rules array, iterating nonetheless for future compatibility
	for _, element := range result.ServerSideEncryptionConfiguration.Rules {
		if element.ApplyServerSideEncryptionByDefault != nil {
			applyServerSideEncryptionByDefault = true
		}
	}

	b.applyServerSideEncryptionByDefault = &applyServerSideEncryptionByDefault

	klog.V(2).Infof("bucket %q has default encryption set to %t", b.name, applyServerSideEncryptionByDefault)

	return applyServerSideEncryptionByDefault
}

/*
Amazon's S3 API provides the GetBucketLocation call to determine the region in which a bucket is located.
This call can however only be used globally by the owner of the bucket, as mentioned on the documentation page.

For S3 buckets that are shared across multiple AWS accounts using bucket policies the call will only work if it is sent
to the correct region in the first place.

This method will attempt to "bruteforce" the bucket location by sending a request to every available region and picking
out the first result.

See also: https://docs.aws.amazon.com/goto/WebAPI/s3-2006-03-01/GetBucketLocationRequest
*/
func bruteforceBucketLocation(region *string, request *s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error) {
	config := &aws.Config{Region: region}
	config = config.WithCredentialsChainVerboseErrors(true)

	session, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("error creating aws session: %v", err)
	}

	regions, err := ec2.New(session).DescribeRegions(nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to list AWS regions: %v", err)
	}

	klog.V(2).Infof("Querying S3 for bucket location for %s", *request.Bucket)

	out := make(chan *s3.GetBucketLocationOutput, len(regions.Regions))
	for _, region := range regions.Regions {
		go func(regionName string) {
			klog.V(8).Infof("Doing GetBucketLocation in %q", regionName)
			s3Client := s3.New(session, &aws.Config{Region: aws.String(regionName)})
			result, bucketError := s3Client.GetBucketLocation(request)
			if bucketError == nil {
				klog.V(8).Infof("GetBucketLocation succeeded in %q", regionName)
				out <- result
			}
		}(*region.RegionName)
	}

	select {
	case bucketLocation := <-out:
		return bucketLocation, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("Could not retrieve location for AWS bucket %s", *request.Bucket)
	}
}

// isRunningOnEC2 determines if we could be running on EC2.
// It is used to avoid a call to the metadata service to get the current region,
// because that call is slow if not running on EC2
func isRunningOnEC2() (bool, error) {
	if runtime.GOOS == "linux" {
		// Approach based on https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/identify_ec2_instances.html
		productUUID, err := ioutil.ReadFile("/sys/devices/virtual/dmi/id/product_uuid")
		if err != nil {
			klog.V(2).Infof("unable to read /sys/devices/virtual/dmi/id/product_uuid, assuming not running on EC2: %v", err)
			return false, err
		}

		s := strings.ToLower(strings.TrimSpace(string(productUUID)))
		if strings.HasPrefix(s, "ec2") {
			klog.V(2).Infof("product_uuid is %q, assuming running on EC2", s)
			return true, nil
		}
		klog.V(2).Infof("product_uuid is %q, assuming not running on EC2", s)
		return false, nil
	}
	klog.V(2).Infof("GOOS=%q, assuming not running on EC2", runtime.GOOS)
	return false, nil
}

// getRegionFromMetadata queries the metadata service for the current region, if running in EC2
func getRegionFromMetadata() (string, error) {
	// Use an even shorter timeout, to minimize impact when not running on EC2
	// Note that we still retry a few times, this works out a little under a 1s delay
	shortTimeout := &aws.Config{
		HTTPClient: &http.Client{
			Timeout: 100 * time.Millisecond,
		},
	}

	metadataSession, err := session.NewSession(shortTimeout)
	if err != nil {
		return "", fmt.Errorf("unable to build session: %v", err)
	}

	metadata := ec2metadata.New(metadataSession)
	metadataRegion, err := metadata.Region()

	if err != nil {
		return "", fmt.Errorf("unable to get region from metadata: %v", err)
	}

	return metadataRegion, nil
}

func validateRegion(region string) error {
	resolver := endpoints.DefaultResolver()
	partitions := resolver.(endpoints.EnumPartitions).Partitions()
	for _, p := range partitions {
		for _, r := range p.Regions() {
			if r.ID() == region {
				return nil
			}
		}
	}
	return fmt.Errorf("%s is not a valid region\nPlease check that your region is formatted correctly (e.g. us-east-1)", region)
}

func VFSPath(url string) (string, error) {
	if !s3UrlRegexp.MatchString(url) {
		return "", fmt.Errorf("%s is not a valid S3 URL", url)
	}
	groupNames := s3UrlRegexp.SubexpNames()
	result := s3UrlRegexp.FindAllStringSubmatch(url, -1)[0]

	captured := map[string]string{}
	for i, value := range result {
		if value != "" {
			captured[groupNames[i]] = value
		}
	}
	bucket := captured["bucket"]
	path := captured["path"]
	if bucket == "" {
		if path == "" {
			return "", fmt.Errorf("%s is not a valid S3 URL. No bucket defined.", url)
		}
		return fmt.Sprintf("s3:/%s", path), nil
	}
	return fmt.Sprintf("s3://%s%s", bucket, path), nil
}
