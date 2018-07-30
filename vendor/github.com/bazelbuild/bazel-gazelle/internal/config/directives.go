/* Copyright 2017 The Bazel Authors. All rights reserved.

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

package config

import (
	"log"

	"github.com/bazelbuild/bazel-gazelle/internal/rule"
)

// ApplyDirectives applies directives that modify the configuration to a copy of
// c, which is returned. If there are no configuration directives, c is returned
// unmodified.
// TODO(jayconrod): delete this function and move all directive handling
// into configuration extensions.
func ApplyDirectives(c *Config, directives []rule.Directive, rel string) *Config {
	modified := *c
	didModify := false
	for _, d := range directives {
		switch d.Key {
		case "build_tags":
			if err := modified.SetBuildTags(d.Value); err != nil {
				log.Print(err)
				modified.GenericTags = c.GenericTags
			} else {
				modified.PreprocessTags()
				didModify = true
			}
		case "importmap_prefix":
			if err := CheckPrefix(d.Value); err != nil {
				log.Print(err)
				continue
			}
			modified.GoImportMapPrefix = d.Value
			modified.GoImportMapPrefixRel = rel
			didModify = true
		case "prefix":
			if err := CheckPrefix(d.Value); err != nil {
				log.Print(err)
				continue
			}
			modified.GoPrefix = d.Value
			modified.GoPrefixRel = rel
			didModify = true
		case "proto":
			protoMode, err := ProtoModeFromString(d.Value)
			if err != nil {
				log.Print(err)
				continue
			}
			modified.ProtoMode = protoMode
			modified.ProtoModeExplicit = true
			didModify = true
		}
	}
	if !didModify {
		return c
	}
	return &modified
}
