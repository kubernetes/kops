/*
Copyright 2017 The Kubernetes Authors.

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
	"sort"

	"github.com/aws/aws-sdk-go/aws"
)

type cloudformationTag struct {
	Key   *string `json:"Key"`
	Value *string `json:"Value"`
}

type cfTagByKey []cloudformationTag

func (a cfTagByKey) Len() int      { return len(a) }
func (a cfTagByKey) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a cfTagByKey) Less(i, j int) bool {
	return aws.StringValue(a[i].Key) < aws.StringValue(a[j].Key)
}

func buildCloudformationTags(tags map[string]string) []cloudformationTag {
	var cfTags []cloudformationTag
	for k, v := range tags {
		cfTag := cloudformationTag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}
		cfTags = append(cfTags, cfTag)
	}
	sort.Sort(cfTagByKey(cfTags))
	return cfTags
}
