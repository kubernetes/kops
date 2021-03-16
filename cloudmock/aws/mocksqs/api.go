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

	mutex     sync.Mutex
	QueueUrls []*string
}

var _ sqsiface.SQSAPI = &MockSQS{}

func (c *MockSQS) DeleteQueue(*sqs.DeleteQueueInput) (*sqs.DeleteQueueOutput, error) {
	panic("Not implemented")
}

func (c *MockSQS) ListQueues(*sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	response := &sqs.ListQueuesOutput{
		QueueUrls: c.QueueUrls,
	}

	return response, nil
}

func (c *MockSQS) ListQueueTags(*sqs.ListQueueTagsInput) (*sqs.ListQueueTagsOutput, error) {
	panic("Not implemented")
}

func (c *MockSQS) CreateQueue(*sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	panic("Not implemented")
}
