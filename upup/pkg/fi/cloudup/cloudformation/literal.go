/*
Copyright 2017 The Kubernetes Authors.

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

package cloudformation

import (
	"encoding/json"
)

type Literal struct {
	json interface{}
}

var _ json.Marshaler = &Literal{}

func (l *Literal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&l.json)
}

func (l *Literal) extractRef() string {
	m, ok := l.json.(map[string]interface{})
	if !ok {
		return ""
	}
	ref := m["Ref"]
	if ref == nil {
		return ""
	}
	s, ok := ref.(string)
	if !ok {
		return ""
	}
	return s
}

func literalRef(s string) *Literal {
	j := make(map[string]interface{})
	j["Ref"] = s
	return &Literal{json: j}
}

func Ref(resourceType, resourceName string) *Literal {
	return literalRef(sanitizeCloudformationResourceName(resourceType + "::" + resourceName))
}

func GetAtt(resourceType, resourceName string, attribute string) *Literal {
	path := []string{
		sanitizeCloudformationResourceName(resourceType + "::" + resourceName),
		attribute,
	}
	j := make(map[string]interface{})
	j["Fn::GetAtt"] = path
	return &Literal{json: j}
}

func LiteralString(v string) *Literal {
	j := &v
	return &Literal{json: j}
}

//
//func LiteralSelfLink(resourceType, resourceName string) *Literal {
//	return LiteralProperty(resourceType, resourceName, "self_link")
//}
//
//
//func DefaultProperty(resourceType, resourceName string) *Literal {
//	return LiteralProperty(resourceType, resourceName, "")
//}
//
//func LiteralProperty(resourceType, resourceName, prop string) *Literal {
//	tfName := sanitizeCloudformationResourceName( resourceType + "::" + resourceName)
//
//	expr := "${" + resourceType + "." + tfName + "." + prop + "}"
//	return LiteralExpression(expr)
//}
//
//func LiteralFromStringValue(s string) *Literal {
//	return &Literal{value: s}
//}
