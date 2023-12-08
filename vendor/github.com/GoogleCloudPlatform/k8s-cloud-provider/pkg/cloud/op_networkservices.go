/*
Copyright 2023 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/networkservices/v1"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
)

type networkServicesOperation struct {
	s         *Service
	projectID string
	key       *meta.Key
	err       error
}

func (o *networkServicesOperation) String() string {
	return fmt.Sprintf("networkServicesOperation{%q, %s}", o.projectID, o.key)
}

func (o *networkServicesOperation) isDone(ctx context.Context) (bool, error) {
	var (
		op  *networkservices.Operation
		err error
	)

	fqname := fmt.Sprintf("projects/%s/locations/global/operations/%s", o.projectID, o.key.Name)
	klog.V(5).Infof("isDone %q", fqname)

	switch o.key.Type() {
	case meta.Global:
		op, err = o.s.NetworkServicesGA.Operations.Get(fqname).Context(ctx).Do()
		klog.V(5).Infof("GA.GlobalOperations.Get(%v, %v) = %+v, %v; ctx = %v", o.projectID, o.key.Name, op, err, ctx)
	default:
		return false, fmt.Errorf("invalid key type: %#v", o.key)
	}

	if err != nil {
		return false, err
	}

	if op == nil || !op.Done {
		return false, nil
	}

	if op.Error != nil {
		o.err = &googleapi.Error{
			Code:    int(op.Error.Code),
			Message: fmt.Sprintf("%v - %v", op.Error.Code, op.Error.Message),
		}
	}
	return true, nil
}

func (o *networkServicesOperation) rateLimitKey() *RateLimitKey {
	return &RateLimitKey{
		ProjectID: o.projectID,
		Operation: "Get",
		Service:   "Operations",
		Version:   meta.VersionGA,
	}
}

func (o *networkServicesOperation) error() error {
	return o.err
}

type networkServiceOpURLParseResult struct {
	projectID string
	key       *meta.Key
}

// parseNetworkServiceOpURL parses the URL of the network services operation.
// This is different than the `compute` API paths.
func parseNetworkServiceOpURL(name string) (*networkServiceOpURLParseResult, error) {
	// Format: projects/<projectID>/locations/global/operations/<Name>
	//         0        1           2         3      4          5
	split := strings.Split(name, "/")
	const pieces = 6
	if len(split) != pieces {
		return nil, fmt.Errorf("invalid op URL %q, want %d pieces, got %d", name, 6, len(split))
	}
	if split[0] != "projects" || split[2] != "locations" || split[4] != "operations" {
		return nil, fmt.Errorf("invalid op URL %q, did not match expected format", name)
	}
	if split[3] != "global" {
		return nil, fmt.Errorf("only global ops are supported (URL was %q)", name)
	}
	return &networkServiceOpURLParseResult{
		projectID: split[1],
		key:       meta.GlobalKey(split[5]),
	}, nil
}
