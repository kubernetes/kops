package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// DomainRecord represents a DomainRecord object
type DomainRecord struct {
	ID       int              `json:"id"`
	Type     DomainRecordType `json:"type"`
	Name     string           `json:"name"`
	Target   string           `json:"target"`
	Priority int              `json:"priority"`
	Weight   int              `json:"weight"`
	Port     int              `json:"port"`
	Service  *string          `json:"service"`
	Protocol *string          `json:"protocol"`
	TTLSec   int              `json:"ttl_sec"`
	Tag      *string          `json:"tag"`
	Created  *time.Time       `json:"-"`
	Updated  *time.Time       `json:"-"`
}

// DomainRecordCreateOptions fields are those accepted by CreateDomainRecord
type DomainRecordCreateOptions struct {
	Type     DomainRecordType `json:"type"`
	Name     string           `json:"name"`
	Target   string           `json:"target"`
	Priority *int             `json:"priority,omitzero"`
	Weight   *int             `json:"weight,omitzero"`
	Port     *int             `json:"port,omitzero"`
	Service  *string          `json:"service,omitzero"`
	Protocol *string          `json:"protocol,omitzero"`
	TTLSec   int              `json:"ttl_sec,omitzero"` // 0 is not accepted by Linode, so can be omitted
	Tag      *string          `json:"tag,omitzero"`
}

// DomainRecordUpdateOptions fields are those accepted by UpdateDomainRecord
type DomainRecordUpdateOptions struct {
	Type     DomainRecordType `json:"type,omitzero"`
	Name     string           `json:"name,omitzero"`
	Target   string           `json:"target,omitzero"`
	Priority *int             `json:"priority,omitzero"` // 0 is valid, so omit only nil values
	Weight   *int             `json:"weight,omitzero"`   // 0 is valid, so omit only nil values
	Port     *int             `json:"port,omitzero"`     // 0 is valid to spec, so omit only nil values
	Service  *string          `json:"service,omitzero"`
	Protocol *string          `json:"protocol,omitzero"`
	TTLSec   int              `json:"ttl_sec,omitzero"` // 0 is not accepted by Linode, so can be omitted
	Tag      *string          `json:"tag,omitzero"`
}

// DomainRecordType constants start with RecordType and include Linode API Domain Record Types
type DomainRecordType string

// DomainRecordType constants are the DNS record types a DomainRecord can assign
const (
	RecordTypeA     DomainRecordType = "A"
	RecordTypeAAAA  DomainRecordType = "AAAA"
	RecordTypeNS    DomainRecordType = "NS"
	RecordTypeMX    DomainRecordType = "MX"
	RecordTypeCNAME DomainRecordType = "CNAME"
	RecordTypeTXT   DomainRecordType = "TXT"
	RecordTypeSRV   DomainRecordType = "SRV"
	RecordTypePTR   DomainRecordType = "PTR"
	RecordTypeCAA   DomainRecordType = "CAA"
)

// UnmarshalJSON for DomainRecord responses
func (d *DomainRecord) UnmarshalJSON(b []byte) error {
	type Mask DomainRecord

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(d),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	d.Created = (*time.Time)(p.Created)
	d.Updated = (*time.Time)(p.Updated)

	return nil
}

// GetUpdateOptions converts a DomainRecord to DomainRecordUpdateOptions for use in UpdateDomainRecord
func (d DomainRecord) GetUpdateOptions() (du DomainRecordUpdateOptions) {
	du.Type = d.Type
	du.Name = d.Name
	du.Target = d.Target
	du.Priority = copyInt(&d.Priority)
	du.Weight = copyInt(&d.Weight)
	du.Port = copyInt(&d.Port)
	du.Service = copyString(d.Service)
	du.Protocol = copyString(d.Protocol)
	du.TTLSec = d.TTLSec
	du.Tag = copyString(d.Tag)

	return du
}

// ListDomainRecords lists DomainRecords
func (c *Client) ListDomainRecords(ctx context.Context, domainID int, opts *ListOptions) ([]DomainRecord, error) {
	return getPaginatedResults[DomainRecord](ctx, c, formatAPIPath("domains/%d/records", domainID), opts)
}

// GetDomainRecord gets the domainrecord with the provided ID
func (c *Client) GetDomainRecord(ctx context.Context, domainID int, recordID int) (*DomainRecord, error) {
	e := formatAPIPath("domains/%d/records/%d", domainID, recordID)
	return doGETRequest[DomainRecord](ctx, c, e)
}

// CreateDomainRecord creates a DomainRecord
func (c *Client) CreateDomainRecord(ctx context.Context, domainID int, opts DomainRecordCreateOptions) (*DomainRecord, error) {
	e := formatAPIPath("domains/%d/records", domainID)
	return doPOSTRequest[DomainRecord](ctx, c, e, opts)
}

// UpdateDomainRecord updates the DomainRecord with the specified id
func (c *Client) UpdateDomainRecord(ctx context.Context, domainID int, recordID int, opts DomainRecordUpdateOptions) (*DomainRecord, error) {
	e := formatAPIPath("domains/%d/records/%d", domainID, recordID)
	return doPUTRequest[DomainRecord](ctx, c, e, opts)
}

// DeleteDomainRecord deletes the DomainRecord with the specified id
func (c *Client) DeleteDomainRecord(ctx context.Context, domainID int, recordID int) error {
	e := formatAPIPath("domains/%d/records/%d", domainID, recordID)
	return doDELETERequest(ctx, c, e)
}
