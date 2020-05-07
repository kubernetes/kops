/*
Copyright 2020 The Kubernetes Authors.

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

package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestBuildProxyRequest(t *testing.T) {
	grid := []struct {
		In  string
		Out string
	}{
		{In: "http://127.0.0.1:8080/readyz", Out: "https://127.0.0.1/readyz"},
		{In: "http://127.0.0.1:8080/livez", Out: "https://127.0.0.1/livez"},
		{In: "http://127.0.0.1:8080/healthz", Out: "https://127.0.0.1/healthz"},
		{In: "http://127.0.0.1:8080/ready", Out: ""},
		{In: "http://127.0.0.1:8080/", Out: ""},
		{In: "http://127.0.0.1:8080/readyz/foo", Out: ""},
		{In: "http://127.0.0.1:8080/readyzfoo", Out: ""},
		{In: "http://127.0.0.1:8080/readyz?", Out: "https://127.0.0.1/readyz"},
		{In: "http://127.0.0.1:8080/readyz?foo=bar", Out: "https://127.0.0.1/readyz"},
		{In: "http://127.0.0.1:8080/readyz?exclude=1", Out: "https://127.0.0.1/readyz?exclude=1"},
		{In: "http://127.0.0.1:8080/readyz?exclude=1&exclude=2", Out: "https://127.0.0.1/readyz?exclude=1&exclude=2"},
		{In: "http://127.0.0.1:8080/readyz?exclude=1&verbose", Out: "https://127.0.0.1/readyz?exclude=1"},
		{In: "http://127.0.0.1:8080/readyz?exclude", Out: "https://127.0.0.1/readyz?exclude="},
	}

	for _, g := range grid {
		g := g
		t.Run(g.In, func(t *testing.T) {
			u, err := url.Parse(g.In)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", g.In, err)
			}
			req := &http.Request{
				Method: "GET",
				URL:    u,
			}
			out := mapToProxyRequest(req)
			actual := ""
			if out != nil {
				if out.Method != "GET" {
					t.Fatalf("unexpected method %q", out.Method)
				}
				if out.URL == nil {
					t.Fatalf("expected URL to be set")
				}
				actual = out.URL.String()
				if actual == "" {
					t.Fatalf("unexpected empty URL")
				}
			}

			if actual != g.Out {
				t.Fatalf("unexpected mapToProxyRequest result %q => %q, expected %q",
					g.In,
					actual,
					g.Out)
			}
		})
	}
}
