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

package mockec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog"
)

func (m *MockEC2) DescribeImageAttributeRequest(*ec2.DescribeImageAttributeInput) (*request.Request, *ec2.DescribeImageAttributeOutput) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeImageAttributeWithContext(aws.Context, *ec2.DescribeImageAttributeInput, ...request.Option) (*ec2.DescribeImageAttributeOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeImageAttribute(*ec2.DescribeImageAttributeInput) (*ec2.DescribeImageAttributeOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeImagesRequest(*ec2.DescribeImagesInput) (*request.Request, *ec2.DescribeImagesOutput) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeImagesWithContext(aws.Context, *ec2.DescribeImagesInput, ...request.Option) (*ec2.DescribeImagesOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeImages(request *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	klog.Infof("DescribeImages: %v", request)

	var images []*ec2.Image

	for _, image := range m.Images {
		matches, err := m.imageMatchesFilter(image, request.Filters)
		if err != nil {
			return nil, err
		}
		if !matches {
			continue
		}

		copy := *image
		copy.Tags = m.getTags(ec2.ResourceTypeImage, *image.ImageId)
		images = append(images, &copy)
	}

	response := &ec2.DescribeImagesOutput{
		Images: images,
	}

	return response, nil
}
func (m *MockEC2) DescribeImportImageTasksRequest(*ec2.DescribeImportImageTasksInput) (*request.Request, *ec2.DescribeImportImageTasksOutput) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeImportImageTasksWithContext(aws.Context, *ec2.DescribeImportImageTasksInput, ...request.Option) (*ec2.DescribeImportImageTasksOutput, error) {
	panic("Not implemented")
}
func (m *MockEC2) DescribeImportImageTasks(*ec2.DescribeImportImageTasksInput) (*ec2.DescribeImportImageTasksOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) imageMatchesFilter(image *ec2.Image, filters []*ec2.Filter) (bool, error) {
	allFiltersMatch := true
	for _, filter := range filters {
		match := false
		switch *filter.Name {

		case "name":
			for _, v := range filter.Values {
				if aws.StringValue(image.Name) == *v {
					match = true
				}
			}

		default:
			if strings.HasPrefix(*filter.Name, "tag:") {
				match = m.hasTag(ec2.ResourceTypeImage, *image.ImageId, filter)
			} else {
				return false, fmt.Errorf("unknown filter name: %q", *filter.Name)
			}
		}

		if !match {
			allFiltersMatch = false
			break
		}
	}

	return allFiltersMatch, nil
}
