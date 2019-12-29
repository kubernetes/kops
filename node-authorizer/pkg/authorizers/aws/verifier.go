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

package aws

import (
	"context"
	"encoding/json"

	"k8s.io/kops/node-authorizer/pkg/server"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

type awsNodeVerifier struct{}

// NewVerifier creates and returns a verifier
func NewVerifier() (server.Verifier, error) {
	return &awsNodeVerifier{}, nil
}

// Verify is responsible for build a identification document
func (a *awsNodeVerifier) VerifyIdentity(ctx context.Context) ([]byte, error) {
	errs := make(chan error)
	doneCh := make(chan []byte)

	go func() {
		encoded, err := func() ([]byte, error) {
			// @step: create a metadata client
			sess, err := session.NewSession()
			if err != nil {
				return []byte{}, err
			}
			client := ec2metadata.New(sess)

			// @step: get the pkcs7 signature from the metadata service
			signature, err := client.GetDynamicData("/instance-identity/pkcs7")
			if err != nil {
				return []byte{}, err
			}

			// @step: construct request for the request
			request := &Request{
				Document: []byte(signature),
			}

			return json.Marshal(request)
		}()
		if err != nil {
			errs <- err
			return
		}

		doneCh <- encoded
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errs:
		return nil, err
	case req := <-doneCh:
		return req, nil
	}
}
