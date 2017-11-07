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

var CategoryTemplate = `
# <strong>{{.Name}}</strong>
`

var CommandTemplate = `
------------

# {{.MainCommand.Name}}

{{.MainCommand.Example}}

{{.MainCommand.Description}}

### Usage

` + "`" + `$ {{.MainCommand.Usage}}` + "`" + `

{{if .MainCommand.Options}}

### Flags

Name | Shorthand | Default | Usage
---- | --------- | ------- | ----- {{range $option := .MainCommand.Options}}
{{$option.Name}} | {{$option.Shorthand}} | {{$option.DefaultValue}} | {{$option.Usage}} {{end}}
{{end}}
{{range $sub := .SubCommands}}
------------

## <em>{{$sub.Path}}</em>

{{$sub.Example}}

{{$sub.Description}}

### Usage

` + "`" + `$ {{$sub.Usage}}` + "`" + `

{{if $sub.Options}}

### Flags

Name | Shorthand | Default | Usage
---- | --------- | ------- | ----- {{range $option := $sub.Options}}
{{$option.Name}} | {{$option.Shorthand}} | {{$option.DefaultValue}} | {{$option.Usage}} {{end}}
{{end}}

{{end}}

`
