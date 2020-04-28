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

package k8sclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	apimachinerynet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

const maxTimeout = 15 * time.Second

// Interface is a wrapper around kubernetes.Interface, that recovers
// better from network failures.  This error handling is only
// available on the helper methods, but for compatability RawClient
// exposes the underlying kubernetes.Interface
type Interface interface {
	// RawClient returns the current kubernetes.Interface; it should not be cached
	// Using wrapper methods is preferrable because we can do richer retry logic and error handling.
	// Deprecated: use wrapper methods instead; this is used to ease the transition.
	RawClient() kubernetes.Interface

	// DeleteNode wraps CoreV1.Nodes.Delete
	DeleteNode(ctx context.Context, nodeName string) error

	// ListNodes wraps CoreV1.Nodes.List
	ListNodes(ctx context.Context) (*corev1.NodeList, error)
}

var _ Interface = &client{}

type client struct {
	inner kubernetes.Interface
}

// NewForConfig creates a client for the specified rest.Config config.
// It is a wrapper around kubernetes.NewForConfig, but ensures a short
// timeout.
func NewForConfig(config *rest.Config) (Interface, error) {
	// Set a lower timeout, to work around
	// https://github.com/kubernetes/client-go/issues/374 We
	// trigger a timeout error and then to recover we need to
	// reset any existing connections
	if config.Timeout < maxTimeout {
		config.Timeout = maxTimeout
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &client{inner: c}, nil
}

func (c *client) RawClient() kubernetes.Interface {
	return c.inner
}

func (c *client) ListNodes(ctx context.Context) (*corev1.NodeList, error) {
	client := c.RawClient()
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	c.handleError(err)
	return nodes, err
}

func (c *client) DeleteNode(ctx context.Context, nodeName string) error {
	client := c.RawClient()
	err := client.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{})
	c.handleError(err)
	return err
}

// TaintNode applies the taint to the specified node
func TaintNode(ctx context.Context, k8sClient Interface, node *corev1.Node, taint corev1.Taint) error {
	oldData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	node.Spec.Taints = append(node.Spec.Taints, taint)

	newData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, node)
	if err != nil {
		return err
	}

	_, err = k8sClient.RawClient().CoreV1().Nodes().Patch(ctx, node.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if c, ok := k8sClient.(*client); ok {
		c.handleError(err)
	} else {
		klog.Warningf("unexpected type for client %T", c)
	}
	return err
}

// handleError is called with all potential errors.
// When it observes a timeout error, it closes all open / idle connections.
func (c *client) handleError(err error) {
	if err == nil {
		return
	}

	isTimeout := false
	s := err.Error()
	// TODO: move this to the chained errors framework when that's available
	if strings.Contains(s, "Client.Timeout exceeded while awaiting headers") {
		isTimeout = true
	}

	if isTimeout {
		restClientInterface := c.inner.CoreV1().RESTClient()
		restClient, ok := restClientInterface.(*rest.RESTClient)
		if !ok {
			klog.Warningf("client timed out, but rest client was not of expected type, was %T", restClientInterface)
			return
		}

		httpTransport := findHTTPTransport(restClient.Client.Transport)
		if httpTransport == nil {
			klog.Warningf("client timed out, but http transport was not of expected type, was %T", restClient.Client.Transport)
			return
		}
		httpTransport.CloseIdleConnections()
		klog.Infof("client timed out; reset connections")
	}
}

// findHTTPTransport returns the http.Transport under a RoundTripper.
// If it cannot be determined, it returns nil.
func findHTTPTransport(transport http.RoundTripper) *http.Transport {
	httpTransport, ok := transport.(*http.Transport)
	if ok {
		return httpTransport
	}

	wrapper, ok := transport.(apimachinerynet.RoundTripperWrapper)
	if ok {
		wrapped := wrapper.WrappedRoundTripper()
		if wrapped != nil {
			return findHTTPTransport(wrapped)
		}
	}

	return nil
}
