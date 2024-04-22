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
	"context"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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
		got := stsRequestInfo.IsValid(u)
		if got != g.IsValid {
			t.Errorf("unexpected result for IsValid(%v); got %v, want %v", g.URL, got, g.IsValid)
		}
	}

}
