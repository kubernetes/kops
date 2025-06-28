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

// PrimaryIP defines a Primary IP.
type PrimaryIP struct {
	ID           int64
	IP           net.IP
	Network      *net.IPNet
	Labels       map[string]string
	Name         string
	Type         PrimaryIPType
	Protection   PrimaryIPProtection
	DNSPtr       map[string]string
	AssigneeID   int64
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

// changeDNSPtr changes or resets the reverse DNS pointer for a IP address.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (p *PrimaryIP) changeDNSPtr(ctx context.Context, client *Client, ip net.IP, ptr *string) (*Action, *Response, error) {
	const opPath = "/primary_ips/%d/actions/change_dns_ptr"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, p.ID)

	reqBody := schema.PrimaryIPActionChangeDNSPtrRequest{
		IP:     ip.String(),
		DNSPtr: ptr,
	}

	respBody, resp, err := postRequest[schema.PrimaryIPActionChangeDNSPtrResponse](ctx, client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
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
	AssigneeID   *int64
	AssigneeType string
	AutoDelete   *bool
	Datacenter   string
	Labels       map[string]string
	Name         string
	Type         PrimaryIPType
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
	AutoDelete *bool
	Labels     *map[string]string
	Name       string
}

// PrimaryIPAssignOpts defines the request to
// assign a Primary IP to an assignee (usually a server).
type PrimaryIPAssignOpts struct {
	ID           int64
	AssigneeID   int64
	AssigneeType string
}

// Deprecated: Please use [schema.PrimaryIPActionAssignResponse] instead.
type PrimaryIPAssignResult = schema.PrimaryIPActionAssignResponse

// PrimaryIPChangeDNSPtrOpts defines the request to
// change a DNS PTR entry from a Primary IP.
type PrimaryIPChangeDNSPtrOpts struct {
	ID     int64
	DNSPtr string
	IP     string
}

// Deprecated: Please use [schema.PrimaryIPChangeDNSPtrResponse] instead.
type PrimaryIPChangeDNSPtrResult = schema.PrimaryIPActionChangeDNSPtrResponse

// PrimaryIPChangeProtectionOpts defines the request to
// change protection configuration of a Primary IP.
type PrimaryIPChangeProtectionOpts struct {
	ID     int64
	Delete bool
}

// Deprecated: Please use [schema.PrimaryIPActionChangeProtectionResponse] instead.
type PrimaryIPChangeProtectionResult = schema.PrimaryIPActionChangeProtectionResponse

// PrimaryIPClient is a client for the Primary IP API.
type PrimaryIPClient struct {
	client *Client
	Action *ResourceActionClient
}

// GetByID retrieves a Primary IP by its ID. If the Primary IP does not exist, nil is returned.
func (c *PrimaryIPClient) GetByID(ctx context.Context, id int64) (*PrimaryIP, *Response, error) {
	const opPath = "/primary_ips/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.PrimaryIPGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return PrimaryIPFromSchema(respBody.PrimaryIP), resp, nil
}

// GetByIP retrieves a Primary IP by its IP Address. If the Primary IP does not exist, nil is returned.
func (c *PrimaryIPClient) GetByIP(ctx context.Context, ip string) (*PrimaryIP, *Response, error) {
	if ip == "" {
		return nil, nil, nil
	}
	return firstBy(func() ([]*PrimaryIP, *Response, error) {
		return c.List(ctx, PrimaryIPListOpts{IP: ip})
	})
}

// GetByName retrieves a Primary IP by its name. If the Primary IP does not exist, nil is returned.
func (c *PrimaryIPClient) GetByName(ctx context.Context, name string) (*PrimaryIP, *Response, error) {
	return firstByName(name, func() ([]*PrimaryIP, *Response, error) {
		return c.List(ctx, PrimaryIPListOpts{Name: name})
	})
}

// Get retrieves a Primary IP by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Primary IP by its name. If the Primary IP does not exist, nil is returned.
func (c *PrimaryIPClient) Get(ctx context.Context, idOrName string) (*PrimaryIP, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// PrimaryIPListOpts specifies options for listing Primary IPs.
type PrimaryIPListOpts struct {
	ListOpts
	Name string
	IP   string
	Sort []string
}

func (l PrimaryIPListOpts) values() url.Values {
	vals := l.ListOpts.Values()
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
	const opPath = "/primary_ips?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.PrimaryIPListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.PrimaryIPs, PrimaryIPFromSchema), resp, nil
}

// All returns all Primary IPs.
func (c *PrimaryIPClient) All(ctx context.Context) ([]*PrimaryIP, error) {
	return c.AllWithOpts(ctx, PrimaryIPListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all Primary IPs for the given options.
func (c *PrimaryIPClient) AllWithOpts(ctx context.Context, opts PrimaryIPListOpts) ([]*PrimaryIP, error) {
	return iterPages(func(page int) ([]*PrimaryIP, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// Create creates a Primary IP.
func (c *PrimaryIPClient) Create(ctx context.Context, opts PrimaryIPCreateOpts) (*PrimaryIPCreateResult, *Response, error) {
	const opPath = "/primary_ips"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := &PrimaryIPCreateResult{}

	reqPath := opPath

	reqBody := SchemaFromPrimaryIPCreateOpts(opts)

	respBody, resp, err := postRequest[schema.PrimaryIPCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.PrimaryIP = PrimaryIPFromSchema(respBody.PrimaryIP)
	if respBody.Action != nil {
		result.Action = ActionFromSchema(*respBody.Action)
	}

	return result, resp, nil
}

// Delete deletes a Primary IP.
func (c *PrimaryIPClient) Delete(ctx context.Context, primaryIP *PrimaryIP) (*Response, error) {
	const opPath = "/primary_ips/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, primaryIP.ID)

	return deleteRequestNoResult(ctx, c.client, reqPath)
}

// Update updates a Primary IP.
func (c *PrimaryIPClient) Update(ctx context.Context, primaryIP *PrimaryIP, opts PrimaryIPUpdateOpts) (*PrimaryIP, *Response, error) {
	const opPath = "/primary_ips/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, primaryIP.ID)

	reqBody := SchemaFromPrimaryIPUpdateOpts(opts)

	respBody, resp, err := putRequest[schema.PrimaryIPUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return PrimaryIPFromSchema(respBody.PrimaryIP), resp, nil
}

// Assign a Primary IP to a resource.
func (c *PrimaryIPClient) Assign(ctx context.Context, opts PrimaryIPAssignOpts) (*Action, *Response, error) {
	const opPath = "/primary_ips/%d/actions/assign"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.ID)

	reqBody := SchemaFromPrimaryIPAssignOpts(opts)

	respBody, resp, err := postRequest[schema.PrimaryIPActionAssignResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// Unassign a Primary IP from a resource.
func (c *PrimaryIPClient) Unassign(ctx context.Context, id int64) (*Action, *Response, error) {
	const opPath = "/primary_ips/%d/actions/unassign"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := postRequest[schema.PrimaryIPActionUnassignResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// ChangeDNSPtr Change the reverse DNS from a Primary IP.
func (c *PrimaryIPClient) ChangeDNSPtr(ctx context.Context, opts PrimaryIPChangeDNSPtrOpts) (*Action, *Response, error) {
	const opPath = "/primary_ips/%d/actions/change_dns_ptr"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.ID)

	reqBody := SchemaFromPrimaryIPChangeDNSPtrOpts(opts)

	respBody, resp, err := postRequest[schema.PrimaryIPActionChangeDNSPtrResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// ChangeProtection Changes the protection configuration of a Primary IP.
func (c *PrimaryIPClient) ChangeProtection(ctx context.Context, opts PrimaryIPChangeProtectionOpts) (*Action, *Response, error) {
	const opPath = "/primary_ips/%d/actions/change_protection"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.ID)

	reqBody := SchemaFromPrimaryIPChangeProtectionOpts(opts)

	respBody, resp, err := postRequest[schema.PrimaryIPActionChangeProtectionResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}
