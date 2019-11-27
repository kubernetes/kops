/*
Copyright 2019 The Kubernetes Authors.

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

package templater

import (
	"io/ioutil"
	"testing"

	"k8s.io/kops/pkg/diff"

	yaml "gopkg.in/yaml.v2"
)

func TestRenderGeneralOK(t *testing.T) {
	cases := []renderTest{
		{
			Template: "hello",
			Expected: "hello",
		},
		{
			Template: `{{ lower "Hello" }}`,
			Expected: "hello",
		},
		{
			Template: `{{ upper "Hello" }}`,
			Expected: "HELLO",
		},
		{
			Context:  map[string]interface{}{"list": []string{"a", "b", "c"}},
			Template: `{{ .list | join "," }}`,
			Expected: "a,b,c",
		},
	}
	makeRenderTests(t, cases)
}

func TestRenderMissingValue(t *testing.T) {
	cases := []renderTest{
		{
			Context:  map[string]interface{}{"missing": "no"},
			Template: `{{ .missing  }}`,
			Expected: "no",
		},
		{
			Context:  map[string]interface{}{"is_missing": "no"},
			Template: `{{ .missing  }}`,
			NotOK:    true,
		},
		{
			Context:  map[string]interface{}{"missing": "no"},
			Snippets: map[string]string{"snip": "{{ .is_missing }}"},
			Template: `{{ .missing  }}{{ include "snip" . }}`,
			NotOK:    true,
		},
	}
	makeRenderTests(t, cases)
}

func TestRenderIndent(t *testing.T) {
	cases := []renderTest{
		{
			Context:  map[string]interface{}{"line": "this is a line of\ntext"},
			Template: `{{ .line | indent 2 }}`,
			Expected: "this is a line of\n  text",
		},
	}
	makeRenderTests(t, cases)
}

func TestRenderSnippet(t *testing.T) {
	cases := []renderTest{
		{
			Context:  map[string]interface{}{"name": "world"},
			Snippets: map[string]string{"snip": "hello world"},
			Template: `this should say {{ include "snip" . }}`,
			Expected: "this should say hello world",
		},
		{
			Context: map[string]interface{}{"name": "world"},
			Snippets: map[string]string{
				"one": "hello world",
				"two": "hello everyone",
			},
			Template: `this should say {{ include "one" . }} {{ include "two" . }}`,
			Expected: "this should say hello world hello everyone",
		},
	}
	makeRenderTests(t, cases)
}

func TestRenderContext(t *testing.T) {
	cases := []renderTest{
		{
			Context:  map[string]interface{}{"name": "world"},
			Template: `hello {{ .name }}`,
			Expected: "hello world",
		},
		{
			Context:  map[string]interface{}{"name": "world", "id": 99},
			Template: `hello {{ .name }} {{.id}}`,
			Expected: "hello world 99",
		},
		{
			Context: map[string]interface{}{
				"struct": map[string]interface{}{
					"id":   1,
					"name": "test",
				},
			},
			Template: `hello {{ .struct.name }} {{ .struct.id }}`,
			Expected: "hello test 1",
		},
		{
			Context: map[string]interface{}{
				"members": []struct {
					Name    string
					Members []string
				}{
					{
						Name:    "etcd0",
						Members: []string{"1", "2"},
					},
					{
						Name:    "etcd1",
						Members: []string{"1", "2"},
					},
				},
			},
			Template: `{{ range .members }}{{ .Name }},{{ end }}`,
			Expected: "etcd0,etcd1,",
		},
	}
	makeRenderTests(t, cases)
}

func TestAllowForMissingVars(t *testing.T) {
	cases := []renderTest{
		{
			Context:        map[string]interface{}{},
			Template:       `{{ default "is missing" .name }}`,
			Expected:       "is missing",
			DisableMissing: true,
		},
	}
	makeRenderTests(t, cases)
}

func TestRenderIntegration(t *testing.T) {
	var cases []renderTest
	content, err := ioutil.ReadFile("integration_tests.yml")
	if err != nil {
		t.Fatalf("unable to load the integration tests, error: %s", err)
	}
	if err := yaml.Unmarshal(content, &cases); err != nil {
		t.Fatalf("unable to decode the integration tests, error: %s", err)
	}

	makeRenderTests(t, cases)
}

type renderTest struct {
	Context        map[string]interface{}
	DisableMissing bool
	Expected       string
	NotOK          bool
	Snippets       map[string]string
	Template       string
}

func makeRenderTests(t *testing.T, tests []renderTest) {
	r := NewTemplater()
	for i, x := range tests {
		render, err := r.Render(x.Template, x.Context, x.Snippets, !x.DisableMissing)
		if x.NotOK {
			if err == nil {
				t.Errorf("case %d: should have thrown an error", i)
			}
			continue
		}
		if err != nil {
			t.Errorf("case %d: failed to render template, error: %s", i, err)
			continue
		}
		if x.Expected != render {
			t.Logf("diff:\n%s\n", diff.FormatDiff(x.Expected, render))
			t.Errorf("case %d failed, policy output differed from expected.", i)
		}
	}
}
