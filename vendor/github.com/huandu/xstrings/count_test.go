// Copyright 2015 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package xstrings

import (
	"fmt"
	"testing"
)

func TestLen(t *testing.T) {
	runner := func(str string) string {
		return fmt.Sprint(Len(str))
	}

	runTestCases(t, runner, _M{
		"abcdef":    "6",
		"中文":        "2",
		"中yin文hun排": "9",
		"":          "0",
	})
}

func TestWordCount(t *testing.T) {
	runner := func(str string) string {
		return fmt.Sprint(WordCount(str))
	}

	runTestCases(t, runner, _M{
		"one word: λ":             "3",
		"中文":                      "0",
		"你好，sekai！":               "1",
		"oh, it's super-fancy!!a": "4",
		"":        "0",
		"-":       "0",
		"it's-'s": "1",
	})
}

func TestWidth(t *testing.T) {
	runner := func(str string) string {
		return fmt.Sprint(Width(str))
	}

	runTestCases(t, runner, _M{
		"abcd\t0123\n7890": "12",
		"中zh英eng文混排":       "15",
		"": "0",
	})
}

func TestRuneWidth(t *testing.T) {
	runner := func(str string) string {
		return fmt.Sprint(RuneWidth([]rune(str)[0]))
	}

	runTestCases(t, runner, _M{
		"a":    "1",
		"中":    "2",
		"\x11": "0",
	})
}
