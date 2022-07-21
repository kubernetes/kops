package scaleway

import (
	"errors"
	"fmt"
	"k8s.io/kops/pkg/apis/kops"
	"net/http"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultInstanceWaitRetryInterval = 5 * time.Second
	defaultInstanceWaitTimeout       = 10 * time.Minute
)

func FindRegion(cluster *kops.Cluster) (string, error) {
	region := ""

	for _, subnet := range cluster.Spec.Subnets {
		zoneRegion := ""

		switch subnet.Zone {
		case "fr-par-1", "fr-par-2", "fr-par-3":
			zoneRegion = "fr-par"
		case "nl-ams-1", "nl-ams-2":
			zoneRegion = "nl-ams"
		case "pl-waw-1":
			zoneRegion = "pl-waw"
		default:
			return "", fmt.Errorf("unknown zone: %s", subnet.Zone)
		}

		if region != "" && region != zoneRegion {
			return "", fmt.Errorf("cluster cannot span multiple regions (found zone %s, but region is %s)", subnet.Zone, region)
		}
		region = subnet.Region
	}
	return region, nil
}

func reachState(instanceAPI *instance.API, zone scw.Zone, serverID string, toState instance.ServerState) error {
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

func waitForInstanceServer(api *instance.API, zone scw.Zone, id string) (*instance.Server, error) {
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
