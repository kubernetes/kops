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
	//"gopkg.in/yaml.v2"
	"encoding/json"
	"fmt"
	"strings"
)

type ExampleProvider interface {
	GetTab() string
	GetRequestMessage() string
	GetResponseMessage() string
	GetRequestType() string
	GetResponseType() string
	GetSampleType() string
	GetSample(d *Definition) string
	GetRequest(o *Operation) string
	GetResponse(o *Operation) string
}

var ExampleProviders = []ExampleProvider{
	KubectlExample{},
	CurlExample{},
}

var EmptyExampleProviders = []ExampleProvider {
	EmptyExample{},
}

func GetExampleProviders() []ExampleProvider {
	if *BuildOps {
		return ExampleProviders
	} else {
		return EmptyExampleProviders
	}
}

type EmptyExample struct {
}

var _ ExampleProvider = &EmptyExample{}

func (ce EmptyExample) GetSample(d *Definition) string {
	return d.Sample.Sample
}

func (ce EmptyExample) GetRequestMessage() string {
	return ""
}

func (ce EmptyExample) GetResponseMessage() string {
	return ""
}

func (ce EmptyExample) GetTab() string {
	return "bdocs-tab:example"
}

func (ce EmptyExample) GetRequestType() string {
	return "bdocs-tab:example_shell"
}

func (ce EmptyExample) GetResponseType() string {
	return "bdocs-tab:example_json"
}

func (ce EmptyExample) GetSampleType() string {
	return "bdocs-tab:example_yaml"
}

func (ce EmptyExample) GetRequest(o *Operation) string {
	return ""
}

func (ce EmptyExample) GetResponse(o *Operation) string {
	return ""
}

var _ ExampleProvider = &CurlExample{}

type CurlExample struct {
}

func (ce CurlExample) GetSample(d *Definition) string {
	return d.Sample.Sample
}

func (ce CurlExample) GetRequestMessage() string {
	return "`curl` Command (*requires `kubectl proxy` to be running*)"
}

func (ce CurlExample) GetResponseMessage() string {
	return "Response Body"
}

func (ce CurlExample) GetTab() string {
	return "bdocs-tab:curl"
}

func (ce CurlExample) GetRequestType() string {
	return "bdocs-tab:curl_shell"
}

func (ce CurlExample) GetResponseType() string {
	return "bdocs-tab:curl_json"
}

func (ce CurlExample) GetSampleType() string {
	return "bdocs-tab:curl_yaml"
}

func (ce CurlExample) GetRequest(o *Operation) string {
	c := o.ExampleConfig
	y := c.Request
	if len(y) <= 0 && len(c.Name) <= 0 {
		return "Coming Soon"
	}

	switch o.Type.Name {
	case "Create":
		return fmt.Sprintf("$ kubectl proxy\n$ curl -X POST -H 'Content-Type: application/yaml' --data '\n%s' http://127.0.0.1:8001%s", y, strings.Replace(o.Path, "{namespace}", "default", -1))
	case "Delete":
		path := strings.Replace(o.Path, "{namespace}", o.ExampleConfig.Namespace, -1)
		path = strings.Replace(path, "{name}", o.ExampleConfig.Name, -1)
		return fmt.Sprintf("$ kubectl proxy\n$ curl -X DELETE -H 'Content-Type: application/yaml' --data '\n%s' 'http://127.0.0.1:8001%s'", y, path)
	case "List":
		path := strings.Replace(o.Path, "{namespace}", o.ExampleConfig.Namespace, -1)
		path = strings.Replace(path, "{name}", o.ExampleConfig.Name, -1)
		return fmt.Sprintf("$ kubectl proxy\n$ curl -X GET 'http://127.0.0.1:8001%s'", path)
	case "Patch":
		path := strings.Replace(o.Path, "{namespace}", o.ExampleConfig.Namespace, -1)
		path = strings.Replace(path, "{name}", o.ExampleConfig.Name, -1)
		return fmt.Sprintf("$ kubectl proxy\n$ curl -X PATCH -H 'Content-Type: application/strategic-merge-patch+json' --data '\n%s' \\\n\t'http://127.0.0.1:8001%s'", y, path)
	case "Read":
		path := strings.Replace(o.Path, "{namespace}", o.ExampleConfig.Namespace, -1)
		path = strings.Replace(path, "{name}", o.ExampleConfig.Name, -1)
		return fmt.Sprintf("$ kubectl proxy\n$ curl -X GET http://127.0.0.1:8001%s", path)
	case "Replace":
		path := strings.Replace(o.Path, "{namespace}", o.ExampleConfig.Namespace, -1)
		path = strings.Replace(path, "{name}", o.ExampleConfig.Name, -1)
		return fmt.Sprintf("$ kubectl proxy\n$ curl -X PUT -H 'Content-Type: application/yaml' --data '\n%s' http://127.0.0.1:8001%s", y, path)
	case "Watch":
		path := strings.Replace(o.Path, "{namespace}", o.ExampleConfig.Namespace, -1)
		path = strings.Replace(path, "{name}", o.ExampleConfig.Name, -1)
		return fmt.Sprintf("$ kubectl proxy\n$ curl -X GET 'http://127.0.0.1:8001%s'", path)
	}
	return ""
}

func (ce CurlExample) GetResponse(o *Operation) string {
	c := o.ExampleConfig
	j := o.ExampleConfig.Response
	if len(j) <= 0 && len(c.Name) <= 0 {
		return "Coming Soon"
	}
	switch o.Type.Name {
	case "Create":
		return fmt.Sprintf("%s", j)
	case "Delete":
		return fmt.Sprintf("%s", j)
	case "List":
		return fmt.Sprintf("%s", j)
	case "Patch":
		return fmt.Sprintf("%s", j)
	case "Read":
		return fmt.Sprintf("%s", j)
	case "Replace":
		return fmt.Sprintf("%s", j)
	case "Watch":
		return fmt.Sprintf("%s", j)
	}
	return ""
}

var _ ExampleProvider = &KubectlExample{}

type KubectlExample struct{}

func (ke KubectlExample) GetSample(d *Definition) string {
	return d.Sample.Sample
}

func (ke KubectlExample) GetRequestMessage() string {
	return "`kubectl` Command"
}

func (ke KubectlExample) GetResponseMessage() string {
	return "Output"
}

func (ke KubectlExample) GetTab() string {
	return "bdocs-tab:kubectl"
}

func (ke KubectlExample) GetRequestType() string {
	return "bdocs-tab:kubectl_shell"
}

func (ke KubectlExample) GetResponseType() string {
	return "bdocs-tab:kubectl_json"
}

func (ke KubectlExample) GetSampleType() string {
	return "bdocs-tab:kubectl_yaml"
}

func (ke KubectlExample) GetRequest(o *Operation) string {
	c := o.ExampleConfig
	t := strings.ToLower(o.Definition.Name)
	y := c.Request
	if len(y) <= 0 && len(c.Name) <= 0 {
		return "Coming Soon"
	}
	switch o.Type.Name {
	case "Create":
		return fmt.Sprintf("$ echo '%s' | kubectl create -f -", y)
	case "Delete":
		return fmt.Sprintf("$ kubectl delete %s %s", t, c.Name)
	case "List":
		return fmt.Sprintf("$ kubectl get %s -o json", t)
	case "Patch":
		return fmt.Sprintf("$ kubectl patch %s %s -p \\\n\t'%s'", t, c.Name, c.Request)
	case "Read":
		return fmt.Sprintf("$ kubectl get %s %s -o json", t, c.Name)
	case "Replace":
		return fmt.Sprintf("$ echo '%s' | kubectl replace -f -", y)
	case "Watch":
		return fmt.Sprintf("$ kubectl get %s %s --watch -o json", t, c.Name)
	}
	return ""
}

func (ke KubectlExample) GetResponse(o *Operation) string {
	c := o.ExampleConfig
	name := o.ExampleConfig.Name
	t := strings.ToLower(o.Definition.Name)
	j := o.ExampleConfig.Response
	if len(j) <= 0 && len(c.Name) <= 0 {
		return "Coming Soon"
	}
	switch o.Type.Name {
	case "Create":
		return fmt.Sprintf("%s \"%s\" created", t, name)
	case "Delete":
		return fmt.Sprintf("%s \"%s\" deleted", t, name)
	case "List":
		return fmt.Sprintf("%s", j)
	case "Patch":
		return fmt.Sprintf("\"%s\" patched", name)
	case "Read":
		return string(j)
	case "Replace":
		return fmt.Sprintf("%s \"%s\" replaced", t, name)
	case "Watch":
		return string(j)
	}
	return ""
}

func GetName(parsed map[string]interface{}) string {
	meta := parsed["metadata"].(map[string]interface{})
	name := meta["name"].(string)
	return name
}

func ParseJson(j []byte) map[string]interface{} {
	var parsed interface{}
	err := json.Unmarshal(j, &parsed)
	if err != nil {
		panic(fmt.Sprintf("Could not parse json %s y: %v\n", j, err))
	}
	return parsed.(map[string]interface{})
}
