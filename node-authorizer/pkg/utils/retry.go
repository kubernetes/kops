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

package utils

import (
	"context"
	"errors"
	"time"

	"github.com/jpillora/backoff"
	"go.uber.org/zap"
)

// Retry attempts to perform an operation for x time
func Retry(ctx context.Context, interval, timeout time.Duration, fn func() error) (err error) {
	j := &backoff.Backoff{Min: interval / 2, Max: interval, Factor: 1, Jitter: true}
	to := time.NewTimer(timeout)
	defer to.Stop()

	for {
		if err = fn(); err == nil {
			return nil
		}

		Logger.Error("operation failed to execute", zap.Error(err))

		// @check if the context has been cancelled or wait
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-to.C:
			return errors.New("operation timed out")
		case <-time.After(j.Duration()):
		}
	}
}
