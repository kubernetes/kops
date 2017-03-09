package members

import (
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
)

// ImageMember model
type ImageMember struct {
	CreatedAt time.Time `mapstructure:"-"`
	ImageID   string    `mapstructure:"image_id"`
	MemberID  string    `mapstructure:"member_id"`
	Schema    string
	// Status could be one of pending, accepted, reject
	Status    string
	UpdatedAt time.Time `mapstructure:"-"`
}

// CreateMemberResult result model
type CreateMemberResult struct {
	gophercloud.Result
}

// Extract ImageMember model from request if possible
func (cm CreateMemberResult) Extract() (*ImageMember, error) {
	if cm.Err != nil {
		return nil, cm.Err
	}
	casted := cm.Body.(map[string]interface{})
	var results ImageMember

	if err := mapstructure.Decode(casted, &results); err != nil {
		return nil, err
	}

	if t, ok := casted["created_at"].(string); ok && t != "" {
		createdAt, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return &results, err
		}
		results.CreatedAt = createdAt
	}

	if t, ok := casted["updated_at"].(string); ok && t != "" {
		updatedAt, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return &results, err
		}
		results.UpdatedAt = updatedAt
	}

	return &results, nil
}

// ListMembersResult model
type ListMembersResult struct {
	gophercloud.Result
}

// Extract returns list of image members
func (lm ListMembersResult) Extract() ([]ImageMember, error) {
	if lm.Err != nil {
		return nil, lm.Err
	}
	casted := lm.Body.(map[string]interface{})

	var results struct {
		ImageMembers []ImageMember `mapstructure:"members"`
	}

	err := mapstructure.Decode(casted, &results)
	return results.ImageMembers, err
}

// MemberPage is a single page of Members results.
type MemberPage struct {
	pagination.SinglePageBase
}

// ExtractMembers returns a slice of Members contained in a single page of results.
func ExtractMembers(page pagination.Page) ([]ImageMember, error) {
	casted := page.(MemberPage).Body
	var response struct {
		ImageMembers []ImageMember `mapstructure:"members"`
	}

	err := mapstructure.Decode(casted, &response)
	return response.ImageMembers, err
}

// IsEmpty determines whether or not a page of Members contains any results.
func (page MemberPage) IsEmpty() (bool, error) {
	tenants, err := ExtractMembers(page)
	if err != nil {
		return false, err
	}
	return len(tenants) == 0, nil
}

// MemberDetailsResult model
type MemberDetailsResult struct {
	CreateMemberResult
}

// MemberDeleteResult model
type MemberDeleteResult struct {
	gophercloud.Result
}

// MemberUpdateResult model
type MemberUpdateResult struct {
	CreateMemberResult
}
