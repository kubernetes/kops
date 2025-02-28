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

// Client contains S3 and STS clients that are used to perform bucket and object actions.
type Client struct {
	s3Client  *s3.Client
	stsClient *sts.Client
}

// NewAWSClient returns a new instance of awsClient configured to work in the default region (us-east-2).
func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(defaultRegion))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	return &Client{
		s3Client:  s3.NewFromConfig(cfg),
		stsClient: sts.NewFromConfig(cfg),
	}, nil
}

// BucketName constructs an unique bucket name using the AWS account ID in the default region (us-east-2).
func (c Client) BucketName(ctx context.Context) (string, error) {
	callerIdentity, err := c.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("building AWS STS presigned request: %w", err)
	}

	// Construct the bucket name based on the AWS account ID and the current timestamp
	timestamp := time.Now().Format("20060102150405")
	bucket := fmt.Sprintf("k8s-infra-kops-%s-%s", *callerIdentity.Account, timestamp)

	bucket = strings.ToLower(bucket)
	// Only allow lowercase letters, numbers, and hyphens
	bucket = regexp.MustCompile("[^a-z0-9-]").ReplaceAllString(bucket, "")

	if len(bucket) > 63 {
		bucket = bucket[:63] // Max length is 63
	}

	return bucket, nil
}

// EnsureS3Bucket creates a new S3 bucket with the given name and public read permissions.
func (c Client) EnsureS3Bucket(ctx context.Context, bucketName string, publicRead bool) error {
	_, err := c.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: defaultRegion,
		},
	})
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
	_, err := c.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var noBucket *types.NoSuchBucket
		if errors.As(err, &noBucket) {
			klog.Infof("Bucket %s does not exits.", bucketName)

			return nil
		} else {
			klog.Infof("Couldn't delete bucket %s, err: %v", bucketName, err)

			return fmt.Errorf("deleting bucket %s: %w", bucketName, err)
		}
	}

	err = s3.NewBucketNotExistsWaiter(c.s3Client).Wait(
		ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		},
		time.Minute)
	if err != nil {
		klog.Infof("Failed attempt to wait for bucket %s to be deleted, err: %v", bucketName, err)

		return fmt.Errorf("waiting for bucket %s to be deleted, err: %w", bucketName, err)
	}

	klog.Infof("Bucket %s deleted", bucketName)

	return nil
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
