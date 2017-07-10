package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

var (
	listEmptyJSON = `
	{
		"tags": [
		],
		"meta": {
			"total": 0
		}
	}
	`

	listJSON = `
	{
		"tags": [
		{
			"name": "testing-1",
			"resources": {
				"droplets": {
					"count": 0,
					"last_tagged": null
				}
			}
		},
		{
			"name": "testing-2",
			"resources": {
				"droplets": {
					"count": 0,
					"last_tagged": null
				}
			}
		}
		],
		"links": {
			"pages":{
				"next":"http://example.com/v2/tags/?page=3",
				"prev":"http://example.com/v2/tags/?page=1",
				"last":"http://example.com/v2/tags/?page=3",
				"first":"http://example.com/v2/tags/?page=1"
			}
		},
		"meta": {
			"total": 2
		}
	}
	`

	createJSON = `
	{
		"tag": {
			"name": "testing-1",
			"resources": {
				"droplets": {
					"count": 0,
					"last_tagged": null
				}
			}
		}
	}
	`

	getJSON = `
	{
		"tag": {
			"name": "testing-1",
			"resources": {
				"droplets": {
					"count": 1,
					"last_tagged": {
						"id": 1,
						"name": "test.example.com",
						"memory": 1024,
						"vcpus": 2,
						"disk": 20,
						"region": {
							"slug": "nyc1",
							"name": "New York",
							"sizes": [
							"1024mb",
							"512mb"
							],
							"available": true,
							"features": [
							"virtio",
							"private_networking",
							"backups",
							"ipv6"
							]
						},
						"image": {
							"id": 119192817,
							"name": "Ubuntu 13.04",
							"distribution": "ubuntu",
							"slug": "ubuntu1304",
							"public": true,
							"regions": [
							"nyc1"
							],
							"created_at": "2014-07-29T14:35:37Z"
						},
						"size_slug": "1024mb",
						"locked": false,
						"status": "active",
						"networks": {
							"v4": [
							{
								"ip_address": "10.0.0.19",
								"netmask": "255.255.0.0",
								"gateway": "10.0.0.1",
								"type": "private"
							},
							{
								"ip_address": "127.0.0.19",
								"netmask": "255.255.255.0",
								"gateway": "127.0.0.20",
								"type": "public"
							}
							],
							"v6": [
							{
								"ip_address": "2001::13",
								"cidr": 124,
								"gateway": "2400:6180:0000:00D0:0000:0000:0009:7000",
								"type": "public"
							}
							]
						},
						"kernel": {
							"id": 485432985,
							"name": "DO-recovery-static-fsck",
							"version": "3.8.0-25-generic"
						},
						"created_at": "2014-07-29T14:35:37Z",
						"features": [
						"ipv6"
						],
						"backup_ids": [
						449676382
						],
						"snapshot_ids": [
						449676383
						],
						"action_ids": [
						],
						"tags": [
						"tag-1",
						"tag-2"
						]
					}
				}
			}
		}
	}
	`
)

func TestTags_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listJSON)
	})

	tags, _, err := client.Tags.List(ctx, nil)
	if err != nil {
		t.Errorf("Tags.List returned error: %v", err)
	}

	expected := []Tag{{Name: "testing-1", Resources: &TaggedResources{Droplets: &TaggedDropletsResources{Count: 0, LastTagged: nil}}},
		{Name: "testing-2", Resources: &TaggedResources{Droplets: &TaggedDropletsResources{Count: 0, LastTagged: nil}}}}
	if !reflect.DeepEqual(tags, expected) {
		t.Errorf("Tags.List returned %+v, expected %+v", tags, expected)
	}
}

func TestTags_ListEmpty(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listEmptyJSON)
	})

	tags, _, err := client.Tags.List(ctx, nil)
	if err != nil {
		t.Errorf("Tags.List returned error: %v", err)
	}

	expected := []Tag{}
	if !reflect.DeepEqual(tags, expected) {
		t.Errorf("Tags.List returned %+v, expected %+v", tags, expected)
	}
}

func TestTags_ListPaging(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, listJSON)
	})

	_, resp, err := client.Tags.List(ctx, nil)
	if err != nil {
		t.Errorf("Tags.List returned error: %v", err)
	}
	checkCurrentPage(t, resp, 2)
}

func TestTags_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/tags/testing-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, getJSON)
	})

	tag, _, err := client.Tags.Get(ctx, "testing-1")
	if err != nil {
		t.Errorf("Tags.Get returned error: %v", err)
	}

	if tag.Name != "testing-1" {
		t.Errorf("Tags.Get return an incorrect name, got %+v, expected %+v", tag.Name, "testing-1")
	}

	if tag.Resources.Droplets.Count != 1 {
		t.Errorf("Tags.Get return an incorrect droplet resource count, got %+v, expected %+v", tag.Resources.Droplets.Count, 1)
	}

	if tag.Resources.Droplets.LastTagged.ID != 1 {
		t.Errorf("Tags.Get return an incorrect last tagged droplet %+v, expected %+v", tag.Resources.Droplets.LastTagged.ID, 1)
	}
}

func TestTags_Create(t *testing.T) {
	setup()
	defer teardown()

	createRequest := &TagCreateRequest{
		Name: "testing-1",
	}

	mux.HandleFunc("/v2/tags", func(w http.ResponseWriter, r *http.Request) {
		v := new(TagCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, createRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, createRequest)
		}

		fmt.Fprintf(w, createJSON)
	})

	tag, _, err := client.Tags.Create(ctx, createRequest)
	if err != nil {
		t.Errorf("Tags.Create returned error: %v", err)
	}

	expected := &Tag{Name: "testing-1", Resources: &TaggedResources{Droplets: &TaggedDropletsResources{Count: 0, LastTagged: nil}}}
	if !reflect.DeepEqual(tag, expected) {
		t.Errorf("Tags.Create returned %+v, expected %+v", tag, expected)
	}
}

func TestTags_Delete(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/tags/testing-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Tags.Delete(ctx, "testing-1")
	if err != nil {
		t.Errorf("Tags.Delete returned error: %v", err)
	}
}

func TestTags_TagResource(t *testing.T) {
	setup()
	defer teardown()

	tagResourcesRequest := &TagResourcesRequest{
		Resources: []Resource{{ID: "1", Type: DropletResourceType}},
	}

	mux.HandleFunc("/v2/tags/testing-1/resources", func(w http.ResponseWriter, r *http.Request) {
		v := new(TagResourcesRequest)

		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, tagResourcesRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, tagResourcesRequest)
		}

	})

	_, err := client.Tags.TagResources(ctx, "testing-1", tagResourcesRequest)
	if err != nil {
		t.Errorf("Tags.TagResources returned error: %v", err)
	}
}

func TestTags_UntagResource(t *testing.T) {
	setup()
	defer teardown()

	untagResourcesRequest := &UntagResourcesRequest{
		Resources: []Resource{{ID: "1", Type: DropletResourceType}},
	}

	mux.HandleFunc("/v2/tags/testing-1/resources", func(w http.ResponseWriter, r *http.Request) {
		v := new(UntagResourcesRequest)

		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodDelete)
		if !reflect.DeepEqual(v, untagResourcesRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, untagResourcesRequest)
		}

	})

	_, err := client.Tags.UntagResources(ctx, "testing-1", untagResourcesRequest)
	if err != nil {
		t.Errorf("Tags.UntagResources returned error: %v", err)
	}
}
