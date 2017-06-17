package kubemanifest

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
)

type visitorBase struct {
}

func (m *visitorBase) VisitString(path []string, v string, mutator func(string)) error {
	glog.Infof("string value at %s: %s", strings.Join(path, "."), v)
	return nil
}

func (m *visitorBase) VisitBool(path []string, v bool, mutator func(bool)) error {
	glog.Infof("string value at %s: %s", strings.Join(path, "."), v)
	return nil
}

func (m *visitorBase) VisitFloat64(path []string, v float64, mutator func(float64)) error {
	glog.Infof("float64 value at %s: %s", strings.Join(path, "."), v)
	return nil
}

type Visitor interface {
	VisitBool(path []string, v bool, mutator func(bool)) error
	VisitString(path []string, v string, mutator func(string)) error
	VisitFloat64(path []string, v float64, mutator func(float64)) error
}

func visit(visitor Visitor, data interface{}, path []string, mutator func(interface{})) error {
	switch data.(type) {
	case string:
		err := visitor.VisitString(path, data.(string), func(v string) {
			mutator(v)
		})
		if err != nil {
			return err
		}

	case bool:
		err := visitor.VisitBool(path, data.(bool), func(v bool) {
			mutator(v)
		})
		if err != nil {
			return err
		}

	case float64:
		err := visitor.VisitFloat64(path, data.(float64), func(v float64) {
			mutator(v)
		})
		if err != nil {
			return err
		}

	case map[string]interface{}:
		m := data.(map[string]interface{})
		for k, v := range m {
			path = append(path, k)

			err := visit(visitor, v, path, func(v interface{}) {
				m[k] = v
			})
			if err != nil {
				return err
			}
			path = path[:len(path)-1]
		}

	case []interface{}:
		s := data.([]interface{})
		for i, v := range s {
			path = append(path, fmt.Sprintf("[%d]", i))

			err := visit(visitor, v, path, func(v interface{}) {
				s[i] = v
			})
			if err != nil {
				return err
			}
			path = path[:len(path)-1]
		}

	default:
		return fmt.Errorf("unhandled type in manifest: %T", data)
	}

	return nil
}
