package containers

import "github.com/gophercloud/gophercloud"

// ErrInvalidContainerName signals a container name containing an illegal
// character.
type ErrInvalidContainerName struct {
	gophercloud.BaseError
}

func (e ErrInvalidContainerName) Error() string {
	return "A container name must not contain: " + forbiddenContainerRunes
}
