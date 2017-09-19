// Copyright 2015 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package xstrings

import (
	"strconv"
	"strings"
	"testing"
)

func TestExpandTabs(t *testing.T) {
	runner := func(str string) (result string) {
		defer func() {
			if e := recover(); e != nil {
				result = e.(string)
			}
		}()

		input := strings.Split(str, separator)
		n, _ := strconv.Atoi(input[1])
		return ExpandTabs(input[0], n)
	}

	runTestCases(t, runner, _M{
		sep("a\tbc\tdef\tghij\tk", "4"): "a   bc  def ghij    k",
		sep("abcdefg\thij\nk\tl", "4"):  "abcdefg hij\nk   l",
		sep("z中\t文\tw", "4"):            "z中 文  w",
		sep("abcdef", "4"):              "abcdef",

		sep("abc\td\tef\tghij\nk\tl", "3"): "abc   d  ef ghij\nk  l",
		sep("abc\td\tef\tghij\nk\tl", "1"): "abc d ef ghij\nk l",

		sep("abc", "0"):  "tab size must be positive",
		sep("abc", "-1"): "tab size must be positive",
	})
}

func TestLeftJustify(t *testing.T) {
	runner := func(str string) string {
		input := strings.Split(str, separator)
		n, _ := strconv.Atoi(input[1])
		return LeftJustify(input[0], n, input[2])
	}

	runTestCases(t, runner, _M{
		sep("hello", "4", " "):    "hello",
		sep("hello", "10", " "):   "hello     ",
		sep("hello", "10", "123"): "hello12312",

		sep("hello中文test", "4", " "):    "hello中文test",
		sep("hello中文test", "12", " "):   "hello中文test ",
		sep("hello中文test", "18", "测试！"): "hello中文test测试！测试！测",

		sep("hello中文test", "0", "123"): "hello中文test",
		sep("hello中文test", "18", ""):   "hello中文test",
	})
}

func TestRightJustify(t *testing.T) {
	runner := func(str string) string {
		input := strings.Split(str, separator)
		n, _ := strconv.Atoi(input[1])
		return RightJustify(input[0], n, input[2])
	}

	runTestCases(t, runner, _M{
		sep("hello", "4", " "):    "hello",
		sep("hello", "10", " "):   "     hello",
		sep("hello", "10", "123"): "12312hello",

		sep("hello中文test", "4", " "):    "hello中文test",
		sep("hello中文test", "12", " "):   " hello中文test",
		sep("hello中文test", "18", "测试！"): "测试！测试！测hello中文test",

		sep("hello中文test", "0", "123"): "hello中文test",
		sep("hello中文test", "18", ""):   "hello中文test",
	})
}

func TestCenter(t *testing.T) {
	runner := func(str string) string {
		input := strings.Split(str, separator)
		n, _ := strconv.Atoi(input[1])
		return Center(input[0], n, input[2])
	}

	runTestCases(t, runner, _M{
		sep("hello", "4", " "):    "hello",
		sep("hello", "10", " "):   "  hello   ",
		sep("hello", "10", "123"): "12hello123",

		sep("hello中文test", "4", " "):    "hello中文test",
		sep("hello中文test", "12", " "):   "hello中文test ",
		sep("hello中文test", "18", "测试！"): "测试！hello中文test测试！测",

		sep("hello中文test", "0", "123"): "hello中文test",
		sep("hello中文test", "18", ""):   "hello中文test",
	})
}
