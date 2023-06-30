//go:build !windows

// Copyright (c) 2018, Google LLC All rights reserved.
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

package tpmutil

import (
	"fmt"
	"io"
	"net"
	"os"
)

// OpenTPM opens a channel to the TPM at the given path. If the file is a
// device, then it treats it like a normal TPM device, and if the file is a
// Unix domain socket, then it opens a connection to the socket.
func OpenTPM(path string) (io.ReadWriteCloser, error) {
	// If it's a regular file, then open it
	var rwc io.ReadWriteCloser
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.Mode()&os.ModeDevice != 0 {
		var f *os.File
		f, err = os.OpenFile(path, os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
		rwc = io.ReadWriteCloser(f)
	} else if fi.Mode()&os.ModeSocket != 0 {
		rwc = NewEmulatorReadWriteCloser(path)
	} else {
		return nil, fmt.Errorf("unsupported TPM file mode %s", fi.Mode().String())
	}

	return rwc, nil
}

// dialer abstracts the net.Dial call so test code can provide its own net.Conn
// implementation.
type dialer func(network, path string) (net.Conn, error)

// EmulatorReadWriteCloser manages connections with a TPM emulator over a Unix
// domain socket. These emulators often operate in a write/read/disconnect
// sequence, so the Write method always connects, and the Read method always
// closes. EmulatorReadWriteCloser is not thread safe.
type EmulatorReadWriteCloser struct {
	path   string
	conn   net.Conn
	dialer dialer
}

// NewEmulatorReadWriteCloser stores information about a Unix domain socket to
// write to and read from.
func NewEmulatorReadWriteCloser(path string) *EmulatorReadWriteCloser {
	return &EmulatorReadWriteCloser{
		path:   path,
		dialer: net.Dial,
	}
}

// Read implements io.Reader by reading from the Unix domain socket and closing
// it.
func (erw *EmulatorReadWriteCloser) Read(p []byte) (int, error) {
	// Read is always the second operation in a Write/Read sequence.
	if erw.conn == nil {
		return 0, fmt.Errorf("must call Write then Read in an alternating sequence")
	}
	n, err := erw.conn.Read(p)
	erw.conn.Close()
	erw.conn = nil
	return n, err
}

// Write implements io.Writer by connecting to the Unix domain socket and
// writing.
func (erw *EmulatorReadWriteCloser) Write(p []byte) (int, error) {
	if erw.conn != nil {
		return 0, fmt.Errorf("must call Write then Read in an alternating sequence")
	}
	var err error
	erw.conn, err = erw.dialer("unix", erw.path)
	if err != nil {
		return 0, err
	}
	return erw.conn.Write(p)
}

// Close implements io.Closer by closing the Unix domain socket if one is open.
func (erw *EmulatorReadWriteCloser) Close() error {
	if erw.conn == nil {
		return fmt.Errorf("cannot call Close when no connection is open")
	}
	err := erw.conn.Close()
	erw.conn = nil
	return err
}
