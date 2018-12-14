package sprig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	tpl := `{{"" | default "foo"}}`
	if err := runt(tpl, "foo"); err != nil {
		t.Error(err)
	}
	tpl = `{{default "foo" 234}}`
	if err := runt(tpl, "234"); err != nil {
		t.Error(err)
	}
	tpl = `{{default "foo" 2.34}}`
	if err := runt(tpl, "2.34"); err != nil {
		t.Error(err)
	}

	tpl = `{{ .Nothing | default "123" }}`
	if err := runt(tpl, "123"); err != nil {
		t.Error(err)
	}
	tpl = `{{ default "123" }}`
	if err := runt(tpl, "123"); err != nil {
		t.Error(err)
	}
}

func TestEmpty(t *testing.T) {
	tpl := `{{if empty 1}}1{{else}}0{{end}}`
	if err := runt(tpl, "0"); err != nil {
		t.Error(err)
	}

	tpl = `{{if empty 0}}1{{else}}0{{end}}`
	if err := runt(tpl, "1"); err != nil {
		t.Error(err)
	}
	tpl = `{{if empty ""}}1{{else}}0{{end}}`
	if err := runt(tpl, "1"); err != nil {
		t.Error(err)
	}
	tpl = `{{if empty 0.0}}1{{else}}0{{end}}`
	if err := runt(tpl, "1"); err != nil {
		t.Error(err)
	}
	tpl = `{{if empty false}}1{{else}}0{{end}}`
	if err := runt(tpl, "1"); err != nil {
		t.Error(err)
	}

	dict := map[string]interface{}{"top": map[string]interface{}{}}
	tpl = `{{if empty .top.NoSuchThing}}1{{else}}0{{end}}`
	if err := runtv(tpl, "1", dict); err != nil {
		t.Error(err)
	}
	tpl = `{{if empty .bottom.NoSuchThing}}1{{else}}0{{end}}`
	if err := runtv(tpl, "1", dict); err != nil {
		t.Error(err)
	}
}
func TestCoalesce(t *testing.T) {
	tests := map[string]string{
		`{{ coalesce 1 }}`:                            "1",
		`{{ coalesce "" 0 nil 2 }}`:                   "2",
		`{{ $two := 2 }}{{ coalesce "" 0 nil $two }}`: "2",
		`{{ $two := 2 }}{{ coalesce "" $two 0 0 0 }}`: "2",
		`{{ $two := 2 }}{{ coalesce "" $two 3 4 5 }}`: "2",
		`{{ coalesce }}`:                              "<no value>",
	}
	for tpl, expect := range tests {
		assert.NoError(t, runt(tpl, expect))
	}

	dict := map[string]interface{}{"top": map[string]interface{}{}}
	tpl := `{{ coalesce .top.NoSuchThing .bottom .bottom.dollar "airplane"}}`
	if err := runtv(tpl, "airplane", dict); err != nil {
		t.Error(err)
	}
}

func TestToJson(t *testing.T) {
	dict := map[string]interface{}{"Top": map[string]interface{}{"bool": true, "string": "test", "number": 42}}

	tpl := `{{.Top | toJson}}`
	expected := `{"bool":true,"number":42,"string":"test"}`
	if err := runtv(tpl, expected, dict); err != nil {
		t.Error(err)
	}
}

func TestToPrettyJson(t *testing.T) {
	dict := map[string]interface{}{"Top": map[string]interface{}{"bool": true, "string": "test", "number": 42}}
	tpl := `{{.Top | toPrettyJson}}`
	expected := `{
  "bool": true,
  "number": 42,
  "string": "test"
}`
	if err := runtv(tpl, expected, dict); err != nil {
		t.Error(err)
	}
}
