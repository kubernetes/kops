/*
Copyright 2017 The Kubernetes Authors.

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

package fi

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/testutils/testcontext"
)

func Test_tryResourceAsString(t *testing.T) {
	var sr *StringResource
	grid := []struct {
		Resource interface{}
		Expected string
	}{
		{
			Resource: NewStringResource("hello"),
			Expected: "hello",
		},
		{
			Resource: sr,
			Expected: "",
		},
		{
			Resource: nil,
			Expected: "",
		},
	}
	for i, g := range grid {
		v := reflect.ValueOf(g.Resource)
		actual, _ := tryResourceAsString(v)
		if actual != g.Expected {
			t.Errorf("unexpected result from %d.  Expected=%q, got %q", i, g.Expected, actual)
		}
	}
}

type testTask struct {
	Name      *string
	Lifecycle Lifecycle
	Tags      map[string]string
}

var _ CloudupTask = &testTask{}

func (*testTask) Run(_ *CloudupContext) error {
	panic("not implemented")
}

func Test_DryrunTarget_PrintReport(t *testing.T) {
	ctx := testcontext.ForTest(t)
	builder := assets.NewAssetBuilder(&api.Cluster{
		Spec: api.ClusterSpec{
			KubernetesVersion: "1.17.3",
		},
	}, false)
	var stdout bytes.Buffer
	target := newDryRunTarget[CloudupSubContext](builder, &stdout)
	context, err := NewCloudupContext(ctx, target, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("error from NewCloudupContext: %v", err)
	}
	tasks := map[string]CloudupTask{}
	a := &testTask{
		Name:      PtrTo("TestName"),
		Lifecycle: LifecycleSync,
		Tags:      map[string]string{"key": "value"},
	}
	e := &testTask{
		Name:      PtrTo("TestName"),
		Lifecycle: LifecycleSync,
		Tags:      map[string]string{"key": "value"},
	}
	changes := reflect.New(reflect.TypeOf(e).Elem()).Interface().(CloudupTask)
	_ = BuildChanges(a, e, changes)
	err = target.Render(context, a, e, changes)
	tasks[*e.Name] = e
	assert.NoError(t, err, "target.Render()")

	var out bytes.Buffer
	err = target.PrintReport(tasks, &out)
	assert.NoError(t, err, "target.PrintReport()")
}
