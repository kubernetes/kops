package utils

import (
	"bytes"
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
