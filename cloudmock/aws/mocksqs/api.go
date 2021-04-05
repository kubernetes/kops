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
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

type MockSQS struct {
	sqsiface.SQSAPI
	mutex sync.Mutex

	Queues map[string]mockQueue
}

type mockQueue struct {
	url        *string
	attributes map[string]*string
	tags       map[string]*string
}

var _ sqsiface.SQSAPI = &MockSQS{}

func (m *MockSQS) CreateQueue(input *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := *input.QueueName
	url := "https://sqs.us-east-1.amazonaws.com/123456789123/" + name

	if m.Queues == nil {
		m.Queues = make(map[string]mockQueue)
	}
	queue := mockQueue{
		url:        &url,
		attributes: input.Attributes,
		tags:       input.Tags,
	}

	m.Queues[name] = queue

	response := &sqs.CreateQueueOutput{
		QueueUrl: &url,
	}
	return response, nil
}

func (m *MockSQS) ListQueues(input *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &sqs.ListQueuesOutput{}

	if queue, ok := m.Queues[*input.QueueNamePrefix]; ok {
		response.QueueUrls = []*string{queue.url}
	}
	return response, nil
}

func (m *MockSQS) GetQueueAttributes(input *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &sqs.GetQueueAttributesOutput{}

	for _, v := range m.Queues {
		if *v.url == *input.QueueUrl {
			response.Attributes = v.attributes
			return response, nil
		}
	}
	return response, nil
}

func (m *MockSQS) ListQueueTags(input *sqs.ListQueueTagsInput) (*sqs.ListQueueTagsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	response := &sqs.ListQueueTagsOutput{}

	for _, v := range m.Queues {
		if *v.url == *input.QueueUrl {
			response.Tags = v.tags
			return response, nil
		}
	}
	return response, nil
}

func (m *MockSQS) DeleteQueue(*sqs.DeleteQueueInput) (*sqs.DeleteQueueOutput, error) {
	panic("Not implemented")
}
