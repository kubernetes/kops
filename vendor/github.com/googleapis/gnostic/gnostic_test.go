package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func test_compiler(t *testing.T, input_file string, reference_file string, expect_errors bool) {
	text_file := strings.Replace(filepath.Base(input_file), filepath.Ext(input_file), ".text", 1)
	errors_file := strings.Replace(filepath.Base(input_file), filepath.Ext(input_file), ".errors", 1)
	// remove any preexisting output files
	os.Remove(text_file)
	os.Remove(errors_file)
	// run the compiler
	var err error
	var cmd = exec.Command(
		"gnostic",
		input_file,
		"--text_out=.",
		"--errors_out=.",
		"--resolve_refs")
	t.Log(cmd.Args)
	err = cmd.Run()
	if err != nil && !expect_errors {
		t.Logf("Compile failed: %+v", err)
		t.FailNow()
	}
	// verify the output against a reference
	var output_file string
	if expect_errors {
		output_file = errors_file
	} else {
		output_file = text_file
	}
	err = exec.Command("diff", output_file, reference_file).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(text_file)
		os.Remove(errors_file)
	}
}

func test_normal(t *testing.T, input_file string, reference_file string) {
	test_compiler(t, input_file, reference_file, false)
}

func test_errors(t *testing.T, input_file string, reference_file string) {
	test_compiler(t, input_file, reference_file, true)
}

func TestPetstoreJSON(t *testing.T) {
	test_normal(t,
		"examples/v2.0/json/petstore.json",
		"test/v2.0/petstore.text")
}

func TestPetstoreYAML(t *testing.T) {
	test_normal(t,
		"examples/v2.0/yaml/petstore.yaml",
		"test/v2.0/petstore.text")
}

func TestSeparateYAML(t *testing.T) {
	test_normal(t,
		"examples/v2.0/yaml/petstore-separate/spec/swagger.yaml",
		"test/v2.0/yaml/petstore-separate/spec/swagger.text")
}

func TestSeparateJSON(t *testing.T) {
	test_normal(t,
		"examples/v2.0/json/petstore-separate/spec/swagger.json",
		"test/v2.0/yaml/petstore-separate/spec/swagger.text") // yaml and json results should be identical
}

func TestRemotePetstoreJSON(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/v2.0/json/petstore.json",
		"test/v2.0/petstore.text")
}

func TestRemotePetstoreYAML(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/v2.0/yaml/petstore.yaml",
		"test/v2.0/petstore.text")
}

func TestRemoteSeparateYAML(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/v2.0/yaml/petstore-separate/spec/swagger.yaml",
		"test/v2.0/yaml/petstore-separate/spec/swagger.text")
}

func TestRemoteSeparateJSON(t *testing.T) {
	test_normal(t,
		"https://raw.githubusercontent.com/googleapis/openapi-compiler/master/examples/v2.0/json/petstore-separate/spec/swagger.json",
		"test/v2.0/yaml/petstore-separate/spec/swagger.text")
}

func TestErrorBadProperties(t *testing.T) {
	test_errors(t,
		"examples/errors/petstore-badproperties.yaml",
		"test/errors/petstore-badproperties.errors")
}

func TestErrorUnresolvedRefs(t *testing.T) {
	test_errors(t,
		"examples/errors/petstore-unresolvedrefs.yaml",
		"test/errors/petstore-unresolvedrefs.errors")
}

func test_plugin(t *testing.T, plugin string, input_file string, output_file string, reference_file string) {
	// remove any preexisting output files
	os.Remove(output_file)
	// run the compiler
	var err error
	output, err := exec.Command(
		"gnostic",
		"--"+plugin+"_out=-",
		input_file).Output()
	if err != nil {
		t.Logf("Compile failed: %+v", err)
		t.FailNow()
	}
	_ = ioutil.WriteFile(output_file, output, 0644)
	err = exec.Command("diff", output_file, reference_file).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(output_file)
	}
}

func TestSamplePluginWithPetstore(t *testing.T) {
	test_plugin(t,
		"go_sample",
		"examples/v2.0/yaml/petstore.yaml",
		"sample-petstore.out",
		"test/v2.0/yaml/sample-petstore.out")
}

func TestErrorInvalidPluginInvocations(t *testing.T) {
	var err error
	output, err := exec.Command(
		"gnostic",
		"examples/v2.0/yaml/petstore.yaml",
		"--errors_out=-",
		"--plugin_out=foo=bar,:abc",
		"--plugin_out=,foo=bar:abc",
		"--plugin_out=foo=:abc",
		"--plugin_out==bar:abc",
		"--plugin_out=,,:abc",
		"--plugin_out=foo=bar=baz:abc",
	).Output()
	if err == nil {
		t.Logf("Invalid invocations were accepted")
		t.FailNow()
	}
	output_file := "invalid-plugin-invocation.errors"
	_ = ioutil.WriteFile(output_file, output, 0644)
	err = exec.Command("diff", output_file, "test/errors/invalid-plugin-invocation.errors").Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(output_file)
	}
}

func TestValidPluginInvocations(t *testing.T) {
	var err error
	output, err := exec.Command(
		"gnostic",
		"examples/v2.0/yaml/petstore.yaml",
		"--errors_out=-",
		// verify an invocation with no parameters
		"--go_sample_out=!", // "!" indicates that no output should be generated
		// verify single pair of parameters
		"--go_sample_out=a=b:!",
		// verify multiple parameters
		"--go_sample_out=a=b,c=123,xyz=alphabetagammadelta:!",
		// verify that special characters / . - _ can be included in parameter keys and values
		"--go_sample_out=a/b/c=x/y/z:!",
		"--go_sample_out=a.b.c=x.y.z:!",
		"--go_sample_out=a-b-c=x-y-z:!",
		"--go_sample_out=a_b_c=x_y_z:!",
	).Output()
	if len(output) != 0 {
		t.Logf("Valid invocations generated invalid errors\n%s", string(output))
		t.FailNow()
	}
	if err != nil {
		t.Logf("Valid invocations were not accepted")
		t.FailNow()
	}
}

func TestExtensionHandlerWithLibraryExample(t *testing.T) {
	output_file := "library-example-with-ext.text.out"
	input_file := "test/library-example-with-ext.json"
	reference_file := "test/library-example-with-ext.text.out"

	os.Remove(output_file)
	// run the compiler
	var err error

	command := exec.Command(
		"gnostic",
		"--extension=samplecompanyone",
		"--extension=samplecompanytwo",
		"--text_out="+output_file,
		"--resolve_refs",
		input_file)

	_, err = command.Output()
	if err != nil {
		t.Logf("Compile failed for command %v: %+v", command, err)
		t.FailNow()
	}
	//_ = ioutil.WriteFile(output_file, output, 0644)
	err = exec.Command("diff", output_file, reference_file).Run()
	if err != nil {
		t.Logf("Diff failed: %+v", err)
		t.FailNow()
	} else {
		// if the test succeeded, clean up
		os.Remove(output_file)
	}
}

// OpenAPI 3.0 tests

func TestPetstoreYAML_30(t *testing.T) {
	test_normal(t,
		"examples/v3.0/yaml/petstore.yaml",
		"test/v3.0/petstore.text")
}

func TestPetstoreJSON_30(t *testing.T) {
	test_normal(t,
		"examples/v3.0/json/petstore.json",
		"test/v3.0/petstore.text")
}
