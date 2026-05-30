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
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"k8s.io/klog/v2"

	"k8s.io/kops/pkg/wellknownports"
)

type applyChannelReadiness struct {
	ready atomic.Bool
	addr  string // resolved listen address; read only by tests (which bind :0)
}

func (r *applyChannelReadiness) recordApplyResult(err error) {
	r.ready.Store(err == nil)
}

// serveReadiness serves /readyz on loopback for the kubelet readiness probe until ctx is cancelled:
// 200 when ready is true, 503 otherwise. The pod runs with hostNetwork, so the kubelet reaches it
// via 127.0.0.1 in the host network namespace.
func serveReadiness(ctx context.Context) (*applyChannelReadiness, error) {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(wellknownports.KopsChannelsHealthCheck))
	return serveReadinessOnAddr(ctx, addr)
}

func serveReadinessOnAddr(ctx context.Context, addr string) (*applyChannelReadiness, error) {
	readiness := &applyChannelReadiness{}

	mux := http.NewServeMux()
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if readiness.ready.Load() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok\n"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("apply iterations are failing\n"))
		}
	})

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listening on %s: %w", addr, err)
	}
	readiness.addr = listener.Addr().String()

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			klog.Fatalf("kops-channels readiness server stopped: %v", err)
		}
	}()
	return readiness, nil
}
