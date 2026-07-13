package linodego

import "context"

// LinodeEntity is anything in Linode with IAM Permissions
type LinodeEntity struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// ListEntities returns a paginated list of all entities
func (c *Client) ListEntities(ctx context.Context, opts *ListOptions) ([]LinodeEntity, error) {
	return getPaginatedResults[LinodeEntity](ctx, c, "entities", opts)
}

// GetEntityRoles returns a list of roles for the entity and user
func (c *Client) GetEntityRoles(ctx context.Context, username string, entityType string, entityID int) ([]string, error) {
	perms, err := doGETRequest[[]string](ctx, c,
		formatAPIPath("iam/users/%s/permissions/%s/%d", username, entityType, entityID))
	if err != nil || perms == nil {
		return nil, err
	}

	return (*perms), err
}
