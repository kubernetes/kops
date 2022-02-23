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

package gce

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	compute "google.golang.org/api/compute/v1"
	oauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/storage/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/google/clouddns"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcemetadata"
)

type GCECloud interface {
	fi.Cloud
	Compute() ComputeClient
	Storage() *storage.Service
	IAM() IamClient
	CloudDNS() DNSClient
	Project() string
	WaitForOp(op *compute.Operation) error
	Labels() map[string]string
	Zones() ([]string, error)

	// ServiceAccount returns the email for the service account that the instances will run under
	ServiceAccount() (string, error)

	// CloudResourceManager returns the client for the cloudresourcemanager API
	CloudResourceManager() *cloudresourcemanager.Service
}

type gceCloudImplementation struct {
	compute *computeClientImpl
	storage *storage.Service
	iam     *iamClientImpl
	dns     *dnsClientImpl

	// cloudResourceManager is the client for the cloudresourcemanager API
	cloudResourceManager *cloudresourcemanager.Service

	region  string
	project string

	// projectInfo caches the project info from the compute API
	projectInfo *compute.Project

	labels map[string]string
}

var _ fi.Cloud = &gceCloudImplementation{}

func (c *gceCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderGCE
}

var gceCloudInstances map[string]GCECloud = make(map[string]GCECloud)

// DefaultProject returns the current project configured in the gcloud SDK, ("", nil) if no project was set
func DefaultProject() (string, error) {
	// The default project isn't usually defined by the google cloud APIs,
	// for example the Application Default Credential won't have ProjectID set.
	// If we're running on a GCP instance, we can get it from the metadata service,
	// but the normal kops CLI usage is running locally with gcloud configuration with a project,
	// so we use that value.
	cmd := exec.Command("gcloud", "config", "get-value", "project")

	env := os.Environ()
	cmd.Env = env
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	human := strings.Join(cmd.Args, " ")
	klog.V(2).Infof("Running command: %s", human)
	err := cmd.Run()
	if err != nil {
		klog.Infof("error running %s", human)
		klog.Info(stdout.String())
		klog.Info(stderr.String())
		return "", fmt.Errorf("error running %s: %v", human, err)
	}

	projectID := strings.TrimSpace(stdout.String())
	return projectID, err
}

func NewGCECloud(region string, project string, labels map[string]string) (GCECloud, error) {
	i := gceCloudInstances[region+"::"+project]
	if i != nil {
		return i.(gceCloudInternal).WithLabels(labels), nil
	}

	c := &gceCloudImplementation{region: region, project: project}

	ctx := context.Background()

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		klog.Infof("Will load GOOGLE_APPLICATION_CREDENTIALS from %s", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	}

	computeClient, err := newComputeClientImpl(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building compute API client: %v", err)
	}
	c.compute = computeClient

	storageService, err := storage.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building storage API client: %v", err)
	}
	c.storage = storageService

	iamService, err := newIamClientImpl(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building IAM API client: %v", err)
	}
	c.iam = iamService

	dnsClient, err := newDNSClientImpl(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building DNS API client: %v", err)
	}
	c.dns = dnsClient

	cloudResourceManager, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error building cloudresourcemanager API client: %w", err)
	}
	c.cloudResourceManager = cloudResourceManager

	CacheGCECloudInstance(region, project, c)

	{
		// Attempt to log the current GCE service account in user, for diagnostic purposes
		// At least until we get e2e running, we're doing this always
		tokenInfo, err := c.getTokenInfo(ctx)
		if err != nil {
			klog.Infof("unable to get token info: %v", err)
		} else {
			klog.V(2).Infof("running with GCE credentials: email=%s, scope=%s", tokenInfo.Email, tokenInfo.Scope)
		}
	}

	return c.WithLabels(labels), nil
}

func CacheGCECloudInstance(region string, project string, c GCECloud) {
	gceCloudInstances[region+"::"+project] = c
}

// gceCloudInternal is an interface for private functions for a gceCloudImplemention or mockGCECloud
type gceCloudInternal interface {
	// WithLabels returns a copy of the GCECloud, bound to the specified labels
	WithLabels(labels map[string]string) GCECloud
}

// WithLabels returns a copy of the GCECloud, bound to the specified labels
func (c *gceCloudImplementation) WithLabels(labels map[string]string) GCECloud {
	i := &gceCloudImplementation{}
	*i = *c
	i.labels = labels
	return i
}

// Compute returns private struct element compute.
func (c *gceCloudImplementation) Compute() ComputeClient {
	return c.compute
}

// Storage returns private struct element storage.
func (c *gceCloudImplementation) Storage() *storage.Service {
	return c.storage
}

// IAM returns the IAM client
func (c *gceCloudImplementation) IAM() IamClient {
	return c.iam
}

// CloudDNS returns the DNS client
func (c *gceCloudImplementation) CloudDNS() DNSClient {
	return c.dns
}

// CloudResourceManager returns the client for the cloudresourcemanager API
func (c *gceCloudImplementation) CloudResourceManager() *cloudresourcemanager.Service {
	return c.cloudResourceManager
}

// Region returns private struct element region.
func (c *gceCloudImplementation) Region() string {
	return c.region
}

// Project returns private struct element project.
func (c *gceCloudImplementation) Project() string {
	return c.project
}

// ServiceAccount returns the email address for the service account that the instances will run under.
func (c *gceCloudImplementation) ServiceAccount() (string, error) {
	if c.projectInfo == nil {
		// Find the project info from the compute API, which includes the default service account
		klog.V(2).Infof("fetching project %q from compute API", c.project)
		p, err := c.compute.Projects().Get(c.project)
		if err != nil {
			return "", fmt.Errorf("error fetching info for project %q: %v", c.project, err)
		}

		c.projectInfo = p
	}

	if c.projectInfo.DefaultServiceAccount == "" {
		return "", fmt.Errorf("compute project %q did not have DefaultServiceAccount", c.project)
	}

	return c.projectInfo.DefaultServiceAccount, nil
}

func (c *gceCloudImplementation) DNS() (dnsprovider.Interface, error) {
	provider, err := clouddns.CreateInterface(c.project, nil)
	if err != nil {
		return nil, fmt.Errorf("error building (k8s) DNS provider: %v", err)
	}
	return provider, nil
}

func (c *gceCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	klog.Warningf("FindVPCInfo not (yet) implemented on GCE")
	return nil, nil
}

func (c *gceCloudImplementation) Labels() map[string]string {
	// Defensive copy
	tags := make(map[string]string)
	for k, v := range c.labels {
		tags[k] = v
	}
	return tags
}

// TODO refactor this out of resources
// this is needed for delete groups and other new methods

// Zones returns the zones in a region
func (c *gceCloudImplementation) Zones() ([]string, error) {
	var zones []string
	// TODO: Only zones in api.Cluster object, if we have one?
	gceZones, err := c.Compute().Zones().List(context.Background(), c.Project())
	if err != nil {
		return nil, fmt.Errorf("error listing zones: %v", err)
	}
	for _, gceZone := range gceZones {
		u, err := ParseGoogleCloudURL(gceZone.Region)
		if err != nil {
			return nil, err
		}
		if u.Name != c.Region() {
			continue
		}
		zones = append(zones, gceZone.Name)
	}
	if len(zones) == 0 {
		return nil, fmt.Errorf("unable to determine zones in region %q", c.Region())
	}

	klog.Infof("Scanning zones: %v", zones)
	return zones, nil
}

func (c *gceCloudImplementation) WaitForOp(op *compute.Operation) error {
	return WaitForOp(c.compute.srv, op)
}

func (c *gceCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	var ingresses []fi.ApiIngressStatus

	// Note that this must match GCEModelContext::NameForForwardingRule
	name := SafeObjectName("api", cluster.ObjectMeta.Name)

	klog.V(2).Infof("Querying GCE to find ForwardingRules for API (%q)", name)
	forwardingRule, err := c.compute.ForwardingRules().Get(c.project, c.region, name)
	if err != nil {
		if !IsNotFound(err) {
			forwardingRule = nil
		} else {
			return nil, fmt.Errorf("error getting ForwardingRule %q: %v", name, err)
		}
	}

	if forwardingRule != nil {
		if forwardingRule.IPAddress == "" {
			return nil, fmt.Errorf("found forward rule %q, but it did not have an IPAddress", name)
		}

		ingresses = append(ingresses, fi.ApiIngressStatus{
			IP: forwardingRule.IPAddress,
		})
	}

	return ingresses, nil
}

// FindInstanceTemplates finds all instance templates that are associated with the current cluster
// It matches them by looking for instance metadata with key='cluster-name' and value of our cluster name
func FindInstanceTemplates(c GCECloud, clusterName string) ([]*compute.InstanceTemplate, error) {
	findClusterName := strings.TrimSpace(clusterName)

	ts, err := c.Compute().InstanceTemplates().List(context.Background(), c.Project())
	if err != nil {
		return nil, fmt.Errorf("error listing instance templates: %v", err)
	}

	var matches []*compute.InstanceTemplate
	for _, t := range ts {
		if !gcemetadata.MetadataMatchesClusterName(findClusterName, t.Properties.Metadata) {
			continue
		}

		matches = append(matches, t)
	}

	return matches, nil
}

// logTokenInfo returns information about the active credential
func (c *gceCloudImplementation) getTokenInfo(ctx context.Context) (*oauth2.Tokeninfo, error) {
	tokenSource, err := google.DefaultTokenSource(ctx, compute.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("error building token source: %v", err)
	}

	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %v", err)
	}

	// Note: do not log token or any portion of it

	service, err := oauth2.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating oauth2 service: %v", err)
	}

	tokenInfo, err := service.Tokeninfo().AccessToken(token.AccessToken).Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching oauth2 token info: %v", err)
	}

	return tokenInfo, nil
}

// SplitServiceAccountEmail splits service account email
func SplitServiceAccountEmail(email string) (string, string, error) {
	accountID := ""
	projectID := ""

	tokens := strings.Split(email, "@")
	if len(tokens) == 2 {
		accountID = tokens[0]
		if strings.HasSuffix(tokens[1], ".iam.gserviceaccount.com") {
			projectID = strings.TrimSuffix(tokens[1], ".iam.gserviceaccount.com")
		}
	}

	if accountID == "" || projectID == "" {
		return "", "", fmt.Errorf("unexpected format for ServiceAccount email %q", email)
	}
	return accountID, projectID, nil
}
