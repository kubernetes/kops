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

package fi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/hashing"
)

// DownloadURL will download the file at the given url and store it as dest.
// If hash is non-nil, it will also verify that it matches the hash of the downloaded file.
func DownloadURL(url string, dest string, hash *hashing.Hash) (*hashing.Hash, error) {
	if hash != nil {
		match, err := fileHasHash(dest, hash)
		if err != nil {
			return nil, err
		}
		if match {
			return hash, nil
		}
	}

	dirMode := os.FileMode(0o755)
	err := downloadURLAlways(url, dest, dirMode)
	if err != nil {
		return nil, err
	}

	if hash != nil {
		match, err := fileHasHash(dest, hash)
		if err != nil {
			return nil, err
		}
		if !match {
			return nil, fmt.Errorf("downloaded from %q but hash did not match expected %q", url, hash)
		}
	} else {
		hash, err = hashing.HashAlgorithmSHA256.HashFile(dest)
		if err != nil {
			return nil, err
		}
	}

	return hash, nil
}

func addAuthHeadersFromEnvironment(client *http.Client, req *http.Request) error {
	// We need to figure out if we're running in a GCP VM, making a request to
	// google cloud storage.  If so, we need
	// to get an auth token from the metadata server.
	if req.URL.Host == "storage.googleapis.com" {
		cmd := exec.Command("systemctl", "check", "oem-gce")
		out, err := cmd.CombinedOutput()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() == 4 {
					klog.Infof("systemctl did not find oem-gce, assuming we are not on GCE: %s", out)
				}
				klog.Warningf("systemctl error: %v; %s", exitErr, out)
				// just return because it is perfectly fine to not be on GCE!
				return nil
			} else {
				klog.Errorf("failed to run systemctl: %w", err)
				return err
			}
		}
		// The "metadata.google.internal" URL is reachable from a GCE VM, but not from anywhere else.
		// If we fail to reach it, we should assume that we are, in fact, not on GCE, despite previous
		// indications to the contrary.
		metadataAuthReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
		if err != nil {
			klog.Warningf("Failed to construct request to metadata.google.internal - are we not on GCE?: %v", err)
			return nil
		}
		metadataAuthReq.Header.Add("Metadata-Flavor", "Google")
		metadataAuthResp, err := client.Do(metadataAuthReq)
		if err != nil {
			klog.Warningf("Failed to send request to metadata.google.internal - are we not on GCE?: %v", err)
			return nil
		}
		authFromMetadata, err := ioutil.ReadAll(metadataAuthResp.Body)
		if err != nil {
			return fmt.Errorf("failed to read auth metadata from gce: %w", err)
		}

		if metadataAuthResp.StatusCode != 200 {
			klog.Warning("Got non-200 status code %v from metadata server with body: %s", metadataAuthResp.StatusCode, authFromMetadata)
			return fmt.Errorf("non-200 status code while reading auth metadata from gce: %v", authFromMetadata)
		}
		var auth map[string]interface{}
		if err := json.Unmarshal(authFromMetadata, &auth); err != nil {
			klog.Errorf("Received response from metadata.google.internal, but it was not parsable. Assuming we are not on GCE: %v", err)
			return fmt.Errorf("unparsable response while reading auth metadata from gce: %w", err)
		}
		klog.Infof("Found auth token: %v", auth)

		req.Header.Add("Authorization", "Bearer "+auth["access_token"].(string))
	}

	return nil
}

func downloadURLAlways(url string, destPath string, dirMode os.FileMode) error {
	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return fmt.Errorf("error creating directories for destination file %q: %w", destPath, err)
	}

	output, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file for download %q: %w", destPath, err)
	}
	defer output.Close()

	klog.Infof("Downloading %q", url)

	// Create a client with custom timeouts
	// to avoid idle downloads to hang the program
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       30 * time.Second,
		},
	}

	// this will stop slow downloads after 3 minutes
	// and interrupt reading of the Response.Body
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Cannot create request: %w", err)
	}
	responseWithoutAuth, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error doing HTTP fetch of %q: %w", url, err)
	}
	defer responseWithoutAuth.Body.Close()
	response := responseWithoutAuth

	if response.StatusCode == 401 || response.StatusCode == 403 {
		klog.Infof("Detected that authentication is required.  Attempting to find auth token.")
		// Create the same request again, but this time add auth headers.
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("Cannot create request: %w", err)
		}
		if err := addAuthHeadersFromEnvironment(httpClient, req); err != nil {
			return fmt.Errorf("Cannot determine authentication headers: %w", err)
		}
		responseWithAuth, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error doing authenticated HTTP fetch of %q: %w", url, err)
		}
		response = responseWithAuth
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("error response from %q: HTTP %v", url, response.StatusCode)
	}

	defer response.Body.Close()
	start := time.Now()
	defer klog.Infof("Copying %q to %q took %q seconds", url, destPath, time.Since(start))

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return fmt.Errorf("error downloading HTTP content from %q: %w", url, err)
	}
	return nil
}
