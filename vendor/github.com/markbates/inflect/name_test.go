package inflect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Name_Camel(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "foo_bar", E: "FooBar"},
		{V: "widget", E: "Widget"},
		{V: "User", E: "User"},
		{V: "user_id", E: "UserID"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).Camel())
	}
}

func Test_Name_ParamID(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "foo_bar", E: "foo_bar_id"},
		{V: "admin/widget", E: "admin_widget_id"},
		{V: "widget", E: "widget_id"},
		{V: "User", E: "user_id"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).ParamID())
	}
}

func Test_Name_Title(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "foo_bar", E: "Foo Bar"},
		{V: "admin/widget", E: "Admin Widget"},
		{V: "widget", E: "Widget"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).Title())
	}
}

func Test_Name_Model(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "foo_bar", E: "FooBar"},
		{V: "admin/widget", E: "AdminWidget"},
		{V: "widget", E: "Widget"},
		{V: "widgets", E: "Widget"},
		{V: "status", E: "Status"},
		{V: "Statuses", E: "Status"},
		{V: "statuses", E: "Status"},
		{V: "People", E: "Person"},
		{V: "people", E: "Person"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).Model())
	}
}

func Test_Name_Resource(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "Person", E: "People"},
		{V: "foo_bar", E: "FooBars"},
		{V: "admin/widget", E: "AdminWidgets"},
		{V: "widget", E: "Widgets"},
		{V: "widgets", E: "Widgets"},
		{V: "greatPerson", E: "GreatPeople"},
		{V: "great/person", E: "GreatPeople"},
		{V: "status", E: "Statuses"},
		{V: "Status", E: "Statuses"},
		{V: "Statuses", E: "Statuses"},
		{V: "statuses", E: "Statuses"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).Resource())
	}
}

func Test_Name_ModelPlural(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "foo_bar", E: "FooBars"},
		{V: "admin/widget", E: "AdminWidgets"},
		{V: "widget", E: "Widgets"},
		{V: "widgets", E: "Widgets"},
		{V: "status", E: "Statuses"},
		{V: "statuses", E: "Statuses"},
		{V: "people", E: "People"},
		{V: "person", E: "People"},
		{V: "People", E: "People"},
		{V: "Status", E: "Statuses"},
	}

	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).ModelPlural())
	}
}

func Test_Name_File(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "foo_bar", E: "foo_bar"},
		{V: "admin/widget", E: "admin/widget"},
		{V: "widget", E: "widget"},
		{V: "widgets", E: "widgets"},
		{V: "User", E: "user"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).File())
	}
}

func Test_Name_VarCaseSingular(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "foo_bar", E: "fooBar"},
		{V: "admin/widget", E: "adminWidget"},
		{V: "widget", E: "widget"},
		{V: "widgets", E: "widget"},
		{V: "User", E: "user"},
		{V: "FooBar", E: "fooBar"},
		{V: "status", E: "status"},
		{V: "statuses", E: "status"},
		{V: "Status", E: "status"},
		{V: "Statuses", E: "status"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).VarCaseSingular())
	}
}

func Test_Name_VarCasePlural(t *testing.T) {
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: "foo_bar", E: "fooBars"},
		{V: "admin/widget", E: "adminWidgets"},
		{V: "widget", E: "widgets"},
		{V: "widgets", E: "widgets"},
		{V: "User", E: "users"},
		{V: "FooBar", E: "fooBars"},
		{V: "status", E: "statuses"},
		{V: "statuses", E: "statuses"},
		{V: "Status", E: "statuses"},
		{V: "Statuses", E: "statuses"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).VarCasePlural())
	}
}

func Test_Name_Package(t *testing.T) {
	gp := os.Getenv("GOPATH")
	r := require.New(t)
	table := []struct {
		V string
		E string
	}{
		{V: filepath.Join(gp, "src", "admin/widget"), E: "admin/widget"},
		{V: filepath.Join(gp, "admin/widget"), E: "admin/widget"},
		{V: "admin/widget", E: "admin/widget"},
	}
	for _, tt := range table {
		r.Equal(tt.E, Name(tt.V).Package())
	}
}

func Test_Name_Char(t *testing.T) {
	r := require.New(t)

	n := Name("Foo")
	r.Equal("f", n.Char())
}
