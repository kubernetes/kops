package utils

import (
	"bytes"
	"os"
	"strings"
)

func SanitizeString(s string) string {
	var out bytes.Buffer
	allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	for _, c := range s {
		if strings.IndexRune(allowed, c) != -1 {
			out.WriteRune(c)
		} else {
			out.WriteRune('_')
		}
	}
	return string(out.Bytes())
}

// ExpandPath replaces common path aliases: ~ -> $HOME
func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		p = os.Getenv("HOME") + p[1:]
	}
	return p
}
