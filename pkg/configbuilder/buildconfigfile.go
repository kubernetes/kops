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

package configbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"

	"k8s.io/klog"
	"k8s.io/kops/util/pkg/reflectutils"
)

// ClientConnectionConfig for kube-scheduler
type ClientConnectionConfig struct {
	Burst      *int32   `yaml:"burst,omitempty"`
	Kubeconfig *string  `yaml:"kubeconfig"`
	QPS        *float32 `yaml:"qps,omitempty"`
}

// SchedulerConfig used to generate the config file
type SchedulerConfig struct {
	APIVersion         string                  `yaml:"apiVersion"`
	Kind               string                  `yaml:"Kind"`
	BindTimeoutSeconds *int64                  `yaml:"bindTimeoutSeconds,omitempty"`
	ClientConnection   *ClientConnectionConfig `yaml:"clientConnection,omitempty"`
}

// BuildConfigYaml reflects the options interface and extracts the parameters for the config file
func BuildConfigYaml(options interface{}) ([]byte, error) {

	schedConfig := new(SchedulerConfig)
	schedConfig.APIVersion = "kubescheduler.config.k8s.io/v1alpha1"
	schedConfig.Kind = "KubeSchedulerConfiguration"
	schedConfig.ClientConnection = new(ClientConnectionConfig)

	walker := func(path string, field *reflect.StructField, val reflect.Value) error {
		if field == nil {
			klog.V(8).Infof("ignoring non-field: %s", path)
			return nil
		}
		tag := field.Tag.Get("configfile")
		if tag == "" {
			klog.V(4).Infof("not writing field with no flag tag: %s", path)
			// We want to descend - it could be a structure containing flags
			return nil
		}
		if tag == "-" {
			klog.V(4).Infof("skipping field with %q flag tag: %s", tag, path)
			return reflectutils.SkipReflection
		}

		tokens := strings.Split(tag, ",")

		flagName := tokens[0]

		targetValue, error := getValueFromStruct(flagName, schedConfig)
		if error != nil {
			return fmt.Errorf("conversion error for field %s: %s", flagName, error)
		}
		// We do have to do this, even though the recursive walk will do it for us
		// because when we descend we won't have `field` set
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return nil
			}
		}
		targetValue.Set(val)

		return reflectutils.SkipReflection
	}

	err := reflectutils.ReflectRecursive(reflect.ValueOf(options), walker)
	if err != nil {
		return nil, fmt.Errorf("BuildFlagsList to reflect value: %s", err)
	}

	configFile, err := yaml.Marshal(schedConfig)
	if err != nil {
		return nil, err
	}

	return configFile, nil
}

func getValueFromStruct(keyWithDots string, object *SchedulerConfig) (*reflect.Value, error) {
	keySlice := strings.Split(keyWithDots, ".")
	v := reflect.ValueOf(object)
	// iterate through field names ,ignore the first name as it might be the current instance name
	// you can make it recursive also if want to support types like slice,map etc along with struct
	for _, key := range keySlice {
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		// we only accept structs
		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("only accepts structs; got %T", v)
		}
		v = v.FieldByName(key)
	}

	return &v, nil
}
