package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// InstanceConfig represents all of the settings that control the boot and run configuration of a Linode Instance
type InstanceConfig struct {
	ID          int                       `json:"id"`
	Label       string                    `json:"label"`
	Comments    string                    `json:"comments"`
	Devices     *InstanceConfigDeviceMap  `json:"devices"`
	Helpers     *InstanceConfigHelpers    `json:"helpers"`
	Interfaces  []InstanceConfigInterface `json:"interfaces"`
	MemoryLimit int                       `json:"memory_limit"`
	Kernel      string                    `json:"kernel"`
	InitRD      *int                      `json:"init_rd"`
	RootDevice  string                    `json:"root_device"`
	RunLevel    string                    `json:"run_level"`
	VirtMode    string                    `json:"virt_mode"`
	Created     *time.Time                `json:"-"`
	Updated     *time.Time                `json:"-"`
}

// InstanceConfigDevice contains either the DiskID or VolumeID assigned to a Config Device
type InstanceConfigDevice struct {
	DiskID   int `json:"disk_id,omitzero"`
	VolumeID int `json:"volume_id,omitzero"`
}

// InstanceConfigDeviceMap contains SDA-SDH InstanceConfigDevice settings
type InstanceConfigDeviceMap struct {
	// sda-sdz
	SDA *InstanceConfigDevice `json:"sda,omitzero"`
	SDB *InstanceConfigDevice `json:"sdb,omitzero"`
	SDC *InstanceConfigDevice `json:"sdc,omitzero"`
	SDD *InstanceConfigDevice `json:"sdd,omitzero"`
	SDE *InstanceConfigDevice `json:"sde,omitzero"`
	SDF *InstanceConfigDevice `json:"sdf,omitzero"`
	SDG *InstanceConfigDevice `json:"sdg,omitzero"`
	SDH *InstanceConfigDevice `json:"sdh,omitzero"`
	SDI *InstanceConfigDevice `json:"sdi,omitzero"`
	SDJ *InstanceConfigDevice `json:"sdj,omitzero"`
	SDK *InstanceConfigDevice `json:"sdk,omitzero"`
	SDL *InstanceConfigDevice `json:"sdl,omitzero"`
	SDM *InstanceConfigDevice `json:"sdm,omitzero"`
	SDN *InstanceConfigDevice `json:"sdn,omitzero"`
	SDO *InstanceConfigDevice `json:"sdo,omitzero"`
	SDP *InstanceConfigDevice `json:"sdp,omitzero"`
	SDQ *InstanceConfigDevice `json:"sdq,omitzero"`
	SDR *InstanceConfigDevice `json:"sdr,omitzero"`
	SDS *InstanceConfigDevice `json:"sds,omitzero"`
	SDT *InstanceConfigDevice `json:"sdt,omitzero"`
	SDU *InstanceConfigDevice `json:"sdu,omitzero"`
	SDV *InstanceConfigDevice `json:"sdv,omitzero"`
	SDW *InstanceConfigDevice `json:"sdw,omitzero"`
	SDX *InstanceConfigDevice `json:"sdx,omitzero"`
	SDY *InstanceConfigDevice `json:"sdy,omitzero"`
	SDZ *InstanceConfigDevice `json:"sdz,omitzero"`

	// sdaa-sdaz
	SDAA *InstanceConfigDevice `json:"sdaa,omitzero"`
	SDAB *InstanceConfigDevice `json:"sdab,omitzero"`
	SDAC *InstanceConfigDevice `json:"sdac,omitzero"`
	SDAD *InstanceConfigDevice `json:"sdad,omitzero"`
	SDAE *InstanceConfigDevice `json:"sdae,omitzero"`
	SDAF *InstanceConfigDevice `json:"sdaf,omitzero"`
	SDAG *InstanceConfigDevice `json:"sdag,omitzero"`
	SDAH *InstanceConfigDevice `json:"sdah,omitzero"`
	SDAI *InstanceConfigDevice `json:"sdai,omitzero"`
	SDAJ *InstanceConfigDevice `json:"sdaj,omitzero"`
	SDAK *InstanceConfigDevice `json:"sdak,omitzero"`
	SDAL *InstanceConfigDevice `json:"sdal,omitzero"`
	SDAM *InstanceConfigDevice `json:"sdam,omitzero"`
	SDAN *InstanceConfigDevice `json:"sdan,omitzero"`
	SDAO *InstanceConfigDevice `json:"sdao,omitzero"`
	SDAP *InstanceConfigDevice `json:"sdap,omitzero"`
	SDAQ *InstanceConfigDevice `json:"sdaq,omitzero"`
	SDAR *InstanceConfigDevice `json:"sdar,omitzero"`
	SDAS *InstanceConfigDevice `json:"sdas,omitzero"`
	SDAT *InstanceConfigDevice `json:"sdat,omitzero"`
	SDAU *InstanceConfigDevice `json:"sdau,omitzero"`
	SDAV *InstanceConfigDevice `json:"sdav,omitzero"`
	SDAW *InstanceConfigDevice `json:"sdaw,omitzero"`
	SDAX *InstanceConfigDevice `json:"sdax,omitzero"`
	SDAY *InstanceConfigDevice `json:"sday,omitzero"`
	SDAZ *InstanceConfigDevice `json:"sdaz,omitzero"`

	// sdba-sdbl
	SDBA *InstanceConfigDevice `json:"sdba,omitzero"`
	SDBB *InstanceConfigDevice `json:"sdbb,omitzero"`
	SDBC *InstanceConfigDevice `json:"sdbc,omitzero"`
	SDBD *InstanceConfigDevice `json:"sdbd,omitzero"`
	SDBE *InstanceConfigDevice `json:"sdbe,omitzero"`
	SDBF *InstanceConfigDevice `json:"sdbf,omitzero"`
	SDBG *InstanceConfigDevice `json:"sdbg,omitzero"`
	SDBH *InstanceConfigDevice `json:"sdbh,omitzero"`
	SDBI *InstanceConfigDevice `json:"sdbi,omitzero"`
	SDBJ *InstanceConfigDevice `json:"sdbj,omitzero"`
	SDBK *InstanceConfigDevice `json:"sdbk,omitzero"`
	SDBL *InstanceConfigDevice `json:"sdbl,omitzero"`
}

// InstanceConfigHelpers are Instance Config options that control Linux distribution specific tweaks
type InstanceConfigHelpers struct {
	UpdateDBDisabled  bool `json:"updatedb_disabled"`
	Distro            bool `json:"distro"`
	ModulesDep        bool `json:"modules_dep"`
	Network           bool `json:"network"`
	DevTmpFsAutomount bool `json:"devtmpfs_automount"`
}

// ConfigInterfacePurpose options start with InterfacePurpose and include all known interface purpose types
type ConfigInterfacePurpose string

const (
	InterfacePurposePublic ConfigInterfacePurpose = "public"
	InterfacePurposeVLAN   ConfigInterfacePurpose = "vlan"
	InterfacePurposeVPC    ConfigInterfacePurpose = "vpc"
)

// InstanceConfigCreateOptions are InstanceConfig settings that can be used at creation
type InstanceConfigCreateOptions struct {
	Label       string                                 `json:"label,omitzero"`
	Comments    string                                 `json:"comments,omitzero"`
	Devices     InstanceConfigDeviceMap                `json:"devices"`
	Helpers     *InstanceConfigHelpers                 `json:"helpers,omitzero"`
	Interfaces  []InstanceConfigInterfaceCreateOptions `json:"interfaces"`
	MemoryLimit int                                    `json:"memory_limit,omitzero"`
	Kernel      string                                 `json:"kernel,omitzero"`
	InitRD      int                                    `json:"init_rd,omitzero"`
	RootDevice  *string                                `json:"root_device,omitzero"`
	RunLevel    string                                 `json:"run_level,omitzero"`
	VirtMode    string                                 `json:"virt_mode,omitzero"`
}

// InstanceConfigUpdateOptions are InstanceConfig settings that can be used in updates
type InstanceConfigUpdateOptions struct {
	Label      string                                 `json:"label,omitzero"`
	Comments   string                                 `json:"comments"`
	Devices    *InstanceConfigDeviceMap               `json:"devices,omitzero"`
	Helpers    *InstanceConfigHelpers                 `json:"helpers,omitzero"`
	Interfaces []InstanceConfigInterfaceCreateOptions `json:"interfaces"`
	// MemoryLimit 0 means unlimitted, this is not omitted
	MemoryLimit int    `json:"memory_limit"`
	Kernel      string `json:"kernel,omitzero"`
	// InitRD is nullable, permit the sending of null
	InitRD     *int   `json:"init_rd"`
	RootDevice string `json:"root_device,omitzero"`
	RunLevel   string `json:"run_level,omitzero"`
	VirtMode   string `json:"virt_mode,omitzero"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *InstanceConfig) UnmarshalJSON(b []byte) error {
	type Mask InstanceConfig

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.Updated = (*time.Time)(p.Updated)

	return nil
}

// GetCreateOptions converts a InstanceConfig to InstanceConfigCreateOptions for use in CreateInstanceConfig
func (i InstanceConfig) GetCreateOptions() InstanceConfigCreateOptions {
	result := InstanceConfigCreateOptions{
		Label:       i.Label,
		Comments:    i.Comments,
		Helpers:     i.Helpers,
		Interfaces:  getInstanceConfigInterfacesCreateOptionsList(i.Interfaces),
		MemoryLimit: i.MemoryLimit,
		Kernel:      i.Kernel,
		RootDevice:  copyString(&i.RootDevice),
		RunLevel:    i.RunLevel,
		VirtMode:    i.VirtMode,
	}

	if i.InitRD != nil {
		result.InitRD = *i.InitRD
	}

	if i.Devices != nil {
		result.Devices = *i.Devices
	}

	return result
}

// GetUpdateOptions converts a InstanceConfig to InstanceConfigUpdateOptions for use in UpdateInstanceConfig
func (i InstanceConfig) GetUpdateOptions() InstanceConfigUpdateOptions {
	return InstanceConfigUpdateOptions{
		Label:       i.Label,
		Comments:    i.Comments,
		Devices:     i.Devices,
		Helpers:     i.Helpers,
		Interfaces:  getInstanceConfigInterfacesCreateOptionsList(i.Interfaces),
		MemoryLimit: i.MemoryLimit,
		Kernel:      i.Kernel,
		InitRD:      copyInt(i.InitRD),
		RootDevice:  i.RootDevice,
		RunLevel:    i.RunLevel,
		VirtMode:    i.VirtMode,
	}
}

// ListInstanceConfigs lists InstanceConfigs
func (c *Client) ListInstanceConfigs(ctx context.Context, linodeID int, opts *ListOptions) ([]InstanceConfig, error) {
	return getPaginatedResults[InstanceConfig](ctx, c, formatAPIPath("linode/instances/%d/configs", linodeID), opts)
}

// GetInstanceConfig gets the template with the provided ID
func (c *Client) GetInstanceConfig(ctx context.Context, linodeID int, configID int) (*InstanceConfig, error) {
	e := formatAPIPath("linode/instances/%d/configs/%d", linodeID, configID)
	return doGETRequest[InstanceConfig](ctx, c, e)
}

// CreateInstanceConfig creates a new InstanceConfig for the given Instance
func (c *Client) CreateInstanceConfig(ctx context.Context, linodeID int, opts InstanceConfigCreateOptions) (*InstanceConfig, error) {
	e := formatAPIPath("linode/instances/%d/configs", linodeID)
	return doPOSTRequest[InstanceConfig](ctx, c, e, opts)
}

// UpdateInstanceConfig update an InstanceConfig for the given Instance
func (c *Client) UpdateInstanceConfig(ctx context.Context, linodeID int, configID int, opts InstanceConfigUpdateOptions) (*InstanceConfig, error) {
	e := formatAPIPath("linode/instances/%d/configs/%d", linodeID, configID)
	return doPUTRequest[InstanceConfig](ctx, c, e, opts)
}

// RenameInstanceConfig renames an InstanceConfig
func (c *Client) RenameInstanceConfig(ctx context.Context, linodeID int, configID int, label string) (*InstanceConfig, error) {
	return c.UpdateInstanceConfig(ctx, linodeID, configID, InstanceConfigUpdateOptions{Label: label})
}

// DeleteInstanceConfig deletes a Linode InstanceConfig
func (c *Client) DeleteInstanceConfig(ctx context.Context, linodeID int, configID int) error {
	e := formatAPIPath("linode/instances/%d/configs/%d", linodeID, configID)
	return doDELETERequest(ctx, c, e)
}
