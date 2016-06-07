package fi

import "time"

func BuildTimestampString() string {
	now := time.Now()
	return now.UTC().Format("20060102150405")
}
