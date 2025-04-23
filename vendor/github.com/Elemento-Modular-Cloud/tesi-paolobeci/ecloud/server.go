package ecloud

import (
	"context"
	"time"

	// "github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema" // TODO
)

// VM representation inside Elemento
type Server struct {
	ID              int
	Name            string
	Status          ServerStatus
	Created         time.Time
	PublicNet       ServerPublicNet
	ServerType      *ServerType
	Datacenter      string
	BackupWindow    string
	Labels          map[string]string
	Volumes         []*Volume
}

// ServerType represents a server type in the Elemento.
type ServerType struct {
	ID           int
	Name         string
	Description  string
	Cores        int
	Memory       float32
	Disk         int
	Architecture Architecture
}

// Architecture specifies the architecture of the CPU.
type Architecture string

const (
	// ArchitectureX86 is the architecture for Intel/AMD x86 CPUs.
	ArchitectureX86 Architecture = "x86"

	// ArchitectureARM is the architecture for ARM CPUs.
	ArchitectureARM Architecture = "arm"
)


// ServerStatus specifies a server's status.
type ServerStatus string

const (
	// ServerStatusInitializing is the status when a server is initializing.
	ServerStatusInitializing ServerStatus = "initializing"

	// ServerStatusOff is the status when a server is off.
	ServerStatusOff ServerStatus = "off"

	// ServerStatusRunning is the status when a server is running.
	ServerStatusRunning ServerStatus = "running"

	// ServerStatusStarting is the status when a server is being started.
	ServerStatusStarting ServerStatus = "starting"

	// ServerStatusStopping is the status when a server is being stopped.
	ServerStatusStopping ServerStatus = "stopping"

	// ServerStatusMigrating is the status when a server is being migrated.
	ServerStatusMigrating ServerStatus = "migrating"

	// ServerStatusRebuilding is the status when a server is being rebuilt.
	ServerStatusRebuilding ServerStatus = "rebuilding"

	// ServerStatusDeleting is the status when a server is being deleted.
	ServerStatusDeleting ServerStatus = "deleting"

	// ServerStatusUnknown is the status when a server's state is unknown.
	ServerStatusUnknown ServerStatus = "unknown"
)

// FirewallStatus specifies a Firewall's status.
type FirewallStatus string

const (
	// FirewallStatusPending is the status when a Firewall is pending.
	FirewallStatusPending FirewallStatus = "pending"

	// FirewallStatusApplied is the status when a Firewall is applied.
	FirewallStatusApplied FirewallStatus = "applied"
)

// ServerPublicNet represents a server's public network.
type ServerPublicNet struct {
	IPv4        string
	IPv6        string
	FloatingIPs []*string
	Firewalls   []*string
}

// ServerRescueType represents rescue types.
type ServerRescueType string

// List of rescue types.
const (
	// Deprecated: Use ServerRescueTypeLinux64 instead.
	ServerRescueTypeLinux32 ServerRescueType = "linux32"
	ServerRescueTypeLinux64 ServerRescueType = "linux64"
)

// GetByID retrieves a server by its ID. If the server does not exist, nil is returned.
func (c *Client) GetByID(ctx context.Context, id int) (*Server, *Response, error) {
	// TODO
	// serve GetByID (info server by id, vedi hetzner)
	return nil, nil, nil
}

