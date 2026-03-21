/*
Copyright The Kubernetes Authors.

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

package cloudprovider

import (
	"context"
	"fmt"
	"os"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/awslabs/operatorpkg/status"
	"k8s.io/kops/karpenter-provider-lambda-ai/pkg/apis/v1alpha1"
	"k8s.io/kops/karpenter-provider-lambda-ai/pkg/lambdaapi"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

type LambdaAICloudProvider struct {
	kubeClient   client.Client
	lambdaClient *lambdaapi.Client
}

func New(kubeClient client.Client) *LambdaAICloudProvider {
	apiKey := os.Getenv("LAMBDA_API_KEY")
	if apiKey == "" {
		// In a real scenario, we might want to fail or wait,
		// but here we'll just log or return a provider that might fail calls.
		fmt.Println("Warning: LAMBDA_API_KEY environment variable is not set")
	}

	return &LambdaAICloudProvider{
		kubeClient:   kubeClient,
		lambdaClient: lambdaapi.NewClient(apiKey),
	}
}

// Create launches an instance
func (c *LambdaAICloudProvider) Create(ctx context.Context, nodeClaim *karpv1.NodeClaim) (*karpv1.NodeClaim, error) {
	// 1. Resolve Instance Type
	// For simplicity, we assume the NodeClaim requirements strictly match one instance type or we pick the first valid one.
	// In a real provider, we'd do complex packing/selection.
	instanceTypes, err := c.lambdaClient.ListInstanceTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to list instance types: %w", err)
	}

	// Simple selection: Pick the first one that matches requirements (mock logic)
	// We really need to look at nodeClaim.Spec.Requirements
	// For this skeleton, let's just pick a default or hardcoded one if not specified,
	// or assume the NodeClaim has a label/requirement for "instance-type".

	chosenType := ""
	for name := range instanceTypes {
		chosenType = name
		break // just pick one for now
	}

	// Real logic: iterate requirements
	reqs := scheduling.NewNodeSelectorRequirementsWithMinValues(nodeClaim.Spec.Requirements...)

	if req := reqs.Get(v1.LabelInstanceTypeStable); req != nil {
		vals := req.Values()
		if len(vals) > 0 {
			chosenType = vals[0]
		}
	}

	if chosenType == "" {
		return nil, fmt.Errorf("could not resolve instance type from requirements")
	}

	// 2. Resolve Region
	// Similar logic for region
	region := "us-east-1" // Default
	if req := reqs.Get(v1.LabelTopologyZone); req != nil {
		vals := req.Values()
		if len(vals) > 0 {
			region = vals[0]
		}
	}

	// 3. Launch
	launchReq := lambdaapi.LaunchRequest{
		RegionName:       region,
		InstanceTypeName: chosenType,
		SSHKeyNames:      []string{}, // TODO: Configure SSH keys
		Quantity:         1,
		Name:             nodeClaim.Name,
	}

	ids, err := c.lambdaClient.LaunchInstance(launchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to launch instance: %w", err)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no instance IDs returned")
	}

	id := ids[0]

	// 4. Return NodeClaim with ProviderID
	nodeClaim.Status.ProviderID = fmt.Sprintf("lambda-ai://%s", id)
	// We might want to set other capacity fields here

	return nodeClaim, nil
}

// Delete removes an instance
func (c *LambdaAICloudProvider) Delete(ctx context.Context, nodeClaim *karpv1.NodeClaim) error {
	id := nodeClaim.Status.ProviderID
	// Parse ID (remove prefix if strictly needed, though lambda API takes IDs)
	// Assuming format "lambda-ai://<id>"
	var instanceID string
	_, err := fmt.Sscanf(id, "lambda-ai://%s", &instanceID)
	if err != nil {
		// Try using the whole string if scan fails, or handle error
		instanceID = id
	}

	// If empty, maybe it wasn't provisioned yet
	if instanceID == "" {
		return nil
	}

	return c.lambdaClient.TerminateInstances([]string{instanceID})
}

// Get retrieves an instance
func (c *LambdaAICloudProvider) Get(ctx context.Context, providerID string) (*karpv1.NodeClaim, error) {
	var instanceID string
	_, err := fmt.Sscanf(providerID, "lambda-ai://%s", &instanceID)
	if err != nil {
		instanceID = providerID
	}

	// Inefficient: List all to find one. API doesn't seem to have Get(ID)
	instances, err := c.lambdaClient.ListInstances()
	if err != nil {
		return nil, err
	}

	for _, inst := range instances {
		if inst.ID == instanceID {
			nc := &karpv1.NodeClaim{}
			nc.Status.ProviderID = providerID
			// Populate other status fields
			return nc, nil
		}
	}

	return nil, cloudprovider.NewNodeClaimNotFoundError(fmt.Errorf("node claim with providerID %s not found", providerID))
}

// List retrieves all instances
func (c *LambdaAICloudProvider) List(ctx context.Context) ([]*karpv1.NodeClaim, error) {
	instances, err := c.lambdaClient.ListInstances()
	if err != nil {
		return nil, err
	}

	var nodeClaims []*karpv1.NodeClaim
	for _, inst := range instances {
		nc := &karpv1.NodeClaim{}
		nc.Status.ProviderID = fmt.Sprintf("lambda-ai://%s", inst.ID)
		// Map other fields
		nodeClaims = append(nodeClaims, nc)
	}

	return nodeClaims, nil
}

// GetInstanceTypes returns available instance types
func (c *LambdaAICloudProvider) GetInstanceTypes(ctx context.Context, nodePool *karpv1.NodePool) ([]*cloudprovider.InstanceType, error) {
	types, err := c.lambdaClient.ListInstanceTypes()
	if err != nil {
		return nil, err
	}

	regions, err := c.lambdaClient.ListRegions()
	if err != nil {
		return nil, err
	}

	var result []*cloudprovider.InstanceType
	for name, details := range types {
		resources := v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%d", details.Specs.VCPUs)),
			v1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dGi", details.Specs.MemoryGib)),
		}

		if details.Specs.GPUs > 0 {
			resources["nvidia.com/gpu"] = resource.MustParse(fmt.Sprintf("%d", details.Specs.GPUs))
		}

		it := &cloudprovider.InstanceType{
			Name: name,
			Requirements: scheduling.NewRequirements(
				// Add basic requirements like Arch, OS, etc.
				scheduling.NewRequirement(v1.LabelArchStable, v1.NodeSelectorOpIn, "amd64"), // Assumption
				scheduling.NewRequirement(v1.LabelOSStable, v1.NodeSelectorOpIn, "linux"),
			),
			Capacity: resources,
		}

		// Create offerings for each region
		for _, r := range regions {
			it.Offerings = append(it.Offerings, &cloudprovider.Offering{
				Requirements: scheduling.NewRequirements(
					scheduling.NewRequirement(karpv1.CapacityTypeLabelKey, v1.NodeSelectorOpIn, karpv1.CapacityTypeOnDemand), // Lambda mainly does on-demand
					scheduling.NewRequirement(v1.LabelTopologyZone, v1.NodeSelectorOpIn, r.Name),
				),
				Price:     float64(details.PriceCentsPerHour) / 100.0,
				Available: true,
			})
		}

		result = append(result, it)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (c *LambdaAICloudProvider) Name() string {
	return "lambda-ai"
}

// IsDrifted returns whether a NodeClaim has drifted from the provisioning requirements
func (c *LambdaAICloudProvider) IsDrifted(ctx context.Context, nodeClaim *karpv1.NodeClaim) (cloudprovider.DriftReason, error) {
	// Not implemented
	return "", nil
}

// RepairPolicies returns the repair policies for the cloud provider
func (c *LambdaAICloudProvider) RepairPolicies() []cloudprovider.RepairPolicy {
	// Not implemented
	return nil
}

// GetSupportedNodeClasses returns CloudProvider NodeClass that implements status.Object
func (c *LambdaAICloudProvider) GetSupportedNodeClasses() []status.Object {
	return []status.Object{&v1alpha1.LambdaAINodeClass{}}
}
