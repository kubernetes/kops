/*
Copyright 2016 Google Inc. All Rights Reserved.
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

package logger

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/eventlog"
)

type writer struct {
	pri severity
	src string
	el  *eventlog.Log
}

// Write sends a log message to the Event Log.
func (w *writer) Write(b []byte) (int, error) {
	switch w.pri {
	case sInfo:
		return len(b), w.el.Info(1, string(b))
	case sWarning:
		return len(b), w.el.Warning(3, string(b))
	case sError:
		return len(b), w.el.Error(2, string(b))
	}
	return 0, fmt.Errorf("unrecognized severity: %v", w.pri)
}

func (w *writer) Close() error {
	return w.el.Close()
}

func newW(pri severity, src string) (io.WriteCloser, error) {
	// Continue if we receive "registry key already exists" or if we get
	// ERROR_ACCESS_DENIED so that we can log without administrative permissions
	// for pre-existing eventlog sources.
	if err := eventlog.InstallAsEventCreate(src, eventlog.Info|eventlog.Warning|eventlog.Error); err != nil {
		if !strings.Contains(err.Error(), "registry key already exists") && err != windows.ERROR_ACCESS_DENIED {
			return nil, err
		}
	}
	el, err := eventlog.Open(src)
	if err != nil {
		return nil, err
	}
	return &writer{
		pri: pri,
		src: src,
		el:  el,
	}, nil
}

func setup(src string) (io.WriteCloser, io.WriteCloser, io.WriteCloser, error) {
	infoL, err := newW(sInfo, src)
	if err != nil {
		return nil, nil, nil, err
	}
	warningL, err := newW(sWarning, src)
	if err != nil {
		return nil, nil, nil, err
	}
	errL, err := newW(sError, src)
	if err != nil {
		return nil, nil, nil, err
	}
	return infoL, warningL, errL, nil
}
