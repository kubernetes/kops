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

package trace_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/context"
	api "google.golang.org/api/cloudtrace/v1"
	"google.golang.org/cloud"
	"google.golang.org/cloud/trace"
)

const testProjectID = "testproject"

type fakeRoundTripper struct {
	reqc chan *http.Request
}

func newFakeRoundTripper() *fakeRoundTripper {
	return &fakeRoundTripper{reqc: make(chan *http.Request)}
}

func (rt *fakeRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.reqc <- r
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader("{}")),
	}
	return resp, nil
}

func newTestClient(rt http.RoundTripper) *trace.Client {
	t, err := trace.NewClient(context.Background(), testProjectID, cloud.WithBaseHTTP(&http.Client{Transport: rt}))
	if err != nil {
		panic(err)
	}
	return t
}

// makeRequests makes some requests.
// req is an incoming request used to construct the trace.  traceClient is the
// client used to upload the trace.  rt is the trace client's http client's
// transport.  This is used to retrieve the trace uploaded by the client, if
// any.  If expectTrace is true, we expect a trace will be uploaded.  If
// synchronous is true, the call to Finish is expected not to return before the
// client has uploaded any traces.
func makeRequests(t *testing.T, req *http.Request, traceClient *trace.Client, rt *fakeRoundTripper, synchronous bool, expectTrace bool) *http.Request {
	span := traceClient.SpanFromRequest(req)

	{
		req2, err := http.NewRequest("GET", "http://example.com/bar", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp := &http.Response{StatusCode: 200}
		s := span.NewRemoteChild(req2)
		s.Finish(trace.WithResponse(resp))
	}

	done := make(chan struct{})
	go func() {
		if synchronous {
			err := span.FinishWait()
			if err != nil {
				t.Errorf("Unexpected error from span.FinishWait: %v", err)
			}
		} else {
			span.Finish()
		}
		done <- struct{}{}
	}()
	if !expectTrace {
		<-done
		select {
		case <-rt.reqc:
			t.Errorf("Got a trace, expected none.")
		case <-time.After(5 * time.Millisecond):
		}
		return nil
	} else if !synchronous {
		<-done
		return <-rt.reqc
	} else {
		select {
		case <-done:
			t.Errorf("Synchronous Finish didn't wait for trace upload.")
			return <-rt.reqc
		case <-time.After(5 * time.Millisecond):
			r := <-rt.reqc
			<-done
			return r
		}
	}
}

func TestTrace(t *testing.T) {
	testTrace(t, false)
}

func TestTraceWithWait(t *testing.T) {
	testTrace(t, true)
}

func testTrace(t *testing.T, synchronous bool) {
	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header["X-Cloud-Trace-Context"] = []string{`0123456789ABCDEF0123456789ABCDEF/42;o=3`}

	rt := newFakeRoundTripper()
	traceClient := newTestClient(rt)

	uploaded := makeRequests(t, req, traceClient, rt, synchronous, true)

	if uploaded == nil {
		t.Fatalf("No trace uploaded, expected one.")
	}

	expected := api.Traces{
		Traces: []*api.Trace{
			&api.Trace{
				ProjectId: testProjectID,
				Spans: []*api.TraceSpan{
					&api.TraceSpan{
						Kind: "RPC_CLIENT",
						Labels: map[string]string{
							"trace.cloud.google.com/http/host":        "example.com",
							"trace.cloud.google.com/http/method":      "GET",
							"trace.cloud.google.com/http/status_code": "200",
							"trace.cloud.google.com/http/url":         "http://example.com/bar",
						},
						Name: "/bar",
					},
					&api.TraceSpan{
						Kind: "RPC_SERVER",
						Labels: map[string]string{
							"trace.cloud.google.com/http/host":   "example.com",
							"trace.cloud.google.com/http/method": "GET",
							"trace.cloud.google.com/http/url":    "http://example.com/foo",
						},
						Name: "/foo",
					},
				},
				TraceId: "0123456789ABCDEF0123456789ABCDEF",
			},
		},
	}

	body, err := ioutil.ReadAll(uploaded.Body)
	if err != nil {
		t.Fatal(err)
	}
	var patch api.Traces
	err = json.Unmarshal(body, &patch)
	if err != nil {
		t.Fatal(err)
	}

	if len(patch.Traces) != len(expected.Traces) || len(patch.Traces[0].Spans) != len(expected.Traces[0].Spans) {
		got, _ := json.Marshal(patch)
		want, _ := json.Marshal(expected)
		t.Fatalf("PatchTraces request: got %s want %s", got, want)
	}

	n := len(patch.Traces[0].Spans)
	rootSpan := patch.Traces[0].Spans[n-1]
	for i, s := range patch.Traces[0].Spans {
		if a, b := s.StartTime, s.EndTime; a > b {
			t.Errorf("span %d start time is later than its end time (%q, %q)", i, a, b)
		}
		if a, b := rootSpan.StartTime, s.StartTime; a > b {
			t.Errorf("trace start time is later than span %d start time (%q, %q)", i, a, b)
		}
		if a, b := s.EndTime, rootSpan.EndTime; a > b {
			t.Errorf("span %d end time is later than trace end time (%q, %q)", i, a, b)
		}
		if i > 1 && i < n-1 {
			if a, b := patch.Traces[0].Spans[i-1].EndTime, s.StartTime; a > b {
				t.Errorf("span %d end time is later than span %d start time (%q, %q)", i-1, i, a, b)
			}
		}
	}

	if x := rootSpan.ParentSpanId; x != 42 {
		t.Errorf("Incorrect ParentSpanId: got %d want %d", x, 42)
	}
	for i, s := range patch.Traces[0].Spans {
		if x, y := rootSpan.SpanId, s.ParentSpanId; i < n-1 && x != y {
			t.Errorf("Incorrect ParentSpanId in span %d: got %d want %d", i, y, x)
		}
	}
	for i, s := range patch.Traces[0].Spans {
		s.EndTime = ""
		labels := &expected.Traces[0].Spans[i].Labels
		for key, value := range *labels {
			if v, ok := s.Labels[key]; !ok {
				t.Errorf("Span %d is missing Label %q:%q", i, key, value)
			} else if v != value {
				t.Errorf("Span %d Label %q: got value %q want %q", i, key, v, value)
			}
		}
		for key, _ := range s.Labels {
			if _, ok := (*labels)[key]; key != "trace.cloud.google.com/stacktrace" && !ok {
				t.Errorf("Span %d: unexpected label %q", i, key)
			}
		}
		*labels = nil
		s.Labels = nil
		s.ParentSpanId = 0
		if s.SpanId == 0 {
			t.Errorf("Incorrect SpanId: got 0 want nonzero")
		}
		s.SpanId = 0
		s.StartTime = ""
	}
	if !reflect.DeepEqual(patch, expected) {
		got, _ := json.Marshal(patch)
		want, _ := json.Marshal(expected)
		t.Errorf("PatchTraces request: got %s want %s", got, want)
	}
}

func TestNoTrace(t *testing.T) {
	testNoTrace(t, false)
}

func TestNoTraceWithWait(t *testing.T) {
	testNoTrace(t, true)
}

func testNoTrace(t *testing.T, synchronous bool) {
	for _, header := range []string{
		`0123456789ABCDEF0123456789ABCDEF/42;o=2`,
		`0123456789ABCDEF0123456789ABCDEF/42;o=0`,
		`0123456789ABCDEF0123456789ABCDEF/42`,
		`0123456789ABCDEF0123456789ABCDEF`,
		``,
	} {
		req, err := http.NewRequest("GET", "http://example.com/foo", nil)
		if header != "" {
			req.Header["X-Cloud-Trace-Context"] = []string{header}
		}
		if err != nil {
			t.Fatal(err)
		}
		rt := newFakeRoundTripper()
		traceClient := newTestClient(rt)
		uploaded := makeRequests(t, req, traceClient, rt, synchronous, false)
		if uploaded != nil {
			t.Errorf("Got a trace, expected none.")
		}
	}
}
