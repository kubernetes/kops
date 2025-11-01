/*
Copyright 2025 The Kubernetes Authors.

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

package deployer

import (
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/boskos"
)

func (d *deployer) acquireBoskosAWSAccount() error {
	klog.V(1).Info("Acquiring AWS account from Boskos")

	boskosClient, err := boskos.NewClient(d.BoskosLocation)
	if err != nil {
		return fmt.Errorf("failed to make boskos client: %s", err)
	}
	d.boskos = boskosClient

	resource, err := boskos.Acquire(
		d.boskos,
		d.BoskosAWSResourceType,
		d.BoskosAcquireTimeout,
		d.BoskosHeartbeatInterval,
		d.boskosHeartbeatClose,
	)
	if err != nil {
		return fmt.Errorf("init failed to get aws account from boskos: %s", err)
	}
	d.boskosAWSAccount = resource.Name
	klog.V(1).Infof("Got aws account %s from boskos", d.boskosAWSAccount)

	if resource.UserData == nil {
		return fmt.Errorf("boskos resource %s has nil UserData", resource.Name)
	}

	jsonBytes, err := json.Marshal(resource.UserData)
	if err != nil {
		// Use %w to wrap the error for better debugging
		return fmt.Errorf("failed to marshal boskos user data: %w", err)
	}

	var data struct {
		AccessKeyID     string `json:"access_key_id"`
		SecretAccessKey string `json:"secret_access_key"`
		SessionToken    string `json:"session_token"`
	}

	// 2. Unmarshal the JSON bytes (from the map) into the 'data' struct
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return fmt.Errorf("failed to unmarshal boskos user data: %w", err)
	}

	os.Setenv("AWS_ACCESS_KEY_ID", data.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", data.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", data.SessionToken)

	return nil
}

func (d *deployer) acquireBoskosGCPProject() error {
	klog.V(1).Info("No GCP project provided, acquiring from Boskos")

	boskosClient, err := boskos.NewClient(d.BoskosLocation)
	if err != nil {
		return fmt.Errorf("failed to make boskos client: %s", err)
	}
	d.boskos = boskosClient

	resource, err := boskos.Acquire(
		d.boskos,
		d.BoskosGCPResourceType,
		d.BoskosAcquireTimeout,
		d.BoskosHeartbeatInterval,
		d.boskosHeartbeatClose,
	)
	if err != nil {
		return fmt.Errorf("init failed to get project from boskos: %s", err)
	}
	d.GCPProject = resource.Name
	klog.V(1).Infof("Got project %s from boskos", d.GCPProject)
	return nil
}

func (d *deployer) releaseBoskosResources() error {
	if d.boskos == nil {
		return nil
	}
	klog.V(2).Info("releasing boskos project")
	resources := []string{}
	if d.GCPProject != "" {
		resources = append(resources, d.GCPProject)
	}
	if d.boskosAWSAccount != "" {
		resources = append(resources, d.boskosAWSAccount)
	}
	err := boskos.Release(
		d.boskos,
		resources,
		d.boskosHeartbeatClose,
	)
	if err != nil {
		return fmt.Errorf("down failed to release boskos project: %s", err)
	}
	return nil
}
