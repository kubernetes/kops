package variant

import (
	"context"
	"fmt"
	"sync"

	v1 "k8s.io/api/core/v1"
	cloudprovider "k8s.io/cloud-provider"

	"github.com/aws/aws-sdk-go-v2/aws"

	"k8s.io/cloud-provider-aws/pkg/providers/v1/config"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
)

var variantsLock sync.Mutex
var variants = make(map[string]Variant)

// Variant is a slightly different type of node
type Variant interface {
	Initialize(cloudConfig *config.CloudConfig, credentials aws.CredentialsProvider,
		provider config.SDKProvider, ec2API iface.EC2, region string) error
	IsSupportedNode(nodeName string) bool
	NodeAddresses(ctx context.Context, instanceID, vpcID string) ([]v1.NodeAddress, error)
	GetZone(ctx context.Context, instanceID, vpcID, region string) (cloudprovider.Zone, error)
	InstanceExists(ctx context.Context, instanceID, vpcID string) (bool, error)
	InstanceShutdown(ctx context.Context, instanceID, vpcID string) (bool, error)
	InstanceTypeByProviderID(id string) (string, error)
}

// RegisterVariant is used to register code that needs to be called for a specific variant
func RegisterVariant(name string, variant Variant) {
	variantsLock.Lock()
	defer variantsLock.Unlock()
	if _, found := variants[name]; found {
		panic(fmt.Sprintf("%q was registered twice", name))
	}
	variants[name] = variant
}

// IsVariantNode helps evaluate if a specific variant handles a given instance
func IsVariantNode(instanceID string) bool {
	variantsLock.Lock()
	defer variantsLock.Unlock()
	for _, v := range variants {
		if v.IsSupportedNode(instanceID) {
			return true
		}
	}
	return false
}

// NodeType returns the type name example: "fargate"
func NodeType(instanceID string) string {
	variantsLock.Lock()
	defer variantsLock.Unlock()
	for key, v := range variants {
		if v.IsSupportedNode(instanceID) {
			return key
		}
	}
	return ""
}

// GetVariant returns the interface that can then be used to handle a specific instance
func GetVariant(instanceID string) Variant {
	variantsLock.Lock()
	defer variantsLock.Unlock()
	for _, v := range variants {
		if v.IsSupportedNode(instanceID) {
			return v
		}
	}
	return nil
}

// GetVariants returns the names of all the variants registered
func GetVariants() []Variant {
	variantsLock.Lock()
	defer variantsLock.Unlock()
	var values []Variant

	// Iterate over the map and collect all values
	for _, v := range variants {
		values = append(values, v)
	}
	return values
}
