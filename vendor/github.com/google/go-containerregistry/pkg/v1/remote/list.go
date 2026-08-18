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

package remote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// ListWithContext calls List with the given context.
//
// Deprecated: Use List and WithContext. This will be removed in a future release.
func ListWithContext(ctx context.Context, repo name.Repository, options ...Option) ([]string, error) {
	return List(repo, append(options, WithContext(ctx))...)
}

// List calls /tags/list for the given repository, returning the list of tags
// in the "tags" property.
func List(repo name.Repository, options ...Option) ([]string, error) {
	o, err := makeOptions(options...)
	if err != nil {
		return nil, err
	}
	return newPuller(o).List(o.context, repo)
}

type Tags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
	Next string   `json:"next,omitempty"`
}

func (f *fetcher) listPage(ctx context.Context, repo name.Repository, next string, pageSize int) (*Tags, error) {
	if next == "" {
		uri := &url.URL{
			Scheme: repo.Scheme(),
			Host:   repo.RegistryStr(),
			Path:   fmt.Sprintf("/v2/%s/tags/list", repo.RepositoryStr()),
		}
		if pageSize > 0 {
			uri.RawQuery = fmt.Sprintf("n=%d", pageSize)
		}
		next = uri.String()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", next, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}

	if err := transport.CheckError(resp, http.StatusOK); err != nil {
		return nil, err
	}

	parsed := Tags{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	uri, err := getNextPageURL(resp, repo)
	if err != nil {
		return nil, err
	}

	if uri != nil {
		parsed.Next = uri.String()
	}

	return &parsed, nil
}

// getNextPageURL checks if there is a Link header in a http.Response which
// contains a link to the next page. If yes it returns the url.URL of the next
// page otherwise it returns nil. It validates that the resolved URL points
// to the same registry to prevent SSRF attacks.
func getNextPageURL(resp *http.Response, repo name.Repository) (*url.URL, error) {
	link := resp.Header.Get("Link")
	if link == "" {
		return nil, nil
	}

	if link[0] != '<' {
		return nil, fmt.Errorf("failed to parse link header: missing '<' in: %s", link)
	}

	end := strings.Index(link, ">")
	if end == -1 {
		return nil, fmt.Errorf("failed to parse link header: missing '>' in: %s", link)
	}
	link = link[1:end]
	if link == "" {
		return nil, nil
	}

	linkURL, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if resp.Request == nil || resp.Request.URL == nil {
		return nil, nil
	}
	linkURL = resp.Request.URL.ResolveReference(linkURL)

	// Validate that the pagination URL points to the same registry to prevent SSRF.
	if err := validatePaginationURL(linkURL, repo); err != nil {
		return nil, err
	}

	return linkURL, nil
}

// validatePaginationURL checks that a pagination URL is safe to follow.
func validatePaginationURL(u *url.URL, repo name.Repository) error {
	return validatePaginationURLHost(u, repo.Scheme(), repo.RegistryStr())
}

// validatePaginationURLHost checks that a pagination URL is safe to follow.
func validatePaginationURLHost(u *url.URL, scheme, host string) error {
	if u.Scheme != scheme {
		return fmt.Errorf("pagination URL scheme %q does not match registry scheme %q", u.Scheme, scheme)
	}
	if u.Host != host {
		return errors.New("pagination URL host does not match registry host: potential SSRF attack")
	}
	return nil
}

type Lister struct {
	f        *fetcher
	repo     name.Repository
	pageSize int

	page *Tags
	err  error

	needMore bool
}

func (l *Lister) Next(ctx context.Context) (*Tags, error) {
	if l.needMore {
		l.page, l.err = l.f.listPage(ctx, l.repo, l.page.Next, l.pageSize)
	} else {
		l.needMore = true
	}
	return l.page, l.err
}

func (l *Lister) HasNext() bool {
	return l.page != nil && (!l.needMore || l.page.Next != "")
}

// getNextPageURLForRegistry is like getNextPageURL but for name.Registry.
func getNextPageURLForRegistry(resp *http.Response, reg name.Registry) (*url.URL, error) {
	link := resp.Header.Get("Link")
	if link == "" {
		return nil, nil
	}

	if link[0] != '<' {
		return nil, fmt.Errorf("failed to parse link header: missing '<' in: %s", link)
	}

	end := strings.Index(link, ">")
	if end == -1 {
		return nil, fmt.Errorf("failed to parse link header: missing '>' in: %s", link)
	}
	link = link[1:end]
	if link == "" {
		return nil, nil
	}

	linkURL, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if resp.Request == nil || resp.Request.URL == nil {
		return nil, nil
	}
	linkURL = resp.Request.URL.ResolveReference(linkURL)

	// Validate that the pagination URL points to the same registry to prevent SSRF.
	if err := validatePaginationURLHost(linkURL, reg.Scheme(), reg.RegistryStr()); err != nil {
		return nil, err
	}

	return linkURL, nil
}
