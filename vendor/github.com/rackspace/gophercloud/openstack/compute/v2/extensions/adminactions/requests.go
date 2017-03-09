package adminactions

import (
	"fmt"

	"github.com/rackspace/gophercloud"
)

func actionURL(client *gophercloud.ServiceClient, id string) string {
	return client.ServiceURL("servers", id, "action")
}

type CreateBackupOpts struct {
	// Name: required, name of the backup.
	Name string

	// BackupType: required, type of the backup, such as "daily".
	BackupType string

	// Rotation: the number of backups to retain.
	Rotation int
}

// ToBackupCreateMap assembles a request body based on the contents of a CreateOpts.
func (opts CreateBackupOpts) ToCreateBackupMap() (map[string]interface{}, error) {
	backup := make(map[string]interface{})

	if opts.Name == "" {
		return nil, fmt.Errorf("CreateBackupOpts.Name cannot be blank.")
	}
	if opts.BackupType == "" {
		return nil, fmt.Errorf("CreateBackupOpts.BackupType cannot be blank.")
	}
	if opts.Rotation < 0 {
		return nil, fmt.Errorf("CreateBackupOpts.Rotation must 0 or greater.")
	}
	backup["name"] = opts.Name
	backup["backup_type"] = opts.BackupType
	backup["rotation"] = opts.Rotation

	return map[string]interface{}{"createBackup": backup}, nil
}

// ResetNetwork is the admin operation to create a backup of a Compute Server.
func CreateBackup(client *gophercloud.ServiceClient, id string, opts CreateBackupOpts) gophercloud.ErrResult {
	var res gophercloud.ErrResult

	req, err := opts.ToCreateBackupMap()
	if err != nil {
		res.Err = err
		return res
	}
	_, res.Err = client.Post(actionURL(client, id), req, nil, nil)
	return res

}

// InjectNetworkInfo is the admin operation which injects network info into a Compute Server.
func InjectNetworkInfo(client *gophercloud.ServiceClient, id string) gophercloud.ErrResult {
	var req struct {
		InjectNetworkInfo string `json:"injectNetworkInfo"`
	}

	var res gophercloud.ErrResult
	_, res.Err = client.Post(actionURL(client, id), req, nil, nil)
	return res
}

// Migrate is the admin operation to migrate a Compute Server.
func Migrate(client *gophercloud.ServiceClient, id string) gophercloud.ErrResult {
	var req struct {
		Migrate string `json:"migrate"`
	}

	var res gophercloud.ErrResult
	_, res.Err = client.Post(actionURL(client, id), req, nil, nil)
	return res
}

type LiveMigrateOpts struct {
	// Host: optional, If you omit this parameter, the scheduler chooses a host.
	Host string

	// BlockMigration:  defaults to false. Set to true to migrate local disks
	// by using block migration. If the source or destination host uses shared storage
	// and you set this value to true, the live migration fails.
	BlockMigration bool

	//DiskOverCommit: defaults to false. Set to true to enable over commit when the
	// destination host is checked for available disk space.
	DiskOverCommit bool
}

// ToServerCreateMap assembles a request body based on the contents of a CreateOpts.
func (opts LiveMigrateOpts) ToLiveMigrateMap() (map[string]interface{}, error) {
	migration := make(map[string]interface{})

	migration["host"] = opts.Host
	migration["block_migration"] = opts.BlockMigration
	migration["disk_over_commit"] = opts.DiskOverCommit

	return map[string]interface{}{"os-migrateLive": migration}, nil
}

// ResetNetwork is the admin operation to reset the network on a Compute Server.
func LiveMigrate(client *gophercloud.ServiceClient, id string, opts LiveMigrateOpts) gophercloud.ErrResult {
	var res gophercloud.ErrResult

	req, err := opts.ToLiveMigrateMap()
	if err != nil {
		res.Err = err
		return res
	}

	_, res.Err = client.Post(actionURL(client, id), req, nil, nil)
	return res

}

// ResetNetwork is the admin operation to reset the network on a Compute Server.
func ResetNetwork(client *gophercloud.ServiceClient, id string) gophercloud.ErrResult {
	var req struct {
		ResetNetwork string `json:"resetNetwork"`
	}

	var res gophercloud.ErrResult
	_, res.Err = client.Post(actionURL(client, id), req, nil, nil)
	return res
}

// ResetState is the admin operation to reset the state of a server.
func ResetState(client *gophercloud.ServiceClient, id string, state string) gophercloud.ErrResult {
	var res gophercloud.ErrResult
	var req struct {
		ResetState struct {
			State string `json:"state"`
		} `json:"os-resetState"`
	}
	req.ResetState.State = state

	_, res.Err = client.Post(actionURL(client, id), req, nil, nil)
	return res
}
