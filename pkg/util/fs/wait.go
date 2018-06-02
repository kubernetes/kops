/*
Copyright 2018 The Kubernetes Authors.

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

package fs

import (
	"context"
	"errors"
	"os"
	"time"
)

// ErrTimeout indicates the operation has timed out
var ErrTimeout = errors.New("operation timeout")

// WaitForFile is responsible for waiting for file to appear or timeout
func WaitForFile(ctx context.Context, path string) error {
	doneCh := make(chan struct{}, 0)

	// @step: we wait for the bootstrap token file to appear
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if found, _ := FileExists(path); found {
					doneCh <- struct{}{}
					return
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
