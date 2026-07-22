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

package gcetpmverifier

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kops/pkg/nodeidentity/clusterapi/capimanager"
	gcetpm "k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testProject     = "testproject"
	testZone        = "us-test1-a"
	testRegion      = "us-test1"
	testClusterName = "cluster.example.com"
	testInstance    = "node1"
	testInstanceID  = uint64(1234567890)
	testMIG         = "nodes-us-test1-a"
	testTemplate    = "nodes-us-test1-a-template"
)

// fakeComputeAPI serves the subset of the GCE compute API used by the TPM verifier.
type fakeComputeAPI struct {
	instance         *compute.Instance
	signingKey       *rsa.PrivateKey
	migExists        bool
	managedInstances []*compute.ManagedInstance
	instanceTemplate *compute.InstanceTemplate

	// requests records the paths of all requests served
	requests []string
}

func (f *fakeComputeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.requests = append(f.requests, r.URL.Path)

	var response any

	switch r.URL.Path {
	case "/projects/" + testProject + "/zones/" + testZone + "/instances/" + testInstance:
		response = f.instance
	case "/projects/" + testProject + "/zones/" + testZone + "/instances/" + testInstance + "/getShieldedInstanceIdentity":
		der, err := x509.MarshalPKIXPublicKey(&f.signingKey.PublicKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ekPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
		response = &compute.ShieldedInstanceIdentity{
			SigningKey: &compute.ShieldedInstanceIdentityEntry{
				EkPub: string(ekPub),
			},
		}
	case "/projects/" + testProject + "/zones/" + testZone + "/instanceGroupManagers/" + testMIG + "/listManagedInstances":
		if !f.migExists {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		response = &compute.InstanceGroupManagersListManagedInstancesResponse{
			ManagedInstances: f.managedInstances,
		}
	case "/projects/" + testProject + "/global/instanceTemplates/" + testTemplate:
		response = f.instanceTemplate
	default:
		http.Error(w, "unexpected request "+r.URL.Path, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// defaultFakeComputeAPI returns a fake where the instance is a genuine member of a kOps-managed
// MIG whose instance template carries the expected metadata.
func defaultFakeComputeAPI(t *testing.T) *fakeComputeAPI {
	t.Helper()

	signingKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}

	return &fakeComputeAPI{
		signingKey: signingKey,
		instance: &compute.Instance{
			Name:     testInstance,
			Id:       testInstanceID,
			SelfLink: "https://www.googleapis.com/compute/v1/projects/" + testProject + "/zones/" + testZone + "/instances/" + testInstance,
			Zone:     "https://www.googleapis.com/compute/v1/projects/" + testProject + "/zones/" + testZone,
			Metadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					{Key: "created-by", Value: ptr.To("https://www.googleapis.com/compute/v1/projects/" + testProject + "/zones/" + testZone + "/instanceGroupManagers/" + testMIG)},
					// Live instance metadata is settable by whoever created the instance; the
					// verifier must not trust these values.
					{Key: "cluster-name", Value: ptr.To("spoofed.example.com")},
					{Key: "kops-k8s-io-instance-group-name", Value: ptr.To("spoofed-ig")},
				},
			},
			NetworkInterfaces: []*compute.NetworkInterface{
				{NetworkIP: "10.0.0.1"},
			},
		},
		migExists: true,
		managedInstances: []*compute.ManagedInstance{
			{
				Id: testInstanceID,
				Version: &compute.ManagedInstanceVersion{
					InstanceTemplate: "https://www.googleapis.com/compute/v1/projects/" + testProject + "/global/instanceTemplates/" + testTemplate,
				},
			},
		},
		instanceTemplate: &compute.InstanceTemplate{
			Name: testTemplate,
			Properties: &compute.InstanceProperties{
				Metadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{Key: "cluster-name", Value: ptr.To(testClusterName)},
						{Key: "kops-k8s-io-instance-group-name", Value: ptr.To("nodes")},
					},
				},
			},
		},
	}
}

func buildAuthToken(t *testing.T, signingKey *rsa.PrivateKey, body []byte) string {
	t.Helper()

	requestHash := sha256.Sum256(body)
	tokenData := gcetpm.AuthTokenData{
		GCPProjectID: testProject,
		Zone:         testZone,
		Instance:     testInstance,
		RequestHash:  requestHash[:],
		Timestamp:    time.Now().Unix(),
		Audience:     gcetpm.AudienceNodeAuthentication,
	}
	data, err := json.Marshal(&tokenData)
	if err != nil {
		t.Fatalf("marshalling token data: %v", err)
	}

	dataHash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, signingKey, crypto.SHA256, dataHash[:])
	if err != nil {
		t.Fatalf("signing token data: %v", err)
	}

	token, err := json.Marshal(&gcetpm.AuthToken{
		Signature: signature,
		Data:      data,
	})
	if err != nil {
		t.Fatalf("marshalling token: %v", err)
	}

	return gcetpm.GCETPMAuthenticationTokenPrefix + base64.StdEncoding.EncodeToString(token)
}

func newTestVerifier(t *testing.T, fake *fakeComputeAPI) *tpmVerifier {
	t.Helper()

	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	computeClient, err := compute.NewService(context.Background(), option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("building compute client: %v", err)
	}

	return &tpmVerifier{
		opt: gcetpm.TPMVerifierOptions{
			ProjectID:   testProject,
			Region:      testRegion,
			ClusterName: testClusterName,
			MaxTimeSkew: 300,
		},
		computeClient: computeClient,
	}
}

func TestVerifyToken(t *testing.T) {
	body := []byte("test-request-body")

	tests := []struct {
		name              string
		setup             func(fake *fakeComputeAPI)
		expectedError     string
		expectedNodeName  string
		expectedGroupName string
	}{
		{
			name:              "valid MIG member",
			setup:             func(fake *fakeComputeAPI) {},
			expectedNodeName:  testInstance,
			expectedGroupName: "nodes",
		},
		{
			name: "instance not a member of the MIG",
			setup: func(fake *fakeComputeAPI) {
				// The instance spoofs created-by (and cluster-name) in its metadata, but the MIG
				// does not actually manage it.
				fake.managedInstances = nil
			},
			expectedError: "not managed by mig",
		},
		{
			name: "instance without created-by metadata",
			setup: func(fake *fakeComputeAPI) {
				fake.instance.Metadata.Items = fake.instance.Metadata.Items[1:]
			},
			expectedError: "cannot find owner",
		},
		{
			name: "created-by references MIG that does not exist",
			setup: func(fake *fakeComputeAPI) {
				fake.migExists = false
			},
			expectedError: "error fetching GCE managed instance group",
		},
		{
			name: "instance template belongs to another cluster",
			setup: func(fake *fakeComputeAPI) {
				fake.instanceTemplate.Properties.Metadata.Items[0].Value = ptr.To("other.example.com")
			},
			expectedError: "clusterName does not match expected",
		},
		{
			name: "instance template without cluster name",
			setup: func(fake *fakeComputeAPI) {
				fake.instanceTemplate.Properties.Metadata.Items = fake.instanceTemplate.Properties.Metadata.Items[1:]
			},
			expectedError: "could not determine cluster",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			fake := defaultFakeComputeAPI(t)
			tc.setup(fake)
			verifier := newTestVerifier(t, fake)

			authToken := buildAuthToken(t, fake.signingKey, body)
			request := httptest.NewRequest(http.MethodPost, "/bootstrap", nil)

			result, err := verifier.VerifyToken(ctx, request, authToken, body)
			if tc.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got result %+v", tc.expectedError, result)
				}
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Fatalf("expected error containing %q, got %q", tc.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.NodeName != tc.expectedNodeName {
				t.Errorf("expected NodeName %q, got %q", tc.expectedNodeName, result.NodeName)
			}
			if result.InstanceGroupName != tc.expectedGroupName {
				t.Errorf("expected InstanceGroupName %q, got %q", tc.expectedGroupName, result.InstanceGroupName)
			}
		})
	}
}

func TestVerifyTokenRejectsBadSignature(t *testing.T) {
	ctx := context.Background()
	body := []byte("test-request-body")

	fake := defaultFakeComputeAPI(t)
	verifier := newTestVerifier(t, fake)

	// Sign with a key that does not match the instance's TPM endorsement key.
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}
	authToken := buildAuthToken(t, otherKey, body)
	request := httptest.NewRequest(http.MethodPost, "/bootstrap", nil)

	result, err := verifier.VerifyToken(ctx, request, authToken, body)
	if err == nil {
		t.Fatalf("expected signature verification error, got result %+v", result)
	}
	if !strings.Contains(err.Error(), "failed to verify claim signature") {
		t.Fatalf("expected signature verification error, got %q", err.Error())
	}

	// Requests that fail signature verification must not trigger authorization lookups.
	for _, path := range fake.requests {
		if strings.Contains(path, "instanceGroupManagers") || strings.Contains(path, "instanceTemplates") {
			t.Errorf("request with invalid signature reached authorization endpoint %s", path)
		}
	}
}

// fakeKubeClient implements the parts of client.Client used by capimanager.
type fakeKubeClient struct {
	client.Client
	machines []unstructured.Unstructured
}

func (f *fakeKubeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	ul, ok := list.(*unstructured.UnstructuredList)
	if !ok {
		return fmt.Errorf("unexpected list type %T", list)
	}
	ul.Items = f.machines
	return nil
}

func capiMachineObject(providerID, clusterName string) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "cluster.x-k8s.io/v1beta1",
		"kind":       "Machine",
		"metadata": map[string]any{
			"name":      "machine-1",
			"namespace": "default",
		},
		"spec": map[string]any{
			"providerID":  providerID,
			"clusterName": clusterName,
		},
	}}
}

func TestVerifyTokenCAPI(t *testing.T) {
	body := []byte("test-request-body")
	providerID := "gce://" + testProject + "/" + testZone + "/" + testInstance

	tests := []struct {
		name               string
		machineClusterName string
		expectedError      string
	}{
		{
			name: "machine in this cluster",
			// The CAPI cluster name is the kOps cluster name escaped for GCE
			machineClusterName: "cluster-example-com",
		},
		{
			// A Machine from another cluster must not match; without a Machine the instance
			// falls back to the MIG check, which fails as CAPI instances are not MIG members.
			name:               "machine in another cluster",
			machineClusterName: "other-example-com",
			expectedError:      "cannot find owner",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			fake := defaultFakeComputeAPI(t)
			fake.instance.Labels = map[string]string{"capg-role": "node"}
			// CAPI instances are not members of a kOps MIG and have no created-by metadata.
			fake.instance.Metadata.Items = nil
			fake.migExists = false
			fake.managedInstances = nil

			verifier := newTestVerifier(t, fake)
			verifier.capiManager = capimanager.NewManager(&fakeKubeClient{
				machines: []unstructured.Unstructured{capiMachineObject(providerID, tc.machineClusterName)},
			})

			authToken := buildAuthToken(t, fake.signingKey, body)
			request := httptest.NewRequest(http.MethodPost, "/bootstrap", nil)

			result, err := verifier.VerifyToken(ctx, request, authToken, body)
			if tc.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got result %+v", tc.expectedError, result)
				}
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Fatalf("expected error containing %q, got %q", tc.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.CAPIMachine == nil {
				t.Error("expected CAPIMachine to be set")
			}
			if result.InstanceGroupName != "" {
				t.Errorf("expected empty InstanceGroupName, got %q", result.InstanceGroupName)
			}
		})
	}
}
