/*
Copyright 2016 The Kubernetes Authors.

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

package main

import (
	"io"
	"os"

	"fmt"

	"bytes"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kops/cmd/kops/util"
	kops_iam "k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"k8s.io/kubernetes/pkg/util/i18n"
)

var (
	auth_long = templates.LongDesc(i18n.T(`
		Create various IAM policies for kops.
		Types include:
		- admin: an IAM policy with the permissions to run the kops program.
		- master: an IAM policy with the permissions needed for a kubernetes master.
		- node: an IAM policy with the permissions needed for a kubernetes node.

		Use flag "--create-network-perms=false"  exist to remove permissions for the admin
		user to not have create, update and delete capabilities with VPC networking resources.

		Use flag "--create-ecr-perms=false" exist to remove ECR permissions for master and nodes.
		`))

	auth_example = templates.Examples(i18n.T(`
	        # The following example will create a policy called kops-admin-policy and attach
	        # it to the user kops-admin.  If the policy aready exists it will be replaced with the new policy.
	        # This policy will have permisions to modify the Route53 zone Z151KI3YBRQBLD.

		kops create auth-policy -type admin \
		  --hosted-zone Z151KI3YBRQBLD --replace-policy="true" --username kops-admin
		  --account-id 962917490108 --name kops-admin-policy

	        # The following example will create a policy called kops-admin-policy and attach
	        # it to the user kops-install.  If the policy aready exists it will be replaced with the new policy.
	        # This policy will have permisions to modify the Route53 zone Z151KI3YBRQBLD.
	        # This policy will NOT have permission to create VPC networking components.  This example
	        # can be used when using a pre-existing VPC, Subnets, and if needed Nat Gateways.
	        # This user will not be able to create any VPC related resources.

	        kops create auth-policy --name kops-installer-test --hosted-zone Z151KI3YMRFBLD --account-id 942917491108 \
	        --replace-policy=true --username kops-installer --create-network-perms=false

	        # The following example will create a policy called kops-master-policy.
	        # If the policy aready exists it will be replaced with the new policy.
	        # This policy will have permisions to modify the Route53 zone Z151KI3YBRQBLD.

		kops create auth-policy -type admin \
		  --hosted-zone Z151KI3YBRQBLD --replace-policy="true" \
		  --account-id 962917490108 --name kops-master-policy


		`))
)

type CreateAuthPolicyOptions struct {
	PolicyType     string // three different types: admin, master, node
	Name           string // the name of the policy we create
	HostedZone     string // hosted dns zone.  This can be parsed from the cluster
	ReplacePolicy  bool   // If the policy is replaced if it already exists
	UserName       string // auth user name that we can attach a policy to
	AccountId      string // cloud account id
	Output         string // allow for json output
	CreateECRPerms bool   // allow for ecr permission to not be created
	// TODO: should we change this to an IAM condition?
	ResourceARN               string   // allow for a specific resource ARN to be added
	CreateNetworkingPerms     bool     // Specific to admin user.  Allow for the networking perms to be removed.
	CreateIAMPerms            bool     // Specific to the admin user. Allow for the Policy Creations perms to be removed.
	CreateCloudFormationPerms bool     // Remove Cloud Formation Perms.
	Region                    string   // Region for creating aws client.
	KMSKeys                   []string // KMS keys to add.
	ClusterName               string   // Cluster name to use.  Once we have a cluster we will not need this member.
}

// NewCmdCreateAuth creates the cobra command to allow a user to create auth.
func NewCmdCreateAuthPolicy(f *util.Factory, out io.Writer) *cobra.Command {

	options := &CreateAuthPolicyOptions{
		PolicyType:                "admin",
		ReplacePolicy:             false,
		CreateECRPerms:            true,
		ResourceARN:               "*",
		CreateIAMPerms:            true,
		CreateNetworkingPerms:     true,
		CreateCloudFormationPerms: false,
		Region: "us-east-1",
	}

	cmd := &cobra.Command{
		Use:     "auth-policy",
		Short:   i18n.T("Create a IAM policy for an kops admin, k8s master or k8s node."),
		Example: auth_example,
		Long:    auth_long,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunCreateAuthPolicy(f, cmd, args, os.Stdout, options)
			if err != nil {
				exitWithError(err)
			}
		},
		// we are hiding this until we introduce life cycles
		Hidden: true,
	}

	cmd.Flags().StringVar(&options.PolicyType, "type", options.PolicyType, "Type of policy to create")
	cmd.Flags().StringVar(&options.Name, "name", options.Name, "Name of policy to create")
	cmd.Flags().StringVar(&options.HostedZone, "hosted-zone", options.HostedZone, "Hosted Zone Id to add to the policy")
	cmd.Flags().BoolVar(&options.ReplacePolicy, "replace-policy", options.ReplacePolicy, "Replace existing policy")
	cmd.Flags().StringVar(&options.AccountId, "account-id", options.AccountId, "Account Id")
	cmd.Flags().StringVar(&options.UserName, "username", options.UserName, "Username to attach policy")
	cmd.Flags().StringVarP(&options.Output, "output", "o", "", "Output the policy as JSON")
	cmd.Flags().BoolVar(&options.CreateECRPerms, "create-ecr-perms", options.CreateECRPerms, "Create ECR policy for master or node")
	cmd.Flags().BoolVar(&options.CreateNetworkingPerms, "create-network-perms", options.CreateNetworkingPerms, "Give VPC Networking permissions to installer user")
	cmd.Flags().BoolVar(&options.CreateIAMPerms, "create-iam-policy-perms", options.CreateIAMPerms, "Give IAM Role Policy permissions to installer user")
	cmd.Flags().BoolVar(&options.CreateCloudFormationPerms, "create-cf-perms", options.CreateCloudFormationPerms, "Give IAM Cloudformation permissions to installer user")
	cmd.Flags().StringVar(&options.ResourceARN, "policy-resource", options.ResourceARN, "The resource to use for the policy.")
	cmd.Flags().StringVar(&options.Region, "region", options.Region, "aws region")
	cmd.Flags().StringArrayVar(&options.KMSKeys, "kms-keys", options.KMSKeys, "KMS keys for encrypting EBS volumes")
	cmd.Flags().StringVar(&options.ClusterName, "cluster-name", options.ClusterName, "Cluster name that is used to tighten S3 permissions")

	return cmd
}

// RunCreateAuth creates various auth profiles.
func RunCreateAuthPolicy(f *util.Factory, cmd *cobra.Command, args []string, out io.Writer, options *CreateAuthPolicyOptions) error {

	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return err
	}

	// TODO check args more

	bucket := strings.TrimPrefix(rootCommand.RegistryPath, "s3://")
	bucket = strings.TrimSuffix(rootCommand.RegistryPath, "/")

	i := &kops_iam.IAMPolicyBuilder{
		S3Bucket:              bucket,
		HostedZoneID:          options.HostedZone,
		CreatePolicyPerms:     options.CreateNetworkingPerms,
		CloudFormationPerms:   options.CreateCloudFormationPerms,
		CreateECRPerms:        options.CreateECRPerms,
		CreateNetworkingPerms: options.CreateNetworkingPerms,
		ResourceARN:           &options.ResourceARN,
		KMSKeys:               options.KMSKeys,
		ClusterName:           options.ClusterName,
	}

	var policy *kops_iam.IAMPolicy
	switch options.PolicyType {
	case "admin":
		policy, err = i.BuildAWSIAMPolicyInstaller()
	case "node":
		policy, err = i.BuildAWSIAMPolicyNode()
	case "master":
		policy, err = i.BuildAWSIAMPolicyMaster()
	}

	if err != nil {
		return fmt.Errorf("error building policy %v", err)
	}

	j, err := policy.AsJSON()

	if err != nil {
		return fmt.Errorf("error building policy json %v", err)
	}

	request := &iam.CreatePolicyInput{
		PolicyDocument: aws.String(j),
		PolicyName:     aws.String(options.Name),
	}

	if options.Output == "json" {
		var sb bytes.Buffer
		fmt.Fprintf(&sb, "\n")
		fmt.Fprintf(&sb, "Policy:\n")
		fmt.Fprintf(&sb, "%s", j)
		fmt.Fprintf(&sb, "\n")
		_, err := out.Write(sb.Bytes())

		if err != nil {
			return fmt.Errorf("error writing to output: %v", err)
		}

		return nil
	}
	// just in case the use provides account id with dashes

	if options.UserName != "" && options.AccountId == "" {
		return fmt.Errorf("In order to detach or attach policy please seth the account-id flag")
	}

	account := strings.Replace(options.AccountId, "-", "", -1)
	userName := aws.String(options.UserName)
	policyArn := aws.String("arn:aws:iam::" + account + ":policy/" + options.Name)

	cloud, err := awsup.NewAWSCloud(options.Region, nil)

	target := awsup.NewAWSAPITarget(cloud.(awsup.AWSCloud))

	policyExists := false

	if options.ReplacePolicy {
		r, err := target.Cloud.IAM().GetPolicy(&iam.GetPolicyInput{
			PolicyArn: policyArn,
		})

		glog.V(4).Infof("policy %q", r)

		if err == nil && *r.Policy.PolicyId != "" {
			policyExists = true
			glog.V(4).Infof("policy found")
		}
	}

	if policyExists && options.UserName != "" && options.ReplacePolicy {
		_, err = target.Cloud.IAM().DetachUserPolicy(&iam.DetachUserPolicyInput{
			UserName:  userName,
			PolicyArn: policyArn,
		})

		if err != nil {
			glog.Warningf("unable to detach IAMPolicy: %v", err)
		} else {
			glog.V(2).Infof("policy detached")
		}
	}

	if policyExists && options.ReplacePolicy {
		r, err := target.Cloud.IAM().DeletePolicy(&iam.DeletePolicyInput{
			PolicyArn: policyArn,
		})

		if err != nil {
			glog.Warningf("unable to delete IAMPolicy: %v", err)
		} else {
			glog.V(2).Infof("policy deleted %s", r.String())
		}
	}

	response, err := target.Cloud.IAM().CreatePolicy(request)

	if err != nil {
		return fmt.Errorf("error creating IAMRole: %v", err)
	}

	id := response.Policy.PolicyId

	glog.V(2).Infof("role created %s", id)

	if options.UserName != "" {
		_, err = target.Cloud.IAM().AttachUserPolicy(&iam.AttachUserPolicyInput{
			PolicyArn: response.Policy.Arn,
			UserName:  userName,
		})

		if err != nil {
			return fmt.Errorf("error attaching user to policy: %v", err)
		}
	}

	fmt.Fprintf(out, "\nPolicy Created\n")
	return nil
}
