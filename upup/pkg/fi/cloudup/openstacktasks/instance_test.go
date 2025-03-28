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

package openstacktasks

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
	"k8s.io/kops/pkg/truncate"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

func TestFilterPortsReturnsAllPorts(t *testing.T) {
	clusterName := "fakeCluster"

	allPorts := []ports.Port{
		{
			ID: "fakeID_1",
		},
		{
			ID: "fakeID_2",
		},
	}

	actualPorts := filterInstancePorts(allPorts, clusterName)

	if !reflect.DeepEqual(allPorts, actualPorts) {
		t.Fatalf("expected '%+v', but got '%+v", allPorts, actualPorts)
	}
}

func TestFilterPortsReturnsOnlyTaggedPort(t *testing.T) {
	clusterName := "fakeCluster"
	clusterNameTag := fmt.Sprintf("%s=%s", openstack.TagClusterName, clusterName)

	allPorts := []ports.Port{
		{
			ID: "fakeID_1",
		},
		{
			ID: "fakeID_2",
			Tags: []string{
				clusterNameTag,
			},
		},
	}

	expectedPorts := []ports.Port{
		allPorts[1],
	}
	actualPorts := filterInstancePorts(allPorts, clusterName)

	if !reflect.DeepEqual(expectedPorts, actualPorts) {
		t.Fatalf("expected '%+v', but got '%+v", expectedPorts, actualPorts)
	}
}

func TestFilterPortsReturnsOnlyTaggedPorts(t *testing.T) {
	clusterName := "fakeCluster"
	clusterNameTag := fmt.Sprintf("%s=%s", openstack.TagClusterName, clusterName)

	allPorts := []ports.Port{
		{
			ID: "fakeID_1",
			Tags: []string{
				clusterNameTag,
			},
		},
		{
			ID: "fakeID_2",
		},
		{
			ID: "fakeID_3",
			Tags: []string{
				clusterNameTag,
			},
		},
	}

	expectedPorts := []ports.Port{
		allPorts[0],
		allPorts[2],
	}
	actualPorts := filterInstancePorts(allPorts, clusterName)

	if !reflect.DeepEqual(expectedPorts, actualPorts) {
		t.Fatalf("expected '%+v', but got '%+v", expectedPorts, actualPorts)
	}
}

func TestFilterPortsReturnsOnlyTaggedPortsWithLongClustername(t *testing.T) {
	clusterName := "tom-software-dev-playground-real33-k8s-local"
	clusterNameTag := truncate.TruncateString(fmt.Sprintf("%s=%s", openstack.TagClusterName, clusterName), TRUNCATE_OPT)

	allPorts := []ports.Port{
		{
			ID: "fakeID_1",
			Tags: []string{
				clusterNameTag,
			},
		},
		{
			ID: "fakeID_2",
		},
		{
			ID: "fakeID_3",
			Tags: []string{
				clusterNameTag,
			},
		},
	}

	expectedPorts := []ports.Port{
		allPorts[0],
		allPorts[2],
	}
	actualPorts := filterInstancePorts(allPorts, clusterName)

	if !reflect.DeepEqual(expectedPorts, actualPorts) {
		t.Fatalf("expected '%+v', but got '%+v", expectedPorts, actualPorts)
	}
}
