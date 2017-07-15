/*
Copyright 2016 The Kubernetes Authors.

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

package digitalocean

import (
	"errors"
	"os"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"k8s.io/kops/pkg/resources/digitalocean/dns"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
)

// TokenSource implements oauth2.TokenSource
type TokenSource struct {
	AccessToken string
}

// Token() returns oauth2.Token
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// Cloud exposes all the interfaces required to operate on DigitalOcean resources
type Cloud struct {
	client *godo.Client

	Region string
	tags   map[string]string
}

// NewCloud returns a Cloud, expecting the env var DO_ACCESS_TOKEN
// NewCloud will return an err if DO_ACCESS_TOKEN is not defined
func NewCloud() (*Cloud, error) {
	accessToken := os.Getenv("DO_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("DO_ACCESS_TOKEN is required")
	}

	tokenSource := &TokenSource{
		AccessToken: accessToken,
	}

	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)

	return &Cloud{
		client: client,
	}, nil
}

// DNS returns a DO implementation for dnsprovider.Interface
func (c *Cloud) DNS() (dnsprovider.Interface, error) {
	provider := dns.NewProvider(c.client)
	return provider, nil
}
