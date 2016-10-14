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

package terraform

import "encoding/json"

type Literal struct {
	value string
}

var _ json.Marshaler = &Literal{}

func (l *Literal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&l.value)
}

func LiteralExpression(s string) *Literal {
	return &Literal{value: s}
}

func LiteralSelfLink(resourceType, resourceName string) *Literal {
	return LiteralProperty(resourceType, resourceName, "self_link")
}

func LiteralProperty(resourceType, resourceName, prop string) *Literal {
	tfName := tfSanitize(resourceName)

	expr := "${" + resourceType + "." + tfName + "." + prop + "}"
	return LiteralExpression(expr)
}

func LiteralFromStringValue(s string) *Literal {
	return &Literal{value: s}
}
