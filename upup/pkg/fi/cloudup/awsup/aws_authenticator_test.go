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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func TestAWSPresign(t *testing.T) {
	mockSTSServer := &mockHTTPClient{t: t}
	awsConfig := aws.Config{}
	awsConfig.Region = "us-east-1"
	awsConfig.Credentials = credentials.NewStaticCredentialsProvider("fakeaccesskey", "fakesecretkey", "")
	awsConfig.HTTPClient = mockSTSServer
	sts := sts.NewFromConfig(awsConfig)

	a := &awsAuthenticator{
		sts: sts,
	}

	body := []byte("test-body")
	bodyHash := sha256.Sum256(body)
	bodyHashBase64 := base64.RawStdEncoding.EncodeToString(bodyHash[:])
	if bodyHashBase64 != "2dhlzFTsYGePGxGQhK15rn+TV9HEUZxkV94zFLf7uoo" {
		t.Fatalf("unexpected hash of body; got %q", bodyHashBase64)
	}

	token, err := a.CreateToken(body)
	if err != nil {
		t.Fatalf("error from CreateToken: %v", err)
	}
	if !strings.HasPrefix(token, "x-aws-sts ") {
		t.Fatalf("expected token to start with x-aws-sts; got %q", token)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(token, "x-aws-sts "))
	if err != nil {
		t.Fatalf("decoding token as base64: %v", err)
	}
	decoded := &awsV2Token{}
	if err := json.Unmarshal([]byte(data), &decoded); err != nil {
		t.Fatalf("decoding token as json: %v", err)
	}

	t.Logf("decoded: %+v", decoded)

	amzSignature := ""
	amzSignedHeaders := ""
	amzDate := ""
	amzAlgorithm := ""
	amzCredential := ""

	authorization := ""
	for header, values := range decoded.SignedHeader {
		got := strings.Join(values, "||")
		switch header {
		case "User-Agent":
			// Ignore
			// TODO: Should we (can we) override the useragent?
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

		case "Host":
			if want := "sts.us-east-1.amazonaws.com"; got != want {
				t.Errorf("unexpected %q header: got %q, want %q", header, got, want)
			}

		case "X-Kops-Request-Sha":
			if want := bodyHashBase64; got != want {
				t.Errorf("unexpected %q header: got %q, want %q", header, got, want)
			}
		case "Authorization":
			// Validated more deeply below
			authorization = got
		default:
			t.Errorf("unexpected header %q: %q", header, got)
		}
	}

	// TODO: This is only aws-sdk-go V1
	if authorization != "" {
		if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 ") {
			t.Errorf("unexpected authorization prefix, got %q", authorization)
		}

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
			amzDate = k
		case "X-Amz-Signature":
			amzSignature = k
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

	if len(amzDate) < 10 {
		t.Errorf("expected amz-date of at least 10 characters, got %q", amzDate)
	}

	if len(amzSignature) < 10 {
		t.Errorf("expected amzSignature value of at least 10 characters, got %q", amzSignature)
	}

	if want := "AWS4-HMAC-SHA256"; amzAlgorithm != want {
		t.Errorf("unexpected amzAlgorithm: got %q, want %q", amzAlgorithm, want)
	}

	if want := "host;x-kops-request-sha"; amzSignedHeaders != want {
		t.Errorf("unexpected amzSignedHeaders: got %q, want %q", amzSignedHeaders, want)
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
