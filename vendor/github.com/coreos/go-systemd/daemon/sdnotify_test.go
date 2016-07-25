// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package daemon

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
)

// TestSdNotify
func TestSdNotify(t *testing.T) {
	notificationSupportedDataSent := "Notification supported, data sent"
	notificationSupportedFailure := "Notification supported, but failure happened"
	notificationNotSupported := "Notification not supported"

	testDir, e := ioutil.TempDir("/tmp/", "test-")
	if e != nil {
		panic(e)
	}
	defer os.RemoveAll(testDir)

	notifySocket := testDir + "/notify-socket.sock"
	laddr := net.UnixAddr{
		Name: notifySocket,
		Net:  "unixgram",
	}
	_, e = net.ListenUnixgram("unixgram", &laddr)
	if e != nil {
		panic(e)
	}

	// (true, nil) - notification supported, data has been sent
	e = os.Setenv("NOTIFY_SOCKET", notifySocket)
	if e != nil {
		panic(e)
	}
	sent, err := SdNotify(notificationSupportedDataSent)
	if !sent || err != nil {
		t.Errorf("TEST: %s FAILED", notificationSupportedDataSent)
	}

	// (false, err) - notification supported, but failure happened
	e = os.Setenv("NOTIFY_SOCKET", testDir+"/not-exist.sock")
	if e != nil {
		panic(e)
	}
	sent, err = SdNotify(notificationSupportedFailure)
	if sent && err == nil {
		t.Errorf("TEST: %s FAILED", notificationSupportedFailure)
	}

	// (false, nil) - notification not supported
	e = os.Unsetenv("NOTIFY_SOCKET")
	if e != nil {
		panic(e)
	}
	sent, err = SdNotify(notificationNotSupported)
	if sent || err != nil {
		t.Errorf("TEST: %s FAILED", notificationNotSupported)
	}
}
