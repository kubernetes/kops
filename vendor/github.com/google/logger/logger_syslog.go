// +build linux darwin freebsd

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
	"io"
	"log/syslog"
)

func setup(src string) (io.Writer, io.Writer, io.Writer, error) {
	const facility = syslog.LOG_USER
	il, err := syslog.New(facility|syslog.LOG_NOTICE, src)
	if err != nil {
		return nil, nil, nil, err
	}
	wl, err := syslog.New(facility|syslog.LOG_WARNING, src)
	if err != nil {
		return nil, nil, nil, err
	}
	el, err := syslog.New(facility|syslog.LOG_ERR, src)
	if err != nil {
		return nil, nil, nil, err
	}
	return il, wl, el, nil
}
