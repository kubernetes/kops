/*
Copyright 2024 The Kubernetes Authors.

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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/go-cmp/cmp"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
)

func TestGetSTSRequestInfo(t *testing.T) {
	ctx := context.TODO()

	awsConfig := aws.Config{}
	awsConfig.Region = "us-east-1"
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider("fakeaccesskey", "fakesecretkey", "")
	sts := sts.NewFromConfig(awsConfig)

	stsRequestInfo, err := buildSTSRequestValidator(ctx, sts)
	if err != nil {
		t.Fatalf("error from getSTSRequestInfo: %v", err)
	}

	if got, want := stsRequestInfo.Host, "sts.us-east-1.amazonaws.com"; got != want {
		t.Errorf("unexpected host in sts request info; got %q, want %q", got, want)
	}

	grid := []struct {
		URL     string
		IsValid bool
	}{
		{
			URL:     "https://sts.us-east-1.amazonaws.com/",
			IsValid: false,
		},
		{
			URL:     "https://sts.us-east-1.amazonaws.com/Foo",
			IsValid: false,
		},
		{
			URL:     "https://sts.us-east-1.amazonaws.com/?Action=GetCallerIdentity",
			IsValid: true,
		},
		{
			URL:     "https://sts.us-east-1.amazonaws.com/Foo?Action=GetCallerIdentity",
			IsValid: false,
		},
		{
			URL:     "https://sts.us-east-1.amazonaws.com/?Action=GetCallerIdentity&Action=GetCallerIdentity",
			IsValid: false,
		},
	}

	for _, g := range grid {
		u, err := url.Parse(g.URL)
		if err != nil {
			t.Fatalf("parsing url %q: %v", g.URL, err)
		}
		got := stsRequestInfo.isValidV2(u)
		if got != g.IsValid {
			t.Errorf("unexpected result for IsValid(%v); got %v, want %v", g.URL, got, g.IsValid)
		}
	}

}

func TestVerifyTokenV1(t *testing.T) {
	mockSTS := &mockSTS{
		region: "us-east-1",
	}
	ctx := context.TODO()

	awsConfig := aws.Config{}
	awsConfig.Region = "us-east-1"
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider("fakeaccesskey", "fakesecretkey", "")
	awsConfig.HTTPClient = mockSTS.HTTPClient()
	stsClient := sts.NewFromConfig(awsConfig)

	body := []byte("test-body")

	var token string
	{

		a := &awsAuthenticator{
			sts:                 stsClient,
			credentialsProvider: awsConfig.Credentials,
			region:              awsConfig.Region,
		}

		v, err := a.createTokenV1(ctx, body)
		if err != nil {
			t.Fatalf("createTokenV1 failed: %v", err)
		}
		token = v
		t.Logf("token is %v", token)
	}

	{

		stsRequestValidator, err := buildSTSRequestValidator(ctx, stsClient)
		if err != nil {
			t.Fatalf("error from buildSTSRequestValidator: %v", err)
		}

		verifier := &awsVerifier{
			stsRequestValidator: stsRequestValidator,
			client:              *mockSTS.HTTPClient(),
		}
		var gotCallerIdentity *GetCallerIdentityResponse
		mockVerify := func(ctx context.Context, callerIdentity *GetCallerIdentityResponse) (*bootstrap.VerifyResult, error) {
			gotCallerIdentity = callerIdentity
			return nil, nil
		}
		if _, err := verifier.verifyTokenV1(ctx, token, body, mockVerify); err != nil {
			t.Fatalf("verifyTokenV1 failed: %v", err)
		}

		t.Logf("gotCallerIdentity is %+v", gotCallerIdentity)
		want := GetCallerIdentityResult{
			Arn:     "arn:aws:iam::123456789012:user/Alice",
			UserId:  "AIDACKCEVSQ6C2EXAMPLE",
			Account: "123456789012",
		}
		if diff := cmp.Diff(want, gotCallerIdentity.GetCallerIdentityResult[0]); diff != "" {
			t.Errorf("diff in GetCallerIdentityResult.  diff=%v", diff)
		}
	}
}

func TestVerifyTokenV2(t *testing.T) {
	mockSTS := &mockSTS{region: "us-east-1"}
	ctx := context.TODO()

	awsConfig := aws.Config{}
	awsConfig.Region = "us-east-1"
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider("fakeaccesskey", "fakesecretkey", "")
	awsConfig.HTTPClient = mockSTS.HTTPClient()
	stsClient := sts.NewFromConfig(awsConfig)

	body := []byte("test-body")

	var token string
	{

		a := &awsAuthenticator{
			sts:                 stsClient,
			credentialsProvider: awsConfig.Credentials,
			region:              awsConfig.Region,
		}

		v, err := a.createTokenV2(ctx, body)
		if err != nil {
			t.Fatalf("createTokenV2 failed: %v", err)
		}
		token = v
		t.Logf("token is %v", token)
	}

	{

		stsRequestValidator, err := buildSTSRequestValidator(ctx, stsClient)
		if err != nil {
			t.Fatalf("error from buildSTSRequestValidator: %v", err)
		}

		verifier := &awsVerifier{
			stsRequestValidator: stsRequestValidator,
			client:              *mockSTS.HTTPClient(),
		}
		var gotCallerIdentity *GetCallerIdentityResponse
		mockVerify := func(ctx context.Context, callerIdentity *GetCallerIdentityResponse) (*bootstrap.VerifyResult, error) {
			gotCallerIdentity = callerIdentity
			return nil, nil
		}
		if _, err := verifier.verifyTokenV2(ctx, token, body, mockVerify); err != nil {
			t.Fatalf("verifyTokenV1 failed: %v", err)
		}

		t.Logf("gotCallerIdentity is %+v", gotCallerIdentity)
		want := GetCallerIdentityResult{
			Arn:     "arn:aws:iam::123456789012:user/Alice",
			UserId:  "AIDACKCEVSQ6C2EXAMPLE",
			Account: "123456789012",
		}
		if diff := cmp.Diff(want, gotCallerIdentity.GetCallerIdentityResult[0]); diff != "" {
			t.Errorf("diff in GetCallerIdentityResult.  diff=%v", diff)
		}
	}
}

// TestVerifyAWSSDKV1Token verifies a token that was generated by 1.29.0, which uses aws-sdk-go v1
func TestVerifyAWSSDKV1Token(t *testing.T) {
	mockSTS := &mockSTS{region: "us-east-1"}
	ctx := context.TODO()

	awsConfig := aws.Config{}
	awsConfig.Region = "us-east-1"
	awsConfig.HTTPClient = mockSTS.HTTPClient()
	stsClient := sts.NewFromConfig(awsConfig)

	body := []byte("test-body")

	token := "x-aws-sts eyJBdXRob3JpemF0aW9uIjpbIkFXUzQtSE1BQy1TSEEyNTYgQ3JlZGVudGlhbD1mYWtlYWNjZXNza2V5LzIwMjQwNTE4L3VzLWVhc3QtMS9zdHMvYXdzNF9yZXF1ZXN0LCBTaWduZWRIZWFkZXJzPWNvbnRlbnQtbGVuZ3RoO2NvbnRlbnQtdHlwZTtob3N0O3gtYW16LWRhdGU7eC1rb3BzLXJlcXVlc3Qtc2hhLCBTaWduYXR1cmU9MGJkYWYyZTFmYjYxMGRjMjBlMzgyN2MzYWIwYzdiN2Q3ZWUzZjY5ZGMwZTUwNmQ3Y2Y3MTFlZjU5YTA4MzljZCJdLCJDb250ZW50LUxlbmd0aCI6WyI0MyJdLCJDb250ZW50LVR5cGUiOlsiYXBwbGljYXRpb24veC13d3ctZm9ybS11cmxlbmNvZGVkOyBjaGFyc2V0PXV0Zi04Il0sIlVzZXItQWdlbnQiOlsiYXdzLXNkay1nby8xLjUyLjYgKGdvMS4yMi4zOyBsaW51eDsgYW1kNjQpIl0sIlgtQW16LURhdGUiOlsiMjAyNDA1MThUMTg1OTA0WiJdLCJYLUtvcHMtUmVxdWVzdC1TaGEiOlsiMmRobHpGVHNZR2VQR3hHUWhLMTVybitUVjlIRVVaeGtWOTR6RkxmN3VvbyJdfQ=="

	{
		stsRequestValidator, err := buildSTSRequestValidator(ctx, stsClient)
		if err != nil {
			t.Fatalf("error from buildSTSRequestValidator: %v", err)
		}

		verifier := &awsVerifier{
			stsRequestValidator: stsRequestValidator,
			client:              *mockSTS.HTTPClient(),
		}
		var gotCallerIdentity *GetCallerIdentityResponse
		mockVerify := func(ctx context.Context, callerIdentity *GetCallerIdentityResponse) (*bootstrap.VerifyResult, error) {
			gotCallerIdentity = callerIdentity
			return nil, nil
		}
		if _, err := verifier.verifyTokenV1(ctx, token, body, mockVerify); err != nil {
			t.Fatalf("verifyTokenV1 failed: %v", err)
		}

		t.Logf("gotCallerIdentity is %+v", gotCallerIdentity)
		want := GetCallerIdentityResult{
			Arn:     "arn:aws:iam::123456789012:user/Alice",
			UserId:  "AIDACKCEVSQ6C2EXAMPLE",
			Account: "123456789012",
		}
		if diff := cmp.Diff(want, gotCallerIdentity.GetCallerIdentityResult[0]); diff != "" {
			t.Errorf("diff in GetCallerIdentityResult.  diff=%v", diff)
		}
	}

}

type mockSTS struct {
	region string
}

func (s *mockSTS) HTTPClient() *http.Client {
	return &http.Client{
		Transport: s,
	}
}

func (s *mockSTS) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	klog.Infof("mockSTS request: %+v", req)

	var payload []byte

	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			klog.Errorf("reading body: %v", err)
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
			}, nil
		}
		payload = b
		req.Body = io.NopCloser(bytes.NewReader(payload))
	}

	if err := req.ParseForm(); err != nil {
		klog.Warningf("error from ParseForm: %v", err)
		return &http.Response{
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	req.Body = io.NopCloser(bytes.NewReader(payload))
	klog.Infof("form is %+v", req.Form)

	action := req.Form.Get("Action")

	switch action {
	case "GetCallerIdentity":
		return s.GetCallerIdentity(ctx, req)
	}

	return &http.Response{
		StatusCode: http.StatusNotFound,
	}, nil
}

func (s *mockSTS) GetCallerIdentity(ctx context.Context, req *http.Request) (*http.Response, error) {
	klog.Infof("GetCallerIdentity request: %+v", req)

	authn := s.validateAuthorization(ctx, req)
	if authn == nil {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	body := `
<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
    <Arn>arn:aws:iam::123456789012:user/Alice</Arn>
    <UserId>AIDACKCEVSQ6C2EXAMPLE</UserId>
    <Account>123456789012</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>
`
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}, nil
}

type mockAuthInfo struct {
	credential string
}

func (s *mockSTS) validateAuthorization(ctx context.Context, req *http.Request) *mockAuthInfo {
	service := "sts"

	amzSignature := req.Form.Get("X-Amz-Signature")
	amzCredential := req.Form.Get("X-Amz-Credential")
	// amzSignedHeaders := ""

	authorization := req.Header.Get("authorization")
	if authorization != "" {
		if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 ") {
			klog.Errorf("unexpected authorization value %q", authorization)
			return nil
		}
		authorization = strings.TrimPrefix(authorization, "AWS4-HMAC-SHA256 ")

		for _, token := range strings.Split(strings.TrimPrefix(authorization, "AWS4-HMAC-SHA256 "), ", ") {
			kv := strings.SplitN(token, "=", 2)
			if len(kv) == 1 {
				return nil
			}
			got := kv[1]
			switch kv[0] {
			case "Signature":
				amzSignature = got
			case "Credential":
				amzCredential = got
			case "SignedHeaders":
				// amzSignedHeaders = got
			default:
				klog.Errorf("unexpected authorization value %q", authorization)
				return nil
			}
		}
	}

	if amzCredential == "" {
		klog.Errorf("credential not supplied")
		return nil
	}
	// If this was not a mock, we would look up the credential here
	credentials := aws.Credentials{
		AccessKeyID:     "fakeaccesskey",
		SecretAccessKey: "fakesecretkey",
	}

	computedSignature, err := computeSignatureV4(req, s.region, service, credentials.SecretAccessKey)
	if err != nil {
		klog.Errorf("computing signature: %v", err)
		return nil
	}

	if computedSignature != amzSignature {
		klog.Errorf("authorization did not match")
		return nil
	}
	return &mockAuthInfo{
		credential: amzCredential,
	}
}

// computeSignatureV4 computes the expected AWS4-HMAC-SHA256 signature
func computeSignatureV4(req *http.Request, region string, service string, secretAccessKey string) (string, error) {
	var payload []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return "", fmt.Errorf("reading body: %w", err)
		}
		payload = b
	}
	payloadHash := sha256.Sum256(payload)
	payloadHashHex := hex.EncodeToString(payloadHash[:])

	canonicalURI := req.URL.Path
	var canonicalQueryString string
	{
		var b strings.Builder

		var keys []string
		for k := range req.URL.Query() {
			if k == "X-Amz-Signature" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i != 0 {
				b.WriteString("&")
			}
			values := req.URL.Query()[k]
			b.WriteString(url.QueryEscape(k))
			b.WriteString("=")
			for j, v := range values {
				if j != 0 {
					b.WriteString(",")
				}
				b.WriteString(url.QueryEscape(v))
			}
		}
		canonicalQueryString = b.String()
	}

	signedHeaders := ""
	if s := req.Form.Get("X-Amz-SignedHeaders"); s != "" {
		signedHeaders = s
	}

	authorization := req.Header.Get("authorization")
	if authorization != "" {
		if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 ") {
			return "", fmt.Errorf("unexpected authorization prefix")
		}
		authorization = strings.TrimPrefix(authorization, "AWS4-HMAC-SHA256 ")

		for _, token := range strings.Split(strings.TrimPrefix(authorization, "AWS4-HMAC-SHA256 "), ", ") {
			kv := strings.SplitN(token, "=", 2)
			if len(kv) == 1 {
				return "", fmt.Errorf("unexpected authorization token")
			}
			got := kv[1]
			switch kv[0] {
			case "Signature":
				// amzSignature = got
			case "Credential":
				// amzCredential = got
			case "SignedHeaders":
				signedHeaders = got
			default:
				return "", fmt.Errorf("unexpected authorization token")
			}
		}
	}

	var canonicalHeaders string

	{
		var b strings.Builder

		for _, key := range strings.Split(signedHeaders, ";") {
			k := strings.ToLower(key)
			var values []string
			switch k {
			case "host":
				values = []string{req.URL.Host}

			case "content-length":
				values = []string{strconv.FormatInt(req.ContentLength, 10)}

			default:
				values = req.Header.Values(k)
			}

			b.WriteString(url.QueryEscape(k))
			b.WriteString(":")
			for j, v := range values {
				if j != 0 {
					b.WriteString(",")
				}
				v = strings.TrimSpace(v)
				b.WriteString(v)
			}

			b.WriteString("\n")
		}
		canonicalHeaders = b.String()
	}
	canonicalRequest := req.Method + "\n" + canonicalURI + "\n" + canonicalQueryString + "\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + payloadHashHex

	reqTime := req.Header.Get("X-Amz-Date")
	if reqTime == "" {
		reqTime = req.Form.Get("X-Amz-Date")
	}
	if reqTime == "" {
		return "", fmt.Errorf("cannot determine signature date")
	}
	timeStampISO8601Format := reqTime

	today := reqTime[:strings.Index(reqTime, "T")]

	scope := today + "/" + region + "/" + service + "/aws4_request"

	canonicalRequestSHA256 := sha256.Sum256([]byte(canonicalRequest))
	stringToSign := "AWS4-HMAC-SHA256" + "\n" + timeStampISO8601Format + "\n" + scope + "\n" + hex.EncodeToString(canonicalRequestSHA256[:])

	hmacSHA256 := func(key []byte, data string) ([]byte, error) {
		signer := hmac.New(sha256.New, key)
		if _, err := signer.Write([]byte(data)); err != nil {
			return nil, err
		}
		return signer.Sum(nil), nil
	}

	dateKey, err := hmacSHA256([]byte("AWS4"+secretAccessKey), today)
	if err != nil {
		return "", nil
	}

	dateRegionKey, err := hmacSHA256(dateKey, region)
	if err != nil {
		return "", nil
	}

	dateRegionServiceKey, err := hmacSHA256(dateRegionKey, service)
	if err != nil {
		return "", nil
	}
	signingKey, err := hmacSHA256(dateRegionServiceKey, "aws4_request")
	if err != nil {
		return "", nil
	}
	signature, err := hmacSHA256(signingKey, stringToSign)
	if err != nil {
		return "", nil
	}

	return hex.EncodeToString(signature), nil
}
