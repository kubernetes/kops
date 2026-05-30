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

package cmd

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServeReadinessReportsApplyOutcome(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	readiness, err := serveReadinessOnAddr(ctx, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("serveReadinessOnAddr returned error: %v", err)
	}

	assertReadinessStatus(t, readiness.addr, http.StatusServiceUnavailable)

	readiness.recordApplyResult(nil)
	assertReadinessStatus(t, readiness.addr, http.StatusOK)

	readiness.recordApplyResult(errors.New("apply failed"))
	assertReadinessStatus(t, readiness.addr, http.StatusServiceUnavailable)
}

func TestServeReadinessReturnsBindError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve test port: %v", err)
	}
	defer listener.Close()

	_, err = serveReadinessOnAddr(context.Background(), listener.Addr().String())
	if err == nil {
		t.Fatalf("expected bind error")
	}
}

func assertReadinessStatus(t *testing.T, addr string, expectedStatus int) {
	t.Helper()

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + addr + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}
