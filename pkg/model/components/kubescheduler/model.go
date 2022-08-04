/*
Copyright 2019 The Kubernetes Authors.

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

package kubescheduler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/reflectutils"
	"sigs.k8s.io/yaml"
)

// KubeSchedulerConfigPath is the path where we write the kube-scheduler config file (on the control-plane nodes)
const KubeSchedulerConfigPath = "/var/lib/kube-scheduler/config.yaml"

// Kubeconfig is the path where we write the kube-scheduler kubeconfig file (on the control-plane nodes)
const KubeConfigPath = "/var/lib/kube-scheduler/kubeconfig"

// KubeSchedulerBuilder builds the configuration file for kube-scheduler
type KubeSchedulerBuilder struct {
	*model.KopsModelContext
	Lifecycle    fi.Lifecycle
	AssetBuilder *assets.AssetBuilder
}

var _ fi.ModelBuilder = &KubeSchedulerBuilder{}

// Build creates the tasks relating to kube-scheduler
func (b *KubeSchedulerBuilder) Build(c *fi.ModelBuilderContext) error {
	configYAML, err := b.buildSchedulerConfig()
	if err != nil {
		return err
	}

	b.AssetBuilder.StaticFiles = append(b.AssetBuilder.StaticFiles, &assets.StaticFile{
		Path:    KubeSchedulerConfigPath,
		Content: string(configYAML),
		Roles:   []kops.InstanceGroupRole{kops.InstanceGroupRoleMaster, kops.InstanceGroupRoleAPIServer},
	})
	return nil
}

func (b *KubeSchedulerBuilder) buildSchedulerConfig() ([]byte, error) {
	var matches []*kubemanifest.Object
	for _, additionalObject := range b.AdditionalObjects {
		gvk := additionalObject.GroupVersionKind()
		if gvk.Group != "kubescheduler.config.k8s.io" {
			continue
		}
		if gvk.Kind != "KubeSchedulerConfiguration" {
			continue
		}
		matches = append(matches, additionalObject)
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("found multiple KubeSchedulerConfiguration objects in cluster configuration; expected at most one")
	}

	var config *unstructured.Unstructured
	if len(matches) == 1 {
		config = matches[0].ToUnstructured()
	} else {
		config = &unstructured.Unstructured{}
		config.SetKind("KubeSchedulerConfiguration")
		if b.IsKubernetesGTE("1.22") {
			config.SetAPIVersion("kubescheduler.config.k8s.io/v1beta2")
		} else {
			config.SetAPIVersion("kubescheduler.config.k8s.io/v1beta1")
		}
		// We need to store the object, because we are often called repeatedly (until we converge)
		b.AdditionalObjects = append(b.AdditionalObjects, kubemanifest.NewObject(config.Object))
	}

	// TODO: Handle different versions? e.g. gvk := config.GroupVersionKind()

	if err := unstructured.SetNestedField(config.Object, KubeConfigPath, "clientConnection", "kubeconfig"); err != nil {
		return nil, fmt.Errorf("error setting clientConnection.kubeconfig in kube-scheduler configuration: %w", err)
	}

	kubeScheduler := b.Cluster.Spec.KubeScheduler
	if kubeScheduler != nil {
		if err := MapToUnstructured(kubeScheduler, config); err != nil {
			return nil, err
		}
	}

	configYAML, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}
	return configYAML, nil
}

// MapToUnstructured reflects the options interface and extracts the parameters for the config file
func MapToUnstructured(options interface{}, target *unstructured.Unstructured) error {
	setValue := func(targetPath string, val interface{}) error {
		fields := strings.Split(targetPath, ".")
		// Cannot use unstructured.SetNestedField, because it fails with e.g. "cannot deep copy int32"
		parent := target.Object
		for i := 0; i < len(fields)-1; i++ {
			v := parent[fields[i]]
			if v == nil {
				v = make(map[string]interface{})
				parent[fields[i]] = v
			}
			m, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("value was not a map at position %d in %s", i, targetPath)
			}
			parent = m
		}
		parent[fields[len(fields)-1]] = val
		return nil
	}

	walker := func(path *reflectutils.FieldPath, field *reflect.StructField, val reflect.Value) error {
		if field == nil {
			klog.V(8).Infof("ignoring non-field: %s", path)
			return nil
		}

		tag := field.Tag.Get("config")
		if tag == "" {
			klog.V(4).Infof("not writing field with no config tag: %s", path)
			// We want to descend - it could be a structure containing flags
			return nil
		}
		if tag == "-" {
			klog.V(4).Infof("skipping field with %q config tag: %s", tag, path)
			return reflectutils.SkipReflection
		}

		tagTokens := strings.Split(tag, ",")
		omitEmpty := false
		for _, token := range tagTokens {
			if token == "omitempty" {
				omitEmpty = true
			}
		}
		targetPath := tagTokens[0]

		// We do have to do this, even though the recursive walk will do it for us
		// because when we descend we won't have `field` set
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return nil
			}
		}

		isEmpty := val.IsZero()

		if !isEmpty || !omitEmpty {
			switch v := val.Interface().(type) {
			case *resource.Quantity:
				floatVal, err := strconv.ParseFloat(v.AsDec().String(), 64)
				if err != nil {
					return fmt.Errorf("unable to convert from Quantity %v to float", v)
				}
				if err := setValue(targetPath, floatVal); err != nil {
					return err
				}
				// Clear the field, so we don't set the flag
				val.Set(reflect.ValueOf(nil))
			default:
				if err := setValue(targetPath, val.Interface()); err != nil {
					return err
				}
				// Clear the field, so we don't set the flag
				empty := reflect.New(val.Type()).Elem()
				val.Set(empty)
			}
		}

		return reflectutils.SkipReflection
	}

	err := reflectutils.ReflectRecursive(reflect.ValueOf(options), walker, &reflectutils.ReflectOptions{DeprecatedDoubleVisit: true, JSONNames: true})
	if err != nil {
		return fmt.Errorf("error walking over %T: %w", options, err)
	}

	return nil
}
