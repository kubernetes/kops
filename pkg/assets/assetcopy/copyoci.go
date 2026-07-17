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

package assetcopy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/util/pkg/vfs"
)

// CopyFileToOCI copies a file from a source file repository to an OCI registry,
// pushing it as a single-layer artifact. The layer blob digest is the sha256 hash
// of the file, so clients can download the blob directly by the file's hash.
type CopyFileToOCI struct {
	Name       string
	SourceFile string
	// TargetRef is the target location, in the form oci://<registry>/<repository>.
	TargetRef  string
	SHA        string
	VFSContext *vfs.VFSContext
	Keychain   authn.Keychain
}

func (e *CopyFileToOCI) Run() error {
	repository := strings.TrimPrefix(e.TargetRef, "oci://")

	// Tag the artifact with the file's hash; this also makes the check for an
	// already-pushed artifact a single HEAD request.
	ref, err := name.NewTag(strings.ToLower(repository) + ":" + e.SHA)
	if err != nil {
		return fmt.Errorf("parsing reference for %q: %w", e.TargetRef, err)
	}

	options := []remote.Option{remote.WithAuthFromKeychain(e.Keychain)}

	if _, err := remote.Head(ref, options...); err == nil {
		klog.Infof("no need to copy file from %v to %v", e.SourceFile, e.TargetRef)
		return nil
	}

	data, err := e.VFSContext.ReadFile(e.SourceFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found %q: %w", e.SourceFile, err)
		}
		return fmt.Errorf("error downloading file %q: %w", e.SourceFile, err)
	}

	digest := sha256.Sum256(data)
	actualSHA := hex.EncodeToString(digest[:])
	if actualSHA != e.SHA {
		return fmt.Errorf("hash mismatch for %q: expected %q, got %q", e.SourceFile, e.SHA, actualSHA)
	}

	layer := static.NewLayer(data, types.MediaType("application/octet-stream"))
	image, err := mutate.AppendLayers(empty.Image, layer)
	if err != nil {
		return fmt.Errorf("building artifact for %q: %w", e.TargetRef, err)
	}

	klog.V(2).Infof("copying bits from %q to %q", e.SourceFile, e.TargetRef)

	if err := remote.Write(ref, image, options...); err != nil {
		return fmt.Errorf("unable to transfer %q to %q: %w", e.SourceFile, e.TargetRef, err)
	}

	return nil
}

// acrKeychain authenticates to Azure Container Registries by exchanging an
// Entra ID token for a registry refresh token.
type acrKeychain struct {
	credential azcore.TokenCredential
}

// NewACRKeychain returns a keychain that authenticates to Azure Container Registries
// with the default Azure credential, falling back to anonymous elsewhere.
func NewACRKeychain() (authn.Keychain, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating an identity: %w", err)
	}
	return &acrKeychain{credential: credential}, nil
}

var _ authn.Keychain = (*acrKeychain)(nil)

func (k *acrKeychain) Resolve(target authn.Resource) (authn.Authenticator, error) {
	ctx := context.TODO()

	registry := target.RegistryStr()
	if !strings.HasSuffix(registry, ".azurecr.io") {
		return authn.Anonymous, nil
	}

	token, err := k.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("getting an identity token: %w", err)
	}

	refreshToken, err := acrExchangeToken(ctx, registry, token.Token)
	if err != nil {
		return nil, fmt.Errorf("exchanging the identity token for a registry refresh token: %w", err)
	}

	// The registry accepts its refresh token as the password for this well-known user.
	return authn.FromConfig(authn.AuthConfig{
		Username: "00000000-0000-0000-0000-000000000000",
		Password: refreshToken,
	}), nil
}

// acrExchangeToken exchanges an Entra ID access token for a registry refresh token.
func acrExchangeToken(ctx context.Context, registry, aadToken string) (string, error) {
	values := url.Values{
		"grant_type":   []string{"access_token"},
		"service":      []string{registry},
		"access_token": []string{aadToken},
	}
	endpoint := fmt.Sprintf("https://%s/oauth2/exchange", registry)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error posting to %q: %w", endpoint, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return "", fmt.Errorf("unexpected response from %q: HTTP %s", endpoint, response.Status)
	}

	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("cannot decode response from %q: %w", endpoint, err)
	}
	if body.RefreshToken == "" {
		return "", fmt.Errorf("response from %q did not contain a refresh token", endpoint)
	}
	return body.RefreshToken, nil
}
