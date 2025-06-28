package hcloud

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// Firewall represents a Firewall in the Hetzner Cloud.
type Firewall struct {
	ID        int64
	Name      string
	Labels    map[string]string
	Created   time.Time
	Rules     []FirewallRule
	AppliedTo []FirewallResource
}

// FirewallRule represents a Firewall's rules.
type FirewallRule struct {
	Direction      FirewallRuleDirection
	SourceIPs      []net.IPNet
	DestinationIPs []net.IPNet
	Protocol       FirewallRuleProtocol
	Port           *string
	Description    *string
}

// FirewallRuleDirection specifies the direction of a Firewall rule.
type FirewallRuleDirection string

const (
	// FirewallRuleDirectionIn specifies a rule for inbound traffic.
	FirewallRuleDirectionIn FirewallRuleDirection = "in"

	// FirewallRuleDirectionOut specifies a rule for outbound traffic.
	FirewallRuleDirectionOut FirewallRuleDirection = "out"
)

// FirewallRuleProtocol specifies the protocol of a Firewall rule.
type FirewallRuleProtocol string

const (
	// FirewallRuleProtocolTCP specifies a TCP rule.
	FirewallRuleProtocolTCP FirewallRuleProtocol = "tcp"
	// FirewallRuleProtocolUDP specifies a UDP rule.
	FirewallRuleProtocolUDP FirewallRuleProtocol = "udp"
	// FirewallRuleProtocolICMP specifies an ICMP rule.
	FirewallRuleProtocolICMP FirewallRuleProtocol = "icmp"
	// FirewallRuleProtocolESP specifies an esp rule.
	FirewallRuleProtocolESP FirewallRuleProtocol = "esp"
	// FirewallRuleProtocolGRE specifies an gre rule.
	FirewallRuleProtocolGRE FirewallRuleProtocol = "gre"
)

// FirewallResourceType specifies the resource to apply a Firewall on.
type FirewallResourceType string

const (
	// FirewallResourceTypeServer specifies a Server.
	FirewallResourceTypeServer FirewallResourceType = "server"
	// FirewallResourceTypeLabelSelector specifies a LabelSelector.
	FirewallResourceTypeLabelSelector FirewallResourceType = "label_selector"
)

// FirewallResource represents a resource to apply the new Firewall on.
type FirewallResource struct {
	Type          FirewallResourceType
	Server        *FirewallResourceServer
	LabelSelector *FirewallResourceLabelSelector
}

// FirewallResourceServer represents a Server to apply a Firewall on.
type FirewallResourceServer struct {
	ID int64
}

// FirewallResourceLabelSelector represents a LabelSelector to apply a Firewall on.
type FirewallResourceLabelSelector struct {
	Selector string
}

// FirewallClient is a client for the Firewalls API.
type FirewallClient struct {
	client *Client
	Action *ResourceActionClient
}

// GetByID retrieves a Firewall by its ID. If the Firewall does not exist, nil is returned.
func (c *FirewallClient) GetByID(ctx context.Context, id int64) (*Firewall, *Response, error) {
	const opPath = "/firewalls/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.FirewallGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return FirewallFromSchema(respBody.Firewall), resp, nil
}

// GetByName retrieves a Firewall by its name. If the Firewall does not exist, nil is returned.
func (c *FirewallClient) GetByName(ctx context.Context, name string) (*Firewall, *Response, error) {
	return firstByName(name, func() ([]*Firewall, *Response, error) {
		return c.List(ctx, FirewallListOpts{Name: name})
	})
}

// Get retrieves a Firewall by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Firewall by its name. If the Firewall does not exist, nil is returned.
func (c *FirewallClient) Get(ctx context.Context, idOrName string) (*Firewall, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// FirewallListOpts specifies options for listing Firewalls.
type FirewallListOpts struct {
	ListOpts
	Name string
	Sort []string
}

func (l FirewallListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of Firewalls for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *FirewallClient) List(ctx context.Context, opts FirewallListOpts) ([]*Firewall, *Response, error) {
	const opPath = "/firewalls?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.FirewallListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Firewalls, FirewallFromSchema), resp, nil
}

// All returns all Firewalls.
func (c *FirewallClient) All(ctx context.Context) ([]*Firewall, error) {
	return c.AllWithOpts(ctx, FirewallListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all Firewalls for the given options.
func (c *FirewallClient) AllWithOpts(ctx context.Context, opts FirewallListOpts) ([]*Firewall, error) {
	return iterPages(func(page int) ([]*Firewall, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// FirewallCreateOpts specifies options for creating a new Firewall.
type FirewallCreateOpts struct {
	Name    string
	Labels  map[string]string
	Rules   []FirewallRule
	ApplyTo []FirewallResource
}

// Validate checks if options are valid.
func (o FirewallCreateOpts) Validate() error {
	if o.Name == "" {
		return missingField(o, "Name")
	}
	return nil
}

// FirewallCreateResult is the result of a create Firewall call.
type FirewallCreateResult struct {
	Firewall *Firewall
	Actions  []*Action
}

// Create creates a new Firewall.
func (c *FirewallClient) Create(ctx context.Context, opts FirewallCreateOpts) (FirewallCreateResult, *Response, error) {
	const opPath = "/firewalls"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := FirewallCreateResult{}

	reqPath := opPath

	if err := opts.Validate(); err != nil {
		return result, nil, err
	}

	reqBody := firewallCreateOptsToSchema(opts)

	respBody, resp, err := postRequest[schema.FirewallCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Firewall = FirewallFromSchema(respBody.Firewall)
	result.Actions = ActionsFromSchema(respBody.Actions)

	return result, resp, nil
}

// FirewallUpdateOpts specifies options for updating a Firewall.
type FirewallUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a Firewall.
func (c *FirewallClient) Update(ctx context.Context, firewall *Firewall, opts FirewallUpdateOpts) (*Firewall, *Response, error) {
	const opPath = "/firewalls/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, firewall.ID)

	reqBody := schema.FirewallUpdateRequest{}
	if opts.Name != "" {
		reqBody.Name = &opts.Name
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := putRequest[schema.FirewallUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return FirewallFromSchema(respBody.Firewall), resp, nil
}

// Delete deletes a Firewall.
func (c *FirewallClient) Delete(ctx context.Context, firewall *Firewall) (*Response, error) {
	const opPath = "/firewalls/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, firewall.ID)

	return deleteRequestNoResult(ctx, c.client, reqPath)
}

// FirewallSetRulesOpts specifies options for setting rules of a Firewall.
type FirewallSetRulesOpts struct {
	Rules []FirewallRule
}

// SetRules sets the rules of a Firewall.
func (c *FirewallClient) SetRules(ctx context.Context, firewall *Firewall, opts FirewallSetRulesOpts) ([]*Action, *Response, error) {
	const opPath = "/firewalls/%d/actions/set_rules"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, firewall.ID)

	reqBody := firewallSetRulesOptsToSchema(opts)

	respBody, resp, err := postRequest[schema.FirewallActionSetRulesResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionsFromSchema(respBody.Actions), resp, nil
}

func (c *FirewallClient) ApplyResources(ctx context.Context, firewall *Firewall, resources []FirewallResource) ([]*Action, *Response, error) {
	const opPath = "/firewalls/%d/actions/apply_to_resources"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, firewall.ID)

	applyTo := make([]schema.FirewallResource, len(resources))
	for i, r := range resources {
		applyTo[i] = firewallResourceToSchema(r)
	}

	reqBody := schema.FirewallActionApplyToResourcesRequest{ApplyTo: applyTo}

	respBody, resp, err := postRequest[schema.FirewallActionApplyToResourcesResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionsFromSchema(respBody.Actions), resp, nil
}

func (c *FirewallClient) RemoveResources(ctx context.Context, firewall *Firewall, resources []FirewallResource) ([]*Action, *Response, error) {
	const opPath = "/firewalls/%d/actions/remove_from_resources"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, firewall.ID)

	removeFrom := make([]schema.FirewallResource, len(resources))
	for i, r := range resources {
		removeFrom[i] = firewallResourceToSchema(r)
	}

	reqBody := schema.FirewallActionRemoveFromResourcesRequest{RemoveFrom: removeFrom}

	respBody, resp, err := postRequest[schema.FirewallActionRemoveFromResourcesResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionsFromSchema(respBody.Actions), resp, nil
}
