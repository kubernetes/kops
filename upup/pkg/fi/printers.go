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

package fi

import (
	"fmt"

	"k8s.io/kops/pkg/values"
	"k8s.io/kops/util/pkg/reflectutils"
)

func init() {
	// Register our custom printer functions
	reflectutils.RegisterPrinter(PrintResource)
	reflectutils.RegisterPrinter(PrintResourceHolder)
	reflectutils.RegisterPrinter(PrintCompareWithID)
}

func PrintResource(o interface{}) (string, bool) {
	if _, ok := o.(Resource); !ok {
		return "", false
	}
	return "<resource>", true
}

func PrintResourceHolder(o interface{}) (string, bool) {
	if _, ok := o.(*ResourceHolder); !ok {
		return "", false
	}
	return "<resource>", true
}

func PrintCompareWithID(o interface{}) (string, bool) {
	compareWithID, ok := o.(CompareWithID)
	if !ok {
		return "", false
	}

	id := compareWithID.CompareWithID()
	name := ""
	hasName, ok := o.(HasName)
	if ok {
		name = values.StringValue(hasName.GetName())
	}
	if id == nil {
		// Uninformative, but we can often print the name instead
		if name != "" {
			return fmt.Sprintf("name:%s", name), true
		}
		return "id:<nil>", true
	}
	// Uninformative, but we can often print the name instead
	if name != "" {
		return fmt.Sprintf("name:%s id:%s", name, *id), true
	}
	return fmt.Sprintf("id:%s", *id), true
}
