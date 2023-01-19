/*
Copyright 2021 The Kubernetes Authors.

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
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/nodeidentity/gce"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcemetadata"
	gcetpm "k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm"
)

type tpmVerifier struct {
	opt gcetpm.TPMVerifierOptions

	computeClient *compute.Service
}

// NewTPMVerifier constructs a new TPM verifier for GCE.
func NewTPMVerifier(opt *gcetpm.TPMVerifierOptions) (bootstrap.Verifier, error) {
	ctx := context.Background()

	computeClient, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %w", err)
	}

	return &tpmVerifier{
		opt:           *opt,
		computeClient: computeClient,
	}, nil
}

var _ bootstrap.Verifier = &tpmVerifier{}

func (v *tpmVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, authToken string, body []byte, useInstanceIDForNodeName bool) (*bootstrap.VerifyResult, error) {
	// Reminder: we shouldn't trust any data we get from the client until we've checked the signature (and even then...)
	// Thankfully the GCE SDK does seem to escape the parameters correctly, for example.

	if !strings.HasPrefix(authToken, gcetpm.GCETPMAuthenticationTokenPrefix) {
		return nil, fmt.Errorf("incorrect authorization type")
	}
	authToken = strings.TrimPrefix(authToken, gcetpm.GCETPMAuthenticationTokenPrefix)

	tokenBytes, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return nil, fmt.Errorf("decoding authorization token: %w", err)
	}

	token := &gcetpm.AuthToken{}
	if err = json.Unmarshal(tokenBytes, token); err != nil {
		return nil, fmt.Errorf("unmarshalling authorization token: %w", err)
	}

	tokenData := gcetpm.AuthTokenData{}
	if err := json.Unmarshal(token.Data, &tokenData); err != nil {
		return nil, fmt.Errorf("unmarshalling authorization token data: %w", err)
	}

	// Guard against replay attacks
	if tokenData.Audience != gcetpm.AudienceNodeAuthentication {
		return nil, fmt.Errorf("incorrect Audience")
	}
	timeSkew := math.Abs(time.Since(time.Unix(tokenData.Timestamp, 0)).Seconds())
	if timeSkew > float64(v.opt.MaxTimeSkew) {
		return nil, fmt.Errorf("incorrect Timestamp %v", tokenData.Timestamp)
	}

	// Verify the token has signed the body content.
	requestHash := sha256.Sum256(body)
	if !bytes.Equal(requestHash[:], tokenData.RequestHash) {
		return nil, fmt.Errorf("incorrect RequestHash")
	}

	// Some basic validation to avoid requesting invalid instances.
	if tokenData.GCPProjectID == "" {
		return nil, fmt.Errorf("gcpProjectID is required")
	}
	if tokenData.Zone == "" {
		return nil, fmt.Errorf("zone is required")
	}
	if tokenData.Instance == "" {
		return nil, fmt.Errorf("instance is required")
	}

	// Verify node is in our cluster
	if tokenData.GCPProjectID != v.opt.ProjectID {
		return nil, fmt.Errorf("projectID does not match expected: got %q, want %q", tokenData.GCPProjectID, v.opt.ProjectID)
	}

	instance, err := v.computeClient.Instances.Get(tokenData.GCPProjectID, tokenData.Zone, tokenData.Instance).Do()
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("unable to find instance in compute API: %w", err)
		}
		return nil, fmt.Errorf("error fetching instance from compute API: %w", err)
	}

	if !strings.HasPrefix(lastComponent(instance.Zone), v.opt.Region+"-") {
		return nil, fmt.Errorf("instance was in zone %q, expected region %q", instance.Zone, v.opt.Region)
	}

	clusterName := ""
	instanceGroupName := ""
	for _, item := range instance.Metadata.Items {
		switch item.Key {
		case gce.MetadataKeyInstanceGroupName:
			instanceGroupName = fi.ValueOf(item.Value)
		case gcemetadata.MetadataKeyClusterName:
			clusterName = fi.ValueOf(item.Value)
		}
	}

	if clusterName == "" {
		return nil, fmt.Errorf("could not determine cluster for instance %s", instance.SelfLink)
	}

	if clusterName != v.opt.ClusterName {
		return nil, fmt.Errorf("clusterName does not match expected: got %q, want %q", clusterName, v.opt.ClusterName)
	}
	if instanceGroupName == "" {
		return nil, fmt.Errorf("could not determine instance group for instance %s", instance.SelfLink)
	}

	// Verify the token has a valid GCE TPM signature.
	{
		// Note - we might be able to avoid this call by including the attestation certificate (signed by GCE) in the claim.
		tpmSigningKey, err := v.getTPMSigningKey(ctx, &tokenData)
		if err != nil {
			return nil, err
		}

		if !verifySignature(tpmSigningKey, token.Data, token.Signature) {
			return nil, fmt.Errorf("failed to verify claim signature for node: %w", err)
		}
	}

	sans, err := GetInstanceCertificateAlternateNames(instance)
	if err != nil {
		return nil, err
	}

	result := &bootstrap.VerifyResult{
		NodeName:          instance.Name,
		InstanceGroupName: instanceGroupName,
		CertificateNames:  sans,
	}

	return result, nil
}

func (v *tpmVerifier) getTPMSigningKey(ctx context.Context, data *gcetpm.AuthTokenData) (*rsa.PublicKey, error) {
	response, err := v.computeClient.Instances.GetShieldedInstanceIdentity(data.GCPProjectID, data.Zone, data.Instance).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get shield instance identity: %w", err)
	}

	if response.SigningKey == nil {
		return nil, fmt.Errorf("instance doesn't have a signing key in ShieldedVmIdentity")
	}

	block, _ := pem.Decode([]byte(response.SigningKey.EkPub))
	if block == nil {
		return nil, fmt.Errorf("failed parsing PEM block from EkPub %q", response.SigningKey.EkPub)
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing EK public key: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("EK public key is %T, expected *rsa.PublickKey", pub)
	}
	return rsaPub, nil
}

// GetInstanceCertificateAlternateNames returns the instance hostname and addresses that should go into certificates.
// The first value is the node name and any additional values are IP addresses.
func GetInstanceCertificateAlternateNames(instance *compute.Instance) ([]string, error) {
	var sans []string

	for _, iface := range instance.NetworkInterfaces {
		if iface.NetworkIP != "" {
			sans = append(sans, iface.NetworkIP)
		}
		if iface.Ipv6Address != "" {
			sans = append(sans, iface.Ipv6Address)
		}
		// We only use data for the first interface, and only the first IP
		if len(sans) > 0 {
			break
		}
	}

	return sans, nil
}

func isNotFound(err error) bool {
	gerr, ok := err.(*googleapi.Error)
	return ok && gerr.Code == http.StatusNotFound
}

// lastComponent returns the last component of a URL, i.e. anything after the last slash
// If there is no slash, returns the whole string
func lastComponent(s string) string {
	lastSlash := strings.LastIndex(s, "/")
	if lastSlash != -1 {
		s = s[lastSlash+1:]
	}
	return s
}

func verifySignature(signingKey *rsa.PublicKey, payload []byte, signature []byte) bool {
	attestHash := sha256.Sum256(payload)
	if err := rsa.VerifyPKCS1v15(signingKey, crypto.SHA256, attestHash[:], signature); err != nil {
		return false
	}

	return true
}
