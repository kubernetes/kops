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

package do

import (
	"fmt"
	"io"
	"net/http"

	"k8s.io/kops/pkg/bootstrap"
)

const DOAuthenticationTokenPrefix = "x-digitalocean-droplet-id "

type doAuthenticator struct {
}

var _ bootstrap.Authenticator = &doAuthenticator{}

func NewAuthenticator() (bootstrap.Authenticator, error) {
	return &doAuthenticator{}, nil
}

func (o *doAuthenticator) CreateToken(body []byte) (string, error) {
	dropletID, err := getMetadataDropletID()
	if err != nil {
		return "", fmt.Errorf("unable to fetch droplet id: %w", err)
	}
	return DOAuthenticationTokenPrefix + dropletID, nil
}

const (
	dropletIDMetadataURL = "http://169.254.169.254/metadata/v1/id"
)

func getMetadataDropletID() (string, error) {
	return getMetadata(dropletIDMetadataURL)
}

func getMetadata(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error querying droplet metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("droplet metadata returned non-200 status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading droplet metadata: %w", err)
	}

	return string(bodyBytes), nil
}
