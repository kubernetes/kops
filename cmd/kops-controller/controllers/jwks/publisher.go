/*
Copyright 2020 The Kubernetes Authors.

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

package jwks

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/util/pkg/vfs"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// recheckDestinationInterval is the frequency with which we re-read the destination
	recheckDestinationInterval = 1 * time.Hour

	// publishInterval controls how often we should poll to republish the jwks information
	publishInterval = 1 * time.Minute
)

// ConfigurePublisher will run the jwks publisher, if configured in options
// It runs it until ctx expires.
func ConfigurePublisher(ctx context.Context, mgr manager.Manager, options *config.Options) error {
	if options.PublicDiscovery == nil {
		return nil
	}

	publisher, err := NewJWKSPublisher(mgr, options.PublicDiscovery.PublishBase, options.PublicDiscovery.PublishACL)
	if err != nil {
		return fmt.Errorf("failed to build jwks published: %v", err)
	}

	go publisher.runForever(ctx)

	return nil
}

// NewJWKSPublisher is the constructor for a JWKSPublisher
func NewJWKSPublisher(mgr manager.Manager, publishPath string, publishACL string) (*publisher, error) {
	restClient, err := rest.RESTClientFor(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("error building client for apiserver: %v", err)
	}

	c := &publisher{
		restClient: restClient,
		cache:      vfs.NewCache(),
	}

	publishBase, err := vfs.Context.BuildVfsPath(publishPath)
	if err != nil {
		return nil, fmt.Errorf("cannot parse publishPath %q: %v", publishPath, err)
	}
	c.publishBase = publishBase

	if publishACL != "" {
		switch publishBase.(type) {
		case *vfs.S3Path:
			c.publishACL = &vfs.S3Acl{
				RequestACL: aws.String(publishACL),
			}
		default:
			return nil, fmt.Errorf("ACL not handled for path %q", publishBase)
		}
	}

	return c, nil
}

// publisher republishes the jwks documents to a vfs.Path
// This lets us expose the JWKS information publicly without exposing the apiserver.
type publisher struct {
	restClient *rest.RESTClient

	// publishBase is the base path we should publish to
	publishBase vfs.Path

	// publishACL sets the permissions we should use to publish with (likely world-readable)
	publishACL vfs.ACL

	// cache caches the published values, to avoid repeated GCS/S3 calls
	cache *vfs.Cache
}

// runForever will reattempt the publish step every publishInterval
func (c *publisher) runForever(ctx context.Context) {
	for {
		err := c.republish(ctx)
		if err != nil {
			klog.Warningf("error publishing jwks information: %v", err)
		}

		if ctx.Err() != nil {
			break
		}

		time.Sleep(publishInterval)
	}
}

// republish copies the jwks information from the apiserver to the destination vfs path
func (c *publisher) republish(ctx context.Context) error {
	expected, err := c.getJWKSDocuments(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch jwks information: %v", err)
	}

	changed := false
	for p, b := range expected {
		actual, err := c.cache.Read(p, recheckDestinationInterval)
		if err != nil {
			return fmt.Errorf("error reading %s: %v", p, err)
		}

		if bytes.Equal(actual, b) {
			klog.V(4).Infof("file %s is up to date", p)
			continue
		}

		if err := p.WriteFile(bytes.NewReader(b), c.publishACL); err != nil {
			return fmt.Errorf("error writing file %s: %v", p, err)
		}
		klog.Infof("updated jwks file %s")
		changed = true
	}

	if !changed {
		klog.Infof("no jwks changes detected")
	}

	return nil
}

// getJWKSDocuments queries the apiserver to retrieve the jwks documents we should publish
func (c *publisher) getJWKSDocuments(ctx context.Context) (map[vfs.Path][]byte, error) {
	paths := []string{
		".well-known/openid-configuration",
		"openid/v1/jwks",
	}

	expected := make(map[vfs.Path][]byte)
	for _, p := range paths {
		contents, err := c.restClient.Get().RequestURI(p).DoRaw(ctx)
		if err != nil {
			return nil, fmt.Errorf("error fetching %s: %v", p, err)
		}

		destPath := c.publishBase.Join(p)
		expected[destPath] = contents
	}

	return expected, nil
}
