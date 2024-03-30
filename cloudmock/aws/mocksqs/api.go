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

package mocksqs

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"k8s.io/kops/util/pkg/awsinterfaces"
)

type MockSQS struct {
	awsinterfaces.SQSAPI
	mutex sync.Mutex

	Queues map[string]mockQueue
}

type mockQueue struct {
	url        *string
	attributes map[string]*string
	tags       map[string]*string
}

var _ awsinterfaces.SQSAPI = &MockSQS{}

func (m *MockSQS) CreateQueue(ctx context.Context, input *sqs.CreateQueueInput, optFns ...func(*sqs.Options)) (*sqs.CreateQueueOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := *input.QueueName
	url := "https://sqs.us-east-1.amazonaws.com/123456789123/" + name

	if m.Queues == nil {
		m.Queues = make(map[string]mockQueue)
	}
	queue := mockQueue{
		url:        &url,
		attributes: aws.StringMap(input.Attributes),
		tags:       aws.StringMap(input.Tags),
	}

	arn := fmt.Sprintf("arn:aws-test:sqs:us-test-1:000000000000:queue/%v", aws.ToString(input.QueueName))
	queue.attributes["QueueArn"] = &arn

	m.Queues[name] = queue

	response := &sqs.CreateQueueOutput{
		QueueUrl: &url,
	}
	return response, nil
}

func (m *MockSQS) ListQueues(ctx context.Context, input *sqs.ListQueuesInput, optFns ...func(*sqs.Options)) (*sqs.ListQueuesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &sqs.ListQueuesOutput{}

	if queue, ok := m.Queues[*input.QueueNamePrefix]; ok {
		response.QueueUrls = []string{aws.ToString(queue.url)}
	}
	return response, nil
}

func (m *MockSQS) GetQueueAttributes(ctx context.Context, input *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &sqs.GetQueueAttributesOutput{}

	for _, v := range m.Queues {
		if *v.url == *input.QueueUrl {
			response.Attributes = aws.ToStringMap(v.attributes)
			return response, nil
		}
	}
	return response, nil
}

func (m *MockSQS) ListQueueTags(ctx context.Context, input *sqs.ListQueueTagsInput, optFns ...func(*sqs.Options)) (*sqs.ListQueueTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &sqs.ListQueueTagsOutput{}

	for _, v := range m.Queues {
		if *v.url == *input.QueueUrl {
			response.Tags = aws.ToStringMap(v.tags)
			return response, nil
		}
	}
	return response, nil
}

func (m *MockSQS) DeleteQueue(ctx context.Context, input *sqs.DeleteQueueInput, optFns ...func(*sqs.Options)) (*sqs.DeleteQueueOutput, error) {
	panic("Not implemented")
}
