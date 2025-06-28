package schema

// ISO defines the schema of an ISO image.
type ISO struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Type         string  `json:"type"`
	Architecture *string `json:"architecture"`
	DeprecatableResource
}

// ISOGetResponse defines the schema of the response when retrieving a single ISO.
type ISOGetResponse struct {
	ISO ISO `json:"iso"`
}

// ISOListResponse defines the schema of the response when listing ISOs.
type ISOListResponse struct {
	ISOs []ISO `json:"isos"`
}
