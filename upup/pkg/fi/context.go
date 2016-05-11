package fi

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

type Context struct {
	Tmpdir string

	Target  Target
	Cloud   Cloud
	CAStore CAStore

	CheckExisting bool
}

func NewContext(target Target, cloud Cloud, castore CAStore, checkExisting bool) (*Context, error) {
	c := &Context{
		Cloud:         cloud,
		Target:        target,
		CAStore:       castore,
		CheckExisting: checkExisting,
	}

	t, err := ioutil.TempDir("", "deploy")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary directory: %v", err)
	}
	c.Tmpdir = t

	return c, nil
}

func (c *Context) RunTasks(taskMap map[string]Task) error {
	taskOrder := TopologicalSort(taskMap)

	for _, stage := range taskOrder {
		for _, k := range stage {
			task := taskMap[k]
			glog.V(2).Infof("Executing task: %v\n", task)
			err := task.Run(c)
			if err != nil {
				return fmt.Errorf("error running tasks (%s): %v", task, err)
			}
		}
	}

	return nil
}

func (c *Context) Close() {
	glog.V(2).Infof("deleting temp dir: %q", c.Tmpdir)
	if c.Tmpdir != "" {
		err := os.RemoveAll(c.Tmpdir)
		if err != nil {
			glog.Warningf("unable to delete temporary directory %q: %v", c.Tmpdir, err)
		}
	}
}

//func (c *Context) MergeOptions(options Options) error {
//	return c.Options.Merge(options)
//}

func (c *Context) NewTempDir(prefix string) (string, error) {
	t, err := ioutil.TempDir(c.Tmpdir, prefix)
	if err != nil {
		return "", fmt.Errorf("error creating temporary directory: %v", err)
	}
	return t, nil
}

func (c *Context) Render(a, e, changes Task) error {
	if _, ok := c.Target.(*DryRunTarget); ok {
		return c.Target.(*DryRunTarget).Render(a, e, changes)
	}

	v := reflect.ValueOf(e)
	vType := v.Type()

	targetType := reflect.ValueOf(c.Target).Type()

	var renderer *reflect.Method
	for i := 0; i < vType.NumMethod(); i++ {
		method := vType.Method(i)
		if !strings.HasPrefix(method.Name, "Render") {
			continue
		}
		match := true
		for j := 0; j < method.Type.NumIn(); j++ {
			arg := method.Type.In(j)
			if arg.ConvertibleTo(vType) {
				continue
			}
			if arg.ConvertibleTo(targetType) {
				continue
			}
			match = false
			break
		}
		if match {
			if renderer != nil {
				return fmt.Errorf("Found multiple Render methods that could be invokved on %T", e)
			}
			renderer = &method
		}

	}
	if renderer == nil {
		return fmt.Errorf("Could not find Render method on type %T (target %T)", e, c.Target)
	}
	var args []reflect.Value
	args = append(args, reflect.ValueOf(c.Target))
	args = append(args, reflect.ValueOf(a))
	args = append(args, reflect.ValueOf(e))
	args = append(args, reflect.ValueOf(changes))
	glog.V(4).Infof("Calling method %s on %T", renderer.Name, e)
	m := v.MethodByName(renderer.Name)
	rv := m.Call(args)
	var rvErr error
	if !rv[0].IsNil() {
		rvErr = rv[0].Interface().(error)
	}
	return rvErr
}
