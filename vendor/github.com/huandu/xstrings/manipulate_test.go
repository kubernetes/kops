// Copyright 2015 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package xstrings

import (
	"strconv"
	"strings"
	"testing"
)

func TestReverse(t *testing.T) {
	runTestCases(t, Reverse, _M{
		"reverse string": "gnirts esrever",
		"中文如何？":          "？何如文中",
		"中en文混~排怎样？a":    "a？样怎排~混文ne中",
	})
}

func TestSlice(t *testing.T) {
	runner := func(str string) (result string) {
		defer func() {
			if e := recover(); e != nil {
				result = e.(string)
			}
		}()

		strs := split(str)
		start, _ := strconv.ParseInt(strs[1], 10, 0)
		end, _ := strconv.ParseInt(strs[2], 10, 0)

		result = Slice(strs[0], int(start), int(end))
		return
	}

	runTestCases(t, runner, _M{
		sep("abcdefghijk", "3", "8"):      "defgh",
		sep("来点中文如何？", "2", "7"):          "中文如何？",
		sep("中en文混~排总是少不了的a", "2", "8"):   "n文混~排总",
		sep("中en文混~排总是少不了的a", "0", "0"):   "",
		sep("中en文混~排总是少不了的a", "14", "14"): "",
		sep("中en文混~排总是少不了的a", "5", "-1"):  "~排总是少不了的a",
		sep("中en文混~排总是少不了的a", "14", "-1"): "",

		sep("let us slice out of range", "-3", "3"): "out of range",
		sep("超出范围哦", "2", "6"):                      "out of range",
		sep("don't do this", "3", "2"):              "out of range",
		sep("千gan万de不piao要liang", "19", "19"):       "out of range",
	})
}

func TestPartition(t *testing.T) {
	runner := func(str string) string {
		input := strings.Split(str, separator)
		head, match, tail := Partition(input[0], input[1])
		return sep(head, match, tail)
	}

	runTestCases(t, runner, _M{
		sep("hello", "l"):           sep("he", "l", "lo"),
		sep("中文总少不了", "少"):          sep("中文总", "少", "不了"),
		sep("z这个zh英文混排hao不", "h英文"): sep("z这个z", "h英文", "混排hao不"),
		sep("边界tiao件zen能忘", "边界"):   sep("", "边界", "tiao件zen能忘"),
		sep("尾巴ye别忘le", "忘le"):      sep("尾巴ye别", "忘le", ""),

		sep("hello", "x"):     sep("hello", "", ""),
		sep("不是晩香玉", "晚"):     sep("不是晩香玉", "", ""), // Hint: 晩 is not 晚 :)
		sep("来ge混排ba", "e 混"): sep("来ge混排ba", "", ""),
	})
}

func TestLastPartition(t *testing.T) {
	runner := func(str string) string {
		input := strings.Split(str, separator)
		head, match, tail := LastPartition(input[0], input[1])
		return sep(head, match, tail)
	}

	runTestCases(t, runner, _M{
		sep("hello", "l"):               sep("hel", "l", "o"),
		sep("少量中文总少不了", "少"):            sep("少量中文总", "少", "不了"),
		sep("z这个zh英文ch英文混排hao不", "h英文"): sep("z这个zh英文c", "h英文", "混排hao不"),
		sep("边界tiao件zen能忘边界", "边界"):     sep("边界tiao件zen能忘", "边界", ""),
		sep("尾巴ye别忘le", "尾巴"):           sep("", "尾巴", "ye别忘le"),

		sep("hello", "x"):     sep("", "", "hello"),
		sep("不是晩香玉", "晚"):     sep("", "", "不是晩香玉"), // Hint: 晩 is not 晚 :)
		sep("来ge混排ba", "e 混"): sep("", "", "来ge混排ba"),
	})
}

func TestInsert(t *testing.T) {
	runner := func(str string) (result string) {
		defer func() {
			if e := recover(); e != nil {
				result = e.(string)
			}
		}()

		strs := split(str)
		index, _ := strconv.ParseInt(strs[2], 10, 0)
		result = Insert(strs[0], strs[1], int(index))
		return
	}

	runTestCases(t, runner, _M{
		sep("abcdefg", "hi", "3"):    "abchidefg",
		sep("少量中文是必须的", "混pai", "4"): "少量中文混pai是必须的",
		sep("zh英文hun排", "~！", "5"):   "zh英文h~！un排",
		sep("插在begining", "我", "0"):  "我插在begining",
		sep("插在ending", "我", "8"):    "插在ending我",

		sep("超tian出yuan边tu界po", "foo", "-1"): "out of range",
		sep("超tian出yuan边tu界po", "foo", "17"): "out of range",
	})
}

func TestScrub(t *testing.T) {
	runner := func(str string) string {
		strs := split(str)
		return Scrub(strs[0], strs[1])
	}

	runTestCases(t, runner, _M{
		sep("ab\uFFFDcd\xFF\xCEefg\xFF\xFC\xFD\xFAhijk", "*"): "ab*cd*efg*hijk",
		sep("no错误です", "*"):                                    "no错误です",
		sep("", "*"):                                          "",
	})
}

func TestWordSplit(t *testing.T) {
	runner := func(str string) string {
		return sep(WordSplit(str)...)
	}

	runTestCases(t, runner, _M{
		"one word":                   sep("one", "word"),
		"一个字：把他给我拿下！":                "",
		"it's a super-fancy one!!!a": sep("it's", "a", "super-fancy", "one", "a"),
		"a -b-c' 'd'e":               sep("a", "b-c'", "d'e"),
	})
}
