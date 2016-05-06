package fi

//
//import (
//	"fmt"
//	"github.com/golang/glog"
//	"sort"
//	"strings"
//)
//
//type Options map[string]interface{}
//
//func NewOptions() Options {
//	m := make(map[string]interface{})
//	return Options(m)
//}
//
//func (o Options) Merge(r Options) error {
//	return merge(o, r)
//}
//
//// TODO: What do we do about this...?
//func (o Options) Token(key string) string {
//	return "secret-" + key
//}
//
//func merge(l, r map[string]interface{}) error {
//	for k, v := range r {
//		if v == nil {
//			delete(l, k)
//			continue
//		}
//
//		switch v := v.(type) {
//		case string, int, bool:
//			l[k] = v
//
//		case map[string]interface{}:
//			existing, found := l[k]
//			if !found {
//				l[k] = v
//			} else {
//				switch existing := existing.(type) {
//				case map[string]interface{}:
//					err := merge(existing, v)
//					if err != nil {
//						return err
//					}
//
//				default:
//					return fmt.Errorf("cannot merge object into target of type %T", v)
//
//				}
//			}
//
//		default:
//			return fmt.Errorf("merging of option type not handled: %T", v)
//		}
//	}
//	return nil
//}
//
//func (o Options) BuildFlags(path string) string {
//	if path != "" {
//		options := o.Navigate(path)
//		return options.BuildFlags("")
//	}
//
//	var flags []string
//	for k, v := range o {
//		var flag string
//		switch v := v.(type) {
//		case string, int, bool, float32, float64:
//			flag = fmt.Sprintf("--%s=%v", k, v)
//
//		default:
//			// TODO: Better error handling (with templates)
//			glog.Exitf("BuildFlags of value type not handled: %T %s=%v", v, k, v)
//			return ""
//		}
//		if flag != "" {
//			flags = append(flags, flag)
//		}
//	}
//	sort.Strings(flags)
//
//	return strings.Join(flags, " ")
//}
//
//func (o Options) Navigate(path string) Options {
//	if path == "" {
//		return o
//	}
//
//	tokens := strings.SplitN(path, ".", 2)
//
//	child, found := o[tokens[0]]
//	if !found {
//		return NewOptions()
//	}
//
//	var childOptions Options
//	switch child := child.(type) {
//
//	case map[string]interface{}:
//		childOptions = Options(child)
//
//	default:
//		glog.Warningf("Navigate of chjild type not handled: %T", child)
//		childOptions = NewOptions()
//	}
//
//	if len(tokens) == 1 {
//		return childOptions
//	}
//	return childOptions.Navigate(tokens[1])
//}
