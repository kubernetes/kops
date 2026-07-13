package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

type MySQLDatabaseTarget string

type MySQLDatabaseMaintenanceWindow = DatabaseMaintenanceWindow

const (
	MySQLDatabaseTargetPrimary   MySQLDatabaseTarget = "primary"
	MySQLDatabaseTargetSecondary MySQLDatabaseTarget = "secondary"
)

// A MySQLDatabase is an instance of Linode MySQL Managed Databases
type MySQLDatabase struct {
	ID          int              `json:"id"`
	Status      DatabaseStatus   `json:"status"`
	Label       string           `json:"label"`
	Hosts       DatabaseHost     `json:"hosts"`
	Region      string           `json:"region"`
	Type        string           `json:"type"`
	Engine      string           `json:"engine"`
	Version     string           `json:"version"`
	ClusterSize int              `json:"cluster_size"`
	Platform    DatabasePlatform `json:"platform"`

	// Members has dynamic keys so it is a map
	Members map[string]DatabaseMemberType `json:"members"`

	SSLConnection     bool                      `json:"ssl_connection"`
	Encrypted         bool                      `json:"encrypted"`
	AllowList         []string                  `json:"allow_list"`
	Created           *time.Time                `json:"-"`
	Updated           *time.Time                `json:"-"`
	Updates           DatabaseMaintenanceWindow `json:"updates"`
	Fork              *DatabaseFork             `json:"fork"`
	OldestRestoreTime *time.Time                `json:"-"`
	UsedDiskSizeGB    int                       `json:"used_disk_size_gb"`
	TotalDiskSizeGB   int                       `json:"total_disk_size_gb"`
	Port              int                       `json:"port"`

	EngineConfig   MySQLDatabaseEngineConfig `json:"engine_config"`
	PrivateNetwork *DatabasePrivateNetwork   `json:"private_network,omitzero"`
}

type MySQLDatabaseEngineConfig struct {
	MySQL                 *MySQLDatabaseEngineConfigMySQL `json:"mysql,omitzero"`
	BinlogRetentionPeriod *int                            `json:"binlog_retention_period,omitzero"`
}

type MySQLDatabaseEngineConfigMySQL struct {
	ConnectTimeout               *int     `json:"connect_timeout,omitzero"`
	DefaultTimeZone              *string  `json:"default_time_zone,omitzero"`
	GroupConcatMaxLen            *float64 `json:"group_concat_max_len,omitzero"`
	InformationSchemaStatsExpiry *int     `json:"information_schema_stats_expiry,omitzero"`
	InnoDBChangeBufferMaxSize    *int     `json:"innodb_change_buffer_max_size,omitzero"`
	InnoDBFlushNeighbors         *int     `json:"innodb_flush_neighbors,omitzero"`
	InnoDBFTMinTokenSize         *int     `json:"innodb_ft_min_token_size,omitzero"`
	InnoDBFTServerStopwordTable  **string `json:"innodb_ft_server_stopword_table,omitzero"`
	InnoDBLockWaitTimeout        *int     `json:"innodb_lock_wait_timeout,omitzero"`
	InnoDBLogBufferSize          *int     `json:"innodb_log_buffer_size,omitzero"`
	InnoDBOnlineAlterLogMaxSize  *int     `json:"innodb_online_alter_log_max_size,omitzero"`
	InnoDBReadIOThreads          *int     `json:"innodb_read_io_threads,omitzero"`
	InnoDBRollbackOnTimeout      *bool    `json:"innodb_rollback_on_timeout,omitzero"`
	InnoDBThreadConcurrency      *int     `json:"innodb_thread_concurrency,omitzero"`
	InnoDBWriteIOThreads         *int     `json:"innodb_write_io_threads,omitzero"`
	InteractiveTimeout           *int     `json:"interactive_timeout,omitzero"`
	InternalTmpMemStorageEngine  *string  `json:"internal_tmp_mem_storage_engine,omitzero"`
	MaxAllowedPacket             *int     `json:"max_allowed_packet,omitzero"`
	MaxHeapTableSize             *int     `json:"max_heap_table_size,omitzero"`
	NetBufferLength              *int     `json:"net_buffer_length,omitzero"`
	NetReadTimeout               *int     `json:"net_read_timeout,omitzero"`
	NetWriteTimeout              *int     `json:"net_write_timeout,omitzero"`
	SortBufferSize               *int     `json:"sort_buffer_size,omitzero"`
	SQLMode                      *string  `json:"sql_mode,omitzero"`
	SQLRequirePrimaryKey         *bool    `json:"sql_require_primary_key,omitzero"`
	TmpTableSize                 *int     `json:"tmp_table_size,omitzero"`
	WaitTimeout                  *int     `json:"wait_timeout,omitzero"`
}

type MySQLDatabaseConfigInfo struct {
	MySQL                 MySQLDatabaseConfigInfoMySQL                 `json:"mysql"`
	BinlogRetentionPeriod MySQLDatabaseConfigInfoBinlogRetentionPeriod `json:"binlog_retention_period"`
}

type MySQLDatabaseConfigInfoMySQL struct {
	ConnectTimeout               ConnectTimeout               `json:"connect_timeout"`
	DefaultTimeZone              DefaultTimeZone              `json:"default_time_zone"`
	GroupConcatMaxLen            GroupConcatMaxLen            `json:"group_concat_max_len"`
	InformationSchemaStatsExpiry InformationSchemaStatsExpiry `json:"information_schema_stats_expiry"`
	InnoDBChangeBufferMaxSize    InnoDBChangeBufferMaxSize    `json:"innodb_change_buffer_max_size"`
	InnoDBFlushNeighbors         InnoDBFlushNeighbors         `json:"innodb_flush_neighbors"`
	InnoDBFTMinTokenSize         InnoDBFTMinTokenSize         `json:"innodb_ft_min_token_size"`
	InnoDBFTServerStopwordTable  InnoDBFTServerStopwordTable  `json:"innodb_ft_server_stopword_table"`
	InnoDBLockWaitTimeout        InnoDBLockWaitTimeout        `json:"innodb_lock_wait_timeout"`
	InnoDBLogBufferSize          InnoDBLogBufferSize          `json:"innodb_log_buffer_size"`
	InnoDBOnlineAlterLogMaxSize  InnoDBOnlineAlterLogMaxSize  `json:"innodb_online_alter_log_max_size"`
	InnoDBReadIOThreads          InnoDBReadIOThreads          `json:"innodb_read_io_threads"`
	InnoDBRollbackOnTimeout      InnoDBRollbackOnTimeout      `json:"innodb_rollback_on_timeout"`
	InnoDBThreadConcurrency      InnoDBThreadConcurrency      `json:"innodb_thread_concurrency"`
	InnoDBWriteIOThreads         InnoDBWriteIOThreads         `json:"innodb_write_io_threads"`
	InteractiveTimeout           InteractiveTimeout           `json:"interactive_timeout"`
	InternalTmpMemStorageEngine  InternalTmpMemStorageEngine  `json:"internal_tmp_mem_storage_engine"`
	MaxAllowedPacket             MaxAllowedPacket             `json:"max_allowed_packet"`
	MaxHeapTableSize             MaxHeapTableSize             `json:"max_heap_table_size"`
	NetBufferLength              NetBufferLength              `json:"net_buffer_length"`
	NetReadTimeout               NetReadTimeout               `json:"net_read_timeout"`
	NetWriteTimeout              NetWriteTimeout              `json:"net_write_timeout"`
	SortBufferSize               SortBufferSize               `json:"sort_buffer_size"`
	SQLMode                      SQLMode                      `json:"sql_mode"`
	SQLRequirePrimaryKey         SQLRequirePrimaryKey         `json:"sql_require_primary_key"`
	TmpTableSize                 TmpTableSize                 `json:"tmp_table_size"`
	WaitTimeout                  WaitTimeout                  `json:"wait_timeout"`
}

type ConnectTimeout struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type DefaultTimeZone struct {
	Description     string `json:"description"`
	Example         string `json:"example"`
	MaxLength       int    `json:"maxLength"`
	MinLength       int    `json:"minLength"`
	Pattern         string `json:"pattern"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type GroupConcatMaxLen struct {
	Description     string  `json:"description"`
	Example         float64 `json:"example"`
	Maximum         float64 `json:"maximum"`
	Minimum         float64 `json:"minimum"`
	RequiresRestart bool    `json:"requires_restart"`
	Type            string  `json:"type"`
}

type InformationSchemaStatsExpiry struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBChangeBufferMaxSize struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBFlushNeighbors struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBFTMinTokenSize struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBFTServerStopwordTable struct {
	Description     string   `json:"description"`
	Example         string   `json:"example"`
	MaxLength       int      `json:"maxLength"`
	Pattern         string   `json:"pattern"`
	RequiresRestart bool     `json:"requires_restart"`
	Type            []string `json:"type"`
}

type InnoDBLockWaitTimeout struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBLogBufferSize struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBOnlineAlterLogMaxSize struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBReadIOThreads struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBRollbackOnTimeout struct {
	Description     string `json:"description"`
	Example         bool   `json:"example"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBThreadConcurrency struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InnoDBWriteIOThreads struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InteractiveTimeout struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type InternalTmpMemStorageEngine struct {
	Description     string   `json:"description"`
	Enum            []string `json:"enum"`
	Example         string   `json:"example"`
	RequiresRestart bool     `json:"requires_restart"`
	Type            string   `json:"type"`
}

type MaxAllowedPacket struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type MaxHeapTableSize struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type NetBufferLength struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type NetReadTimeout struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type NetWriteTimeout struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type SortBufferSize struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type SQLMode struct {
	Description     string `json:"description"`
	Example         string `json:"example"`
	MaxLength       int    `json:"maxLength"`
	Pattern         string `json:"pattern"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type SQLRequirePrimaryKey struct {
	Description     string `json:"description"`
	Example         bool   `json:"example"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type TmpTableSize struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type WaitTimeout struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

type MySQLDatabaseConfigInfoBinlogRetentionPeriod struct {
	Description     string `json:"description"`
	Example         int    `json:"example"`
	Maximum         int    `json:"maximum"`
	Minimum         int    `json:"minimum"`
	RequiresRestart bool   `json:"requires_restart"`
	Type            string `json:"type"`
}

func (d *MySQLDatabase) UnmarshalJSON(b []byte) error {
	type Mask MySQLDatabase

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

// MySQLCreateOptions fields are used when creating a new MySQL Database
type MySQLCreateOptions struct {
	Label       string   `json:"label"`
	Region      string   `json:"region"`
	Type        string   `json:"type"`
	Engine      string   `json:"engine"`
	AllowList   []string `json:"allow_list,omitzero"`
	ClusterSize int      `json:"cluster_size,omitzero"`

	Fork           *DatabaseFork              `json:"fork,omitzero"`
	EngineConfig   *MySQLDatabaseEngineConfig `json:"engine_config,omitzero"`
	PrivateNetwork *DatabasePrivateNetwork    `json:"private_network,omitzero"`
}

// MySQLUpdateOptions fields are used when altering the existing MySQL Database
type MySQLUpdateOptions struct {
	Label          string                     `json:"label,omitzero"`
	AllowList      []string                   `json:"allow_list,omitzero"`
	Updates        *DatabaseMaintenanceWindow `json:"updates,omitzero"`
	Type           string                     `json:"type,omitzero"`
	ClusterSize    int                        `json:"cluster_size,omitzero"`
	Version        string                     `json:"version,omitzero"`
	EngineConfig   *MySQLDatabaseEngineConfig `json:"engine_config,omitzero"`
	PrivateNetwork **DatabasePrivateNetwork   `json:"private_network,omitzero"`
}

// MySQLDatabaseCredential is the Root Credentials to access the Linode Managed Database
type MySQLDatabaseCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// MySQLDatabaseSSL is the SSL Certificate to access the Linode Managed MySQL Database
type MySQLDatabaseSSL struct {
	CACertificate []byte `json:"ca_certificate"`
}

// ListMySQLDatabases lists all MySQL Databases associated with the account
func (c *Client) ListMySQLDatabases(ctx context.Context, opts *ListOptions) ([]MySQLDatabase, error) {
	return getPaginatedResults[MySQLDatabase](ctx, c, "databases/mysql/instances", opts)
}

// GetMySQLDatabase returns a single MySQL Database matching the id
func (c *Client) GetMySQLDatabase(ctx context.Context, databaseID int) (*MySQLDatabase, error) {
	e := formatAPIPath("databases/mysql/instances/%d", databaseID)
	return doGETRequest[MySQLDatabase](ctx, c, e)
}

// CreateMySQLDatabase creates a new MySQL Database using the createOpts as configuration, returns the new MySQL Database
func (c *Client) CreateMySQLDatabase(ctx context.Context, opts MySQLCreateOptions) (*MySQLDatabase, error) {
	return doPOSTRequest[MySQLDatabase](ctx, c, "databases/mysql/instances", opts)
}

// DeleteMySQLDatabase deletes an existing MySQL Database with the given id
func (c *Client) DeleteMySQLDatabase(ctx context.Context, databaseID int) error {
	e := formatAPIPath("databases/mysql/instances/%d", databaseID)
	return doDELETERequest(ctx, c, e)
}

// UpdateMySQLDatabase updates the given MySQL Database with the provided opts, returns the MySQLDatabase with the new settings
func (c *Client) UpdateMySQLDatabase(ctx context.Context, databaseID int, opts MySQLUpdateOptions) (*MySQLDatabase, error) {
	e := formatAPIPath("databases/mysql/instances/%d", databaseID)
	return doPUTRequest[MySQLDatabase](ctx, c, e, opts)
}

// GetMySQLDatabaseSSL returns the SSL Certificate for the given MySQL Database
func (c *Client) GetMySQLDatabaseSSL(ctx context.Context, databaseID int) (*MySQLDatabaseSSL, error) {
	e := formatAPIPath("databases/mysql/instances/%d/ssl", databaseID)
	return doGETRequest[MySQLDatabaseSSL](ctx, c, e)
}

// GetMySQLDatabaseCredentials returns the Root Credentials for the given MySQL Database
func (c *Client) GetMySQLDatabaseCredentials(ctx context.Context, databaseID int) (*MySQLDatabaseCredential, error) {
	e := formatAPIPath("databases/mysql/instances/%d/credentials", databaseID)
	return doGETRequest[MySQLDatabaseCredential](ctx, c, e)
}

// ResetMySQLDatabaseCredentials returns the Root Credentials for the given MySQL Database (may take a few seconds to work)
func (c *Client) ResetMySQLDatabaseCredentials(ctx context.Context, databaseID int) error {
	e := formatAPIPath("databases/mysql/instances/%d/credentials/reset", databaseID)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}

// PatchMySQLDatabase applies security patches and updates to the underlying operating system of the Managed MySQL Database
func (c *Client) PatchMySQLDatabase(ctx context.Context, databaseID int) error {
	e := formatAPIPath("databases/mysql/instances/%d/patch", databaseID)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}

// SuspendMySQLDatabase suspends a MySQL Managed Database, releasing idle resources and keeping only necessary data.
// All service data is lost if there are no backups available.
func (c *Client) SuspendMySQLDatabase(ctx context.Context, databaseID int) error {
	e := formatAPIPath("databases/mysql/instances/%d/suspend", databaseID)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}

// ResumeMySQLDatabase resumes a suspended MySQL Managed Database
func (c *Client) ResumeMySQLDatabase(ctx context.Context, databaseID int) error {
	e := formatAPIPath("databases/mysql/instances/%d/resume", databaseID)
	return doPOSTRequestNoRequestResponseBody(ctx, c, e)
}

// GetMySQLDatabaseConfig returns a detailed list of all the configuration options for MySQL Databases
func (c *Client) GetMySQLDatabaseConfig(ctx context.Context) (*MySQLDatabaseConfigInfo, error) {
	return doGETRequest[MySQLDatabaseConfigInfo](ctx, c, "databases/mysql/config")
}
