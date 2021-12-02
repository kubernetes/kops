/*
Copyright 2021 The Kubernetes Authors.

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

package watchers

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/kops/dns-controller/pkg/dns"
)

func TestPodController(t *testing.T) {
	ctx := context.Background()
	pspec := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "somepod",
			Namespace: "kube-system",
			Annotations: map[string]string{
				"dns.alpha.kubernetes.io/internal": "internal.a.foo.com",
				"dns.alpha.kubernetes.io/external": "a.foo.com",
			},
		},
		Spec: corev1.PodSpec{
			HostNetwork: true,
			NodeName:    "my-node",
		},
		Status: corev1.PodStatus{
			PodIP: "10.0.0.1",
			PodIPs: []corev1.PodIP{
				{IP: "10.0.0.1"},
				{IP: "2001:db8:0:0:0:ff00:42:8329"},
			},
		},
	}

	client := fake.NewSimpleClientset()
	pods := client.CoreV1().Pods("kube-system")

	_, err := pods.Create(ctx, pspec, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ch := make(chan struct{})
	scope := &fakeScope{
		readyCh: ch,
		records: make(map[string][]dns.Record),
	}

	dnsctx := &fakeDNSContext{
		scope: scope,
	}

	c, err := NewPodController(client, dnsctx, "kube-system")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go c.Run()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("update was not marked as complete")
	}

	c.Stop()

	want := map[string][]dns.Record{
		"kube-system/somepod": {
			{RecordType: "_alias", FQDN: "a.foo.com.", Value: "node/my-node/external"},
			{RecordType: "_alias", FQDN: "internal.a.foo.com.", Value: "node/my-node/internal"},
		},
	}
	if diff := cmp.Diff(scope.records, want); diff != "" {
		t.Fatalf("generated records did not match expected; diff=%s", diff)
	}
}
