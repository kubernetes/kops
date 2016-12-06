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

package test

import (
	"fmt"
	"strings"
	"os"

	"k8s.io/kops/util/pkg/vfs"
)

// KopsTest
type KopsTest struct {
	ClusterName  string
	StateStore   string
	S3BucketName string

	Image      string
	DomainName string
	NodeUpURL  string

	NodeCount   int
	MasterCount int
	NodeZones   string
	MasterZones string
	NodeSize    string
	MasterSize  string
	Verbosity   int
	Networking  string
	Topology    string

	K8sVersion string
	AWSRegion  string
}

// Basic Pre Test
func (t *KopsTest) Pre() (*KopsTest, error) {
	// TODO: setup dynamically
	t.NodeCount = 3
	t.MasterCount = 3
	t.NodeZones = "eu-west-1a,eu-west-1b,eu-west-1c"
	t.MasterZones = "eu-west-1a,eu-west-1b,eu-west-1c"
	t.NodeSize = "m3.large"
	t.MasterSize = "m3.large"
	t.Verbosity = 10
	t.Networking = "weave"
	t.Topology = "private"
	t.AWSRegion = "eu-west-1"

	t, err := t.basicPreCheck()
	if err != nil {
		return nil, fmt.Errorf("error in precheck %v", err)
	}
	t, err = t.createBucketAndNames()
	if err != nil {
		return nil, fmt.Errorf("error in create bucket %v", err)
	}

	return t, nil
}

// Basic Post Test
func (t *KopsTest) Post() error {
	if t.S3BucketName == "" {
		return fmt.Errorf("error in deleting bucket bucket name not defined")
	}
	err := t.deleteBucket()
	if err != nil {
		return fmt.Errorf("error in deleting bucket %v", err)
	}
	return nil
}

// Base setup function to check that a template, and nic information is set
func (t *KopsTest) basicPreCheck() (*KopsTest, error) {

	if v := os.Getenv("KOPS_TEST_IMAGE"); v == "" {
		return nil, fmt.Errorf("env variable KOPS_TEST_IMAGE must be set for acceptance tests")
	}

	t.Image = os.Getenv("KOPS_TEST_IMAGE")

	if v := os.Getenv("KOPS_TEST_DOMAIN"); v == "" {
		return nil, fmt.Errorf("env variable KOPS_TEST_DOMAIN must be set for acceptance tests")
	}

	t.DomainName = os.Getenv("KOPS_TEST_DOMAIN")

	if v := os.Getenv("KOPS_TEST_NODEUP_URL"); v == "" {
		return nil, fmt.Errorf("env variable KOPS_TEST_NODEUP_URL must be set for acceptance tests")
	}

	t.NodeUpURL = os.Getenv("KOPS_TEST_NODEUP_URL")

	if v := os.Getenv("KOPS_TEST_K8S_VERSION"); v == "" {
		return nil, fmt.Errorf("env variable KOPS_TEST_K8S_VERSION must be set for acceptance tests")
	}

	t.K8sVersion = os.Getenv("KOPS_TEST_K8S_VERSION")

	return t, nil
}

func (t *KopsTest) createBucketAndNames() (*KopsTest, error) {

	name, err := GetRandomClusterName(t.DomainName)
	if err != nil {
		return nil, fmt.Errorf("Unable to create name: %v", err)
	}

	bucketName := "kops-test-" + strings.Replace(name, "."+t.DomainName, "", -1)
	s3Context := vfs.NewS3Context()

	s3 := vfs.NewS3PathWithRegion(s3Context, bucketName, t.AWSRegion)

	// TODO pass in region
	err = s3.EnsureBucketExists(bucketName, t.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("Unable to create s3 bucket: %v", err)
	}

	t.S3BucketName = bucketName
	t.StateStore = "s3://" + bucketName // refactor to a method

	t.ClusterName = fmt.Sprintf("kops-test-%s", name)
	return t, nil
}

func (t *KopsTest) deleteBucket() error {

	s3Context := vfs.NewS3Context()

	s3 := vfs.NewS3PathWithRegion(s3Context, t.S3BucketName, t.AWSRegion)

	err := s3.DeleteBucket(t.S3BucketName)
	if err != nil {
		return fmt.Errorf("Unable to create s3 bucket: %v", err)
	}

	return nil

}
