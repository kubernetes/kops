package schema

import "time"

type DeprecationInfo struct {
	Announced        time.Time `json:"announced"`
	UnavailableAfter time.Time `json:"unavailable_after"`
}

type DeprecatableResource struct {
	Deprecation *DeprecationInfo `json:"deprecation"`
}
