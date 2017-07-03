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

//go:generate ./COMPILE-PROTOS.sh

// Gnostic is a tool for building better REST APIs through knowledge.
//
// Gnostic reads declarative descriptions of REST APIs that conform
// to the OpenAPI Specification, reports errors, resolves internal
// dependencies, and puts the results in a binary form that can
// be used in any language that is supported by the Protocol Buffer
// tools.
//
// Gnostic models are validated and typed. This allows API tool
// developers to focus on their product and not worry about input
// validation and type checking.
//
// Gnostic calls plugins that implement a variety of API implementation
// and support features including generation of client and server
// support code.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/googleapis/gnostic/OpenAPIv3"
	"github.com/googleapis/gnostic/compiler"
	plugins "github.com/googleapis/gnostic/plugins"
)

const ( // OpenAPI Version
	OpenAPIvUnknown = 0
	OpenAPIv2       = 2
	OpenAPIv3       = 3
)

func openapi_version(info interface{}) int {
	m, ok := compiler.UnpackMap(info)
	if !ok {
		return OpenAPIvUnknown
	}
	swagger, ok := compiler.MapValueForKey(m, "swagger").(string)
	if ok && swagger == "2.0" {
		return OpenAPIv2
	}
	openapi, ok := compiler.MapValueForKey(m, "openapi").(string)
	if ok && openapi == "3.0" {
		return OpenAPIv3
	}
	return OpenAPIvUnknown
}

type PluginCall struct {
	Name       string
	Invocation string
}

func (pluginCall *PluginCall) perform(document proto.Message, openapi_version int, sourceName string) error {
	if pluginCall.Name != "" {
		request := &plugins.Request{}

		// Infer the name of the executable by adding the prefix.
		executableName := "gnostic_" + pluginCall.Name

		// validate invocation string with regular expression
		invocation := pluginCall.Invocation

		//
		// Plugin invocations must consist of
		// zero or more comma-separated key=value pairs followed by a path.
		// If pairs are present, a colon separates them from the path.
		// Keys and values must be alphanumeric strings and may contain
		// dashes, underscores, periods, or forward slashes.
		// A path can contain any characters other than the separators ',', ':', and '='.
		//
		invocation_regex := regexp.MustCompile(`^([\w-_\/\.]+=[\w-_\/\.]+(,[\w-_\/\.]+=[\w-_\/\.]+)*:)?[^,:=]+$`)
		if !invocation_regex.Match([]byte(pluginCall.Invocation)) {
			return errors.New(fmt.Sprintf("Invalid invocation of %s: %s", executableName, invocation))
		}

		invocationParts := strings.Split(pluginCall.Invocation, ":")
		var outputLocation string
		switch len(invocationParts) {
		case 1:
			outputLocation = invocationParts[0]
		case 2:
			parameters := strings.Split(invocationParts[0], ",")
			for _, keyvalue := range parameters {
				pair := strings.Split(keyvalue, "=")
				if len(pair) == 2 {
					request.Parameters = append(request.Parameters, &plugins.Parameter{Name: pair[0], Value: pair[1]})
				}
			}
			outputLocation = invocationParts[1]
		default:
			// badly-formed request
			outputLocation = invocationParts[len(invocationParts)-1]
		}

		version := &plugins.Version{}
		version.Major = 0
		version.Minor = 1
		version.Patch = 0
		request.CompilerVersion = version

		request.OutputPath = outputLocation

		wrapper := &plugins.Wrapper{}
		wrapper.Name = sourceName
		switch openapi_version {
		case OpenAPIv2:
			wrapper.Version = "v2"
		case OpenAPIv3:
			wrapper.Version = "v3"
		default:
			wrapper.Version = "unknown"
		}
		protoBytes, _ := proto.Marshal(document)
		wrapper.Value = protoBytes
		request.Wrapper = wrapper
		requestBytes, _ := proto.Marshal(request)

		cmd := exec.Command(executableName)
		cmd.Stdin = bytes.NewReader(requestBytes)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		response := &plugins.Response{}
		err = proto.Unmarshal(output, response)
		if err != nil {
			return err
		}

		if response.Errors != nil {
			return errors.New(fmt.Sprintf("Plugin error: %+v", response.Errors))
		}

		// write files to the specified directory
		var writer io.Writer
		if outputLocation == "!" {
			// write nothing
		} else if outputLocation == "-" {
			writer = os.Stdout
			for _, file := range response.Files {
				writer.Write([]byte("\n\n" + file.Name + " -------------------- \n"))
				writer.Write(file.Data)
			}
		} else if isFile(outputLocation) {
			return errors.New(fmt.Sprintf("Error, unable to overwrite %s\n", outputLocation))
		} else {
			if !isDirectory(outputLocation) {
				os.Mkdir(outputLocation, 0755)
			}
			for _, file := range response.Files {
				p := outputLocation + "/" + file.Name
				dir := path.Dir(p)
				os.MkdirAll(dir, 0755)
				f, _ := os.Create(p)
				defer f.Close()
				f.Write(file.Data)
			}
		}
	}
	return nil
}

func isFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func writeFile(name string, bytes []byte, source string, extension string) {
	var writer io.Writer
	if name == "!" {
		return
	} else if name == "-" {
		writer = os.Stdout
	} else if name == "=" {
		writer = os.Stderr
	} else if isDirectory(name) {
		base := filepath.Base(source)
		// remove the original source extension
		base = base[0 : len(base)-len(filepath.Ext(base))]
		// build the path that puts the result in the passed-in directory
		filename := name + "/" + base + "." + extension
		file, _ := os.Create(filename)
		defer file.Close()
		writer = file
	} else {
		file, _ := os.Create(name)
		defer file.Close()
		writer = file
	}
	writer.Write(bytes)
	if name == "-" || name == "=" {
		writer.Write([]byte("\n"))
	}
}

func main() {
	usage := `
Usage: gnostic OPENAPI_SOURCE [OPTIONS]
  OPENAPI_SOURCE is the filename or URL of an OpenAPI description to read.
Options:
  --pb_out=PATH       Write a binary proto to the specified location.
  --json_out=PATH     Write a json proto to the specified location.
  --text_out=PATH     Write a text proto to the specified location.
  --errors_out=PATH   Write compilation errors to the specified location.
  --PLUGIN_out=PATH   Run the plugin named gnostic_PLUGIN and write results
                      to the specified location.
  --resolve_refs      Explicitly resolve $ref references.
                      This could have problems with recursive definitions.
  --extension=NAME    Run the specified gnostic extension to process
                      extensions found in OpenAPI descriptions.
`
	// default values for all options
	sourceName := ""
	binaryProtoPath := ""
	jsonProtoPath := ""
	textProtoPath := ""
	errorPath := ""
	pluginCalls := make([]*PluginCall, 0)
	resolveReferences := false
	extensionHandlers := make([]compiler.ExtensionHandler, 0)

	// arg processing matches patterns of the form "--PLUGIN_out=PATH"
	plugin_regex := regexp.MustCompile("--(.+)_out=(.+)")

	// arg processing matches patterns of the form "--extension=GENERATOR_NAME"
	extensionHandler_regex, err := regexp.Compile("--extension=(.+)")
	defaultPrefixForExtensions := "openapi_extensions_"

	for i, arg := range os.Args {
		if i == 0 {
			continue // skip the tool name
		}
		var m [][]byte
		if m = plugin_regex.FindSubmatch([]byte(arg)); m != nil {
			pluginName := string(m[1])
			invocation := string(m[2])
			switch pluginName {
			case "pb":
				binaryProtoPath = invocation
			case "json":
				jsonProtoPath = invocation
			case "text":
				textProtoPath = invocation
			case "errors":
				errorPath = invocation
			default:
				pluginCall := &PluginCall{Name: pluginName, Invocation: invocation}
				pluginCalls = append(pluginCalls, pluginCall)
			}
		} else if m = extensionHandler_regex.FindSubmatch([]byte(arg)); m != nil {
			extensionHandlers = append(extensionHandlers, compiler.ExtensionHandler{Name: defaultPrefixForExtensions + string(m[1])})
		} else if arg == "--resolve_refs" {
			resolveReferences = true
		} else if arg[0] == '-' {
			fmt.Fprintf(os.Stderr, "Unknown option: %s.\n%s\n", arg, usage)
			os.Exit(-1)
		} else {
			sourceName = arg
		}
	}

	if binaryProtoPath == "" &&
		jsonProtoPath == "" &&
		textProtoPath == "" &&
		errorPath == "" &&
		len(pluginCalls) == 0 {
		fmt.Fprintf(os.Stderr, "Missing output directives.\n%s\n", usage)
		os.Exit(-1)
	}

	if sourceName == "" {
		fmt.Fprintf(os.Stderr, "No input specified.\n%s\n", usage)
		os.Exit(-1)
	}

	errorPrefix := "Errors reading " + sourceName + "\n"

	// If we get here and the error output is unspecified, write errors to stderr.
	if errorPath == "" {
		errorPath = "="
	}

	// Read the OpenAPI source.
	info, err := compiler.ReadInfoForFile(sourceName)
	if err != nil {
		writeFile(errorPath, []byte(errorPrefix+err.Error()), sourceName, "errors")
		os.Exit(-1)
	}

	// Determine the OpenAPI version.
	openapi_version := openapi_version(info)
	if openapi_version == OpenAPIvUnknown {
		fmt.Fprintf(os.Stderr, "Unknown OpenAPI Version\n")
		os.Exit(-1)
	}

	var message proto.Message
	if openapi_version == OpenAPIv2 {
		document, err := openapi_v2.NewDocument(info, compiler.NewContextWithExtensions("$root", nil, &extensionHandlers))
		if err != nil {
			writeFile(errorPath, []byte(errorPrefix+err.Error()), sourceName, "errors")
			os.Exit(-1)
		}
		// optionally resolve internal references
		if resolveReferences {
			_, err = document.ResolveReferences(sourceName)
			if err != nil {
				writeFile(errorPath, []byte(errorPrefix+err.Error()), sourceName, "errors")
				os.Exit(-1)
			}
		}
		message = document
	} else if openapi_version == OpenAPIv3 {
		document, err := openapi_v3.NewDocument(info, compiler.NewContextWithExtensions("$root", nil, &extensionHandlers))
		if err != nil {
			writeFile(errorPath, []byte(errorPrefix+err.Error()), sourceName, "errors")
			os.Exit(-1)
		}
		// optionally resolve internal references
		if resolveReferences {
			_, err = document.ResolveReferences(sourceName)
			if err != nil {
				writeFile(errorPath, []byte(errorPrefix+err.Error()), sourceName, "errors")
				os.Exit(-1)
			}
		}
		message = document
	}

	// perform all specified actions
	if binaryProtoPath != "" {
		// write proto in binary format
		protoBytes, _ := proto.Marshal(message)
		writeFile(binaryProtoPath, protoBytes, sourceName, "pb")
	}
	if jsonProtoPath != "" {
		// write proto in json format
		jsonBytes, _ := json.Marshal(message)
		writeFile(jsonProtoPath, jsonBytes, sourceName, "json")
	}
	if textProtoPath != "" {
		// write proto in text format
		bytes := []byte(proto.MarshalTextString(message))
		writeFile(textProtoPath, bytes, sourceName, "text")
	}
	for _, pluginCall := range pluginCalls {
		err = pluginCall.perform(message, openapi_version, sourceName)
		if err != nil {
			writeFile(errorPath, []byte(errorPrefix+err.Error()), sourceName, "errors")
			defer os.Exit(-1)
		}
	}
}
