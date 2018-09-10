package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestImageActions_Transfer(t *testing.T) {
	setup()
	defer teardown()

	transferRequest := &ActionRequest{}

	mux.HandleFunc("/v2/images/12345/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, transferRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, transferRequest)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)

	})

	transfer, _, err := client.ImageActions.Transfer(ctx, 12345, transferRequest)
	if err != nil {
		t.Errorf("ImageActions.Transfer returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(transfer, expected) {
		t.Errorf("ImageActions.Transfer returned %+v, expected %+v", transfer, expected)
	}
}

func TestImageActions_Convert(t *testing.T) {
	setup()
	defer teardown()

	convertRequest := &ActionRequest{
		"type": "convert",
	}

	mux.HandleFunc("/v2/images/12345/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatalf("decode json: %v", err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, convertRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, convertRequest)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)

	})

	transfer, _, err := client.ImageActions.Convert(ctx, 12345)
	if err != nil {
		t.Errorf("ImageActions.Transfer returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(transfer, expected) {
		t.Errorf("ImageActions.Transfer returned %+v, expected %+v", transfer, expected)
	}
}

func TestImageActions_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images/123/actions/456", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)
	})

	action, _, err := client.ImageActions.Get(ctx, 123, 456)
	if err != nil {
		t.Errorf("ImageActions.Get returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("ImageActions.Get returned %+v, expected %+v", action, expected)
	}
}
