// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elasticsearch

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/olivere/elastic.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_api "k8s.io/client-go/pkg/api/v1"
	esCommon "k8s.io/heapster/common/elasticsearch"
	"k8s.io/heapster/events/core"
)

type dataSavedToES struct {
	data string
}

type fakeESSink struct {
	core.EventSink
	savedData []dataSavedToES
}

var FakeESSink fakeESSink

func SaveDataIntoES_Stub(date time.Time, sinkData []interface{}) error {
	for _, data := range sinkData {
		jsonItems, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to transform the items to json : %s", err)
		}
		FakeESSink.savedData = append(FakeESSink.savedData, dataSavedToES{string(jsonItems)})
	}
	return nil
}

// Returns a fake ES sink.
func NewFakeSink() fakeESSink {
	savedData := make([]dataSavedToES, 0)
	return fakeESSink{
		&elasticSearchSink{
			saveData:  SaveDataIntoES_Stub,
			flushData: func() error { return nil },
			esSvc: esCommon.ElasticSearchService{
				EsClient:    &elastic.Client{},
				ClusterName: esCommon.ESClusterName,
			},
		},
		savedData,
	}
}

func TestStoreDataEmptyInput(t *testing.T) {
	fakeSink := NewFakeSink()
	dataBatch := core.EventBatch{}
	fakeSink.ExportEvents(&dataBatch)
	assert.Equal(t, 0, len(fakeSink.savedData))
}

func TestStoreMultipleDataInput(t *testing.T) {
	fakeSink := NewFakeSink()
	timestamp := time.Now()
	now := time.Now()
	event1 := kube_api.Event{
		Message:        "event1",
		Count:          100,
		LastTimestamp:  metav1.NewTime(now),
		FirstTimestamp: metav1.NewTime(now),
	}
	event2 := kube_api.Event{
		Message:        "event2",
		Count:          101,
		LastTimestamp:  metav1.NewTime(now),
		FirstTimestamp: metav1.NewTime(now),
	}
	data := core.EventBatch{
		Timestamp: timestamp,
		Events: []*kube_api.Event{
			&event1,
			&event2,
		},
	}
	fakeSink.ExportEvents(&data)
	// expect msg string
	assert.Equal(t, 2, len(FakeESSink.savedData))

	var expectMsgTemplate = [2]string{
		`{"Count":100,"Metadata":{"creationTimestamp":null},"InvolvedObject":{},"Source":{},"FirstOccurrenceTimestamp":%s,"LastOccurrenceTimestamp":%s,"Message":"event1","Reason":"","Type":"","EventTags":{"cluster_name":"default","eventID":"","hostname":""}}`,
		`{"Count":101,"Metadata":{"creationTimestamp":null},"InvolvedObject":{},"Source":{},"FirstOccurrenceTimestamp":%s,"LastOccurrenceTimestamp":%s,"Message":"event2","Reason":"","Type":"","EventTags":{"cluster_name":"default","eventID":"","hostname":""}}`,
	}

	msgsString := fmt.Sprintf("%s", FakeESSink.savedData)
	ts, _ := json.Marshal(metav1.NewTime(now).Time.UTC())

	for _, mgsTemplate := range expectMsgTemplate {
		expectMsg := fmt.Sprintf(mgsTemplate, ts, ts)
		assert.Contains(t, msgsString, expectMsg)
	}

	FakeESSink = fakeESSink{}
}
