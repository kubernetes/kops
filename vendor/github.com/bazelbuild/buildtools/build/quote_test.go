/*
Copyright 2016 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package build

import (
	"strings"
	"testing"
)

var quoteTests = []struct {
	q   string // quoted
	s   string // unquoted (actual string)
	std bool   // q is standard form for s
}{
	{`""`, "", true},
	{`''`, "", false},
	{`"hello"`, `hello`, true},
	{`'hello'`, `hello`, false},
	{`"quote\"here"`, `quote"here`, true},
	{`'quote\"here'`, `quote"here`, false},
	{`'quote"here'`, `quote"here`, false},
	{`"quote'here"`, `quote'here`, true},
	{`"quote\'here"`, `quote'here`, false},
	{`'quote\'here'`, `quote'here`, false},
	{`"""hello " ' world "" asdf ''' foo"""`, `hello " ' world "" asdf ''' foo`, true},
	{`"foo\(bar"`, `foo\(bar`, true},
	{`"""hello
world"""`, "hello\nworld", true},

	{`"\a\b\f\n\r\t\v\000\377"`, "\a\b\f\n\r\t\v\000\xFF", true},
	{`"\a\b\f\n\r\t\v\x00\xff"`, "\a\b\f\n\r\t\v\000\xFF", false},
	{`"\a\b\f\n\r\t\v\000\xFF"`, "\a\b\f\n\r\t\v\000\xFF", false},
	{`"\a\b\f\n\r\t\v\000\377\"'\\\003\200"`, "\a\b\f\n\r\t\v\x00\xFF\"'\\\x03\x80", true},
	{`"\a\b\f\n\r\t\v\x00\xff\"'\\\x03\x80"`, "\a\b\f\n\r\t\v\x00\xFF\"'\\\x03\x80", false},
	{`"\a\b\f\n\r\t\v\000\xFF\"'\\\x03\x80"`, "\a\b\f\n\r\t\v\x00\xFF\"'\\\x03\x80", false},
	{`"\a\b\f\n\r\t\v\000\xFF\"\'\\\x03\x80"`, "\a\b\f\n\r\t\v\x00\xFF\"'\\\x03\x80", false},
	{
		`"cat $(SRCS) | grep '\s*ip_block:' | sed -e 's/\s*ip_block: \"\([^ ]*\)\"/    \x27\\1\x27,/g' >> $@; "`,
		"cat $(SRCS) | grep '\\s*ip_block:' | sed -e 's/\\s*ip_block: \"\\([^ ]*\\)\"/    '\\1',/g' >> $@; ",
		false,
	},
	{
		`"cat $(SRCS) | grep '\\s*ip_block:' | sed -e 's/\\s*ip_block: \"\([^ ]*\)\"/    '\\1',/g' >> $@; "`,
		"cat $(SRCS) | grep '\\s*ip_block:' | sed -e 's/\\s*ip_block: \"\\([^ ]*\\)\"/    '\\1',/g' >> $@; ",
		true,
	},
}

func TestQuote(t *testing.T) {
	for _, tt := range quoteTests {
		if !tt.std {
			continue
		}
		q := quote(tt.s, strings.HasPrefix(tt.q, `"""`))
		if q != tt.q {
			t.Errorf("quote(%#q) = %s, want %s", tt.s, q, tt.q)
		}
	}
}

func TestUnquote(t *testing.T) {
	for _, tt := range quoteTests {
		s, triple, err := unquote(tt.q)
		wantTriple := strings.HasPrefix(tt.q, `"""`) || strings.HasPrefix(tt.q, `'''`)
		if s != tt.s || triple != wantTriple || err != nil {
			t.Errorf("unquote(%s) = %#q, %v, %v want %#q, %v, nil", tt.q, s, triple, err, tt.s, wantTriple)
		}
	}
}
