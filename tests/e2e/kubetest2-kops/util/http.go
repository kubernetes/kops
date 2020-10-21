/*
Copyright 2020 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"io"
	"net/http"

	"k8s.io/klog/v2"
)

var httpTransport *http.Transport

func init() {
	httpTransport = new(http.Transport)
	httpTransport.Proxy = http.ProxyFromEnvironment
	httpTransport.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
}

// httpGETWithHeaders writes the response of an HTTP GET request
func httpGETWithHeaders(url string, headers map[string]string, writer io.Writer) error {
	klog.Infof("curl %s", url)
	c := &http.Client{Transport: httpTransport}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	r, err := c.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode >= 400 {
		return fmt.Errorf("%v returned %d", url, r.StatusCode)
	}
	_, err = io.Copy(writer, r.Body)
	if err != nil {
		return err
	}
	return nil
}
