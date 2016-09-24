// Copyright 2014 Google Inc. All Rights Reserved.
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

package datastore_test

import (
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"golang.org/x/net/context"
)

// TODO(jbd): Document other authorization methods and refer to them here.
func Example_auth() {
	ctx := context.Background()
	// Use Google Application Default Credentials to authorize and authenticate the client.
	// More information about Application Default Credentials and how to enable is at
	// https://developers.google.com/identity/protocols/application-default-credentials.
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	// Use the client (see other examples).

	// Close the client when finished.
	client.Close()
}

func ExampleNewClient() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	_ = client // TODO: Use client.
}

func ExampleClient_Get() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	type Article struct {
		Title       string
		Description string
		Body        string `datastore:",noindex"`
		Author      *datastore.Key
		PublishedAt time.Time
	}
	key := datastore.NewKey(ctx, "Article", "articled1", 0, nil)
	article := &Article{}
	if err := client.Get(ctx, key, article); err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_Put() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	type Article struct {
		Title       string
		Description string
		Body        string `datastore:",noindex"`
		Author      *datastore.Key
		PublishedAt time.Time
	}
	newKey := datastore.NewIncompleteKey(ctx, "Article", nil)
	_, err = client.Put(ctx, newKey, &Article{
		Title:       "The title of the article",
		Description: "The description of the article...",
		Body:        "...",
		Author:      datastore.NewKey(ctx, "Author", "jbd", 0, nil),
		PublishedAt: time.Now(),
	})
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_Delete() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	key := datastore.NewKey(ctx, "Article", "articled1", 0, nil)
	if err := client.Delete(ctx, key); err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_DeleteMulti() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	var keys []*datastore.Key
	for i := 1; i <= 10; i++ {
		keys = append(keys, datastore.NewKey(ctx, "Article", "", int64(i), nil))
	}
	if err := client.DeleteMulti(ctx, keys); err != nil {
		// TODO: Handle error.
	}
}

type Post struct {
	Title       string
	PublishedAt time.Time
	Comments    int
}

func ExampleClient_GetMulti() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	keys := []*datastore.Key{
		datastore.NewKey(ctx, "Post", "post1", 0, nil),
		datastore.NewKey(ctx, "Post", "post2", 0, nil),
		datastore.NewKey(ctx, "Post", "post3", 0, nil),
	}
	posts := make([]Post, 3)
	if err := client.GetMulti(ctx, keys, posts); err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_PutMulti_slice() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	keys := []*datastore.Key{
		datastore.NewKey(ctx, "Post", "post1", 0, nil),
		datastore.NewKey(ctx, "Post", "post2", 0, nil),
	}

	// PutMulti with a Post slice.
	posts := []*Post{
		{Title: "Post 1", PublishedAt: time.Now()},
		{Title: "Post 2", PublishedAt: time.Now()},
	}
	if _, err := client.PutMulti(ctx, keys, posts); err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_PutMulti_interfaceSlice() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	keys := []*datastore.Key{
		datastore.NewKey(ctx, "Post", "post1", 0, nil),
		datastore.NewKey(ctx, "Post", "post2", 0, nil),
	}

	// PutMulti with an empty interface slice.
	posts := []interface{}{
		&Post{Title: "Post 1", PublishedAt: time.Now()},
		&Post{Title: "Post 2", PublishedAt: time.Now()},
	}
	if _, err := client.PutMulti(ctx, keys, posts); err != nil {
		// TODO: Handle error.
	}
}

func ExampleNewQuery() {
	// Query for Post entities.
	q := datastore.NewQuery("Post")
	_ = q // TODO: Use the query with Client.Run.
}

func ExampleNewQuery_options() {
	// Query to order the posts by the number of comments they have recieved.
	q := datastore.NewQuery("Post").Order("-Comments")
	// Start listing from an offset and limit the results.
	q = q.Offset(20).Limit(10)
	_ = q // TODO: Use the query.
}

func ExampleClient_Count() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	// Count the number of the post entities.
	q := datastore.NewQuery("Post")
	n, err := client.Count(ctx, q)
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Printf("There are %d posts.", n)
}

func ExampleClient_Run() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	// List the posts published since yesterday.
	yesterday := time.Now().Add(-24 * time.Hour)
	q := datastore.NewQuery("Post").Filter("PublishedAt >", yesterday)
	it := client.Run(ctx, q)
	_ = it // TODO: iterate using Next.
}

func ExampleClient_NewTransaction() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	const retries = 3

	// Increment a counter.
	// See https://cloud.google.com/appengine/articles/sharding_counters for
	// a more scalable solution.
	type Counter struct {
		Count int
	}

	key := datastore.NewKey(ctx, "counter", "CounterA", 0, nil)
	var tx *datastore.Transaction
	for i := 0; i < retries; i++ {
		tx, err = client.NewTransaction(ctx)
		if err != nil {
			break
		}

		var c Counter
		if err = tx.Get(key, &c); err != nil && err != datastore.ErrNoSuchEntity {
			break
		}
		c.Count++
		if _, err = tx.Put(key, &c); err != nil {
			break
		}

		// Attempt to commit the transaction. If there's a conflict, try again.
		if _, err = tx.Commit(); err != datastore.ErrConcurrentTransaction {
			break
		}
	}
	if err != nil {
		// TODO: Handle error.
	}
}

func ExampleClient_RunInTransaction() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}

	// Increment a counter.
	// See https://cloud.google.com/appengine/articles/sharding_counters for
	// a more scalable solution.
	type Counter struct {
		Count int
	}

	var count int
	key := datastore.NewKey(ctx, "Counter", "singleton", 0, nil)
	_, err = client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var x Counter
		if err := tx.Get(key, &x); err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}
		x.Count++
		if _, err := tx.Put(key, &x); err != nil {
			return err
		}
		count = x.Count
		return nil
	}, nil)
	if err != nil {
		// TODO: Handle error.
	}
	// The value of count is only valid once the transaction is successful
	// (RunInTransaction has returned nil).
	fmt.Printf("Count=%d\n", count)
}

func ExampleClient_AllocateIDs() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "project-id")
	if err != nil {
		// TODO: Handle error.
	}
	var keys []*datastore.Key
	for i := 0; i < 10; i++ {
		keys = append(keys, datastore.NewIncompleteKey(ctx, "Article", nil))
	}
	keys, err = client.AllocateIDs(ctx, keys)
	if err != nil {
		// TODO: Handle error.
	}
	_ = keys // TODO: Use keys.
}

func ExampleKey_Encode() {
	ctx := context.Background()
	key := datastore.NewKey(ctx, "Article", "", 1, nil)
	encoded := key.Encode()
	fmt.Println(encoded)
	// Output: EgsKB0FydGljbGUQAQ
}

func ExampleDecodeKey() {
	const encoded = "EgsKB0FydGljbGUQAQ"
	key, err := datastore.DecodeKey(encoded)
	if err != nil {
		// TODO: Handle error.
	}
	fmt.Println(key)
	// Output: /Article,1
}

func ExampleNewKey() {
	ctx := context.Background()
	// Key with numeric ID.
	k1 := datastore.NewKey(ctx, "Article", "", 1, nil)
	// Key with string ID.
	k2 := datastore.NewKey(ctx, "Article", "article8", 0, nil)
	_, _ = k1, k2 // TODO: Use keys.
}

func ExampleNewIncompleteKey() {
	ctx := context.Background()
	k := datastore.NewIncompleteKey(ctx, "Article", nil)
	_ = k // TODO: Use incomplete key.
}

func ExampleWithNamespace() {
	ctx := context.Background()
	// k1 is in the default namespace.
	k1 := datastore.NewKey(ctx, "Article", "", 1, nil)
	// k2 is in the "other" namespace.
	ctx2 := datastore.WithNamespace(ctx, "other")
	k2 := datastore.NewKey(ctx2, "Article", "", 1, nil)
	// k1 and k2 can refer to different entities, despite the same kind and ID.
	fmt.Printf("k1: %q\n", k1.Namespace())
	fmt.Printf("k2: %q\n", k2.Namespace())
	// Output:
	// k1: ""
	// k2: "other"
}
