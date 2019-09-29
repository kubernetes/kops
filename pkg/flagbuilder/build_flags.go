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

package flagbuilder

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kops/util/pkg/reflectutils"
)

// BuildFlags returns a space separated list arguments
// @deprecated: please use BuildFlagsList
func BuildFlags(options interface{}) (string, error) {
	flags, err := BuildFlagsList(options)
	if err != nil {
		return "", err
	}

	return strings.Join(flags, " "), nil
}

// BuildFlagsList reflects the options interface and extracts the flags from struct tags
func BuildFlagsList(options interface{}) ([]string, error) {
	var flags []string

	walker := func(path string, field *reflect.StructField, val reflect.Value) error {
		if field == nil {
			klog.V(8).Infof("ignoring non-field: %s", path)
			return nil
		}
		tag := field.Tag.Get("flag")
		if tag == "" {
			klog.V(4).Infof("not writing field with no flag tag: %s", path)
			// We want to descend - it could be a structure containing flags
			return nil
		}
		if tag == "-" {
			klog.V(4).Infof("skipping field with %q flag tag: %s", tag, path)
			return reflectutils.SkipReflection
		}

		// If we specify the repeat option, we will repeat the flag rather than joining it with commas
		repeatFlag := false

		tokens := strings.Split(tag, ",")
		if len(tokens) > 1 {
			for i, t := range tokens {
				if i == 0 {
					continue
				}
				if t == "repeat" {
					repeatFlag = true
				} else {
					return fmt.Errorf("cannot parse flag spec: %q", tag)
				}
			}
		}
		flagName := tokens[0]

		// If the "unset" value is not empty string, by setting this tag we avoid passing spurious flag values
		flagEmpty := field.Tag.Get("flag-empty")

		flagIncludeEmpty, _ := strconv.ParseBool(field.Tag.Get("flag-include-empty"))

		// We do have to do this, even though the recursive walk will do it for us
		// because when we descend we won't have `field` set
		if val.Kind() == reflect.Ptr && reflect.TypeOf(val.Interface()).String() != "*string" {
			if val.IsNil() {
				return nil
			}
			val = val.Elem()
		}

		if val.Kind() == reflect.Map {
			if val.IsNil() {
				return nil
			}
			// We handle a map[string]string like --node-labels=k1=v1,k2=v2 etc
			// As we need more formats we can add additional spec to the flags tag
			if stringStringMap, ok := val.Interface().(map[string]string); ok {
				var args []string
				for k, v := range stringStringMap {
					arg := fmt.Sprintf("%s=%s", k, v)
					args = append(args, arg)
				}
				sort.Strings(args)
				if len(args) != 0 {
					flag := fmt.Sprintf("--%s=%s", flagName, strings.Join(args, ","))
					flags = append(flags, flag)
				}
				return reflectutils.SkipReflection
			}

			return fmt.Errorf("BuildFlags of value type not handled: %T %s=%v", val.Interface(), path, val.Interface())
		}

		if val.Kind() == reflect.Slice {
			if val.IsNil() {
				return nil
			}
			// We handle a []string like --admission-control=v1,v2 etc
			if stringSlice, ok := val.Interface().([]string); ok {
				if len(stringSlice) != 0 {
					if repeatFlag {
						for _, v := range stringSlice {
							flag := fmt.Sprintf("--%s=%s", flagName, v)
							flags = append(flags, flag)
						}
					} else {
						flag := fmt.Sprintf("--%s=%s", flagName, strings.Join(stringSlice, ","))
						flags = append(flags, flag)
					}
				}
				return reflectutils.SkipReflection
			}

			return fmt.Errorf("BuildFlags of value type not handled: %T %s=%v", val.Interface(), path, val.Interface())
		}

		var flag string
		switch v := val.Interface().(type) {
		case string:
			vString := fmt.Sprintf("%v", v)
			if vString != "" && vString != flagEmpty {
				flag = fmt.Sprintf("--%s=%s", flagName, vString)
			}

		case *string:
			if v != nil {
				// If flagIncludeEmpty is specified, include anything, including empty strings. Otherwise, behave
				// just like the string case above.
				if flagIncludeEmpty {
					vString := fmt.Sprintf("%v", *v)
					flag = fmt.Sprintf("--%s=%s", flagName, vString)
				} else {
					vString := fmt.Sprintf("%v", *v)
					if vString != "" && vString != flagEmpty {
						flag = fmt.Sprintf("--%s=%s", flagName, vString)
					}
				}
			}

		case bool, int, int32, int64:
			vString := fmt.Sprintf("%v", v)
			if vString != flagEmpty {
				flag = fmt.Sprintf("--%s=%s", flagName, vString)
			}

		case float32, float64:
			// Because these types don't round-trip, we should use resource.Quantity instead
			klog.Warningf("use of unsafe float type for flag %q; use resource.Quantity instead", flagName)
			vString := fmt.Sprintf("%v", v)
			if vString != flagEmpty {
				flag = fmt.Sprintf("--%s=%s", flagName, vString)
			}

		case metav1.Duration:
			vString := v.Duration.String()

			// See https://github.com/kubernetes/kubernetes/issues/40783
			// Go renders a time.Duration to `0` in <= 1.6, and `0s` in >= 1.7
			// We force it to be `0s`, regardless of value
			if vString == "0" {
				vString = "0s"
			}

			if vString != flagEmpty {
				flag = fmt.Sprintf("--%s=%s", flagName, vString)
			}

		case resource.Quantity:
			// Format as a floating point value (i.e. 3.14, not 3140m)
			vString := v.AsDec().String()
			if vString != flagEmpty {
				flag = fmt.Sprintf("--%s=%s", flagName, vString)
			}

		default:
			return fmt.Errorf("BuildFlagsList of value type not handled: %T %s=%v", v, path, v)
		}
		if flag != "" {
			flags = append(flags, flag)
		}

		return reflectutils.SkipReflection
	}
	err := reflectutils.ReflectRecursive(reflect.ValueOf(options), walker)
	if err != nil {
		return nil, fmt.Errorf("BuildFlagsList to reflect value: %s", err)
	}
	// Sort so that the order is stable across runs
	sort.Strings(flags)

	return flags, nil
}
