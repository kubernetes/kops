// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transport

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"google.golang.org/cloud"
	"google.golang.org/grpc"
)

// ErrHTTP is returned when on a non-200 HTTP response.
type ErrHTTP struct {
	StatusCode int
	Body       []byte
	err        error
}

func (e *ErrHTTP) Error() string {
	if e.err == nil {
		return fmt.Sprintf("error during call, http status code: %v %s", e.StatusCode, e.Body)
	}
	return e.err.Error()
}

// NewHTTPClient returns an HTTP client for use communicating with a Google cloud
// service, configured with the given ClientOptions. It also returns the endpoint
// for the service as specified in the options.
func NewHTTPClient(ctx context.Context, opt ...cloud.ClientOption) (*http.Client, string, error) {
	o := make([]option.ClientOption, 0, len(opt))
	for _, opt := range opt {
		o = append(o, opt.Resolve())
	}
	return transport.NewHTTPClient(ctx, o...)
}

// DialGRPC returns a GRPC connection for use communicating with a Google cloud
// service, configured with the given ClientOptions.
func DialGRPC(ctx context.Context, opt ...cloud.ClientOption) (*grpc.ClientConn, error) {
	o := make([]option.ClientOption, 0, len(opt))
	for _, opt := range opt {
		o = append(o, opt.Resolve())
	}
	return transport.DialGRPC(ctx, o...)
}
