package ecloud

import (
	"context"
	"time"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema" 
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

// ServerPublicNet represents a server's public network.
type ServerPublicNet struct {
	IPv4        string
	IPv6        string
	FloatingIPs []*string
	Firewalls   []*string
}

// ServerClient is a client for the servers API.
type ServerClient struct {
	client *Client
}

// GetByID retrieves a server by its ID. If the server does not exist, nil is returned.
func (c *ServerClient) GetByID(ctx context.Context, id string) (*schema.Server, error) {
    // Call ComputeStatus and get the full status response
    statusResp, err := c.client.ComputeStatus()
    if err != nil {
        return nil, err
    }

    // Search for the server with the matching ID
    for _, server := range statusResp.Servers {
        if server.UniqueID == id {
            return &server, nil
        }
    }

    // Server not found
    return nil, nil
}
