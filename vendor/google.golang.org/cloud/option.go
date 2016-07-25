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

package cloud

import (
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// ClientOption is used when construct clients for each cloud service.
type ClientOption interface {
	// Resolve returns the equivalent option from the
	// google.golang.org/api/option package.
	Resolve() option.ClientOption
}

type wrapOpt struct {
	o option.ClientOption
}

func (w wrapOpt) Resolve() option.ClientOption {
	return w.o
}

// WithTokenSource returns a ClientOption that specifies an OAuth2 token
// source to be used as the basis for authentication.
func WithTokenSource(s oauth2.TokenSource) ClientOption {
	return wrapOpt{option.WithTokenSource(s)}
}

// WithEndpoint returns a ClientOption that overrides the default endpoint
// to be used for a service.
func WithEndpoint(url string) ClientOption {
	return wrapOpt{option.WithEndpoint(url)}
}

// WithScopes returns a ClientOption that overrides the default OAuth2 scopes
// to be used for a service.
func WithScopes(scope ...string) ClientOption {
	return wrapOpt{option.WithScopes(scope...)}
}

// WithUserAgent returns a ClientOption that sets the User-Agent.
func WithUserAgent(ua string) ClientOption {
	return wrapOpt{option.WithUserAgent(ua)}
}

// WithBaseHTTP returns a ClientOption that specifies the HTTP client to
// use as the basis of communications. This option may only be used with
// services that support HTTP as their communication transport.
func WithBaseHTTP(client *http.Client) ClientOption {
	return wrapOpt{option.WithHTTPClient(client)}
}

// WithBaseGRPC returns a ClientOption that specifies the gRPC client
// connection to use as the basis of communications. This option many only be
// used with services that support gRPC as their communication transport.
func WithBaseGRPC(conn *grpc.ClientConn) ClientOption {
	return wrapOpt{option.WithGRPCConn(conn)}
}

// WithGRPCDialOption returns a ClientOption that appends a new grpc.DialOption
// to an underlying gRPC dial. It does not work with WithBaseGRPC.
func WithGRPCDialOption(o grpc.DialOption) ClientOption {
	return wrapOpt{option.WithGRPCDialOption(o)}
}
