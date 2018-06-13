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

package aws

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"k8s.io/kops/node-authorizer/pkg/server"
	"k8s.io/kops/node-authorizer/pkg/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/fullsailor/pkcs7"
	"go.uber.org/zap"
)

const (
	// the default tag used to indicates a kubernetes cluster name
	kubernetesClusterTag = "KubernetesCluster"
)

// a collection of aws public signing certificates
var publicCertificates []*x509.Certificate

// awsNodeAuthorizer is the implementation for a node authorizer
type awsNodeAuthorizer struct {
	// client is the ec2 interface
	client ec2iface.EC2API
	// config is the service configuration
	config *server.Config
	// vpcID is our vpc id
	vpcID string
}

func init() {
	for i := range awsCertificates {
		block, _ := pem.Decode([]byte(awsCertificates[i]))

		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			panic("failed to parse the embedded aws certificate")
		}

		publicCertificates = append(publicCertificates, c)
	}
}

// NewAuthorizer creates and returns a aws node authorizer
func NewAuthorizer(config *server.Config) (server.Authorizer, error) {
	// @step: get the identity document for the instance we are running
	document, err := getInstanceIdentityDocument()
	if err != nil {
		return nil, err
	}

	utils.Logger.Info("running node authorizer on instance",
		zap.String("instance-id", document.InstanceID),
		zap.String("region", document.Region))

	// @step: we create a ec2 client
	client := ec2.New(session.New(&aws.Config{
		Region: aws.String(document.Region),
	}))

	// @step: get information on the instance we are running
	instance, err := getInstance(client, document.InstanceID)
	if err != nil {
		return nil, err
	}

	return &awsNodeAuthorizer{
		client: client,
		config: config,
		vpcID:  aws.StringValue(instance.VpcId),
	}, nil
}

// Authorize is responsible for accepting the request
func (a *awsNodeAuthorizer) Authorize(ctx context.Context, r *server.NodeRegistration) error {
	identity := &ec2metadata.EC2InstanceIdentityDocument{}

	// @step: decode the request
	request, err := decodeRequest(r.Spec.Request)
	if err != nil {
		return err
	}

	// @step: extract and validate the document
	if reason, err := func() (string, error) {
		if reason, err := a.validateIdentityDocument(ctx, request.Document, identity); err != nil {
			return "", err
		} else if reason != "" {
			return reason, nil
		}

		if reason, err := a.validateNodeInstance(ctx, identity, r); err != nil {
			return "", err
		} else if reason != "" {
			return reason, nil
		}

		r.Status.Allowed = true

		return "", nil
	}(); err != nil {
		return err
	} else if reason != "" {
		r.Deny(reason)
	}

	return nil
}

// validateNodeInstance is responsible for checking the instance exists and it part of the cluster
// - check instance exists
// - check the instance is running
// - check the instance is running in our vpc
// - check the instance run tagged with our kubernetes cluster
func (a *awsNodeAuthorizer) validateNodeInstance(ctx context.Context, doc *ec2metadata.EC2InstanceIdentityDocument, spec *server.NodeRegistration) (string, error) {
	// @check we found some instances
	instance, err := getInstance(a.client, doc.InstanceID)
	if err != nil {
		return "", err
	}
	if aws.StringValue(instance.State.Name) != ec2.InstanceStateNameRunning {
		return "instance is not running", nil
	}

	// @check the instance is running in our vpc
	if aws.StringValue(instance.VpcId) != a.vpcID {
		return "instance not running in vpc", nil
	}

	// @check the instance is tagged with our kubernetes cluster id
	if !hasInstanceTags(kubernetesClusterTag, a.config.ClusterName, instance.Tags) {
		return "missing cluster tag", nil
	}

	// @check the requester is as expected
	if a.config.EnableAddressCheck {
		if spec.Spec.RemoteAddr != aws.StringValue(instance.PrivateIpAddress) {
			return fmt.Sprintf("ip address conflict, expected: %s, got: %s",
				aws.StringValue(instance.PrivateIpAddress), spec.Spec.RemoteAddr), nil
		}
	}

	return "", nil
}

// validateIdentityDocument is responsible for validate the aws identity document
func (a *awsNodeAuthorizer) validateIdentityDocument(_ context.Context, signed []byte, document interface{}) (string, error) {
	// @step: decode the signed document
	decoded, err := base64.StdEncoding.DecodeString(string(signed))
	if err != nil {
		return "", err
	}

	// @step: get the digest
	for _, x := range publicCertificates {
		parsed, err := pkcs7.Parse(decoded)
		if err != nil {
			return "", err
		}

		parsed.Certificates = []*x509.Certificate{x}
		if err := parsed.Verify(); err == nil {
			return "", json.NewDecoder(bytes.NewReader(parsed.Content)).Decode(document)
		}
	}

	return "invalid signature", nil
}

// validateNodeRegistrationRequest is responsible for validating the request itself
func validateNodeRegistrationRequest(request *Request) error {
	err := func() error {
		if len(request.Document) <= 0 {
			return errors.New("missing identity document")
		}

		return nil
	}()
	if err != nil {
		return fmt.Errorf("invalid verification request: %s", err)
	}

	return nil
}

// decodeRequest is responsible for decoding the request
func decodeRequest(in []byte) (*Request, error) {
	request := &Request{}

	if err := json.NewDecoder(bytes.NewReader(in)).Decode(request); err != nil {
		return nil, err
	}

	// @step: validate the node request
	if err := validateNodeRegistrationRequest(request); err != nil {
		return nil, err
	}

	return request, nil
}

func (a *awsNodeAuthorizer) Close() error {
	return nil
}

// Name returns the name of the authozier
func (a *awsNodeAuthorizer) Name() string {
	return "aws"
}

// hasInstanceTags checks the tags exists on the cluster
func hasInstanceTags(name, value string, tags []*ec2.Tag) bool {
	for _, x := range tags {
		if aws.StringValue(x.Key) == name && aws.StringValue(x.Value) == value {
			return true
		}
	}

	return false
}

// getInstanceIdentityDocument is responsible for retrieving the instance identity document
func getInstanceIdentityDocument() (ec2metadata.EC2InstanceIdentityDocument, error) {
	var document ec2metadata.EC2InstanceIdentityDocument

	client := ec2metadata.New(session.New())
	maxInterval := 500 * time.Millisecond
	maxTime := 5 * time.Second

	err := utils.Retry(context.TODO(), maxInterval, maxTime, func() error {
		x, err := client.GetInstanceIdentityDocument()
		if err != nil {
			return err
		}
		document = x

		return nil
	})

	return document, err
}

// getInstance is responsible for getting the instance
func getInstance(client ec2iface.EC2API, instanceID string) (*ec2.Instance, error) {
	// @step: describe the instance
	resp, err := client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})
	if err != nil {
		return nil, err
	}

	// @check we found some instances
	if len(resp.Reservations) <= 0 || len(resp.Reservations[0].Instances) <= 0 {
		return nil, fmt.Errorf("missing instance id: %s", instanceID)
	}
	if len(resp.Reservations[0].Instances) > 1 {
		return nil, fmt.Errorf("found multiple instances with instance id: %s", instanceID)
	}

	// @check the instance is running
	instance := resp.Reservations[0].Instances[0]
	if instance.State == nil {
		return nil, errors.New("missing instance status")
	}

	return instance, nil
}
