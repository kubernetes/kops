package vfs

type WriteOption string

const (
	WriteOptionCreate       WriteOption = "Create"
	WriteOptionOnlyIfExists WriteOption = "IfExists"
)
