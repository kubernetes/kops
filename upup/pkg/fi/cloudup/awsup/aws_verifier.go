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
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
	"k8s.io/kops/upup/pkg/fi"
)

type AWSVerifierOptions struct {
	// NodesRoles are the IAM roles that worker nodes are permitted to have.
	NodesRoles []string `json:"nodesRoles"`
}

type awsVerifier struct {
	accountId string
	partition string
	opt       AWSVerifierOptions

	ec2    *ec2.EC2
	sts    *sts.STS
	client http.Client
}

var _ fi.Verifier = &awsVerifier{}

func NewAWSVerifier(opt *AWSVerifierOptions) (fi.Verifier, error) {
	config := aws.NewConfig().WithCredentialsChainVerboseErrors(true)
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	stsClient := sts.New(sess)
	identity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	partition := strings.Split(aws.StringValue(identity.Arn), ":")[1]

	metadata := ec2metadata.New(sess, config)
	region, err := metadata.Region()
	if err != nil {
		return nil, fmt.Errorf("error querying ec2 metadata service (for region): %v", err)
	}

	ec2Client := ec2.New(sess, config.WithRegion(region))

	return &awsVerifier{
		accountId: aws.StringValue(identity.Account),
		partition: partition,
		opt:       *opt,
		ec2:       ec2Client,
		sts:       stsClient,
		client: http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				DisableKeepAlives:     true,
				MaxIdleConnsPerHost:   -1,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

type GetCallerIdentityResponse struct {
	XMLName                 xml.Name                  `xml:"GetCallerIdentityResponse"`
	GetCallerIdentityResult []GetCallerIdentityResult `xml:"GetCallerIdentityResult"`
	ResponseMetadata        []ResponseMetadata        `xml:"ResponseMetadata"`
}

type GetCallerIdentityResult struct {
	Arn     string `xml:"Arn"`
	UserId  string `xml:"UserId"`
	Account string `xml:"Account"`
}

type ResponseMetadata struct {
	RequestId string `xml:"RequestId"`
}

func (a awsVerifier) VerifyToken(token string, body []byte) (*fi.VerifyResult, error) {
	if !strings.HasPrefix(token, AWSAuthenticationTokenPrefix) {
		return nil, fmt.Errorf("incorrect authorization type")
	}
	token = strings.TrimPrefix(token, AWSAuthenticationTokenPrefix)

	// We rely on the client and server using the same version of the same STS library.
	stsRequest, _ := a.sts.GetCallerIdentityRequest(nil)
	err := stsRequest.Sign()
	if err != nil {
		return nil, fmt.Errorf("creating identity request: %v", err)
	}

	stsRequest.HTTPRequest.Header = nil
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("decoding authorization token: %v", err)
	}
	err = json.Unmarshal(tokenBytes, &stsRequest.HTTPRequest.Header)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling authorization token: %v", err)
	}

	// Verify the token has signed the body content.
	sha := sha256.Sum256(body)
	if stsRequest.HTTPRequest.Header.Get("X-Kops-Request-SHA") != base64.RawStdEncoding.EncodeToString(sha[:]) {
		return nil, fmt.Errorf("incorrect SHA")
	}

	requestBytes, _ := ioutil.ReadAll(stsRequest.Body)
	_, _ = stsRequest.Body.Seek(0, io.SeekStart)
	if stsRequest.HTTPRequest.Header.Get("Content-Length") != strconv.Itoa(len(requestBytes)) {
		return nil, fmt.Errorf("incorrect content-length")
	}

	response, err := a.client.Do(stsRequest.HTTPRequest)
	if err != nil {
		return nil, fmt.Errorf("sending STS request: %v", err)
	}
	if response != nil {
		defer response.Body.Close()
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading STS response: %v", err)
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("received status code %d from STS: %s", response.StatusCode, string(responseBody))
	}

	result := GetCallerIdentityResponse{}
	err = xml.NewDecoder(bytes.NewReader(responseBody)).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("decoding STS response: %v", err)
	}

	if result.GetCallerIdentityResult[0].Account != a.accountId {
		return nil, fmt.Errorf("incorrect account %s", result.GetCallerIdentityResult[0].Account)
	}

	arn := result.GetCallerIdentityResult[0].Arn
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return nil, fmt.Errorf("arn %q contains unexpected number of colons", arn)
	}
	if parts[0] != "arn" {
		return nil, fmt.Errorf("arn %q doesn't start with \"arn:\"", arn)
	}
	if parts[1] != a.partition {
		return nil, fmt.Errorf("arn %q not in partion %q", arn, a.partition)
	}
	if parts[2] != "iam" && parts[2] != "sts" {
		return nil, fmt.Errorf("arn %q has unrecognized service", arn)
	}
	// parts[3] is region
	// parts[4] is account
	resource := strings.Split(parts[5], "/")
	if resource[0] != "assumed-role" {
		return nil, fmt.Errorf("arn %q has unrecognized type", arn)
	}
	if len(resource) < 3 {
		return nil, fmt.Errorf("arn %q contains too few slashes", arn)
	}
	found := false
	for _, role := range a.opt.NodesRoles {
		if resource[1] == role {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("arn %q does not contain acceptable node role", arn)
	}

	instanceID := resource[2]
	instances, err := a.ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{instanceID}),
	})
	if err != nil {
		return nil, fmt.Errorf("describing instance for arn %q", arn)
	}

	if len(instances.Reservations) <= 0 || len(instances.Reservations[0].Instances) <= 0 {
		return nil, fmt.Errorf("missing instance id: %s", instanceID)
	}
	if len(instances.Reservations[0].Instances) > 1 {
		return nil, fmt.Errorf("found multiple instances with instance id: %s", instanceID)
	}

	return &fi.VerifyResult{
		NodeName: aws.StringValue(instances.Reservations[0].Instances[0].PrivateDnsName),
	}, nil
}
