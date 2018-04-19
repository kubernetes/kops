/*
Copyright 2016 The Kubernetes Authors.

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

package api

import (
	"flag"
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/loads"
)

var BuildOps = flag.Bool("build-operations", true, "If true build operations in the docs.")

// OperationCategory defines a group of related operations
type OperationCategory struct {
	// Name is the display name of this group
	Name string `yaml:",omitempty"`
	// Operations are the collection of Operations in this group
	OperationTypes []OperationType `yaml:"operation_types,omitempty"`
	// Default is true if this is the default operation group for operations that do not match any other groups
	Default bool `yaml:",omitempty"`

	Operations []*Operation
}

// Operation defines a highlevel operation type such as Read, Replace, Patch
type OperationType struct {
	// Name is the display name of this operation
	Name string `yaml:",omitempty"`
	// Match is the regular expression of operation IDs that match this group where '${resource}' matches the resource name.
	Match string `yaml:",omitempty"`
}

// GetOperationId returns the ID of the operation for the given definition
func (ot OperationType) GetOperationId(definition string) string {
	return strings.Replace(ot.Match, "${resource}", definition, -1)
}

type Operations map[string]*Operation

type Operation struct {
	item          spec.PathItem
	op            *spec.Operation
	ID            string
	Type          OperationType
	Path          string
	HttpMethod    string
	Definition    *Definition
	BodyParams    Fields
	QueryParams   Fields
	PathParams    Fields
	HttpResponses HttpResponses

	ExampleConfig ExampleConfig
}

type ExampleText struct {
	Tab  string
	Type string
	Text string
	Msg  string
}

func (o *Operation) GetExampleRequests() []ExampleText {
	r := []ExampleText{}
	for _, p := range GetExampleProviders() {
		r = append(r, ExampleText{
			Tab: p.GetTab(),
			Type: p.GetRequestType(),
			Text: p.GetRequest(o),
			Msg:  p.GetRequestMessage(),
		})
	}
	return r
}

func (o *Operation) GetExampleResponses() []ExampleText {
	r := []ExampleText{}
	for _, p := range GetExampleProviders() {
		r = append(r, ExampleText{
			Tab: p.GetTab(),
			Type: p.GetResponseType(),
			Text: p.GetResponse(o),
			Msg: p.GetResponseMessage(),
		})
	}
	return r
}

func (o *Operation) Description() string {
	return o.op.Description
}

type HttpResponses []*HttpResponse

func (a HttpResponses) Len() int      { return len(a) }
func (a HttpResponses) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a HttpResponses) Less(i, j int) bool {
	return a[i].Code < a[j].Code
}

type HttpResponse struct {
	Field
	Code string
}

// VisitOperations calls fn once for each operation found in the collection of Documents
func VisitOperations(specs []*loads.Document, fn func(operation Operation)) {
	for _, d := range specs {
		for path, item := range d.Spec().Paths.Paths {
			for method, operation := range getOperationsForItem(item) {
				if operation != nil && !IsBlacklistedOperation(operation) {
					fn(Operation{
						item:       item,
						op:         operation,
						Path:       path,
						HttpMethod: method,
						ID:         operation.ID,
					})
				}
			}
		}
	}
}

func IsBlacklistedOperation(o *spec.Operation) bool {
	return strings.HasSuffix(o.ID, "APIGroup") || // These are just the API group meta datas.  Ignore for now.
		strings.HasSuffix(o.ID, "APIResources") || // These are just the API group meta datas.  Ignore for now.
		strings.HasSuffix(o.ID, "APIVersions") // || // These are just the API group meta datas.  Ignore for now.
		//strings.HasPrefix(o.ID, "connect") || // Skip pod connect apis for now.  There are too many.
		//strings.HasPrefix(o.ID, "proxy")
}

// Get all operations from the pathitem so we cacn iterate over them
func getOperationsForItem(pathItem spec.PathItem) map[string]*spec.Operation {
	return map[string]*spec.Operation{
		"GET":    pathItem.Get,
		"DELETE": pathItem.Delete,
		"PATCH":  pathItem.Patch,
		"PUT":    pathItem.Put,
		"POST":   pathItem.Post,
		"HEAD":   pathItem.Head,
	}
}

func (operation *Operation) GetDisplayHttp() string {
	return fmt.Sprintf("%s %s", operation.HttpMethod, operation.Path)
}
