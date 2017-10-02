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

package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/markbates/inflect"
)

var Domain string
var Repo string
var GoSrc string

// writeIfNotFound returns true if the file was created and false if it already exists
func WriteIfNotFound(path, templateName, templateValue string, data interface{}) bool {
	// Make sure the directory exists
	os.MkdirAll(filepath.Dir(path), 0700)

	// Don't create the doc.go if it exists
	if _, err := os.Stat(path); err == nil {
		return false
	} else if !os.IsNotExist(err) {
		log.Fatalf("Could not stat %s: %v", path, err)

	}
	create(path)

	t := template.Must(template.New(templateName).Funcs(
		template.FuncMap{
			"title":  strings.Title,
			"lower":  strings.ToLower,
			"plural": inflect.NewDefaultRuleset().Pluralize,
		},
	).Parse(templateValue))

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", path, err)
	}
	defer f.Close()

	err = t.Execute(f, data)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", path, err)
	}

	return true
}

func GetCopyright(file string) string {
	if len(file) == 0 {
		file = "boilerplate.go.txt"
	}
	cr, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Must create boilerplate.go.txt file with copyright and file headers")
	}
	return string(cr)
}

func GetDomain() string {
	b, err := ioutil.ReadFile(filepath.Join("pkg", "apis", "doc.go"))
	if err != nil {
		log.Fatalf("Could not find pkg/apis/doc.go.  First run `apiserver-boot init --domain <domain>`.")
	}
	r := regexp.MustCompile("\\+domain=(.*)")
	l := r.FindSubmatch(b)
	if len(l) < 2 {
		log.Fatalf("pkg/apis/doc.go does not contain the domain (// +domain=.*)", l)
	}
	Domain = string(l[1])
	return Domain
}

func create(path string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
}

func DoCmd(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	log.Printf("%s\n", strings.Join(c.Args, " "))
	err := c.Run()
	if err != nil {
		log.Fatalf("command failed %v", err)
	}
}

func CheckInstall() {
	bins := []string{"apiregister-gen", "client-gen", "deepcopy-gen", "gen-apidocs", "informer-gen",
		"openapi-gen", "apiserver-boot", "conversion-gen", "defaulter-gen", "lister-gen"}
	missing := []string{}

	e, err := os.Executable()
	if err != nil {
		log.Fatal("unable to get directory of apiserver-builder tools")
	}

	dir := filepath.Dir(e)
	for _, b := range bins {
		_, err = os.Stat(filepath.Join(dir, b))
		if err != nil {
			missing = append(missing, b)
		}
	}
	if len(missing) > 0 {
		log.Fatalf("Error running apiserver-boot."+
			"\nThe following files are missing [%s]"+
			"\napiserver-boot must be installed using a release tar.gz downloaded from the git repo.",
			strings.Join(missing, ","))
	}
}
