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
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/bootstrap"
	nodeidentityaws "k8s.io/kops/pkg/nodeidentity/aws"
	"k8s.io/kops/pkg/wellknownports"
)

type AWSVerifierOptions struct {
	// NodesRoles are the IAM roles that worker nodes are permitted to have.
	NodesRoles []string `json:"nodesRoles"`
	// Region is the AWS region of the cluster.
	Region string
}

type awsVerifier struct {
	accountId string
	partition string
	opt       AWSVerifierOptions

	ec2    *ec2.Client
	client http.Client

	stsRequestValidator *stsRequestValidator
}

var _ bootstrap.Verifier = &awsVerifier{}

func NewAWSVerifier(ctx context.Context, opt *AWSVerifierOptions) (bootstrap.Verifier, error) {
	config, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(opt.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	stsClient := sts.NewFromConfig(config)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}

	partition := strings.Split(aws.ToString(identity.Arn), ":")[1]

	ec2Client := ec2.NewFromConfig(config)

	stsRequestValidator, err := buildSTSRequestValidator(ctx, stsClient)
	if err != nil {
		return nil, err
	}

	return &awsVerifier{
		accountId:           aws.ToString(identity.Account),
		partition:           partition,
		opt:                 *opt,
		ec2:                 ec2Client,
		stsRequestValidator: stsRequestValidator,
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

func (a awsVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, AWSAuthenticationTokenPrefix) {
		return nil, bootstrap.ErrNotThisVerifier
	}
	token = strings.TrimPrefix(token, AWSAuthenticationTokenPrefix)

	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("decoding authorization token: %v", err)
	}
	var decoded awsV2Token
	if err := json.Unmarshal(tokenBytes, &decoded); err != nil {
		return nil, fmt.Errorf("unmarshalling authorization token: %v", err)
	}

	// Verify the token has signed the body content.
	sha := sha256.Sum256(body)
	if decoded.SignedHeader.Get("X-Kops-Request-SHA") != base64.RawStdEncoding.EncodeToString(sha[:]) {
		return nil, fmt.Errorf("incorrect SHA")
	}

	reqURL, err := url.Parse(decoded.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing STS request URL: %v", err)
	}
	signedHeaders := sets.New(strings.Split(reqURL.Query().Get("X-Amz-SignedHeaders"), ",")...)
	if !signedHeaders.Has("x-kops-request-sha") {
		return nil, fmt.Errorf("unexpected signed headers value")
	}

	if !a.stsRequestValidator.IsValid(reqURL) {
		return nil, fmt.Errorf("invalid STS url: host=%q, path=%q", reqURL.Host, reqURL.Path)
	}

	callerIdentity, err := a.stsRequestValidator.GetCallerIdentity(ctx, &a.client, &decoded)
	if err != nil {
		return nil, err
	}
	if callerIdentity.GetCallerIdentityResult[0].Account != a.accountId {
		return nil, fmt.Errorf("incorrect account %s", callerIdentity.GetCallerIdentityResult[0].Account)
	}

	arn := callerIdentity.GetCallerIdentityResult[0].Arn
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
	instances, err := a.ec2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
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

	instance := instances.Reservations[0].Instances[0]

	addrs, err := GetInstanceCertificateNames(instances)
	if err != nil {
		return nil, err
	}

	var challengeEndpoints []string
	for _, nic := range instance.NetworkInterfaces {
		if ip := aws.ToString(nic.PrivateIpAddress); ip != "" {
			challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(ip, strconv.Itoa(wellknownports.NodeupChallenge)))
		}
		for _, a := range nic.PrivateIpAddresses {
			if ip := aws.ToString(a.PrivateIpAddress); ip != "" {
				challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(ip, strconv.Itoa(wellknownports.NodeupChallenge)))
			}
		}

		for _, a := range nic.Ipv6Addresses {
			if ip := aws.ToString(a.Ipv6Address); ip != "" {
				challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(ip, strconv.Itoa(wellknownports.NodeupChallenge)))
			}
		}
	}

	if len(challengeEndpoints) == 0 {
		return nil, fmt.Errorf("cannot determine challenge endpoint for instance id: %s", instanceID)
	}

	result := &bootstrap.VerifyResult{
		NodeName:          addrs[0],
		CertificateNames:  addrs,
		ChallengeEndpoint: challengeEndpoints[0],
	}

	for _, tag := range instance.Tags {
		tagKey := aws.ToString(tag.Key)
		if tagKey == nodeidentityaws.CloudTagInstanceGroupName {
			result.InstanceGroupName = aws.ToString(tag.Value)
		}
	}

	return result, nil
}

// stsRequestValidator describes valid STS Presigned URLs, and is used to validate client authentication requests.
type stsRequestValidator struct {
	Host string
}

// IsValid performs some basic pre-validation of the request URL.
func (s *stsRequestValidator) IsValid(u *url.URL) bool {
	if u.Host != s.Host {
		return false
	}
	if u.Path != "/" {
		return false
	}
	if u.Query().Get("Action") != "GetCallerIdentity" {
		return false
	}
	if len(u.Query()["Action"]) != 1 {
		return false
	}

	return true
}

// GetCallerIdentity will request the presigned token URL, and decode the returned identity.
func (s *stsRequestValidator) GetCallerIdentity(ctx context.Context, httpClient *http.Client, decoded *awsV2Token) (*GetCallerIdentityResponse, error) {
	reqURL, err := url.Parse(decoded.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing STS request URL: %w", err)
	}

	if !s.IsValid(reqURL) {
		return nil, fmt.Errorf("url not valid for STS request")
	}

	req := &http.Request{
		URL:    reqURL,
		Method: decoded.Method,
		Header: decoded.SignedHeader,
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending STS request: %v", err)
	}
	if response != nil {
		defer response.Body.Close()
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading STS response: %v", err)
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("received status code %d from STS: %s", response.StatusCode, string(responseBody))
	}

	callerIdentity := &GetCallerIdentityResponse{}
	err = xml.NewDecoder(bytes.NewReader(responseBody)).Decode(callerIdentity)
	if err != nil {
		return nil, fmt.Errorf("decoding STS response: %v", err)
	}

	return callerIdentity, nil
}

// buildSTSRequestValidator determines the form of a valid STS presigned URL.
func buildSTSRequestValidator(ctx context.Context, stsClient *sts.Client) (*stsRequestValidator, error) {
	// We build a presigned token ourselves, primarily to get the expected hostname for the endpoint.
	signed, err := sts.NewPresignClient(stsClient).PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("building presigned request: %w", err)
	}
	u, err := url.Parse(signed.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing presigned url: %w", err)
	}
	return &stsRequestValidator{Host: u.Host}, nil
}
