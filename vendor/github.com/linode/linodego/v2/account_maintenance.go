package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// AccountMaintenance represents a Maintenance object for any entity a user has permissions to view
type AccountMaintenance struct {
	Entity *Entity `json:"entity"`
	Reason string  `json:"reason"`
	Status string  `json:"status"`
	Type   string  `json:"type"`

	MaintenancePolicySet string `json:"maintenance_policy_set"`

	Description  string     `json:"description"`
	Source       string     `json:"source"`
	NotBefore    *time.Time `json:"-"`
	StartTime    *time.Time `json:"-"`
	CompleteTime *time.Time `json:"-"`
}

// Entity represents the entity being affected by maintenance
type Entity struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (accountMaintenance *AccountMaintenance) UnmarshalJSON(b []byte) error {
	type Mask AccountMaintenance

	p := struct {
		*Mask

		NotBefore    *parseabletime.ParseableTime `json:"not_before"`
		StartTime    *parseabletime.ParseableTime `json:"start_time"`
		CompleteTime *parseabletime.ParseableTime `json:"complete_time"`
		When         *parseabletime.ParseableTime `json:"when"`
	}{
		Mask: (*Mask)(accountMaintenance),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	accountMaintenance.NotBefore = (*time.Time)(p.NotBefore)
	accountMaintenance.StartTime = (*time.Time)(p.StartTime)
	accountMaintenance.CompleteTime = (*time.Time)(p.CompleteTime)

	return nil
}

// ListMaintenances lists Account Maintenance objects for any entity a user has permissions to view
func (c *Client) ListMaintenances(ctx context.Context, opts *ListOptions) ([]AccountMaintenance, error) {
	return getPaginatedResults[AccountMaintenance](ctx, c, "account/maintenance", opts)
}
