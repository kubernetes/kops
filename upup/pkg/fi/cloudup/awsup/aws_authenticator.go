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

package awsup

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"k8s.io/kops/upup/pkg/fi"
)

const AWSAuthenticationTokenPrefix = "x-aws-sts "

type awsAuthenticator struct {
	sts *sts.STS
}

var _ fi.Authenticator = &awsAuthenticator{}

func NewAWSAuthenticator() (fi.Authenticator, error) {
	config := aws.NewConfig().WithCredentialsChainVerboseErrors(true)
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return &awsAuthenticator{
		sts: sts.New(sess),
	}, nil
}

func (a awsAuthenticator) CreateToken(body []byte) (string, error) {
	sha := sha256.Sum256(body)

	stsRequest, _ := a.sts.GetCallerIdentityRequest(nil)

	// Ensure the signature is only valid for this particular body content.
	stsRequest.HTTPRequest.Header.Add("X-Kops-Request-SHA", base64.RawStdEncoding.EncodeToString(sha[:]))

	err := stsRequest.Sign()
	if err != nil {
		return "", err
	}

	headers, _ := json.Marshal(stsRequest.HTTPRequest.Header)
	return AWSAuthenticationTokenPrefix + base64.StdEncoding.EncodeToString(headers), nil
}
