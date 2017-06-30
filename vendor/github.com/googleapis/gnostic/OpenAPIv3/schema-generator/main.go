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

// schema-generator is a support tool that generates the OpenAPI v3 JSON schema.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/googleapis/gnostic/jsonschema"
)

// convert the first character of a string to lower case
func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

// model a section of the OpenAPI specification text document
type Section struct {
	Level    int
	Text     string
	Title    string
	Children []*Section
}

// read a section of the OpenAPI Specification, recursively dividing it into subsections
func ReadSection(text string, level int) (section *Section) {
	titlePattern := regexp.MustCompile("^" + strings.Repeat("#", level) + " .*$")
	subtitlePattern := regexp.MustCompile("^" + strings.Repeat("#", level+1) + " .*$")

	section = &Section{Level: level, Text: text}
	lines := strings.Split(string(text), "\n")
	subsection := ""
	for i, line := range lines {
		if i == 0 && titlePattern.Match([]byte(line)) {
			section.Title = line
		} else if subtitlePattern.Match([]byte(line)) {
			// we've found a subsection title.
			// if there's a subsection that we've already been reading, save it
			if len(subsection) != 0 {
				child := ReadSection(subsection, level+1)
				section.Children = append(section.Children, child)
			}
			// start a new subsection
			subsection = line + "\n"
		} else {
			// add to the subsection we've been reading
			subsection += line + "\n"
		}
	}
	// if this section has subsections, save the last one
	if len(section.Children) > 0 {
		child := ReadSection(subsection, level+1)
		section.Children = append(section.Children, child)
	}
	return
}

// recursively display a section of the specification
func (s *Section) Display(section string) {
	if len(s.Children) == 0 {
		//fmt.Printf("%s\n", s.Text)
	} else {
		for i, child := range s.Children {
			var subsection string
			if section == "" {
				subsection = fmt.Sprintf("%d", i)
			} else {
				subsection = fmt.Sprintf("%s.%d", section, i)
			}
			fmt.Printf("%-12s %s\n", subsection, child.NiceTitle())
			child.Display(subsection)
		}
	}
}

// remove a link from a string, leaving only the text that follows it
// if there is no link, just return the string
func stripLink(input string) (output string) {
	stringPattern := regexp.MustCompile("^(.*)$")
	stringWithLinkPattern := regexp.MustCompile("^<a .*</a>(.*)$")
	if matches := stringWithLinkPattern.FindSubmatch([]byte(input)); matches != nil {
		return string(matches[1])
	} else if matches := stringPattern.FindSubmatch([]byte(input)); matches != nil {
		return string(matches[1])
	} else {
		return input
	}
}

// return a nice-to-display title for a section by removing the opening "###" and any links
func (s *Section) NiceTitle() string {
	titlePattern := regexp.MustCompile("^#+ (.*)$")
	titleWithLinkPattern := regexp.MustCompile("^#+ <a .*</a>(.*)$")
	if matches := titleWithLinkPattern.FindSubmatch([]byte(s.Title)); matches != nil {
		return string(matches[1])
	} else if matches := titlePattern.FindSubmatch([]byte(s.Title)); matches != nil {
		return string(matches[1])
	} else {
		return ""
	}
}

// replace markdown links with their link text (removing the URL part)
func removeMarkdownLinks(input string) (output string) {
	markdownLink := regexp.MustCompile("\\[([^\\]]*)\\]\\(([^\\)]*)\\)") // matches [link title](link url)
	output = string(markdownLink.ReplaceAll([]byte(input), []byte("$1")))
	return
}

// extract the fixed fields from a table in a section
func parseFixedFields(input string, schemaObject *SchemaObject) {
	lines := strings.Split(input, "\n")
	for _, line := range lines {

		line = strings.Replace(line, " \\| ", " OR ", -1)

		parts := strings.Split(line, "|")
		if len(parts) > 1 {
			fieldName := strings.Trim(stripLink(parts[0]), " ")
			if fieldName != "Field Name" && fieldName != "---" {

				if len(parts) == 3 || len(parts) == 4 {
					// this is what we expect
				} else {
					log.Printf("ERROR: %+v", parts)
				}

				typeName := parts[1]
				typeName = strings.Trim(typeName, " ")
				typeName = strings.Replace(typeName, "`", "", -1)
				typeName = removeMarkdownLinks(typeName)
				typeName = strings.Replace(typeName, " ", "", -1)
				typeName = strings.Replace(typeName, "Object", "", -1)
				typeName = strings.Replace(typeName, "{expression}", "Expression", -1)
				isArray := false
				if typeName[0] == '[' && typeName[len(typeName)-1] == ']' {
					typeName = typeName[1 : len(typeName)-1]
					isArray = true
				}
				isMap := false
				mapPattern := regexp.MustCompile("^Mapstring,\\[(.*)\\]$")
				if matches := mapPattern.FindSubmatch([]byte(typeName)); matches != nil {
					typeName = string(matches[1])
					isMap = true
				}
				description := strings.Trim(parts[len(parts)-1], " ")
				description = removeMarkdownLinks(description)
				description = strings.Replace(description, "\n", " ", -1)

				requiredLabel := "**Required.** "
				if strings.Contains(description, requiredLabel) {
					// only include required values if their "Validity" is "Any" or if no validity is specified
					valid := true
					if len(parts) == 4 {
						validity := parts[2]
						if strings.Contains(validity, "Any") {
							valid = true
						} else {
							valid = false
						}
					}
					if valid {
						schemaObject.RequiredFields = append(schemaObject.RequiredFields, fieldName)
					}
					description = strings.Replace(description, requiredLabel, "", -1)
				}
				schemaField := SchemaObjectField{
					Name:        fieldName,
					Type:        typeName,
					IsArray:     isArray,
					IsMap:       isMap,
					Description: description,
				}
				schemaObject.FixedFields = append(schemaObject.FixedFields, schemaField)
			}
		}
	}
}

// extract the patterned fields from a table in a section
func parsePatternedFields(input string, schemaObject *SchemaObject) {
	lines := strings.Split(input, "\n")
	for _, line := range lines {

		line = strings.Replace(line, " \\| ", " OR ", -1)

		parts := strings.Split(line, "|")
		if len(parts) > 1 {
			fieldName := strings.Trim(stripLink(parts[0]), " ")
			fieldName = removeMarkdownLinks(fieldName)
			if fieldName == "HTTP Status Code" {
				fieldName = "^([0-9]{3})$"
			}
			if fieldName != "Field Pattern" && fieldName != "---" {
				typeName := parts[1]
				typeName = strings.Trim(typeName, " ")
				typeName = strings.Replace(typeName, "`", "", -1)
				typeName = removeMarkdownLinks(typeName)
				typeName = strings.Replace(typeName, " ", "", -1)
				typeName = strings.Replace(typeName, "Object", "", -1)
				typeName = strings.Replace(typeName, "{expression}", "Expression", -1)
				isArray := false
				if typeName[0] == '[' && typeName[len(typeName)-1] == ']' {
					typeName = typeName[1 : len(typeName)-1]
					isArray = true
				}
				isMap := false
				mapPattern := regexp.MustCompile("^Mapstring,\\[(.*)\\]$")
				if matches := mapPattern.FindSubmatch([]byte(typeName)); matches != nil {
					typeName = string(matches[1])
					isMap = true
				}
				description := strings.Trim(parts[len(parts)-1], " ")
				description = removeMarkdownLinks(description)
				description = strings.Replace(description, "\n", " ", -1)

				schemaField := SchemaObjectField{
					Name:        fieldName,
					Type:        typeName,
					IsArray:     isArray,
					IsMap:       isMap,
					Description: description,
				}
				schemaObject.PatternedFields = append(schemaObject.PatternedFields, schemaField)
			}
		}
	}
}

type SchemaObjectField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	IsArray     bool   `json:"is_array"`
	IsMap       bool   `json:"is_map"`
	Description string `json:"description"`
}

type SchemaObject struct {
	Name            string              `json:"name"`
	Id              string              `json:"id"`
	Description     string              `json:"description"`
	Extendable      bool                `json:"extendable"`
	RequiredFields  []string            `json:"required"`
	FixedFields     []SchemaObjectField `json:"fixed"`
	PatternedFields []SchemaObjectField `json:"patterned"`
}

type SchemaModel struct {
	Objects []SchemaObject
}

func (m *SchemaModel) objectWithId(id string) *SchemaObject {
	for _, object := range m.Objects {
		if object.Id == id {
			return &object
		}
	}
	return nil
}

func NewSchemaModel(filename string) (schemaModel *SchemaModel, err error) {

	b, err := ioutil.ReadFile("3.0.md")
	if err != nil {
		return nil, err
	}

	// divide the specification into sections
	document := ReadSection(string(b), 1)
	//document.Display("")

	// read object names and their details
	specification := document.Children[4] // fragile!
	schema := specification.Children[5]   // fragile!
	anchor := regexp.MustCompile("^#### <a name=\"(.*)Object\"")
	schemaObjects := make([]SchemaObject, 0)
	for _, section := range schema.Children {
		if matches := anchor.FindSubmatch([]byte(section.Title)); matches != nil {

			id := string(matches[1])

			schemaObject := SchemaObject{
				Name:           section.NiceTitle(),
				Id:             id,
				RequiredFields: nil,
			}

			if len(section.Children) > 0 {
				description := section.Children[0].Text
				description = removeMarkdownLinks(description)
				description = strings.Trim(description, " \t\n")
				description = strings.Replace(description, "\n", " ", -1)
				schemaObject.Description = description
			}

			// is the object extendable?
			if strings.Contains(section.Text, "Specification Extensions") {
				schemaObject.Extendable = true
			}

			// look for fixed fields
			for _, child := range section.Children {
				if child.NiceTitle() == "Fixed Fields" {
					parseFixedFields(child.Text, &schemaObject)
				}
			}

			// look for patterned fields
			for _, child := range section.Children {
				if child.NiceTitle() == "Patterned Fields" {
					parsePatternedFields(child.Text, &schemaObject)
				}
			}

			schemaObjects = append(schemaObjects, schemaObject)
		}
	}

	return &SchemaModel{Objects: schemaObjects}, nil
}

type UnionType struct {
	Name        string
	ObjectType1 string
	ObjectType2 string
}

var unionTypes map[string]*UnionType

func noteUnionType(typeName, objectType1, objectType2 string) {
	if unionTypes == nil {
		unionTypes = make(map[string]*UnionType, 0)
	}
	unionTypes[typeName] = &UnionType{
		Name:        typeName,
		ObjectType1: objectType1,
		ObjectType2: objectType2,
	}
}

type MapType struct {
	Name       string
	ObjectType string
}

var mapTypes map[string]*MapType

func noteMapType(typeName, objectType string) {
	if mapTypes == nil {
		mapTypes = make(map[string]*MapType, 0)
	}
	mapTypes[typeName] = &MapType{
		Name:       typeName,
		ObjectType: objectType,
	}
}

func definitionNameForType(typeName string) string {
	name := typeName
	switch typeName {
	case "OAuthFlows":
		name = "oauthFlows"
	case "OAuthFlow":
		name = "oauthFlow"
	case "XML":
		name = "xml"
	case "ExternalDocumentation":
		name = "externalDocs"
	default:
		// does the name contain an "OR"
		if parts := strings.Split(typeName, "OR"); len(parts) > 1 {
			name = lowerFirst(parts[0]) + "Or" + parts[1]
			noteUnionType(name, parts[0], parts[1])
		} else {
			name = lowerFirst(typeName)
		}
	}
	return "#/definitions/" + name
}

func definitionNameForMapOfType(typeName string) string {
	// pluralize the type name to get the name of an object representing a map of them
	name := lowerFirst(typeName)
	if name[len(name)-1] == 'y' {
		name = name[0:len(name)-1] + "ies"
	} else {
		name = name + "s"
	}
	noteMapType(name, typeName)
	return "#/definitions/" + name
}

func updateSchemaFieldWithModelField(schemaField *jsonschema.Schema, modelField *SchemaObjectField) {
	// fmt.Printf("IN %s:%+v\n", name, schemaField)
	// update the attributes of the schema field
	if modelField.IsArray {
		// is array
		itemSchema := &jsonschema.Schema{}
		switch modelField.Type {
		case "string":
			itemSchema.Type = jsonschema.NewStringOrStringArrayWithString("string")
		case "boolean":
			itemSchema.Type = jsonschema.NewStringOrStringArrayWithString("boolean")
		case "primitive":
			itemSchema.Ref = stringptr(definitionNameForType("Primitive"))
		default:
			itemSchema.Ref = stringptr(definitionNameForType(modelField.Type))
		}
		schemaField.Items = jsonschema.NewSchemaOrSchemaArrayWithSchema(itemSchema)
		schemaField.Type = jsonschema.NewStringOrStringArrayWithString("array")
		boolValue := true // not sure about this
		schemaField.UniqueItems = &boolValue
	} else if modelField.IsMap {
		schemaField.Ref = stringptr(definitionNameForMapOfType(modelField.Type))
	} else {
		// is scalar
		switch modelField.Type {
		case "string":
			schemaField.Type = jsonschema.NewStringOrStringArrayWithString("string")
		case "boolean":
			schemaField.Type = jsonschema.NewStringOrStringArrayWithString("boolean")
		case "primitive":
			schemaField.Ref = stringptr(definitionNameForType("Primitive"))
		default:
			schemaField.Ref = stringptr(definitionNameForType(modelField.Type))
		}
	}
}

func buildSchemaWithModel(modelObject *SchemaObject) (schema *jsonschema.Schema) {

	schema = &jsonschema.Schema{}
	schema.Type = jsonschema.NewStringOrStringArrayWithString("object")

	if modelObject.RequiredFields != nil && len(modelObject.RequiredFields) > 0 {
		// copy array
		arrayCopy := modelObject.RequiredFields
		schema.Required = &arrayCopy
	}

	schema.Description = stringptr(modelObject.Description)

	// handle fixed fields
	if modelObject.FixedFields != nil {
		newNamedSchemas := make([]*jsonschema.NamedSchema, 0)
		for _, modelField := range modelObject.FixedFields {
			schemaField := schema.PropertyWithName(modelField.Name)
			if schemaField == nil {
				// create and add the schema field
				schemaField = &jsonschema.Schema{}
				namedSchema := &jsonschema.NamedSchema{Name: modelField.Name, Value: schemaField}
				newNamedSchemas = append(newNamedSchemas, namedSchema)
			}
			updateSchemaFieldWithModelField(schemaField, &modelField)
		}
		for _, pair := range newNamedSchemas {
			if schema.Properties == nil {
				properties := make([]*jsonschema.NamedSchema, 0)
				schema.Properties = &properties
			}
			*(schema.Properties) = append(*(schema.Properties), pair)
		}

	} else {
		if schema.Properties != nil {
			fmt.Printf("SCHEMA SHOULD NOT HAVE PROPERTIES %s\n", modelObject.Id)
		}
	}

	// handle patterned fields
	if modelObject.PatternedFields != nil {
		newNamedSchemas := make([]*jsonschema.NamedSchema, 0)

		for _, modelField := range modelObject.PatternedFields {
			schemaField := schema.PatternPropertyWithName(modelField.Name)
			if schemaField == nil {
				// create and add the schema field
				schemaField = &jsonschema.Schema{}
				namedSchema := &jsonschema.NamedSchema{Name: modelField.Name, Value: schemaField}
				newNamedSchemas = append(newNamedSchemas, namedSchema)
			}
			updateSchemaFieldWithModelField(schemaField, &modelField)
		}

		for _, pair := range newNamedSchemas {
			if schema.PatternProperties == nil {
				properties := make([]*jsonschema.NamedSchema, 0)
				schema.PatternProperties = &properties
			}
			*(schema.PatternProperties) = append(*(schema.PatternProperties), pair)
		}

	} else {
		if schema.PatternProperties != nil && !modelObject.Extendable {
			fmt.Printf("SCHEMA SHOULD NOT HAVE PATTERN PROPERTIES %s\n", modelObject.Id)
		}
	}

	if modelObject.Extendable {
		schemaField := schema.PatternPropertyWithName("^x-")
		if schemaField != nil {
			schemaField.Ref = stringptr("#/definitions/specificationExtension")
		} else {
			schemaField = &jsonschema.Schema{}
			schemaField.Ref = stringptr("#/definitions/specificationExtension")
			namedSchema := &jsonschema.NamedSchema{Name: "^x-", Value: schemaField}
			if schema.PatternProperties == nil {
				properties := make([]*jsonschema.NamedSchema, 0)
				schema.PatternProperties = &properties
			}
			*(schema.PatternProperties) = append(*(schema.PatternProperties), namedSchema)
		}
	} else {
		schemaField := schema.PatternPropertyWithName("^x-")
		if schemaField != nil {
			fmt.Printf("INVALID EXTENSION SUPPORT %s:%s\n", modelObject.Id, "^x-")
		}
	}

	return schema
}

// return a pointer to a copy of a passed-in string
func stringptr(input string) (output *string) {
	return &input
}

func int64ptr(input int64) (output *int64) {
	return &input
}

func arrayOfSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:     jsonschema.NewStringOrStringArrayWithString("array"),
		MinItems: int64ptr(1),
		Items:    jsonschema.NewSchemaOrSchemaArrayWithSchema(&jsonschema.Schema{Ref: stringptr("#/definitions/schemaOrReference")}),
	}
}

func main() {
	// read and parse the text specification into a model structure
	model, err := NewSchemaModel("3.0.md")
	if err != nil {
		panic(err)
	}

	// write the model as JSON (for debugging)
	modelJSON, _ := json.MarshalIndent(model, "", "  ")
	err = ioutil.WriteFile("model.json", modelJSON, 0644)
	if err != nil {
		panic(err)
	}

	// build the top-level schema using the "OAS" model
	schema := buildSchemaWithModel(model.objectWithId("oas"))

	// manually set a few fields
	schema.Title = stringptr("A JSON Schema for OpenAPI 3.0.")
	schema.Id = stringptr("http://openapis.org/v3/schema.json#")
	schema.Schema = stringptr("http://json-schema.org/draft-04/schema#")

	// loop over all models and create the corresponding schema objects
	definitions := make([]*jsonschema.NamedSchema, 0)
	schema.Definitions = &definitions

	for _, modelObject := range model.Objects {
		if modelObject.Id == "oas" {
			continue
		}
		definitionSchema := buildSchemaWithModel(&modelObject)
		name := modelObject.Id
		if name == "externalDocumentation" {
			name = "externalDocs"
		}
		*schema.Definitions = append(*schema.Definitions, jsonschema.NewNamedSchema(name, definitionSchema))
	}

	// copy the properties of headerObject from parameterObject
	headerObject := schema.DefinitionWithName("header")
	parameterObject := schema.DefinitionWithName("parameter")
	if parameterObject != nil {
		// "So a shorthand for copying array arr would be tmp := append([]int{}, arr...)"
		newArray := make([]*jsonschema.NamedSchema, 0)
		newArray = append(newArray, *(parameterObject.Properties)...)
		headerObject.Properties = &newArray
		// we need to remove a few properties...
	}

	// generate implied union types
	unionTypeKeys := make([]string, 0, len(unionTypes))
	for key := range unionTypes {
		unionTypeKeys = append(unionTypeKeys, key)
	}
	sort.Strings(unionTypeKeys)
	for _, unionTypeKey := range unionTypeKeys {
		unionType := unionTypes[unionTypeKey]
		objectSchema := schema.DefinitionWithName(unionType.Name)
		if objectSchema == nil {
			objectSchema = &jsonschema.Schema{}
			oneOf := make([]*jsonschema.Schema, 0)
			oneOf = append(oneOf, &jsonschema.Schema{Ref: stringptr("#/definitions/" + lowerFirst(unionType.ObjectType1))})
			oneOf = append(oneOf, &jsonschema.Schema{Ref: stringptr("#/definitions/" + lowerFirst(unionType.ObjectType2))})
			objectSchema.OneOf = &oneOf
			*schema.Definitions = append(*schema.Definitions, jsonschema.NewNamedSchema(unionType.Name, objectSchema))
		}
	}

	// generate implied map types
	mapTypeKeys := make([]string, 0, len(mapTypes))
	for key := range mapTypes {
		mapTypeKeys = append(mapTypeKeys, key)
	}
	sort.Strings(mapTypeKeys)
	for _, mapTypeKey := range mapTypeKeys {
		mapType := mapTypes[mapTypeKey]
		objectSchema := schema.DefinitionWithName(mapType.Name)
		if objectSchema == nil {
			objectSchema = &jsonschema.Schema{}
			objectSchema.Type = jsonschema.NewStringOrStringArrayWithString("object")
			additionalPropertiesSchema := &jsonschema.Schema{}
			additionalPropertiesSchema.Ref = stringptr("#/definitions/" + lowerFirst(mapType.ObjectType))
			objectSchema.AdditionalProperties = jsonschema.NewSchemaOrBooleanWithSchema(additionalPropertiesSchema)
			*schema.Definitions = append(*schema.Definitions, jsonschema.NewNamedSchema(mapType.Name, objectSchema))
		}
	}

	// add schema objects for "object", "any", and "expression"
	if true {
		objectSchema := &jsonschema.Schema{}
		objectSchema.Type = jsonschema.NewStringOrStringArrayWithString("object")
		objectSchema.AdditionalProperties = jsonschema.NewSchemaOrBooleanWithBoolean(true)
		objectSchema.AdditionalItems = jsonschema.NewSchemaOrBooleanWithBoolean(true)
		*schema.Definitions = append(*schema.Definitions, jsonschema.NewNamedSchema("object", objectSchema))
	}
	if true {
		objectSchema := &jsonschema.Schema{}
		objectSchema.AdditionalProperties = jsonschema.NewSchemaOrBooleanWithBoolean(true)
		objectSchema.AdditionalItems = jsonschema.NewSchemaOrBooleanWithBoolean(true)
		*schema.Definitions = append(*schema.Definitions, jsonschema.NewNamedSchema("any", objectSchema))
	}
	if true {
		objectSchema := &jsonschema.Schema{}
		objectSchema.Type = jsonschema.NewStringOrStringArrayWithString("object")
		objectSchema.AdditionalProperties = jsonschema.NewSchemaOrBooleanWithBoolean(true)
		objectSchema.AdditionalItems = jsonschema.NewSchemaOrBooleanWithBoolean(true)
		*schema.Definitions = append(*schema.Definitions, jsonschema.NewNamedSchema("expression", objectSchema))
	}

	// add schema objects for "specificationExtension"
	if true {
		objectSchema := &jsonschema.Schema{}
		objectSchema.Description = stringptr("Any property starting with x- is valid.")
		oneOf := make([]*jsonschema.Schema, 0)
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("integer")})
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("number")})
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("boolean")})
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("string")})
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("object")})
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("array")})
		objectSchema.OneOf = &oneOf
		*schema.Definitions = append(*schema.Definitions, jsonschema.NewNamedSchema("specificationExtension", objectSchema))
	}

	// add schema objects for "primitive"
	if true {
		objectSchema := &jsonschema.Schema{}
		oneOf := make([]*jsonschema.Schema, 0)
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("integer")})
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("number")})
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("boolean")})
		oneOf = append(oneOf, &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("string")})
		objectSchema.OneOf = &oneOf
		*schema.Definitions = append(*schema.Definitions, jsonschema.NewNamedSchema("primitive", objectSchema))
	}

	// force a few more things into the "schema" schema
	schemaObject := schema.DefinitionWithName("schema")
	schemaObject.CopyOfficialSchemaProperties(
		[]string{
			"title",
			"multipleOf",
			"maximum",
			"exclusiveMaximum",
			"minimum",
			"exclusiveMinimum",
			"maxLength",
			"minLength",
			"pattern",
			"maxItems",
			"minItems",
			"uniqueItems",
			"maxProperties",
			"minProperties",
			"required",
			"enum",
		})
	schemaObject.AddProperty("type", &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("string")})
	schemaObject.AddProperty("allOf", arrayOfSchema())
	schemaObject.AddProperty("oneOf", arrayOfSchema())
	schemaObject.AddProperty("anyOf", arrayOfSchema())
	schemaObject.AddProperty("not", &jsonschema.Schema{Ref: stringptr("#/definitions/schema")})
	anyOf := make([]*jsonschema.Schema, 0)
	anyOf = append(anyOf, &jsonschema.Schema{Ref: stringptr("#/definitions/schemaOrReference")})
	anyOf = append(anyOf, arrayOfSchema())
	schemaObject.AddProperty("items",
		&jsonschema.Schema{AnyOf: &anyOf})
	schemaObject.AddProperty("properties", &jsonschema.Schema{
		Type: jsonschema.NewStringOrStringArrayWithString("object"),
		AdditionalProperties: jsonschema.NewSchemaOrBooleanWithSchema(
			&jsonschema.Schema{Ref: stringptr("#/definitions/schema")})})
	schemaObject.AddProperty("description", &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("string")})
	schemaObject.AddProperty("format", &jsonschema.Schema{Type: jsonschema.NewStringOrStringArrayWithString("string")})

	// fix the content object
	contentObject := schema.DefinitionWithName("content")
	pairs := make([]*jsonschema.NamedSchema, 0)
	contentObject.PatternProperties = &pairs
	namedSchema := &jsonschema.NamedSchema{Name: "{media-type}", Value: &jsonschema.Schema{Ref: stringptr("#/definitions/mediaType")}}
	*(contentObject.PatternProperties) = append(*(contentObject.PatternProperties), namedSchema)

	// write the updated schema
	output := schema.JSONString()
	err = ioutil.WriteFile("schema.json", []byte(output), 0644)
	if err != nil {
		panic(err)
	}
}
