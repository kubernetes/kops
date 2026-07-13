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

package linode

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode/linodemetadata"
)

func TestLinodeVerifierVerifyToken(t *testing.T) {
	verifier := &linodeVerifier{
		client: &fakeLinodeVerifierClient{
			instance: &linodego.Instance{
				ID:    101,
				Label: "nodes-us-ord.example.k8s.local-1",
				IPv4:  []net.IP{*ipPtr("203.0.113.15"), *ipPtr("192.168.152.10")},
				Tags:  []string{"kops.k8s.io/instance-group:nodes-us-ord"},
			},
		},
	}

	result, err := verifier.VerifyToken(context.Background(), &http.Request{}, linodemetadata.LinodeAuthenticationTokenPrefix+"101", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.NodeName != "nodes-us-ord.example.k8s.local-1" {
		t.Fatalf("unexpected node name %q", result.NodeName)
	}
	if result.InstanceGroupName != "nodes-us-ord" {
		t.Fatalf("unexpected instance group %q", result.InstanceGroupName)
	}
	if want := net.JoinHostPort("192.168.152.10", strconv.Itoa(wellknownports.NodeupChallenge)); result.ChallengeEndpoint != want {
		t.Fatalf("expected challenge endpoint %q, got %q", want, result.ChallengeEndpoint)
	}
}

func TestLinodeVerifierVerifyTokenWrongPrefix(t *testing.T) {
	verifier := &linodeVerifier{client: &fakeLinodeVerifierClient{}}
	_, err := verifier.VerifyToken(context.Background(), &http.Request{}, "x-other 1", nil)
	if !errors.Is(err, bootstrap.ErrNotThisVerifier) {
		t.Fatalf("expected ErrNotThisVerifier, got %v", err)
	}
}

func TestLinodeVerifierVerifyTokenNoPrivateIP(t *testing.T) {
	verifier := &linodeVerifier{
		client: &fakeLinodeVerifierClient{
			instance: &linodego.Instance{ID: 101, Label: "node-1", IPv4: []net.IP{*ipPtr("203.0.113.15")}},
		},
	}

	_, err := verifier.VerifyToken(context.Background(), &http.Request{}, linodemetadata.LinodeAuthenticationTokenPrefix+"101", nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestInstanceGroupNameFromTags(t *testing.T) {
	tags := []string{"foo:bar", "kops.k8s.io/instance-group:nodes-us-ord"}
	if got := instanceGroupNameFromTags(tags); got != "nodes-us-ord" {
		t.Fatalf("unexpected instance group %q", got)
	}
}

type fakeLinodeVerifierClient struct {
	instance *linodego.Instance
	err      error
}

func (f *fakeLinodeVerifierClient) GetInstance(ctx context.Context, linodeID int) (*linodego.Instance, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.instance, nil
}

func ipPtr(s string) *net.IP {
	ip := net.ParseIP(s)
	return &ip
}
