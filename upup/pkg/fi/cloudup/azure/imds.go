/*
Copyright 2026 The Kubernetes Authors.

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

package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"k8s.io/klog/v2"
)

const (
	// imdsBaseURL is the base URL for the Azure Instance Metadata Service.
	imdsBaseURL = "http://169.254.169.254"

	// imdsAPIVersion is the IMDS API version.
	// https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service#versions
	imdsAPIVersion = "2025-04-07"
)

// imdsHTTPClient is shared by all IMDS queries. The explicit Transport{Proxy: nil}
// bypasses any system proxy since IMDS lives at the link-local 169.254.169.254
// address and must not be routed through one.
var imdsHTTPClient = &http.Client{
	Transport: &http.Transport{Proxy: nil},
	Timeout:   10 * time.Second,
}

// InstanceMetadata contains compute instance metadata from the Azure IMDS.
type InstanceMetadata struct {
	SubscriptionID    string `json:"subscriptionId"`
	ResourceGroupName string `json:"resourceGroupName"`
	ResourceID        string `json:"resourceId"`
	VMID              string `json:"vmId"`
}

// attestedDocument is the JSON response from the IMDS attested/document endpoint.
type attestedDocument struct {
	Encoding  string `json:"encoding"`
	Signature string `json:"signature"`
}

// queryIMDS queries an Azure IMDS endpoint and unmarshals the JSON response.
// https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service
func queryIMDS(path string, params url.Values, result any) error {
	if path == "" {
		return fmt.Errorf("IMDS path is required")
	}
	if result == nil {
		return fmt.Errorf("result is required")
	}

	req, err := http.NewRequest("GET", imdsBaseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating IMDS request: %w", err)
	}
	req.Header.Add("Metadata", "True")

	params.Set("api-version", imdsAPIVersion)
	req.URL.RawQuery = params.Encode()

	klog.V(4).Infof("Azure IMDS query: %q", req.URL.String())

	resp, err := imdsHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("querying IMDS %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("querying IMDS %s: status %d", path, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading IMDS response: %w", err)
	}
	klog.V(4).Infof("Azure IMDS response: %d bytes", len(body))

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("unmarshalling IMDS response: %w", err)
	}

	return nil
}

// QueryComputeInstanceMetadata queries Azure IMDS for compute instance metadata.
// https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service#instance-metadata
func QueryComputeInstanceMetadata() (*InstanceMetadata, error) {
	metadata := &InstanceMetadata{}
	params := url.Values{"format": {"json"}}
	if err := queryIMDS("/metadata/instance/compute", params, metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// queryIMDSAttestedDocument queries the Azure IMDS attested document endpoint.
// The nonce is included in the PKCS7 signed content for replay protection.
// https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service#attested-data
func queryIMDSAttestedDocument(nonce string) (*attestedDocument, error) {
	if nonce == "" {
		return nil, fmt.Errorf("nonce is required")
	}

	doc := &attestedDocument{}
	params := url.Values{"nonce": {nonce}}
	if err := queryIMDS("/metadata/attested/document", params, doc); err != nil {
		return nil, err
	}
	return doc, nil
}
