/*
Copyright 2026 The Kubernetes Authors.

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

package linodemetadata

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLinodeAuthenticatorCreateToken(t *testing.T) {
	const metadataToken = "test-token"

	h := http.NewServeMux()
	h.HandleFunc("/v1/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method %q", r.Method)
		}
		if got := r.Header.Get("Metadata-Token-Expiry-Seconds"); got == "" {
			t.Fatalf("expected Metadata-Token-Expiry-Seconds header")
		}
		fmt.Fprint(w, metadataToken)
	})
	h.HandleFunc("/v1/instance", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Metadata-Token"); got != metadataToken {
			t.Fatalf("unexpected Metadata-Token %q", got)
		}
		fmt.Fprint(w, "id: 123\nlabel: node-1\nregion: us-ord\n")
	})

	ts := httptest.NewServer(h)
	defer ts.Close()

	a := &linodeAuthenticator{client: ts.Client(), metadataBaseURL: ts.URL}
	token, err := a.CreateToken(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := LinodeAuthenticationTokenPrefix + "123"; token != want {
		t.Fatalf("expected token %q, got %q", want, token)
	}
}

func TestGetLinodeMetadataValueMissingKey(t *testing.T) {
	h := http.NewServeMux()
	h.HandleFunc("/v1/token", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "token")
	})
	h.HandleFunc("/v1/instance", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "label: node-1\nregion: us-ord\n")
	})

	ts := httptest.NewServer(h)
	defer ts.Close()

	_, err := getLinodeMetadataValue(context.Background(), ts.Client(), ts.URL, "id")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "instance id") {
		t.Fatalf("unexpected error: %v", err)
	}
}
