/*
Copyright 2021 The Kubernetes Authors.

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

package gce

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/iam"
	"cloud.google.com/go/storage"
	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

const (
	defaultRegion = "us-central1"
)

func GCSBucketName(projectID, prefix string) string {
	var s string
	if jobID := os.Getenv("PROW_JOB_ID"); len(jobID) >= 4 {
		s = jobID[:4]
	} else {
		b := make([]byte, 2)
		rand.Read(b)
		s = hex.EncodeToString(b)
	}
	bucket := strings.Join([]string{projectID, prefix, s}, "-")
	return bucket
}

func EnsureGCSBucket(bucketPath, projectID string, public bool) error {
	// TODO: Detect the GCP region used
	return EnsureGCSBucketWithRegion(bucketPath, projectID, defaultRegion, public)
}

func EnsureGCSBucketWithRegion(bucketPath, projectID string, region string, public bool) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	// Extract bucket name from gs:// path if the bucket's URI is provided
	bucketName := strings.TrimPrefix(bucketPath, "gs://")
	bucketName = strings.TrimSuffix(bucketName, "/")

	bucket := client.Bucket(bucketName)

	// Check if bucket exists
	klog.Infof("Checking if bucket %s exists", bucketName)
	_, err = bucket.Attrs(ctx)
	if err == nil {
		klog.Infof("Bucket %s already exists", bucketName)
		return nil
	}

	// If error is not "bucket doesn't exist", return error
	if errors.Is(err, storage.ErrBucketNotExist) {
		return fmt.Errorf("error checking bucket: %w", err)
	}

	// Create the bucket
	klog.Infof("Creating bucket %s in project %s", bucketName, projectID)
	bucketAttrs := &storage.BucketAttrs{
		Location:     region,
		LocationType: "region",
		UniformBucketLevelAccess: storage.UniformBucketLevelAccess{
			Enabled: true,
		},
		SoftDeletePolicy: &storage.SoftDeletePolicy{
			RetentionDuration: 0,
		},
	}

	if err := bucket.Create(ctx, projectID, bucketAttrs); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	if public {
		klog.Infof("Making bucket %s public", bucketName)
		policy, err := bucket.IAM().Policy(ctx)
		if err != nil {
			return fmt.Errorf("failed to get bucket IAM policy: %w", err)
		}

		// Add allUsers as objectViewer
		policy.Add(iam.AllUsers, "roles/storage.objectViewer")

		if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
			return fmt.Errorf("failed to set bucket IAM policy: %w", err)
		}
	}

	return nil
}

func DeleteGCSBucket(bucketPath, projectID string) error {
	rmArgs := []string{
		"gsutil",
		"-u", projectID,
		"rm", "-r", bucketPath,
	}

	klog.Info(strings.Join(rmArgs, " "))
	cmd := exec.Command(rmArgs[0], rmArgs[1:]...)

	exec.InheritOutput(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
