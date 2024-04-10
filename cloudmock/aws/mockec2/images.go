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
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
)

func (m *MockEC2) DescribeImages(ctx context.Context, request *ec2.DescribeImagesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeImagesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeImagesPages: %v", request)

	var images []ec2types.Image

	for _, image := range m.Images {
		matches, err := m.imageMatchesFilter(image, request.Filters)
		if err != nil {
			return nil, err
		}
		if !matches {
			continue
		}

		copy := *image
		copy.Tags = m.getTags(ec2types.ResourceTypeImage, *image.ImageId)
		images = append(images, copy)
	}

	response := &ec2.DescribeImagesOutput{
		Images: images,
	}
	return response, nil
}

func (m *MockEC2) imageMatchesFilter(image *ec2types.Image, filters []ec2types.Filter) (bool, error) {
	allFiltersMatch := true
	for _, filter := range filters {
		match := false
		switch *filter.Name {

		case "name":
			for _, v := range filter.Values {
				if aws.ToString(image.Name) == v {
					match = true
				}
			}

		default:
			if strings.HasPrefix(*filter.Name, "tag:") {
				match = m.hasTag(ec2types.ResourceTypeImage, *image.ImageId, filter)
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
