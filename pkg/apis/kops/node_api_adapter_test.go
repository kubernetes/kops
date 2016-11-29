/*
Copyright 2016 The Kubernetes Authors.

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

package kops

import (
	"testing"
	//"time"
	//"github.com/golang/glog"
	//k8sapi "k8s.io/kubernetes/pkg/api"
	//"k8s.io/kubernetes/pkg/client/unversioned/testclient"
	//"k8s.io/kubernetes/pkg/client/unversioned/testclient/simple"
	//"k8s.io/kubernetes/pkg/api/testapi"
)

func TestBuildNodeAPIAdapter(t *testing.T) {

}

func TestGetReadySchedulableNodes(t *testing.T) {

}

// TODO not working since they changed the darn api

/*
func TestWaitForNodeToBeReady(t *testing.T) {
	conditions := make([]k8sapi.NodeCondition,1)
	conditions[0] = k8sapi.NodeCondition{Type:"Ready",Status:"True"}

	nodeAA := setupNodeAA(t,conditions)

	test, err := nodeAA.WaitForNodeToBeReady()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if test != true {
		t.Fatalf("unexpected error WaitForNodeToBeReady Failed: %v", test)
	}
}


func TestWaitForNodeToBeNotReady(t *testing.T) {
	conditions := make([]k8sapi.NodeCondition,1)
	conditions[0] = k8sapi.NodeCondition{Type:"Ready",Status:"False"}

	nodeAA := setupNodeAA(t,conditions)

	test, err := nodeAA.WaitForNodeToBeNotReady()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if test != true {
		t.Fatalf("unexpected error WaitForNodeToBeReady Failed: %v", test)
	}
}

func TestIsNodeConditionUnset(t *testing.T) {

}

func setupNodeAA(t *testing.T, conditions []k8sapi.NodeCondition)(*NodeAPIAdapter) {

	c := &simple.Client{
		Request: simple.Request{
			Method: "GET",
			Path:   testapi.Default.ResourcePath(getNodesResourceName(), "", "foo"),
		},
		Response: simple.Response{
			StatusCode: 200,
			Body: &k8sapi.Node{
				ObjectMeta: k8sapi.ObjectMeta{Name: "node-foo"},
				Spec: k8sapi.NodeSpec{ Unschedulable: false },
				Status: k8sapi.NodeStatus{ Conditions: conditions},
			},
		},
	}
	c.Setup(t).Clientset.Nodes().Get("foo")
	//c.Validate(t, response, err)
	return &NodeAPIAdapter{
		client:   c.Clientset,
		timeout:  time.Duration(10)*time.Second,
		nodeName: "foo",
	}
}
*/

/*
func mockClient() *testclient.Fake {
	return testclient.NewSimpleFake(dummyNode())
}

// Create a NodeAPIAdapter with K8s client based on the current kubectl config
func buildMockNodeAPIAdapter(nodeName string, t *testing.T) *NodeAPIAdapter {
	s := simple.Client{}
	c := s.Setup(t)
	c.Nodes().Create(dummyNode())
	node, err := c.Client.Nodes().Get("foo")
	glog.V(4).Infof("node call %v, %v", node, err)
	return &NodeAPIAdapter{
		client:   c.Client,
		timeout:  time.Duration(10)*time.Second,
		nodeName: nodeName,
	}
}

func dummyNode() *api.Node {
	return &api.Node{
		ObjectMeta: api.ObjectMeta{
			Name: "foo",
		},
		Spec: api.NodeSpec{
			Unschedulable: false,
		},
	}
}*/

func getNodesResourceName() string {
	return "nodes"
}

/// Example mocking of api
/*
type secretsClient struct {
	unversioned.Interface
}

// dummySecret generates a secret with one user inside the auth key
// foo:md5(bar)
func dummySecret() *api.Secret {
	return &api.Secret{
		ObjectMeta: api.ObjectMeta{
			Namespace: api.NamespaceDefault,
			Name:      "demo-secret",
		},
		Data: map[string][]byte{"auth": []byte("foo:$apr1$OFG3Xybp$ckL0FHDAkoXYIlH9.cysT0")},
	}
}

func mockClient() *testclient.Fake {
	return testclient.NewSimpleFake(dummySecret())
}

func TestIngressWithoutAuth(t *testing.T) {
	ing := buildIngress()
	client := mockClient()
	_, err := ParseAnnotations(client, ing, "")
	if err == nil {
		t.Error("Expected error with ingress without annotations")
	}

	if err == ErrMissingAuthType {
		t.Errorf("Expected MissingAuthType error but returned %v", err)
	}
}


*/
