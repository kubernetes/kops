package main

import (
	"strings"
	"os"
)

// expandPath replaces common path aliases: ~ -> $HOME
func expandPath(p string) (string) {
	if strings.HasPrefix(p, "~/") {
		p = os.Getenv("HOME") + p[1:]
	}
	return p
}