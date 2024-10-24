package v1

import (
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud/v2"
)

func CheckContainerName(s string) error {
	if len(s) < 1 {
		return ErrEmptyContainerName{}
	}
	if strings.ContainsRune(s, '/') {
		return ErrInvalidContainerName{name: s}
	}
	return nil
}

func CheckObjectName(s string) error {
	if s == "" {
		return ErrEmptyObjectName{}
	}
	return nil
}

// ErrInvalidContainerName signals a container name containing an illegal
// character.
type ErrInvalidContainerName struct {
	name string
	gophercloud.BaseError
}

func (e ErrInvalidContainerName) Error() string {
	return fmt.Sprintf("invalid name %q: a container name cannot contain a slash (/) character", e.name)
}

// ErrEmptyContainerName signals an empty container name.
type ErrEmptyContainerName struct {
	gophercloud.BaseError
}

func (e ErrEmptyContainerName) Error() string {
	return "a container name must not be empty"
}

// ErrEmptyObjectName signals an empty container name.
type ErrEmptyObjectName struct {
	gophercloud.BaseError
}

func (e ErrEmptyObjectName) Error() string {
	return "an object name must not be empty"
}
