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

package bigquery_test

import (
	"fmt"

	"cloud.google.com/go/bigquery"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

func ExampleNewClient() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	_ = client // TODO: Use client.
}

func ExampleClient_Dataset() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	ds := client.Dataset("my-dataset")
	fmt.Println(ds)
}

func ExampleClient_DatasetInProject() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	ds := client.DatasetInProject("their-project-id", "their-dataset")
	fmt.Println(ds)
}

func ExampleClient_Datasets() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Datasets(ctx)
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func ExampleClient_DatasetsInProject() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.DatasetsInProject(ctx, "their-project-id")
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func getJobID() string { return "" }

func ExampleClient_JobFromID() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	jobID := getJobID() // Get a job ID using Job.ID, the console or elsewhere.
	job, err := client.JobFromID(ctx, jobID)
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(job)
}

func ExampleDataset_Create() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	if err := client.Dataset("new-dataset").Create(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleDataset_Table() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	// Table creates a reference to the table. It does not create the actual
	// table in BigQuery; to do so, use Table.Create.
	t := client.Dataset("my-dataset").Table("my-table")
	fmt.Println(t)
}

func ExampleDataset_Tables() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Dataset("my-dataset").Tables(ctx)
	_ = it // TODO: iterate using Next or iterator.Pager.
}

func ExampleDatasetIterator_Next() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Datasets(ctx)
	for {
		ds, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		fmt.Println(ds)
	}
}

func ExampleInferSchema() {
	type Item struct {
		Name  string
		Size  float64
		Count int
	}
	schema, err := bigquery.InferSchema(Item{})
	if err != nil {
		fmt.Println(err)
		// TODO: Handle error.
	}
	for _, fs := range schema {
		fmt.Println(fs.Name, fs.Type)
	}
	// Output:
	// Name STRING
	// Size FLOAT
	// Count INTEGER
}

func ExampleTable_Create() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	t := client.Dataset("my-dataset").Table("new-table")
	if err := t.Create(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_Delete() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	if err := client.Dataset("my-dataset").Table("my-table").Delete(ctx); err != nil {
		// TODO: Handle error.
	}
}

func ExampleTable_Metadata() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	md, err := client.Dataset("my-dataset").Table("my-table").Metadata(ctx)
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(md)
}

func ExampleTable_NewUploader() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	u := client.Dataset("my-dataset").Table("my-table").NewUploader()
	_ = u // TODO: Use u.
}

func ExampleTableIterator_Next() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	it := client.Dataset("my-dataset").Tables(ctx)
	for {
		t, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		fmt.Println(t)
	}
}

type Item struct {
	Name  string
	Size  float64
	Count int
}

// Save implements the ValueSaver interface.
func (i *Item) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"Name":  i.Name,
		"Size":  i.Size,
		"Count": i.Count,
	}, "", nil
}

func ExampleUploader_Put() {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	u := client.Dataset("my-dataset").Table("my-table").NewUploader()
	// Item implements the ValueSaver interface.
	items := []*Item{
		{Name: "n1", Size: 32.6, Count: 7},
		{Name: "n2", Size: 4, Count: 2},
		{Name: "n3", Size: 101.5, Count: 1},
	}
	if err := u.Put(ctx, items); err != nil {
		// TODO: Handle error.
	}
}
