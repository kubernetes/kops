package fi

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
	"sync"
	"time"
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
// It will perform some re-execution on error, retrying as long as progress is still being made
func (e *executor) RunTasks(taskMap map[string]Task, maxAttemptsWithNoProgress int) error {
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

	noProgressCount := 0
	for {
		var canRun []*taskState
		doneCount := 0
		for _, ts := range taskStates {
			if ts.done {
				doneCount++
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

		glog.Infof("Tasks: %d done / %d total; %d can run", doneCount, len(taskStates), len(canRun))
		if len(canRun) == 0 {
			break
		}

		progress := false

		var tasks []*taskState
		for _, ts := range canRun {
			tasks = append(tasks, ts)
		}

		errors := e.forkJoin(tasks)
		for i, err := range errors {
			ts := tasks[i]
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
				noProgressCount++
				if noProgressCount == maxAttemptsWithNoProgress {
					return fmt.Errorf("did not make any progress executing task.  Example error: %v", errors[0])
				} else {
					glog.Infof("No progress made, sleeping before retrying failed tasks")
					time.Sleep(10 * time.Second)
				}
			} else {
				// Logic error!
				panic("did not make progress executing tasks; but no errors reported")
			}
		} else {
			noProgressCount = 0
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

type runnable func() error

func (e *executor) forkJoin(tasks []*taskState) []error {
	if len(tasks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	results := make([]error, len(tasks))
	for i := 0; i < len(tasks); i++ {
		wg.Add(1)
		go func(ts *taskState, index int) {
			results[index] = fmt.Errorf("function panic")
			defer wg.Done()
			glog.V(2).Infof("Executing task %q: %v\n", ts.key, ts.task)
			results[index] = ts.task.Run(e.context)
		}(tasks[i], i)
	}

	wg.Wait()

	return results
}
