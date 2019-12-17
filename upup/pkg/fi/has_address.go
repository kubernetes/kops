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

// HasAddress is implemented by elastic/floating IP addresses, to expose the address
// For example, this is used so that the master SSL certificate can be configured with the dynamically allocated IP
type HasAddress interface {
	// FindIPAddress returns the address associated with the implementor.  If there is no address, returns (nil, nil)
	FindIPAddress(context *Context) (*string, error)
}
