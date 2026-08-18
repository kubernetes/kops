// Copyright 2018 Google LLC All Rights Reserved.
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

package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/internal/limit"
	"github.com/google/go-containerregistry/internal/redact"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/internal/authchallenge"
)

// maxTokenBodySize limits bearer token response body reads to prevent OOM
// when a token endpoint returns an unexpectedly large body.
const maxTokenBodySize = 64 * 1024 // 64 KiB

type Token struct {
	Token        string `json:"token"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Exchange requests a registry Token with the given scopes.
func Exchange(ctx context.Context, reg name.Registry, auth authn.Authenticator, t http.RoundTripper, scopes []string, pr *Challenge) (*Token, error) {
	if strings.ToLower(pr.Scheme) != "bearer" {
		// TODO: Pretend token for basic?
		return nil, fmt.Errorf("challenge scheme %q is not bearer", pr.Scheme)
	}
	bt, err := fromChallenge(reg, auth, t, pr, scopes...)
	if err != nil {
		return nil, err
	}
	authcfg, err := authn.Authorization(ctx, auth)
	if err != nil {
		return nil, err
	}
	tok, err := bt.Refresh(ctx, authcfg)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// FromToken returns a transport given a Challenge + Token.
func FromToken(reg name.Registry, auth authn.Authenticator, t http.RoundTripper, pr *Challenge, tok *Token) (http.RoundTripper, error) {
	if strings.ToLower(pr.Scheme) != "bearer" {
		return &Wrapper{&basicTransport{inner: t, auth: auth, target: reg.RegistryStr()}}, nil
	}
	bt, err := fromChallenge(reg, auth, t, pr)
	if err != nil {
		return nil, err
	}
	if tok.Token != "" {
		bt.bearer.RegistryToken = tok.Token
	}
	return &Wrapper{bt}, nil
}

func fromChallenge(reg name.Registry, auth authn.Authenticator, t http.RoundTripper, pr *Challenge, scopes ...string) (*bearerTransport, error) {
	// We require the realm, which tells us where to send our Basic auth to turn it into Bearer auth.
	realm, ok := pr.Parameters["realm"]
	if !ok {
		return nil, fmt.Errorf("malformed www-authenticate, missing realm: %v", pr.Parameters)
	}
	// Validate the realm URL before storing it. A malicious or compromised
	// registry can supply a realm pointing at an internal service or cloud
	// metadata endpoint (e.g. 169.254.169.254), causing SSRF when the client
	// subsequently fetches a token.
	if err := validateRealmURL(realm, reg.RegistryStr(), pr.Insecure); err != nil {
		return nil, fmt.Errorf("invalid realm in www-authenticate: %w", err)
	}
	service := pr.Parameters["service"]
	scheme := "https"
	if pr.Insecure {
		scheme = "http"
	}
	return &bearerTransport{
		inner:    t,
		basic:    auth,
		realm:    realm,
		registry: reg,
		service:  service,
		scopes:   scopes,
		scheme:   scheme,
	}, nil
}

// realmRedirectCheck mimics the default http.Client redirect policy but also
// validates each redirect URL with validateRealmURL.
func realmRedirectCheck(registryHost string, insecure bool) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		if err := validateRealmURL(req.URL.String(), registryHost, insecure); err != nil {
			return fmt.Errorf("refusing token-server redirect to %q: %w", req.URL, err)
		}
		return nil
	}
}

// validateRealmURL returns an error if the realm URL uses a disallowed scheme
// or resolves to a private / link-local IP address. Realm URLs matching the
// registry host:port are always allowed. See #2258.
func validateRealmURL(realm, registryHost string, insecure bool) error {
	u, err := url.Parse(realm)
	if err != nil {
		return fmt.Errorf("parsing realm %q: %w", realm, err)
	}
	switch u.Scheme {
	case "https":
		// always allowed
	case "http":
		if !insecure {
			return fmt.Errorf("realm scheme %q not allowed for a secure registry; use https", u.Scheme)
		}
	default:
		return fmt.Errorf("realm scheme %q not allowed; must be https (or http for insecure registries)", u.Scheme)
	}
	// Always allow realms matching the registry host:port.
	if registryHost != "" && u.Host == registryHost {
		return nil
	}
	// Reject IP literals that resolve to private or link-local ranges.
	// This blocks direct references to RFC 1918 addresses, loopback, and
	// link-local ranges including the cloud instance metadata service
	// (169.254.169.254 / fd00:ec2::254).  DNS-based SSRF is out of scope
	// here; callers should apply network-level controls if needed.
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() || ip.IsUnspecified() {
			return fmt.Errorf("realm host %q is a private or link-local address", host)
		}
	}
	return nil
}

type bearerTransport struct {
	mx sync.RWMutex
	// Wrapped by bearerTransport.
	inner http.RoundTripper
	// Basic credentials that we exchange for bearer tokens.
	basic authn.Authenticator
	// Holds the bearer response from the token service.
	bearer authn.AuthConfig
	// Registry to which we send bearer tokens.
	registry name.Registry
	// See https://tools.ietf.org/html/rfc6750#section-3
	realm string
	// See https://docs.docker.com/registry/spec/auth/token/
	service string
	scopes  []string
	// Scheme we should use, determined by ping response.
	scheme string
}

var _ http.RoundTripper = (*bearerTransport)(nil)

var portMap = map[string]string{
	"http":  "80",
	"https": "443",
}

func stringSet(ss []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, s := range ss {
		set[s] = struct{}{}
	}
	return set
}

// RoundTrip implements http.RoundTripper
func (bt *bearerTransport) RoundTrip(in *http.Request) (*http.Response, error) {
	sendRequest := func() (*http.Response, error) {
		// http.Client handles redirects at a layer above the http.RoundTripper
		// abstraction, so to avoid forwarding Authorization headers to places
		// we are redirected, only set it when the authorization header matches
		// the registry with which we are interacting.
		// In case of redirect http.Client can use an empty Host, check URL too.
		if matchesHost(bt.registry.RegistryStr(), in, bt.scheme) {
			bt.mx.RLock()
			localToken := bt.bearer.RegistryToken
			bt.mx.RUnlock()
			hdr := fmt.Sprintf("Bearer %s", localToken)
			in.Header.Set("Authorization", hdr)
		}
		return bt.inner.RoundTrip(in)
	}

	res, err := sendRequest()
	if err != nil {
		return nil, err
	}

	// If we hit a WWW-Authenticate challenge, it might be due to expired tokens or insufficient scope.
	if challenges := authchallenge.ResponseChallenges(res); len(challenges) != 0 {
		// close out old response, since we will not return it.
		res.Body.Close()

		// For cross-host challenges (the request was redirected to another host),
		// never mutate bt's shared state: accumulating this host's scope into
		// bt.scopes or refreshing bt's token from bt.realm would pollute future
		// same-host requests with a scope/token that belongs to a request we
		// only ever intended to send to the redirected host. Instead, attempt a
		// fresh per-host token exchange scoped entirely to this one request.
		if !matchesHost(bt.registry.RegistryStr(), in, bt.scheme) {
			return bt.handleCrossHostChallenge(in, challenges)
		}

		newScopes := []string{}
		bt.mx.Lock()
		got := stringSet(bt.scopes)
		for _, wac := range challenges {
			// TODO(jonjohnsonjr): Should we also update "realm" or "service"?
			if want, ok := wac.Parameters["scope"]; ok {
				// Add any scopes that we don't already request.
				if _, ok := got[want]; !ok {
					newScopes = append(newScopes, want)
				}
			}
		}

		// Some registries seem to only look at the first scope parameter during a token exchange.
		// If a request fails because it's missing a scope, we should put those at the beginning,
		// otherwise the registry might just ignore it :/
		newScopes = append(newScopes, bt.scopes...)
		bt.scopes = newScopes
		bt.mx.Unlock()

		// TODO(jonjohnsonjr): Teach transport.Error about "error" and "error_description" from challenge.

		// Retry the request to attempt to get a valid token.
		if err = bt.refresh(in.Context()); err != nil {
			return nil, err
		}

		bt.mx.RLock()
		tok := bt.bearer.RegistryToken
		bt.mx.RUnlock()
		in.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok))
		return bt.inner.RoundTrip(in)
	}

	return res, err
}

// handleCrossHostChallenge performs a per-host bearer token exchange when
// the request has been redirected to a host different from bt.registry. It
// parses the redirected host's own WWW-Authenticate Bearer challenge, fetches
// a token from that host's realm using anonymous auth (the original registry's
// credentials are never forwarded cross-host), and applies the token only to
// this request.
//
// If no usable Bearer challenge is present, or the token exchange fails, the
// request is retried without an Authorization header.
func (bt *bearerTransport) handleCrossHostChallenge(in *http.Request, challenges []authchallenge.Challenge) (*http.Response, error) {
	for _, wac := range challenges {
		if strings.ToLower(wac.Scheme) != "bearer" {
			continue
		}
		if _, ok := wac.Parameters["realm"]; !ok {
			continue
		}

		redirectedReg, err := name.NewRegistry(in.URL.Host, name.WeakValidation)
		if err != nil {
			logs.Warn.Printf("cross-host redirect: invalid host %q: %v", in.URL.Host, err)
			continue
		}

		// Build a transport.Challenge from the redirected host's WWW-Authenticate
		// parameters. fromChallenge validates the realm URL (SSRF guard).
		pr := &Challenge{
			Scheme:     wac.Scheme,
			Parameters: wac.Parameters,
			Insecure:   in.URL.Scheme == "http",
		}
		scope := wac.Parameters["scope"]
		// TODO: use a keychain to resolve credentials for the redirected host so
		// that token endpoints requiring auth are also supported. For now, anonymous
		// auth covers the common case where the redirected host's token endpoint is
		// public. See https://github.com/google/go-containerregistry/issues/2359.
		tmpBt, err := fromChallenge(redirectedReg, authn.Anonymous, bt.inner, pr, scope)
		if err != nil {
			logs.Warn.Printf("cross-host bearer challenge setup for %q failed: %v", in.URL.Host, err)
			continue
		}

		if err := tmpBt.refresh(in.Context()); err != nil {
			logs.Warn.Printf("cross-host bearer exchange for %q failed: %v", in.URL.Host, err)
			continue
		}

		tmpBt.mx.RLock()
		tok := tmpBt.bearer.RegistryToken
		tmpBt.mx.RUnlock()

		in.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok))
		return bt.inner.RoundTrip(in)
	}

	// No usable bearer challenge; retry without credentials.
	return bt.inner.RoundTrip(in)
}

// It's unclear which authentication flow to use based purely on the protocol,
// so we rely on heuristics and fallbacks to support as many registries as possible.
// The basic token exchange is attempted first, falling back to the oauth flow.
// If the IdentityToken is set, this indicates that we should start with the oauth flow.
func (bt *bearerTransport) refresh(ctx context.Context) error {
	auth, err := authn.Authorization(ctx, bt.basic)
	if err != nil {
		return err
	}

	if auth.RegistryToken != "" {
		bt.mx.Lock()
		bt.bearer.RegistryToken = auth.RegistryToken
		bt.mx.Unlock()
		return nil
	}

	response, err := bt.Refresh(ctx, auth)
	if err != nil {
		return err
	}

	// Some registries set access_token instead of token. See #54.
	if response.AccessToken != "" {
		response.Token = response.AccessToken
	}

	// Find a token to turn into a Bearer authenticator
	if response.Token != "" {
		bt.mx.Lock()
		bt.bearer.RegistryToken = response.Token
		bt.mx.Unlock()
	}

	// If we obtained a refresh token from the oauth flow, use that for refresh() now.
	if response.RefreshToken != "" {
		bt.basic = authn.FromConfig(authn.AuthConfig{
			IdentityToken: response.RefreshToken,
		})
	}

	return nil
}

func (bt *bearerTransport) Refresh(ctx context.Context, auth *authn.AuthConfig) (*Token, error) {
	var (
		content []byte
		err     error
	)
	if auth.IdentityToken != "" {
		// If the secret being stored is an identity token,
		// the Username should be set to <token>, which indicates
		// we are using an oauth flow.
		content, err = bt.refreshOauth(ctx)
		var terr *Error
		if errors.As(err, &terr) && terr.StatusCode == http.StatusNotFound {
			// Note: Not all token servers implement oauth2.
			// If the request to the endpoint returns 404 using the HTTP POST method,
			// refer to Token Documentation for using the HTTP GET method supported by all token servers.
			content, err = bt.refreshBasic(ctx)
		}
	} else {
		content, err = bt.refreshBasic(ctx)
	}
	if err != nil {
		return nil, err
	}

	var response Token
	if err := json.Unmarshal(content, &response); err != nil {
		return nil, err
	}

	if response.Token == "" && response.AccessToken == "" {
		return &response, fmt.Errorf("no token in bearer response:\n%s", content)
	}

	return &response, nil
}

func matchesHost(host string, in *http.Request, scheme string) bool {
	canonicalHeaderHost := canonicalAddress(in.Host, scheme)
	canonicalURLHost := canonicalAddress(in.URL.Host, scheme)
	canonicalRegistryHost := canonicalAddress(host, scheme)
	return canonicalHeaderHost == canonicalRegistryHost || canonicalURLHost == canonicalRegistryHost
}

func canonicalAddress(host, scheme string) (address string) {
	// The host may be any one of:
	// - hostname
	// - hostname:port
	// - ipv4
	// - ipv4:port
	// - ipv6
	// - [ipv6]:port
	// As net.SplitHostPort returns an error if the host does not contain a port, we should only attempt
	// to call it when we know that the address contains a port
	if strings.Count(host, ":") == 1 || (strings.Count(host, ":") >= 2 && strings.Contains(host, "]:")) {
		hostname, port, err := net.SplitHostPort(host)
		if err != nil {
			return host
		}
		if port == "" {
			port = portMap[scheme]
		}

		return net.JoinHostPort(hostname, port)
	}

	return net.JoinHostPort(host, portMap[scheme])
}

// https://docs.docker.com/registry/spec/auth/oauth/
func (bt *bearerTransport) refreshOauth(ctx context.Context) ([]byte, error) {
	auth, err := authn.Authorization(ctx, bt.basic)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(bt.realm)
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	bt.mx.RLock()
	v.Set("scope", strings.Join(bt.scopes, " "))
	bt.mx.RUnlock()
	if bt.service != "" {
		v.Set("service", bt.service)
	}
	v.Set("client_id", defaultUserAgent)
	if auth.IdentityToken != "" {
		v.Set("grant_type", "refresh_token")
		v.Set("refresh_token", auth.IdentityToken)
	} else if auth.Username != "" && auth.Password != "" {
		// TODO(#629): This is unreachable.
		v.Set("grant_type", "password")
		v.Set("username", auth.Username)
		v.Set("password", auth.Password)
		v.Set("access_type", "offline")
	}

	allowInsecure := bt.scheme == "http"
	client := http.Client{Transport: bt.inner, CheckRedirect: realmRedirectCheck(bt.registry.RegistryStr(), allowInsecure)}
	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// We don't want to log credentials.
	ctx = redact.NewContext(ctx, "oauth token response contains credentials")

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := CheckError(resp, http.StatusOK); err != nil {
		if bt.basic == authn.Anonymous {
			logs.Warn.Printf("No matching credentials were found for %q", bt.registry)
		}
		return nil, err
	}

	return limit.ReadAll(resp.Body, maxTokenBodySize)
}

// https://docs.docker.com/registry/spec/auth/token/
func (bt *bearerTransport) refreshBasic(ctx context.Context) ([]byte, error) {
	u, err := url.Parse(bt.realm)
	if err != nil {
		return nil, err
	}
	b := &basicTransport{
		inner:  bt.inner,
		auth:   bt.basic,
		target: u.Host,
	}
	allowInsecure := bt.scheme == "http"
	client := http.Client{Transport: b, CheckRedirect: realmRedirectCheck(bt.registry.RegistryStr(), allowInsecure)}

	v := u.Query()
	bt.mx.RLock()
	v["scope"] = bt.scopes
	bt.mx.RUnlock()
	v.Set("service", bt.service)
	u.RawQuery = v.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// We don't want to log credentials.
	ctx = redact.NewContext(ctx, "basic token response contains credentials")

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := CheckError(resp, http.StatusOK); err != nil {
		if bt.basic == authn.Anonymous {
			logs.Warn.Printf("No matching credentials were found for %q", bt.registry)
		}
		return nil, err
	}

	return limit.ReadAll(resp.Body, maxTokenBodySize)
}
