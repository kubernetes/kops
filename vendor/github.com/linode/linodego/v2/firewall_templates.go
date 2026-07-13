package linodego

import (
	"context"
)

type FirewallTemplate struct {
	Slug  string        `json:"slug"`
	Rules FirewallRules `json:"rules"`
}

// GetFirewallTemplate gets a FirewallTemplate given a slug.
// NOTE: This feature may not currently be available to all users.
func (c *Client) GetFirewallTemplate(ctx context.Context, slug string) (*FirewallTemplate, error) {
	e := formatAPIPath("networking/firewalls/templates/%s", slug)
	return doGETRequest[FirewallTemplate](ctx, c, e)
}

// ListFirewallTemplates gets all available firewall templates for the account.
// NOTE: This feature may not currently be available to all users.
func (c *Client) ListFirewallTemplates(ctx context.Context, opts *ListOptions) ([]FirewallTemplate, error) {
	return getPaginatedResults[FirewallTemplate](ctx, c, "networking/firewalls/templates", opts)
}
