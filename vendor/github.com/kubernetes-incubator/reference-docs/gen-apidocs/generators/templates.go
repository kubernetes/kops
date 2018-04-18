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

package generators

var DefinitionTemplate = `
{{define "definition.template"}}##` + "`{{.Name}}` [`{{.GroupDisplayName}}`/`{{.Version}}`]" + `

Group        | Version     | Kind
------------ | ---------- | -----------
` + "`{{.GroupDisplayName}}` | `{{.Version}}` | `{{.Name}}`" + `

{{if .OtherVersions}}<aside class="notice">Other api versions of this object exist: {{range $v := .OtherVersions}}{{$v.VersionLink}} {{end}}</aside>{{end}}

{{.DescriptionWithEntities}}

{{if .AppearsIn}}<aside class="notice">
Appears In:

<ul> {{range $appearsin := .AppearsIn}}
<li>{{$appearsin.FullHrefLink}}</li>{{end}}
</ul></aside>{{end}}

Field        | Description
------------ | -----------
{{range $field := .Fields}}` + "`{{$field.Name}}`" + `{{if $field.Link}}<br /> *{{$field.Link}}* {{end}} {{if $field.PatchStrategy}}<br /> **patch type**: *{{$field.PatchStrategy}}* {{end}} {{if $field.PatchMergeKey}}<br /> **patch merge key**: *{{$field.PatchMergeKey}}* {{end}} | {{$field.DescriptionWithEntities}}
{{end}}
{{end}}
`

var ConceptTemplate = `
{{define "pathparams"}}{{if .PathParams }}### Path Parameters

Parameter    | Description
------------ | -----------
{{range $param := .PathParams}}` + "`{{$param.Name}}`" + `{{if $param.Link}}<br /> *{{$param.Link}}* {{end}} | {{$param.Description}}
{{end}}{{end}}{{end}}

{{define "queryparams"}}{{if .QueryParams }}### Query Parameters

Parameter    | Description
------------ | -----------
{{range $param := .QueryParams}}` + "`{{$param.Name}}`" + `{{if $param.Link}}<br /> *{{$param.Link}}* {{end}} | {{$param.Description}}
{{end}}{{end}}{{end}}

{{define "bodyparams"}}{{if .BodyParams }}### Body Parameters

Parameter    | Description
------------ | -----------
{{range $param := .BodyParams}}` + "`{{$param.Name}}`" + `{{if $param.Link}}<br /> *{{$param.Link}}* {{end}} | {{$param.Description}}
{{end}}{{end}}{{end}}

{{define "responsebody"}}{{if .HttpResponses}}### Response

Code         | Description
------------ | -----------
{{range $i, $response := .HttpResponses}}{{$response.Code}} {{if $response.Field.Link}}<br /> *{{$response.Field.Link}}* {{end}} | {{$response.Field.Description}}
{{end}}{{end}}{{end}}

{{define "concept.template"}}

-----------
# {{.Name}} [{{if .Definition.ShowGroup}}{{.Definition.GroupDisplayName}}/{{end}}{{.Definition.Version}}] 

{{if .Definition.Sample.Sample}}{{$n := .Definition.Sample.Note}}{{range $e := .Definition.GetSamples}}>{{$e.Tab}} {{$n}}

` + "```" + `{{$e.Type}}

{{$e.Text}}

` + "```" + `
{{end}}{{end}}

Group        | Version     | Kind
------------ | ---------- | -----------
` + "`{{.Definition.GroupDisplayName}}` | `{{.Definition.Version}}` | `{{.Name}}`" + `

{{if .DescriptionWarning}}<aside class="warning">{{.DescriptionWarning}}</aside>{{end}}
{{if .DescriptionNote}}<aside class="notice">{{.DescriptionNote}}</aside>{{end}}

{{if .Definition.OtherVersions}}<aside class="notice">Other api versions of this object exist: {{range $v := .Definition.OtherVersions}}{{$v.VersionLink}} {{end}}</aside>{{end}}


{{.Definition.DescriptionWithEntities}}

{{if .Definition.AppearsIn}}<aside class="notice">
Appears In:

<ul> {{range $appearsin := .Definition.AppearsIn}}
<li>{{$appearsin.FullHrefLink}}</li>{{end}}
</ul> </aside>{{end}}

Field        | Description
------------ | -----------
{{range $field := .Definition.Fields}}` + "`{{$field.Name}}`" + `{{if $field.Link}}<br /> *{{$field.Link}}* {{end}} {{if $field.PatchStrategy}}<br /> **patch type**: *{{$field.PatchStrategy}}* {{end}} {{if $field.PatchMergeKey}}<br /> **patch merge key**: *{{$field.PatchMergeKey}}* {{end}} | {{$field.DescriptionWithEntities}}
{{end}}

{{if .Definition.Inline}}{{range $inline := .Definition.Inline}}### {{$inline.Name}} {{$inline.Version}} {{$inline.Group}}

{{if $inline.AppearsIn}}<aside class="notice">
Appears In:

<ul>{{range $appearsin := $inline.AppearsIn}}
<li>{{$appearsin.FullHrefLink}}</li>{{end}}
</ul></aside>{{end}}

Field        | Description
------------ | -----------
{{range $field := $inline.Fields}}` + "`{{$field.Name}}`" + `{{if $field.Link}}<br /> *{{$field.Link}}* {{end}} {{if $field.PatchStrategy}}<br /> **patch type**: *{{$field.PatchStrategy}}* {{end}} {{if $field.PatchMergeKey}}<br /> **patch merge key**: *{{$field.PatchMergeKey}}* {{end}} | {{$field.DescriptionWithEntities}}
{{end}}
{{end}}{{end}}

{{if .Definition.OperationCategories}}{{range $category := .Definition.OperationCategories}}{{if $category.Operations}}
## <strong>{{$category.Name}}</strong>

See supported operations below...

{{range $operation := $category.Operations}}## {{$operation.Type.Name}}

{{if $operation.GetExampleRequests}}{{range $er := $operation.GetExampleRequests}}>{{$er.Tab}} {{$er.Msg}}

` + "```" + `{{$er.Type}}

{{$er.Text}}

` + "```" + `

{{end}}{{end}}{{if $operation.GetExampleResponses}}{{range $er := $operation.GetExampleResponses}}>{{$er.Tab}} {{$er.Msg}}

` + "```" + `{{$er.Type}}

{{$er.Text}}

` + "```" + `
{{end}}{{end}}


{{$operation.Description}}

### HTTP Request

` + "`" + `{{$operation.GetDisplayHttp}}` + "`" + `

{{template "pathparams" $operation}}
{{template "queryparams" $operation}}
{{template "bodyparams" $operation}}
{{template "responsebody" $operation}}

{{end}}{{end}}{{end}}{{end}}

{{end}}

`
