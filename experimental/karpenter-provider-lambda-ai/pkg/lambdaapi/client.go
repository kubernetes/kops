/*
Copyright The Kubernetes Authors.

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

package lambdaapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BaseURL = "https://cloud.lambda-labs.com/api/v1"

type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BaseURL,
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) ListInstanceTypes() (map[string]InstanceType, error) {
	var resp ListInstanceTypesResponse
	if err := c.doRequest(http.MethodGet, "/instance-types", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ListRegions() ([]Region, error) {
	var resp ListRegionsResponse
	if err := c.doRequest(http.MethodGet, "/regions", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) LaunchInstance(req LaunchRequest) ([]string, error) {
	var resp LaunchResponse
	if err := c.doRequest(http.MethodPost, "/instance-operations/launch", req, &resp); err != nil {
		return nil, err
	}
	return resp.Data.InstanceIDs, nil
}

func (c *Client) TerminateInstances(instanceIDs []string) error {
	req := TerminateRequest{InstanceIDs: instanceIDs}
	var resp TerminateResponse
	return c.doRequest(http.MethodPost, "/instance-operations/terminate", req, &resp)
}

func (c *Client) ListInstances() ([]Instance, error) {
	var resp ListInstancesResponse
	if err := c.doRequest(http.MethodGet, "/instances", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) ListSSHKeys() ([]SSHKey, error) {
	var resp ListSSHKeysResponse
	if err := c.doRequest(http.MethodGet, "/ssh-keys", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
