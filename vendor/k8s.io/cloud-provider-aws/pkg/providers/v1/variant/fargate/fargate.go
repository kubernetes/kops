package fargate

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	v1 "k8s.io/api/core/v1"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"

	"k8s.io/cloud-provider-aws/pkg/providers/v1/config"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/iface"
	"k8s.io/cloud-provider-aws/pkg/providers/v1/variant"
)

const (
	// fargateNodeNamePrefix string is added to awsInstance nodeName and providerID of Fargate nodes.
	fargateNodeNamePrefix = "fargate-"
)

type fargateVariant struct {
	cloudConfig *config.CloudConfig
	ec2API      iface.EC2
	credentials aws.CredentialsProvider
	provider    config.SDKProvider
}

func (f *fargateVariant) Initialize(cloudConfig *config.CloudConfig, credentials aws.CredentialsProvider, provider config.SDKProvider, ec2API iface.EC2, region string) error {
	f.cloudConfig = cloudConfig
	f.ec2API = ec2API
	f.credentials = credentials
	f.provider = provider
	return nil
}

func (f *fargateVariant) InstanceTypeByProviderID(instanceID string) (string, error) {
	return "", nil
}

func (f *fargateVariant) GetZone(ctx context.Context, instanceID, vpcID, region string) (cloudprovider.Zone, error) {
	eni, err := f.DescribeNetworkInterfaces(ctx, f.ec2API, instanceID, vpcID)
	if eni == nil || err != nil {
		return cloudprovider.Zone{}, err
	}
	return cloudprovider.Zone{
		FailureDomain: *eni.AvailabilityZone,
		Region:        region,
	}, nil
}

func (f *fargateVariant) IsSupportedNode(nodeName string) bool {
	return strings.HasPrefix(nodeName, fargateNodeNamePrefix)
}

func (f *fargateVariant) NodeAddresses(ctx context.Context, instanceID, vpcID string) ([]v1.NodeAddress, error) {
	eni, err := f.DescribeNetworkInterfaces(ctx, f.ec2API, instanceID, vpcID)
	if eni == nil || err != nil {
		return nil, err
	}

	var addresses []v1.NodeAddress

	// Assign NodeInternalIP based on IP family
	for _, family := range f.cloudConfig.Global.NodeIPFamilies {
		switch family {
		case "ipv4":
			nodeAddresses := getNodeAddressesForFargateNode(aws.ToString(eni.PrivateDnsName), aws.ToString(eni.PrivateIpAddress))
			addresses = append(addresses, nodeAddresses...)
		case "ipv6":
			if eni.Ipv6Addresses == nil || len(eni.Ipv6Addresses) == 0 {
				klog.Errorf("no Ipv6Addresses associated with the eni")
				continue
			}
			internalIPv6Address := eni.Ipv6Addresses[0].Ipv6Address
			nodeAddresses := getNodeAddressesForFargateNode(aws.ToString(eni.PrivateDnsName), aws.ToString(internalIPv6Address))
			addresses = append(addresses, nodeAddresses...)
		}
	}
	return addresses, nil
}

func (f *fargateVariant) InstanceExists(ctx context.Context, instanceID, vpcID string) (bool, error) {
	eni, err := f.DescribeNetworkInterfaces(ctx, f.ec2API, instanceID, vpcID)
	return eni != nil, err
}

func (f *fargateVariant) InstanceShutdown(ctx context.Context, instanceID, vpcID string) (bool, error) {
	eni, err := f.DescribeNetworkInterfaces(ctx, f.ec2API, instanceID, vpcID)
	return eni != nil, err
}

func newEc2Filter(name string, values ...string) ec2types.Filter {
	filter := ec2types.Filter{
		Name: aws.String(name),
	}
	for _, value := range values {
		filter.Values = append(filter.Values, value)
	}
	return filter
}

const (
	// privateDNSNamePrefix is the prefix added to ENI Private DNS Name.
	privateDNSNamePrefix = "ip-"
)

// extract private ip address from node name
func nodeNameToIPAddress(nodeName string) string {
	nodeName = strings.TrimPrefix(nodeName, privateDNSNamePrefix)
	nodeName = strings.Split(nodeName, ".")[0]
	return strings.ReplaceAll(nodeName, "-", ".")
}

// DescribeNetworkInterfaces returns network interface information for the given DNS name.
func (f *fargateVariant) DescribeNetworkInterfaces(ctx context.Context, ec2API iface.EC2, instanceID, vpcID string) (*ec2types.NetworkInterface, error) {
	eniEndpoint := strings.TrimPrefix(instanceID, fargateNodeNamePrefix)

	filters := []ec2types.Filter{
		newEc2Filter("attachment.status", "attached"),
		newEc2Filter("vpc-id", vpcID),
	}

	// when enableDnsSupport is set to false in a VPC, interface will not have private DNS names.
	// convert node name to ip address because ip-name based and resource-named EC2 resources
	// may have different privateDNSName formats but same privateIpAddress format
	if strings.HasPrefix(eniEndpoint, privateDNSNamePrefix) {
		eniEndpoint = nodeNameToIPAddress(eniEndpoint)
	}

	filters = append(filters, newEc2Filter("private-ip-address", eniEndpoint))

	request := &ec2.DescribeNetworkInterfacesInput{
		Filters: filters,
	}

	eni, err := ec2API.DescribeNetworkInterfaces(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(eni.NetworkInterfaces) == 0 {
		return nil, nil
	}
	if len(eni.NetworkInterfaces) != 1 {
		// This should not be possible - ids should be unique
		return nil, fmt.Errorf("multiple interfaces found with same id %+v", eni.NetworkInterfaces)
	}
	return &eni.NetworkInterfaces[0], nil
}

func init() {
	v := &fargateVariant{}
	variant.RegisterVariant(
		"fargate",
		v,
	)
}

// getNodeAddressesForFargateNode generates list of Node addresses for Fargate node.
func getNodeAddressesForFargateNode(privateDNSName, privateIP string) []v1.NodeAddress {
	addresses := []v1.NodeAddress{}
	addresses = append(addresses, v1.NodeAddress{Type: v1.NodeInternalIP, Address: privateIP})
	if privateDNSName != "" {
		addresses = append(addresses, v1.NodeAddress{Type: v1.NodeInternalDNS, Address: privateDNSName})
	}
	return addresses
}
