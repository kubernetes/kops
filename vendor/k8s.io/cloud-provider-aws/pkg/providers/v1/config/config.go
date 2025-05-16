package config

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws/endpoints"

	"k8s.io/klog/v2"
)

const (
	// ClusterServiceLoadBalancerHealthProbeModeShared is the shared health probe mode for cluster service load balancer.
	ClusterServiceLoadBalancerHealthProbeModeShared = "Shared"

	// ClusterServiceLoadBalancerHealthProbeModeServiceNodePort is the service node port health probe mode for cluster service load balancer.
	ClusterServiceLoadBalancerHealthProbeModeServiceNodePort = "ServiceNodePort"
)

// CloudConfig wraps the settings for the AWS cloud provider.
// NOTE: Cloud config files should follow the same Kubernetes deprecation policy as
// flags or CLIs. Config fields should not change behavior in incompatible ways and
// should be deprecated for at least 2 release prior to removing.
// See https://kubernetes.io/docs/reference/using-api/deprecation-policy/#deprecating-a-flag-or-cli
// for more details.
type CloudConfig struct {
	Global struct {
		// TODO: Is there any use for this?  We can get it from the instance metadata service
		// Maybe if we're not running on AWS, e.g. bootstrap; for now it is not very useful
		Zone string

		Region string

		// The AWS VPC flag enables the possibility to run the master components
		// on a different aws account, on a different cloud provider or on-premises.
		// If the flag is set also the KubernetesClusterTag must be provided
		VPC string
		// SubnetID enables using a specific subnet to use for ELB's
		SubnetID string
		// RouteTableID enables using a specific RouteTable
		RouteTableID string

		// RoleARN is the IAM role to assume when interaction with AWS APIs.
		RoleARN string
		// SourceARN is value which is passed while assuming role specified by RoleARN. When a service
		// assumes a role in your account, you can include the aws:SourceAccount and aws:SourceArn global
		// condition context keys in your role trust policy to limit access to the role to only requests that are generated
		// by expected resources. https://docs.aws.amazon.com/IAM/latest/UserGuide/confused-deputy.html
		SourceARN string

		// KubernetesClusterTag is the legacy cluster id we'll use to identify our cluster resources
		KubernetesClusterTag string
		// KubernetesClusterID is the cluster id we'll use to identify our cluster resources
		KubernetesClusterID string

		//The aws provider creates an inbound rule per load balancer on the node security
		//group. However, this can run into the AWS security group rule limit of 50 if
		//many LoadBalancers are created.
		//
		//This flag disables the automatic ingress creation. It requires that the user
		//has setup a rule that allows inbound traffic on kubelet ports from the
		//local VPC subnet (so load balancers can access it). E.g. 10.82.0.0/16 30000-32000.
		DisableSecurityGroupIngress bool

		//AWS has a hard limit of 500 security groups. For large clusters creating a security group for each ELB
		//can cause the max number of security groups to be reached. If this is set instead of creating a new
		//Security group for each ELB this security group will be used instead.
		ElbSecurityGroup string

		// NodeIPFamilies determines which IP addresses are added to node objects and their ordering.
		NodeIPFamilies []string

		// ClusterServiceLoadBalancerHealthProbeMode determines the health probe mode for cluster service load balancer.
		// Supported values are `Shared` and `ServiceNodePort`.
		// `ServiceeNodePort`: the health probe will be created against each port of each service by watching the backend application (default).
		// `Shared`: all cluster services shares one HTTP probe targeting the kube-proxy on the node (<nodeIP>/healthz:10256).
		ClusterServiceLoadBalancerHealthProbeMode string `json:"clusterServiceLoadBalancerHealthProbeMode,omitempty" yaml:"clusterServiceLoadBalancerHealthProbeMode,omitempty"`

		// ClusterServiceSharedLoadBalancerHealthProbePort defines the target port of the shared health probe. Default to 10256.
		ClusterServiceSharedLoadBalancerHealthProbePort int32 `json:"clusterServiceSharedLoadBalancerHealthProbePort,omitempty" yaml:"clusterServiceSharedLoadBalancerHealthProbePort,omitempty"`

		// ClusterServiceSharedLoadBalancerHealthProbePath defines the target path of the shared health probe. Default to `/healthz`.
		ClusterServiceSharedLoadBalancerHealthProbePath string `json:"clusterServiceSharedLoadBalancerHealthProbePath,omitempty" yaml:"clusterServiceSharedLoadBalancerHealthProbePath,omitempty"`

		// Override to regex validating whether or not instance types require instance topology
		// to get a definitive response. This will impact whether or not the node controller will
		// block on getting instance topology information for nodes.
		// See pkg/resourcemanagers/topology.go for more details.
		//
		// WARNING: Updating the default behavior and corresponding unit tests would be a much safer option.
		SupportedTopologyInstanceTypePattern string `json:"supportedTopologyInstanceTypePattern,omitempty" yaml:"supportedTopologyInstanceTypePattern,omitempty"`
	}
	// [ServiceOverride "1"]
	//  Service = s3
	//  Region = region1
	//  URL = https://s3.foo.bar
	//  SigningRegion = signing_region
	//  SigningMethod = signing_method
	//
	//  [ServiceOverride "2"]
	//     Service = ec2
	//     Region = region2
	//     URL = https://ec2.foo.bar
	//     SigningRegion = signing_region
	//     SigningMethod = signing_method
	ServiceOverride map[string]*struct {
		Service       string
		Region        string
		URL           string
		SigningRegion string
		SigningMethod string
		SigningName   string
	}
}

// EC2Metadata is an abstraction over the AWS metadata service.
type EC2Metadata interface {
	// Query the EC2 metadata service (used to discover instance-id etc)
	GetMetadata(path string) (string, error)
	Region() (string, error)
}

// GetRegion returns the AWS region from the config, if set, or gets it from the metadata
// service if unset and sets in config
func (cfg *CloudConfig) GetRegion(metadata EC2Metadata) (string, error) {
	if cfg.Global.Region != "" {
		return cfg.Global.Region, nil
	}

	klog.Info("Loading region from metadata service")
	region, err := metadata.Region()
	if err != nil {
		return "", err
	}

	cfg.Global.Region = region
	return region, nil
}

// ValidateOverrides ensures overrides are correct
func (cfg *CloudConfig) ValidateOverrides() error {
	if len(cfg.ServiceOverride) == 0 {
		return nil
	}
	set := make(map[string]bool)
	for onum, ovrd := range cfg.ServiceOverride {
		// Note: gcfg does not space trim, so we have to when comparing to empty string ""
		name := strings.TrimSpace(ovrd.Service)
		if name == "" {
			return fmt.Errorf("service name is missing [Service is \"\"] in override %s", onum)
		}
		// insure the map service name is space trimmed
		ovrd.Service = name

		region := strings.TrimSpace(ovrd.Region)
		if region == "" {
			return fmt.Errorf("service region is missing [Region is \"\"] in override %s", onum)
		}
		// insure the map region is space trimmed
		ovrd.Region = region

		url := strings.TrimSpace(ovrd.URL)
		if url == "" {
			return fmt.Errorf("url is missing [URL is \"\"] in override %s", onum)
		}
		signingRegion := strings.TrimSpace(ovrd.SigningRegion)
		if signingRegion == "" {
			return fmt.Errorf("signingRegion is missing [SigningRegion is \"\"] in override %s", onum)
		}
		signature := name + "_" + region
		if set[signature] {
			return fmt.Errorf("duplicate entry found for service override [%s] (%s in %s)", onum, name, region)
		}
		set[signature] = true
	}
	return nil
}

// GetResolver computes the correct resolver to use
func (cfg *CloudConfig) GetResolver() endpoints.ResolverFunc {
	defaultResolver := endpoints.DefaultResolver()
	defaultResolverFn := func(service, region string,
		optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		return defaultResolver.EndpointFor(service, region, optFns...)
	}
	if len(cfg.ServiceOverride) == 0 {
		return defaultResolverFn
	}

	return func(service, region string,
		optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		for _, override := range cfg.ServiceOverride {
			if override.Service == service && override.Region == region {
				return endpoints.ResolvedEndpoint{
					URL:           override.URL,
					SigningRegion: override.SigningRegion,
					SigningMethod: override.SigningMethod,
					SigningName:   override.SigningName,
				}, nil
			}
		}
		return defaultResolver.EndpointFor(service, region, optFns...)
	}
}

// SDKProvider can be used by variants to add their own handlers
type SDKProvider interface {
	AddHandlers(regionName string, h *request.Handlers)
}
