package members

import (
	"encoding/json"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// Member model
type Member struct {
	CreatedAt time.Time `json:"-"`
	ImageID   string    `json:"image_id"`
	MemberID  string    `json:"member_id"`
	Schema    string    `json:"schema"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"-"`
}

func (s *Member) UnmarshalJSON(b []byte) error {
	type tmp Member
	var p *struct {
		tmp
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	err := json.Unmarshal(b, &p)
	if err != nil {
		return err
	}

	*s = Member(p.tmp)
	s.CreatedAt, err = time.Parse(time.RFC3339, p.CreatedAt)
	if err != nil {
		return err
	}
	s.UpdatedAt, err = time.Parse(time.RFC3339, p.UpdatedAt)
	return err
}

// Extract Member model from request if possible
func (r commonResult) Extract() (*Member, error) {
	var s *Member
	err := r.ExtractInto(&s)
	return s, err
}

// MemberPage is a single page of Members results.
type MemberPage struct {
	pagination.SinglePageBase
}

// ExtractMembers returns a slice of Members contained in a single page of results.
func ExtractMembers(r pagination.Page) ([]Member, error) {
	var s struct {
		Members []Member `json:"members"`
	}
	err := r.(MemberPage).ExtractInto(&s)
	return s.Members, err
}

// IsEmpty determines whether or not a page of Members contains any results.
func (r MemberPage) IsEmpty() (bool, error) {
	members, err := ExtractMembers(r)
	return len(members) == 0, err
}

type commonResult struct {
	gophercloud.Result
}

// CreateResult result model
type CreateResult struct {
	commonResult
}

// DetailsResult model
type DetailsResult struct {
	commonResult
}

// UpdateResult model
type UpdateResult struct {
	commonResult
}

// DeleteResult model
type DeleteResult struct {
	gophercloud.ErrResult
}
