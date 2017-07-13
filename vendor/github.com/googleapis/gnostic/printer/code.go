// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package printer provides support for writing generated code.
package printer

import (
	"fmt"
)

const INDENT = "  "

type Code struct {
	text   string
	indent int
}

func (c *Code) Print(args ...interface{}) {
	if len(args) > 0 {
		for i := 0; i < c.indent; i++ {
			c.text += INDENT
		}
		c.text += fmt.Sprintf(args[0].(string), args[1:]...)
	}
	c.text += "\n"
}

func (c *Code) String() string {
	return c.text
}

func (c *Code) Indent() {
	c.indent += 1
}

func (c *Code) Outdent() {
	c.indent -= 1
	if c.indent < 0 {
		c.indent = 0
	}
}
