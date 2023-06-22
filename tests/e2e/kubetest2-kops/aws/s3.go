/*
Copyright 2023 The Kubernetes Authors.

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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"k8s.io/klog/v2"
)

// We need to pick some region to query the AWS APIs through, even if we are not running on AWS.
const defaultRegion = "us-east-2"

type awsClient struct {
	sts *sts.STS
	s3  *s3.S3
}

func newAWSClient(ctx context.Context, creds *credentials.Credentials) (*awsClient, error) {
	awsConfig := aws.NewConfig().WithRegion(defaultRegion).WithUseDualStack(true)
	awsConfig = awsConfig.WithCredentialsChainVerboseErrors(true)
	if creds != nil {
		awsConfig = awsConfig.WithCredentials(creds)
	}

	awsSession, err := session.NewSessionWithOptions(session.Options{
		Config:            *awsConfig,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("error starting new AWS session: %v", err)
	}

	return &awsClient{
		sts: sts.New(awsSession, awsConfig),
		s3:  s3.New(awsSession, awsConfig),
	}, nil
}

// AWSBucketName constructs a bucket name that is unique to the AWS account.
func AWSBucketName(ctx context.Context, creds *credentials.Credentials) (string, error) {
	client, err := newAWSClient(ctx, creds)
	if err != nil {
		return "", err
	}

	callerIdentity, err := client.sts.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("error getting AWS caller identity from STS: %w", err)
	}
	bucket := "kops-test-" + aws.StringValue(callerIdentity.Account)
	return bucket, nil
}

// EnsureAWSBucket creates a bucket if it does not exist in the account.
// If a different account has already created the bucket, that is treated as an error to prevent "preimage" attacks.
func EnsureAWSBucket(ctx context.Context, creds *credentials.Credentials, bucketName string) error {
	// These don't need to be in the same region, so we pick a region arbitrarily
	location := "us-east-2"

	client, err := newAWSClient(ctx, creds)
	if err != nil {
		return err
	}

	// Note that this lists only our buckets, so we know that someone else hasn't created the bucket
	buckets, err := client.s3.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("error listing buckets: %w", err)
	}

	var existingBucket *s3.Bucket
	for _, bucket := range buckets.Buckets {
		if aws.StringValue(bucket.Name) == bucketName {
			existingBucket = bucket
		}
	}

	if existingBucket == nil {
		klog.Infof("creating S3 bucket s3://%s", bucketName)
		if _, err := client.s3.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
			Bucket: &bucketName,
			CreateBucketConfiguration: &s3.CreateBucketConfiguration{
				LocationConstraint: &location,
			},
		}); err != nil {
			return fmt.Errorf("error creating bucket s3://%v: %w", bucketName, err)
		}
	}

	return nil
}

// DeleteAWSBucket deletes an AWS bucket.
func DeleteAWSBucket(ctx context.Context, creds *credentials.Credentials, bucketName string) error {
	client, err := newAWSClient(ctx, creds)
	if err != nil {
		return err
	}

	klog.Infof("deleting S3 bucket s3://%s", bucketName)
	if _, err := client.s3.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{Bucket: &bucketName}); err != nil {
		return fmt.Errorf("error deleting bucket: %w", err)
	}
	return nil
}
