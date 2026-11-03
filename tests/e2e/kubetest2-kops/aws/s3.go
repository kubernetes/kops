/*
Copyright 2024 The Kubernetes Authors.

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

package aws

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"k8s.io/klog/v2"
)

// defaultRegion is the region to query the AWS APIs through, this can be any AWS region is required even if we are not
// running on AWS.
const defaultRegion = "us-east-2"

var bucketNameRegex = regexp.MustCompile("[^a-z0-9-]")

// Client contains S3 and STS clients that are used to perform bucket and object actions.
type Client struct {
	s3Client  *s3.Client
	stsClient *sts.Client
}

type BucketType string

const (
	BucketTypeStateStore     BucketType = "state"
	BucketTypeDiscoveryStore BucketType = "discovery"
	BucketTypeStagingStore   BucketType = "staging"
)

// NewAWSClient returns a new instance of awsClient configured to work in the default region (us-east-2).
func NewClient(ctx context.Context, region string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	return &Client{
		s3Client:  s3.NewFromConfig(cfg),
		stsClient: sts.NewFromConfig(cfg),
	}, nil
}

// BucketName constructs an unique bucket name using the AWS account ID in the default region (us-east-2).
func (c Client) BucketName(ctx context.Context, bucketType BucketType) (string, error) {
	// Construct the bucket name based on the ProwJob ID (if running in Prow) or AWS account ID (if running outside
	// Prow) and a timestamp. When BUILD_ID is set we use it in place of time.Now() so that multiple kubetest2-kops
	// invocations within the same CI job (e.g. upgrade tests) resolve to the same bucket name.
	var suffix string
	if jobID := os.Getenv("BUILD_ID"); jobID != "" {
		suffix = jobID
	} else {
		callerIdentity, err := c.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			return "", fmt.Errorf("building AWS STS presigned request: %w", err)
		}
		suffix = *callerIdentity.Account + "-" + time.Now().Format("20060102150405")
	}
	bucket := fmt.Sprintf("k8s-infra-kops-%s-%s", bucketType, suffix)

	bucket = strings.ToLower(bucket)
	// Only allow lowercase letters, numbers, and hyphens
	bucket = bucketNameRegex.ReplaceAllString(bucket, "")

	if len(bucket) > 63 {
		bucket = bucket[:63] // Max length is 63
	}

	return bucket, nil
}

// EnsureS3Bucket creates a new S3 bucket with the given name and public read permissions.
func (c Client) EnsureS3Bucket(ctx context.Context, region, bucketName string, publicRead bool) error {
	bucketName = strings.TrimPrefix(bucketName, "s3://")
	klog.Infof("Creating bucket %s in region %s", bucketName, region)
	bucketConfig := &types.CreateBucketConfiguration{}
	if region != "us-east-1" {
		bucketConfig.LocationConstraint = types.BucketLocationConstraint(region)
	}
	_, err := c.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket:                    aws.String(bucketName),
		CreateBucketConfiguration: bucketConfig,
	},
	)
	var owned *types.BucketAlreadyOwnedByYou
	if errors.As(err, &owned) {
		// The SDK retried a CreateBucket whose first attempt had succeeded.
		klog.Infof("Bucket %s already exists in this account", bucketName)
		err = nil
	}
	if err != nil {
		var exists *types.BucketAlreadyExists
		if errors.As(err, &exists) {
			klog.Infof("Bucket %s already exists\n", bucketName)
		} else {
			klog.Infof("Error creating bucket %s, err: %v\n", bucketName, err)
		}

		return fmt.Errorf("creating bucket %s: %w", bucketName, err)
	}

	// Wait for the bucket to be created
	err = s3.NewBucketExistsWaiter(c.s3Client).Wait(
		ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		},
		time.Minute)
	if err != nil {
		klog.Infof("Failed attempt to wait for bucket %s to exist, err: %v", bucketName, err)

		return fmt.Errorf("waiting for bucket %s to exist: %w", bucketName, err)
	}

	klog.Infof("Bucket %s created successfully", bucketName)

	if publicRead {
		err = c.setPublicAccessBlock(ctx, bucketName)
		if err != nil {
			klog.Errorf("Failed to disable public access block policies on bucket %s, err: %v", bucketName, err)

			return fmt.Errorf("disabling public access block policies for bucket %s: %w", bucketName, err)
		}

		// Wait for public access block settings to propagate before setting the policy
		time.Sleep(10 * time.Second)

		err = c.setPublicReadPolicy(ctx, bucketName)
		if err != nil {
			klog.Errorf("Failed to set public read policy on bucket %s, err: %v", bucketName, err)

			return fmt.Errorf("setting public read policy for bucket %s: %w", bucketName, err)
		}

		klog.Infof("Public read policy set on bucket %s", bucketName)
	}

	return nil
}

// DeleteS3Bucket deletes a S3 bucket with the given name.
func (c Client) DeleteS3Bucket(ctx context.Context, bucketName string) error {
	bucketName = strings.TrimPrefix(bucketName, "s3://")

	regionOpt := c.bucketRegionOption(ctx, bucketName)

	_, err := c.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}, regionOpt)
	if err != nil {
		var noBucket *types.NoSuchBucket
		if errors.As(err, &noBucket) {
			klog.Infof("Bucket %s does not exist.", bucketName)

			return nil
		}
		klog.Infof("Couldn't delete bucket %s, err: %v", bucketName, err)

		return fmt.Errorf("deleting bucket %s: %w", bucketName, err)
	}

	err = s3.NewBucketNotExistsWaiter(c.s3Client).Wait(
		ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		},
		time.Minute, func(o *s3.BucketNotExistsWaiterOptions) {
			o.ClientOptions = append(o.ClientOptions, regionOpt)
		})
	if err != nil {
		klog.Infof("Failed attempt to wait for bucket %s to be deleted, err: %v", bucketName, err)

		return fmt.Errorf("waiting for bucket %s to be deleted, err: %w", bucketName, err)
	}

	klog.Infof("Bucket %s deleted", bucketName)

	return nil
}

// DeleteS3BucketAndObjects deletes all objects in the S3 bucket with the given name, and then the
// bucket. A bucket that does not exist is not an error.
func (c Client) DeleteS3BucketAndObjects(ctx context.Context, bucketName string) error {
	bucketName = strings.TrimPrefix(bucketName, "s3://")
	regionOpt := c.bucketRegionOption(ctx, bucketName)

	// A page holds at most 1000 keys, which is also the most that DeleteObjects accepts.
	paginator := s3.NewListObjectsV2Paginator(c.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	deleted := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx, regionOpt)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				klog.Infof("Bucket %s does not exist.", bucketName)

				return nil
			}

			return fmt.Errorf("listing objects in bucket %s: %w", bucketName, err)
		}
		if len(page.Contents) == 0 {
			continue
		}

		objects := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, object := range page.Contents {
			objects = append(objects, types.ObjectIdentifier{Key: object.Key})
		}
		resp, err := c.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		}, regionOpt)
		if err != nil {
			return fmt.Errorf("deleting objects in bucket %s: %w", bucketName, err)
		}
		if len(resp.Errors) > 0 {
			return fmt.Errorf("deleting object %q in bucket %s: %s", aws.ToString(resp.Errors[0].Key), bucketName, aws.ToString(resp.Errors[0].Message))
		}
		deleted += len(objects)
	}
	klog.Infof("Deleted %d objects from bucket %s", deleted, bucketName)

	return c.DeleteS3Bucket(ctx, bucketName)
}

// bucketRegionOption returns an option that sends requests to the region where the bucket resides,
// to avoid 301 PermanentRedirect errors. During teardown the deployer may derive the S3 client
// region from random zones, which can differ from the region where the bucket was created.
func (c Client) bucketRegionOption(ctx context.Context, bucketName string) func(*s3.Options) {
	bucketRegion, err := c.BucketRegion(ctx, bucketName)
	if err != nil {
		klog.Infof("Could not determine region for bucket %s: %v", bucketName, err)
	}

	return func(o *s3.Options) {
		if bucketRegion != "" {
			o.Region = bucketRegion
		}
	}
}

// BucketRegion resolves the AWS region where the bucket resides.
// GetBucketLocation is region-agnostic and can locate buckets in any region
// from any endpoint. We pin to defaultRegion for consistency.
func (c Client) BucketRegion(ctx context.Context, bucketName string) (string, error) {
	bucketName = strings.TrimPrefix(bucketName, "s3://")
	resp, err := c.s3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}, func(o *s3.Options) {
		o.Region = defaultRegion
	})
	if err != nil {
		return "", fmt.Errorf("getting bucket location for %s: %w", bucketName, err)
	}
	region := string(resp.LocationConstraint)
	switch resp.LocationConstraint {
	case "":
		// GetBucketLocation returns empty for us-east-1.
		return "us-east-1", nil
	case "EU":
		// GetBucketLocation may return the legacy EU alias for eu-west-1.
		return "eu-west-1", nil
	default:
		return region, nil
	}
}

func (c Client) setPublicReadPolicy(ctx context.Context, bucketName string) error {
	policy := fmt.Sprintf(`{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Sid": "PublicReadGetObject",
        "Effect": "Allow",
        "Principal": "*",
        "Action": "s3:GetObject",
        "Resource": "arn:aws:s3:::%s/*"
      }
    ]
  }`, bucketName)

	_, err := c.s3Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(policy),
	})
	if err != nil {
		return fmt.Errorf("putting bucket policy for %s: %w", bucketName, err)
	}

	return nil
}

func (c Client) setPublicAccessBlock(ctx context.Context, bucketName string) error {
	_, err := c.s3Client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(false),
			RestrictPublicBuckets: aws.Bool(false),
		},
	})
	return err
}

// SplitS3URL splits an s3://<bucket>/<prefix> location into the bucket name and the key prefix,
// without leading or trailing slashes.
func SplitS3URL(location string) (bucket, prefix string, err error) {
	rest, ok := strings.CutPrefix(location, "s3://")
	if !ok {
		return "", "", fmt.Errorf("%q is not an s3:// location", location)
	}
	bucket, prefix, _ = strings.Cut(rest, "/")
	if bucket == "" {
		return "", "", fmt.Errorf("%q does not contain a bucket name", location)
	}

	return bucket, strings.Trim(prefix, "/"), nil
}

// PublicObjectURL returns the https:// URL that serves urlPath from a publicly readable bucket in
// the given region. urlPath is kept as is, so a "+" stays literal; see k8s.io/release/pkg/build.S3ObjectKey
// for the key that S3 serves at that URL.
func PublicObjectURL(ctx context.Context, region, bucket, urlPath string) (string, error) {
	endpoint, err := s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, s3.EndpointParameters{
		Bucket: aws.String(bucket),
		Region: aws.String(region),
	})
	if err != nil {
		return "", fmt.Errorf("resolving S3 endpoint for bucket %q in region %q: %w", bucket, region, err)
	}

	u := endpoint.URI
	// The bucket is part of the path for buckets that cannot be addressed as a subdomain.
	u.Path = path.Join("/", u.Path, urlPath)

	return u.String(), nil
}

// EndpointObjectURL returns the https:// URL that serves urlPath from a publicly readable bucket at
// an S3-compatible endpoint outside AWS, which addresses the bucket as a subdomain.
func EndpointObjectURL(endpoint, bucket, urlPath string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid S3 endpoint %q", endpoint)
	}
	u.Host = bucket + "." + u.Host
	u.Path = path.Join("/", urlPath)

	return u.String(), nil
}
