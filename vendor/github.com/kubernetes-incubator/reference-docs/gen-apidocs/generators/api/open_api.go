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
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-openapi/loads"
)

// Loads all of the open-api documents
func LoadOpenApiSpec() []*loads.Document {
	dir := filepath.Join(*ConfigDir, "openapi-spec/")
	docs := []*loads.Document{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext != ".json" {
			return nil
		}
		var d *loads.Document
		d, err = loads.JSONSpec(path)
		if err != nil {
			return fmt.Errorf("Could not load json file %s as api-spec: %v\n", path, err)
		}
		docs = append(docs, d)
		return nil
	})
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
	return docs
}
