// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AUTO-GENERATED CODE. DO NOT EDIT.

package monitoring_test

import (
	"cloud.google.com/go/monitoring/apiv3"
	"golang.org/x/net/context"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func ExampleNewAgentTranslationClient() {
	ctx := context.Background()
	c, err := monitoring.NewAgentTranslationClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	// TODO: Use client.
	_ = c
}

func ExampleAgentTranslationClient_CreateCollectdTimeSeries() {
	ctx := context.Background()
	c, err := monitoring.NewAgentTranslationClient(ctx)
	if err != nil {
		// TODO: Handle error.
	}

	req := &monitoringpb.CreateCollectdTimeSeriesRequest{
	// TODO: Fill request struct fields.
	}
	err = c.CreateCollectdTimeSeries(ctx, req)
	if err != nil {
		// TODO: Handle error.
	}
}
