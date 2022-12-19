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
	"os"
	"strings"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/exec"
)

// GCSBucketName constructs a bucket name to use for the state store.
// Note that this must be deterministic across invocations of kubetest2,
// otherwise we won't use the same bucket between runs.
func GCSBucketName(projectID string) string {
	if jobID := os.Getenv("PROW_JOB_ID"); len(jobID) >= 2 {
		s := jobID[:2]
		bucket := strings.Join([]string{projectID, "state", s}, "-")
		return bucket
	} else {
		bucket := "kops-state-" + projectID
		return bucket
	}
}

func EnsureGCSBucket(bucketPath, projectID string) error {
	lsArgs := []string{
		"gsutil", "ls", "-b",
	}
	if projectID != "" {
		lsArgs = append(lsArgs, "-p", projectID)
	}
	lsArgs = append(lsArgs, bucketPath)

	klog.Info(strings.Join(lsArgs, " "))
	cmd := exec.Command(lsArgs[0], lsArgs[1:]...)

	output, err := exec.CombinedOutputLines(cmd)
	if err == nil {
		return nil
	} else if len(output) != 1 || !strings.Contains(output[0], "BucketNotFound") {
		klog.Info(output)
		return err
	}

	mbArgs := []string{
		"gsutil", "mb",
	}
	if projectID != "" {
		mbArgs = append(mbArgs, "-p", projectID)
	}
	mbArgs = append(mbArgs, bucketPath)

	klog.Info(strings.Join(mbArgs, " "))
	cmd = exec.Command(mbArgs[0], mbArgs[1:]...)

	exec.InheritOutput(cmd)
	err = cmd.Run()
	if err != nil {
		return err
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
