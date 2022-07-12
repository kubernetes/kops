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

package scaleway

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/pkg/apis/kops"
)

const (
	defaultInstanceWaitRetryInterval = 5 * time.Second
	defaultInstanceWaitTimeout       = 10 * time.Minute
)

func reachState(instanceAPI *instance.API, zone scw.Zone, serverID string, toState instance.ServerState) error {
	// TODO(Mia-Cross): this function is not that useful, remove it
	response, err := instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: serverID,
	})
	if err != nil {
		return err
	}
	fromState := response.Server.State

	if response.Server.State == toState {
		return nil
	}

	transitionMap := map[[2]instance.ServerState][]instance.ServerAction{
		{instance.ServerStateStopped, instance.ServerStateRunning}:        {instance.ServerActionPoweron},
		{instance.ServerStateStopped, instance.ServerStateStoppedInPlace}: {instance.ServerActionPoweron, instance.ServerActionStopInPlace},
		{instance.ServerStateRunning, instance.ServerStateStopped}:        {instance.ServerActionPoweroff},
		{instance.ServerStateRunning, instance.ServerStateStoppedInPlace}: {instance.ServerActionStopInPlace},
		{instance.ServerStateStoppedInPlace, instance.ServerStateRunning}: {instance.ServerActionPoweron},
		{instance.ServerStateStoppedInPlace, instance.ServerStateStopped}: {instance.ServerActionPoweron, instance.ServerActionPoweroff},
	}

	actions, exist := transitionMap[[2]instance.ServerState{fromState, toState}]
	if !exist {
		return fmt.Errorf("don't know how to reach state %s from state %s for server %s", toState, fromState, serverID)
	}

	retryInterval := defaultInstanceWaitRetryInterval

	// We need to check that all volumes are ready
	for _, volume := range response.Server.Volumes {
		if volume.State != instance.VolumeServerStateAvailable {
			_, err = instanceAPI.WaitForVolume(&instance.WaitForVolumeRequest{
				Zone:          zone,
				VolumeID:      volume.ID,
				RetryInterval: &retryInterval,
			})
			if err != nil {
				return err
			}
		}
	}

	for _, a := range actions {
		err = instanceAPI.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			ServerID:      serverID,
			Action:        a,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(defaultInstanceWaitTimeout),
			RetryInterval: &retryInterval,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func WaitForInstanceServer(api *instance.API, zone scw.Zone, id string) (*instance.Server, error) {
	// TODO(Mia-Cross): this function is not that useful, replace all calls by api.WaitForServer ??
	retryInterval := defaultInstanceWaitRetryInterval
	timeout := defaultInstanceWaitTimeout

	server, err := api.WaitForServer(&instance.WaitForServerRequest{
		Zone:          zone,
		ServerID:      id,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	})

	return server, err
}

// isHTTPCodeError returns true if err is an http error with code statusCode
func isHTTPCodeError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	responseError := &scw.ResponseError{}
	if errors.As(err, &responseError) && responseError.StatusCode == statusCode {
		return true
	}
	return false
}

// is404Error returns true if err is an HTTP 404 error
func is404Error(err error) bool {
	notFoundError := &scw.ResourceNotFoundError{}
	return isHTTPCodeError(err, http.StatusNotFound) || errors.As(err, &notFoundError)
}

// parseZonedID parses a zonedID and extracts the resource zone and id.
func parseZonedID(zonedID string) (zone scw.Zone, id string, err error) {
	tab := strings.Split(zonedID, "/")
	if len(tab) != 2 {
		return "", zonedID, fmt.Errorf("can't parse zoned id: %s", zonedID)
	}
	locality := tab[0]
	id = tab[1]
	zone, err = scw.ParseZone(locality)
	return zone, id, err
}

func displayEnv() {
	fmt.Printf("******************* Scaleway credentials *******************\n\n")

	fmt.Printf(fmt.Sprintf("SCW_ACCESS_KEY = %s\n", os.Getenv("SCW_ACCESS_KEY")))
	fmt.Printf(fmt.Sprintf("SCW_SECRET_KEY = %s\n", os.Getenv("SCW_SECRET_KEY")))
	fmt.Printf(fmt.Sprintf("SCW_DEFAULT_REGION = %s\n", os.Getenv("SCW_DEFAULT_REGION")))
	fmt.Printf(fmt.Sprintf("SCW_DEFAULT_ZONE = %s\n", os.Getenv("SCW_DEFAULT_ZONE")))
	fmt.Printf(fmt.Sprintf("SCW_DEFAULT_PROJECT_ID = %s\n", os.Getenv("SCW_DEFAULT_PROJECT_ID")))

	fmt.Printf("\n********************* S3 credentials *********************\n\n")

	fmt.Printf(fmt.Sprintf("S3_REGION = %s\n", os.Getenv("S3_REGION")))
	fmt.Printf(fmt.Sprintf("S3_ENDPOINT = %s\n", os.Getenv("S3_ENDPOINT")))
	fmt.Printf(fmt.Sprintf("S3_ACCESS_KEY_ID = %s\n", os.Getenv("S3_ACCESS_KEY_ID")))
	fmt.Printf(fmt.Sprintf("S3_SECRET_ACCESS_KEY = %s\n", os.Getenv("S3_SECRET_ACCESS_KEY")))

	fmt.Printf("\n\t*********** State-store bucket *************\n\n")

	fmt.Printf(fmt.Sprintf("KOPS_STATE_STORE = %s\n", os.Getenv("KOPS_STATE_STORE")))
	fmt.Printf(fmt.Sprintf("S3_BUCKET_NAME = %s\n", os.Getenv("S3_BUCKET_NAME")))

	fmt.Printf("\n\t*********** State-store bucket *************\n\n")

	fmt.Printf(fmt.Sprintf("NODEUP_BUCKET = %s\n", os.Getenv("NODEUP_BUCKET")))
	fmt.Printf(fmt.Sprintf("UPLOAD_DEST = %s\n", os.Getenv("UPLOAD_DEST")))
	fmt.Printf(fmt.Sprintf("KOPS_BASE_URL = %s\n", os.Getenv("KOPS_BASE_URL")))
	fmt.Printf(fmt.Sprintf("KOPSCONTROLLER_IMAGE = %s\n", os.Getenv("KOPSCONTROLLER_IMAGE")))
	fmt.Printf(fmt.Sprintf("DNSCONTROLLER_IMAGE = %s\n", os.Getenv("DNSCONTROLLER_IMAGE")))

	fmt.Printf("\n********************* Registry access *********************\n\n")

	fmt.Printf(fmt.Sprintf("DOCKER_REGISTRY = %s\n", os.Getenv("DOCKER_REGISTRY")))
	fmt.Printf(fmt.Sprintf("DOCKER_IMAGE_PREFIX = %s\n", os.Getenv("DOCKER_IMAGE_PREFIX")))

	fmt.Printf("\n********************* Other *********************\n\n")

	fmt.Printf(fmt.Sprintf("KOPS_FEATURE_FLAGS = %s\n", os.Getenv("KOPS_FEATURE_FLAGS")))
	fmt.Printf(fmt.Sprintf("KOPS_ARCH = %s\n", os.Getenv("KOPS_ARCH")))
	fmt.Printf(fmt.Sprintf("KOPS_VERSION = %s\n\n", os.Getenv("KOPS_VERSION")))
}

func ParseZoneFromClusterSpec(clusterSpec kops.ClusterSpec) (scw.Zone, error) {
	zone := ""
	for _, subnet := range clusterSpec.Networking.Subnets {
		if zone == "" {
			zone = subnet.Zone
		} else if zone != subnet.Zone {
			return "", fmt.Errorf("scaleway currently only supports clusters in the same zone")
		}
	}
	return scw.Zone(zone), nil
}

func ParseRegionFromZone(zone scw.Zone) (region scw.Region, err error) {
	region, err = scw.ParseRegion(strings.TrimRight(string(zone), "-123"))
	if err != nil {
		return "", fmt.Errorf("could not determine region from zone %s: %w", zone, err)
	}
	return region, nil
}
