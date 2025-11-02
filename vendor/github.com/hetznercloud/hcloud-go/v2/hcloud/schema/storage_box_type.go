package schema

// StorageBoxType represents a Storage Box type as returned by the Hetzner Cloud API.
type StorageBoxType struct {
	ID                     int64                 `json:"id"`
	Name                   string                `json:"name"`
	Description            string                `json:"description"`
	SnapshotLimit          *int                  `json:"snapshot_limit"`
	AutomaticSnapshotLimit *int                  `json:"automatic_snapshot_limit"`
	SubaccountsLimit       int                   `json:"subaccounts_limit"`
	Size                   int64                 `json:"size"`
	Prices                 []StorageBoxTypePrice `json:"prices"`
	DeprecatableResource
}

// StorageBoxTypePrice represents pricing for a Storage Box type in a specific location.
type StorageBoxTypePrice struct {
	Location     string `json:"location"`
	PriceHourly  Price  `json:"price_hourly"`
	PriceMonthly Price  `json:"price_monthly"`
	SetupFee     Price  `json:"setup_fee"`
}

// StorageBoxTypeListResponse represents the response for listing Storage Box Types.
type StorageBoxTypeListResponse struct {
	StorageBoxTypes []StorageBoxType `json:"storage_box_types"`
}

// StorageBoxTypeGetResponse represents the response for getting a single Storage Box Type.
type StorageBoxTypeGetResponse struct {
	StorageBoxType StorageBoxType `json:"storage_box_type"`
}
