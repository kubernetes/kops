/*
Copyright 2026 The Kubernetes Authors.

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

package main

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

// fakeEC2 answers GetConsoleOutput, optionally rejecting requests for the latest output.
type fakeEC2 struct {
	output      *string
	latestErr   error
	snapshotErr error

	latestRequests   int
	snapshotRequests int
	instanceIDs      []string
}

func (f *fakeEC2) GetConsoleOutput(ctx context.Context, params *ec2.GetConsoleOutputInput, optFns ...func(*ec2.Options)) (*ec2.GetConsoleOutputOutput, error) {
	f.instanceIDs = append(f.instanceIDs, aws.ToString(params.InstanceId))
	if aws.ToBool(params.Latest) {
		f.latestRequests++
		if f.latestErr != nil {
			return nil, f.latestErr
		}
	} else {
		f.snapshotRequests++
		if f.snapshotErr != nil {
			return nil, f.snapshotErr
		}
	}
	return &ec2.GetConsoleOutputOutput{InstanceId: params.InstanceId, Output: f.output}, nil
}

func TestGetEC2ConsoleOutput(t *testing.T) {
	const instanceID = "i-0123"
	encoded := aws.String(base64.StdEncoding.EncodeToString([]byte("[    0.000000] Booting Linux\n")))
	unsupported := &smithy.GenericAPIError{Code: "UnsupportedOperation", Message: "latest output is not supported"}

	grid := []struct {
		name    string
		ec2     *fakeEC2
		want    string
		wantErr bool
		// wantErrCode is the API error code the returned error must wrap.
		wantErrCode          string
		wantLatestRequests   int
		wantSnapshotRequests int
	}{
		{
			name:               "latest output",
			ec2:                &fakeEC2{output: encoded},
			want:               "[    0.000000] Booting Linux\n",
			wantLatestRequests: 1,
		},
		{
			name:               "nothing captured yet",
			ec2:                &fakeEC2{},
			wantLatestRequests: 1,
		},
		{
			name:                 "latest unsupported falls back to snapshot",
			ec2:                  &fakeEC2{output: encoded, latestErr: unsupported},
			want:                 "[    0.000000] Booting Linux\n",
			wantLatestRequests:   1,
			wantSnapshotRequests: 1,
		},
		{
			name:                 "snapshot fallback also fails",
			ec2:                  &fakeEC2{latestErr: unsupported, snapshotErr: &smithy.GenericAPIError{Code: "InvalidInstanceID.NotFound"}},
			wantErr:              true,
			wantErrCode:          "InvalidInstanceID.NotFound",
			wantLatestRequests:   1,
			wantSnapshotRequests: 1,
		},
		{
			name:               "other API errors are not retried",
			ec2:                &fakeEC2{latestErr: &smithy.GenericAPIError{Code: "UnauthorizedOperation"}},
			wantErr:            true,
			wantErrCode:        "UnauthorizedOperation",
			wantLatestRequests: 1,
		},
		{
			name:               "non-API error is not retried",
			ec2:                &fakeEC2{latestErr: errors.New("connection reset")},
			wantErr:            true,
			wantLatestRequests: 1,
		},
		{
			name:               "invalid base64",
			ec2:                &fakeEC2{output: aws.String("not base64!")},
			wantErr:            true,
			wantLatestRequests: 1,
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			got, err := getEC2ConsoleOutput(t.Context(), g.ec2, instanceID)
			if (err != nil) != g.wantErr {
				t.Errorf("getEC2ConsoleOutput(ctx, fake, %q) error = %v, wantErr %t", instanceID, err, g.wantErr)
			}
			if code := awsup.AWSErrorCode(err); g.wantErrCode != "" && code != g.wantErrCode {
				t.Errorf("getEC2ConsoleOutput(ctx, fake, %q) error code = %q, want %q (err = %v)", instanceID, code, g.wantErrCode, err)
			}
			if string(got) != g.want {
				t.Errorf("getEC2ConsoleOutput(ctx, fake, %q) = %q, want %q", instanceID, got, g.want)
			}
			if g.ec2.latestRequests != g.wantLatestRequests || g.ec2.snapshotRequests != g.wantSnapshotRequests {
				t.Errorf("getEC2ConsoleOutput(ctx, fake, %q) made %d latest and %d snapshot requests, want %d and %d", instanceID, g.ec2.latestRequests, g.ec2.snapshotRequests, g.wantLatestRequests, g.wantSnapshotRequests)
			}
			for _, id := range g.ec2.instanceIDs {
				if id != instanceID {
					t.Errorf("getEC2ConsoleOutput(ctx, fake, %q) sent InstanceId %q", instanceID, id)
				}
			}
		})
	}
}
