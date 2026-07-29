package schema

// Datacenter defines the schema of a datacenter.
//
// Deprecated: [Datacenter] is deprecated and will be removed after the 2026-10-01. See
// https://docs.hetzner.cloud/changelog#2026-06-02-datacenters-deprecated.
type Datacenter struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Location    Location `json:"location"`

	// Deprecated: [Datacenter.ServerTypes] is deprecated and will not be returned after 2026-10-01.
	// Use [ServerType.Locations] instead.
	ServerTypes *DatacenterServerTypes `json:"server_types,omitempty"`
}

// DatacenterServerTypes defines the schema of the server types available in a datacenter.
//
// Deprecated: [DatacenterServerTypes] is deprecated and will not be returned after 2026-10-01.
// Use [ServerType.Locations] instead.
type DatacenterServerTypes struct {
	Supported             []int64 `json:"supported"`
	AvailableForMigration []int64 `json:"available_for_migration"`
	Available             []int64 `json:"available"`
}

// DatacenterGetResponse defines the schema of the response when retrieving a single datacenter.
//
// Deprecated: [Datacenter] is deprecated and will be removed after the 2026-10-01. See
// https://docs.hetzner.cloud/changelog#2026-06-02-datacenters-deprecated.
type DatacenterGetResponse struct {
	Datacenter Datacenter `json:"datacenter"`
}

// DatacenterListResponse defines the schema of the response when listing datacenters.
//
// Deprecated: [Datacenter] is deprecated and will be removed after the 2026-10-01. See
// https://docs.hetzner.cloud/changelog#2026-06-02-datacenters-deprecated.
type DatacenterListResponse struct {
	Datacenters []Datacenter `json:"datacenters"`
}
