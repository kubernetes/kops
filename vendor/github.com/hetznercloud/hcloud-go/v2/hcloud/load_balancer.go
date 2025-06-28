package hcloud

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/schema"
)

// LoadBalancer represents a Load Balancer in the Hetzner Cloud.
type LoadBalancer struct {
	ID               int64
	Name             string
	PublicNet        LoadBalancerPublicNet
	PrivateNet       []LoadBalancerPrivateNet
	Location         *Location
	LoadBalancerType *LoadBalancerType
	Algorithm        LoadBalancerAlgorithm
	Services         []LoadBalancerService
	Targets          []LoadBalancerTarget
	Protection       LoadBalancerProtection
	Labels           map[string]string
	Created          time.Time
	IncludedTraffic  uint64
	OutgoingTraffic  uint64
	IngoingTraffic   uint64
}

// LoadBalancerPublicNet represents a Load Balancer's public network.
type LoadBalancerPublicNet struct {
	Enabled bool
	IPv4    LoadBalancerPublicNetIPv4
	IPv6    LoadBalancerPublicNetIPv6
}

// LoadBalancerPublicNetIPv4 represents a Load Balancer's public IPv4 address.
type LoadBalancerPublicNetIPv4 struct {
	IP     net.IP
	DNSPtr string
}

// LoadBalancerPublicNetIPv6 represents a Load Balancer's public IPv6 address.
type LoadBalancerPublicNetIPv6 struct {
	IP     net.IP
	DNSPtr string
}

// LoadBalancerPrivateNet represents a Load Balancer's private network.
type LoadBalancerPrivateNet struct {
	Network *Network
	IP      net.IP
}

// LoadBalancerService represents a Load Balancer service.
type LoadBalancerService struct {
	Protocol        LoadBalancerServiceProtocol
	ListenPort      int
	DestinationPort int
	Proxyprotocol   bool
	HTTP            LoadBalancerServiceHTTP
	HealthCheck     LoadBalancerServiceHealthCheck
}

// LoadBalancerServiceHTTP stores configuration for a service using the HTTP protocol.
type LoadBalancerServiceHTTP struct {
	CookieName     string
	CookieLifetime time.Duration
	Certificates   []*Certificate
	RedirectHTTP   bool
	StickySessions bool
}

// LoadBalancerServiceHealthCheck stores configuration for a service health check.
type LoadBalancerServiceHealthCheck struct {
	Protocol LoadBalancerServiceProtocol
	Port     int
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
	HTTP     *LoadBalancerServiceHealthCheckHTTP
}

// LoadBalancerServiceHealthCheckHTTP stores configuration for a service health check
// using the HTTP protocol.
type LoadBalancerServiceHealthCheckHTTP struct {
	Domain      string
	Path        string
	Response    string
	StatusCodes []string
	TLS         bool
}

// LoadBalancerAlgorithmType specifies the algorithm type a Load Balancer
// uses for distributing requests.
type LoadBalancerAlgorithmType string

const (
	// LoadBalancerAlgorithmTypeRoundRobin is an algorithm which distributes
	// requests to targets in a round-robin fashion.
	LoadBalancerAlgorithmTypeRoundRobin LoadBalancerAlgorithmType = "round_robin"
	// LoadBalancerAlgorithmTypeLeastConnections is an algorithm which distributes
	// requests to targets with the least number of connections.
	LoadBalancerAlgorithmTypeLeastConnections LoadBalancerAlgorithmType = "least_connections"
)

// LoadBalancerAlgorithm configures the algorithm a Load Balancer uses
// for distributing requests.
type LoadBalancerAlgorithm struct {
	Type LoadBalancerAlgorithmType
}

// LoadBalancerTargetType specifies the type of Load Balancer target.
type LoadBalancerTargetType string

const (
	// LoadBalancerTargetTypeServer is a target type which points to a specific
	// server.
	LoadBalancerTargetTypeServer LoadBalancerTargetType = "server"

	// LoadBalancerTargetTypeLabelSelector is a target type which selects the
	// servers a Load Balancer points to using labels assigned to the servers.
	LoadBalancerTargetTypeLabelSelector LoadBalancerTargetType = "label_selector"

	// LoadBalancerTargetTypeIP is a target type which points to an IP.
	LoadBalancerTargetTypeIP LoadBalancerTargetType = "ip"
)

// LoadBalancerServiceProtocol specifies the protocol of a Load Balancer service.
type LoadBalancerServiceProtocol string

const (
	// LoadBalancerServiceProtocolTCP specifies a TCP service.
	LoadBalancerServiceProtocolTCP LoadBalancerServiceProtocol = "tcp"
	// LoadBalancerServiceProtocolHTTP specifies an HTTP service.
	LoadBalancerServiceProtocolHTTP LoadBalancerServiceProtocol = "http"
	// LoadBalancerServiceProtocolHTTPS specifies an HTTPS service.
	LoadBalancerServiceProtocolHTTPS LoadBalancerServiceProtocol = "https"
)

// LoadBalancerTarget represents a Load Balancer target.
type LoadBalancerTarget struct {
	Type          LoadBalancerTargetType
	Server        *LoadBalancerTargetServer
	LabelSelector *LoadBalancerTargetLabelSelector
	IP            *LoadBalancerTargetIP
	HealthStatus  []LoadBalancerTargetHealthStatus
	Targets       []LoadBalancerTarget
	UsePrivateIP  bool
}

// LoadBalancerTargetServer configures a Load Balancer target
// pointing at a specific server.
type LoadBalancerTargetServer struct {
	Server *Server
}

// LoadBalancerTargetLabelSelector configures a Load Balancer target pointing
// at the servers matching the selector. This includes the target pointing at
// nothing, if no servers match the Selector.
type LoadBalancerTargetLabelSelector struct {
	Selector string
}

// LoadBalancerTargetIP configures a Load Balancer target pointing to a Hetzner
// Online IP address.
type LoadBalancerTargetIP struct {
	IP string
}

// LoadBalancerTargetHealthStatusStatus describes a target's health status.
type LoadBalancerTargetHealthStatusStatus string

const (
	// LoadBalancerTargetHealthStatusStatusUnknown denotes that the health status is unknown.
	LoadBalancerTargetHealthStatusStatusUnknown LoadBalancerTargetHealthStatusStatus = "unknown"
	// LoadBalancerTargetHealthStatusStatusHealthy denotes a healthy target.
	LoadBalancerTargetHealthStatusStatusHealthy LoadBalancerTargetHealthStatusStatus = "healthy"
	// LoadBalancerTargetHealthStatusStatusUnhealthy denotes an unhealthy target.
	LoadBalancerTargetHealthStatusStatusUnhealthy LoadBalancerTargetHealthStatusStatus = "unhealthy"
)

// LoadBalancerTargetHealthStatus describes a target's health for a specific service.
type LoadBalancerTargetHealthStatus struct {
	ListenPort int
	Status     LoadBalancerTargetHealthStatusStatus
}

// LoadBalancerProtection represents the protection level of a Load Balancer.
type LoadBalancerProtection struct {
	Delete bool
}

// changeDNSPtr changes or resets the reverse DNS pointer for an IP address.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (lb *LoadBalancer) changeDNSPtr(ctx context.Context, client *Client, ip net.IP, ptr *string) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/change_dns_ptr"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, lb.ID)

	reqBody := schema.LoadBalancerActionChangeDNSPtrRequest{
		IP:     ip.String(),
		DNSPtr: ptr,
	}

	respBody, resp, err := postRequest[schema.LoadBalancerActionChangeDNSPtrResponse](ctx, client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// GetDNSPtrForIP searches for the dns assigned to the given IP address.
// It returns an error if there is no dns set for the given IP address.
func (lb *LoadBalancer) GetDNSPtrForIP(ip net.IP) (string, error) {
	if net.IP.Equal(lb.PublicNet.IPv4.IP, ip) {
		return lb.PublicNet.IPv4.DNSPtr, nil
	} else if net.IP.Equal(lb.PublicNet.IPv6.IP, ip) {
		return lb.PublicNet.IPv6.DNSPtr, nil
	}

	return "", DNSNotFoundError{ip}
}

// LoadBalancerClient is a client for the Load Balancers API.
type LoadBalancerClient struct {
	client *Client
	Action *ResourceActionClient
}

// GetByID retrieves a Load Balancer by its ID. If the Load Balancer does not exist, nil is returned.
func (c *LoadBalancerClient) GetByID(ctx context.Context, id int64) (*LoadBalancer, *Response, error) {
	const opPath = "/load_balancers/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.LoadBalancerGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return LoadBalancerFromSchema(respBody.LoadBalancer), resp, nil
}

// GetByName retrieves a Load Balancer by its name. If the Load Balancer does not exist, nil is returned.
func (c *LoadBalancerClient) GetByName(ctx context.Context, name string) (*LoadBalancer, *Response, error) {
	return firstByName(name, func() ([]*LoadBalancer, *Response, error) {
		return c.List(ctx, LoadBalancerListOpts{Name: name})
	})
}

// Get retrieves a Load Balancer by its ID if the input can be parsed as an integer, otherwise it
// retrieves a Load Balancer by its name. If the Load Balancer does not exist, nil is returned.
func (c *LoadBalancerClient) Get(ctx context.Context, idOrName string) (*LoadBalancer, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// LoadBalancerListOpts specifies options for listing Load Balancers.
type LoadBalancerListOpts struct {
	ListOpts
	Name string
	Sort []string
}

func (l LoadBalancerListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of Load Balancers for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *LoadBalancerClient) List(ctx context.Context, opts LoadBalancerListOpts) ([]*LoadBalancer, *Response, error) {
	const opPath = "/load_balancers?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.LoadBalancerListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.LoadBalancers, LoadBalancerFromSchema), resp, nil
}

// All returns all Load Balancers.
func (c *LoadBalancerClient) All(ctx context.Context) ([]*LoadBalancer, error) {
	return c.AllWithOpts(ctx, LoadBalancerListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all Load Balancers for the given options.
func (c *LoadBalancerClient) AllWithOpts(ctx context.Context, opts LoadBalancerListOpts) ([]*LoadBalancer, error) {
	return iterPages(func(page int) ([]*LoadBalancer, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// LoadBalancerUpdateOpts specifies options for updating a Load Balancer.
type LoadBalancerUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a Load Balancer.
func (c *LoadBalancerClient) Update(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerUpdateOpts) (*LoadBalancer, *Response, error) {
	const opPath = "/load_balancers/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := schema.LoadBalancerUpdateRequest{}
	if opts.Name != "" {
		reqBody.Name = &opts.Name
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := putRequest[schema.LoadBalancerUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return LoadBalancerFromSchema(respBody.LoadBalancer), resp, nil
}

// LoadBalancerCreateOpts specifies options for creating a new Load Balancer.
type LoadBalancerCreateOpts struct {
	Name             string
	LoadBalancerType *LoadBalancerType
	Algorithm        *LoadBalancerAlgorithm
	Location         *Location
	NetworkZone      NetworkZone
	Labels           map[string]string
	Targets          []LoadBalancerCreateOptsTarget
	Services         []LoadBalancerCreateOptsService
	PublicInterface  *bool
	Network          *Network
}

// LoadBalancerCreateOptsTarget holds options for specifying a target
// when creating a new Load Balancer.
type LoadBalancerCreateOptsTarget struct {
	Type          LoadBalancerTargetType
	Server        LoadBalancerCreateOptsTargetServer
	LabelSelector LoadBalancerCreateOptsTargetLabelSelector
	IP            LoadBalancerCreateOptsTargetIP
	UsePrivateIP  *bool
}

// LoadBalancerCreateOptsTargetServer holds options for specifying a server target
// when creating a new Load Balancer.
type LoadBalancerCreateOptsTargetServer struct {
	Server *Server
}

// LoadBalancerCreateOptsTargetLabelSelector holds options for specifying a label selector target
// when creating a new Load Balancer.
type LoadBalancerCreateOptsTargetLabelSelector struct {
	Selector string
}

// LoadBalancerCreateOptsTargetIP holds options for specifying an IP target
// when creating a new Load Balancer.
type LoadBalancerCreateOptsTargetIP struct {
	IP string
}

// LoadBalancerCreateOptsService holds options for specifying a service
// when creating a new Load Balancer.
type LoadBalancerCreateOptsService struct {
	Protocol        LoadBalancerServiceProtocol
	ListenPort      *int
	DestinationPort *int
	Proxyprotocol   *bool
	HTTP            *LoadBalancerCreateOptsServiceHTTP
	HealthCheck     *LoadBalancerCreateOptsServiceHealthCheck
}

// LoadBalancerCreateOptsServiceHTTP holds options for specifying an HTTP service
// when creating a new Load Balancer.
type LoadBalancerCreateOptsServiceHTTP struct {
	CookieName     *string
	CookieLifetime *time.Duration
	Certificates   []*Certificate
	RedirectHTTP   *bool
	StickySessions *bool
}

// LoadBalancerCreateOptsServiceHealthCheck holds options for specifying a service
// health check when creating a new Load Balancer.
type LoadBalancerCreateOptsServiceHealthCheck struct {
	Protocol LoadBalancerServiceProtocol
	Port     *int
	Interval *time.Duration
	Timeout  *time.Duration
	Retries  *int
	HTTP     *LoadBalancerCreateOptsServiceHealthCheckHTTP
}

// LoadBalancerCreateOptsServiceHealthCheckHTTP holds options for specifying a service
// HTTP health check when creating a new Load Balancer.
type LoadBalancerCreateOptsServiceHealthCheckHTTP struct {
	Domain      *string
	Path        *string
	Response    *string
	StatusCodes []string
	TLS         *bool
}

// LoadBalancerCreateResult is the result of a create Load Balancer call.
type LoadBalancerCreateResult struct {
	LoadBalancer *LoadBalancer
	Action       *Action
}

// Create creates a new Load Balancer.
func (c *LoadBalancerClient) Create(ctx context.Context, opts LoadBalancerCreateOpts) (LoadBalancerCreateResult, *Response, error) {
	const opPath = "/load_balancers"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := LoadBalancerCreateResult{}

	reqPath := opPath

	reqBody := loadBalancerCreateOptsToSchema(opts)

	respBody, resp, err := postRequest[schema.LoadBalancerCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.LoadBalancer = LoadBalancerFromSchema(respBody.LoadBalancer)
	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, nil
}

// Delete deletes a Load Balancer.
func (c *LoadBalancerClient) Delete(ctx context.Context, loadBalancer *LoadBalancer) (*Response, error) {
	const opPath = "/load_balancers/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	return deleteRequestNoResult(ctx, c.client, reqPath)
}

func (c *LoadBalancerClient) addTarget(ctx context.Context, loadBalancer *LoadBalancer, reqBody schema.LoadBalancerActionAddTargetRequest) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/add_target"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	respBody, resp, err := postRequest[schema.LoadBalancerActionAddTargetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

func (c *LoadBalancerClient) removeTarget(ctx context.Context, loadBalancer *LoadBalancer, reqBody schema.LoadBalancerActionRemoveTargetRequest) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/remove_target"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	respBody, resp, err := postRequest[schema.LoadBalancerActionRemoveTargetResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerAddServerTargetOpts specifies options for adding a server target
// to a Load Balancer.
type LoadBalancerAddServerTargetOpts struct {
	Server       *Server
	UsePrivateIP *bool
}

// AddServerTarget adds a server target to a Load Balancer.
func (c *LoadBalancerClient) AddServerTarget(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerAddServerTargetOpts) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionAddTargetRequest{
		Type: string(LoadBalancerTargetTypeServer),
		Server: &schema.LoadBalancerActionAddTargetRequestServer{
			ID: opts.Server.ID,
		},
		UsePrivateIP: opts.UsePrivateIP,
	}
	return c.addTarget(ctx, loadBalancer, reqBody)
}

// RemoveServerTarget removes a server target from a Load Balancer.
func (c *LoadBalancerClient) RemoveServerTarget(ctx context.Context, loadBalancer *LoadBalancer, server *Server) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionRemoveTargetRequest{
		Type: string(LoadBalancerTargetTypeServer),
		Server: &schema.LoadBalancerActionRemoveTargetRequestServer{
			ID: server.ID,
		},
	}
	return c.removeTarget(ctx, loadBalancer, reqBody)
}

// LoadBalancerAddLabelSelectorTargetOpts specifies options for adding a label selector target
// to a Load Balancer.
type LoadBalancerAddLabelSelectorTargetOpts struct {
	Selector     string
	UsePrivateIP *bool
}

// AddLabelSelectorTarget adds a label selector target to a Load Balancer.
func (c *LoadBalancerClient) AddLabelSelectorTarget(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerAddLabelSelectorTargetOpts) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionAddTargetRequest{
		Type: string(LoadBalancerTargetTypeLabelSelector),
		LabelSelector: &schema.LoadBalancerActionAddTargetRequestLabelSelector{
			Selector: opts.Selector,
		},
		UsePrivateIP: opts.UsePrivateIP,
	}
	return c.addTarget(ctx, loadBalancer, reqBody)
}

// RemoveLabelSelectorTarget removes a label selector target from a Load Balancer.
func (c *LoadBalancerClient) RemoveLabelSelectorTarget(ctx context.Context, loadBalancer *LoadBalancer, labelSelector string) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionRemoveTargetRequest{
		Type: string(LoadBalancerTargetTypeLabelSelector),
		LabelSelector: &schema.LoadBalancerActionRemoveTargetRequestLabelSelector{
			Selector: labelSelector,
		},
	}
	return c.removeTarget(ctx, loadBalancer, reqBody)
}

// LoadBalancerAddIPTargetOpts specifies options for adding an IP target to a
// Load Balancer.
type LoadBalancerAddIPTargetOpts struct {
	IP net.IP
}

// AddIPTarget adds an IP target to a Load Balancer.
func (c *LoadBalancerClient) AddIPTarget(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerAddIPTargetOpts) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionAddTargetRequest{
		Type: string(LoadBalancerTargetTypeIP),
		IP:   &schema.LoadBalancerActionAddTargetRequestIP{IP: opts.IP.String()},
	}
	return c.addTarget(ctx, loadBalancer, reqBody)
}

// RemoveIPTarget removes an IP target from a Load Balancer.
func (c *LoadBalancerClient) RemoveIPTarget(ctx context.Context, loadBalancer *LoadBalancer, ip net.IP) (*Action, *Response, error) {
	reqBody := schema.LoadBalancerActionRemoveTargetRequest{
		Type: string(LoadBalancerTargetTypeIP),
		IP: &schema.LoadBalancerActionRemoveTargetRequestIP{
			IP: ip.String(),
		},
	}
	return c.removeTarget(ctx, loadBalancer, reqBody)
}

// LoadBalancerAddServiceOpts specifies options for adding a service to a Load Balancer.
type LoadBalancerAddServiceOpts struct {
	Protocol        LoadBalancerServiceProtocol
	ListenPort      *int
	DestinationPort *int
	Proxyprotocol   *bool
	HTTP            *LoadBalancerAddServiceOptsHTTP
	HealthCheck     *LoadBalancerAddServiceOptsHealthCheck
}

// LoadBalancerAddServiceOptsHTTP holds options for specifying an HTTP service
// when adding a service to a Load Balancer.
type LoadBalancerAddServiceOptsHTTP struct {
	CookieName     *string
	CookieLifetime *time.Duration
	Certificates   []*Certificate
	RedirectHTTP   *bool
	StickySessions *bool
}

// LoadBalancerAddServiceOptsHealthCheck holds options for specifying a health check
// when adding a service to a Load Balancer.
type LoadBalancerAddServiceOptsHealthCheck struct {
	Protocol LoadBalancerServiceProtocol
	Port     *int
	Interval *time.Duration
	Timeout  *time.Duration
	Retries  *int
	HTTP     *LoadBalancerAddServiceOptsHealthCheckHTTP
}

// LoadBalancerAddServiceOptsHealthCheckHTTP holds options for specifying an
// HTTP health check when adding a service to a Load Balancer.
type LoadBalancerAddServiceOptsHealthCheckHTTP struct {
	Domain      *string
	Path        *string
	Response    *string
	StatusCodes []string
	TLS         *bool
}

// AddService adds a service to a Load Balancer.
func (c *LoadBalancerClient) AddService(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerAddServiceOpts) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/add_service"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := loadBalancerAddServiceOptsToSchema(opts)

	respBody, resp, err := postRequest[schema.LoadBalancerActionAddServiceResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerUpdateServiceOpts specifies options for updating a service.
type LoadBalancerUpdateServiceOpts struct {
	Protocol        LoadBalancerServiceProtocol
	DestinationPort *int
	Proxyprotocol   *bool
	HTTP            *LoadBalancerUpdateServiceOptsHTTP
	HealthCheck     *LoadBalancerUpdateServiceOptsHealthCheck
}

// LoadBalancerUpdateServiceOptsHTTP specifies options for updating an HTTP(S) service.
type LoadBalancerUpdateServiceOptsHTTP struct {
	CookieName     *string
	CookieLifetime *time.Duration
	Certificates   []*Certificate
	RedirectHTTP   *bool
	StickySessions *bool
}

// LoadBalancerUpdateServiceOptsHealthCheck specifies options for updating
// a service's health check.
type LoadBalancerUpdateServiceOptsHealthCheck struct {
	Protocol LoadBalancerServiceProtocol
	Port     *int
	Interval *time.Duration
	Timeout  *time.Duration
	Retries  *int
	HTTP     *LoadBalancerUpdateServiceOptsHealthCheckHTTP
}

// LoadBalancerUpdateServiceOptsHealthCheckHTTP specifies options for updating
// the HTTP-specific settings of a service's health check.
type LoadBalancerUpdateServiceOptsHealthCheckHTTP struct {
	Domain      *string
	Path        *string
	Response    *string
	StatusCodes []string
	TLS         *bool
}

// UpdateService updates a Load Balancer service.
func (c *LoadBalancerClient) UpdateService(ctx context.Context, loadBalancer *LoadBalancer, listenPort int, opts LoadBalancerUpdateServiceOpts) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/update_service"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := loadBalancerUpdateServiceOptsToSchema(opts)
	reqBody.ListenPort = listenPort

	respBody, resp, err := postRequest[schema.LoadBalancerActionUpdateServiceResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// DeleteService deletes a Load Balancer service.
func (c *LoadBalancerClient) DeleteService(ctx context.Context, loadBalancer *LoadBalancer, listenPort int) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/delete_service"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := schema.LoadBalancerDeleteServiceRequest{
		ListenPort: listenPort,
	}

	respBody, resp, err := postRequest[schema.LoadBalancerDeleteServiceResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerChangeProtectionOpts specifies options for changing the resource protection level of a Load Balancer.
type LoadBalancerChangeProtectionOpts struct {
	Delete *bool
}

// ChangeProtection changes the resource protection level of a Load Balancer.
func (c *LoadBalancerClient) ChangeProtection(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerChangeProtectionOpts) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/change_protection"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := schema.LoadBalancerActionChangeProtectionRequest{
		Delete: opts.Delete,
	}

	respBody, resp, err := postRequest[schema.LoadBalancerActionChangeProtectionResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerChangeAlgorithmOpts specifies options for changing the algorithm of a Load Balancer.
type LoadBalancerChangeAlgorithmOpts struct {
	Type LoadBalancerAlgorithmType
}

// ChangeAlgorithm changes the algorithm of a Load Balancer.
func (c *LoadBalancerClient) ChangeAlgorithm(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerChangeAlgorithmOpts) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/change_algorithm"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := schema.LoadBalancerActionChangeAlgorithmRequest{
		Type: string(opts.Type),
	}

	respBody, resp, err := postRequest[schema.LoadBalancerActionChangeAlgorithmResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerAttachToNetworkOpts specifies options for attaching a Load Balancer to a network.
type LoadBalancerAttachToNetworkOpts struct {
	Network *Network
	IP      net.IP
}

// AttachToNetwork attaches a Load Balancer to a network.
func (c *LoadBalancerClient) AttachToNetwork(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerAttachToNetworkOpts) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/attach_to_network"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := schema.LoadBalancerActionAttachToNetworkRequest{
		Network: opts.Network.ID,
	}
	if opts.IP != nil {
		reqBody.IP = Ptr(opts.IP.String())
	}

	respBody, resp, err := postRequest[schema.LoadBalancerActionAttachToNetworkResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerDetachFromNetworkOpts specifies options for detaching a Load Balancer from a network.
type LoadBalancerDetachFromNetworkOpts struct {
	Network *Network
}

// DetachFromNetwork detaches a Load Balancer from a network.
func (c *LoadBalancerClient) DetachFromNetwork(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerDetachFromNetworkOpts) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/detach_from_network"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := schema.LoadBalancerActionDetachFromNetworkRequest{
		Network: opts.Network.ID,
	}

	respBody, resp, err := postRequest[schema.LoadBalancerActionDetachFromNetworkResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// EnablePublicInterface enables the Load Balancer's public network interface.
func (c *LoadBalancerClient) EnablePublicInterface(ctx context.Context, loadBalancer *LoadBalancer) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/enable_public_interface"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	respBody, resp, err := postRequest[schema.LoadBalancerActionEnablePublicInterfaceResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// DisablePublicInterface disables the Load Balancer's public network interface.
func (c *LoadBalancerClient) DisablePublicInterface(ctx context.Context, loadBalancer *LoadBalancer) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/disable_public_interface"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	respBody, resp, err := postRequest[schema.LoadBalancerActionDisablePublicInterfaceResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerChangeTypeOpts specifies options for changing a Load Balancer's type.
type LoadBalancerChangeTypeOpts struct {
	LoadBalancerType *LoadBalancerType // new Load Balancer type
}

// ChangeType changes a Load Balancer's type.
func (c *LoadBalancerClient) ChangeType(ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerChangeTypeOpts) (*Action, *Response, error) {
	const opPath = "/load_balancers/%d/actions/change_type"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID)

	reqBody := schema.LoadBalancerActionChangeTypeRequest{}
	if opts.LoadBalancerType.ID != 0 || opts.LoadBalancerType.Name != "" {
		reqBody.LoadBalancerType = schema.IDOrName{ID: opts.LoadBalancerType.ID, Name: opts.LoadBalancerType.Name}
	}

	respBody, resp, err := postRequest[schema.LoadBalancerActionChangeTypeResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// LoadBalancerMetricType is the type of available metrics for Load Balancers.
type LoadBalancerMetricType string

// Available types of Load Balancer metrics. See Hetzner Cloud API
// documentation for details.
const (
	LoadBalancerMetricOpenConnections      LoadBalancerMetricType = "open_connections"
	LoadBalancerMetricConnectionsPerSecond LoadBalancerMetricType = "connections_per_second"
	LoadBalancerMetricRequestsPerSecond    LoadBalancerMetricType = "requests_per_second"
	LoadBalancerMetricBandwidth            LoadBalancerMetricType = "bandwidth"
)

// LoadBalancerGetMetricsOpts configures the call to get metrics for a Load
// Balancer.
type LoadBalancerGetMetricsOpts struct {
	Types []LoadBalancerMetricType
	Start time.Time
	End   time.Time
	Step  int
}

func (o LoadBalancerGetMetricsOpts) Validate() error {
	if len(o.Types) == 0 {
		return missingField(o, "Types")
	}
	if o.Start.IsZero() {
		return missingField(o, "Start")
	}
	if o.End.IsZero() {
		return missingField(o, "End")
	}
	return nil
}

func (o LoadBalancerGetMetricsOpts) values() url.Values {
	query := url.Values{}

	for _, typ := range o.Types {
		query.Add("type", string(typ))
	}

	query.Add("start", o.Start.Format(time.RFC3339))
	query.Add("end", o.End.Format(time.RFC3339))

	if o.Step > 0 {
		query.Add("step", strconv.Itoa(o.Step))
	}

	return query
}

// LoadBalancerMetrics contains the metrics requested for a Load Balancer.
type LoadBalancerMetrics struct {
	Start      time.Time
	End        time.Time
	Step       float64
	TimeSeries map[string][]LoadBalancerMetricsValue
}

// LoadBalancerMetricsValue represents a single value in a time series of metrics.
type LoadBalancerMetricsValue struct {
	Timestamp float64
	Value     string
}

// GetMetrics obtains metrics for a Load Balancer.
func (c *LoadBalancerClient) GetMetrics(
	ctx context.Context, loadBalancer *LoadBalancer, opts LoadBalancerGetMetricsOpts,
) (*LoadBalancerMetrics, *Response, error) {
	const opPath = "/load_balancers/%d/metrics?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	if loadBalancer == nil {
		return nil, nil, missingArgument("loadBalancer", loadBalancer)
	}

	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	reqPath := fmt.Sprintf(opPath, loadBalancer.ID, opts.values().Encode())

	respBody, resp, err := getRequest[schema.LoadBalancerGetMetricsResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	metrics, err := loadBalancerMetricsFromSchema(&respBody)
	if err != nil {
		return nil, nil, fmt.Errorf("convert response body: %w", err)
	}

	return metrics, resp, nil
}

// ChangeDNSPtr changes or resets the reverse DNS pointer for a Load Balancer.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (c *LoadBalancerClient) ChangeDNSPtr(ctx context.Context, lb *LoadBalancer, ip string, ptr *string) (*Action, *Response, error) {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return nil, nil, InvalidIPError{ip}
	}
	return lb.changeDNSPtr(ctx, c.client, net.ParseIP(ip), ptr)
}
