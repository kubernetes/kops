package schema

import "time"

type Zone struct {
	ID                 int64                   `json:"id"`
	Name               string                  `json:"name"`
	Created            time.Time               `json:"created"`
	TTL                int                     `json:"ttl"`
	Mode               string                  `json:"mode"`
	PrimaryNameservers []ZonePrimaryNameserver `json:"primary_nameservers"`
	Protection         ZoneProtection          `json:"protection"`
	Labels             map[string]string       `json:"labels"`

	AuthoritativeNameservers ZoneAuthoritativeNameservers `json:"authoritative_nameservers"`
	Registrar                string                       `json:"registrar"`
	Status                   string                       `json:"status"`
	RecordCount              int                          `json:"record_count"`
}

type ZoneProtection struct {
	Delete bool `json:"delete"`
}

type ZonePrimaryNameserver struct {
	Address       string `json:"address"`
	Port          int    `json:"port"`
	TSIGAlgorithm string `json:"tsig_algorithm"`
	TSIGKey       string `json:"tsig_key"`
}

type ZoneAuthoritativeNameservers struct {
	Assigned            []string  `json:"assigned"`
	Delegated           []string  `json:"delegated"`
	DelegationLastCheck time.Time `json:"delegation_last_check"`
	DelegationStatus    string    `json:"delegation_status"`
}

type ZoneListResponse struct {
	Zones []Zone `json:"zones"`
}

type ZoneCreateRequest struct {
	Name               string                               `json:"name"`
	Mode               string                               `json:"mode"`
	TTL                *int                                 `json:"ttl,omitempty"`
	Labels             *map[string]string                   `json:"labels,omitempty"`
	PrimaryNameservers []ZoneCreateRequestPrimaryNameserver `json:"primary_nameservers,omitempty"`
	RRSets             []ZoneCreateRequestRRSet             `json:"rrsets,omitempty"`
	Zonefile           string                               `json:"zonefile,omitempty"`
}

type ZoneCreateRequestPrimaryNameserver struct {
	Address       string `json:"address"`
	Port          int    `json:"port,omitempty"`
	TSIGAlgorithm string `json:"tsig_algorithm,omitempty"`
	TSIGKey       string `json:"tsig_key,omitempty"`
}

type ZoneCreateRequestRRSet struct {
	Type    string             `json:"type"`
	Name    string             `json:"name"`
	TTL     *int               `json:"ttl,omitempty"`
	Labels  *map[string]string `json:"labels,omitempty"`
	Records []ZoneRRSetRecord  `json:"records,omitempty"`
}

type ZoneCreateResponse struct {
	Zone   Zone   `json:"zone"`
	Action Action `json:"action"`
}

type ZoneGetResponse struct {
	Zone Zone `json:"zone"`
}

type ZoneUpdateRequest struct {
	Labels *map[string]string `json:"labels,omitempty"`
}

type ZoneUpdateResponse struct {
	Zone Zone `json:"zone"`
}

type ZoneExportZonefileResponse struct {
	Zonefile string `json:"zonefile"`
}

type ZoneChangeProtectionRequest struct {
	Delete *bool `json:"delete,omitempty"`
}

type ZoneChangePrimaryNameserversRequest struct {
	PrimaryNameservers []ZoneChangePrimaryNameserversRequestPrimaryNameserver `json:"primary_nameservers"`
}

type ZoneChangePrimaryNameserversRequestPrimaryNameserver struct {
	Address       string `json:"address"`
	Port          int    `json:"port,omitempty"`
	TSIGAlgorithm string `json:"tsig_algorithm,omitempty"`
	TSIGKey       string `json:"tsig_key,omitempty"`
}

type ZoneChangeTTLRequest struct {
	TTL int `json:"ttl"`
}

type ZoneImportZonefileRequest struct {
	Zonefile string `json:"zonefile"`
}
