package nodeup

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"reflect"
	"sort"
	"strings"
)

func buildFlags(options interface{}) (string, error) {
	var flags []string

	walker := func(path string, field *reflect.StructField, val reflect.Value) error {
		if field == nil {
			glog.V(4).Infof("not writing non-field: %s", path)
			return nil
		}
		tag := field.Tag.Get("flag")
		if tag == "" {
			glog.V(4).Infof("not writing field with no flag tag: %s", path)
			return nil
		}
		if tag == "-" {
			glog.V(4).Infof("skipping field with %q flag tag: %s", tag, path)
			return utils.SkipReflection
		}
		flagName := tag

		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				return nil
			}
			val = val.Elem()
		}

		var flag string
		switch v := val.Interface().(type) {
		case string, int, bool, float32, float64:
			vString := fmt.Sprintf("%v", v)
			if vString != "" {
				flag = fmt.Sprintf("--%s=%s", flagName, vString)
			}

		default:
			return fmt.Errorf("BuildFlags of value type not handled: %T %s=%v", v, path, v)
		}
		if flag != "" {
			flags = append(flags, flag)
		}
		return nil
	}
	err := utils.WalkRecursive(reflect.ValueOf(options), walker)
	if err != nil {
		return "", err
	}
	// Sort so that the order is stable across runs
	sort.Strings(flags)

	return strings.Join(flags, " "), nil
}
