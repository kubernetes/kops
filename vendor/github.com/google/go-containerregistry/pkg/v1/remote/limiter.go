// Copyright 2026 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"context"
	"io"
	"sync"
)

type pullLimiter struct {
	tokens chan struct{}
}

func newPullLimiter(jobs int) *pullLimiter {
	return &pullLimiter{
		tokens: make(chan struct{}, jobs),
	}
}

func (l *pullLimiter) acquire(ctx context.Context) (func(), error) {
	if l == nil {
		return func() {}, nil
	}
	select {
	case l.tokens <- struct{}{}:
		return func() { <-l.tokens }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type limitedReadCloser struct {
	io.ReadCloser
	release func()
	once    sync.Once
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
	n, err := l.ReadCloser.Read(p)
	if err != nil {
		l.once.Do(l.release)
	}
	return n, err
}

func (l *limitedReadCloser) Close() error {
	err := l.ReadCloser.Close()
	l.once.Do(l.release)
	return err
}
