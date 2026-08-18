// Copyright 2023 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/internal/limit"
	"github.com/google/go-containerregistry/internal/redact"
	"github.com/google/go-containerregistry/internal/verify"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

const (
	kib           = 1024
	mib           = 1024 * kib
	manifestLimit = 100 * mib
)

// fetcher implements methods for reading from a registry.
type fetcher struct {
	target  resource
	client  *http.Client
	limiter *pullLimiter
}

func makeFetcher(ctx context.Context, target resource, o *options) (*fetcher, error) {
	auth := o.auth
	if o.keychain != nil {
		kauth, err := authn.Resolve(ctx, o.keychain, target)
		if err != nil {
			return nil, err
		}
		auth = kauth
	}

	reg, ok := target.(name.Registry)
	if !ok {
		repo, ok := target.(name.Repository)
		if !ok {
			return nil, fmt.Errorf("unexpected resource: %T", target)
		}
		reg = repo.Registry
	}

	tr, err := transport.NewWithContext(ctx, reg, auth, o.transport, []string{target.Scope(transport.PullScope)})
	if err != nil {
		return nil, err
	}
	return &fetcher{
		target: target,
		client: &http.Client{
			Transport:     tr,
			CheckRedirect: checkRedirectSSRF,
		},
		limiter: o.limiter,
	}, nil
}

// checkRedirectSSRF rejects HTTP redirects that cross from a public host to a
// private or link-local IP literal. This prevents a malicious registry from
// issuing a 302 to a cloud instance metadata service (e.g. 169.254.169.254)
// or another internal network address during blob or manifest downloads.
//
// Same-host redirects and redirects to non-IP hostnames (including DNS names
// that may resolve to private addresses) are allowed. The first redirect in
// the chain uses the original request URL as the "origin host" via
// req.Response.Request, falling back to req.URL when no prior response exists.
func checkRedirectSSRF(req *http.Request, via []*http.Request) error {
	if len(via) == 0 || req.Response == nil {
		return nil
	}
	origHost := via[0].URL.Hostname()
	destHost := req.URL.Hostname()
	if destHost == origHost {
		return nil // same-host redirect is always allowed
	}
	if ip := net.ParseIP(destHost); ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() || ip.IsUnspecified() {
			return fmt.Errorf("SSRF protection: redirect from %q to private/link-local host %q denied", origHost, destHost)
		}
	}
	return nil
}

func (f *fetcher) Do(req *http.Request) (*http.Response, error) {
	return f.client.Do(req)
}

type resource interface {
	Scheme() string
	RegistryStr() string
	Scope(string) string

	authn.Resource
}

// url returns a url.Url for the specified path in the context of this remote image reference.
func (f *fetcher) url(resource, identifier string) url.URL {
	u := url.URL{
		Scheme: f.target.Scheme(),
		Host:   f.target.RegistryStr(),
		// Default path if this is not a repository.
		Path: "/v2/_catalog",
	}
	if repo, ok := f.target.(name.Repository); ok {
		u.Path = fmt.Sprintf("/v2/%s/%s/%s", repo.RepositoryStr(), resource, identifier)
	}
	return u
}

func (f *fetcher) get(ctx context.Context, ref name.Reference, acceptable []types.MediaType, platform v1.Platform) (*Descriptor, error) {
	b, desc, err := f.fetchManifest(ctx, ref, acceptable)
	if err != nil {
		return nil, err
	}
	return &Descriptor{
		ref:        ref,
		ctx:        ctx,
		fetcher:    *f,
		Manifest:   b,
		Descriptor: *desc,
		platform:   platform,
	}, nil
}

func (f *fetcher) fetchManifest(ctx context.Context, ref name.Reference, acceptable []types.MediaType) ([]byte, *v1.Descriptor, error) {
	u := f.url("manifests", ref.Identifier())
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	accept := []string{}
	for _, mt := range acceptable {
		accept = append(accept, string(mt))
	}
	req.Header.Set("Accept", strings.Join(accept, ","))

	resp, err := f.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if err := transport.CheckError(resp, http.StatusOK); err != nil {
		return nil, nil, err
	}

	manifest, err := limit.ReadAll(resp.Body, manifestLimit)
	if err != nil {
		return nil, nil, err
	}

	digest, size, err := v1.SHA256(bytes.NewReader(manifest))
	if err != nil {
		return nil, nil, err
	}

	mediaType := types.MediaType(resp.Header.Get("Content-Type"))
	contentDigest, err := v1.NewHash(resp.Header.Get("Docker-Content-Digest"))
	if err == nil && mediaType == types.DockerManifestSchema1Signed {
		// If we can parse the digest from the header, and it's a signed schema 1
		// manifest, let's use that for the digest to appease older registries.
		digest = contentDigest
	}

	// Validate the digest matches what we asked for, if pulling by digest.
	if dgst, ok := ref.(name.Digest); ok {
		if digest.String() != dgst.DigestStr() {
			return nil, nil, fmt.Errorf("manifest digest: %q does not match requested digest: %q for %q", digest, dgst.DigestStr(), ref)
		}
	}

	var artifactType string
	var annotations map[string]string
	mf, _ := v1.ParseManifest(bytes.NewReader(manifest))
	// Failing to parse as a manifest should just be ignored.
	// The manifest might not be valid, and that's okay.
	if mf != nil {
		// Per the OCI distribution spec, artifactType on the descriptor is
		// set to the manifest's artifactType if present, otherwise it falls
		// back to the config descriptor's mediaType.
		if mf.ArtifactType != "" {
			artifactType = mf.ArtifactType
		} else {
			artifactType = string(mf.Config.MediaType)
		}
		annotations = mf.Annotations
	}

	// Do nothing for tags; I give up.
	//
	// We'd like to validate that the "Docker-Content-Digest" header matches what is returned by the registry,
	// but so many registries implement this incorrectly that it's not worth checking.
	//
	// For reference:
	// https://github.com/GoogleContainerTools/kaniko/issues/298

	// Return all this info since we have to calculate it anyway.
	desc := v1.Descriptor{
		Digest:       digest,
		Size:         size,
		MediaType:    mediaType,
		ArtifactType: artifactType,
		Annotations:  annotations,
	}

	return manifest, &desc, nil
}

func (f *fetcher) headManifest(ctx context.Context, ref name.Reference, acceptable []types.MediaType) (*v1.Descriptor, error) {
	u := f.url("manifests", ref.Identifier())
	req, err := http.NewRequest(http.MethodHead, u.String(), nil)
	if err != nil {
		return nil, err
	}
	accept := []string{}
	for _, mt := range acceptable {
		accept = append(accept, string(mt))
	}
	req.Header.Set("Accept", strings.Join(accept, ","))

	resp, err := f.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := transport.CheckError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	mth := resp.Header.Get("Content-Type")
	if mth == "" {
		return nil, fmt.Errorf("HEAD %s: response did not include Content-Type header", u.String())
	}
	mediaType := types.MediaType(mth)

	size := resp.ContentLength
	if size == -1 {
		return nil, fmt.Errorf("GET %s: response did not include Content-Length header", u.String())
	}

	dh := resp.Header.Get("Docker-Content-Digest")
	if dh == "" {
		return nil, fmt.Errorf("HEAD %s: response did not include Docker-Content-Digest header", u.String())
	}
	digest, err := v1.NewHash(dh)
	if err != nil {
		return nil, err
	}

	// Validate the digest matches what we asked for, if pulling by digest.
	if dgst, ok := ref.(name.Digest); ok {
		if digest.String() != dgst.DigestStr() {
			return nil, fmt.Errorf("manifest digest: %q does not match requested digest: %q for %q", digest, dgst.DigestStr(), ref)
		}
	}

	// Return all this info since we have to calculate it anyway.
	return &v1.Descriptor{
		Digest:    digest,
		Size:      size,
		MediaType: mediaType,
	}, nil
}

func (f *fetcher) fetchBlob(ctx context.Context, size int64, h v1.Hash) (io.ReadCloser, error) {
	u := f.url("blobs", h.String())
	return f.fetchBlobURL(ctx, u, size, h)
}

func (f *fetcher) fetchBlobURL(ctx context.Context, u url.URL, size int64, h v1.Hash) (io.ReadCloser, error) {
	release, err := f.limiter.acquire(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		release()
		return nil, err
	}

	resp, err := f.client.Do(req.WithContext(ctx))
	if err != nil {
		release()
		return nil, redact.Error(err)
	}

	if err := transport.CheckError(resp, http.StatusOK); err != nil {
		resp.Body.Close()
		release()
		return nil, err
	}

	// Do whatever we can.
	// If we have an expected size and Content-Length doesn't match, return an error.
	// If we don't have an expected size and we do have a Content-Length, use Content-Length.
	if hsize := resp.ContentLength; hsize != -1 {
		if size == verify.SizeUnknown {
			size = hsize
		} else if hsize != size {
			resp.Body.Close()
			release()
			return nil, fmt.Errorf("GET %s: Content-Length header %d does not match expected size %d", u.String(), hsize, size)
		}
	}

	rc, err := verify.ReadCloser(resp.Body, size, h)
	if err != nil {
		resp.Body.Close()
		release()
		return nil, err
	}
	return &limitedReadCloser{
		ReadCloser: rc,
		release:    release,
	}, nil
}

func (f *fetcher) headBlob(ctx context.Context, h v1.Hash) (*http.Response, error) {
	u := f.url("blobs", h.String())
	req, err := http.NewRequest(http.MethodHead, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, redact.Error(err)
	}

	if err := transport.CheckError(resp, http.StatusOK); err != nil {
		resp.Body.Close()
		return nil, err
	}

	return resp, nil
}

// validateForeignURL rejects foreign layer URLs that use a disallowed scheme
// or resolve to a private / link-local IP address (SSRF protection). DNS-based
// SSRF is out of scope, matching transport.validateRealmURL.
func validateForeignURL(rawURL string, insecure bool) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parsing foreign layer URL %q: %w", rawURL, err)
	}
	switch u.Scheme {
	case "https":
	case "http":
		if !insecure {
			return fmt.Errorf("foreign layer URL scheme %q not allowed for a secure registry; use https", u.Scheme)
		}
	default:
		return fmt.Errorf("foreign layer URL scheme %q not allowed; must be https (or http for insecure registries)", u.Scheme)
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() || ip.IsUnspecified() {
			return fmt.Errorf("foreign layer URL host %q is a private or link-local address", host)
		}
	}
	return nil
}

// fetchForeignBlobURL fetches a foreign-layer blob, validating every redirect
// destination through validateForeignURL (SSRF protection).
func (f *fetcher) fetchForeignBlobURL(ctx context.Context, u url.URL, size int64, h v1.Hash, insecure bool) (io.ReadCloser, error) {
	safeClient := &http.Client{
		Transport: f.client.Transport,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			return validateForeignURL(req.URL.String(), insecure)
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := safeClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("GET %s: unexpected status %s", u.String(), resp.Status)
	}
	return verify.ReadCloser(resp.Body, size, h)
}

func (f *fetcher) blobExists(ctx context.Context, h v1.Hash) (bool, error) {
	u := f.url("blobs", h.String())
	req, err := http.NewRequest(http.MethodHead, u.String(), nil)
	if err != nil {
		return false, err
	}

	resp, err := f.client.Do(req.WithContext(ctx))
	if err != nil {
		return false, redact.Error(err)
	}
	defer resp.Body.Close()

	if err := transport.CheckError(resp, http.StatusOK, http.StatusNotFound); err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}
