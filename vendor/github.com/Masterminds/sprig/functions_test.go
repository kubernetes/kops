package sprig

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"text/template"

	"github.com/aokoli/goutils"
	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	os.Setenv("FOO", "bar")
	tpl := `{{env "FOO"}}`
	if err := runt(tpl, "bar"); err != nil {
		t.Error(err)
	}
}

func TestExpandEnv(t *testing.T) {
	os.Setenv("FOO", "bar")
	tpl := `{{expandenv "Hello $FOO"}}`
	if err := runt(tpl, "Hello bar"); err != nil {
		t.Error(err)
	}
}

func TestBase(t *testing.T) {
	assert.NoError(t, runt(`{{ base "foo/bar" }}`, "bar"))
}

func TestDir(t *testing.T) {
	assert.NoError(t, runt(`{{ dir "foo/bar/baz" }}`, "foo/bar"))
}

func TestIsAbs(t *testing.T) {
	assert.NoError(t, runt(`{{ isAbs "/foo" }}`, "true"))
	assert.NoError(t, runt(`{{ isAbs "foo" }}`, "false"))
}

func TestClean(t *testing.T) {
	assert.NoError(t, runt(`{{ clean "/foo/../foo/../bar" }}`, "/bar"))
}

func TestExt(t *testing.T) {
	assert.NoError(t, runt(`{{ ext "/foo/bar/baz.txt" }}`, ".txt"))
}

func TestSnakeCase(t *testing.T) {
	assert.NoError(t, runt(`{{ snakecase "FirstName" }}`, "first_name"))
	assert.NoError(t, runt(`{{ snakecase "HTTPServer" }}`, "http_server"))
	assert.NoError(t, runt(`{{ snakecase "NoHTTPS" }}`, "no_https"))
	assert.NoError(t, runt(`{{ snakecase "GO_PATH" }}`, "go_path"))
	assert.NoError(t, runt(`{{ snakecase "GO PATH" }}`, "go_path"))
	assert.NoError(t, runt(`{{ snakecase "GO-PATH" }}`, "go_path"))
}

func TestCamelCase(t *testing.T) {
	assert.NoError(t, runt(`{{ camelcase "http_server" }}`, "HttpServer"))
	assert.NoError(t, runt(`{{ camelcase "_camel_case" }}`, "_CamelCase"))
	assert.NoError(t, runt(`{{ camelcase "no_https" }}`, "NoHttps"))
	assert.NoError(t, runt(`{{ camelcase "_complex__case_" }}`, "_Complex_Case_"))
	assert.NoError(t, runt(`{{ camelcase "all" }}`, "All"))
}

func TestShuffle(t *testing.T) {
	goutils.RANDOM = rand.New(rand.NewSource(1))
	// Because we're using a random number generator, we need these to go in
	// a predictable sequence:
	assert.NoError(t, runt(`{{ shuffle "Hello World" }}`, "rldo HWlloe"))
}

// runt runs a template and checks that the output exactly matches the expected string.
func runt(tpl, expect string) error {
	return runtv(tpl, expect, map[string]string{})
}

// runtv takes a template, and expected return, and values for substitution.
//
// It runs the template and verifies that the output is an exact match.
func runtv(tpl, expect string, vars interface{}) error {
	fmap := TxtFuncMap()
	t := template.Must(template.New("test").Funcs(fmap).Parse(tpl))
	var b bytes.Buffer
	err := t.Execute(&b, vars)
	if err != nil {
		return err
	}
	if expect != b.String() {
		return fmt.Errorf("Expected '%s', got '%s'", expect, b.String())
	}
	return nil
}

// runRaw runs a template with the given variables and returns the result.
func runRaw(tpl string, vars interface{}) (string, error) {
	fmap := TxtFuncMap()
	t := template.Must(template.New("test").Funcs(fmap).Parse(tpl))
	var b bytes.Buffer
	err := t.Execute(&b, vars)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
