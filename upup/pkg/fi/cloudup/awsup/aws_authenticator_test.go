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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/go-cmp/cmp"
)

func TestAWSV1Request(t *testing.T) {
	ctx := context.TODO()

	var wantRequest *http.Request
	// This is a well-known V1 value corresponding to the "test-body" kops-request body
	// along with credentials.NewStaticCredentialsProvider("fakeaccesskey", "fakesecretkey", "").
	// It was generated using the _old_ signing code (that used aws-sdk-go v1).
	{
		body := []byte("Action=GetCallerIdentity&Version=2011-06-15")
		r, err := http.NewRequest("POST", "https://sts.amazonaws.com/", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("building http request: %v", err)
		}

		auth := []string{
			"AWS4-HMAC-SHA256 Credential=fakeaccesskey/20240518/us-east-1/sts/aws4_request",
			"SignedHeaders=content-length;content-type;host;x-amz-date;x-kops-request-sha",
			"Signature=198684464845d2d52947df10171b6291e1b2223ce4bd82a380087761d91246f9",
		}
		r.Header.Add("Authorization", strings.Join(auth, ", "))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		r.Header.Add("X-Amz-Date", "20240518T131902Z")
		r.Header.Add("X-Kops-Request-Sha", "2dhlzFTsYGePGxGQhK15rn+TV9HEUZxkV94zFLf7uoo")
		wantRequest = r
	}

	credentials, err := credentials.NewStaticCredentialsProvider("fakeaccesskey", "fakesecretkey", "").Retrieve(ctx)
	if err != nil {
		t.Fatalf("getting credentials: %v", err)
	}
	signingTime := time.Date(2024, time.May, 18, 13, 19, 02, 0, time.UTC)

	body := []byte("test-body")
	stsURL := "https://sts.amazonaws.com/"
	region := "us-east-1"
	signedRequest, err := signV1Request(ctx, stsURL, region, credentials, signingTime, body)
	if err != nil {
		t.Fatalf("error from signV1Request: %v", err)
	}
	t.Logf("signedRequest is %+v", signedRequest)
	if diff := cmp.Diff(signedRequest.Header, wantRequest.Header); diff != "" {
		t.Errorf("headers did not match: %v", diff)
	}
}

func TestCreateTokenV1(t *testing.T) {
	ctx := context.TODO()

	mockSTSServer := &mockHTTPClient{t: t}
	awsConfig := aws.Config{}
	awsConfig.Region = "us-east-1"
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider("fakeaccesskey", "fakesecretkey", "")
	awsConfig.HTTPClient = mockSTSServer
	sts := sts.NewFromConfig(awsConfig)

	a := &awsAuthenticator{
		sts:                 sts,
		credentialsProvider: awsConfig.Credentials,
		region:              awsConfig.Region,
	}

	body := []byte("test-body")
	bodyHash := sha256.Sum256(body)
	bodyHashBase64 := base64.RawStdEncoding.EncodeToString(bodyHash[:])
	if bodyHashBase64 != "2dhlzFTsYGePGxGQhK15rn+TV9HEUZxkV94zFLf7uoo" {
		t.Fatalf("unexpected hash of body; got %q", bodyHashBase64)
	}

	token, err := a.createTokenV1(ctx, body)
	if err != nil {
		t.Fatalf("error from createTokenV1: %v", err)
	}
	if !strings.HasPrefix(token, "x-aws-sts ") {
		t.Fatalf("expected token to start with x-aws-sts; got %q", token)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(token, "x-aws-sts "))
	if err != nil {
		t.Fatalf("decoding token as base64: %v", err)
	}
	var decoded http.Header
	if err := json.Unmarshal([]byte(data), &decoded); err != nil {
		t.Fatalf("decoding token as json: %v", err)
	}
	t.Logf("decoded: %+v", decoded)

	signatureDate := time.Now().Format("20060102") // This might fail if we cross midnight, it seems very unlikely

	amzSignature := ""
	amzSignedHeaders := ""
	amzDate := ""
	amzAlgorithm := ""
	amzCredential := ""
	host := ""
	authorization := ""
	kopsRequestSHA := ""

	for header, values := range decoded {
		got := strings.Join(values, "||")
		switch header {
		case "X-Amz-Date":
			amzDate = got

		case "Content-Length":
			if want := "43"; got != want {
				t.Errorf("unexpected %q header: got %q, want %q", header, got, want)
			}
		case "Content-Type":
			if want := "application/x-www-form-urlencoded; charset=utf-8"; got != want {
				t.Errorf("unexpected %q header: got %q, want %q", header, got, want)
			}

		case "X-Kops-Request-Sha":
			kopsRequestSHA = got

		case "Authorization":
			authorization = got

		default:
			t.Errorf("unexpected header %q: %q", header, got)
		}
	}

	//  This is only for the version 1 token
	if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 ") {
		t.Errorf("unexpected authorization prefix, got %q", authorization)
	}
	amzAlgorithm = "AWS4-HMAC-SHA256"

	for _, token := range strings.Split(strings.TrimPrefix(authorization, "AWS4-HMAC-SHA256 "), ", ") {
		kv := strings.SplitN(token, "=", 2)
		if len(kv) == 1 {
			t.Errorf("invalid token %q in authorization header", token)
			continue
		}
		got := kv[1]
		switch kv[0] {
		case "Signature":
			amzSignature = got
		case "Credential":
			amzCredential = got
		case "SignedHeaders":
			amzSignedHeaders = got
		default:
			t.Errorf("unknown token %q in authorization header", token)
		}
	}

	// host is not included in v1 token (but is signed) - a shortcoming addressed in v2
	stsHost, err := a.getSTSHost(ctx)
	if err != nil {
		t.Fatalf("getting AWS STS url: %v", err)
	}
	host = stsHost

	if len(amzCredential) < 10 {
		t.Errorf("expected amzCredential value of at least 10 characters, got %q", amzCredential)
	}

	if want := "fakeaccesskey/" + signatureDate + "/" + awsConfig.Region + "/sts/aws4_request"; amzCredential != want {
		t.Errorf("unexpected amzCredential: got %q, want %q", amzCredential, want)
	}

	if len(amzDate) < 10 {
		t.Errorf("expected amz-date of at least 10 characters, got %q", amzDate)
	}
	if wantPrefix := signatureDate + "T"; !strings.HasPrefix(amzDate, wantPrefix) {
		t.Errorf("expected amz-date to have prefix %q; got %q", wantPrefix, amzDate)
	}

	if len(amzSignature) < 10 {
		t.Errorf("expected amzSignature value of at least 10 characters, got %q", amzSignature)
	}

	if want := "AWS4-HMAC-SHA256"; amzAlgorithm != want {
		t.Errorf("unexpected amzAlgorithm: got %q, want %q", amzAlgorithm, want)
	}

	if want := "sts.us-east-1.amazonaws.com"; host != want {
		t.Errorf("unexpected host %q, want %q", host, want)
	}

	if want := "content-length;content-type;host;x-amz-date;x-kops-request-sha"; amzSignedHeaders != want {
		t.Errorf("unexpected amzSignedHeaders: got %q, want %q", amzSignedHeaders, want)
	}

	if want := bodyHashBase64; kopsRequestSHA != want {
		t.Errorf("unexpected kopsRequestSHA: got %q, want %q", kopsRequestSHA, want)
	}
}

func TestCreateTokenV2(t *testing.T) {
	ctx := context.TODO()

	mockSTSServer := &mockHTTPClient{t: t}
	awsConfig := aws.Config{}
	awsConfig.Region = "us-east-1"
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider("fakeaccesskey", "fakesecretkey", "")
	awsConfig.HTTPClient = mockSTSServer
	sts := sts.NewFromConfig(awsConfig)

	a := &awsAuthenticator{
		sts:                 sts,
		credentialsProvider: awsConfig.Credentials,
		region:              awsConfig.Region,
	}

	body := []byte("test-body")
	bodyHash := sha256.Sum256(body)
	bodyHashBase64 := base64.RawStdEncoding.EncodeToString(bodyHash[:])
	if bodyHashBase64 != "2dhlzFTsYGePGxGQhK15rn+TV9HEUZxkV94zFLf7uoo" {
		t.Fatalf("unexpected hash of body; got %q", bodyHashBase64)
	}

	token, err := a.createTokenV2(ctx, body)
	if err != nil {
		t.Fatalf("error from CreateToken: %v", err)
	}
	if !strings.HasPrefix(token, "x-aws-sts-v2 ") {
		t.Fatalf("expected token to start with x-aws-sts-v2; got %q", token)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(token, "x-aws-sts-v2 "))
	if err != nil {
		t.Fatalf("decoding token as base64: %v", err)
	}
	decoded := &awsV2Token{}
	if err := json.Unmarshal([]byte(data), &decoded); err != nil {
		t.Fatalf("decoding token as json: %v", err)
	}
	t.Logf("decoded: %+v", decoded)

	if want := "GET"; decoded.Method != want {
		t.Errorf("unexpected http method: got %q, want %q", decoded.Method, want)
	}

	signatureDate := time.Now().Format("20060102") // This might fail if we cross midnight, it seems very unlikely

	amzSignature := ""
	amzSignedHeaders := ""
	amzDate := ""
	amzAlgorithm := ""
	amzCredential := ""
	host := ""
	kopsRequestSHA := ""

	for header, values := range decoded.SignedHeader {
		got := strings.Join(values, "||")
		switch header {
		case "X-Amz-Date":
			amzDate = got

		case "Host":
			host = got

		case "X-Kops-Request-Sha":
			kopsRequestSHA = got

		default:
			t.Errorf("unexpected header %q: %q", header, got)
		}
	}

	u, err := url.Parse(decoded.URL)
	if err != nil {
		t.Errorf("error parsing url %q: %v", decoded.URL, err)
	}
	for k, values := range u.Query() {
		got := strings.Join(values, "||")

		switch k {
		case "Action":
			if want := "GetCallerIdentity"; got != want {
				t.Errorf("unexpected %q query param: got %q, want %q", k, got, want)
			}
		case "Version":
			if want := "2011-06-15"; got != want {
				t.Errorf("unexpected %q query param: got %q, want %q", k, got, want)
			}
		case "X-Amz-Date":
			amzDate = got
		case "X-Amz-Signature":
			amzSignature = got
		case "X-Amz-Credential":
			amzCredential = got
		case "X-Amz-SignedHeaders":
			amzSignedHeaders = got
		case "X-Amz-Algorithm":
			amzAlgorithm = got
		default:
			t.Errorf("unknown token %q=%q in query", k, got)
		}
	}

	if len(amzCredential) < 10 {
		t.Errorf("expected amzCredential value of at least 10 characters, got %q", amzCredential)
	}

	if want := "fakeaccesskey/" + signatureDate + "/" + awsConfig.Region + "/sts/aws4_request"; amzCredential != want {
		t.Errorf("unexpected amzCredential: got %q, want %q", amzCredential, want)
	}

	if len(amzDate) < 10 {
		t.Errorf("expected amz-date of at least 10 characters, got %q", amzDate)
	}
	if wantPrefix := signatureDate + "T"; !strings.HasPrefix(amzDate, wantPrefix) {
		t.Errorf("expected amz-date to have prefix %q; got %q", wantPrefix, amzDate)
	}

	if len(amzSignature) < 10 {
		t.Errorf("expected amzSignature value of at least 10 characters, got %q", amzSignature)
	}

	if want := "AWS4-HMAC-SHA256"; amzAlgorithm != want {
		t.Errorf("unexpected amzAlgorithm: got %q, want %q", amzAlgorithm, want)
	}

	if want := "sts.us-east-1.amazonaws.com"; host != want {
		t.Errorf("unexpected host %q, want %q", host, want)
	}

	if want := "host;x-kops-request-sha"; amzSignedHeaders != want {
		t.Errorf("unexpected amzSignedHeaders: got %q, want %q", amzSignedHeaders, want)
	}

	if want := bodyHashBase64; kopsRequestSHA != want {
		t.Errorf("unexpected kopsRequestSHA: got %q, want %q", kopsRequestSHA, want)
	}
}

type mockHTTPClient struct {
	t *testing.T
}

func (s *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	s.t.Fatalf("unexpected request %+v", req)
	return nil, fmt.Errorf("unexpected request")
}

type mockHTTPTransport struct {
	t *testing.T
}

func (s *mockHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	s.t.Fatalf("unexpected request %+v", req)
	return nil, fmt.Errorf("unexpected request")
}
