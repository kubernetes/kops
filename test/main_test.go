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

////
// This awesome tweaked code is from the heart and mind of @kris-nova
////

package test

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"testing"
)

var TestClusterName, TestStateStore, HomeDirectory string
var TestVerbosity int
var KopsPath = "kops"

func TestMain(m *testing.M) {
	flag.Parse()

	existingPath := os.Getenv("PATH")
	os.Setenv("PATH", existingPath+":/go/bin/")

	// Get the users $HOME directory
	usr, err := user.Current()
	if err != nil {
		banner("Unable to get current user!")
		os.Exit(-1)
	}
	HomeDirectory = usr.HomeDir
	EnsurePublicKey(HomeDirectory + "/.ssh/id_rsa.pub")

	kopsTesting := &KopsTest{}
	kopsTesting, err = kopsTesting.Pre()

	if err != nil {
		banner(fmt.Sprintf("Pre test Failure: %v", err))
		os.Exit(-1)
	}

	TestClusterName = kopsTesting.ClusterName
	TestStateStore = kopsTesting.StateStore
	TestVerbosity = kopsTesting.Verbosity

	banner(fmt.Sprintf("kops testing: %+v", kopsTesting))

	// Create cluster
	err = CreateCluster(kopsTesting)

	if err != nil {
		banner(fmt.Sprintf("Create Cluster Failure: %v", err))
		err = kopsTesting.Post()
		if err != nil {
			banner(fmt.Sprintf("Test Cleanup Failure: %v", err))
		}

		os.Exit(-1)
	}

	banner(fmt.Sprintf("Created Cluster: %s", TestClusterName))

	err = Validate()

	if err != nil {
		banner(fmt.Sprintf("Validate Cluster Failure: %v", err))
		err = kopsTesting.Post()
		if err != nil {
			banner(fmt.Sprintf("Test Cleanup Failure: %v", err))
		}
		os.Exit(-1)
	}

	banner(fmt.Sprintf("Validated Cluster: %s", TestClusterName))

	// Run tests
	n := m.Run()

	// Delete cluster
	err = DeleteCluster(kopsTesting)
	if err != nil {
		banner(fmt.Sprintf("Delete Cluster Failure: %v", err))
		err := kopsTesting.Post()
		if err != nil {
			banner(fmt.Sprintf("Test Cleanup Failure: %v", err))
		}
		os.Exit(-1)
	}
	banner(fmt.Sprintf("Deleted Cluster: %s", TestClusterName))
	kopsTesting.Post()
	if err != nil {
		banner(fmt.Sprintf("Test Cleanup Failure: %v", err))
		os.Exit(-1)
	}
	banner(fmt.Sprintf("Deleted StateStore S3 bucket: %s", TestClusterName))
	os.Exit(n)
}

func banner(msg string) {
	fmt.Println("---------------------------------------------------------------------------------------------------------")
	fmt.Println(msg)
	fmt.Println("---------------------------------------------------------------------------------------------------------")
}

const CREATE_CLUSTER = `create cluster
--name %s
--state %s
--node-count %d
--zones %s
--master-zones %s
--cloud aws
--node-size %s
--master-size %s
--topology %s
--networking %s
--kubernetes-version %s
-v %d
--image %s
--yes`

const DELETE_CLUSTER = `delete cluster
--name %s
--state %s
-v %d
--yes`

// CreateCluster will actually create a Kops kubernetes cluster
// and store the name - we should be able to use the name for other test cases
// This is the point in the test case where we actually create a cluster :)
func CreateCluster(kopsTesting *KopsTest) error {

	kopsCreateCommand := fmt.Sprintf(CREATE_CLUSTER,
		kopsTesting.ClusterName,
		kopsTesting.StateStore,
		kopsTesting.NodeCount,
		kopsTesting.NodeZones,
		kopsTesting.MasterZones,
		kopsTesting.NodeSize,
		kopsTesting.MasterSize,
		kopsTesting.Topology,
		kopsTesting.Networking,
		kopsTesting.K8sVersion,
		kopsTesting.Verbosity,
		kopsTesting.Image)

	env := []string{
		fmt.Sprintf("NODE_UP_URL=%s", kopsTesting.NodeUpURL),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}

	banner(kopsCreateCommand)
	stdoutExec, err := ExecOutput(KopsPath, kopsCreateCommand, env)

	if err != nil {
		return fmt.Errorf("Unable to create cluster: %v\n", err)
	}

	banner(stdoutExec)
	TestClusterName = kopsTesting.ClusterName
	return nil
}

// TestCliDeleteClusterHappy will actually delete the cluster created earlier in the
// testing process
func DeleteCluster(kopsTest *KopsTest) error {
	kopsDeleteCommand := fmt.Sprintf(DELETE_CLUSTER, kopsTest.ClusterName, kopsTest.StateStore, kopsTest.Verbosity)
	banner(kopsDeleteCommand)
	stdoutExec, err := ExecOutput(KopsPath, kopsDeleteCommand, []string{})
	if err != nil {
		return fmt.Errorf("Unable to delete cluster: %v\n", err)
	}
	banner(stdoutExec)
	return nil
}
