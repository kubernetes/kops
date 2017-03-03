//
// Copyright (c) 2015 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/heketi/tests"
	"strings"
	"testing"
)

func TestLogLevel(t *testing.T) {
	var testbuffer bytes.Buffer

	defer tests.Patch(&stdout, &testbuffer).Restore()

	l := NewLogger("[testing]", LEVEL_INFO)
	tests.Assert(t, LEVEL_INFO == l.level)
	tests.Assert(t, LEVEL_INFO == l.Level())

	l.SetLevel(LEVEL_CRITICAL)
	tests.Assert(t, LEVEL_CRITICAL == l.level)
	tests.Assert(t, LEVEL_CRITICAL == l.Level())

}

func TestLogInfo(t *testing.T) {
	var testbuffer bytes.Buffer

	defer tests.Patch(&stdout, &testbuffer).Restore()

	l := NewLogger("[testing]", LEVEL_INFO)

	l.Info("Hello %v", "World")
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] INFO "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "Hello World"), testbuffer.String())
	testbuffer.Reset()

	l.SetLevel(LEVEL_WARNING)
	l.Info("TEXT")
	tests.Assert(t, testbuffer.Len() == 0)
}

func TestLogDebug(t *testing.T) {
	var testbuffer bytes.Buffer

	defer tests.Patch(&stdout, &testbuffer).Restore()

	l := NewLogger("[testing]", LEVEL_DEBUG)

	l.Debug("Hello %v", "World")
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] DEBUG "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "Hello World"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "log_test.go"), testbuffer.String())

	// [testing] DEBUG 2016/04/28 15:25:08 /src/github.com/heketi/heketi/pkg/utils/log_test.go:66: Hello World
	fileinfo := strings.Split(testbuffer.String(), " ")[4]
	filename := strings.Split(fileinfo, ":")[0]

	// Need to check that it starts with /src/github.com
	tests.Assert(t, strings.HasPrefix(filename, "/src/github.com/"))
	tests.Assert(t, strings.HasSuffix(filename, "/pkg/utils/log_test.go"))
	testbuffer.Reset()

	l.SetLevel(LEVEL_INFO)
	l.Debug("TEXT")
	tests.Assert(t, testbuffer.Len() == 0)
}

func TestLogWarning(t *testing.T) {
	var testbuffer bytes.Buffer

	defer tests.Patch(&stdout, &testbuffer).Restore()

	l := NewLogger("[testing]", LEVEL_DEBUG)

	l.Warning("Hello %v", "World")
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] WARNING "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "Hello World"), testbuffer.String())
	testbuffer.Reset()

	l.SetLevel(LEVEL_ERROR)
	l.Warning("TEXT")
	tests.Assert(t, testbuffer.Len() == 0)
}

func TestLogWarnErr(t *testing.T) {
	var testbuffer bytes.Buffer

	defer tests.Patch(&stdout, &testbuffer).Restore()

	l := NewLogger("[testing]", LEVEL_DEBUG)

	ErrSample := errors.New("TEST ERROR")
	err := l.WarnErr(ErrSample)
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] WARNING "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "TEST ERROR"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "log_test.go"), testbuffer.String())
	tests.Assert(t, err == ErrSample)
	testbuffer.Reset()

	err = l.WarnErr(fmt.Errorf("GOT %v", err))
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] WARNING "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "TEST ERROR"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "log_test.go"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "GOT"), testbuffer.String())
	tests.Assert(t, err != ErrSample)
	tests.Assert(t, err != nil)
	tests.Assert(t, strings.Contains(err.Error(), "GOT TEST ERROR"), err)
	testbuffer.Reset()

	l.SetLevel(LEVEL_ERROR)
	l.WarnErr(ErrSample)
	tests.Assert(t, testbuffer.Len() == 0)
}

func TestLogError(t *testing.T) {
	var testbuffer bytes.Buffer

	defer tests.Patch(&stderr, &testbuffer).Restore()

	l := NewLogger("[testing]", LEVEL_DEBUG)

	err := l.LogError("Hello %v", "World")
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] ERROR "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "Hello World"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "log_test.go"), testbuffer.String())
	testbuffer.Reset()
	testbuffer.Reset()
	tests.Assert(t, err != nil)
	tests.Assert(t, strings.Contains(err.Error(), "Hello World"), err)

	err = errors.New("BAD")
	l.Err(err)
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] ERROR "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "BAD"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "log_test.go"), testbuffer.String())
	testbuffer.Reset()

	l.SetLevel(LEVEL_CRITICAL)
	l.LogError("TEXT")
	tests.Assert(t, testbuffer.Len() == 0)

}

func TestLogCritical(t *testing.T) {
	var testbuffer bytes.Buffer

	defer tests.Patch(&stderr, &testbuffer).Restore()

	l := NewLogger("[testing]", LEVEL_DEBUG)

	l.LogError("Hello %v", "World")
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] ERROR "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "Hello World"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "log_test.go"), testbuffer.String())
	testbuffer.Reset()

	l.SetLevel(LEVEL_NOLOG)
	l.LogError("TEXT")
	tests.Assert(t, testbuffer.Len() == 0)

}

func TestLogErr(t *testing.T) {
	var testbuffer bytes.Buffer

	defer tests.Patch(&stderr, &testbuffer).Restore()

	l := NewLogger("[testing]", LEVEL_DEBUG)

	ErrSample := errors.New("TEST ERROR")
	err := l.Err(ErrSample)
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] ERROR "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "TEST ERROR"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "log_test.go"), testbuffer.String())
	tests.Assert(t, err == ErrSample)
	testbuffer.Reset()

	err = l.Err(fmt.Errorf("GOT %v", err))
	tests.Assert(t, strings.Contains(testbuffer.String(), "[testing] ERROR "), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "TEST ERROR"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "log_test.go"), testbuffer.String())
	tests.Assert(t, strings.Contains(testbuffer.String(), "GOT"), testbuffer.String())
	tests.Assert(t, err != ErrSample)
	tests.Assert(t, err != nil)
	tests.Assert(t, strings.Contains(err.Error(), "GOT TEST ERROR"), err)
	testbuffer.Reset()

	l.SetLevel(LEVEL_NOLOG)
	l.Err(ErrSample)
	tests.Assert(t, testbuffer.Len() == 0)
}
