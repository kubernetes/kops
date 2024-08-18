/*
Copyright 2024 The Kubernetes Authors.

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

package karpenter

import (
	"fmt"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	karpenterawsapis "github.com/aws/karpenter-provider-aws/pkg/apis"
	karpenterawsv1 "github.com/aws/karpenter-provider-aws/pkg/apis/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cloudprovider "k8s.io/cloud-provider/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/upup/pkg/fi"
	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/utils/resources"
)

type KarpenterModelBuilder struct {
	awsmodel.AWSModelContext
}

func (k KarpenterModelBuilder) EC2NodeClass(ig *kops.InstanceGroup, image *ec2types.Image) (*karpenterawsv1.EC2NodeClass, error) {
	tags, err := k.CloudTagsForInstanceGroup(ig)
	if err != nil {
		return nil, fmt.Errorf("failed to build tags for instance group %s, %v", ig.Name, err)
	}
	instanceProfile, err := k.LinkToIAMInstanceProfile(ig)
	if err != nil {
		return nil, fmt.Errorf("failed to build instance profile for instance group %s, %v", ig.Name, err)
	}

	subnets, err := k.GatherSubnets(ig)
	if err != nil {
		return nil, fmt.Errorf("failed to gather subnets for instance group %s, %v", ig.Name, err)
	}

	subnetSelectors := make([]karpenterawsv1.SubnetSelectorTerm, 0)
	addedTags := false
	for _, subnet := range subnets {
		if subnet.ID != "" {
			subnetSelectors = append(subnetSelectors, karpenterawsv1.SubnetSelectorTerm{
				ID: subnet.ID,
			})
		} else if !addedTags {
			// Only add the tag selectors once
			subnetSelectors = append(subnetSelectors, karpenterawsv1.SubnetSelectorTerm{
				Tags: map[string]string{
					fmt.Sprintf("kops.k8s.io/instance-group/%v", ig.Name):   "*",
					fmt.Sprintf("kubernetes.io/cluster/%v", k.Cluster.Name): "*",
				},
			})
			addedTags = true
		}
	}

	sgs, err := k.GetSecurityGroups(ig.Spec.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to get security groups for instance group %s, %v", ig.Name, err)
	}
	securityGroupSelectors := make([]karpenterawsv1.SecurityGroupSelectorTerm, 0)
	for _, sg := range sgs {
		if sg.Task != nil && sg.Task.ID != nil {
			securityGroupSelectors = append(securityGroupSelectors, karpenterawsv1.SecurityGroupSelectorTerm{
				ID: *sg.Task.ID,
			})
		} else {
			securityGroupSelectors = append(securityGroupSelectors, karpenterawsv1.SecurityGroupSelectorTerm{
				Tags: sg.Task.Tags,
			})
		}
	}
	nc := &karpenterawsv1.EC2NodeClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EC2NodeClass",
			APIVersion: "karpenter.k8s.aws/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ig.Name,
		},
		Spec: karpenterawsv1.EC2NodeClassSpec{
			SubnetSelectorTerms: subnetSelectors,
			AMIFamily:           &karpenterawsv1.AMIFamilyCustom,
			AMISelectorTerms: []karpenterawsv1.AMISelectorTerm{
				{
					ID: fi.ValueOf(image.ImageId),
				},
			},
			BlockDeviceMappings:        ec2BlockDeviceMappings(ig, *image),
			SecurityGroupSelectorTerms: securityGroupSelectors,
			UserData:                   nil, // TODO: get userdata from pkg/model/bootstrapscript.go
			InstanceProfile:            instanceProfile.Name,
			Tags:                       tags,
			DetailedMonitoring:         ig.Spec.DetailedInstanceMonitoring,
		},
	}
	if im := ig.Spec.InstanceMetadata; im != nil {
		nc.Spec.MetadataOptions = &karpenterawsv1.MetadataOptions{
			HTTPTokens:              im.HTTPTokens,
			HTTPPutResponseHopLimit: im.HTTPPutResponseHopLimit,
		}
	}
	if kubelet := ig.Spec.Kubelet; kubelet != nil {
		nc.Spec.Kubelet = &karpenterawsv1.KubeletConfiguration{
			SystemReserved: kubelet.SystemReserved,
			KubeReserved:   kubelet.KubeReserved,
			MaxPods:        kubelet.MaxPods,
		}
	}
	return nc, nil
}

func (k KarpenterModelBuilder) NodePool(ig *kops.InstanceGroup, imageArchictectures, instanceTypes []string) (*karpenterv1.NodePool, error) {
	taints := make([]corev1.Taint, 0)
	for _, t := range ig.Spec.Taints {
		taint, err := util.ParseTaint(t)
		if err != nil {
			return nil, fmt.Errorf("failed to parse taint %s, %v", t, err)
		}
		taints = append(taints, corev1.Taint{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: corev1.TaintEffect(taint.Effect),
		})
	}

	startupTaints := make([]corev1.Taint, 0)
	if k.Cluster.Spec.ExternalCloudControllerManager != nil {
		startupTaints = append(startupTaints, corev1.Taint{
			Key:    cloudprovider.TaintExternalCloudProvider,
			Effect: corev1.TaintEffectNoSchedule,
		})
	}

	// The CRD defaults to 720h but the nil value marshals to `Never` so we set 720h explicitly
	thirtyDays := 720 * time.Hour

	np := karpenterv1.NodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NodePool",
			APIVersion: "karpenter.sh/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ig.Name,
		},
		Spec: karpenterv1.NodePoolSpec{
			Disruption: karpenterv1.Disruption{
				ConsolidateAfter: karpenterv1.NillableDuration{
					Duration: &thirtyDays,
				},
			},
			Template: karpenterv1.NodeClaimTemplate{
				ObjectMeta: karpenterv1.ObjectMeta{
					Labels: ig.Spec.NodeLabels,
				},
				Spec: karpenterv1.NodeClaimTemplateSpec{
					Taints:        taints,
					StartupTaints: startupTaints,
					Requirements: []karpenterv1.NodeSelectorRequirementWithMinValues{
						{
							NodeSelectorRequirement: corev1.NodeSelectorRequirement{
								Key:      "karpenter.sh/capacity-type",
								Operator: "In",
								Values:   []string{"spot", "on-demand"},
							},
						},
						{
							NodeSelectorRequirement: corev1.NodeSelectorRequirement{
								Key:      "kubernetes.io/arch",
								Operator: "In",
								Values:   imageArchictectures,
							},
						},
						{
							NodeSelectorRequirement: corev1.NodeSelectorRequirement{
								Key:      "node.kubernetes.io/instance-type",
								Operator: "In",
								Values:   instanceTypes,
							},
						},
					},
					NodeClassRef: &karpenterv1.NodeClassReference{
						Kind:  "EC2NodeClass",
						Name:  ig.Name,
						Group: karpenterawsapis.Group,
					},
					ExpireAfter: karpenterv1.NillableDuration{
						Duration: &thirtyDays,
					},
				},
			},
		},
	}
	return &np, nil
}

func ec2BlockDeviceMappings(ig *kops.InstanceGroup, image ec2types.Image) []*karpenterawsv1.BlockDeviceMapping {
	bdms := make([]*karpenterawsv1.BlockDeviceMapping, 0)

	rootBDM := &karpenterawsv1.BlockDeviceMapping{
		DeviceName: image.RootDeviceName,
		RootVolume: true,
		EBS:        &karpenterawsv1.BlockDevice{},
	}
	if rv := ig.Spec.RootVolume; rv != nil {
		rootBDM.EBS.VolumeType = rv.Type
		rootBDM.EBS.Encrypted = rv.Encryption
		rootBDM.EBS.KMSKeyID = rv.EncryptionKey
		if rv.IOPS != nil {
			rootBDM.EBS.IOPS = fi.PtrTo(int64(fi.ValueOf(rv.IOPS)))
		}
		if rv.Throughput != nil {
			rootBDM.EBS.Throughput = fi.PtrTo(int64(fi.ValueOf(rv.Throughput)))
		}
		if rv.Size != nil {
			rootBDM.EBS.VolumeSize = resources.Quantity(fmt.Sprintf("%dG", fi.ValueOf(rv.Size)))
		}
	}
	bdms = append(bdms, rootBDM)

	for _, vol := range ig.Spec.Volumes {
		bdm := &karpenterawsv1.BlockDeviceMapping{
			DeviceName: fi.PtrTo(vol.Device),
			EBS: &karpenterawsv1.BlockDevice{
				DeleteOnTermination: vol.DeleteOnTermination,
				Encrypted:           vol.Encrypted,
				IOPS:                vol.IOPS,
				KMSKeyID:            vol.Key,
				VolumeType:          fi.PtrTo(vol.Type),
			},
		}
		if vol.Size > 0 {
			bdm.EBS.VolumeSize = resources.Quantity(fmt.Sprintf("%dG", vol.Size))
		}
		bdms = append(bdms, bdm)
	}
	return bdms
}
