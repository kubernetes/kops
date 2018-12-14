// Copyright 2015 go-swagger maintainers
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

package analysis

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/assert"
)

func schemeNames(schemes [][]SecurityRequirement) []string {
	var names []string
	for _, scheme := range schemes {
		for _, v := range scheme {
			names = append(names, v.Name)
		}
	}
	sort.Sort(sort.StringSlice(names))
	return names
}

func TestAnalyzer(t *testing.T) {
	formatParam := spec.QueryParam("format").Typed("string", "")

	limitParam := spec.QueryParam("limit").Typed("integer", "int32")
	limitParam.Extensions = spec.Extensions(map[string]interface{}{})
	limitParam.Extensions.Add("go-name", "Limit")

	skipParam := spec.QueryParam("skip").Typed("integer", "int32")
	pi := spec.PathItem{}
	pi.Parameters = []spec.Parameter{*limitParam}

	op := &spec.Operation{}
	op.Consumes = []string{"application/x-yaml"}
	op.Produces = []string{"application/x-yaml"}
	op.Security = []map[string][]string{
		map[string][]string{"oauth2": []string{}},
		map[string][]string{"basic": nil},
	}
	op.ID = "someOperation"
	op.Parameters = []spec.Parameter{*skipParam}
	pi.Get = op

	pi2 := spec.PathItem{}
	pi2.Parameters = []spec.Parameter{*limitParam}
	op2 := &spec.Operation{}
	op2.ID = "anotherOperation"
	op2.Parameters = []spec.Parameter{*skipParam}
	pi2.Get = op2

	spec := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
			Security: []map[string][]string{
				map[string][]string{"apikey": nil},
			},
			SecurityDefinitions: map[string]*spec.SecurityScheme{
				"basic":  spec.BasicAuth(),
				"apiKey": spec.APIKeyAuth("api_key", "query"),
				"oauth2": spec.OAuth2AccessToken("http://authorize.com", "http://token.com"),
			},
			Parameters: map[string]spec.Parameter{"format": *formatParam},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/":      pi,
					"/items": pi2,
				},
			},
		},
	}
	analyzer := New(spec)

	assert.Len(t, analyzer.consumes, 2)
	assert.Len(t, analyzer.produces, 2)
	assert.Len(t, analyzer.operations, 1)
	assert.Equal(t, analyzer.operations["GET"]["/"], spec.Paths.Paths["/"].Get)

	expected := []string{"application/x-yaml"}
	sort.Sort(sort.StringSlice(expected))
	consumes := analyzer.ConsumesFor(spec.Paths.Paths["/"].Get)
	sort.Sort(sort.StringSlice(consumes))
	assert.Equal(t, expected, consumes)

	produces := analyzer.ProducesFor(spec.Paths.Paths["/"].Get)
	sort.Sort(sort.StringSlice(produces))
	assert.Equal(t, expected, produces)

	expected = []string{"application/json"}
	sort.Sort(sort.StringSlice(expected))
	consumes = analyzer.ConsumesFor(spec.Paths.Paths["/items"].Get)
	sort.Sort(sort.StringSlice(consumes))
	assert.Equal(t, expected, consumes)

	produces = analyzer.ProducesFor(spec.Paths.Paths["/items"].Get)
	sort.Sort(sort.StringSlice(produces))
	assert.Equal(t, expected, produces)

	expectedSchemes := [][]SecurityRequirement{
		[]SecurityRequirement{SecurityRequirement{"oauth2", []string{}}, SecurityRequirement{"basic", nil}},
	}
	schemes := analyzer.SecurityRequirementsFor(spec.Paths.Paths["/"].Get)
	assert.Equal(t, schemeNames(expectedSchemes), schemeNames(schemes))

	securityDefinitions := analyzer.SecurityDefinitionsFor(spec.Paths.Paths["/"].Get)
	assert.Equal(t, *spec.SecurityDefinitions["basic"], securityDefinitions["basic"])
	assert.Equal(t, *spec.SecurityDefinitions["oauth2"], securityDefinitions["oauth2"])

	parameters := analyzer.ParamsFor("GET", "/")
	assert.Len(t, parameters, 2)

	operations := analyzer.OperationIDs()
	assert.Len(t, operations, 2)

	producers := analyzer.RequiredProduces()
	assert.Len(t, producers, 2)
	consumers := analyzer.RequiredConsumes()
	assert.Len(t, consumers, 2)
	authSchemes := analyzer.RequiredSecuritySchemes()
	assert.Len(t, authSchemes, 3)

	ops := analyzer.Operations()
	assert.Len(t, ops, 1)
	assert.Len(t, ops["GET"], 2)

	op, ok := analyzer.OperationFor("get", "/")
	assert.True(t, ok)
	assert.NotNil(t, op)

	op, ok = analyzer.OperationFor("delete", "/")
	assert.False(t, ok)
	assert.Nil(t, op)
}

func TestDefinitionAnalysis(t *testing.T) {
	doc, err := loadSpec(filepath.Join("fixtures", "definitions.yml"))
	if assert.NoError(t, err) {
		analyzer := New(doc)
		definitions := analyzer.allSchemas
		// parameters
		assertSchemaRefExists(t, definitions, "#/parameters/someParam/schema")
		assertSchemaRefExists(t, definitions, "#/paths/~1some~1where~1{id}/parameters/1/schema")
		assertSchemaRefExists(t, definitions, "#/paths/~1some~1where~1{id}/get/parameters/1/schema")
		// responses
		assertSchemaRefExists(t, definitions, "#/responses/someResponse/schema")
		assertSchemaRefExists(t, definitions, "#/paths/~1some~1where~1{id}/get/responses/default/schema")
		assertSchemaRefExists(t, definitions, "#/paths/~1some~1where~1{id}/get/responses/200/schema")
		// definitions
		assertSchemaRefExists(t, definitions, "#/definitions/tag")
		assertSchemaRefExists(t, definitions, "#/definitions/tag/properties/id")
		assertSchemaRefExists(t, definitions, "#/definitions/tag/properties/value")
		assertSchemaRefExists(t, definitions, "#/definitions/tag/definitions/category")
		assertSchemaRefExists(t, definitions, "#/definitions/tag/definitions/category/properties/id")
		assertSchemaRefExists(t, definitions, "#/definitions/tag/definitions/category/properties/value")
		assertSchemaRefExists(t, definitions, "#/definitions/withAdditionalProps")
		assertSchemaRefExists(t, definitions, "#/definitions/withAdditionalProps/additionalProperties")
		assertSchemaRefExists(t, definitions, "#/definitions/withAdditionalItems")
		assertSchemaRefExists(t, definitions, "#/definitions/withAdditionalItems/items/0")
		assertSchemaRefExists(t, definitions, "#/definitions/withAdditionalItems/items/1")
		assertSchemaRefExists(t, definitions, "#/definitions/withAdditionalItems/additionalItems")
		assertSchemaRefExists(t, definitions, "#/definitions/withNot")
		assertSchemaRefExists(t, definitions, "#/definitions/withNot/not")
		assertSchemaRefExists(t, definitions, "#/definitions/withAnyOf")
		assertSchemaRefExists(t, definitions, "#/definitions/withAnyOf/anyOf/0")
		assertSchemaRefExists(t, definitions, "#/definitions/withAnyOf/anyOf/1")
		assertSchemaRefExists(t, definitions, "#/definitions/withAllOf")
		assertSchemaRefExists(t, definitions, "#/definitions/withAllOf/allOf/0")
		assertSchemaRefExists(t, definitions, "#/definitions/withAllOf/allOf/1")
		allOfs := analyzer.allOfs
		assert.Len(t, allOfs, 1)
		assert.Contains(t, allOfs, "#/definitions/withAllOf")
	}
}

func loadSpec(path string) (*spec.Swagger, error) {
	spec.PathLoader = func(path string) (json.RawMessage, error) {
		ext := filepath.Ext(path)
		if ext == ".yml" || ext == ".yaml" {
			return fmts.YAMLDoc(path)
		}
		data, err := swag.LoadFromFileOrHTTP(path)
		if err != nil {
			return nil, err
		}
		return json.RawMessage(data), nil
	}
	data, err := fmts.YAMLDoc(path)
	if err != nil {
		return nil, err
	}

	var sw spec.Swagger
	if err := json.Unmarshal(data, &sw); err != nil {
		return nil, err
	}
	return &sw, nil
}

func TestReferenceAnalysis(t *testing.T) {
	doc, err := loadSpec(filepath.Join("fixtures", "references.yml"))
	if assert.NoError(t, err) {
		definitions := New(doc).references

		// parameters
		assertRefExists(t, definitions.parameters, "#/paths/~1some~1where~1{id}/parameters/0")
		assertRefExists(t, definitions.parameters, "#/paths/~1some~1where~1{id}/get/parameters/0")

		// path items
		assertRefExists(t, definitions.pathItems, "#/paths/~1other~1place")

		// responses
		assertRefExists(t, definitions.responses, "#/paths/~1some~1where~1{id}/get/responses/404")

		// definitions
		assertRefExists(t, definitions.schemas, "#/responses/notFound/schema")
		assertRefExists(t, definitions.schemas, "#/paths/~1some~1where~1{id}/get/responses/200/schema")
		assertRefExists(t, definitions.schemas, "#/definitions/tag/properties/audit")

		// items
		assertRefExists(t, definitions.allRefs, "#/paths/~1some~1where~1{id}/get/parameters/1/items")
	}
}

func assertRefExists(t testing.TB, data map[string]spec.Ref, key string) bool {
	if _, ok := data[key]; !ok {
		return assert.Fail(t, fmt.Sprintf("expected %q to exist in the ref bag", key))
	}
	return true
}

func assertSchemaRefExists(t testing.TB, data map[string]SchemaRef, key string) bool {
	if _, ok := data[key]; !ok {
		return assert.Fail(t, fmt.Sprintf("expected %q to exist in schema ref bag", key))
	}
	return true
}

func TestPatternAnalysis(t *testing.T) {
	doc, err := loadSpec(filepath.Join("fixtures", "patterns.yml"))
	if assert.NoError(t, err) {
		pt := New(doc).patterns

		// parameters
		assertPattern(t, pt.parameters, "#/parameters/idParam", "a[A-Za-Z0-9]+")
		assertPattern(t, pt.parameters, "#/paths/~1some~1where~1{id}/parameters/1", "b[A-Za-z0-9]+")
		assertPattern(t, pt.parameters, "#/paths/~1some~1where~1{id}/get/parameters/0", "[abc][0-9]+")

		// responses
		assertPattern(t, pt.headers, "#/responses/notFound/headers/ContentLength", "[0-9]+")
		assertPattern(t, pt.headers, "#/paths/~1some~1where~1{id}/get/responses/200/headers/X-Request-Id", "d[A-Za-z0-9]+")

		// definitions
		assertPattern(t, pt.schemas, "#/paths/~1other~1place/post/parameters/0/schema/properties/value", "e[A-Za-z0-9]+")
		assertPattern(t, pt.schemas, "#/paths/~1other~1place/post/responses/200/schema/properties/data", "[0-9]+[abd]")
		assertPattern(t, pt.schemas, "#/definitions/named", "f[A-Za-z0-9]+")
		assertPattern(t, pt.schemas, "#/definitions/tag/properties/value", "g[A-Za-z0-9]+")

		// items
		assertPattern(t, pt.items, "#/paths/~1some~1where~1{id}/get/parameters/1/items", "c[A-Za-z0-9]+")
		assertPattern(t, pt.items, "#/paths/~1other~1place/post/responses/default/headers/Via/items", "[A-Za-z]+")
	}
}

func assertPattern(t testing.TB, data map[string]string, key, pattern string) bool {
	if assert.Contains(t, data, key) {
		return assert.Equal(t, pattern, data[key])
	}
	return false
}

func panickerParamsAsMap() {
	s := prepareTestParamsInvalid("fixture-342.yaml")
	if s == nil {
		return
	}
	m := make(map[string]spec.Parameter)
	if pi, ok := s.spec.Paths.Paths["/fixture"]; ok {
		pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
		s.paramsAsMap(pi.Parameters, m, nil)
	}
}

func panickerParamsAsMap2() {
	s := prepareTestParamsInvalid("fixture-342-2.yaml")
	if s == nil {
		return
	}
	m := make(map[string]spec.Parameter)
	if pi, ok := s.spec.Paths.Paths["/fixture"]; ok {
		pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
		s.paramsAsMap(pi.Parameters, m, nil)
	}
}

func panickerParamsAsMap3() {
	s := prepareTestParamsInvalid("fixture-342-3.yaml")
	if s == nil {
		return
	}
	m := make(map[string]spec.Parameter)
	if pi, ok := s.spec.Paths.Paths["/fixture"]; ok {
		pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
		s.paramsAsMap(pi.Parameters, m, nil)
	}
}

func TestAnalyzer_paramsAsMap(Pt *testing.T) {
	s := prepareTestParamsValid()
	if assert.NotNil(Pt, s) {
		m := make(map[string]spec.Parameter)
		pi, ok := s.spec.Paths.Paths["/items"]
		if assert.True(Pt, ok) {
			//func (s *Spec) paramsAsMap(parameters []spec.Parameter, res map[string]spec.Parameter, callmeOnError ErrorOnParamFunc) {
			s.paramsAsMap(pi.Parameters, m, nil)
			// TODO: Assert?
		}
	}

	// An invalid spec, but passes this step (errors are figured out at a higher level)
	s = prepareTestParamsInvalid("fixture-1289-param.yaml")
	if assert.NotNil(Pt, s) {
		m := make(map[string]spec.Parameter)
		pi, ok := s.spec.Paths.Paths["/fixture"]
		if assert.True(Pt, ok) {
			pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
			//func (s *Spec) paramsAsMap(parameters []spec.Parameter, res map[string]spec.Parameter, callmeOnError ErrorOnParamFunc) {
			s.paramsAsMap(pi.Parameters, m, nil)
			// TODO: Assert?
		}
	}
}

func TestAnalyzer_paramsAsMapWithCallback(Pt *testing.T) {
	s := prepareTestParamsInvalid("fixture-342.yaml")
	if assert.NotNil(Pt, s) {
		// No bail out callback
		m := make(map[string]spec.Parameter)
		e := []string{}
		pi, ok := s.spec.Paths.Paths["/fixture"]
		if assert.True(Pt, ok) {
			//func (s *Spec) paramsAsMap(parameters []spec.Parameter, res map[string]spec.Parameter, callmeOnError ErrorOnParamFunc) {
			pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
			s.paramsAsMap(pi.Parameters, m, func(param spec.Parameter, err error) bool {
				//Pt.Logf("ERROR on %+v : %v", param, err)
				e = append(e, err.Error())
				return true // Continue
			})
		}
		assert.Contains(Pt, e, `resolved reference is not a parameter: "#/definitions/sample_info/properties/sid"`)
		assert.Contains(Pt, e, `invalid reference: "#/definitions/sample_info/properties/sids"`)

		// bail out callback
		m = make(map[string]spec.Parameter)
		e = []string{}
		pi, ok = s.spec.Paths.Paths["/fixture"]
		if assert.True(Pt, ok) {
			//func (s *Spec) paramsAsMap(parameters []spec.Parameter, res map[string]spec.Parameter, callmeOnError ErrorOnParamFunc) {
			pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
			s.paramsAsMap(pi.Parameters, m, func(param spec.Parameter, err error) bool {
				//Pt.Logf("ERROR on %+v : %v", param, err)
				e = append(e, err.Error())
				return false // Bail out
			})
		}
		// We got one then bail out
		assert.Len(Pt, e, 1)
	}

	// Bail out after ref failure: exercising another path
	s = prepareTestParamsInvalid("fixture-342-2.yaml")
	if assert.NotNil(Pt, s) {
		// bail out callback
		m := make(map[string]spec.Parameter)
		e := []string{}
		pi, ok := s.spec.Paths.Paths["/fixture"]
		if assert.True(Pt, ok) {
			//func (s *Spec) paramsAsMap(parameters []spec.Parameter, res map[string]spec.Parameter, callmeOnError ErrorOnParamFunc) {
			pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
			s.paramsAsMap(pi.Parameters, m, func(param spec.Parameter, err error) bool {
				//Pt.Logf("ERROR on %+v : %v", param, err)
				e = append(e, err.Error())
				return false // Bail out
			})
		}
		// We got one then bail out
		assert.Len(Pt, e, 1)
	}

	// Bail out after ref failure: exercising another path
	s = prepareTestParamsInvalid("fixture-342-3.yaml")
	if assert.NotNil(Pt, s) {
		// bail out callback
		m := make(map[string]spec.Parameter)
		e := []string{}
		pi, ok := s.spec.Paths.Paths["/fixture"]
		if assert.True(Pt, ok) {
			//func (s *Spec) paramsAsMap(parameters []spec.Parameter, res map[string]spec.Parameter, callmeOnError ErrorOnParamFunc) {
			pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
			s.paramsAsMap(pi.Parameters, m, func(param spec.Parameter, err error) bool {
				//Pt.Logf("ERROR on %+v : %v", param, err)
				e = append(e, err.Error())
				return false // Bail out
			})
		}
		// We got one then bail out
		assert.Len(Pt, e, 1)
	}
}

func TestAnalyzer_paramsAsMap_Panic(Pt *testing.T) {
	assert.Panics(Pt, panickerParamsAsMap)

	// Specifically on invalid resolved type
	assert.Panics(Pt, panickerParamsAsMap2)

	// Specifically on invalid ref
	assert.Panics(Pt, panickerParamsAsMap3)
}

func TestAnalyzer_SafeParamsFor(Pt *testing.T) {
	s := prepareTestParamsInvalid("fixture-342.yaml")
	if assert.NotNil(Pt, s) {
		e := []string{}
		pi, ok := s.spec.Paths.Paths["/fixture"]
		if assert.True(Pt, ok) {
			pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
			//func (s *Spec) SafeParamsFor(method, path string, callmeOnError ErrorOnParamFunc) map[string]spec.Parameter {
			for range s.SafeParamsFor("Get", "/fixture", func(param spec.Parameter, err error) bool {
				e = append(e, err.Error())
				return true // Continue
			}) {
				assert.Fail(Pt, "There should be no safe parameter in this testcase")
			}
		}
		assert.Contains(Pt, e, `resolved reference is not a parameter: "#/definitions/sample_info/properties/sid"`)
		assert.Contains(Pt, e, `invalid reference: "#/definitions/sample_info/properties/sids"`)

	}
}

func panickerParamsFor() {
	s := prepareTestParamsInvalid("fixture-342.yaml")
	pi, ok := s.spec.Paths.Paths["/fixture"]
	if ok {
		pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
		//func (s *Spec) ParamsFor(method, path string) map[string]spec.Parameter {
		s.ParamsFor("Get", "/fixture")
	}
}

func TestAnalyzer_ParamsFor(Pt *testing.T) {
	// Valid example
	s := prepareTestParamsValid()
	if assert.NotNil(Pt, s) {

		params := s.ParamsFor("Get", "/items")
		assert.True(Pt, len(params) > 0)
	}

	// Invalid example
	assert.Panics(Pt, panickerParamsFor)
}

func TestAnalyzer_SafeParametersFor(Pt *testing.T) {
	s := prepareTestParamsInvalid("fixture-342.yaml")
	if assert.NotNil(Pt, s) {
		e := []string{}
		pi, ok := s.spec.Paths.Paths["/fixture"]
		if assert.True(Pt, ok) {
			pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
			//func (s *Spec) SafeParametersFor(operationID string, callmeOnError ErrorOnParamFunc) []spec.Parameter {
			for range s.SafeParametersFor("fixtureOp", func(param spec.Parameter, err error) bool {
				e = append(e, err.Error())
				return true // Continue
			}) {
				assert.Fail(Pt, "There should be no safe parameter in this testcase")
			}
		}
		assert.Contains(Pt, e, `resolved reference is not a parameter: "#/definitions/sample_info/properties/sid"`)
		assert.Contains(Pt, e, `invalid reference: "#/definitions/sample_info/properties/sids"`)
	}
}

func panickerParametersFor() {
	s := prepareTestParamsInvalid("fixture-342.yaml")
	if s == nil {
		return
	}
	pi, ok := s.spec.Paths.Paths["/fixture"]
	if ok {
		pi.Parameters = pi.PathItemProps.Get.OperationProps.Parameters
		//func (s *Spec) ParametersFor(operationID string) []spec.Parameter {
		s.ParametersFor("fixtureOp")
	}
}

func TestAnalyzer_ParametersFor(Pt *testing.T) {
	// Valid example
	s := prepareTestParamsValid()
	params := s.ParamsFor("Get", "/items")
	assert.True(Pt, len(params) > 0)

	// Invalid example
	assert.Panics(Pt, panickerParametersFor)
}

func prepareTestParamsValid() *Spec {
	formatParam := spec.QueryParam("format").Typed("string", "")

	limitParam := spec.QueryParam("limit").Typed("integer", "int32")
	limitParam.Extensions = spec.Extensions(map[string]interface{}{})
	limitParam.Extensions.Add("go-name", "Limit")

	skipParam := spec.QueryParam("skip").Typed("integer", "int32")
	pi := spec.PathItem{}
	pi.Parameters = []spec.Parameter{*limitParam}

	op := &spec.Operation{}
	op.Consumes = []string{"application/x-yaml"}
	op.Produces = []string{"application/x-yaml"}
	op.Security = []map[string][]string{
		map[string][]string{"oauth2": []string{}},
		map[string][]string{"basic": nil},
	}
	op.ID = "someOperation"
	op.Parameters = []spec.Parameter{*skipParam}
	pi.Get = op

	pi2 := spec.PathItem{}
	pi2.Parameters = []spec.Parameter{*limitParam}
	op2 := &spec.Operation{}
	op2.ID = "anotherOperation"
	op2.Parameters = []spec.Parameter{*skipParam}
	pi2.Get = op2

	spec := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
			Security: []map[string][]string{
				map[string][]string{"apikey": nil},
			},
			SecurityDefinitions: map[string]*spec.SecurityScheme{
				"basic":  spec.BasicAuth(),
				"apiKey": spec.APIKeyAuth("api_key", "query"),
				"oauth2": spec.OAuth2AccessToken("http://authorize.com", "http://token.com"),
			},
			Parameters: map[string]spec.Parameter{"format": *formatParam},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/":      pi,
					"/items": pi2,
				},
			},
		},
	}
	analyzer := New(spec)
	return analyzer
}

func prepareTestParamsInvalid(fixture string) *Spec {
	cwd, _ := os.Getwd()
	bp := filepath.Join(cwd, "fixtures", fixture)
	spec, err := loadSpec(bp)
	if err != nil {
		log.Printf("Warning: fixture %s could not be loaded: %v", fixture, err)
		return nil
	}
	analyzer := New(spec)
	return analyzer
}

func TestSecurityDefinitionsFor(t *testing.T) {
	spec := prepareTestParamsAuth()
	pi1 := spec.spec.Paths.Paths["/"].Get
	pi2 := spec.spec.Paths.Paths["/items"].Get

	defs1 := spec.SecurityDefinitionsFor(pi1)
	require.Contains(t, defs1, "oauth2")
	require.Contains(t, defs1, "basic")
	require.NotContains(t, defs1, "apiKey")

	defs2 := spec.SecurityDefinitionsFor(pi2)
	require.Contains(t, defs2, "oauth2")
	require.Contains(t, defs2, "basic")
	require.Contains(t, defs2, "apiKey")
}

func TestSecurityRequirements(t *testing.T) {
	spec := prepareTestParamsAuth()
	pi1 := spec.spec.Paths.Paths["/"].Get
	pi2 := spec.spec.Paths.Paths["/items"].Get
	scopes := []string{"the-scope"}

	reqs1 := spec.SecurityRequirementsFor(pi1)
	require.Len(t, reqs1, 2)
	require.Len(t, reqs1[0], 1)
	require.Equal(t, reqs1[0][0].Name, "oauth2")
	require.Equal(t, reqs1[0][0].Scopes, scopes)
	require.Len(t, reqs1[1], 1)
	require.Equal(t, reqs1[1][0].Name, "basic")
	require.Empty(t, reqs1[1][0].Scopes)

	reqs2 := spec.SecurityRequirementsFor(pi2)
	require.Len(t, reqs2, 3)
	require.Len(t, reqs2[0], 1)
	require.Equal(t, reqs2[0][0].Name, "oauth2")
	require.Equal(t, reqs2[0][0].Scopes, scopes)
	require.Len(t, reqs2[1], 1)
	require.Empty(t, reqs2[1][0].Name)
	require.Empty(t, reqs2[1][0].Scopes)
	require.Len(t, reqs2[2], 2)
	require.Equal(t, reqs2[2][0].Name, "basic")
	require.Empty(t, reqs2[2][0].Scopes)
	require.Equal(t, reqs2[2][1].Name, "apiKey")
	require.Empty(t, reqs2[2][1].Scopes)
}

func TestSecurityRequirementsDefinitions(t *testing.T) {
	spec := prepareTestParamsAuth()
	pi1 := spec.spec.Paths.Paths["/"].Get
	pi2 := spec.spec.Paths.Paths["/items"].Get

	reqs1 := spec.SecurityRequirementsFor(pi1)
	defs11 := spec.SecurityDefinitionsForRequirements(reqs1[0])
	require.Contains(t, defs11, "oauth2")
	defs12 := spec.SecurityDefinitionsForRequirements(reqs1[1])
	require.Contains(t, defs12, "basic")
	require.NotContains(t, defs12, "apiKey")

	reqs2 := spec.SecurityRequirementsFor(pi2)
	defs21 := spec.SecurityDefinitionsForRequirements(reqs2[0])
	require.Len(t, defs21, 1)
	require.Contains(t, defs21, "oauth2")
	require.NotContains(t, defs21, "basic")
	require.NotContains(t, defs21, "apiKey")
	defs22 := spec.SecurityDefinitionsForRequirements(reqs2[1])
	require.NotNil(t, defs22)
	require.Empty(t, defs22)
	defs23 := spec.SecurityDefinitionsForRequirements(reqs2[2])
	require.Len(t, defs23, 2)
	require.NotContains(t, defs23, "oauth2")
	require.Contains(t, defs23, "basic")
	require.Contains(t, defs23, "apiKey")

}

func prepareTestParamsAuth() *Spec {
	formatParam := spec.QueryParam("format").Typed("string", "")

	limitParam := spec.QueryParam("limit").Typed("integer", "int32")
	limitParam.Extensions = spec.Extensions(map[string]interface{}{})
	limitParam.Extensions.Add("go-name", "Limit")

	skipParam := spec.QueryParam("skip").Typed("integer", "int32")
	pi := spec.PathItem{}
	pi.Parameters = []spec.Parameter{*limitParam}

	op := &spec.Operation{}
	op.Consumes = []string{"application/x-yaml"}
	op.Produces = []string{"application/x-yaml"}
	op.Security = []map[string][]string{
		map[string][]string{"oauth2": []string{"the-scope"}},
		map[string][]string{"basic": nil},
	}
	op.ID = "someOperation"
	op.Parameters = []spec.Parameter{*skipParam}
	pi.Get = op

	pi2 := spec.PathItem{}
	pi2.Parameters = []spec.Parameter{*limitParam}
	op2 := &spec.Operation{}
	op2.ID = "anotherOperation"
	op2.Security = []map[string][]string{
		map[string][]string{"oauth2": []string{"the-scope"}},
		map[string][]string{},
		map[string][]string{
			"basic":  []string{},
			"apiKey": []string{},
		},
	}
	op2.Parameters = []spec.Parameter{*skipParam}
	pi2.Get = op2

	oauth := spec.OAuth2AccessToken("http://authorize.com", "http://token.com")
	oauth.AddScope("the-scope", "the scope gives access to ...")
	spec := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
			Security: []map[string][]string{
				map[string][]string{"apikey": nil},
			},
			SecurityDefinitions: map[string]*spec.SecurityScheme{
				"basic":  spec.BasicAuth(),
				"apiKey": spec.APIKeyAuth("api_key", "query"),
				"oauth2": oauth,
			},
			Parameters: map[string]spec.Parameter{"format": *formatParam},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/":      pi,
					"/items": pi2,
				},
			},
		},
	}
	analyzer := New(spec)
	return analyzer
}
