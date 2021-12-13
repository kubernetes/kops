/*
Copyright 2021 The Kubernetes Authors.

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

package gcphttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// errorResponseData is the response body for an error response.
type errorResponseData struct {
	Error errorInfo `json:"error"`
}

type errorInfo struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

// ErrorNotFound builds a NOT_FOUND error response
func ErrorNotFound(message string, args ...interface{}) (*http.Response, error) {
	statusCode := http.StatusNotFound

	e := errorResponseData{
		Error: errorInfo{
			Code:    statusCode,
			Message: fmt.Sprintf(message, args...),
			Status:  "NOT_FOUND",
		},
	}

	return buildJSONResponse(statusCode, e)
}

// ErrorBadRequest builds a BAD_REQUEST error response
func ErrorBadRequest(message string, args ...interface{}) (*http.Response, error) {
	statusCode := http.StatusBadRequest

	// TODO: What does this actually look like?

	e := errorResponseData{
		Error: errorInfo{
			Code:    statusCode,
			Message: fmt.Sprintf(message, args...),
			Status:  "BAD_REQUEST",
		},
	}

	return buildJSONResponse(statusCode, e)
}

// ErrorAlreadyExists builds an ALREADY_EXISTS error response
func ErrorAlreadyExists(message string, args ...interface{}) (*http.Response, error) {
	statusCode := http.StatusConflict

	e := errorResponseData{
		Error: errorInfo{
			Code:    statusCode,
			Message: fmt.Sprintf(message, args...),
			Status:  "ALREADY_EXISTS",
		},
	}

	return buildJSONResponse(statusCode, e)
}

// OKResponse builds an response encoding the provided data.
func OKResponse(obj interface{}) (*http.Response, error) {
	return buildJSONResponse(http.StatusOK, obj)
}

func buildJSONResponse(statusCode int, obj interface{}) (*http.Response, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSON: %w", err)
	}

	r := &http.Response{
		Status:     http.StatusText(statusCode),
		StatusCode: statusCode,
	}

	r.Header = make(http.Header)
	r.Header.Add("Content-Type", "application/json; charset=UTF-8")

	r.Body = ioutil.NopCloser(bytes.NewReader(b))

	return r, nil
}
