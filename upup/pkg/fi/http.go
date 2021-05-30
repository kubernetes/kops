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
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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

	dirMode := os.FileMode(0755)
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

func downloadURLAlways(url string, destPath string, dirMode os.FileMode) error {
	err := os.MkdirAll(path.Dir(destPath), dirMode)
	if err != nil {
		return fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	output, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file for download %q: %v", destPath, err)
	}
	defer output.Close()

	klog.Infof("Downloading %q", url)

	// Create a client with custom timeouts
	// to avoid idle downloads to hang the program
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       30 * time.Second,
		},
	}

	// this will stop slow downloads after 3 minutes
	// and interrupt reading of the Response.Body
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Cannot create request: %v", err)
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error doing HTTP fetch of %q: %v", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("error response from %q: HTTP %v", url, response.StatusCode)
	}

	start := time.Now()
	defer klog.Infof("Copying %q to %q took %q seconds", url, destPath, time.Since(start))

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return fmt.Errorf("error downloading HTTP content from %q: %v", url, err)
	}
	return nil
}
