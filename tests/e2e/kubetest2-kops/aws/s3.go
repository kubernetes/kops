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

// We need to pick some region to query the AWS APIs through, even if we are not running on AWS.
const defaultRegion = "us-east-2"

// It contains S3Client, an Amazon S3 service client that is used to perform bucket
// and object actions.
type awsClient struct {
	S3Client *s3.Client
}

func NewAWSClient(ctx context.Context) (*awsClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(defaultRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return &awsClient{
		S3Client: s3.NewFromConfig(cfg),
	}, nil
}

// AWSBucketName constructs an unique bucket name using the AWS account ID on region us-east-2
func AWSBucketName(ctx context.Context) (string, error) {
	config, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(string(types.BucketLocationConstraintUsEast2)))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsSvc := sts.NewFromConfig(config)
	callerIdentity, err := stsSvc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("building AWS STS presigned request: %w", err)
	}

	// Add timestamp suffix
	timestamp := time.Now().Format("01022006")
	bucket := fmt.Sprintf("k8s-infra-kops-%s", *callerIdentity.Account)
	bucket = fmt.Sprintf("%s-%s", bucket, timestamp)

	bucket = strings.ToLower(bucket)
	bucket = regexp.MustCompile("[^a-z0-9-]").ReplaceAllString(bucket, "") // Only allow lowercase letters, numbers, and hyphens

	if len(bucket) > 63 {
		bucket = bucket[:63] // Max length is 63
	}

	return bucket, nil
}

func (client awsClient) EnsureS3Bucket(ctx context.Context, bucketName string, publicRead bool) error {
	_, err := client.S3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraintUsEast2,
		},
	})

	var exists *types.BucketAlreadyExists
	if err != nil {
		if errors.As(err, &exists) {
			klog.Infof("Bucket %s already exists.\n", bucketName)
			err = exists
		}
	} else {
		err := s3.NewBucketExistsWaiter(client.S3Client).Wait(
			ctx, &s3.HeadBucketInput{
				Bucket: aws.String(bucketName),
			},
			time.Minute)
		if err != nil {
			klog.Infof("Failed attempt to wait for bucket %s to exist.", bucketName)
		}
	}

	klog.Infof("Bucket %s created successfully", bucketName)

	if err == nil && publicRead {
		err = client.setPublicReadPolicy(ctx, bucketName)
		if err != nil {
			klog.Errorf("Failed to set public read policy on bucket %s: %v", bucketName, err)
			return err
		}
		klog.Infof("Public read policy set on bucket %s", bucketName)
	}

	return err
}

func (client awsClient) DeleteS3Bucket(ctx context.Context, bucketName string) error {
	_, err := client.S3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		var noBucket *types.NoSuchBucket
		if errors.As(err, &noBucket) {
			klog.Infof("Bucket %s does not exits", bucketName)
			err = noBucket
		} else {
			klog.Infof("Couldn't delete bucket %s. Reason: %v", bucketName, err)
		}
	} else {
		err = s3.NewBucketNotExistsWaiter(client.S3Client).Wait(
			ctx, &s3.HeadBucketInput{
				Bucket: aws.String(bucketName),
			},
			time.Minute)
		if err != nil {
			klog.Infof("Failed attempt to wait for bucket %s to be deleted", bucketName)
		} else {
			klog.Infof("Bucket %s deleted", bucketName)
		}
	}
	return err
}

func (client awsClient) setPublicReadPolicy(ctx context.Context, bucketName string) error {
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

	_, err := client.S3Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(policy),
	})
	if err != nil {
		return fmt.Errorf("failed to put bucket policy for %s: %w", bucketName, err)
	}

	return err
}
