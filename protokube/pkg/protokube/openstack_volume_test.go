/*
Copyright 2022 The Kubernetes Authors.

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

package protokube

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"k8s.io/mount-utils"
	ue "k8s.io/utils/exec"
	uet "k8s.io/utils/exec/testing"
)

const (
	metadataIDFirst    = MetadataID + ", " + ConfigDriveID
	configDriveIDFirst = ConfigDriveID + ", " + MetadataID
)

// expectedDriveMetadata is the metadata expected from the service endpoint
var expectedServiceMetadata = &InstanceMetadata{
	ServerID:         "01234567-cafe-babe-beef-0123456789ab",
	Hostname:         "test-server.serv.ice",
	Name:             "test-server",
	AvailabilityZone: "nova",
	ProjectID:        "0123456789abcdeffedcba9876543210",
}

// expectedDriveMetadata is the metadata expected from the config drive
var expectedDriveMetadata = &InstanceMetadata{
	ServerID:         "01234567-cafe-babe-beef-0123456789ab",
	Hostname:         "test-server.config.drv",
	Name:             "test-server",
	AvailabilityZone: "nova",
	ProjectID:        "0123456789abcdeffedcba9876543210",
}

func assertTestResults(t *testing.T, err error, expected interface{}, actual interface{}) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %+v, but got %+v", expected, actual)
	}
}

// mockMetadataEndpoint mocks the actual metadata endpoint by returning the passed data
func mockMetadataEndpoint(w http.ResponseWriter, r *http.Request, data string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, data)
}

// createMockServer creates and returns a mock HTTP server to test getting the metadata from a service endpoint
func createMockServer() (*httptest.Server, error) {
	data, err := os.ReadFile("testdata/metadata_service.json")
	if err != nil {
		return nil, err
	}

	// Note(ederst): source of inspiration https://clavinjune.dev/en/blogs/mocking-http-call-in-golang-a-better-way/
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/openstack/latest/meta_data.json") {
			mockMetadataEndpoint(w, r, string(data))
		} else {
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))

	return mockServer, nil
}

// fakeBlkidCmd builds a fake blkid command which returns the device path of the config drive
func fakeBlkidCmd(device string) uet.FakeCommandAction {
	fakeCmd := &uet.FakeCmd{
		CombinedOutputScript: []uet.FakeAction{
			func() ([]byte, []byte, error) {
				if device == "" {
					return nil, nil, uet.FakeExitError{
						Status: 2,
					}
				}
				return []byte(device), nil, nil
			},
		},
	}

	return func(cmd string, args ...string) ue.Cmd {
		return uet.InitFakeCmd(fakeCmd, "blkid", "-l", "-t", "LABEL=config-2", "-o", "device")
	}
}

// getFakeMounter creates and returns a fake mounter to test getting the metadata from a config drive
func getFakeMounter(device string) *mount.SafeFormatAndMount {
	fakeExec := &uet.FakeExec{
		ExactOrder: true,
	}

	fakeExec.CommandScript = append(fakeExec.CommandScript, fakeBlkidCmd(device))

	return &mount.SafeFormatAndMount{
		Interface: &mount.FakeMounter{},
		Exec:      fakeExec,
	}
}

func TestGetMetadataFromMetadataServiceReturnsNotFoundError(t *testing.T) {
	mockServer, err := createMockServer()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer mockServer.Close()
	metadataUrl := fmt.Sprintf("%s/%s", mockServer.URL, "no/meta_data.json")

	expectedErr := fmt.Errorf("fetching metadata from '%s' returned status code '404'", metadataUrl)

	_, actualErr := newMetadataService(metadataUrl, MetadataLatestPath, nil, "", MetadataID).getMetadata()

	assertTestResults(t, nil, expectedErr, actualErr)
}

func TestGetMetadataFromMetadataService(t *testing.T) {
	mockServer, err := createMockServer()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer mockServer.Close()
	metadataUrl := fmt.Sprintf("%s/%s", mockServer.URL, MetadataLatestPath)

	actual, err := newMetadataService(metadataUrl, MetadataLatestPath, nil, "", MetadataID).getMetadata()

	assertTestResults(t, err, expectedServiceMetadata, actual)
}

func TestGetMetadataFromConfigDriveReturnsErrorWhenNoDeviceIsFound(t *testing.T) {
	fakeMounter := getFakeMounter("")

	expectedErr := fmt.Errorf("unable to run blkid: exit 2")

	_, actualErr := newMetadataService("", "testdata/metadata_drive.json", fakeMounter, ".", ConfigDriveID).getMetadata()

	assertTestResults(t, nil, expectedErr, actualErr)
}

func TestGetMetadataFromConfigDrive(t *testing.T) {
	fakeMounter := getFakeMounter("/dev/sr0")

	actual, err := newMetadataService("", "testdata/metadata_drive.json", fakeMounter, ".", ConfigDriveID).getMetadata()

	assertTestResults(t, err, expectedDriveMetadata, actual)
}

func TestGetMetadataReturnsLastErrorWhenNoMetadataWasFound(t *testing.T) {
	fakeMounter := getFakeMounter("")

	mockServer, err := createMockServer()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer mockServer.Close()
	metadataUrl := fmt.Sprintf("%s/%s", mockServer.URL, "no/meta_data.json")

	expectedErr := fmt.Errorf("fetching metadata from '%s' returned status code '404'", metadataUrl)

	_, actualErr := newMetadataService(metadataUrl, "testdata/metadata_drive.json", fakeMounter, ".", configDriveIDFirst).getMetadata()

	assertTestResults(t, nil, expectedErr, actualErr)
}

func TestGetMetadataFromConfigDriveWhenItIsFirstInSearchOrder(t *testing.T) {
	fakeMounter := getFakeMounter("/dev/sr0")

	mockServer, err := createMockServer()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer mockServer.Close()
	metadataUrl := fmt.Sprintf("%s/%s", mockServer.URL, MetadataLatestPath)

	actual, err := newMetadataService(metadataUrl, "testdata/metadata_drive.json", fakeMounter, ".", configDriveIDFirst).getMetadata()

	assertTestResults(t, err, expectedDriveMetadata, actual)
}

func TestGetMetadataFromServiceEndpointWhenConfigDriveFails(t *testing.T) {
	fakeMounter := getFakeMounter("")

	mockServer, err := createMockServer()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer mockServer.Close()
	metadataUrl := fmt.Sprintf("%s/%s", mockServer.URL, MetadataLatestPath)

	actual, err := newMetadataService(metadataUrl, "testdata/metadata_drive.json", fakeMounter, ".", configDriveIDFirst).getMetadata()

	assertTestResults(t, err, expectedServiceMetadata, actual)
}

func TestGetMetadataFromServiceEndpointWhenItIsFirstInSearchOrder(t *testing.T) {
	fakeMounter := getFakeMounter("/dev/sr0")

	mockServer, err := createMockServer()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer mockServer.Close()
	metadataUrl := fmt.Sprintf("%s/%s", mockServer.URL, MetadataLatestPath)

	actual, err := newMetadataService(metadataUrl, "testdata/metadata_drive.json", fakeMounter, ".", metadataIDFirst).getMetadata()

	assertTestResults(t, err, expectedServiceMetadata, actual)
}

func TestGetMetadataFromConfigDriveWhenServiceEndpointFails(t *testing.T) {
	fakeMounter := getFakeMounter("/dev/sr0")

	mockServer, err := createMockServer()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer mockServer.Close()
	metadataUrl := fmt.Sprintf("%s/%s", mockServer.URL, "no/meta_data.json")

	actual, err := newMetadataService(metadataUrl, "testdata/metadata_drive.json", fakeMounter, ".", metadataIDFirst).getMetadata()

	assertTestResults(t, err, expectedDriveMetadata, actual)
}
