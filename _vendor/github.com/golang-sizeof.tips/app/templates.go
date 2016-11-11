package app

import (
	"html/template"

	bin "github.com/gophergala/golang-sizeof.tips/internal/bindata/templates"
)

const templatesDir = "templs/"

var templates map[string]*template.Template

func prepareTemplates() error {
	templates = make(map[string]*template.Template)
	baseData, err := bin.Asset(templatesDir + "parts/base.tmpl")
	if err != nil {
		return err
	}
	var fns = template.FuncMap{
		"unvischunk": func(x int, len int) bool {
			return x > 2 && x < (len-1)
		},
	}
	for _, name := range []string{
		"index", "404", "500",
	} {
		assetData, err := bin.Asset(templatesDir + name + ".tmpl")
		if err != nil {
			return err
		}
		templates[name], err = template.New(name).Funcs(fns).Parse(
			string(baseData) + string(assetData),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
