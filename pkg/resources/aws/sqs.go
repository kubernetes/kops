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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/model"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func DumpSQSQueue(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["name"] = r.Name
	data["type"] = r.Type
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	return nil
}

func DeleteSQSQueue(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(awsup.AWSCloud)

	url := r.ID

	klog.V(2).Infof("Deleting SQS queue %q", url)
	request := &sqs.DeleteQueueInput{
		QueueUrl: &url,
	}
	_, err := c.SQS().DeleteQueue(request)
	if err != nil {
		return fmt.Errorf("error deleting SQS queue %q: %v", url, err)
	}
	return nil
}

func ListSQSQueues(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(awsup.AWSCloud)

	klog.V(2).Infof("Listing SQS queues")
	queueName := model.QueueNamePrefix(clusterName)

	request := &sqs.ListQueuesInput{
		QueueNamePrefix: &queueName,
	}
	response, err := c.SQS().ListQueues(request)
	if err != nil {
		return nil, fmt.Errorf("error listing SQS queues: %v", err)
	}
	if response == nil || len(response.QueueUrls) == 0 {
		return nil, nil
	}

	var resourceTrackers []*resources.Resource

	for _, queue := range response.QueueUrls {
		resourceTracker := &resources.Resource{
			Name:    queueName,
			ID:      *queue,
			Type:    "sqs",
			Deleter: DeleteSQSQueue,
			Dumper:  DumpSQSQueue,
			Obj:     queue,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}
