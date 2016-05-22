package fi

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
)

type executor struct {
	context *Context
}

type taskState struct {
	done         bool
	key          string
	task         Task
	dependencies []*taskState
}

// RunTasks executes all the tasks, considering their dependencies
// It will perform some re-execution on error, retrying as long as progess is still being made
func (e *executor) RunTasks(taskMap map[string]Task) error {
	dependencies := FindTaskDependencies(taskMap)

	taskStates := make(map[string]*taskState)

	for k, task := range taskMap {
		ts := &taskState{
			key:  k,
			task: task,
		}
		taskStates[k] = ts
	}

	for k, ts := range taskStates {
		for _, dep := range dependencies[k] {
			d := taskStates[dep]
			if d == nil {
				glog.Fatalf("did not find task state for dependency: %q", k)
			}
			ts.dependencies = append(ts.dependencies, d)
		}
	}

	for {
		var canRun []*taskState
		for _, ts := range taskStates {
			if ts.done {
				continue
			}
			ready := true
			for _, dep := range ts.dependencies {
				if !dep.done {
					ready = false
					break
				}
			}
			if ready {
				canRun = append(canRun, ts)
			}
		}

		if len(canRun) == 0 {
			break
		}

		progress := false
		var errors []error

		// TODO: Fork/join execution here
		for _, ts := range canRun {
			glog.V(2).Infof("Executing task %q: %v\n", ts.key, ts.task)
			err := ts.task.Run(e.context)
			if err != nil {
				glog.Warningf("error running task %q: %v", ts.key, err)
				errors = append(errors, err)
			} else {
				ts.done = true
				progress = true
			}
		}

		if !progress {
			if len(errors) != 0 {
				// TODO: Sleep and re-attempt?
				return fmt.Errorf("did not make any progress executing task.  Example error: %v", errors[0])
			} else {
				// Logic error!
				panic("did not make progress executing tasks; but no errors reported")
			}
		}
	}

	// Raise error if not all tasks done - this means they depended on each other
	var notDone []string
	for _, ts := range taskStates {
		if !ts.done {
			notDone = append(notDone, ts.key)
		}
	}
	if len(notDone) != 0 {
		return fmt.Errorf("Unable to execute tasks (circular dependency): %s", strings.Join(notDone, ", "))
	}

	return nil
}
