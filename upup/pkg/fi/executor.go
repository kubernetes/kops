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
	"fmt"
	"os"
	"strconv"
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
func (e *executor[T]) RunTasks(taskMap map[string]Task[T]) error {
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

		taskErrors := e.forkJoin(tasks)
		var errors []error
		for i, err := range taskErrors {
			ts := tasks[i]
			if err != nil {
				//  print warning message and continue like the task succeeded
				if _, ok := err.(*ExistsAndWarnIfChangesError); ok {
					klog.Warningf(err.Error())
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
				errors = append(errors, err)
				ts.lastError = err
			} else {
				ts.done = true
				ts.lastError = nil
				progress = true
			}
		}

		if !progress {
			if len(errors) == 0 {
				// Logic error!
				panic("did not make progress executing tasks; but no errors reported")
			}
			klog.Infof("No progress made, sleeping before retrying %d task(s)", len(errors))
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

// getExecutorPoolSize determines the executor pool size for executing tasks in parallel
//
// It will determine the pool size based on the current number of tasks and the maximum pool size.
// If the maximum pool size is not specified or greater than the actual number of tasks, then the function returns the number of tasks as the pool size.
// Otherwise it will return the defined maximum pool size.
//
// Currently, the max pool size can be set by specifying the ENV variable "KOPS_EXECUTOR_POOL_MAX_SIZE".
func (e *executor[T]) getExecutorPoolSize(numberOfTasks int) int {
	// TODO: make configurable via flag rather than env var?
	executorPoolMaxSize, _ := strconv.Atoi(os.Getenv("KOPS_EXECUTOR_POOL_MAX_SIZE"))

	// if executorPoolMaxSize not specified, or the actual number of tasks being less than the wanted number of workers,
	// set the number of workers to the number of tasks, thus also - more or less - preserving the old behavior
	if executorPoolMaxSize <= 0 || executorPoolMaxSize > numberOfTasks {
		return numberOfTasks
	}

	return executorPoolMaxSize
}

func (e *executor[T]) forkJoin(tasks []*taskState[T]) []error {
	if len(tasks) == 0 {
		return nil
	}

	// a worker based execution of the tasks, based on my findings from here:
	//   * https://stackoverflow.com/questions/55203251/limiting-number-of-go-routines-running
	//   * https://golangbot.com/buffered-channels-worker-pools/

	executorPoolSize := e.getExecutorPoolSize(len(tasks))
	// TODO: set to V(>0)
	klog.V(0).Infof("Executor pool size: %d\n", executorPoolSize)

	// the reason why using the indices rather than the tasks themselves is that the resulting error slice has to match
	// the order of the tasks slice
	// TODO: write a test for that (for forkJoin before making those changes, and then see if everything is still the same)
	taskIndices := make(chan int)
	// feed the workers with the indices of the task slice
	go func() {
		for i := 0; i < len(tasks); i++ {
			taskIndices <- i
		}
		// workers will exit from range loop when channel is closed
		close(taskIndices)
	}()

	results := make([]error, len(tasks))

	var resultsMutex sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < executorPoolSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for index := range taskIndices {
				ts := tasks[index]

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
			}
		}()
	}

	wg.Wait()

	return results
}
