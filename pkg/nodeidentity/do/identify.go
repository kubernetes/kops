/*
Copyright 2019 The Kubernetes Authors.

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

package do

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	expirationcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodelabels"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
)

const (
	cacheTTL                 = 60 * time.Minute
	dropletRegionMetadataURL = "http://169.254.169.254/metadata/v1/region"
)

// nodeIdentifier identifies a node from DigitalOcean.
type nodeIdentifier struct {
	doClient     *godo.Client
	cache        expirationcache.Store
	cacheEnabled bool
}

// TokenSource implements oauth2.TokenSource.
type TokenSource struct {
	AccessToken string
}

// Token returns an oauth2.Token for the configured access token.
func (t *TokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: t.AccessToken}, nil
}

// New creates and returns a nodeidentity.Identifier for nodes running on DigitalOcean.
func New(cacheNodeidentityInfo bool) (nodeidentity.Identifier, error) {
	region, err := getMetadataRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet region: %s", err)
	}

	doClient, err := NewCloud(region)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize digitalocean cloud: %s", err)
	}

	return &nodeIdentifier{
		doClient:     doClient,
		cache:        expirationcache.NewTTLStore(stringKeyFunc, cacheTTL),
		cacheEnabled: cacheNodeidentityInfo,
	}, nil
}

func getMetadataRegion() (string, error) {
	return getMetadata(dropletRegionMetadataURL)
}

// NewCloud returns a godo client, expecting the env var DIGITALOCEAN_ACCESS_TOKEN to be set.
func NewCloud(region string) (*godo.Client, error) {
	accessToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("DIGITALOCEAN_ACCESS_TOKEN is required")
	}

	tokenSource := &TokenSource{AccessToken: accessToken}
	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	return godo.NewClient(oauthClient), nil
}

func getMetadata(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata URL %s: %v", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("droplet metadata returned non-200 status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read metadata information %s: %v", url, err)
	}

	return string(bodyBytes), nil
}

// IdentifyNode queries DigitalOcean for the node identity information.
func (i *nodeIdentifier) IdentifyNode(ctx context.Context, node *corev1.Node) (*nodeidentity.Info, error) {
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, errors.New("provider ID cannot be empty")
	}

	const prefix = "digitalocean://"
	if !strings.HasPrefix(providerID, prefix) {
		return nil, fmt.Errorf("provider ID %q is missing prefix %q", providerID, prefix)
	}

	instanceID := strings.TrimPrefix(providerID, prefix)
	if instanceID == "" {
		return nil, errors.New("provider ID number cannot be empty")
	}

	if i.cacheEnabled {
		if obj, exists, err := i.cache.GetByKey(instanceID); err != nil {
			klog.Warningf("Nodeidentity info cache lookup failure: %v", err)
		} else if exists {
			return obj.(*nodeidentity.Info), nil
		}
	}

	dropletID, err := strconv.Atoi(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert provider ID number %q: %s", instanceID, err)
	}

	droplet, _, err := i.doClient.Droplets.Get(ctx, dropletID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve droplet %d: %w", dropletID, err)
	}
	if droplet == nil {
		return nil, fmt.Errorf("droplet %d not found", dropletID)
	}
	if droplet.Status != "active" && droplet.Status != "new" {
		return nil, fmt.Errorf("droplet %d has unexpected status %q", dropletID, droplet.Status)
	}

	info := &nodeidentity.Info{
		InstanceID: instanceID,
		Labels:     labelsFromTags(droplet.Tags),
	}

	if i.cacheEnabled {
		if err := i.cache.Add(info); err != nil {
			klog.Warningf("Failed to add node identity info to cache: %v", err)
		}
	}

	return info, nil
}

// labelsFromTags derives node labels from a droplet's tags.
//
// The explicit kops-instance-role:<Role> tag is preferred. When it is missing
// (e.g. for droplets created before role tagging was added), the presence of
// the KubernetesCluster-Master tag is used to mark the droplet as control
// plane; otherwise we default to the worker role.
func labelsFromTags(tags []string) map[string]string {
	labels := map[string]string{}

	role, hasRoleTag := roleFromTags(tags)
	if !hasRoleTag {
		role = fallbackRoleFromTags(tags)
	}

	switch role {
	case kops.InstanceGroupSubRoleControlPlane.Role():
		labels[nodelabels.RoleLabelControlPlane20] = ""
	case kops.InstanceGroupSubRoleNode.Role():
		labels[nodelabels.RoleLabelNode16] = ""
	case kops.InstanceGroupSubRoleAPIServer.Role():
		labels[nodelabels.RoleLabelAPIServer16] = ""
	case kops.InstanceGroupSubRoleBastion.Role():
		// Bastions don't join the cluster; nothing to label.
	default:
		klog.Warningf("Unknown instance role %q on droplet tags", role)
	}

	return labels
}

func roleFromTags(tags []string) (kops.InstanceGroupRole, bool) {
	for _, tag := range tags {
		if value, ok := strings.CutPrefix(tag, do.TagKubernetesInstanceRole+":"); ok {
			return kops.InstanceGroupRole(value), true
		}
	}
	return "", false
}

func fallbackRoleFromTags(tags []string) kops.InstanceGroupRole {
	for _, tag := range tags {
		if strings.HasPrefix(tag, do.TagKubernetesClusterMasterPrefix+":") {
			return kops.InstanceGroupSubRoleControlPlane.Role()
		}
	}
	return kops.InstanceGroupSubRoleNode.Role()
}

func stringKeyFunc(obj interface{}) (string, error) {
	return obj.(*nodeidentity.Info).InstanceID, nil
}
