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

package azure

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"k8s.io/kops/upup/pkg/fi/cloudup/azure/azuremetadata"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"k8s.io/kops/pkg/bootstrap"
)

func TestVMLogIDFromResource(t *testing.T) {
	testCases := []struct {
		name       string
		resourceID string
		want       string
	}{
		{
			name:       "vm",
			resourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-1",
			want:       "vm-1",
		},
		{
			name:       "vmss vm",
			resourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/nodes-uksouth-3.cluster/virtualMachines/1",
			want:       "nodes-uksouth-3.cluster/1",
		},
		{
			name:       "fallback",
			resourceID: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nic-1",
			want:       "Microsoft.Network/networkInterfaces/nic-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := arm.ParseResourceID(tc.resourceID)
			if err != nil {
				t.Fatalf("parsing resource ID: %v", err)
			}
			if got := vmLogIDFromResource(res); got != tc.want {
				t.Fatalf("vmLogIDFromResource() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestVerifyToken covers the early rejection paths: wrong prefix (different cloud verifier),
// malformed two-part payload, mismatched subscription/RG, and unparseable PKCS7.
func TestVerifyToken(t *testing.T) {
	matchingResourceID := "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm"
	wrongSubResourceID := "/subscriptions/other/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm"
	wrongRGResourceID := "/subscriptions/sub/resourceGroups/other/providers/Microsoft.Compute/virtualMachines/vm"
	invalidPKCS7 := base64.StdEncoding.EncodeToString([]byte("not-pkcs7"))

	testCases := []struct {
		name          string
		token         string
		wantErr       error // explicit error to compare with ==; nil means "any non-nil error"
		wantErrSubstr string
	}{
		{"wrong prefix", "x-aws-sts something", bootstrap.ErrNotThisVerifier, ""},
		{"missing signature", azuremetadata.AzureAuthenticationTokenPrefix + "no-space-here", nil, ""},
		{"subscription mismatch", azuremetadata.AzureAuthenticationTokenPrefix + wrongSubResourceID + " " + invalidPKCS7, nil, ""},
		{"resource group mismatch", azuremetadata.AzureAuthenticationTokenPrefix + wrongRGResourceID + " " + invalidPKCS7, nil, ""},
		{"unsupported resource type", azuremetadata.AzureAuthenticationTokenPrefix + "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nic-1 " + invalidPKCS7, nil, "unsupported resource type"},
		{"short vmss reaches attestation", azuremetadata.AzureAuthenticationTokenPrefix + "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/nodes/virtualMachines/1 " + invalidPKCS7, nil, "parsing PKCS7 signature"},
		{"legacy vmss reaches attestation", azuremetadata.AzureAuthenticationTokenPrefix + "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/nodes.cluster/virtualMachines/1 " + invalidPKCS7, nil, "parsing PKCS7 signature"},
		{"invalid PKCS7", azuremetadata.AzureAuthenticationTokenPrefix + matchingResourceID + " " + invalidPKCS7, nil, ""},
	}

	v := &azureVerifier{
		client: &client{
			subscriptionID: "sub",
			resourceGroup:  "rg",
		},
		clusterName: "cluster",
		attestation: &attestationVerifier{rootCertPool: x509.NewCertPool()},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := v.VerifyToken(context.TODO(), nil, tc.token, nil)
			if tc.wantErr != nil {
				if err != tc.wantErr {
					t.Errorf("expected %v, got: %v", tc.wantErr, err)
				}
				return
			}
			if err == nil {
				t.Error("expected error")
				return
			}
			if tc.wantErrSubstr != "" && !strings.Contains(err.Error(), tc.wantErrSubstr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErrSubstr, err)
			}
		})
	}
}

func TestVerifyToken_SignedSubscriptionMismatch(t *testing.T) {
	caCert, leafCert, leafKey := testPKI(t)
	rootPool := x509.NewCertPool()
	rootPool.AddCert(caCert)

	body := []byte("test-body")
	now := time.Now().UTC()
	sig := testSignature(t, attestedData{
		VMId:           "test-vm-id",
		SubscriptionId: "other-subscription",
		Nonce:          nonceForBody(body),
		TimeStamp: attestedTimeStamp{
			CreatedOn: now.Add(-time.Minute).Format(attestedDocumentTimeFormat),
			ExpiresOn: now.Add(time.Hour).Format(attestedDocumentTimeFormat),
		},
	}, leafCert, leafKey, caCert)

	v := &azureVerifier{
		client: &client{
			subscriptionID: "sub",
			resourceGroup:  "rg",
		},
		clusterName: "cluster",
		attestation: &attestationVerifier{rootCertPool: rootPool},
	}

	token := azuremetadata.AzureAuthenticationTokenPrefix + "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm " + sig
	_, err := v.VerifyToken(context.Background(), nil, token, body)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "attested subscriptionId") {
		t.Fatalf("expected error containing %q, got %v", "attested subscriptionId", err)
	}
}

func TestExtractNodeIdentity(t *testing.T) {
	caCert, leafCert, leafKey := testPKI(t)
	rootPool := x509.NewCertPool()
	rootPool.AddCert(caCert)
	body := []byte("node identity challenge")
	now := time.Now().UTC()
	signature := testSignature(t, attestedData{
		VMId:           "signed-vm-id",
		SubscriptionId: "sub",
		Nonce:          nonceForBody(body),
		TimeStamp: attestedTimeStamp{
			CreatedOn: now.Add(-time.Minute).Format(attestedDocumentTimeFormat),
			ExpiresOn: now.Add(time.Hour).Format(attestedDocumentTimeFormat),
		},
	}, leafCert, leafKey, caCert)
	attestation := &attestationVerifier{rootCertPool: rootPool}
	data, err := attestation.verifyAttestedDocument(signature, body)
	if err != nil {
		t.Fatalf("verifying signed identity: %v", err)
	}
	for _, test := range []struct {
		name      string
		vmID      string
		tags      map[string]*string
		wantError string
	}{
		{name: "matching identity", vmID: "signed-vm-id", tags: map[string]*string{TagClusterName: to.Ptr("cluster"), InstanceGroupNameTag: to.Ptr("nodes")}},
		{name: "missing cluster", vmID: "signed-vm-id", tags: map[string]*string{InstanceGroupNameTag: to.Ptr("nodes")}, wantError: "does not belong to cluster"},
		{name: "nil cluster", vmID: "signed-vm-id", tags: map[string]*string{TagClusterName: nil, InstanceGroupNameTag: to.Ptr("nodes")}, wantError: "does not belong to cluster"},
		{name: "foreign cluster", vmID: "signed-vm-id", tags: map[string]*string{TagClusterName: to.Ptr("other"), InstanceGroupNameTag: to.Ptr("nodes")}, wantError: "does not belong to cluster"},
		{name: "mismatched VMID", vmID: "different-vm-id", tags: map[string]*string{TagClusterName: to.Ptr("cluster"), InstanceGroupNameTag: to.Ptr("nodes")}, wantError: "attested vmId"},
		{name: "missing instance group", vmID: "signed-vm-id", tags: map[string]*string{TagClusterName: to.Ptr("cluster")}, wantError: "determining IG name"},
	} {
		t.Run(test.name, func(t *testing.T) {
			nodeName, igName, err := extractNodeIdentity(data, "VMSS VM", to.Ptr(test.vmID), &compute.OSProfile{ComputerName: to.Ptr("Nodes000000")}, test.tags, "cluster")
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("got error %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("extracting node identity: %v", err)
			}
			if nodeName != "nodes000000" || igName != "nodes" {
				t.Fatalf("got node %q, instance group %q", nodeName, igName)
			}
		})
	}
}
