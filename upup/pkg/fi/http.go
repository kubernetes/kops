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
	"path/filepath"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/hashing"
)

const downloadTimeout = 3 * time.Minute

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
	err := downloadURLAlways(url, dest, dirMode, hash)
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

func downloadURLAlways(url string, destPath string, dirMode os.FileMode, hash *hashing.Hash) error {
	dir := filepath.Dir(destPath)
	err := os.MkdirAll(dir, dirMode)
	if err != nil {
		return fmt.Errorf("error creating directories for destination file %q: %v", destPath, err)
	}

	output, err := os.CreateTemp(dir, "."+filepath.Base(destPath)+".tmp")
	if err != nil {
		return fmt.Errorf("error creating temporary file for download %q: %v", destPath, err)
	}
	tempPath := output.Name()
	defer os.Remove(tempPath)

	_, err = DownloadURLToWriter(url, output, hash)
	if closeErr := output.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err := os.Chmod(tempPath, 0o644); err != nil {
		return fmt.Errorf("error setting mode on downloaded file %q: %v", tempPath, err)
	}
	if err := os.Rename(tempPath, destPath); err != nil {
		return fmt.Errorf("error moving downloaded file %q to %q: %v", tempPath, destPath, err)
	}
	return nil
}

// DownloadURLToWriter streams the file at the given url to dest.
// If hash is non-nil, it will also verify that it matches the downloaded bytes.
func DownloadURLToWriter(url string, dest io.Writer, hash *hashing.Hash) (*hashing.Hash, error) {
	responseBody, err := OpenURL(url)
	if err != nil {
		return nil, err
	}
	defer responseBody.Close()

	klog.V(2).Infof("Downloading %q", url)

	start := time.Now()
	defer func() {
		klog.V(2).Infof("Downloading %q took %q", url, time.Since(start))
	}()

	algorithm := hashing.HashAlgorithmSHA256
	if hash != nil {
		algorithm = hash.Algorithm
	}
	hasher := algorithm.NewHasher()
	writer := io.MultiWriter(dest, hasher)

	if _, err := io.Copy(writer, responseBody); err != nil {
		return nil, fmt.Errorf("error downloading HTTP content from %q: %v", url, err)
	}

	actual := &hashing.Hash{
		Algorithm: algorithm,
		HashValue: hasher.Sum(nil),
	}
	if hash != nil && !actual.Equal(hash) {
		return nil, fmt.Errorf("downloaded from %q but hash did not match expected %q", url, hash)
	}
	return actual, nil
}

// OpenURL opens a hardened HTTP GET stream for url.
func OpenURL(url string) (io.ReadCloser, error) {
	httpClient := newDownloadHTTPClient()

	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("cannot create request: %v", err)
	}

	response, err := httpClient.Do(req)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("error doing HTTP fetch of %q: %v", url, err)
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		response.Body.Close()
		cancel()
		return nil, fmt.Errorf("unexpected response from %q: HTTP %s", url, response.Status)
	}

	return &cancelOnCloseReadCloser{ReadCloser: response.Body, cancel: cancel}, nil
}

func newDownloadHTTPClient() *http.Client {
	return &http.Client{
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
}

type cancelOnCloseReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (r *cancelOnCloseReadCloser) Close() error {
	err := r.ReadCloser.Close()
	r.cancel()
	return err
}
