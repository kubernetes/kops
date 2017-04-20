package net

import (
	"fmt"
	"net"
	"os"
)

func ListenUnixSocket(pathname string) (net.Listener, error) {
	if err := os.Remove(pathname); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("Error removing existing unix socket: %s", err)
	}
	listener, err := net.Listen("unix", pathname)
	if err != nil {
		return nil, fmt.Errorf("ListenUnixSocket failed: %s", err)
	}
	return listener, nil
}
