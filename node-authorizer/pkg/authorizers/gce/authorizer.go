/*
Copyright 2018 The Kubernetes Authors.

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
	"encoding/json"
	"fmt"
	"strconv"

	"k8s.io/kops/node-authorizer/pkg/authorizers"
	"k8s.io/kops/node-authorizer/pkg/server"
	"k8s.io/kops/node-authorizer/pkg/utils"

	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v0.beta"

	"go.uber.org/zap"
)

var (
	// TODO: Promote - these aren't specific to aws/gce

	// CheckIAMProfile indicates we should validate the iam profile
	CheckIAMProfile = "verify-iam-profile"
	// CheckIPAddress indicates we should validate the client ip address
	CheckIPAddress = "verify-ip"
	// CheckSignature indicates we validate the signature of the document
	CheckSignature = "verify-signature"
)

// gceNodeAuthorizer is the implementation for a node authorizer
type gceNodeAuthorizer struct {
	// compute is a client for GCE compute services
	compute *compute.Service
	// config is the service configuration
	config *server.Config
	// identity is our local identity
	identity *gceIdentityData
	// validator performs JWT validattion
	validator *Validator
}

// NewAuthorizer creates and returns a gce node authorizer
func NewAuthorizer(config *server.Config) (server.Authorizer, error) {
	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, fmt.Errorf("error building google API client: %v", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}

	// @step: get the identity document for the instance we are running
	identityClaim, err := getLocalInstanceIdentityClaim(ctx, AudienceNodeBootstrap)
	if err != nil {
		return nil, err
	}

	validator, err := NewValidator(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build validator: %v", err)
	}

	identity, err := validator.ParseAndValidateClaim(ctx, identityClaim, AudienceNodeBootstrap)
	if err != nil {
		return nil, fmt.Errorf("failed to validate own identity: %v", err)
	}

	utils.Logger.Info("running node authorizer on instance",
		zap.String("instance-id", identity.Google.ComputeEngine.InstanceID),
		zap.String("zone", identity.Google.ComputeEngine.Zone))

	// @step: get information on the instance we are running
	/*
		instance, err := computeService.Instances.Get(identity.Google.ComputeEngine.ProjectID, identity.Google.ComputeEngine.Zone, identity.Google.ComputeEngine.InstanceName).Do()
			if err != nil {
				return nil, fmt.Errorf("error getting self instance: %v", err)
			}
	*/

	return &gceNodeAuthorizer{
		compute:   computeService,
		config:    config,
		validator: validator,
		identity:  identity,
		//instance: instance,
		//vpcID:    gce.StringValue(instance.VpcId),
	}, nil
}

// Authorize is responsible for accepting the request
func (a *gceNodeAuthorizer) Authorize(ctx context.Context, r *server.NodeRegistration) error {
	// @step: decode the request
	request, err := decodeRequest(r.Spec.Request)
	if err != nil {
		return err
	}

	if len(request.Document) == 0 {
		r.Deny("identity document not supplied")
		return nil
	}

	identity, err := a.validator.ParseAndValidateClaim(ctx, string(request.Document), AudienceNodeBootstrap)
	if err != nil {
		return err
	}

	if identity == nil {
		r.Deny("identity document not valid")
		return nil
	}

	if reason, err := a.validateNodeInstance(ctx, identity, r); err != nil {
		return err
	} else if reason != "" {
		r.Deny(reason)
		return nil
	}

	r.Status.Allowed = true
	return nil
}

// validateNodeInstance is responsible for checking the instance exists and it part of the cluster
func (a *gceNodeAuthorizer) validateNodeInstance(ctx context.Context, identity *gceIdentityData, spec *server.NodeRegistration) (string, error) {
	// @check we are in the same account
	if a.identity.Google.ComputeEngine.ProjectNumber != identity.Google.ComputeEngine.ProjectNumber {
		return "instance running in different project id", nil
	}

	// @check we found some instances
	instance, err := a.compute.Instances.Get(identity.Google.ComputeEngine.ProjectID, identity.Google.ComputeEngine.Zone, identity.Google.ComputeEngine.InstanceName).Do()
	if err != nil {
		return "instance not found", nil
	}

	if strconv.FormatUint(instance.Id, 10) != identity.Google.ComputeEngine.InstanceID {
		return "instance id mismatch", nil
	}

	if instance.Status != "RUNNING" {
		return "instance is not running", nil
	}

	// TODO: Do we want to do an equivalent of the VPC check, or is that the project check?

	// @check the instance is tagged with our kubernetes cluster id
	clusterTag := a.config.ClusterTag
	if clusterTag == "" {
		clusterTag = "cluster-name"
	}
	if !hasMetadata(clusterTag, a.config.ClusterName, instance.Metadata) {
		return "missing cluster tag", nil
	}

	// TODO: Do we want to do an equivalent of the IAM lookup
	// (It would be the ServiceAccount)
	/*
		// @check the instance has access to the nodes iam profile
		if a.config.UseFeature(CheckIAMProfile) {
			if instance.IamInstanceProfile == nil {
				return "instance does not have an instance profile", nil
			}
			if gce.StringValue(instance.IamInstanceProfile.Arn) == "" {
				return "instance profile arn is empty", nil
			}
			expectedArn := fmt.Sprintf("arn:gce:iam::%s:role/nodes.%s", a.identity.AccountID, a.config.ClusterName)
			if expectedArn != gce.StringValue(instance.IamInstanceProfile.Arn) {
				return fmt.Sprintf("invalid iam instance role, expected: %s, found: %s", expectedArn, gce.StringValue(instance.IamInstanceProfile.Arn)), nil
			}
		}
	*/

	// @check the requester is as expected
	if a.config.UseFeature(CheckIPAddress) {
		found := false
		for _, nic := range instance.NetworkInterfaces {
			if nic.NetworkIP == spec.Spec.RemoteAddr {
				found = true
			}
		}

		if !found {
			return "instance IP mismatch", nil
		}
	}

	return "", nil
}

// hasMetadata checks for a tag-like key/value pair in the metadata
func hasMetadata(key, value string, metadata *compute.Metadata) bool {
	if metadata == nil {
		return false
	}

	for _, mi := range metadata.Items {
		if mi.Key != key {
			continue
		}
		if mi.Value != nil && *mi.Value == value {
			return true
		}
	}

	return false
}

// validateNodeRegistrationRequest is responsible for validating the request itself
func validateNodeRegistrationRequest(request *authorizers.Request) error {
	return nil
}

// decodeRequest is responsible for decoding the request
func decodeRequest(in []byte) (*authorizers.Request, error) {
	request := &authorizers.Request{}
	if err := json.Unmarshal(in, &request); err != nil {
		return nil, fmt.Errorf("error deserializing request: %v", err)
	}

	// @step: validate the node request
	if err := validateNodeRegistrationRequest(request); err != nil {
		return nil, err
	}

	return request, nil
}

func (a *gceNodeAuthorizer) Close() error {
	return nil
}

// Name returns the name of the authozier
func (a *gceNodeAuthorizer) Name() string {
	return "gce"
}
