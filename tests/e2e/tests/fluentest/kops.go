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

package fluentest

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

type GenericObject struct {
	obj *unstructured.Unstructured
}

type KopsInstanceGroup struct {
	GenericObject
}

func executeCommand(cmd *exec.Cmd) ([]byte, []byte, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		klog.Infof("stdout: %v", stdout.String())
		klog.Infof("stderr: %v", stderr.String())
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf("error running command %s: %w", strings.Join(cmd.Args, " "), err)
	}
	return stdout.Bytes(), stderr.Bytes(), nil
}

func KopsGetInstanceGroups(clusterName string) ([]*KopsInstanceGroup, error) {
	cmd := exec.Command("kops", "get", "ig", "--name", clusterName, "-oyaml")
	stdout, _, err := executeCommand(cmd)
	if err != nil {
		return nil, err
	}

	objs, err := ParseGenericObjectList(stdout)
	if err != nil {
		return nil, err
	}
	var igs []*KopsInstanceGroup
	for _, obj := range objs {
		igs = append(igs, &KopsInstanceGroup{
			GenericObject: obj,
		})
	}

	return igs, nil
}

func (g *KopsInstanceGroup) Name() string {
	return g.obj.GetName()
}

func (g *KopsInstanceGroup) String() string {
	return "KopsInstanceGroup:" + g.Name()
}

func (g *KopsInstanceGroup) IsControlPlane() bool {
	return g.Role() == "Master"
}

func (g *KopsInstanceGroup) Role() string {
	return g.StringField("", "spec", "role")
}

func (g *KopsInstanceGroup) MinSize() int {
	return g.IntField(0, "spec", "minSize")
}
func (g *KopsInstanceGroup) MaxSize() int {
	return g.IntField(0, "spec", "maxSize")
}

func ParseGenericObjectList(data []byte) ([]GenericObject, error) {
	var objs []GenericObject
	for _, b := range bytes.Split(data, []byte("\n---\n")) {
		u := &unstructured.Unstructured{}
		if err := yaml.Unmarshal(b, &u.Object); err != nil {
			return nil, fmt.Errorf("error parsing instance group: %w", err)
		}
		objs = append(objs, GenericObject{obj: u})
	}
	return objs, nil
}

func (o *GenericObject) IntField(defaultValue int, fields ...string) int {
	v, found, err := unstructured.NestedFloat64(o.obj.Object, fields...)
	if err != nil {
		klog.Warningf("error fetching fields %s: %v", strings.Join(fields, "."), err)
		return defaultValue
	}
	if !found {
		return defaultValue
	}
	return int(v)
}

func (o *GenericObject) StringField(defaultValue string, fields ...string) string {
	v, found, err := unstructured.NestedString(o.obj.Object, fields...)
	if err != nil {
		klog.Warningf("error fetching fields %s: %v", strings.Join(fields, "."), err)
		return defaultValue
	}
	if !found {
		return defaultValue
	}
	return v
}
