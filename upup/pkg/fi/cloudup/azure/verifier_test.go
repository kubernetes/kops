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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
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

func setTestSystemCertPool(t *testing.T, pool *x509.CertPool) {
	t.Helper()

	cachedSystemMu.Lock()
	previous := cachedSystemCertPool
	cachedSystemCertPool = pool
	cachedSystemMu.Unlock()

	t.Cleanup(func() {
		cachedSystemMu.Lock()
		cachedSystemCertPool = previous
		cachedSystemMu.Unlock()
	})
}

// TestVerifyToken covers the early rejection paths: wrong prefix (different
// cloud verifier), malformed two-part payload, mismatched subscription/RG,
// and unparseable PKCS7.
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
		{"missing signature", AzureAuthenticationTokenPrefix + "no-space-here", nil, ""},
		{"subscription mismatch", AzureAuthenticationTokenPrefix + wrongSubResourceID + " " + invalidPKCS7, nil, ""},
		{"resource group mismatch", AzureAuthenticationTokenPrefix + wrongRGResourceID + " " + invalidPKCS7, nil, ""},
		{"unsupported resource type", AzureAuthenticationTokenPrefix + "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/networkInterfaces/nic-1 " + invalidPKCS7, nil, "unsupported resource type"},
		{"vmss cluster mismatch", AzureAuthenticationTokenPrefix + "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachineScaleSets/nodes.other-cluster/virtualMachines/1 " + invalidPKCS7, nil, "does not match cluster name"},
		{"invalid PKCS7", AzureAuthenticationTokenPrefix + matchingResourceID + " " + invalidPKCS7, nil, ""},
	}

	v := &azureVerifier{
		client: &client{
			subscriptionID: "sub",
			resourceGroup:  "rg",
		},
		clusterName: "cluster",
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
	caCert, _, leafCert, leafKey := testPKI(t)
	rootPool := x509.NewCertPool()
	rootPool.AddCert(caCert)
	setTestSystemCertPool(t, rootPool)

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
	}

	token := AzureAuthenticationTokenPrefix + "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm " + sig
	_, err := v.VerifyToken(context.Background(), nil, token, body)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "attested subscriptionId") {
		t.Fatalf("expected error containing %q, got %v", "attested subscriptionId", err)
	}
}
