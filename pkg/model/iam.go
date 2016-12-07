package model

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model/iam"
	"fmt"
)

// IAMModelBuilder configures IAM objects
type IAMModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &IAMModelBuilder{}

const RolePolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "{{ IAMServiceEC2 }}"},
      "Action": "sts:AssumeRole"
    }
  ]
}`

func (b *IAMModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, role := range []kops.InstanceGroupRole{kops.InstanceGroupRoleNode, kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleBastion} {
		name := b.IAMName(role)

		var iamRole *awstasks.IAMRole
		{

			iamRole = &awstasks.IAMRole{
				Name: s(name),
				RolePolicyDocument: fi.WrapResource(fi.NewStringResource(RolePolicy)),
			}
			c.AddTask(iamRole)

		}

		policy, err := b.buildAWSIAMPolicy(role)
		if err != nil {
			return err
		}
		{
			t := &awstasks.IAMRolePolicy{
				Name: s(name),
				Role: iamRole,
				PolicyDocument: fi.WrapResource(fi.NewStringResource(policy)),
			}
			c.AddTask(t)
		}

		var iamInstanceProfile *awstasks.IAMInstanceProfile
		{
			iamInstanceProfile = &awstasks.IAMInstanceProfile{
				Name: s(name),
			}
			c.AddTask(iamInstanceProfile)
		}

		{
			iamInstanceProfileRole := &awstasks.IAMInstanceProfileRole{
				InstanceProfile: iamInstanceProfile,
				Role: iamRole,
			}
			c.AddTask(iamInstanceProfileRole)
		}
	}

	return nil
}


// buildAWSIAMPolicy produces the AWS IAM policy for the given role
func (b *IAMModelBuilder) buildAWSIAMPolicy(role kops.InstanceGroupRole) (string, error) {
	pb := &iam.IAMPolicyBuilder{
		Cluster: b.Cluster,
		Role:    role,
		Region:  b.Region,
	}

	policy, err := pb.BuildAWSIAMPolicy()
	if err != nil {
		return "", fmt.Errorf("error building IAM policy: %v", err)
	}
	json, err := policy.AsJSON()
	if err != nil {
		return "", fmt.Errorf("error building IAM policy: %v", err)
	}
	return json, nil
}

