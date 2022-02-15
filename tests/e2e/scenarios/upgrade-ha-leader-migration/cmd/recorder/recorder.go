/*
Copyright 2022 The Kubernetes Authors.

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
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/klog/v2"
)

const portForwardSubresource = "portforward"

var ErrConflictDetected = fmt.Errorf("conflict detected")

type recorder struct {
	config         *rest.Config
	clientset      *kubernetes.Clientset
	insecureClient *http.Client
}

type assignment struct {
	controller string
	pod        string
}

// Pods lists pods belonging to KCM and CCM
// because CCM labels are inconsistent among providers, use the pod name instead
func (r *recorder) Pods(ctx context.Context) (kcmPods []v1.Pod, ccmPods []v1.Pod, err error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second) // should be a quick request
	defer cancel()
	pods, err := r.clientset.CoreV1().Pods(Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, KCMPrefix) {
			kcmPods = append(kcmPods, pod)
			continue
		}
		if strings.HasPrefix(pod.Name, CCMPrefix) {
			ccmPods = append(ccmPods, pod)
			continue
		}
		// we don't care about other pods
	}
	return
}

// healthCheckPort returns the port for /healthz by parsing the livenessProbe
// field
func healthCheckPort(pod v1.Pod) (int, error) {
	if len(pod.Spec.Containers) == 0 {
		return 0, fmt.Errorf("pod has no container")
	}
	container := pod.Spec.Containers[0]
	if container.LivenessProbe == nil {
		return 0, fmt.Errorf("container has no liveness probe")
	}
	httpGet := container.LivenessProbe.HTTPGet
	if httpGet == nil {
		return 0, fmt.Errorf("container has no HttpGet probe")
	}
	if httpGet.Port.Type != intstr.Int {
		return 0, fmt.Errorf("container has no numberic probe port")
	}
	return httpGet.Port.IntValue(), nil
}

func (r *recorder) Observe(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	kcmPods, ccmPods, err := r.Pods(ctx)
	if err != nil {
		klog.ErrorS(err, "cannot fetch pods")
	}
	klog.InfoS("starting check", "kcm-pods", len(kcmPods), "ccm-pods", len(ccmPods))
	aChan := make(chan assignment, 128)
	var wg sync.WaitGroup
	observe := func(ctx context.Context, pod v1.Pod, counter *int64, aChan chan<- assignment) {
		defer wg.Done()
		b, err := r.HealthCheck(ctx, pod)
		if err != nil {
			klog.ErrorS(err, "unhealthy", "pod", pod.Name)
		} else {
			klog.InfoS("healthy", "pod", pod.Name)
			controllers, _ := parseHealthz(b)
			for _, controller := range controllers {
				aChan <- assignment{
					controller: controller,
					pod:        pod.Name,
				}
			}
			atomic.AddInt64(counter, 1)
		}
	}
	var kcmHealthy, ccmHealthy int64
	wg.Add(len(kcmPods))
	for _, pod := range kcmPods {
		go observe(ctx, pod, &kcmHealthy, aChan)
	}
	wg.Add(len(ccmPods))
	for _, pod := range ccmPods {
		go observe(ctx, pod, &ccmHealthy, aChan)
	}
	go func() {
		wg.Wait()
		close(aChan)
	}()
	assignments := make(map[string][]string) // controller -> pods map
	for a := range aChan {
		assignments[a.controller] = append(assignments[a.controller], a.pod)
	}
	var conflictingPods sets.String
	for c, pods := range assignments {
		if len(pods) == 1 {
			klog.InfoS("observed", "controller", c, "pods", pods)
		} else {
			klog.InfoS("conflict", "controller", c, "pods", pods)
			conflictingPods.Insert(pods...)
		}
	}
	if conflictingPods.Len() != 0 {
		klog.InfoS("controller may be running under multiple controller managers", "pods", conflictingPods.List())
		return ErrConflictDetected
	}
	return nil
}

// HealthCheck performs a health check on the given Pod by port-forwarding into the Pod
// returns the content from /healthz
func (r *recorder) HealthCheck(ctx context.Context, pod v1.Pod) (string, error) {
	req := r.clientset.RESTClient().Post().Resource("pods").Namespace(Namespace).
		Name(pod.Name).SubResource(portForwardSubresource)
	stopChan, readyChan := make(chan struct{}), make(chan struct{})
	errChan := make(chan error)
	defer close(stopChan)
	transport, upgrader, err := spdy.RoundTripperFor(r.config)
	if err != nil {
		return "", err
	}
	url := req.URL()
	url.Path = "/api/v1" + url.Path // manually add the api root
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, url)
	port, err := healthCheckPort(pod)
	if err != nil {
		return "", err
	}
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", 0, port)}, stopChan, readyChan, io.Discard, io.Discard)
	if err != nil {
		return "", err
	}
	go func() {
		err := fw.ForwardPorts()
		if err != nil {
			errChan <- err
		}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errChan:
		return "", err
	case <-readyChan:
		ports, err := fw.GetPorts()
		if err != nil {
			return "", err
		}
		localPort := ports[0].Local
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("https://localhost:%d/healthz?verbose=1", localPort),
			nil)
		if err != nil {
			return "", err
		}
		resp, err := r.insecureClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("health check status %d", resp.StatusCode)
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

func NewRecorder() (*recorder, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	insecureClient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	return &recorder{config: config, clientset: clientset, insecureClient: insecureClient}, nil
}

func parseHealthz(body string) (controllers []string, err error) {
	const leaderElectionHealthCheck = "leaderElection"
	const healthCheckPrefix = "[+]"
	const healthCheckSuffix = " ok"
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "[+]") {
			controller := strings.TrimSuffix(strings.TrimPrefix(line, healthCheckPrefix), healthCheckSuffix)
			if controller != leaderElectionHealthCheck {
				controllers = append(controllers, controller)
			}
		}
	}
	return
}
