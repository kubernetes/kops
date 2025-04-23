package ecloud

import (
	"time"

	// "github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema" // TODO
)

// Volume represents a volume in the Elemento.
type Volume struct {
	ID          int
	Name        string
	Status      VolumeStatus
	Server      *Server
	Location    string
	Size        int
	Format      *string
	Labels      map[string]string
	LinuxDevice string
	Created     time.Time
}

// VolumeStatus specifies a volume's status.
type VolumeStatus string

const (
	// VolumeStatusCreating is the status when a volume is being created.
	VolumeStatusCreating VolumeStatus = "creating"

	// VolumeStatusAvailable is the status when a volume is available.
	VolumeStatusAvailable VolumeStatus = "available"
)
