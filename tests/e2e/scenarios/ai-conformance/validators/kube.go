/*
Copyright The Kubernetes Authors.

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

package validators

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DynamicClient returns a dynamic client for the kubernetes cluster.
func (h *ValidatorHarness) DynamicClient() dynamic.Interface {
	if h.dynamicClient == nil {
		dynamicClient, err := dynamic.NewForConfig(h.restConfig)
		if err != nil {
			h.Fatalf("failed to create dynamic client: %v", err)
		}
		h.dynamicClient = dynamicClient
	}
	return h.dynamicClient
}

// DeviceClass is a wrapper around the DRA DeviceClass type.
type DeviceClass struct {
	u *unstructured.Unstructured
}

// Name returns the name of the device class.
func (d *DeviceClass) Name() string {
	return d.u.GetName()
}

var deviceClassGVR = schema.GroupVersionResource{
	Group:    "resource.k8s.io",
	Version:  "v1",
	Resource: "deviceclasses",
}

// ListDeviceClasses lists all device classes in the cluster.
func (h *ValidatorHarness) ListDeviceClasses() []*DeviceClass {
	objectList, err := h.DynamicClient().Resource(deviceClassGVR).List(h.Context(), metav1.ListOptions{})
	if err != nil {
		h.Fatalf("failed to list device classes: %v", err)
	}
	var out []*DeviceClass
	for i := range objectList.Items {
		out = append(out, &DeviceClass{u: &objectList.Items[i]})
	}
	return out
}

// HasDeviceClass returns true if a device class with the given name exists.
func (h *ValidatorHarness) HasDeviceClass(name string) bool {
	for _, deviceClass := range h.ListDeviceClasses() {
		if deviceClass.Name() == name {
			return true
		}
	}
	return false
}

// ResourceSlice is a wrapper around the DRA ResourceSlice type.
type ResourceSlice struct {
	u *unstructured.Unstructured
}

// Name returns the name of the resource slice.
func (d *ResourceSlice) Name() string {
	return d.u.GetName()
}

var resourceSliceGVR = schema.GroupVersionResource{
	Group:    "resource.k8s.io",
	Version:  "v1",
	Resource: "resourceslices",
}

// ListResourceSlices lists all resource slices in the cluster.
func (h *ValidatorHarness) ListResourceSlices() []*ResourceSlice {
	objectList, err := h.DynamicClient().Resource(resourceSliceGVR).List(h.Context(), metav1.ListOptions{})
	if err != nil {
		h.Fatalf("failed to list resource slices: %v", err)
	}
	var out []*ResourceSlice
	for i := range objectList.Items {
		out = append(out, &ResourceSlice{u: &objectList.Items[i]})
	}
	return out
}

// CRD is a wrapper around the CustomResourceDefinition type.
type CRD struct {
	u *unstructured.Unstructured
}

// Name returns the name of the CRD.
func (d *CRD) Name() string {
	return d.u.GetName()
}

var crdGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

// ListCRDs lists all CRDs in the cluster.
func (h *ValidatorHarness) ListCRDs() []*CRD {
	objectList, err := h.DynamicClient().Resource(crdGVR).List(h.Context(), metav1.ListOptions{})
	if err != nil {
		h.Fatalf("failed to list CRDs: %v", err)
	}
	var out []*CRD
	for i := range objectList.Items {
		out = append(out, &CRD{u: &objectList.Items[i]})
	}
	return out
}

// HasCRD returns true if a CRD with the given name exists.
func (h *ValidatorHarness) HasCRD(name string) bool {
	for _, crd := range h.ListCRDs() {
		if crd.Name() == name {
			return true
		}
	}
	return false
}

var namespaceGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "namespaces",
}

var namespaceGVK = schema.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "Namespace",
}

func (h *ValidatorHarness) TestNamespace() string {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.testNamespace == "" {
		prefix := h.t.Name()
		prefix = prefix[strings.LastIndex(prefix, "/")+1:]
		prefix = strings.ToLower(prefix)
		prefix = strings.ReplaceAll(prefix, "/", "-")
		prefix = strings.ReplaceAll(prefix, "_", "-")

		ns := fmt.Sprintf("%s-%d", prefix, time.Now().Unix())

		nsObj := &unstructured.Unstructured{}
		nsObj.SetGroupVersionKind(namespaceGVK)
		nsObj.SetName(ns)

		h.Logf("Creating test namespace %q", ns)

		if _, err := h.DynamicClient().Resource(namespaceGVR).Create(h.Context(), nsObj, metav1.CreateOptions{}); err != nil {
			h.Fatalf("failed to create test namespace: %v", err)
		}

		h.testNamespace = ns

		h.t.Cleanup(func() {
			ctx := context.WithoutCancel(h.Context())
			h.dumpNamespaceResources(ctx, ns)

			h.Logf("Deleting test namespace %q", ns)
			err := h.DynamicClient().Resource(namespaceGVR).Delete(ctx, ns, metav1.DeleteOptions{})
			if err != nil {
				h.Logf("failed to delete test namespace: %v", err)
			}
		})
	}

	return h.testNamespace
}

// ApplyManifest applies a Kubernetes manifest from the given file path to the specified namespace.
// It returns the list of objects found in the manifest.
// We use kubectl so that the output is clear and in theory someone could run the same commands themselves to debug.
func (h *ValidatorHarness) ApplyManifest(defaultNamespace string, manifestPath string) []*KubeObjectID {
	h.Logf("Applying manifest %q to namespace %q", manifestPath, defaultNamespace)

	objects, err := h.parseManifestObjects(manifestPath, defaultNamespace)
	if err != nil {
		h.Fatalf("failed to parse manifest %s: %v", manifestPath, err)
	}

	h.objectIDs = append(h.objectIDs, objects...)

	h.ShellExec(fmt.Sprintf("kubectl apply -n %s -f %s", defaultNamespace, manifestPath))

	return objects
}

// dumpNamespaceResources dumps key resources from the namespace to the artifacts directory for debugging.
func (h *ValidatorHarness) dumpNamespaceResources(ctx context.Context, ns string) {
	artifactsDir := os.Getenv("ARTIFACTS")
	if artifactsDir == "" {
		artifactsDir = "_artifacts"
	}

	testName := strings.ReplaceAll(h.t.Name(), "/", "_")
	clusterInfoDir := filepath.Join(artifactsDir, "tests", testName, "cluster-info", ns)
	if err := os.MkdirAll(clusterInfoDir, 0o755); err != nil {
		h.Logf("failed to create cluster-info directory: %v", err)
		return
	}

	resourceTypes := make(map[string]bool)
	for _, objectID := range h.objectIDs {
		gvk := objectID.GVK()
		id := fmt.Sprintf("%s.%s", gvk.Kind, gvk.Group)
		resourceTypes[id] = true
	}

	// Always include Events, Pods: they are usually not in the manifest, but are often critical for understanding failures.
	resourceTypes["Events"] = true
	resourceTypes["Pods"] = true

	for resourceType := range resourceTypes {
		filename := strings.ToLower(resourceType) + ".yaml"
		if err := h.dumpResource(ctx, ns, resourceType, filepath.Join(clusterInfoDir, filename)); err != nil {
			h.Logf("failed to dump resource %s: %v", resourceType, err)
		}
	}

	if err := h.dumpPodLogs(ctx, ns, clusterInfoDir); err != nil {
		h.Logf("failed to dump pod logs: %v", err)
	}
}

// dumpResource runs kubectl get for a resource type and writes the output to a file.
// Errors are logged but do not fail the test.
func (h *ValidatorHarness) dumpResource(ctx context.Context, ns string, resourceType string, outputPath string) error {
	args := []string{"get", resourceType}
	if ns != "" {
		args = append(args, "-n", ns)
	}
	args = append(args, "-o", "yaml")
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to dump %s in namespace %s: %v (stderr: %s)", resourceType, ns, err, stderr.String())
	}

	if err := os.WriteFile(outputPath, stdout.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write %s dump to %s: %w", resourceType, outputPath, err)
	}

	return nil
}

// dumpPodLogs collects logs from all pods in the namespace and writes them to individual files.
func (h *ValidatorHarness) dumpPodLogs(ctx context.Context, ns string, clusterInfoDir string) error {
	podLogsDir := filepath.Join(clusterInfoDir, "pod-logs")

	// List pods in the namespace
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-n", ns, "-o", "jsonpath={.items[*].metadata.name}")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to list pods for log collection in namespace %s (stderr: %s): %w", ns, stderr.String(), err)
	}

	podNames := strings.Fields(stdout.String())

	if err := os.MkdirAll(podLogsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create pod-logs directory: %v", err)
	}

	var errs []error
	for _, podName := range podNames {
		logCmd := exec.CommandContext(ctx, "kubectl", "logs", "-n", ns, podName, "--all-containers", "--ignore-errors")
		var logOut bytes.Buffer
		logCmd.Stdout = &logOut
		var logErr bytes.Buffer
		logCmd.Stderr = &logErr
		if err := logCmd.Run(); err != nil {
			errs = append(errs, fmt.Errorf("failed to get logs for pod %s (stderr: %s): %w", podName, logErr.String(), err))
			continue
		}
		logPath := filepath.Join(podLogsDir, podName+".log")
		if err := os.WriteFile(logPath, logOut.Bytes(), 0o644); err != nil {
			errs = append(errs, fmt.Errorf("failed to write logs for pod %s to %s: %w", podName, logPath, err))
		}
	}

	return errors.Join(errs...)
}
