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
	"testing"
)

var simpleCall *CallExpr = &CallExpr{
	X: &LiteralExpr{
		Token: "java_library",
	},
	List: []Expr{
		&BinaryExpr{
			X: &LiteralExpr{
				Token: "name",
			},
			Op: "=",
			Y: &StringExpr{
				Value: "x",
			},
		},
	},
}

var simpleRule *Rule = &Rule{simpleCall}

var structCall *CallExpr = &CallExpr{
	X: &DotExpr{
		X: &DotExpr{
			X: &LiteralExpr{
				Token: "foo",
			},
			Name: "bar",
		},
		Name: "baz",
	},
	List: []Expr{
		&BinaryExpr{
			X: &LiteralExpr{
				Token: "name",
			},
			Op: "=",
			Y: &StringExpr{
				Value: "x",
			},
		},
	},
}

var structRule *Rule = &Rule{structCall}

func TestKind(t *testing.T) {
	if simpleRule.Kind() != "java_library" {
		t.Errorf(`simpleRule.Kind() = %v, want "java_library"`, simpleRule.Kind())
	}
	if structRule.Kind() != "foo.bar.baz" {
		t.Errorf(`structRule.Kind() = %v, want "foo.bar.baz"`, structRule.Kind())
	}
}

func TestSetKind(t *testing.T) {
	rule := &Rule{
		&CallExpr{
			X: &LiteralExpr{
				Token: "java_library",
			},
			List: []Expr{
				&BinaryExpr{
					X: &LiteralExpr{
						Token: "name",
					},
					Op: "=",
					Y: &StringExpr{
						Value: "x",
					},
				},
			},
		},
	}

	rule.SetKind("java_binary")
	compare(t, rule.Call.X, &LiteralExpr{Token: "java_binary"})

	rule.SetKind("foo.bar.baz")
	compare(t, rule.Call.X, &DotExpr{
		X: &DotExpr{
			X: &LiteralExpr{
				Token: "foo",
			},
			Name: "bar",
		},
		Name: "baz",
	})
}

func TestRules(t *testing.T) {
	f := &File{
		Stmt: []Expr{
			simpleCall,
			structCall,
		},
	}

	compare(t, f.Rules(""), []*Rule{simpleRule, structRule})
	compare(t, f.Rules("java_binary"), []*Rule(nil))
	compare(t, f.Rules("java_library"), []*Rule{simpleRule})
	compare(t, f.Rules("foo.bar.baz"), []*Rule{structRule})
}
