/*
Copyright 2019 The Kubernetes Authors.

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

// Extracted from the k8s GCE cloud provider
// The file contains functions that deal with waiting for GCE operations to complete

import (
	"fmt"
	"time"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

const (
	operationPollInterval        = 3 * time.Second
	operationPollTimeoutDuration = 30 * time.Minute
)

func WaitForOp(client *compute.Service, op *compute.Operation) error {
	u, err := ParseGoogleCloudURL(op.SelfLink)
	if err != nil {
		return fmt.Errorf("error parsing operation URL %q: %v", op.SelfLink, err)
	}

	if u.Zone != "" {
		return waitForZoneOp(client, op)
	}

	if u.Region != "" {
		return waitForRegionOp(client, op)
	}

	return waitForGlobalOp(client, op)
}

func waitForZoneOp(client *compute.Service, op *compute.Operation) error {
	u, err := ParseGoogleCloudURL(op.SelfLink)
	if err != nil {
		return fmt.Errorf("error parsing operation URL %q: %v", op.SelfLink, err)
	}

	return waitForOp(op, func(operationName string) (*compute.Operation, error) {
		return client.ZoneOperations.Get(u.Project, u.Zone, operationName).Do()
	})
}

func waitForRegionOp(client *compute.Service, op *compute.Operation) error {
	u, err := ParseGoogleCloudURL(op.SelfLink)
	if err != nil {
		return fmt.Errorf("error parsing operation URL %q: %v", op.SelfLink, err)
	}

	return waitForOp(op, func(operationName string) (*compute.Operation, error) {
		return client.RegionOperations.Get(u.Project, u.Region, operationName).Do()
	})
}

func waitForGlobalOp(client *compute.Service, op *compute.Operation) error {
	u, err := ParseGoogleCloudURL(op.SelfLink)
	if err != nil {
		return fmt.Errorf("error parsing operation URL %q: %v", op.SelfLink, err)
	}

	return waitForOp(op, func(operationName string) (*compute.Operation, error) {
		return client.GlobalOperations.Get(u.Project, operationName).Do()
	})
}

func opIsDone(op *compute.Operation) bool {
	return op != nil && op.Status == "DONE"
}

func waitForOp(op *compute.Operation, getOperation func(operationName string) (*compute.Operation, error)) error {
	if op == nil {
		return fmt.Errorf("operation must not be nil")
	}

	if opIsDone(op) {
		return getErrorFromOp(op)
	}

	opStart := time.Now()
	opName := op.Name
	return wait.Poll(operationPollInterval, operationPollTimeoutDuration, func() (bool, error) {
		start := time.Now()
		//gce.operationPollRateLimiter.Accept()
		duration := time.Since(start)
		if duration > 5*time.Second {
			klog.Infof("pollOperation: throttled %v for %v", duration, opName)
		}
		pollOp, err := getOperation(opName)
		if err != nil {
			klog.Warningf("GCE poll operation %s failed: pollOp: [%v] err: [%v] getErrorFromOp: [%v]", opName, pollOp, err, getErrorFromOp(pollOp))
		}
		done := opIsDone(pollOp)
		if done {
			duration := time.Since(opStart)
			if duration > 1*time.Minute {
				// Log the JSON. It's cleaner than the %v structure.
				enc, err := pollOp.MarshalJSON()
				if err != nil {
					klog.Warningf("waitForOperation: long operation (%v): %v (failed to encode to JSON: %v)", duration, pollOp, err)
				} else {
					klog.Infof("waitForOperation: long operation (%v): %v", duration, string(enc))
				}
			}
		}
		return done, getErrorFromOp(pollOp)
	})
}

func getErrorFromOp(op *compute.Operation) error {
	if op != nil && op.Error != nil && len(op.Error.Errors) > 0 {
		err := &googleapi.Error{
			Code:    int(op.HttpErrorStatusCode),
			Message: op.Error.Errors[0].Message,
		}
		klog.Errorf("GCE operation failed: %v", err)
		return err
	}

	return nil
}
