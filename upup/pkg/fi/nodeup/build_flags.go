package nodeup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/utils"
	"reflect"
	"sort"
	"strings"
)

// buildFlags is a template helper, which builds a string containing the flags to be passed to a command
func buildFlags(options interface{}) (string, error) {
	var flags []string

	walker := func(path string, field *reflect.StructField, val reflect.Value) error {
		if field == nil {
			glog.V(8).Infof("ignoring non-field: %s", path)
			return nil
		}
		tag := field.Tag.Get("flag")
		if tag == "" {
			glog.V(4).Infof("not writing field with no flag tag: %s", path)
			// We want to descend - it could be a structure containing flags
			return nil
		}
		if tag == "-" {
			glog.V(4).Infof("skipping field with %q flag tag: %s", tag, path)
			return utils.SkipReflection
		}
		flagName := tag

		// We do have to do this, even though the recursive walk will do it for us
		// because when we descend we won't have `field` set
		if val.Kind() == reflect.Ptr {
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
				if len(args) != 0 {
					flag := fmt.Sprintf("--%s=%s", flagName, strings.Join(args, ","))
					flags = append(flags, flag)
				}
				return utils.SkipReflection
			} else {
				return fmt.Errorf("BuildFlags of value type not handled: %T %s=%v", val.Interface(), path, val.Interface())
			}
		}

		var flag string
		switch v := val.Interface().(type) {
		case string:
			vString := fmt.Sprintf("%v", v)
			if vString != "" {
				flag = fmt.Sprintf("--%s=%s", flagName, vString)
			}

		case bool, int, float32, float64:
			vString := fmt.Sprintf("%v", v)
			flag = fmt.Sprintf("--%s=%s", flagName, vString)

		default:
			return fmt.Errorf("BuildFlags of value type not handled: %T %s=%v", v, path, v)
		}
		if flag != "" {
			flags = append(flags, flag)
		}
		// Nothing more to do here
		return utils.SkipReflection
	}
	err := utils.ReflectRecursive(reflect.ValueOf(options), walker)
	if err != nil {
		return "", err
	}
	// Sort so that the order is stable across runs
	sort.Strings(flags)

	return strings.Join(flags, " "), nil
}
