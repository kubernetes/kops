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
	"strings"
	"testing"

	// "github.com/aws/aws-sdk-go-v2/aws"
	// "github.com/aws/aws-sdk-go-v2/credentials"
	// "github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func TestAWSPresign(t *testing.T) {
	// mockSTSServer := &mockSTSServer{t: t}
	// awsConfig := aws.Config{}
	// awsConfig.Region = "us-east-1"
	// awsConfig.Credentials = credentials.NewStaticCredentialsProvider("accesskey", "secretkey", "")
	// awsConfig.HTTPClient = mockSTSServer
	// sts := sts.NewFromConfig(awsConfig)

	mySession := session.Must(session.NewSession())
	mySession.Config.Credentials = credentials.NewStaticCredentials("accesskey", "secretkey", "")
	sts := sts.New(mySession)
	mySession.Config.HTTPClient = &http.Client{Transport: &mockHTTPTransport{}}
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
	headers := make(map[string][]string)
	if err := json.Unmarshal([]byte(data), &headers); err != nil {
		t.Fatalf("decoding token as json: %v", err)
	}

	t.Logf("headers: %+v", headers)

	authorization := ""
	for header, values := range headers {
		got := strings.Join(values, "||")
		switch header {
		case "User-Agent":
			// Ignore
			// TODO: Should we (can we) override the useragent?
		case "X-Amz-Date":
			if len(got) < 10 {
				t.Errorf("expected %q header of at least 10 characters, got %q", header, got)
			}
		case "Content-Length":
			if want := "43"; got != want {
				t.Errorf("unexpected %q header: got %q, want %q", header, got, want)
			}
		case "Content-Type":
			if want := "application/x-www-form-urlencoded; charset=utf-8"; got != want {
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
			t.Errorf("unexpected header %q", header)
		}
	}

	if !strings.HasPrefix(authorization, "AWS4-HMAC-SHA256 ") {
		t.Errorf("unexpected authorization prefix, got %q", authorization)
	}
	for _, token := range strings.Split(strings.TrimPrefix(authorization, "AWS4-HMAC-SHA256 "), ", ") {
		kv := strings.SplitN(token, "=", 2)
		got := kv[1]
		switch kv[0] {
		case "Signature":
			if len(got) < 10 {
				t.Errorf("expected %q Authorization value of at least 10 characters, got %q", kv[0], got)
			}
		case "Credential":
			if len(got) < 10 {
				t.Errorf("expected %q Authorization value of at least 10 characters, got %q", kv[0], got)
			}
		case "SignedHeaders":
			if want := "content-length;content-type;host;x-amz-date;x-kops-request-sha"; got != want {
				t.Errorf("unexpected %q Authorization value: got %q, want %q", kv[0], got, want)
			}
		default:
			t.Errorf("unknown token %q in authorization header", token)
		}
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
