package hcloud

import (
	"testing"
	"time"
)

const apiTimestampFormat = "2006-01-02T15:04:05-07:00"

func mustParseTime(t *testing.T, layout, value string) time.Time {
	t.Helper()

	ts, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("parse time: layout %v: value %v: %v", layout, value, err)
	}
	return ts
}
