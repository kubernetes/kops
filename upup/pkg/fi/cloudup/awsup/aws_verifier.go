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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"k8s.io/kops/upup/pkg/fi"
)

type awsVerifier struct {
	sts    *sts.STS
	client http.Client
}

var _ fi.Verifier = &awsVerifier{}

func NewAWSVerifier() (fi.Verifier, error) {
	config := aws.NewConfig().WithCredentialsChainVerboseErrors(true)
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return &awsVerifier{
		sts: sts.New(sess),
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

func (a awsVerifier) VerifyToken(token string, body []byte) (string, error) {
	if !strings.HasPrefix(token, AWSAuthenticationTokenPrefix) {
		return "", fmt.Errorf("incorrect authorization type")
	}
	token = strings.TrimPrefix(token, AWSAuthenticationTokenPrefix)

	// We rely on the client and server using the same version of the same STS library.
	stsRequest, _ := a.sts.GetCallerIdentityRequest(nil)
	err := stsRequest.Sign()
	if err != nil {
		return "", fmt.Errorf("creating identity request: %v", err)
	}

	stsRequest.HTTPRequest.Header = nil
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("decoding authorization token: %v", err)
	}
	err = json.Unmarshal(tokenBytes, &stsRequest.HTTPRequest.Header)
	if err != nil {
		return "", fmt.Errorf("unmarshalling authorization token: %v", err)
	}

	sha := sha256.Sum256(body)
	if stsRequest.HTTPRequest.Header.Get("X-Kops-Request-SHA") != base64.RawStdEncoding.EncodeToString(sha[:]) {
		return "", fmt.Errorf("incorrect SHA")
	}

	requestBytes, _ := ioutil.ReadAll(stsRequest.Body)
	_, _ = stsRequest.Body.Seek(0, io.SeekStart)
	if stsRequest.HTTPRequest.Header.Get("Content-Length") != strconv.Itoa(len(requestBytes)) {
		return "", fmt.Errorf("incorrect content-length")
	}

	// TODO - implement retry?
	response, err := a.client.Do(stsRequest.HTTPRequest)
	if err != nil {
		return "", fmt.Errorf("sending STS request: %v", err)
	}
	if response != nil {
		defer response.Body.Close()
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("reading STS response: %v", err)
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("received status code %d from STS: %s", response.StatusCode, string(responseBody))
	}

	result := GetCallerIdentityResponse{}
	err = xml.NewDecoder(bytes.NewReader(responseBody)).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("decoding STS response: %v", err)
	}

	marshal, _ := json.Marshal(result)
	return string(marshal), nil
}
