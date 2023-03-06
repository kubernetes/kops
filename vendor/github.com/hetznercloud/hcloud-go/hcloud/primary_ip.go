package hcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

// PrimaryIP defines a Primary IP.
type PrimaryIP struct {
	ID           int
	IP           net.IP
	Network      *net.IPNet
	Labels       map[string]string
	Name         string
	Type         PrimaryIPType
	Protection   PrimaryIPProtection
	DNSPtr       map[string]string
	AssigneeID   int
	AssigneeType string
	AutoDelete   bool
	Blocked      bool
	Created      time.Time
	Datacenter   *Datacenter
}

// PrimaryIPProtection represents the protection level of a Primary IP.
type PrimaryIPProtection struct {
	Delete bool
}

// PrimaryIPDNSPTR contains reverse DNS information for a
// IPv4 or IPv6 Primary IP.
type PrimaryIPDNSPTR struct {
	DNSPtr string
	IP     string
}

// GetDNSPtrForIP searches for the dns assigned to the given IP address.
// It returns an error if there is no dns set for the given IP address.
func (p *PrimaryIP) GetDNSPtrForIP(ip net.IP) (string, error) {
	dns, ok := p.DNSPtr[ip.String()]
	if !ok {
		return "", DNSNotFoundError{ip}
	}

	return dns, nil
}

// PrimaryIPType represents the type of Primary IP.
type PrimaryIPType string

// PrimaryIPType Primary IP types.
const (
	PrimaryIPTypeIPv4 PrimaryIPType = "ipv4"
	PrimaryIPTypeIPv6 PrimaryIPType = "ipv6"
)

// PrimaryIPCreateOpts defines the request to
// create a Primary IP.
type PrimaryIPCreateOpts struct {
	AssigneeID   *int              `json:"assignee_id,omitempty"`
	AssigneeType string            `json:"assignee_type"`
	AutoDelete   *bool             `json:"auto_delete,omitempty"`
	Datacenter   string            `json:"datacenter,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Name         string            `json:"name"`
	Type         PrimaryIPType     `json:"type"`
}

// PrimaryIPCreateResult defines the response
// when creating a Primary IP.
type PrimaryIPCreateResult struct {
	PrimaryIP *PrimaryIP
	Action    *Action
}

// PrimaryIPUpdateOpts defines the request to
// update a Primary IP.
type PrimaryIPUpdateOpts struct {
	AutoDelete *bool              `json:"auto_delete,omitempty"`
	Labels     *map[string]string `json:"labels,omitempty"`
	Name       string             `json:"name,omitempty"`
}

// PrimaryIPAssignOpts defines the request to
// assign a Primary IP to an assignee (usually a server).
type PrimaryIPAssignOpts struct {
	ID           int
	AssigneeID   int    `json:"assignee_id"`
	AssigneeType string `json:"assignee_type"`
}

// PrimaryIPAssignResult defines the response
// when assigning a Primary IP to a assignee.
type PrimaryIPAssignResult struct {
	Action schema.Action `json:"action"`
}

// PrimaryIPChangeDNSPtrOpts defines the request to
// change a DNS PTR entry from a Primary IP.
type PrimaryIPChangeDNSPtrOpts struct {
	ID     int
	DNSPtr string `json:"dns_ptr"`
	IP     string `json:"ip"`
}

// PrimaryIPChangeDNSPtrResult defines the response
// when assigning a Primary IP to a assignee.
type PrimaryIPChangeDNSPtrResult struct {
	Action schema.Action `json:"action"`
}

// PrimaryIPChangeProtectionOpts defines the request to
// change protection configuration of a Primary IP.
type PrimaryIPChangeProtectionOpts struct {
	ID     int
	Delete bool `json:"delete"`
}

// PrimaryIPChangeProtectionResult defines the response
// when changing a protection of a PrimaryIP.
type PrimaryIPChangeProtectionResult struct {
	Action schema.Action `json:"action"`
}

// PrimaryIPClient is a client for the Primary IP API.
type PrimaryIPClient struct {
	client *Client
}

// GetByID retrieves a Primary IP by its ID. If the Primary IP does not exist, nil is returned.
func (c *PrimaryIPClient) GetByID(ctx context.Context, id int) (*PrimaryIP, *Response, error) {
	req, err := c.client.NewRequest(ctx, "GET", fmt.Sprintf("/primary_ips/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.PrimaryIPGetResult
	resp, err := c.client.Do(req, &body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return PrimaryIPFromSchema(body.PrimaryIP), resp, nil
}

// GetByIP retrieves a Primary IP by its IP Address. If the Primary IP does not exist, nil is returned.
func (c *PrimaryIPClient) GetByIP(ctx context.Context, ip string) (*PrimaryIP, *Response, error) {
	if ip == "" {
		return nil, nil, nil
	}
	primaryIPs, response, err := c.List(ctx, PrimaryIPListOpts{IP: ip})
	if len(primaryIPs) == 0 {
		return nil, response, err
	}
	return primaryIPs[0], response, err
}

// GetByName retrieves a Primary IP by its name. If the Primary IP does not exist, nil is returned.
func (c *PrimaryIPClient) GetByName(ctx context.Context, name string) (*PrimaryIP, *Response, error) {
	if name == "" {
		return nil, nil, nil
	}
	primaryIPs, response, err := c.List(ctx, PrimaryIPListOpts{Name: name})
	if len(primaryIPs) == 0 {
		return nil, response, err
	}
	return primaryIPs[0], response, err
}

// Get retrieves a Primary IP by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Primary IP by its name. If the Primary IP does not exist, nil is returned.
func (c *PrimaryIPClient) Get(ctx context.Context, idOrName string) (*PrimaryIP, *Response, error) {
	if id, err := strconv.Atoi(idOrName); err == nil {
		return c.GetByID(ctx, int(id))
	}
	return c.GetByName(ctx, idOrName)
}

// PrimaryIPListOpts specifies options for listing Primary IPs.
type PrimaryIPListOpts struct {
	ListOpts
	Name string
	IP   string
	Sort []string
}

func (l PrimaryIPListOpts) values() url.Values {
	vals := l.ListOpts.values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	if l.IP != "" {
		vals.Add("ip", l.IP)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of Primary IPs for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *PrimaryIPClient) List(ctx context.Context, opts PrimaryIPListOpts) ([]*PrimaryIP, *Response, error) {
	path := "/primary_ips?" + opts.values().Encode()
	req, err := c.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	var body schema.PrimaryIPListResult
	resp, err := c.client.Do(req, &body)
	if err != nil {
		return nil, nil, err
	}
	primaryIPs := make([]*PrimaryIP, 0, len(body.PrimaryIPs))
	for _, s := range body.PrimaryIPs {
		primaryIPs = append(primaryIPs, PrimaryIPFromSchema(s))
	}
	return primaryIPs, resp, nil
}

// All returns all Primary IPs.
func (c *PrimaryIPClient) All(ctx context.Context) ([]*PrimaryIP, error) {
	return c.AllWithOpts(ctx, PrimaryIPListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all Primary IPs for the given options.
func (c *PrimaryIPClient) AllWithOpts(ctx context.Context, opts PrimaryIPListOpts) ([]*PrimaryIP, error) {
	var allPrimaryIPs []*PrimaryIP

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		primaryIPs, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allPrimaryIPs = append(allPrimaryIPs, primaryIPs...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allPrimaryIPs, nil
}

// Create creates a Primary IP.
func (c *PrimaryIPClient) Create(ctx context.Context, reqBody PrimaryIPCreateOpts) (*PrimaryIPCreateResult, *Response, error) {
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return &PrimaryIPCreateResult{}, nil, err
	}

	req, err := c.client.NewRequest(ctx, "POST", "/primary_ips", bytes.NewReader(reqBodyData))
	if err != nil {
		return &PrimaryIPCreateResult{}, nil, err
	}

	var respBody schema.PrimaryIPCreateResponse
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return &PrimaryIPCreateResult{}, resp, err
	}
	var action *Action
	if respBody.Action != nil {
		action = ActionFromSchema(*respBody.Action)
	}
	primaryIP := PrimaryIPFromSchema(respBody.PrimaryIP)
	return &PrimaryIPCreateResult{
		PrimaryIP: primaryIP,
		Action:    action,
	}, resp, nil
}

// Delete deletes a Primary IP.
func (c *PrimaryIPClient) Delete(ctx context.Context, primaryIP *PrimaryIP) (*Response, error) {
	req, err := c.client.NewRequest(ctx, "DELETE", fmt.Sprintf("/primary_ips/%d", primaryIP.ID), nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req, nil)
}

// Update updates a Primary IP.
func (c *PrimaryIPClient) Update(ctx context.Context, primaryIP *PrimaryIP, reqBody PrimaryIPUpdateOpts) (*PrimaryIP, *Response, error) {
	reqBodyData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/primary_ips/%d", primaryIP.ID)
	req, err := c.client.NewRequest(ctx, "PUT", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody schema.PrimaryIPUpdateResult
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return PrimaryIPFromSchema(respBody.PrimaryIP), resp, nil
}

// Assign a Primary IP to a resource.
func (c *PrimaryIPClient) Assign(ctx context.Context, opts PrimaryIPAssignOpts) (*Action, *Response, error) {
	reqBodyData, err := json.Marshal(opts)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/primary_ips/%d/actions/assign", opts.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody PrimaryIPAssignResult
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// Unassign a Primary IP from a resource.
func (c *PrimaryIPClient) Unassign(ctx context.Context, id int) (*Action, *Response, error) {
	path := fmt.Sprintf("/primary_ips/%d/actions/unassign", id)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader([]byte{}))
	if err != nil {
		return nil, nil, err
	}

	var respBody PrimaryIPAssignResult
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// ChangeDNSPtr Change the reverse DNS from a Primary IP.
func (c *PrimaryIPClient) ChangeDNSPtr(ctx context.Context, opts PrimaryIPChangeDNSPtrOpts) (*Action, *Response, error) {
	reqBodyData, err := json.Marshal(opts)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/primary_ips/%d/actions/change_dns_ptr", opts.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody PrimaryIPChangeDNSPtrResult
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}

// ChangeProtection Changes the protection configuration of a Primary IP.
func (c *PrimaryIPClient) ChangeProtection(ctx context.Context, opts PrimaryIPChangeProtectionOpts) (*Action, *Response, error) {
	reqBodyData, err := json.Marshal(opts)
	if err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("/primary_ips/%d/actions/change_protection", opts.ID)
	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
	if err != nil {
		return nil, nil, err
	}

	var respBody PrimaryIPChangeProtectionResult
	resp, err := c.client.Do(req, &respBody)
	if err != nil {
		return nil, resp, err
	}
	return ActionFromSchema(respBody.Action), resp, nil
}
