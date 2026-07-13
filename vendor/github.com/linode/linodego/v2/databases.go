package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

type (
	DatabaseEngineType           string
	DatabaseDayOfWeek            int
	DatabaseMaintenanceFrequency string
	DatabaseStatus               string
	DatabasePlatform             string
	DatabaseMemberType           string
)

const (
	DatabaseMaintenanceDayMonday DatabaseDayOfWeek = iota + 1
	DatabaseMaintenanceDayTuesday
	DatabaseMaintenanceDayWednesday
	DatabaseMaintenanceDayThursday
	DatabaseMaintenanceDayFriday
	DatabaseMaintenanceDaySaturday
	DatabaseMaintenanceDaySunday
)

const (
	DatabaseMaintenanceFrequencyWeekly  DatabaseMaintenanceFrequency = "weekly"
	DatabaseMaintenanceFrequencyMonthly DatabaseMaintenanceFrequency = "monthly"
)

const (
	DatabaseEngineTypeMySQL    DatabaseEngineType = "mysql"
	DatabaseEngineTypePostgres DatabaseEngineType = "postgresql"
)

const (
	DatabaseStatusProvisioning DatabaseStatus = "provisioning"
	DatabaseStatusActive       DatabaseStatus = "active"
	DatabaseStatusDeleting     DatabaseStatus = "deleting"
	DatabaseStatusDeleted      DatabaseStatus = "deleted"
	DatabaseStatusSuspending   DatabaseStatus = "suspending"
	DatabaseStatusSuspended    DatabaseStatus = "suspended"
	DatabaseStatusResuming     DatabaseStatus = "resuming"
	DatabaseStatusRestoring    DatabaseStatus = "restoring"
	DatabaseStatusFailed       DatabaseStatus = "failed"
	DatabaseStatusDegraded     DatabaseStatus = "degraded"
	DatabaseStatusUpdating     DatabaseStatus = "updating"
	DatabaseStatusBackingUp    DatabaseStatus = "backing_up"
)

const (
	DatabasePlatformRDBMSLegacy  DatabasePlatform = "rdbms-legacy"
	DatabasePlatformRDBMSDefault DatabasePlatform = "rdbms-default"
)

const (
	DatabaseMemberTypePrimary  DatabaseMemberType = "primary"
	DatabaseMemberTypeFailover DatabaseMemberType = "failover"
)

// A Database is a instance of Linode Managed Databases
type Database struct {
	ID              int                       `json:"id"`
	Status          DatabaseStatus            `json:"status"`
	Label           string                    `json:"label"`
	Hosts           DatabaseHost              `json:"hosts"`
	Region          string                    `json:"region"`
	Type            string                    `json:"type"`
	Engine          string                    `json:"engine"`
	Version         string                    `json:"version"`
	ClusterSize     int                       `json:"cluster_size"`
	Platform        DatabasePlatform          `json:"platform"`
	Fork            *DatabaseFork             `json:"fork"`
	Updates         DatabaseMaintenanceWindow `json:"updates"`
	UsedDiskSizeGB  int                       `json:"used_disk_size_gb"`
	TotalDiskSizeGB int                       `json:"total_disk_size_gb"`
	Port            int                       `json:"port"`

	// Members has dynamic keys so it is a map
	Members map[string]DatabaseMemberType `json:"members"`

	Encrypted         bool       `json:"encrypted"`
	AllowList         []string   `json:"allow_list"`
	InstanceURI       string     `json:"instance_uri"`
	Created           *time.Time `json:"-"`
	Updated           *time.Time `json:"-"`
	OldestRestoreTime *time.Time `json:"-"`

	PrivateNetwork *DatabasePrivateNetwork `json:"private_network,omitzero"`
}

// DatabaseHost for Primary/Secondary of Database
type DatabaseHost struct {
	Primary string `json:"primary"`
	Standby string `json:"standby"`
}

type DatabasePrivateNetwork struct {
	VPCID        int  `json:"vpc_id"`
	SubnetID     int  `json:"subnet_id"`
	PublicAccess bool `json:"public_access"`
}

// DatabaseEngine is information about Engines supported by Linode Managed Databases
type DatabaseEngine struct {
	ID      string `json:"id"`
	Engine  string `json:"engine"`
	Version string `json:"version"`
}

// DatabaseMaintenanceWindow stores information about a MySQL cluster's maintenance window
type DatabaseMaintenanceWindow struct {
	DayOfWeek DatabaseDayOfWeek            `json:"day_of_week"`
	Duration  int                          `json:"duration"`
	Frequency DatabaseMaintenanceFrequency `json:"frequency"`
	HourOfDay int                          `json:"hour_of_day"`

	Pending []DatabaseMaintenanceWindowPending `json:"pending,omitzero"`
}

type DatabaseMaintenanceWindowPending struct {
	Deadline    *time.Time `json:"-"`
	Description string     `json:"description"`
	PlannedFor  *time.Time `json:"-"`
}

// DatabaseType is information about the supported Database Types by Linode Managed Databases
type DatabaseType struct {
	ID          string                `json:"id"`
	Label       string                `json:"label"`
	Class       string                `json:"class"`
	VirtualCPUs int                   `json:"vcpus"`
	Disk        int                   `json:"disk"`
	Memory      int                   `json:"memory"`
	Engines     DatabaseTypeEngineMap `json:"engines"`
	Deprecated  bool                  `json:"deprecated"`
}

// DatabaseTypeEngineMap stores a list of Database Engine types by engine
type DatabaseTypeEngineMap struct {
	MySQL      []DatabaseTypeEngine `json:"mysql"`
	PostgreSQL []DatabaseTypeEngine `json:"postgresql"`
}

// DatabaseTypeEngine Sizes and Prices
type DatabaseTypeEngine struct {
	Quantity int          `json:"quantity"`
	Price    ClusterPrice `json:"price"`
}

// ClusterPrice for Hourly and Monthly price models
type ClusterPrice struct {
	Hourly  float32 `json:"hourly"`
	Monthly float32 `json:"monthly"`
}

// DatabaseFork describes the source and restore time for the fork for forked DBs
type DatabaseFork struct {
	Source      int        `json:"source"`
	RestoreTime *time.Time `json:"-,omitzero"`
}

func (d *Database) UnmarshalJSON(b []byte) error {
	type Mask Database

	p := struct {
		*Mask

		Created           *parseabletime.ParseableTime `json:"created"`
		Updated           *parseabletime.ParseableTime `json:"updated"`
		OldestRestoreTime *parseabletime.ParseableTime `json:"oldest_restore_time"`
	}{
		Mask: (*Mask)(d),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	d.Created = (*time.Time)(p.Created)
	d.Updated = (*time.Time)(p.Updated)
	d.OldestRestoreTime = (*time.Time)(p.OldestRestoreTime)

	return nil
}

func (d *DatabaseFork) UnmarshalJSON(b []byte) error {
	type Mask DatabaseFork

	p := struct {
		*Mask

		RestoreTime *parseabletime.ParseableTime `json:"restore_time"`
	}{
		Mask: (*Mask)(d),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	d.RestoreTime = (*time.Time)(p.RestoreTime)

	return nil
}

func (d *DatabaseMaintenanceWindowPending) UnmarshalJSON(b []byte) error {
	type Mask DatabaseMaintenanceWindowPending

	p := struct {
		*Mask

		Deadline   *parseabletime.ParseableTime `json:"deadline"`
		PlannedFor *parseabletime.ParseableTime `json:"planned_for"`
	}{
		Mask: (*Mask)(d),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	d.Deadline = (*time.Time)(p.Deadline)
	d.PlannedFor = (*time.Time)(p.PlannedFor)

	return nil
}

// ListDatabases lists all Database instances in Linode Managed Databases for the account
func (c *Client) ListDatabases(ctx context.Context, opts *ListOptions) ([]Database, error) {
	return getPaginatedResults[Database](ctx, c, "databases/instances", opts)
}

// ListDatabaseEngines lists all Database Engines. This endpoint is cached by default.
func (c *Client) ListDatabaseEngines(ctx context.Context, opts *ListOptions) ([]DatabaseEngine, error) {
	return getPaginatedResults[DatabaseEngine](ctx, c, "databases/engines", opts)
}

// GetDatabaseEngine returns a specific Database Engine. This endpoint is cached by default.
func (c *Client) GetDatabaseEngine(ctx context.Context, _ *ListOptions, engineID string) (*DatabaseEngine, error) {
	e := formatAPIPath("databases/engines/%s", engineID)
	return doGETRequest[DatabaseEngine](ctx, c, e)
}

// ListDatabaseTypes lists all Types of Database provided in Linode Managed Databases. This endpoint is cached by default.
func (c *Client) ListDatabaseTypes(ctx context.Context, opts *ListOptions) ([]DatabaseType, error) {
	return getPaginatedResults[DatabaseType](ctx, c, "databases/types", opts)
}

// GetDatabaseType returns a specific Database Type. This endpoint is cached by default.
func (c *Client) GetDatabaseType(ctx context.Context, _ *ListOptions, typeID string) (*DatabaseType, error) {
	e := formatAPIPath("databases/types/%s", typeID)
	return doGETRequest[DatabaseType](ctx, c, e)
}
