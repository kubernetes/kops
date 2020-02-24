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

package drain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func getTestSetup(objects []runtime.Object) *fake.Clientset {
	k8sClient := fake.NewSimpleClientset(objects...)
	return k8sClient
}

func dummyDaemonSet(name string) appsv1.DaemonSet {
	return appsv1.DaemonSet{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
		},
	}
}

func dummyPod(podMap map[string]string) v1.Pod {
	pod := v1.Pod{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      podMap["name"],
			Namespace: "kube-system",
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
		},
		Status: v1.PodStatus{
			Phase: v1.PodPhase(podMap["phase"]),
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:  "container1",
					Ready: podMap["ready"] == "true",
				},
			},
		},
	}
	if podMap["controller"] == "DaemonSet" {
		pod.SetOwnerReferences([]v1meta.OwnerReference{
			{
				Name:       podMap["controllerName"],
				Kind:       appsv1.SchemeGroupVersion.WithKind("DaemonSet").Kind,
				Controller: &[]bool{true}[0],
			},
		})
	}
	return pod
}

// MakePodList constructs api.PodList from a list of pod attributes
func makePodList(pods []map[string]string) []runtime.Object {
	var list []runtime.Object
	for _, pod := range pods {
		p := dummyPod(pod)
		list = append(list, &p)
	}
	return list
}

func TestRollingUpdateDaemonSetMixedPods(t *testing.T) {
	objects := makePodList(
		[]map[string]string{
			{
				"name":           "pod1",
				"ready":          "true",
				"phase":          string(v1.PodRunning),
				"controller":     "DaemonSet",
				"controllerName": "ds1",
			},
			{
				"name":  "pod2",
				"ready": "true",
				"phase": string(v1.PodRunning),
			},
		},
	)
	ds := dummyDaemonSet("ds1")
	objects = append(objects, &ds)
	k8sClient := getTestSetup(objects)
	helper := Helper{
		Client:              k8sClient,
		Force:               true,
		IgnoreAllDaemonSets: false,
		DeleteLocalData:     true,
	}

	podList, _ := helper.GetPodsForDeletion("node1")
	assert.True(t, podList.Warnings() != "")
	assert.True(t, len(podList.errors()) == 0)
	assert.True(t, podList.items[0].status.priority == podDeletionPriorityHighest)
	assert.True(t, podList.items[1].status.priority == podDeletionPriorityLowest)
	assert.NotNil(t, podList)
}

func TestRollingUpdateNoDaemonSets(t *testing.T) {
	objects := makePodList(
		[]map[string]string{
			{
				"name":  "pod1",
				"ready": "true",
				"phase": string(v1.PodRunning),
			},
			{
				"name":  "pod2",
				"ready": "true",
				"phase": string(v1.PodRunning),
			},
		},
	)
	k8sClient := getTestSetup(objects)
	helper := Helper{
		Client:              k8sClient,
		Force:               true,
		IgnoreAllDaemonSets: false,
		DeleteLocalData:     true,
	}

	podList, _ := helper.GetPodsForDeletion("node1")
	assert.True(t, podList.Warnings() != "")
	assert.True(t, len(podList.errors()) == 0)
	assert.True(t, podList.items[0].status.priority == podDeletionPriorityHighest)
	assert.True(t, podList.items[1].status.priority == podDeletionPriorityHighest)
	assert.NotNil(t, podList)
}

func TestRollingUpdateAllDaemonSetPods(t *testing.T) {
	objects := makePodList(
		[]map[string]string{
			{
				"name":           "pod1",
				"ready":          "true",
				"phase":          string(v1.PodRunning),
				"controller":     "DaemonSet",
				"controllerName": "ds1",
			},
			{
				"name":           "pod2",
				"ready":          "true",
				"phase":          string(v1.PodRunning),
				"controller":     "DaemonSet",
				"controllerName": "ds1",
			},
		},
	)
	ds := dummyDaemonSet("ds1")
	objects = append(objects, &ds)
	k8sClient := getTestSetup(objects)
	helper := Helper{
		Client:              k8sClient,
		Force:               true,
		IgnoreAllDaemonSets: false,
		DeleteLocalData:     true,
	}

	podList, _ := helper.GetPodsForDeletion("node1")
	assert.True(t, podList.Warnings() == "")
	assert.True(t, len(podList.errors()) == 0)
	assert.True(t, podList.items[0].status.priority == podDeletionPriorityLowest)
	assert.True(t, podList.items[1].status.priority == podDeletionPriorityLowest)
	assert.NotNil(t, podList)
}
