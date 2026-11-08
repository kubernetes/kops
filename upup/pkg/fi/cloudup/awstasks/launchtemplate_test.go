/*
Copyright 2026 The Kubernetes Authors.

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

package awstasks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/cloudmock/aws/mockssm"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

func setupMockCloudWithSSMAndImage(paramName, amiID string) *awsup.MockAWSCloud {
	cloud := awsup.BuildMockAWSCloud("us-east-1", "a")
	mockEC2 := &mockec2.MockEC2{}
	mockSSM := &mockssm.MockSSM{
		Parameters: map[string]string{
			paramName: amiID,
		},
	}
	cloud.MockEC2 = mockEC2
	cloud.MockSSM = mockSSM

	mockEC2.Images = append(mockEC2.Images, &ec2types.Image{
		ImageId:        aws.String(amiID),
		Name:           aws.String("test-ami"),
		RootDeviceName: aws.String("/dev/xvda"),
		Architecture:   ec2types.ArchitectureValuesX8664,
	})

	return cloud
}

func TestLaunchTemplateRenderAWS_ImageID(t *testing.T) {
	paramPath := "/aws/service/ami-amazon-linux-latest/al2023"
	resolvedAMI := "ami-0123456789abcdef0"

	tests := []struct {
		name        string
		imageID     string
		wantImageID string
	}{
		{
			name:        "dynamic SSM resolution retains resolve:ssm: prefix",
			imageID:     "resolve:ssm:" + paramPath,
			wantImageID: "resolve:ssm:" + paramPath,
		},
		{
			name:        "static SSM resolution resolves to concrete AMI ID",
			imageID:     "ssm:" + paramPath,
			wantImageID: resolvedAMI,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cloud := setupMockCloudWithSSMAndImage(paramPath, resolvedAMI)
			target := awsup.NewAWSAPITarget(cloud)

			lt := &LaunchTemplate{
				Name:         aws.String("nodes-lt"),
				ImageID:      aws.String(tc.imageID),
				InstanceType: new(ec2types.InstanceTypeT3Medium),
				Lifecycle:    fi.LifecycleSync,
			}

			if err := lt.RenderAWS(target, nil, lt, nil); err != nil {
				t.Fatalf("RenderAWS() unexpected error: %v", err)
			}

			out, err := cloud.MockEC2.DescribeLaunchTemplateVersions(context.Background(), &ec2.DescribeLaunchTemplateVersionsInput{
				LaunchTemplateName: aws.String("nodes-lt"),
			})
			if err != nil {
				t.Fatalf("DescribeLaunchTemplateVersions() unexpected error: %v", err)
			}
			if len(out.LaunchTemplateVersions) == 0 || out.LaunchTemplateVersions[0].LaunchTemplateData == nil {
				t.Fatalf("LaunchTemplate 'nodes-lt' was not found in MockEC2")
			}
			foundData := out.LaunchTemplateVersions[0].LaunchTemplateData
			if aws.ToString(foundData.ImageId) != tc.wantImageID {
				t.Errorf("LaunchTemplate ImageId = %q, want %q", aws.ToString(foundData.ImageId), tc.wantImageID)
			}
		})
	}
}

func TestLaunchTemplateRenderTerraform_ImageID(t *testing.T) {
	paramPath := "/aws/service/ami-amazon-linux-latest/al2023"
	resolvedAMI := "ami-0123456789abcdef0"

	tests := []struct {
		name     string
		imageID  string
		wantInTF string
	}{
		{
			name:     "dynamic SSM resolution writes resolve:ssm: to terraform",
			imageID:  "resolve:ssm:" + paramPath,
			wantInTF: `"resolve:ssm:` + paramPath + `"`,
		},
		{
			name:     "static SSM resolution writes resolved AMI ID to terraform",
			imageID:  "ssm:" + paramPath,
			wantInTF: `"` + resolvedAMI + `"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cloud := setupMockCloudWithSSMAndImage(paramPath, resolvedAMI)
			outdir := t.TempDir()
			target := terraform.NewTerraformTarget(cloud, "test", outdir, nil)

			lt := &LaunchTemplate{
				Name:         aws.String("nodes-lt"),
				ImageID:      aws.String(tc.imageID),
				InstanceType: new(ec2types.InstanceTypeT3Medium),
				Lifecycle:    fi.LifecycleSync,
			}

			if err := lt.RenderTerraform(target, nil, lt, nil); err != nil {
				t.Fatalf("RenderTerraform() unexpected error: %v", err)
			}

			if err := target.Finish(make(map[string]fi.CloudupTask)); err != nil {
				t.Fatalf("target.Finish() unexpected error: %v", err)
			}

			content, err := os.ReadFile(filepath.Join(outdir, "kubernetes.tf"))
			if err != nil {
				t.Fatalf("os.ReadFile() unexpected error: %v", err)
			}
			tfContent := string(content)
			if !strings.Contains(tfContent, tc.wantInTF) {
				t.Errorf("Terraform output does not contain %q. Output:\n%s", tc.wantInTF, tfContent)
			}
		})
	}
}

func TestLaunchTemplateFind_ImageIDDiff(t *testing.T) {
	paramPath := "/aws/service/ami-amazon-linux-latest/al2023"
	resolvedAMI := "ami-0123456789abcdef0"

	tests := []struct {
		name            string
		taskImageID     string
		actualImageID   string
		wantActualMatch bool
	}{
		{
			name:            "dynamic SSM when actual is concrete AMI must NOT match so kops updates LT",
			taskImageID:     "resolve:ssm:" + paramPath,
			actualImageID:   resolvedAMI,
			wantActualMatch: false,
		},
		{
			name:            "dynamic SSM when actual is identical resolve:ssm matches",
			taskImageID:     "resolve:ssm:" + paramPath,
			actualImageID:   "resolve:ssm:" + paramPath,
			wantActualMatch: true,
		},
		{
			name:            "static SSM when actual is concrete resolved AMI matches to prevent spurious diff",
			taskImageID:     "ssm:" + paramPath,
			actualImageID:   resolvedAMI,
			wantActualMatch: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cloud := setupMockCloudWithSSMAndImage(paramPath, resolvedAMI)
			mockEC2 := cloud.MockEC2.(*mockec2.MockEC2)

			// Pre-create launch template in mock EC2 with tc.actualImageID
			ltName := "nodes-lt"
			ltID := "lt-12345678"
			mockEC2.CreateLaunchTemplate(context.Background(), &ec2.CreateLaunchTemplateInput{
				LaunchTemplateName: aws.String(ltName),
				LaunchTemplateData: &ec2types.RequestLaunchTemplateData{
					ImageId:      aws.String(tc.actualImageID),
					InstanceType: ec2types.InstanceTypeT3Medium,
				},
				TagSpecifications: []ec2types.TagSpecification{
					{
						ResourceType: ec2types.ResourceTypeLaunchTemplate,
						Tags: []ec2types.Tag{
							{Key: aws.String("Name"), Value: aws.String(ltName)},
						},
					},
				},
			})

			target := awsup.NewAWSAPITarget(cloud)
			ctx := &fi.CloudupContext{
				Target: target,
				T: fi.CloudupSubContext{
					Cloud: cloud,
				},
			}

			lt := &LaunchTemplate{
				ID:           aws.String(ltID),
				Name:         aws.String(ltName),
				ImageID:      aws.String(tc.taskImageID),
				InstanceType: new(ec2types.InstanceTypeT3Medium),
				Lifecycle:    fi.LifecycleSync,
			}

			actual, err := lt.Find(ctx)
			if err != nil {
				t.Fatalf("Find() unexpected error: %v", err)
			}
			if actual == nil {
				t.Fatalf("Find() expected LaunchTemplate, got nil")
			}

			matches := aws.ToString(actual.ImageID) == tc.taskImageID
			if matches != tc.wantActualMatch {
				t.Errorf("Find() actual.ImageID = %q, task.ImageID = %q; match = %v, want %v",
					aws.ToString(actual.ImageID), tc.taskImageID, matches, tc.wantActualMatch)
			}
		})
	}
}
