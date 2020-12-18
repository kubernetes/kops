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

package gce

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/idtoken"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
)

type VerifierOptions struct {
	// Audience is the expected audience
	Audience string `json:"audience,omitempty"`

	// ProjectID is the project of the cluster (we verify the nodes are in this project)
	ProjectID string `json:"projectID,omitempty"`

	// ServiceAccounts are the emails of the service accounts that worker nodes are permitted to have.
	//	ServiceAccounts []string `json:"serviceAccounts,omitempty"`
}

type verifier struct {
	opt VerifierOptions

	nodeIdentifier *nodeIdentifier

	validator *idtoken.Validator
}

var _ fi.Verifier = &verifier{}

func NewVerifier(opt *VerifierOptions) (fi.Verifier, error) {
	ctx := context.Background()

	v := &verifier{
		opt: *opt,
	}

	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building token validator: %w", err)
	}
	v.validator = validator

	computeService, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}

	v.nodeIdentifier = &nodeIdentifier{
		computeService: computeService,
		project:        opt.ProjectID,
	}

	return v, nil

}

func (v *verifier) VerifyToken(ctx context.Context, token string, body []byte) (*fi.VerifyResult, error) {
	// Verify the token has signed the body content.
	sha := sha256.Sum256(body)
	audience := v.opt.Audience + "//" + base64.URLEncoding.EncodeToString(sha[:])

	payload, err := v.validator.Validate(ctx, token, audience)
	if err != nil {
		return nil, fmt.Errorf("token did not validate: %w", err)
	}

	klog.V(2).Infof("claims: %+v", payload.Claims)

	klog.Infof("TODO: Implement serviceAccountEmail validation")
	/*
		serviceAccountEmail, err := readClaim(payload.Claims, "email")
			if err != nil {
				return nil, err
			}

				foundServiceAccount := false
				for _, s := range a.opt.ServiceAccounts {
					if serviceAccountEmail == s {
						found = true
						break
					}
				}
				if !foundServiceAccouunt {
					return nil, fmt.Errorf("serviceAccount %q is not in allow-list of service accounts", serviceAccountEmail)
				}
	*/

	instanceName, err := readClaim(payload.Claims, "google", "compute_engine", "instance_name")
	if err != nil {
		return nil, err
	}
	zone, err := readClaim(payload.Claims, "google", "compute_engine", "zone")
	if err != nil {
		return nil, err
	}
	projectID, err := readClaim(payload.Claims, "google", "compute_engine", "project_id")
	if err != nil {
		return nil, err
	}

	id, err := v.nodeIdentifier.verifyInstance(ctx, projectID, zone, instanceName)
	if err != nil {
		return nil, err
	}

	result := &fi.VerifyResult{
		NodeName:          instanceName,
		InstanceGroupName: id.InstanceGroup,
	}

	return result, nil
}

// readClaim parses
func readClaim(claims map[string]interface{}, path ...string) (string, error) {
	current := claims
	for i, k := range path {
		v, ok := current[k]
		if !ok {
			return "", fmt.Errorf("%q claim not found", k)
		}
		if i+1 == len(path) {
			s, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("%q claim was of unexpected type %T", k, v)
			}
			return s, nil
		} else {
			m, ok := v.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("claim %q was of unexpected type %T", k, v)
			}
			current = m
		}
	}
	return "", fmt.Errorf("path is required")
}
