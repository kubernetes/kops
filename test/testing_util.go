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
	"os"
	"os/exec"

	"fmt"

	"k8s.io/kops/util/pkg/vfs"
	"strings"
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

	K8sVersion  string
}

// Basic Pre Test
func (t *KopsTest) Pre() (*KopsTest, error) {
	t, err := t.basicPreCheck()
	if err != nil {
		return nil, fmt.Errorf("error in precheck %v", err)
	}
	t, err = t.createBucket()
	if err != nil {
		return nil, fmt.Errorf("error in create bucket %v", err)
	}
	t, err = t.createClusterName()
	if err != nil {
		return nil, fmt.Errorf("error in create cluster name %v", err)
	}

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

	return t, nil
}

// Basic Post Test
func (t *KopsTest) Post() error {
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

	t.NodeUpURL = os.Getenv("KOPS_NODEUP_URL")

	if v := os.Getenv("KOPS_TEST_K8S_VERSION"); v == "" {
		return nil, fmt.Errorf("env variable KOPS_TEST_K8S_VERSION must be set for acceptance tests")
	}

	t.K8sVersion = os.Getenv("KOPS_TEST_K8S_VERSION")

	return t, nil
}

func (t *KopsTest) createBucket() (*KopsTest, error) {

	bytes, err := exec.Command("uuidgen").Output()
	if err != nil {
		return nil, fmt.Errorf("Unable to create s3 bucket name: %v", err)
	}
	bucketName := strings.ToLower("kops-testing-" + string(bytes[:])) + "/"
	vfsContext := vfs.NewS3Context()

	// TODO: pass in region or pick it up from profile
	vfsPath := vfs.NewS3PathWithRegion(vfsContext, bucketName, "key","us-west-1")

	err = vfsPath.CreateNewBucket()
	if err != nil {
		return  nil,fmt.Errorf("Unable to create s3 bucket: %v", err)
	}

	t.StateStore = vfsPath.Path()
	return t, nil
}

func (t *KopsTest) createClusterName() (*KopsTest, error) {

	bytes, err := exec.Command("uuidgen").Output()
	if err != nil {
		return  nil,fmt.Errorf("Unable to create s3 bucket name: %v", err)
	}

	cluster := strings.ToLower("kops-testing-cluser"+string(bytes[:]))

	t.ClusterName = fmt.Sprintf("kops-testing-%s.%s", cluster, t.DomainName)
	return t, nil
}

func (t *KopsTest) deleteBucket() error {

	s3Context := vfs.NewS3Context()

	s3 := vfs.NewS3Path(s3Context, t.S3BucketName, "key")

	err := s3.DeleteBucket()
	if err != nil {
		return fmt.Errorf("Unable to create s3 bucket: %v", err)
	}

	return nil

}
