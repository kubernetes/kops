package godo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/digitalocean/godo/context"
)

func TestSnapshots_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"snapshots":[{"id":"1"},{"id":"2", "size_gigabytes": 4.84}]}`)
	})
	ctx := context.Background()
	snapshots, _, err := client.Snapshots.List(ctx, nil)
	if err != nil {
		t.Errorf("Snapshots.List returned error: %v", err)
	}

	expected := []Snapshot{{ID: "1"}, {ID: "2", SizeGigaBytes: 4.84}}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("Snapshots.List returned %+v, expected %+v", snapshots, expected)
	}
}

func TestSnapshots_ListVolume(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "volume"
		actual := r.URL.Query().Get("resource_type")
		if actual != expected {
			t.Errorf("'type' query = %v, expected %v", actual, expected)
		}
		fmt.Fprint(w, `{"snapshots":[{"id":"1"},{"id":"2"}]}`)
	})

	ctx := context.Background()
	snapshots, _, err := client.Snapshots.ListVolume(ctx, nil)
	if err != nil {
		t.Errorf("Snapshots.ListVolume returned error: %v", err)
	}

	expected := []Snapshot{{ID: "1"}, {ID: "2"}}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("Snapshots.ListVolume returned %+v, expected %+v", snapshots, expected)
	}
}

func TestSnapshots_ListDroplet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		expected := "droplet"
		actual := r.URL.Query().Get("resource_type")
		if actual != expected {
			t.Errorf("'resource_type' query = %v, expected %v", actual, expected)
		}

		fmt.Fprint(w, `{"snapshots":[{"id":"1"},{"id":"2", "size_gigabytes": 4.84}]}`)
	})

	ctx := context.Background()
	snapshots, _, err := client.Snapshots.ListDroplet(ctx, nil)
	if err != nil {
		t.Errorf("Snapshots.ListDroplet returned error: %v", err)
	}

	expected := []Snapshot{{ID: "1"}, {ID: "2", SizeGigaBytes: 4.84}}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("Snapshots.ListDroplet returned %+v, expected %+v", snapshots, expected)
	}
}

func TestSnapshots_ListSnapshotsMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"snapshots": [{"id":"1"},{"id":"2"}], "links":{"pages":{"next":"http://example.com/v2/snapshots/?page=2"}}}`)
	})

	ctx := context.Background()
	_, resp, err := client.Snapshots.List(ctx, &ListOptions{Page: 2})
	if err != nil {
		t.Fatal(err)
	}
	checkCurrentPage(t, resp, 1)
}

func TestSnapshots_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
    {
        "snapshots": [{"id":"1"},{"id":"2"}],
        "links":{
            "pages":{
                "next":"http://example.com/v2/snapshots/?page=3",
                "prev":"http://example.com/v2/snapshots/?page=1",
                "last":"http://example.com/v2/snapshots/?page=3",
                "first":"http://example.com/v2/snapshots/?page=1"
            }
        }
    }`

	mux.HandleFunc("/v2/snapshots", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jBlob)
	})

	ctx := context.Background()
	opt := &ListOptions{Page: 2}
	_, resp, err := client.Snapshots.List(ctx, opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestSnapshots_GetSnapshotByID(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"snapshot":{"id":"12345"}}`)
	})

	ctx := context.Background()
	snapshots, _, err := client.Snapshots.Get(ctx, "12345")
	if err != nil {
		t.Errorf("Snapshot.GetByID returned error: %v", err)
	}

	expected := &Snapshot{ID: "12345"}
	if !reflect.DeepEqual(snapshots, expected) {
		t.Errorf("Snapshots.GetByID returned %+v, expected %+v", snapshots, expected)
	}
}

func TestSnapshots_Destroy(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/snapshots/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	ctx := context.Background()
	_, err := client.Snapshots.Delete(ctx, "12345")
	if err != nil {
		t.Errorf("Snapshot.Delete returned error: %v", err)
	}
}

func TestSnapshot_String(t *testing.T) {
	snapshot := &Snapshot{
		ID:            "1",
		Name:          "Snapsh176ot",
		ResourceID:    "0",
		ResourceType:  "droplet",
		Regions:       []string{"one"},
		MinDiskSize:   20,
		SizeGigaBytes: 4.84,
		Created:       "2013-11-27T09:24:55Z",
	}

	stringified := snapshot.String()
	expected := `godo.Snapshot{ID:"1", Name:"Snapsh176ot", ResourceID:"0", ResourceType:"droplet", Regions:["one"], MinDiskSize:20, SizeGigaBytes:4.84, Created:"2013-11-27T09:24:55Z"}`
	if expected != stringified {
		t.Errorf("Snapshot.String returned %+v, expected %+v", stringified, expected)
	}
}
