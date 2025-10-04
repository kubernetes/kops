package zones

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/pagination"
)

type commonResult struct {
	gophercloud.Result
}

// Extract interprets a GetResult, CreateResult or UpdateResult as a Zone.
// An error is returned if the original call or the extraction failed.
func (r commonResult) Extract() (*Zone, error) {
	var s *Zone
	err := r.ExtractInto(&s)
	return s, err
}

// CreateResult is the result of a Create request. Call its Extract method
// to interpret the result as a Zone.
type CreateResult struct {
	commonResult
}

// GetResult is the result of a Get request. Call its Extract method
// to interpret the result as a Zone.
type GetResult struct {
	commonResult
}

// UpdateResult is the result of an Update request. Call its Extract method
// to interpret the result as a Zone.
type UpdateResult struct {
	commonResult
}

// DeleteResult is the result of a Delete request. Call its ExtractErr method
// to determine if the request succeeded or failed.
type DeleteResult struct {
	commonResult
}

// ZonePage is a single page of Zone results.
type ZonePage struct {
	pagination.LinkedPageBase
}

// ErrResult represents a generic error result.
type ErrResult struct {
	gophercloud.ErrResult
}

// IsEmpty returns true if the page contains no results.
func (r ZonePage) IsEmpty() (bool, error) {
	if r.StatusCode == 204 {
		return true, nil
	}

	s, err := ExtractZones(r)
	return len(s) == 0, err
}

// ExtractZones extracts a slice of Zones from a List result.
func ExtractZones(r pagination.Page) ([]Zone, error) {
	var s struct {
		Zones []Zone `json:"zones"`
	}
	err := (r.(ZonePage)).ExtractInto(&s)
	return s.Zones, err
}

// Zone represents a DNS zone.
type Zone struct {
	// ID uniquely identifies this zone amongst all other zones, including those
	// not accessible to the current tenant.
	ID string `json:"id"`

	// PoolID is the ID for the pool hosting this zone.
	PoolID string `json:"pool_id"`

	// ProjectID identifies the project/tenant owning this resource.
	ProjectID string `json:"project_id"`

	// Name is the DNS Name for the zone.
	Name string `json:"name"`

	// Email for the zone. Used in SOA records for the zone.
	Email string `json:"email"`

	// Description for this zone.
	Description string `json:"description"`

	// TTL is the Time to Live for the zone.
	TTL int `json:"ttl"`

	// Serial is the current serial number for the zone.
	Serial int `json:"-"`

	// Status is the status of the resource.
	Status string `json:"status"`

	// Action is the current action in progress on the resource.
	Action string `json:"action"`

	// Version of the resource.
	Version int `json:"version"`

	// Attributes for the zone.
	Attributes map[string]string `json:"attributes"`

	// Type of zone. Primary is controlled by Designate.
	// Secondary zones are slaved from another DNS Server.
	// Defaults to Primary.
	Type string `json:"type"`

	// Masters is the servers for slave servers to get DNS information from.
	Masters []string `json:"masters"`

	// CreatedAt is the date when the zone was created.
	CreatedAt time.Time `json:"-"`

	// UpdatedAt is the date when the last change was made to the zone.
	UpdatedAt time.Time `json:"-"`

	// TransferredAt is the last time an update was retrieved from the
	// master servers.
	TransferredAt time.Time `json:"-"`

	// Links includes HTTP references to the itself, useful for passing along
	// to other APIs that might want a server reference.
	Links map[string]any `json:"links"`
}

func (r *Zone) UnmarshalJSON(b []byte) error {
	type tmp Zone
	var s struct {
		tmp
		CreatedAt     gophercloud.JSONRFC3339MilliNoZ `json:"created_at"`
		UpdatedAt     gophercloud.JSONRFC3339MilliNoZ `json:"updated_at"`
		TransferredAt gophercloud.JSONRFC3339MilliNoZ `json:"transferred_at"`
		Serial        any                             `json:"serial"`
	}
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = Zone(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.UpdatedAt = time.Time(s.UpdatedAt)
	r.TransferredAt = time.Time(s.TransferredAt)

	switch t := s.Serial.(type) {
	case float64:
		r.Serial = int(t)
	case string:
		switch t {
		case "":
			r.Serial = 0
		default:
			serial, err := strconv.ParseFloat(t, 64)
			if err != nil {
				return err
			}
			r.Serial = int(serial)
		}
	}

	return err
}

// ZoneShare represents a shared zone.
type ZoneShare struct {
	// ID uniquely identifies this zone share.
	ID string `json:"id"`

	// ZoneID is the ID of the zone being shared.
	ZoneID string `json:"zone_id"`

	// ProjectID is the ID of the project with which the zone is shared.
	ProjectID string `json:"project_id"`

	// TargetProjectID is the ID of the project with which the zone is shared.
	TargetProjectID string `json:"target_project_id"`

	// CreatedAt is the date when the zone share was created.
	CreatedAt time.Time `json:"-"`

	// UpdatedAt is the date when the zone share was last updated.
	UpdatedAt time.Time `json:"-"`
}

func (r *ZoneShare) UnmarshalJSON(b []byte) error {
	type tmp ZoneShare
	var s struct {
		tmp
		CreatedAt gophercloud.JSONRFC3339MilliNoZ `json:"created_at"`
		UpdatedAt gophercloud.JSONRFC3339MilliNoZ `json:"updated_at"`
	}

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	*r = ZoneShare(s.tmp)

	r.CreatedAt = time.Time(s.CreatedAt)
	r.UpdatedAt = time.Time(s.UpdatedAt)

	return nil
}

// ZoneShareResult is the result of a GetZoneShare request.
type ZoneShareResult struct {
	gophercloud.Result
}

// Extract interprets a GetResult, CreateResult or UpdateResult as a Zone.
// An error is returned if the original call or the extraction failed.
func (r ZoneShareResult) Extract() (*ZoneShare, error) {
	var s *ZoneShare
	err := r.ExtractInto(&s)
	return s, err
}

// ZoneSharePage is a single page of ZoneShare results.
type ZoneSharePage struct {
	pagination.LinkedPageBase
}

// IsEmpty returns true if the page contains no results.
func (r ZoneSharePage) IsEmpty() (bool, error) {
	if r.StatusCode == 204 {
		return true, nil
	}

	s, err := ExtractZoneShares(r)
	return len(s) == 0, err
}

// ExtractZoneShares extracts a slice of ZoneShares from a List result.
func ExtractZoneShares(r pagination.Page) ([]ZoneShare, error) {
	var s struct {
		ZoneShares []ZoneShare `json:"shared_zones"`
	}
	err := (r.(ZoneSharePage)).ExtractInto(&s)
	return s.ZoneShares, err
}
