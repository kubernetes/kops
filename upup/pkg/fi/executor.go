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

package fi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

type executor[T SubContext] struct {
	context *Context[T]

	options RunTasksOptions
}

type taskState[T SubContext] struct {
	done         bool
	key          string
	task         Task[T]
	deadline     time.Time
	lastError    error
	dependencies []*taskState[T]
}

type RunTasksOptions struct {
	MaxTaskDuration         time.Duration
	WaitAfterAllTasksFailed time.Duration
}

func (o *RunTasksOptions) InitDefaults() {
	o.MaxTaskDuration = 10 * time.Minute
	o.WaitAfterAllTasksFailed = 10 * time.Second
}

// RunTasks executes all the tasks, considering their dependencies
// It will perform some re-execution on error, retrying as long as progress is still being made
func (e *executor[T]) RunTasks(ctx context.Context, taskMap map[string]Task[T]) error {
	dependencies := FindTaskDependencies(taskMap)

	for _, task := range taskMap {
		if taskPreRun, ok := task.(TaskPreRun[T]); ok {
			if err := taskPreRun.PreRun(e.context); err != nil {
				return err
			}
		}
	}

	taskStates := make(map[string]*taskState[T])

	for k, task := range taskMap {
		ts := &taskState[T]{
			key:  k,
			task: task,
		}
		taskStates[k] = ts
	}

	for k, ts := range taskStates {
		for _, dep := range dependencies[k] {
			d := taskStates[dep]
			if d == nil {
				klog.Fatalf("did not find task state for dependency: %q", k)
			}
			ts.dependencies = append(ts.dependencies, d)
		}
	}

	for {
		var canRun []*taskState[T]
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
				if ts.deadline.IsZero() {
					ts.deadline = time.Now().Add(e.options.MaxTaskDuration)
				} else if time.Now().After(ts.deadline) {
					return fmt.Errorf("deadline exceeded executing task %v. Example error: %v", ts.key, ts.lastError)
				}
				canRun = append(canRun, ts)
			}
		}

		klog.Infof("Tasks: %d done / %d total; %d can run", doneCount, len(taskStates), len(canRun))
		if len(canRun) == 0 {
			break
		}

		progress := false

		var tasks []*taskState[T]
		tasks = append(tasks, canRun...)

		taskErrors := e.forkJoin(ctx, tasks)
		var errs []error
		for i, err := range taskErrors {
			ts := tasks[i]
			if err != nil {
				//  print warning message and continue like the task succeeded
				if _, ok := err.(*ExistsAndWarnIfChangesError); ok {
					klog.Warning(err.Error())
					ts.done = true
					ts.lastError = nil
					progress = true
					continue
				}

				remaining := time.Second * time.Duration(int(time.Until(ts.deadline).Seconds()))
				if _, ok := err.(*TryAgainLaterError); ok {
					klog.V(2).Infof("Task %q not ready: %v", ts.key, err)
				} else {
					klog.Warningf("error running task %q (%v remaining to succeed): %v", ts.key, remaining, err)
				}
				errs = append(errs, err)
				ts.lastError = err
			} else {
				ts.done = true
				ts.lastError = nil
				progress = true
			}
		}

		if !progress {
			n := len(errs)

			if n == 0 {
				// Logic error!
				panic("did not make progress executing tasks; but no errors reported")
			}

			tryAgainLaterCount := 0
			for _, err := range errs {
				var tryAgainLaterError TryAgainLaterError
				if !errors.Is(err, &tryAgainLaterError) {
					tryAgainLaterCount++
				}
			}
			formatTaskCount := func(n int) string {
				return fmt.Sprintf("%d task(s)", n)
			}
			if tryAgainLaterCount == n {
				klog.Infof("Continuing to run %s", formatTaskCount(tryAgainLaterCount))
			} else {
				klog.Infof("No progress made, sleeping before retrying %s", formatTaskCount(n))
			}
			time.Sleep(e.options.WaitAfterAllTasksFailed)
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

func (e *executor[T]) forkJoin(ctx context.Context, tasks []*taskState[T]) []error {
	if len(tasks) == 0 {
		return nil
	}

	results := make([]error, len(tasks))
	var resultsMutex sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < len(tasks); i++ {
		wg.Add(1)
		go func(ts *taskState[T], index int) {
			defer wg.Done()

			_, span := tracer.Start(ctx, "task-"+ts.key)
			defer span.End()

			resultsMutex.Lock()
			results[index] = fmt.Errorf("function panic")
			resultsMutex.Unlock()

			klog.V(2).Infof("Executing task %q: %v\n", ts.key, ts.task)

			if taskNormalize, ok := ts.task.(TaskNormalize[T]); ok {
				if err := taskNormalize.Normalize(e.context); err != nil {
					results[index] = err
					return
				}
			}

			result := ts.task.Run(e.context)

			resultsMutex.Lock()
			results[index] = result
			resultsMutex.Unlock()
		}(tasks[i], i)
	}

	wg.Wait()

	return results
}
