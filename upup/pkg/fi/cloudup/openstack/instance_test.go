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

package openstack

import (
	"fmt"
	"testing"

	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"k8s.io/kops/upup/pkg/fi"
)

type WaitForStatusActiveMock struct {
	server *servers.Server
	OpenstackCloud
}

func (c WaitForStatusActiveMock) GetInstance(id string) (*servers.Server, error) {
	if id == c.server.ID {
		return c.server, nil
	}

	return nil, fmt.Errorf("Server with ID '%s' not found.", id)
}

func createWaitForStatusActiveMock(serverID string, serverStatus string) *WaitForStatusActiveMock {
	return &WaitForStatusActiveMock{
		server: &servers.Server{
			ID:     serverID,
			Status: serverStatus,
		},
	}
}

func Test_WaitForStatusActiveIsSuccessful(t *testing.T) {
	serverID := "mock-id"
	c := createWaitForStatusActiveMock(serverID, activeStatus)

	err := waitForStatusActive(c, serverID, nil)

	assertTestResults(t, err, nil, nil)
}

func Test_WaitForStatusActiveResultsInInstanceNotFound(t *testing.T) {
	serverID := "mock-id"
	c := createWaitForStatusActiveMock(serverID, activeStatus)

	wrongServerID := "wrong-id"
	actualErr := waitForStatusActive(c, wrongServerID, nil)

	expectedErr := fmt.Errorf("Server with ID '%s' not found.", wrongServerID)
	assertTestResults(t, nil, actualErr, expectedErr)
}

func Test_WaitForStatusActiveResultsInUnableToCreateServer(t *testing.T) {
	serverID := "mock-id"
	c := createWaitForStatusActiveMock(serverID, errorStatus)

	actualErr := waitForStatusActive(c, serverID, nil)

	expectedErr := fmt.Errorf("unable to create server: {0 0001-01-01 00:00:00 +0000 UTC  }")
	assertTestResults(t, nil, actualErr, expectedErr)
}

func Test_WaitForStatusActiveResultsInTimeout(t *testing.T) {
	serverID := "mock-id"
	c := createWaitForStatusActiveMock(serverID, "BUILD")

	actualErr := waitForStatusActive(c, serverID, fi.PtrTo(time.Second))

	expectedErr := fmt.Errorf("A timeout occurred")
	assertTestResults(t, nil, actualErr, expectedErr)
}
