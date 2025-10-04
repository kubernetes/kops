/*
Copyright 2014 The Kubernetes Authors.

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

package aws

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// ProxyProtocolPolicyName is the tag named used for the proxy protocol
	// policy
	ProxyProtocolPolicyName = "k8s-proxyprotocol-enabled"

	// SSLNegotiationPolicyNameFormat is a format string used for the SSL
	// negotiation policy tag name
	SSLNegotiationPolicyNameFormat = "k8s-SSLNegotiationPolicy-%s"

	lbAttrLoadBalancingCrossZoneEnabled = "load_balancing.cross_zone.enabled"
	lbAttrAccessLogsS3Enabled           = "access_logs.s3.enabled"
	lbAttrAccessLogsS3Bucket            = "access_logs.s3.bucket"
	lbAttrAccessLogsS3Prefix            = "access_logs.s3.prefix"

	// defaultEC2InstanceCacheMaxAge is the max age for the EC2 instance cache
	defaultEC2InstanceCacheMaxAge = 10 * time.Minute
)

var (
	// Defaults for ELB Healthcheck
	defaultElbHCHealthyThreshold   = int32(2)
	defaultElbHCUnhealthyThreshold = int32(6)
	defaultElbHCTimeout            = int32(5)
	defaultElbHCInterval           = int32(10)
	defaultNlbHealthCheckInterval  = int32(30)
	defaultNlbHealthCheckTimeout   = int32(10)
	defaultNlbHealthCheckThreshold = int32(3)
	defaultHealthCheckPort         = "traffic-port"
	defaultHealthCheckPath         = "/"

	defaultKubeProxyHealthCheckPort = 10256
	defaultKubeProxyHealthCheckPath = "/healthz"

	// Defaults for ELB Target operations
	defaultRegisterTargetsChunkSize   = 100
	defaultDeregisterTargetsChunkSize = 100
)

func isNLB(annotations map[string]string) bool {
	if annotations[ServiceAnnotationLoadBalancerType] == "nlb" {
		return true
	}
	return false
}

func isLBExternal(annotations map[string]string) bool {
	if val := annotations[ServiceAnnotationLoadBalancerType]; val == "nlb-ip" || val == "external" {
		return true
	}
	return false
}

type healthCheckConfig struct {
	Port               string
	Path               string
	Protocol           elbv2types.ProtocolEnum
	Interval           int32
	Timeout            int32
	HealthyThreshold   int32
	UnhealthyThreshold int32
}

type nlbPortMapping struct {
	FrontendPort     int32
	FrontendProtocol elbv2types.ProtocolEnum

	TrafficPort     int32
	TrafficProtocol elbv2types.ProtocolEnum

	SSLCertificateARN string
	SSLPolicy         string
	HealthCheckConfig healthCheckConfig
}

// getKeyValuePropertiesFromAnnotation converts the comma separated list of key-value
// pairs from the specified annotation and returns it as a map.
func getKeyValuePropertiesFromAnnotation(annotations map[string]string, annotation string) map[string]string {
	additionalTags := make(map[string]string)
	if additionalTagsList, ok := annotations[annotation]; ok {
		additionalTagsList = strings.TrimSpace(additionalTagsList)

		// Break up list of "Key1=Val,Key2=Val2"
		tagList := strings.Split(additionalTagsList, ",")

		// Break up "Key=Val"
		for _, tagSet := range tagList {
			tag := strings.Split(strings.TrimSpace(tagSet), "=")

			// Accept "Key=val" or "Key=" or just "Key"
			if len(tag) >= 2 && len(tag[0]) != 0 {
				// There is a key and a value, so save it
				additionalTags[tag[0]] = tag[1]
			} else if len(tag) == 1 && len(tag[0]) != 0 {
				// Just "Key"
				additionalTags[tag[0]] = ""
			}
		}
	}

	return additionalTags
}

// ensureLoadBalancerv2 ensures a v2 load balancer is created
func (c *Cloud) ensureLoadBalancerv2(ctx context.Context, namespacedName types.NamespacedName, loadBalancerName string, mappings []nlbPortMapping, instanceIDs, discoveredSubnetIDs []string, internalELB bool, annotations map[string]string) (*elbv2types.LoadBalancer, error) {
	loadBalancer, err := c.describeLoadBalancerv2(ctx, loadBalancerName)
	if err != nil {
		return nil, err
	}

	dirty := false

	// Get additional tags set by the user
	tags := getKeyValuePropertiesFromAnnotation(annotations, ServiceAnnotationLoadBalancerAdditionalTags)
	// Add default tags
	tags[TagNameKubernetesService] = namespacedName.String()
	tags = c.tagging.buildTags(ResourceLifecycleOwned, tags)

	if loadBalancer == nil {
		// Create the LB
		createRequest := &elbv2.CreateLoadBalancerInput{
			Type: elbv2types.LoadBalancerTypeEnumNetwork,
			Name: aws.String(loadBalancerName),
		}
		if internalELB {
			createRequest.Scheme = elbv2types.LoadBalancerSchemeEnumInternal
		}

		var allocationIDs []string
		if eipList, present := annotations[ServiceAnnotationLoadBalancerEIPAllocations]; present {
			allocationIDs = strings.Split(eipList, ",")
			if len(allocationIDs) != len(discoveredSubnetIDs) {
				return nil, fmt.Errorf("error creating load balancer: Must have same number of EIP AllocationIDs (%d) and SubnetIDs (%d)", len(allocationIDs), len(discoveredSubnetIDs))
			}
		}

		// We are supposed to specify one subnet per AZ.
		// TODO: What happens if we have more than one subnet per AZ?
		createRequest.SubnetMappings = createSubnetMappings(discoveredSubnetIDs, allocationIDs)

		for k, v := range tags {
			createRequest.Tags = append(createRequest.Tags, elbv2types.Tag{
				Key: aws.String(k), Value: aws.String(v),
			})
		}

		klog.Infof("Creating load balancer for %v with name: %s", namespacedName, loadBalancerName)
		createResponse, err := c.elbv2.CreateLoadBalancer(ctx, createRequest)
		if err != nil {
			return nil, fmt.Errorf("error creating load balancer: %q", err)
		}

		loadBalancer = &createResponse.LoadBalancers[0]
		for i := range mappings {
			// It is easier to keep track of updates by having possibly
			// duplicate target groups where the backend port is the same
			_, err := c.createListenerV2(ctx, createResponse.LoadBalancers[0].LoadBalancerArn, mappings[i], namespacedName, instanceIDs, *createResponse.LoadBalancers[0].VpcId, tags)
			if err != nil {
				return nil, fmt.Errorf("error creating listener: %q", err)
			}
		}
		if err := c.reconcileLBAttributes(ctx, aws.ToString(loadBalancer.LoadBalancerArn), annotations); err != nil {
			return nil, err
		}
	} else {
		// TODO: Sync internal vs non-internal

		// sync mappings
		{
			listenerDescriptions, err := c.elbv2.DescribeListeners(ctx,
				&elbv2.DescribeListenersInput{
					LoadBalancerArn: loadBalancer.LoadBalancerArn,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("error describing listeners: %q", err)
			}

			// actual maps FrontendPort to an elbv2.Listener
			actual := map[int32]map[elbv2types.ProtocolEnum]*elbv2types.Listener{}
			for _, listener := range listenerDescriptions.Listeners {
				if actual[*listener.Port] == nil {
					actual[*listener.Port] = map[elbv2types.ProtocolEnum]*elbv2types.Listener{}
				}
				actual[*listener.Port][listener.Protocol] = &listener
			}

			actualTargetGroups, err := c.elbv2.DescribeTargetGroups(ctx,
				&elbv2.DescribeTargetGroupsInput{
					LoadBalancerArn: loadBalancer.LoadBalancerArn,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("error listing target groups: %q", err)
			}

			nodePortTargetGroup := map[int32]*elbv2types.TargetGroup{}
			for _, targetGroup := range actualTargetGroups.TargetGroups {
				nodePortTargetGroup[*targetGroup.Port] = &targetGroup
			}

			// Handle additions/modifications
			for _, mapping := range mappings {
				frontendPort := mapping.FrontendPort
				frontendProtocol := mapping.FrontendProtocol
				nodePort := mapping.TrafficPort
				// modifications
				if listener, ok := actual[frontendPort][frontendProtocol]; ok {
					listenerNeedsModification := false

					if listener.Protocol != mapping.FrontendProtocol {
						listenerNeedsModification = true
					}
					switch mapping.FrontendProtocol {
					case elbv2types.ProtocolEnumTls:
						{
							if aws.ToString(listener.SslPolicy) != mapping.SSLPolicy {
								listenerNeedsModification = true
							}
							if len(listener.Certificates) == 0 || aws.ToString(listener.Certificates[0].CertificateArn) != mapping.SSLCertificateARN {
								listenerNeedsModification = true
							}
						}
					case elbv2types.ProtocolEnumTcp:
						{
							if aws.ToString(listener.SslPolicy) != "" {
								listenerNeedsModification = true
							}
							if len(listener.Certificates) != 0 {
								listenerNeedsModification = true
							}
						}
					}

					// recreate targetGroup if trafficPort, protocol or HealthCheckProtocol changed
					healthCheckModified := false
					targetGroupRecreated := false
					targetGroup, ok := nodePortTargetGroup[nodePort]

					if targetGroup != nil && (!strings.EqualFold(string(mapping.HealthCheckConfig.Protocol), string(targetGroup.HealthCheckProtocol)) ||
						mapping.HealthCheckConfig.Interval != aws.ToInt32(targetGroup.HealthCheckIntervalSeconds)) {
						healthCheckModified = true
					}

					if !ok || targetGroup.Protocol != mapping.TrafficProtocol || healthCheckModified {
						// create new target group
						targetGroup, err = c.ensureTargetGroup(ctx,
							nil,
							namespacedName,
							mapping,
							instanceIDs,
							*loadBalancer.VpcId,
							tags,
						)
						if err != nil {
							return nil, err
						}
						targetGroupRecreated = true
						listenerNeedsModification = true
					}

					if listenerNeedsModification {
						modifyListenerInput := &elbv2.ModifyListenerInput{
							ListenerArn: listener.ListenerArn,
							Port:        aws.Int32(frontendPort),
							Protocol:    mapping.FrontendProtocol,
							DefaultActions: []elbv2types.Action{{
								TargetGroupArn: targetGroup.TargetGroupArn,
								Type:           elbv2types.ActionTypeEnumForward,
							}},
						}
						if mapping.FrontendProtocol == elbv2types.ProtocolEnumTls {
							if mapping.SSLPolicy != "" {
								modifyListenerInput.SslPolicy = aws.String(mapping.SSLPolicy)
							}
							modifyListenerInput.Certificates = []elbv2types.Certificate{
								{
									CertificateArn: aws.String(mapping.SSLCertificateARN),
								},
							}
						}
						if _, err := c.elbv2.ModifyListener(ctx, modifyListenerInput); err != nil {
							return nil, fmt.Errorf("error updating load balancer listener: %q", err)
						}
					}

					// Delete old targetGroup if needed
					if targetGroupRecreated {
						if _, err := c.elbv2.DeleteTargetGroup(ctx, &elbv2.DeleteTargetGroupInput{
							TargetGroupArn: listener.DefaultActions[0].TargetGroupArn,
						}); err != nil {
							return nil, fmt.Errorf("error deleting old target group: %q", err)
						}
					} else {
						// Run ensureTargetGroup to make sure instances in service are up-to-date
						_, err = c.ensureTargetGroup(ctx,
							targetGroup,
							namespacedName,
							mapping,
							instanceIDs,
							*loadBalancer.VpcId,
							tags,
						)
						if err != nil {
							return nil, err
						}
					}
					dirty = true
					continue
				}

				// Additions
				_, err := c.createListenerV2(ctx, loadBalancer.LoadBalancerArn, mapping, namespacedName, instanceIDs, *loadBalancer.VpcId, tags)
				if err != nil {
					return nil, err
				}
				dirty = true
			}

			frontEndPorts := map[int32]map[elbv2types.ProtocolEnum]bool{}
			for i := range mappings {
				if frontEndPorts[mappings[i].FrontendPort] == nil {
					frontEndPorts[mappings[i].FrontendPort] = map[elbv2types.ProtocolEnum]bool{}
				}
				frontEndPorts[mappings[i].FrontendPort][mappings[i].FrontendProtocol] = true
			}

			// handle deletions
			for port := range actual {
				for protocol := range actual[port] {
					if _, ok := frontEndPorts[port][protocol]; !ok {
						err := c.deleteListenerV2(ctx, actual[port][protocol])
						if err != nil {
							return nil, err
						}
						dirty = true
					}
				}
			}
		}
		if err := c.reconcileLBAttributes(ctx, aws.ToString(loadBalancer.LoadBalancerArn), annotations); err != nil {
			return nil, err
		}

		// Subnets cannot be modified on NLBs
		if dirty {
			loadBalancers, err := c.elbv2.DescribeLoadBalancers(ctx,
				&elbv2.DescribeLoadBalancersInput{
					LoadBalancerArns: []string{
						aws.ToString(loadBalancer.LoadBalancerArn),
					},
				},
			)
			if err != nil {
				return nil, fmt.Errorf("error retrieving load balancer after update: %q", err)
			}
			loadBalancer = &loadBalancers.LoadBalancers[0]
		}
	}
	return loadBalancer, nil
}

func (c *Cloud) reconcileLBAttributes(ctx context.Context, loadBalancerArn string, annotations map[string]string) error {
	desiredLoadBalancerAttributes := map[string]string{}

	desiredLoadBalancerAttributes[lbAttrLoadBalancingCrossZoneEnabled] = "false"
	crossZoneLoadBalancingEnabledAnnotation := annotations[ServiceAnnotationLoadBalancerCrossZoneLoadBalancingEnabled]
	if crossZoneLoadBalancingEnabledAnnotation != "" {
		crossZoneEnabled, err := strconv.ParseBool(crossZoneLoadBalancingEnabledAnnotation)
		if err != nil {
			return fmt.Errorf("error parsing service annotation: %s=%s",
				ServiceAnnotationLoadBalancerCrossZoneLoadBalancingEnabled,
				crossZoneLoadBalancingEnabledAnnotation,
			)
		}

		if crossZoneEnabled {
			desiredLoadBalancerAttributes[lbAttrLoadBalancingCrossZoneEnabled] = "true"
		}
	}

	desiredLoadBalancerAttributes[lbAttrAccessLogsS3Enabled] = "false"
	accessLogsS3EnabledAnnotation := annotations[ServiceAnnotationLoadBalancerAccessLogEnabled]
	if accessLogsS3EnabledAnnotation != "" {
		accessLogsS3Enabled, err := strconv.ParseBool(accessLogsS3EnabledAnnotation)
		if err != nil {
			return fmt.Errorf("error parsing service annotation: %s=%s",
				ServiceAnnotationLoadBalancerAccessLogEnabled,
				accessLogsS3EnabledAnnotation,
			)
		}

		if accessLogsS3Enabled {
			desiredLoadBalancerAttributes[lbAttrAccessLogsS3Enabled] = "true"
		}
	}

	desiredLoadBalancerAttributes[lbAttrAccessLogsS3Bucket] = annotations[ServiceAnnotationLoadBalancerAccessLogS3BucketName]
	desiredLoadBalancerAttributes[lbAttrAccessLogsS3Prefix] = annotations[ServiceAnnotationLoadBalancerAccessLogS3BucketPrefix]

	currentLoadBalancerAttributes := map[string]string{}
	describeAttributesOutput, err := c.elbv2.DescribeLoadBalancerAttributes(ctx, &elbv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(loadBalancerArn),
	})
	if err != nil {
		return fmt.Errorf("unable to retrieve load balancer attributes during attribute sync: %q", err)
	}
	for _, attr := range describeAttributesOutput.Attributes {
		currentLoadBalancerAttributes[aws.ToString(attr.Key)] = aws.ToString(attr.Value)
	}

	var changedAttributes []elbv2types.LoadBalancerAttribute
	if desiredLoadBalancerAttributes[lbAttrLoadBalancingCrossZoneEnabled] != currentLoadBalancerAttributes[lbAttrLoadBalancingCrossZoneEnabled] {
		changedAttributes = append(changedAttributes, elbv2types.LoadBalancerAttribute{
			Key:   aws.String(lbAttrLoadBalancingCrossZoneEnabled),
			Value: aws.String(desiredLoadBalancerAttributes[lbAttrLoadBalancingCrossZoneEnabled]),
		})
	}
	if desiredLoadBalancerAttributes[lbAttrAccessLogsS3Enabled] != currentLoadBalancerAttributes[lbAttrAccessLogsS3Enabled] {
		changedAttributes = append(changedAttributes, elbv2types.LoadBalancerAttribute{
			Key:   aws.String(lbAttrAccessLogsS3Enabled),
			Value: aws.String(desiredLoadBalancerAttributes[lbAttrAccessLogsS3Enabled]),
		})
	}

	// ELBV2 API forbids us to set bucket to an empty bucket, so we keep it unchanged if AccessLogsS3Enabled==false.
	if desiredLoadBalancerAttributes[lbAttrAccessLogsS3Enabled] == "true" {
		if desiredLoadBalancerAttributes[lbAttrAccessLogsS3Bucket] != currentLoadBalancerAttributes[lbAttrAccessLogsS3Bucket] {
			changedAttributes = append(changedAttributes, elbv2types.LoadBalancerAttribute{
				Key:   aws.String(lbAttrAccessLogsS3Bucket),
				Value: aws.String(desiredLoadBalancerAttributes[lbAttrAccessLogsS3Bucket]),
			})
		}
		if desiredLoadBalancerAttributes[lbAttrAccessLogsS3Prefix] != currentLoadBalancerAttributes[lbAttrAccessLogsS3Prefix] {
			changedAttributes = append(changedAttributes, elbv2types.LoadBalancerAttribute{
				Key:   aws.String(lbAttrAccessLogsS3Prefix),
				Value: aws.String(desiredLoadBalancerAttributes[lbAttrAccessLogsS3Prefix]),
			})
		}
	}

	if len(changedAttributes) > 0 {
		klog.V(2).Infof("updating load-balancer attributes for %q", loadBalancerArn)

		_, err = c.elbv2.ModifyLoadBalancerAttributes(ctx, &elbv2.ModifyLoadBalancerAttributesInput{
			LoadBalancerArn: aws.String(loadBalancerArn),
			Attributes:      changedAttributes,
		})
		if err != nil {
			return fmt.Errorf("unable to update load balancer attributes during attribute sync: %q", err)
		}
	}
	return nil
}

var invalidELBV2NameRegex = regexp.MustCompile("[^[:alnum:]]")

// buildTargetGroupName will build unique name for targetGroup of service & port.
// the name is in format k8s-{namespace:8}-{name:8}-{uuid:10} (chosen to benefit most common use cases).
// Note: nodePort & targetProtocol & targetType are included since they cannot be modified on existing targetGroup.
func (c *Cloud) buildTargetGroupName(serviceName types.NamespacedName, servicePort int32, nodePort int32, targetProtocol elbv2types.ProtocolEnum, targetType elbv2types.TargetTypeEnum, mapping nlbPortMapping) string {
	hasher := sha1.New()
	_, _ = hasher.Write([]byte(c.tagging.clusterID()))
	_, _ = hasher.Write([]byte(serviceName.Namespace))
	_, _ = hasher.Write([]byte(serviceName.Name))
	_, _ = hasher.Write([]byte(strconv.FormatInt(int64(servicePort), 10)))
	_, _ = hasher.Write([]byte(strconv.FormatInt(int64(nodePort), 10)))
	_, _ = hasher.Write([]byte(targetProtocol))
	_, _ = hasher.Write([]byte(targetType))
	_, _ = hasher.Write([]byte(mapping.HealthCheckConfig.Protocol))
	_, _ = hasher.Write([]byte(strconv.FormatInt(int64(mapping.HealthCheckConfig.Interval), 10)))
	tgUUID := hex.EncodeToString(hasher.Sum(nil))

	sanitizedNamespace := invalidELBV2NameRegex.ReplaceAllString(serviceName.Namespace, "")
	sanitizedServiceName := invalidELBV2NameRegex.ReplaceAllString(serviceName.Name, "")
	return fmt.Sprintf("k8s-%.8s-%.8s-%.10s", sanitizedNamespace, sanitizedServiceName, tgUUID)
}

func (c *Cloud) createListenerV2(ctx context.Context, loadBalancerArn *string, mapping nlbPortMapping, namespacedName types.NamespacedName, instanceIDs []string, vpcID string, tags map[string]string) (listener *elbv2types.Listener, err error) {
	target, err := c.ensureTargetGroup(ctx,
		nil,
		namespacedName,
		mapping,
		instanceIDs,
		vpcID,
		tags,
	)
	if err != nil {
		return nil, err
	}

	elbTags := []elbv2types.Tag{}
	for k, v := range tags {
		elbTags = append(elbTags, elbv2types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	createListernerInput := &elbv2.CreateListenerInput{
		LoadBalancerArn: loadBalancerArn,
		Port:            aws.Int32(mapping.FrontendPort),
		Protocol:        mapping.FrontendProtocol,
		DefaultActions: []elbv2types.Action{{
			TargetGroupArn: target.TargetGroupArn,
			Type:           elbv2types.ActionTypeEnumForward,
		}},
		Tags: elbTags,
	}
	if mapping.FrontendProtocol == "TLS" {
		if mapping.SSLPolicy != "" {
			createListernerInput.SslPolicy = aws.String(mapping.SSLPolicy)
		}
		createListernerInput.Certificates = []elbv2types.Certificate{
			{
				CertificateArn: aws.String(mapping.SSLCertificateARN),
			},
		}
	}

	klog.Infof("Creating load balancer listener for %v", namespacedName)
	createListenerOutput, err := c.elbv2.CreateListener(ctx, createListernerInput)
	if err != nil {
		return nil, fmt.Errorf("error creating load balancer listener: %q", err)
	}
	return &createListenerOutput.Listeners[0], nil
}

// cleans up listener and corresponding target group
func (c *Cloud) deleteListenerV2(ctx context.Context, listener *elbv2types.Listener) error {
	_, err := c.elbv2.DeleteListener(ctx, &elbv2.DeleteListenerInput{ListenerArn: listener.ListenerArn})
	if err != nil {
		return fmt.Errorf("error deleting load balancer listener: %q", err)
	}
	_, err = c.elbv2.DeleteTargetGroup(ctx, &elbv2.DeleteTargetGroupInput{TargetGroupArn: listener.DefaultActions[0].TargetGroupArn})
	if err != nil {
		return fmt.Errorf("error deleting load balancer target group: %q", err)
	}
	return nil
}

// ensureTargetGroup creates a target group with a set of instances.
func (c *Cloud) ensureTargetGroup(ctx context.Context, targetGroup *elbv2types.TargetGroup, serviceName types.NamespacedName, mapping nlbPortMapping, instances []string, vpcID string, tags map[string]string) (*elbv2types.TargetGroup, error) {
	dirty := false
	expectedTargets := c.computeTargetGroupExpectedTargets(instances, mapping.TrafficPort)
	if targetGroup == nil {
		targetType := elbv2types.TargetTypeEnumInstance
		name := c.buildTargetGroupName(serviceName, mapping.FrontendPort, mapping.TrafficPort, mapping.TrafficProtocol, targetType, mapping)
		klog.Infof("Creating load balancer target group for %v with name: %s", serviceName, name)
		input := &elbv2.CreateTargetGroupInput{
			VpcId:                      aws.String(vpcID),
			Name:                       aws.String(name),
			Port:                       aws.Int32(mapping.TrafficPort),
			Protocol:                   mapping.TrafficProtocol,
			TargetType:                 targetType,
			HealthCheckIntervalSeconds: aws.Int32(mapping.HealthCheckConfig.Interval),
			HealthCheckPort:            aws.String(mapping.HealthCheckConfig.Port),
			HealthCheckProtocol:        mapping.HealthCheckConfig.Protocol,
			HealthyThresholdCount:      aws.Int32(mapping.HealthCheckConfig.HealthyThreshold),
			UnhealthyThresholdCount:    aws.Int32(mapping.HealthCheckConfig.UnhealthyThreshold),
			// HealthCheckTimeoutSeconds:  Currently not configurable, 6 seconds for HTTP, 10 for TCP/HTTPS
		}

		if mapping.HealthCheckConfig.Protocol != elbv2types.ProtocolEnumTcp {
			input.HealthCheckPath = aws.String(mapping.HealthCheckConfig.Path)
		}

		if len(tags) != 0 {
			targetGroupTags := make([]elbv2types.Tag, 0, len(tags))
			for k, v := range tags {
				targetGroupTags = append(targetGroupTags, elbv2types.Tag{
					Key: aws.String(k), Value: aws.String(v),
				})
			}
			input.Tags = targetGroupTags
		}
		result, err := c.elbv2.CreateTargetGroup(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("error creating load balancer target group: %q", err)
		}
		if len(result.TargetGroups) != 1 {
			return nil, fmt.Errorf("expected only one target group on CreateTargetGroup, got %d groups", len(result.TargetGroups))
		}

		tg := result.TargetGroups[0]
		tgARN := aws.ToString(tg.TargetGroupArn)
		if err := c.ensureTargetGroupTargets(ctx, tgARN, expectedTargets, nil); err != nil {
			return nil, err
		}
		return &tg, nil
	}

	// handle instances in service
	{
		tgARN := aws.ToString(targetGroup.TargetGroupArn)
		actualTargets, err := c.obtainTargetGroupActualTargets(ctx, tgARN)
		if err != nil {
			return nil, err
		}
		if err := c.ensureTargetGroupTargets(ctx, tgARN, expectedTargets, actualTargets); err != nil {
			return nil, err
		}
	}

	// ensure the health check is correct
	{
		dirtyHealthCheck := false

		input := &elbv2.ModifyTargetGroupInput{
			TargetGroupArn: targetGroup.TargetGroupArn,
		}
		if mapping.HealthCheckConfig.Port != aws.ToString(targetGroup.HealthCheckPort) {
			input.HealthCheckPort = aws.String(mapping.HealthCheckConfig.Port)
			dirtyHealthCheck = true
		}
		if mapping.HealthCheckConfig.HealthyThreshold != aws.ToInt32(targetGroup.HealthyThresholdCount) {
			dirtyHealthCheck = true
			input.HealthyThresholdCount = aws.Int32(mapping.HealthCheckConfig.HealthyThreshold)
			input.UnhealthyThresholdCount = aws.Int32(mapping.HealthCheckConfig.UnhealthyThreshold)
		}
		if !strings.EqualFold(string(mapping.HealthCheckConfig.Protocol), string(elbv2types.ProtocolEnumTcp)) {
			if mapping.HealthCheckConfig.Path != aws.ToString(input.HealthCheckPath) {
				input.HealthCheckPath = aws.String(mapping.HealthCheckConfig.Path)
				dirtyHealthCheck = true
			}
		}

		if dirtyHealthCheck {
			_, err := c.elbv2.ModifyTargetGroup(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("error modifying target group health check: %q", err)
			}

			dirty = true
		}
	}

	if dirty {
		result, err := c.elbv2.DescribeTargetGroups(ctx, &elbv2.DescribeTargetGroupsInput{
			TargetGroupArns: []string{aws.ToString(targetGroup.TargetGroupArn)},
		})
		if err != nil {
			return nil, fmt.Errorf("error retrieving target group after creation/update: %q", err)
		}
		targetGroup = &result.TargetGroups[0]
	}

	return targetGroup, nil
}

func (c *Cloud) ensureTargetGroupTargets(ctx context.Context, tgARN string, expectedTargets []*elbv2types.TargetDescription, actualTargets []*elbv2types.TargetDescription) error {
	targetsToRegister, targetsToDeregister := c.diffTargetGroupTargets(expectedTargets, actualTargets)
	if len(targetsToRegister) > 0 {
		targetsToRegisterChunks := c.chunkTargetDescriptions(targetsToRegister, defaultRegisterTargetsChunkSize)
		for _, targetsChunk := range targetsToRegisterChunks {
			req := &elbv2.RegisterTargetsInput{
				TargetGroupArn: aws.String(tgARN),
				Targets:        targetsChunk,
			}
			if _, err := c.elbv2.RegisterTargets(ctx, req); err != nil {
				return fmt.Errorf("error trying to register targets in target group: %q", err)
			}
		}
	}
	if len(targetsToDeregister) > 0 {
		targetsToDeregisterChunks := c.chunkTargetDescriptions(targetsToDeregister, defaultDeregisterTargetsChunkSize)
		for _, targetsChunk := range targetsToDeregisterChunks {
			req := &elbv2.DeregisterTargetsInput{
				TargetGroupArn: aws.String(tgARN),
				Targets:        targetsChunk,
			}
			if _, err := c.elbv2.DeregisterTargets(ctx, req); err != nil {
				return fmt.Errorf("error trying to deregister targets in target group: %q", err)
			}
		}
	}
	return nil
}

func (c *Cloud) computeTargetGroupExpectedTargets(instanceIDs []string, port int32) []*elbv2types.TargetDescription {
	expectedTargets := make([]*elbv2types.TargetDescription, 0, len(instanceIDs))
	for _, instanceID := range instanceIDs {
		expectedTargets = append(expectedTargets, &elbv2types.TargetDescription{
			Id:   aws.String(instanceID),
			Port: aws.Int32(port),
		})
	}
	return expectedTargets
}

func (c *Cloud) obtainTargetGroupActualTargets(ctx context.Context, tgARN string) ([]*elbv2types.TargetDescription, error) {
	req := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(tgARN),
	}
	resp, err := c.elbv2.DescribeTargetHealth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error describing target group health: %q", err)
	}
	actualTargets := make([]*elbv2types.TargetDescription, 0, len(resp.TargetHealthDescriptions))
	for _, targetDesc := range resp.TargetHealthDescriptions {
		if targetDesc.TargetHealth.Reason == elbv2types.TargetHealthReasonEnumDeregistrationInProgress {
			continue
		}
		actualTargets = append(actualTargets, targetDesc.Target)
	}
	return actualTargets, nil
}

// diffTargetGroupTargets computes the targets to register and targets to deregister based on existingTargets and desired instances.
func (c *Cloud) diffTargetGroupTargets(expectedTargets []*elbv2types.TargetDescription, actualTargets []*elbv2types.TargetDescription) (targetsToRegister []elbv2types.TargetDescription, targetsToDeregister []elbv2types.TargetDescription) {
	expectedTargetsByUID := make(map[string]elbv2types.TargetDescription, len(expectedTargets))
	for _, target := range expectedTargets {
		targetUID := fmt.Sprintf("%v:%v", aws.ToString(target.Id), aws.ToInt32(target.Port))
		expectedTargetsByUID[targetUID] = *target
	}
	actualTargetsByUID := make(map[string]elbv2types.TargetDescription, len(actualTargets))
	for _, target := range actualTargets {
		targetUID := fmt.Sprintf("%v:%v", aws.ToString(target.Id), aws.ToInt32(target.Port))
		actualTargetsByUID[targetUID] = *target
	}

	expectedTargetsUIDs := sets.StringKeySet(expectedTargetsByUID)
	actualTargetsUIDs := sets.StringKeySet(actualTargetsByUID)
	for _, targetUID := range expectedTargetsUIDs.Difference(actualTargetsUIDs).List() {
		targetsToRegister = append(targetsToRegister, expectedTargetsByUID[targetUID])
	}
	for _, targetUID := range actualTargetsUIDs.Difference(expectedTargetsUIDs).List() {
		targetsToDeregister = append(targetsToDeregister, actualTargetsByUID[targetUID])
	}
	return targetsToRegister, targetsToDeregister
}

// chunkTargetDescriptions will split slice of TargetDescription into chunks
func (c *Cloud) chunkTargetDescriptions(targets []elbv2types.TargetDescription, chunkSize int) [][]elbv2types.TargetDescription {
	var chunks [][]elbv2types.TargetDescription
	for i := 0; i < len(targets); i += chunkSize {
		end := i + chunkSize
		if end > len(targets) {
			end = len(targets)
		}
		chunks = append(chunks, targets[i:end])
	}
	return chunks
}

// updateInstanceSecurityGroupsForNLB will adjust securityGroup's settings to allow inbound traffic into instances from clientCIDRs and portMappings.
// TIP: if either instances or clientCIDRs or portMappings are nil, then the securityGroup rules for lbName are cleared.
func (c *Cloud) updateInstanceSecurityGroupsForNLB(ctx context.Context, lbName string, instances map[InstanceID]*ec2types.Instance, subnetCIDRs []string, clientCIDRs []string, portMappings []nlbPortMapping) error {
	if c.cfg.Global.DisableSecurityGroupIngress {
		return nil
	}

	clusterSGs, err := c.getTaggedSecurityGroups(ctx)
	if err != nil {
		return fmt.Errorf("error querying for tagged security groups: %q", err)
	}
	// scan instances for groups we want to open
	desiredSGIDs := sets.String{}
	for _, instance := range instances {
		sg, err := findSecurityGroupForInstance(instance, clusterSGs)
		if err != nil {
			return err
		}
		if sg == nil {
			klog.Warningf("Ignoring instance without security group: %s", aws.ToString(instance.InstanceId))
			continue
		}
		desiredSGIDs.Insert(aws.ToString(sg.GroupId))
	}

	// TODO(@M00nF1sh): do we really needs to support SG without cluster tag at current version?
	// findSecurityGroupForInstance might return SG that are not tagged.
	{
		for sgID := range desiredSGIDs.Difference(sets.StringKeySet(clusterSGs)) {
			sg, err := c.findSecurityGroup(ctx, sgID)
			if err != nil {
				return fmt.Errorf("error finding instance group: %q", err)
			}
			if sg == nil {
				return fmt.Errorf("error finding security group: %s", sgID)
			}
			clusterSGs[sgID] = sg
		}
	}

	{
		clientPorts := sets.Set[int32]{}
		clientProtocol := "tcp"
		healthCheckPorts := sets.Set[int32]{}
		for _, port := range portMappings {
			clientPorts.Insert(port.TrafficPort)
			hcPort := port.TrafficPort
			if port.HealthCheckConfig.Port != defaultHealthCheckPort {
				hcPort64, err := strconv.ParseInt(port.HealthCheckConfig.Port, 10, 0)
				if err != nil {
					return fmt.Errorf("Invalid health check port %v", port.HealthCheckConfig.Port)
				}
				hcPort = int32(hcPort64)
			}
			healthCheckPorts.Insert(hcPort)
			if port.TrafficProtocol == elbv2types.ProtocolEnumUdp {
				clientProtocol = "udp"
			}
		}
		clientRuleAnnotation := fmt.Sprintf("%s=%s", NLBClientRuleDescription, lbName)
		healthRuleAnnotation := fmt.Sprintf("%s=%s", NLBHealthCheckRuleDescription, lbName)
		for sgID, sg := range clusterSGs {
			sgPerms := NewIPPermissionSet(sg.IpPermissions...).Ungroup()
			if desiredSGIDs.Has(sgID) {
				// If the client rule is 1) all addresses 2) tcp and 3) has same ports as the healthcheck,
				// then the health rules are a subset of the client rule and are not needed.
				if len(clientCIDRs) != 1 || clientCIDRs[0] != "0.0.0.0/0" || clientProtocol != "tcp" || !healthCheckPorts.Equal(clientPorts) {
					if err := c.updateInstanceSecurityGroupForNLBTraffic(ctx, sgID, sgPerms, healthRuleAnnotation, "tcp", healthCheckPorts, subnetCIDRs); err != nil {
						return err
					}
				}
				if err := c.updateInstanceSecurityGroupForNLBTraffic(ctx, sgID, sgPerms, clientRuleAnnotation, clientProtocol, clientPorts, clientCIDRs); err != nil {
					return err
				}
			} else {
				if err := c.updateInstanceSecurityGroupForNLBTraffic(ctx, sgID, sgPerms, healthRuleAnnotation, "tcp", nil, nil); err != nil {
					return err
				}
				if err := c.updateInstanceSecurityGroupForNLBTraffic(ctx, sgID, sgPerms, clientRuleAnnotation, clientProtocol, nil, nil); err != nil {
					return err
				}
			}
			if !sgPerms.Equal(NewIPPermissionSet(sg.IpPermissions...).Ungroup()) {
				if err := c.updateInstanceSecurityGroupForNLBMTU(ctx, sgID, sgPerms); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// updateInstanceSecurityGroupForNLBTraffic will manage permissions set(identified by ruleDesc) on securityGroup to match desired set(allow protocol traffic from ports/cidr).
// Note: sgPerms will be updated to reflect the current permission set on SG after update.
func (c *Cloud) updateInstanceSecurityGroupForNLBTraffic(ctx context.Context, sgID string, sgPerms IPPermissionSet, ruleDesc string, protocol string, ports sets.Set[int32], cidrs []string) error {
	desiredPerms := NewIPPermissionSet()
	for port := range ports {
		for _, cidr := range cidrs {
			desiredPerms.Insert(ec2types.IpPermission{
				IpProtocol: aws.String(protocol),
				FromPort:   aws.Int32(int32(port)),
				ToPort:     aws.Int32(int32(port)),
				IpRanges: []ec2types.IpRange{
					{
						CidrIp:      aws.String(cidr),
						Description: aws.String(ruleDesc),
					},
				},
			})
		}
	}

	permsToGrant := desiredPerms.Difference(sgPerms)
	permsToRevoke := sgPerms.Difference(desiredPerms)
	permsToRevoke.DeleteIf(IPPermissionNotMatch{IPPermissionMatchDesc{ruleDesc}})
	if len(permsToRevoke) > 0 {
		permsToRevokeList := permsToRevoke.List()
		changed, err := c.removeSecurityGroupIngress(ctx, sgID, permsToRevokeList)
		if err != nil {
			klog.Warningf("Error remove traffic permission from security group: %q", err)
			return err
		}
		if !changed {
			klog.Warning("Revoking ingress was not needed; concurrent change? groupId=", sgID)
		}
		sgPerms.Delete(permsToRevokeList...)
	}
	if len(permsToGrant) > 0 {
		permsToGrantList := permsToGrant.List()
		changed, err := c.addSecurityGroupIngress(ctx, sgID, permsToGrantList)
		if err != nil {
			klog.Warningf("Error add traffic permission to security group: %q", err)
			return err
		}
		if !changed {
			klog.Warning("Allowing ingress was not needed; concurrent change? groupId=", sgID)
		}
		sgPerms.Insert(permsToGrantList...)
	}
	return nil
}

// Note: sgPerms will be updated to reflect the current permission set on SG after update.
func (c *Cloud) updateInstanceSecurityGroupForNLBMTU(ctx context.Context, sgID string, sgPerms IPPermissionSet) error {
	desiredPerms := NewIPPermissionSet()
	for _, perm := range sgPerms {
		for _, ipRange := range perm.IpRanges {
			if strings.Contains(aws.ToString(ipRange.Description), NLBClientRuleDescription) {
				desiredPerms.Insert(ec2types.IpPermission{
					IpProtocol: aws.String("icmp"),
					FromPort:   aws.Int32(3),
					ToPort:     aws.Int32(4),
					IpRanges: []ec2types.IpRange{
						{
							CidrIp:      ipRange.CidrIp,
							Description: aws.String(NLBMtuDiscoveryRuleDescription),
						},
					},
				})
			}
		}
	}

	permsToGrant := desiredPerms.Difference(sgPerms)
	permsToRevoke := sgPerms.Difference(desiredPerms)
	permsToRevoke.DeleteIf(IPPermissionNotMatch{IPPermissionMatchDesc{NLBMtuDiscoveryRuleDescription}})
	if len(permsToRevoke) > 0 {
		permsToRevokeList := permsToRevoke.List()
		changed, err := c.removeSecurityGroupIngress(ctx, sgID, permsToRevokeList)
		if err != nil {
			klog.Warningf("Error remove MTU permission from security group: %q", err)
			return err
		}
		if !changed {
			klog.Warning("Revoking ingress was not needed; concurrent change? groupId=", sgID)
		}

		sgPerms.Delete(permsToRevokeList...)
	}
	if len(permsToGrant) > 0 {
		permsToGrantList := permsToGrant.List()
		changed, err := c.addSecurityGroupIngress(ctx, sgID, permsToGrantList)
		if err != nil {
			klog.Warningf("Error add MTU permission to security group: %q", err)
			return err
		}
		if !changed {
			klog.Warning("Allowing ingress was not needed; concurrent change? groupId=", sgID)
		}
		sgPerms.Insert(permsToGrantList...)
	}
	return nil
}

func (c *Cloud) ensureLoadBalancer(ctx context.Context, namespacedName types.NamespacedName, loadBalancerName string, listeners []elbtypes.Listener, subnetIDs []string, securityGroupIDs []string, internalELB, proxyProtocol bool, loadBalancerAttributes *elbtypes.LoadBalancerAttributes, annotations map[string]string) (*elbtypes.LoadBalancerDescription, error) {
	loadBalancer, err := c.describeLoadBalancer(ctx, loadBalancerName)
	if err != nil {
		return nil, err
	}

	dirty := false

	if loadBalancer == nil {
		createRequest := &elb.CreateLoadBalancerInput{}
		createRequest.LoadBalancerName = aws.String(loadBalancerName)

		createRequest.Listeners = listeners

		if internalELB {
			createRequest.Scheme = aws.String("internal")
		}

		// We are supposed to specify one subnet per AZ.
		// TODO: What happens if we have more than one subnet per AZ?
		if subnetIDs == nil {
			createRequest.Subnets = nil
		} else {
			createRequest.Subnets = subnetIDs
		}

		if securityGroupIDs == nil {
			createRequest.SecurityGroups = nil
		} else {
			createRequest.SecurityGroups = securityGroupIDs
		}

		// Get additional tags set by the user
		tags := getKeyValuePropertiesFromAnnotation(annotations, ServiceAnnotationLoadBalancerAdditionalTags)

		// Add default tags
		tags[TagNameKubernetesService] = namespacedName.String()
		tags = c.tagging.buildTags(ResourceLifecycleOwned, tags)

		for k, v := range tags {
			createRequest.Tags = append(createRequest.Tags, elbtypes.Tag{
				Key: aws.String(k), Value: aws.String(v),
			})
		}

		klog.Infof("Creating load balancer for %v with name: %s", namespacedName, loadBalancerName)
		_, err := c.elb.CreateLoadBalancer(ctx, createRequest)
		if err != nil {
			return nil, err
		}

		if proxyProtocol {
			err = c.createProxyProtocolPolicy(ctx, loadBalancerName)
			if err != nil {
				return nil, err
			}

			for _, listener := range listeners {
				klog.V(2).Infof("Adjusting AWS loadbalancer proxy protocol on node port %d. Setting to true", *listener.InstancePort)
				err := c.setBackendPolicies(ctx, loadBalancerName, listener.InstancePort, []*string{aws.String(ProxyProtocolPolicyName)})
				if err != nil {
					return nil, err
				}
			}
		}

		dirty = true
	} else {
		// TODO: Sync internal vs non-internal

		{
			// Sync subnets
			expected := sets.New[string](subnetIDs...)
			actual := sets.New[string](loadBalancer.Subnets...)

			additions := expected.Difference(actual)
			removals := actual.Difference(expected)

			if removals.Len() != 0 {
				request := &elb.DetachLoadBalancerFromSubnetsInput{}
				request.LoadBalancerName = aws.String(loadBalancerName)
				request.Subnets = stringSetToList(removals)
				klog.V(2).Info("Detaching load balancer from removed subnets")
				_, err := c.elb.DetachLoadBalancerFromSubnets(ctx, request)
				if err != nil {
					return nil, fmt.Errorf("error detaching AWS loadbalancer from subnets: %q", err)
				}
				dirty = true
			}

			if additions.Len() != 0 {
				request := &elb.AttachLoadBalancerToSubnetsInput{}
				request.LoadBalancerName = aws.String(loadBalancerName)
				request.Subnets = stringSetToList(additions)
				klog.V(2).Info("Attaching load balancer to added subnets")
				_, err := c.elb.AttachLoadBalancerToSubnets(ctx, request)
				if err != nil {
					return nil, fmt.Errorf("error attaching AWS loadbalancer to subnets: %q", err)
				}
				dirty = true
			}
		}

		{
			// Sync security groups
			expected := sets.New[string](securityGroupIDs...)
			actual := stringSetFromList(loadBalancer.SecurityGroups)

			if !expected.Equal(actual) {
				// This call just replaces the security groups, unlike e.g. subnets (!)
				request := &elb.ApplySecurityGroupsToLoadBalancerInput{}
				request.LoadBalancerName = aws.String(loadBalancerName)
				if securityGroupIDs == nil {
					request.SecurityGroups = nil
				} else {
					request.SecurityGroups = securityGroupIDs
				}
				klog.V(2).Info("Applying updated security groups to load balancer")
				_, err := c.elb.ApplySecurityGroupsToLoadBalancer(ctx, request)
				if err != nil {
					return nil, fmt.Errorf("error applying AWS loadbalancer security groups: %q", err)
				}
				dirty = true
			}
		}

		{
			additions, removals := syncElbListeners(loadBalancerName, listeners, loadBalancer.ListenerDescriptions)

			if len(removals) != 0 {
				request := &elb.DeleteLoadBalancerListenersInput{}
				request.LoadBalancerName = aws.String(loadBalancerName)
				request.LoadBalancerPorts = removals
				klog.V(2).Info("Deleting removed load balancer listeners")
				if _, err := c.elb.DeleteLoadBalancerListeners(ctx, request); err != nil {
					return nil, fmt.Errorf("error deleting AWS loadbalancer listeners: %q", err)
				}
				dirty = true
			}

			if len(additions) != 0 {
				request := &elb.CreateLoadBalancerListenersInput{}
				request.LoadBalancerName = aws.String(loadBalancerName)
				request.Listeners = additions
				klog.V(2).Info("Creating added load balancer listeners")
				if _, err := c.elb.CreateLoadBalancerListeners(ctx, request); err != nil {
					return nil, fmt.Errorf("error creating AWS loadbalancer listeners: %q", err)
				}
				dirty = true
			}
		}

		{
			// Sync proxy protocol state for new and existing listeners

			proxyPolicies := make([]*string, 0)
			if proxyProtocol {
				// Ensure the backend policy exists

				// NOTE The documentation for the AWS API indicates we could get an HTTP 400
				// back if a policy of the same name already exists. However, the aws-sdk does not
				// seem to return an error to us in these cases. Therefore, this will issue an API
				// request every time.
				err := c.createProxyProtocolPolicy(ctx, loadBalancerName)
				if err != nil {
					return nil, err
				}

				proxyPolicies = append(proxyPolicies, aws.String(ProxyProtocolPolicyName))
			}

			foundBackends := make(map[int32]bool)
			proxyProtocolBackends := make(map[int32]bool)
			for _, backendListener := range loadBalancer.BackendServerDescriptions {
				foundBackends[aws.ToInt32(backendListener.InstancePort)] = false
				proxyProtocolBackends[aws.ToInt32(backendListener.InstancePort)] = proxyProtocolEnabled(backendListener)
			}

			for _, listener := range listeners {
				setPolicy := false
				instancePort := *listener.InstancePort

				if currentState, ok := proxyProtocolBackends[instancePort]; !ok {
					// This is a new ELB backend so we only need to worry about
					// potentially adding a policy and not removing an
					// existing one
					setPolicy = proxyProtocol
				} else {
					foundBackends[instancePort] = true
					// This is an existing ELB backend so we need to determine
					// if the state changed
					setPolicy = (currentState != proxyProtocol)
				}

				if setPolicy {
					klog.V(2).Infof("Adjusting AWS loadbalancer proxy protocol on node port %d. Setting to %t", instancePort, proxyProtocol)
					err := c.setBackendPolicies(ctx, loadBalancerName, aws.Int32(instancePort), proxyPolicies)
					if err != nil {
						return nil, err
					}
					dirty = true
				}
			}

			// We now need to figure out if any backend policies need removed
			// because these old policies will stick around even if there is no
			// corresponding listener anymore
			for instancePort, found := range foundBackends {
				if !found {
					klog.V(2).Infof("Adjusting AWS loadbalancer proxy protocol on node port %d. Setting to false", instancePort)
					err := c.setBackendPolicies(ctx, loadBalancerName, aws.Int32(instancePort), []*string{})
					if err != nil {
						return nil, err
					}
					dirty = true
				}
			}
		}

		{
			// Add additional tags
			klog.V(2).Infof("Creating additional load balancer tags for %s", loadBalancerName)
			tags := getKeyValuePropertiesFromAnnotation(annotations, ServiceAnnotationLoadBalancerAdditionalTags)
			if len(tags) > 0 {
				err := c.addLoadBalancerTags(ctx, loadBalancerName, tags)
				if err != nil {
					return nil, fmt.Errorf("unable to create additional load balancer tags: %v", err)
				}
			}
		}
	}

	// Whether the ELB was new or existing, sync attributes regardless. This accounts for things
	// that cannot be specified at the time of creation and can only be modified after the fact,
	// e.g. idle connection timeout.
	{
		describeAttributesRequest := &elb.DescribeLoadBalancerAttributesInput{}
		describeAttributesRequest.LoadBalancerName = aws.String(loadBalancerName)
		describeAttributesOutput, err := c.elb.DescribeLoadBalancerAttributes(ctx, describeAttributesRequest)
		if err != nil {
			klog.Warning("Unable to retrieve load balancer attributes during attribute sync")
			return nil, err
		}

		foundAttributes := &describeAttributesOutput.LoadBalancerAttributes

		// Update attributes if they're dirty
		if !reflect.DeepEqual(loadBalancerAttributes, foundAttributes) {
			klog.V(2).Infof("Updating load-balancer attributes for %q", loadBalancerName)

			modifyAttributesRequest := &elb.ModifyLoadBalancerAttributesInput{}
			modifyAttributesRequest.LoadBalancerName = aws.String(loadBalancerName)
			modifyAttributesRequest.LoadBalancerAttributes = loadBalancerAttributes
			_, err = c.elb.ModifyLoadBalancerAttributes(ctx, modifyAttributesRequest)
			if err != nil {
				return nil, fmt.Errorf("Unable to update load balancer attributes during attribute sync: %q", err)
			}
			dirty = true
		}
	}

	if dirty {
		loadBalancer, err = c.describeLoadBalancer(ctx, loadBalancerName)
		if err != nil {
			klog.Warning("Unable to retrieve load balancer after creation/update")
			return nil, err
		}
	}

	return loadBalancer, nil
}

// syncElbListeners computes a plan to reconcile the desired vs actual state of the listeners on an ELB
// NOTE: there exists an O(nlgn) implementation for this function. However, as the default limit of
//
//	listeners per elb is 100, this implementation is reduced from O(m*n) => O(n).
func syncElbListeners(loadBalancerName string, listeners []elbtypes.Listener, listenerDescriptions []elbtypes.ListenerDescription) ([]elbtypes.Listener, []int32) {
	foundSet := make(map[int]bool)
	removals := []int32{}
	additions := []elbtypes.Listener{}

	for _, listenerDescription := range listenerDescriptions {
		actual := listenerDescription.Listener
		if actual == nil {
			klog.Warning("Ignoring empty listener in AWS loadbalancer: ", loadBalancerName)
			continue
		}

		found := false
		for i, expected := range listeners {
			if elbListenersAreEqual(*actual, expected) {
				// The current listener on the actual
				// elb is in the set of desired listeners.
				foundSet[i] = true
				found = true
				break
			}
		}
		if !found {
			removals = append(removals, actual.LoadBalancerPort)
		}
	}

	for i := range listeners {
		if !foundSet[i] {
			additions = append(additions, listeners[i])
		}
	}

	return additions, removals
}

func elbListenersAreEqual(actual, expected elbtypes.Listener) bool {
	if !elbProtocolsAreEqual(actual.Protocol, expected.Protocol) {
		return false
	}
	if !elbProtocolsAreEqual(actual.InstanceProtocol, expected.InstanceProtocol) {
		return false
	}
	if aws.ToInt32(actual.InstancePort) != aws.ToInt32(expected.InstancePort) {
		return false
	}
	if actual.LoadBalancerPort != expected.LoadBalancerPort {
		return false
	}
	if !awsArnEquals(actual.SSLCertificateId, expected.SSLCertificateId) {
		return false
	}
	return true
}

func createSubnetMappings(subnetIDs []string, allocationIDs []string) []elbv2types.SubnetMapping {
	response := []elbv2types.SubnetMapping{}

	for index, id := range subnetIDs {
		sm := elbv2types.SubnetMapping{SubnetId: aws.String(id)}
		if len(allocationIDs) > 0 {
			sm.AllocationId = aws.String(allocationIDs[index])
		}
		response = append(response, sm)
	}

	return response
}

// elbProtocolsAreEqual checks if two ELB protocol strings are considered the same
// Comparison is case insensitive
func elbProtocolsAreEqual(l, r *string) bool {
	if l == nil || r == nil {
		return l == r
	}
	return strings.EqualFold(aws.ToString(l), aws.ToString(r))
}

// awsArnEquals checks if two ARN strings are considered the same
// Comparison is case insensitive
func awsArnEquals(l, r *string) bool {
	if l == nil || r == nil {
		return l == r
	}
	return strings.EqualFold(aws.ToString(l), aws.ToString(r))
}

// getExpectedHealthCheck returns an elb.Healthcheck for the provided target
// and using either sensible defaults or overrides via Service annotations
func (c *Cloud) getExpectedHealthCheck(target string, annotations map[string]string) (*elbtypes.HealthCheck, error) {
	healthcheck := &elbtypes.HealthCheck{Target: &target}
	getOrDefault := func(annotation string, defaultValue int32) (*int32, error) {
		i32 := defaultValue
		if s, ok := annotations[annotation]; ok {
			i64, err := strconv.ParseInt(s, 10, 0)
			if err != nil {
				return nil, fmt.Errorf("failed parsing health check annotation value: %v", err)
			}
			i32 = int32(i64)
		}
		return &i32, nil
	}
	var err error
	healthcheck.HealthyThreshold, err = getOrDefault(ServiceAnnotationLoadBalancerHCHealthyThreshold, defaultElbHCHealthyThreshold)
	if err != nil {
		return nil, err
	}
	healthcheck.UnhealthyThreshold, err = getOrDefault(ServiceAnnotationLoadBalancerHCUnhealthyThreshold, defaultElbHCUnhealthyThreshold)
	if err != nil {
		return nil, err
	}
	healthcheck.Timeout, err = getOrDefault(ServiceAnnotationLoadBalancerHCTimeout, defaultElbHCTimeout)
	if err != nil {
		return nil, err
	}
	healthcheck.Interval, err = getOrDefault(ServiceAnnotationLoadBalancerHCInterval, defaultElbHCInterval)
	if err != nil {
		return nil, err
	}
	if err = ValidateHealthCheck(healthcheck); err != nil {
		return nil, fmt.Errorf("some of the load balancer health check parameters are invalid: %v", err)
	}
	return healthcheck, nil
}

// Makes sure that the health check for an ELB matches the configured health check node port
func (c *Cloud) ensureLoadBalancerHealthCheck(ctx context.Context, loadBalancer *elbtypes.LoadBalancerDescription, protocol string, port int32, path string, annotations map[string]string) error {
	name := aws.ToString(loadBalancer.LoadBalancerName)

	actual := loadBalancer.HealthCheck
	// Override healthcheck protocol, port and path based on annotations
	if s, ok := annotations[ServiceAnnotationLoadBalancerHealthCheckProtocol]; ok {
		protocol = s
	}
	if s, ok := annotations[ServiceAnnotationLoadBalancerHealthCheckPort]; ok && s != defaultHealthCheckPort {
		p, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return err
		}
		port = int32(p)
	}
	switch strings.ToUpper(protocol) {
	case "HTTP", "HTTPS":
		if path == "" {
			path = defaultHealthCheckPath
		}
		if s := annotations[ServiceAnnotationLoadBalancerHealthCheckPath]; s != "" {
			path = s
		}
	default:
		path = ""
	}
	expectedTarget := protocol + ":" + strconv.FormatInt(int64(port), 10) + path
	expected, err := c.getExpectedHealthCheck(expectedTarget, annotations)
	if err != nil {
		return fmt.Errorf("cannot update health check for load balancer %q: %q", name, err)
	}

	// comparing attributes 1 by 1 to avoid breakage in case a new field is
	// added to the HC which breaks the equality
	if aws.ToString(expected.Target) == aws.ToString(actual.Target) &&
		aws.ToInt32(expected.HealthyThreshold) == aws.ToInt32(actual.HealthyThreshold) &&
		aws.ToInt32(expected.UnhealthyThreshold) == aws.ToInt32(actual.UnhealthyThreshold) &&
		aws.ToInt32(expected.Interval) == aws.ToInt32(actual.Interval) &&
		aws.ToInt32(expected.Timeout) == aws.ToInt32(actual.Timeout) {
		return nil
	}

	request := &elb.ConfigureHealthCheckInput{}
	request.HealthCheck = expected
	request.LoadBalancerName = loadBalancer.LoadBalancerName

	_, err = c.elb.ConfigureHealthCheck(ctx, request)
	if err != nil {
		return fmt.Errorf("error configuring load balancer health check for %q: %q", name, err)
	}

	return nil
}

// Makes sure that exactly the specified hosts are registered as instances with the load balancer
func (c *Cloud) ensureLoadBalancerInstances(ctx context.Context, loadBalancerName string, lbInstances []elbtypes.Instance, instanceIDs map[InstanceID]*ec2types.Instance) error {
	expected := sets.NewString()
	for id := range instanceIDs {
		expected.Insert(string(id))
	}

	actual := sets.NewString()
	for _, lbInstance := range lbInstances {
		actual.Insert(aws.ToString(lbInstance.InstanceId))
	}

	additions := expected.Difference(actual)
	removals := actual.Difference(expected)

	addInstances := []elbtypes.Instance{}
	for _, instanceID := range additions.List() {
		addInstance := elbtypes.Instance{}
		addInstance.InstanceId = aws.String(instanceID)
		addInstances = append(addInstances, addInstance)
	}

	removeInstances := []elbtypes.Instance{}
	for _, instanceID := range removals.List() {
		removeInstance := elbtypes.Instance{}
		removeInstance.InstanceId = aws.String(instanceID)
		removeInstances = append(removeInstances, removeInstance)
	}

	if len(addInstances) > 0 {
		registerRequest := &elb.RegisterInstancesWithLoadBalancerInput{}
		registerRequest.Instances = addInstances
		registerRequest.LoadBalancerName = aws.String(loadBalancerName)
		_, err := c.elb.RegisterInstancesWithLoadBalancer(ctx, registerRequest)
		if err != nil {
			return err
		}
		klog.V(1).Infof("Instances added to load-balancer %s", loadBalancerName)
	}

	if len(removeInstances) > 0 {
		deregisterRequest := &elb.DeregisterInstancesFromLoadBalancerInput{}
		deregisterRequest.Instances = removeInstances
		deregisterRequest.LoadBalancerName = aws.String(loadBalancerName)
		_, err := c.elb.DeregisterInstancesFromLoadBalancer(ctx, deregisterRequest)
		if err != nil {
			return err
		}
		klog.V(1).Infof("Instances removed from load-balancer %s", loadBalancerName)
	}

	return nil
}

func (c *Cloud) getLoadBalancerTLSPorts(loadBalancer *elbtypes.LoadBalancerDescription) []int64 {
	ports := []int64{}

	for _, listenerDescription := range loadBalancer.ListenerDescriptions {
		protocol := aws.ToString(listenerDescription.Listener.Protocol)
		if protocol == "SSL" || protocol == "HTTPS" {
			ports = append(ports, int64(listenerDescription.Listener.LoadBalancerPort))
		}
	}
	return ports
}

func (c *Cloud) ensureSSLNegotiationPolicy(ctx context.Context, loadBalancer *elbtypes.LoadBalancerDescription, policyName string) error {
	klog.V(2).Info("Describing load balancer policies on load balancer")
	result, err := c.elb.DescribeLoadBalancerPolicies(ctx, &elb.DescribeLoadBalancerPoliciesInput{
		LoadBalancerName: loadBalancer.LoadBalancerName,
		PolicyNames: []string{
			fmt.Sprintf(SSLNegotiationPolicyNameFormat, policyName),
		},
	})
	if err != nil {
		// If DescribeLoadBalancerPolicies returns a PolicyNotFoundException, we must proceed and create the policy.
		var notFoundErr *elbtypes.PolicyNotFoundException
		if !errors.As(err, &notFoundErr) {
			return fmt.Errorf("error describing security policies on load balancer: %q", err)
		}
	}

	// If DescribeLoadBalancerPolicies yielded a PolicyNotFoundException, result will be nil,
	// so we must check before dereferencing
	if result != nil && len(result.PolicyDescriptions) > 0 {
		return nil
	}

	klog.V(2).Infof("Creating SSL negotiation policy '%s' on load balancer", fmt.Sprintf(SSLNegotiationPolicyNameFormat, policyName))
	// there is an upper limit of 98 policies on an ELB, we're pretty safe from
	// running into it
	_, err = c.elb.CreateLoadBalancerPolicy(ctx, &elb.CreateLoadBalancerPolicyInput{
		LoadBalancerName: loadBalancer.LoadBalancerName,
		PolicyName:       aws.String(fmt.Sprintf(SSLNegotiationPolicyNameFormat, policyName)),
		PolicyTypeName:   aws.String("SSLNegotiationPolicyType"),
		PolicyAttributes: []elbtypes.PolicyAttribute{
			{
				AttributeName:  aws.String("Reference-Security-Policy"),
				AttributeValue: aws.String(policyName),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating security policy on load balancer: %q", err)
	}
	return nil
}

func (c *Cloud) setSSLNegotiationPolicy(ctx context.Context, loadBalancerName, sslPolicyName string, port int64) error {
	policyName := fmt.Sprintf(SSLNegotiationPolicyNameFormat, sslPolicyName)
	request := &elb.SetLoadBalancerPoliciesOfListenerInput{
		LoadBalancerName: aws.String(loadBalancerName),
		LoadBalancerPort: int32(port),
		PolicyNames:      []string{policyName},
	}
	klog.V(2).Infof("Setting SSL negotiation policy '%s' on load balancer", policyName)
	_, err := c.elb.SetLoadBalancerPoliciesOfListener(ctx, request)
	if err != nil {
		return fmt.Errorf("error setting SSL negotiation policy '%s' on load balancer: %q", policyName, err)
	}
	return nil
}

func (c *Cloud) createProxyProtocolPolicy(ctx context.Context, loadBalancerName string) error {
	request := &elb.CreateLoadBalancerPolicyInput{
		LoadBalancerName: aws.String(loadBalancerName),
		PolicyName:       aws.String(ProxyProtocolPolicyName),
		PolicyTypeName:   aws.String("ProxyProtocolPolicyType"),
		PolicyAttributes: []elbtypes.PolicyAttribute{
			{
				AttributeName:  aws.String("ProxyProtocol"),
				AttributeValue: aws.String("true"),
			},
		},
	}
	klog.V(2).Info("Creating proxy protocol policy on load balancer")
	_, err := c.elb.CreateLoadBalancerPolicy(ctx, request)
	if err != nil {
		return fmt.Errorf("error creating proxy protocol policy on load balancer: %q", err)
	}

	return nil
}

func (c *Cloud) setBackendPolicies(ctx context.Context, loadBalancerName string, instancePort *int32, policies []*string) error {
	request := &elb.SetLoadBalancerPoliciesForBackendServerInput{
		InstancePort:     instancePort,
		LoadBalancerName: aws.String(loadBalancerName),
		PolicyNames:      aws.ToStringSlice(policies),
	}
	if len(policies) > 0 {
		klog.V(2).Infof("Adding AWS loadbalancer backend policies on node port %d", instancePort)
	} else {
		klog.V(2).Infof("Removing AWS loadbalancer backend policies on node port %d", instancePort)
	}
	_, err := c.elb.SetLoadBalancerPoliciesForBackendServer(ctx, request)
	if err != nil {
		return fmt.Errorf("error adjusting AWS loadbalancer backend policies: %q", err)
	}

	return nil
}

func proxyProtocolEnabled(backend elbtypes.BackendServerDescription) bool {
	for _, policy := range backend.PolicyNames {
		if policy == ProxyProtocolPolicyName {
			return true
		}
	}

	return false
}

// findInstancesForELB gets the EC2 instances corresponding to the Nodes, for setting up an ELB
// We ignore Nodes (with a log message) where the instanceid cannot be determined from the provider,
// and we ignore instances which are not found
func (c *Cloud) findInstancesForELB(ctx context.Context, nodes []*v1.Node, annotations map[string]string) (map[InstanceID]*ec2types.Instance, error) {

	targetNodes := filterTargetNodes(nodes, annotations)

	// Map to instance ids ignoring Nodes where we cannot find the id (but logging)
	instanceIDs := mapToAWSInstanceIDsTolerant(targetNodes)

	cacheCriteria := cacheCriteria{
		MaxAge:       defaultEC2InstanceCacheMaxAge,
		HasInstances: instanceIDs, // Refresh if any of the instance ids are missing
	}
	snapshot, err := c.instanceCache.describeAllInstancesCached(ctx, cacheCriteria)
	if err != nil {
		return nil, err
	}

	instances := snapshot.FindInstances(instanceIDs)
	// We ignore instances that cannot be found

	return instances, nil
}

// filterTargetNodes uses node labels to filter the nodes that should be targeted by the ELB,
// checking if all the labels provided in an annotation are present in the nodes
func filterTargetNodes(nodes []*v1.Node, annotations map[string]string) []*v1.Node {

	targetNodeLabels := getKeyValuePropertiesFromAnnotation(annotations, ServiceAnnotationLoadBalancerTargetNodeLabels)

	if len(targetNodeLabels) == 0 {
		return nodes
	}

	targetNodes := make([]*v1.Node, 0, len(nodes))

	for _, node := range nodes {
		if node.Labels != nil && len(node.Labels) > 0 {
			allFiltersMatch := true

			for targetLabelKey, targetLabelValue := range targetNodeLabels {
				if nodeLabelValue, ok := node.Labels[targetLabelKey]; !ok || (nodeLabelValue != targetLabelValue && targetLabelValue != "") {
					allFiltersMatch = false
					break
				}
			}

			if allFiltersMatch {
				targetNodes = append(targetNodes, node)
			}
		}
	}

	return targetNodes
}

// ValidateHealthCheck replaces ELB.HealthCheck.Validate() from AWS SDK Go V1, which has been deprecated in V2
// V1 implementation: https://github.com/aws/aws-sdk-go/blob/v1.55.7/service/elb/api.go#L5346
func ValidateHealthCheck(s *elbtypes.HealthCheck) error {
	var validationErrors []string

	if s == nil {
		validationErrors = append(validationErrors, "HealthCheck is nil")
		return fmt.Errorf("HealthCheck validation errors: %s", strings.Join(validationErrors, "; "))
	}

	if s.HealthyThreshold == nil {
		validationErrors = append(validationErrors, "HealthyThreshold is required")
	} else if *s.HealthyThreshold < 2 {
		validationErrors = append(validationErrors, "HealthyThreshold must be at least 2")
	}

	if s.Interval == nil {
		validationErrors = append(validationErrors, "Interval is required")
	} else if *s.Interval < 5 {
		validationErrors = append(validationErrors, "Interval must be at least 5")
	}

	if s.Target == nil {
		validationErrors = append(validationErrors, "Target is required")
	}

	if s.Timeout == nil {
		validationErrors = append(validationErrors, "Timeout is required")
	} else if *s.Timeout < 2 {
		validationErrors = append(validationErrors, "Timeout must be at least 2")
	}

	if s.UnhealthyThreshold == nil {
		validationErrors = append(validationErrors, "UnhealthyThreshold is required")
	} else if *s.UnhealthyThreshold < 2 {
		validationErrors = append(validationErrors, "UnhealthyThreshold must be at least 2")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("HealthCheck validation errors: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}
