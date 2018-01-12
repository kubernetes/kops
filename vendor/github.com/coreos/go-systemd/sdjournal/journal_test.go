// Copyright 2015 RedHat, Inc.
// Copyright 2015 CoreOS, Inc.
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

package sdjournal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/coreos/go-systemd/journal"
)

func TestJournalFollow(t *testing.T) {
	r, err := NewJournalReader(JournalReaderConfig{
		Since: time.Duration(-15) * time.Second,
		Matches: []Match{
			{
				Field: SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: "NetworkManager.service",
			},
		},
	})

	if err != nil {
		t.Fatalf("Error opening journal: %s", err)
	}

	if r == nil {
		t.Fatal("Got a nil reader")
	}

	defer r.Close()

	// start writing some test entries
	done := make(chan struct{}, 1)
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				if err = journal.Print(journal.PriInfo, "test message %s", time.Now()); err != nil {
					t.Fatalf("Error writing to journal: %s", err)
				}

				time.Sleep(time.Second)
			}
		}
	}()

	// and follow the reader synchronously
	timeout := time.Duration(5) * time.Second
	if err = r.Follow(time.After(timeout), os.Stdout); err != ErrExpired {
		t.Fatalf("Error during follow: %s", err)
	}
}

func TestJournalGetUsage(t *testing.T) {
	j, err := NewJournal()

	if err != nil {
		t.Fatalf("Error opening journal: %s", err)
	}

	if j == nil {
		t.Fatal("Got a nil journal")
	}

	defer j.Close()

	_, err = j.GetUsage()

	if err != nil {
		t.Fatalf("Error getting journal size: %s", err)
	}
}

func TestJournalCursorGetSeekAndTest(t *testing.T) {
	j, err := NewJournal()
	if err != nil {
		t.Fatalf("Error opening journal: %s", err)
	}

	if j == nil {
		t.Fatal("Got a nil journal")
	}

	defer j.Close()

	waitAndNext := func(j *Journal) error {
		r := j.Wait(time.Duration(1) * time.Second)
		if r < 0 {
			return errors.New("Error waiting to journal")
		}

		n, err := j.Next()
		if err != nil {
			return fmt.Errorf("Error reading to journal: %s", err)
		}

		if n == 0 {
			return fmt.Errorf("Error reading to journal: %s", io.EOF)
		}

		return nil
	}

	err = journal.Print(journal.PriInfo, "test message for cursor %s", time.Now())
	if err != nil {
		t.Fatalf("Error writing to journal: %s", err)
	}

	if err = waitAndNext(j); err != nil {
		t.Fatalf(err.Error())
	}

	c, err := j.GetCursor()
	if err != nil {
		t.Fatalf("Error getting cursor from journal: %s", err)
	}

	err = j.SeekCursor(c)
	if err != nil {
		t.Fatalf("Error seeking cursor to journal: %s", err)
	}

	if err = waitAndNext(j); err != nil {
		t.Fatalf(err.Error())
	}

	err = j.TestCursor(c)
	if err != nil {
		t.Fatalf("Error testing cursor to journal: %s", err)
	}
}
