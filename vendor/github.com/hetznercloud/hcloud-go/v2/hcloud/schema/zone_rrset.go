package schema

type ZoneRRSet struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Type       string              `json:"type"`
	TTL        *int                `json:"ttl"`
	Labels     map[string]string   `json:"labels"`
	Protection ZoneRRSetProtection `json:"protection"`
	Records    []ZoneRRSetRecord   `json:"records"`
	Zone       int64               `json:"zone"`
}

type ZoneRRSetProtection struct {
	Change bool `json:"change"`
}

type ZoneRRSetRecord struct {
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

type ZoneRRSetGetResponse struct {
	RRSet ZoneRRSet `json:"rrset"`
}

type ZoneRRSetListResponse struct {
	RRSets []ZoneRRSet `json:"rrsets"`
}

type ZoneRRSetCreateRequest struct {
	Name    string             `json:"name"`
	Type    string             `json:"type"`
	TTL     *int               `json:"ttl,omitempty"`
	Labels  *map[string]string `json:"labels,omitempty"`
	Records []ZoneRRSetRecord  `json:"records,omitempty"`
}

type ZoneRRSetCreateResponse struct {
	RRSet  ZoneRRSet `json:"rrset"`
	Action Action    `json:"action"`
}

type ZoneRRSetUpdateRequest struct {
	Labels *map[string]string `json:"labels,omitempty"`
}

type ZoneRRSetUpdateResponse struct {
	RRSet ZoneRRSet `json:"rrset"`
}

type ZoneRRSetChangeProtectionRequest struct {
	Change *bool `json:"change,omitempty"`
}

type ZoneRRSetChangeTTLRequest struct {
	TTL *int `json:"ttl"`
}

type ZoneRRSetSetRecordsRequest struct {
	Records []ZoneRRSetRecord `json:"records"`
}

type ZoneRRSetAddRecordsRequest struct {
	Records []ZoneRRSetRecord `json:"records"`
	TTL     *int              `json:"ttl,omitempty"`
}

type ZoneRRSetRemoveRecordsRequest struct {
	Records []ZoneRRSetRecord `json:"records"`
}
