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

// Server represents a server in the Hetzner Cloud.
type Server struct {
	ID              int64
	Name            string
	Status          ServerStatus
	Created         time.Time
	PublicNet       ServerPublicNet
	PrivateNet      []ServerPrivateNet
	ServerType      *ServerType
	Datacenter      *Datacenter
	IncludedTraffic uint64
	OutgoingTraffic uint64
	IngoingTraffic  uint64
	BackupWindow    string
	RescueEnabled   bool
	Locked          bool
	ISO             *ISO
	Image           *Image
	Protection      ServerProtection
	Labels          map[string]string
	Volumes         []*Volume
	PrimaryDiskSize int
	PlacementGroup  *PlacementGroup
	LoadBalancers   []*LoadBalancer
}

// ServerProtection represents the protection level of a server.
type ServerProtection struct {
	Delete, Rebuild bool
}

// ServerStatus specifies a server's status.
type ServerStatus string

const (
	// ServerStatusInitializing is the status when a server is initializing.
	ServerStatusInitializing ServerStatus = "initializing"

	// ServerStatusOff is the status when a server is off.
	ServerStatusOff ServerStatus = "off"

	// ServerStatusRunning is the status when a server is running.
	ServerStatusRunning ServerStatus = "running"

	// ServerStatusStarting is the status when a server is being started.
	ServerStatusStarting ServerStatus = "starting"

	// ServerStatusStopping is the status when a server is being stopped.
	ServerStatusStopping ServerStatus = "stopping"

	// ServerStatusMigrating is the status when a server is being migrated.
	ServerStatusMigrating ServerStatus = "migrating"

	// ServerStatusRebuilding is the status when a server is being rebuilt.
	ServerStatusRebuilding ServerStatus = "rebuilding"

	// ServerStatusDeleting is the status when a server is being deleted.
	ServerStatusDeleting ServerStatus = "deleting"

	// ServerStatusUnknown is the status when a server's state is unknown.
	ServerStatusUnknown ServerStatus = "unknown"
)

// FirewallStatus specifies a Firewall's status.
type FirewallStatus string

const (
	// FirewallStatusPending is the status when a Firewall is pending.
	FirewallStatusPending FirewallStatus = "pending"

	// FirewallStatusApplied is the status when a Firewall is applied.
	FirewallStatusApplied FirewallStatus = "applied"
)

// ServerPublicNet represents a server's public network.
type ServerPublicNet struct {
	IPv4        ServerPublicNetIPv4
	IPv6        ServerPublicNetIPv6
	FloatingIPs []*FloatingIP
	Firewalls   []*ServerFirewallStatus
}

// ServerPublicNetIPv4 represents a server's public IPv4 address.
type ServerPublicNetIPv4 struct {
	ID      int64
	IP      net.IP
	Blocked bool
	DNSPtr  string
}

func (n *ServerPublicNetIPv4) IsUnspecified() bool {
	return n.IP == nil || n.IP.Equal(net.IPv4zero)
}

// ServerPublicNetIPv6 represents a Server's public IPv6 network and address.
type ServerPublicNetIPv6 struct {
	ID      int64
	IP      net.IP
	Network *net.IPNet
	Blocked bool
	DNSPtr  map[string]string
}

func (n *ServerPublicNetIPv6) IsUnspecified() bool {
	return n.IP == nil || n.IP.Equal(net.IPv6unspecified)
}

// ServerPrivateNet defines the schema of a Server's private network information.
type ServerPrivateNet struct {
	Network    *Network
	IP         net.IP
	Aliases    []net.IP
	MACAddress string
}

// DNSPtrForIP returns the reverse dns pointer of the ip address.
func (n *ServerPublicNetIPv6) DNSPtrForIP(ip net.IP) string {
	return n.DNSPtr[ip.String()]
}

// ServerFirewallStatus represents a Firewall and its status on a Server's
// network interface.
type ServerFirewallStatus struct {
	Firewall Firewall
	Status   FirewallStatus
}

// ServerRescueType represents rescue types.
type ServerRescueType string

// List of rescue types.
const (
	// Deprecated: Use ServerRescueTypeLinux64 instead.
	ServerRescueTypeLinux32 ServerRescueType = "linux32"
	ServerRescueTypeLinux64 ServerRescueType = "linux64"
)

// changeDNSPtr changes or resets the reverse DNS pointer for a IP address.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (s *Server) changeDNSPtr(ctx context.Context, client *Client, ip net.IP, ptr *string) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/change_dns_ptr"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, s.ID)

	reqBody := schema.ServerActionChangeDNSPtrRequest{
		IP:     ip.String(),
		DNSPtr: ptr,
	}

	respBody, resp, err := postRequest[schema.ServerActionChangeDNSPtrResponse](ctx, client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// GetDNSPtrForIP searches for the dns assigned to the given IP address.
// It returns an error if there is no dns set for the given IP address.
func (s *Server) GetDNSPtrForIP(ip net.IP) (string, error) {
	if net.IP.Equal(s.PublicNet.IPv4.IP, ip) {
		return s.PublicNet.IPv4.DNSPtr, nil
	} else if dns, ok := s.PublicNet.IPv6.DNSPtr[ip.String()]; ok {
		return dns, nil
	}

	return "", DNSNotFoundError{ip}
}

// ServerClient is a client for the servers API.
type ServerClient struct {
	client *Client
	Action *ResourceActionClient
}

// GetByID retrieves a server by its ID. If the server does not exist, nil is returned.
func (c *ServerClient) GetByID(ctx context.Context, id int64) (*Server, *Response, error) {
	const opPath = "/servers/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, id)

	respBody, resp, err := getRequest[schema.ServerGetResponse](ctx, c.client, reqPath)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, resp, err
	}

	return ServerFromSchema(respBody.Server), resp, nil
}

// GetByName retrieves a server by its name. If the server does not exist, nil is returned.
func (c *ServerClient) GetByName(ctx context.Context, name string) (*Server, *Response, error) {
	return firstByName(name, func() ([]*Server, *Response, error) {
		return c.List(ctx, ServerListOpts{Name: name})
	})
}

// Get retrieves a server by its ID if the input can be parsed as an integer, otherwise it
// retrieves a server by its name. If the server does not exist, nil is returned.
func (c *ServerClient) Get(ctx context.Context, idOrName string) (*Server, *Response, error) {
	return getByIDOrName(ctx, c.GetByID, c.GetByName, idOrName)
}

// ServerListOpts specifies options for listing servers.
type ServerListOpts struct {
	ListOpts
	Name   string
	Status []ServerStatus
	Sort   []string
}

func (l ServerListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, status := range l.Status {
		vals.Add("status", string(status))
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of servers for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *ServerClient) List(ctx context.Context, opts ServerListOpts) ([]*Server, *Response, error) {
	const opPath = "/servers?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, opts.values().Encode())

	respBody, resp, err := getRequest[schema.ServerListResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	return allFromSchemaFunc(respBody.Servers, ServerFromSchema), resp, nil
}

// All returns all servers.
func (c *ServerClient) All(ctx context.Context) ([]*Server, error) {
	return c.AllWithOpts(ctx, ServerListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all servers for the given options.
func (c *ServerClient) AllWithOpts(ctx context.Context, opts ServerListOpts) ([]*Server, error) {
	return iterPages(func(page int) ([]*Server, *Response, error) {
		opts.Page = page
		return c.List(ctx, opts)
	})
}

// ServerCreateOpts specifies options for creating a new server.
type ServerCreateOpts struct {
	Name             string
	ServerType       *ServerType
	Image            *Image
	SSHKeys          []*SSHKey
	Location         *Location
	Datacenter       *Datacenter
	UserData         string
	StartAfterCreate *bool
	Labels           map[string]string
	Automount        *bool
	Volumes          []*Volume
	Networks         []*Network
	Firewalls        []*ServerCreateFirewall
	PlacementGroup   *PlacementGroup
	PublicNet        *ServerCreatePublicNet
}

type ServerCreatePublicNet struct {
	EnableIPv4 bool
	EnableIPv6 bool
	IPv4       *PrimaryIP
	IPv6       *PrimaryIP
}

// ServerCreateFirewall defines which Firewalls to apply when creating a Server.
type ServerCreateFirewall struct {
	Firewall Firewall
}

// Validate checks if options are valid.
func (o ServerCreateOpts) Validate() error {
	if o.Name == "" {
		return missingField(o, "Name")
	}
	if o.ServerType == nil || (o.ServerType.ID == 0 && o.ServerType.Name == "") {
		return missingField(o, "ServerType")
	}
	if o.Image == nil || (o.Image.ID == 0 && o.Image.Name == "") {
		return missingField(o, "Image")
	}
	if o.Location != nil && o.Datacenter != nil {
		return mutuallyExclusiveFields(o, "Location", "Datacenter")
	}
	return nil
}

// ServerCreateResult is the result of a create server call.
type ServerCreateResult struct {
	Server       *Server
	Action       *Action
	RootPassword string
	NextActions  []*Action
}

// Create creates a new server.
func (c *ServerClient) Create(ctx context.Context, opts ServerCreateOpts) (ServerCreateResult, *Response, error) {
	const opPath = "/servers"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ServerCreateResult{}

	reqPath := opPath

	if err := opts.Validate(); err != nil {
		return result, nil, err
	}

	var reqBody schema.ServerCreateRequest
	reqBody.UserData = opts.UserData
	reqBody.Name = opts.Name
	reqBody.Automount = opts.Automount
	reqBody.StartAfterCreate = opts.StartAfterCreate
	if opts.ServerType.ID != 0 || opts.ServerType.Name != "" {
		reqBody.ServerType = schema.IDOrName{ID: opts.ServerType.ID, Name: opts.ServerType.Name}
	}
	if opts.Image.ID != 0 || opts.Image.Name != "" {
		reqBody.Image = schema.IDOrName{ID: opts.Image.ID, Name: opts.Image.Name}
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}
	for _, sshKey := range opts.SSHKeys {
		reqBody.SSHKeys = append(reqBody.SSHKeys, sshKey.ID)
	}
	for _, volume := range opts.Volumes {
		reqBody.Volumes = append(reqBody.Volumes, volume.ID)
	}
	for _, network := range opts.Networks {
		reqBody.Networks = append(reqBody.Networks, network.ID)
	}
	for _, firewall := range opts.Firewalls {
		reqBody.Firewalls = append(reqBody.Firewalls, schema.ServerCreateFirewalls{
			Firewall: firewall.Firewall.ID,
		})
	}

	if opts.PublicNet != nil {
		reqBody.PublicNet = &schema.ServerCreatePublicNet{
			EnableIPv4: opts.PublicNet.EnableIPv4,
			EnableIPv6: opts.PublicNet.EnableIPv6,
		}
		if opts.PublicNet.IPv4 != nil {
			reqBody.PublicNet.IPv4ID = opts.PublicNet.IPv4.ID
		}
		if opts.PublicNet.IPv6 != nil {
			reqBody.PublicNet.IPv6ID = opts.PublicNet.IPv6.ID
		}
	}
	if opts.Location != nil {
		if opts.Location.ID != 0 {
			reqBody.Location = strconv.FormatInt(opts.Location.ID, 10)
		} else {
			reqBody.Location = opts.Location.Name
		}
	}
	if opts.Datacenter != nil {
		if opts.Datacenter.ID != 0 {
			reqBody.Datacenter = strconv.FormatInt(opts.Datacenter.ID, 10)
		} else {
			reqBody.Datacenter = opts.Datacenter.Name
		}
	}
	if opts.PlacementGroup != nil {
		reqBody.PlacementGroup = opts.PlacementGroup.ID
	}

	respBody, resp, err := postRequest[schema.ServerCreateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Server = ServerFromSchema(respBody.Server)
	result.Action = ActionFromSchema(respBody.Action)
	result.NextActions = ActionsFromSchema(respBody.NextActions)
	if respBody.RootPassword != nil {
		result.RootPassword = *respBody.RootPassword
	}

	return result, resp, nil
}

// ServerDeleteResult is the result of a delete server call.
type ServerDeleteResult struct {
	Action *Action
}

// Delete deletes a server.
//
// Deprecated: Use [ServerClient.DeleteWithResult] instead.
func (c *ServerClient) Delete(ctx context.Context, server *Server) (*Response, error) {
	_, resp, err := c.DeleteWithResult(ctx, server)
	return resp, err
}

// DeleteWithResult deletes a server and returns the parsed response containing the action.
func (c *ServerClient) DeleteWithResult(ctx context.Context, server *Server) (*ServerDeleteResult, *Response, error) {
	const opPath = "/servers/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := &ServerDeleteResult{}

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := deleteRequest[schema.ServerDeleteResponse](ctx, c.client, reqPath)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, nil
}

// ServerUpdateOpts specifies options for updating a server.
type ServerUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// Update updates a server.
func (c *ServerClient) Update(ctx context.Context, server *Server, opts ServerUpdateOpts) (*Server, *Response, error) {
	const opPath = "/servers/%d"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerUpdateRequest{
		Name: opts.Name,
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	respBody, resp, err := putRequest[schema.ServerUpdateResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ServerFromSchema(respBody.Server), resp, nil
}

// Poweron starts a server.
func (c *ServerClient) Poweron(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/poweron"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionPoweronResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// Reboot reboots a server.
func (c *ServerClient) Reboot(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/reboot"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionRebootResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// Reset resets a server.
func (c *ServerClient) Reset(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/reset"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionResetResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// Shutdown shuts down a server.
func (c *ServerClient) Shutdown(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/shutdown"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionShutdownResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// Poweroff stops a server.
func (c *ServerClient) Poweroff(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/poweroff"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionPoweroffResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// ServerResetPasswordResult is the result of resetting a server's password.
type ServerResetPasswordResult struct {
	Action       *Action
	RootPassword string
}

// ResetPassword resets a server's password.
func (c *ServerClient) ResetPassword(ctx context.Context, server *Server) (ServerResetPasswordResult, *Response, error) {
	const opPath = "/servers/%d/actions/reset_password"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ServerResetPasswordResult{}

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionResetPasswordResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)
	result.RootPassword = respBody.RootPassword

	return result, resp, nil
}

// ServerCreateImageOpts specifies options for creating an image from a server.
type ServerCreateImageOpts struct {
	Type        ImageType
	Description *string
	Labels      map[string]string
}

// Validate checks if options are valid.
func (o ServerCreateImageOpts) Validate() error {
	switch o.Type {
	case ImageTypeSnapshot, ImageTypeBackup:
		break
	case "":
		break
	default:
		return invalidFieldValue(o, "Type", o.Type)
	}

	return nil
}

// ServerCreateImageResult is the result of creating an image from a server.
type ServerCreateImageResult struct {
	Action *Action
	Image  *Image
}

// CreateImage creates an image from a server.
func (c *ServerClient) CreateImage(ctx context.Context, server *Server, opts *ServerCreateImageOpts) (ServerCreateImageResult, *Response, error) {
	const opPath = "/servers/%d/actions/create_image"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ServerCreateImageResult{}

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionCreateImageRequest{}
	if opts != nil {
		if err := opts.Validate(); err != nil {
			return result, nil, err
		}
		if opts.Description != nil {
			reqBody.Description = opts.Description
		}
		if opts.Type != "" {
			reqBody.Type = Ptr(string(opts.Type))
		}
		if opts.Labels != nil {
			reqBody.Labels = &opts.Labels
		}
	}

	respBody, resp, err := postRequest[schema.ServerActionCreateImageResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Image = ImageFromSchema(respBody.Image)
	result.Action = ActionFromSchema(respBody.Action)

	return result, resp, nil
}

// ServerEnableRescueOpts specifies options for enabling rescue mode for a server.
type ServerEnableRescueOpts struct {
	Type    ServerRescueType
	SSHKeys []*SSHKey
}

// ServerEnableRescueResult is the result of enabling rescue mode for a server.
type ServerEnableRescueResult struct {
	Action       *Action
	RootPassword string
}

// EnableRescue enables rescue mode for a server.
func (c *ServerClient) EnableRescue(ctx context.Context, server *Server, opts ServerEnableRescueOpts) (ServerEnableRescueResult, *Response, error) {
	const opPath = "/servers/%d/actions/enable_rescue"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ServerEnableRescueResult{}

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionEnableRescueRequest{
		Type: Ptr(string(opts.Type)),
	}
	for _, sshKey := range opts.SSHKeys {
		reqBody.SSHKeys = append(reqBody.SSHKeys, sshKey.ID)
	}

	respBody, resp, err := postRequest[schema.ServerActionEnableRescueResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)
	result.RootPassword = respBody.RootPassword

	return result, resp, nil
}

// DisableRescue disables rescue mode for a server.
func (c *ServerClient) DisableRescue(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/disable_rescue"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionDisableRescueResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// ServerRebuildOpts specifies options for rebuilding a server.
type ServerRebuildOpts struct {
	Image *Image
}

// ServerRebuildResult is the result of a create server call.
type ServerRebuildResult struct {
	Action       *Action
	RootPassword string
}

// Rebuild rebuilds a server.
//
// Deprecated: Use [ServerClient.RebuildWithResult] instead.
func (c *ServerClient) Rebuild(ctx context.Context, server *Server, opts ServerRebuildOpts) (*Action, *Response, error) {
	result, resp, err := c.RebuildWithResult(ctx, server, opts)

	return result.Action, resp, err
}

// RebuildWithResult rebuilds a server.
func (c *ServerClient) RebuildWithResult(ctx context.Context, server *Server, opts ServerRebuildOpts) (ServerRebuildResult, *Response, error) {
	const opPath = "/servers/%d/actions/rebuild"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ServerRebuildResult{}

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionRebuildRequest{}
	if opts.Image.ID != 0 || opts.Image.Name != "" {
		reqBody.Image = schema.IDOrName{ID: opts.Image.ID, Name: opts.Image.Name}
	}

	respBody, resp, err := postRequest[schema.ServerActionRebuildResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)
	if respBody.RootPassword != nil {
		result.RootPassword = *respBody.RootPassword
	}

	return result, resp, nil
}

// AttachISO attaches an ISO to a server.
func (c *ServerClient) AttachISO(ctx context.Context, server *Server, iso *ISO) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/attach_iso"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionAttachISORequest{}
	if iso.ID != 0 || iso.Name != "" {
		reqBody.ISO = schema.IDOrName{ID: iso.ID, Name: iso.Name}
	}

	respBody, resp, err := postRequest[schema.ServerActionAttachISOResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// DetachISO detaches the currently attached ISO from a server.
func (c *ServerClient) DetachISO(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/detach_iso"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionDetachISOResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// EnableBackup enables backup for a server.
// The window parameter is deprecated and will be ignored.
func (c *ServerClient) EnableBackup(ctx context.Context, server *Server, window string) (*Action, *Response, error) {
	_ = window

	const opPath = "/servers/%d/actions/enable_backup"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionEnableBackupResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// DisableBackup disables backup for a server.
func (c *ServerClient) DisableBackup(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/disable_backup"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionDisableBackupResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// ServerChangeTypeOpts specifies options for changing a server's type.
type ServerChangeTypeOpts struct {
	ServerType  *ServerType // new server type
	UpgradeDisk bool        // whether disk should be upgraded
}

// ChangeType changes a server's type.
func (c *ServerClient) ChangeType(ctx context.Context, server *Server, opts ServerChangeTypeOpts) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/change_type"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionChangeTypeRequest{
		UpgradeDisk: opts.UpgradeDisk,
	}
	if opts.ServerType.ID != 0 || opts.ServerType.Name != "" {
		reqBody.ServerType = schema.IDOrName{ID: opts.ServerType.ID, Name: opts.ServerType.Name}
	}

	respBody, resp, err := postRequest[schema.ServerActionChangeTypeResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// ChangeDNSPtr changes or resets the reverse DNS pointer for a server IP address.
// Pass a nil ptr to reset the reverse DNS pointer to its default value.
func (c *ServerClient) ChangeDNSPtr(ctx context.Context, server *Server, ip string, ptr *string) (*Action, *Response, error) {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return nil, nil, InvalidIPError{ip}
	}
	return server.changeDNSPtr(ctx, c.client, net.ParseIP(ip), ptr)
}

// ServerChangeProtectionOpts specifies options for changing the resource protection level of a server.
type ServerChangeProtectionOpts struct {
	Rebuild *bool
	Delete  *bool
}

// ChangeProtection changes the resource protection level of a server.
func (c *ServerClient) ChangeProtection(ctx context.Context, server *Server, opts ServerChangeProtectionOpts) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/change_protection"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionChangeProtectionRequest{
		Rebuild: opts.Rebuild,
		Delete:  opts.Delete,
	}

	respBody, resp, err := postRequest[schema.ServerActionChangeProtectionResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

// ServerRequestConsoleResult is the result of requesting a WebSocket VNC console.
type ServerRequestConsoleResult struct {
	Action   *Action
	WSSURL   string
	Password string
}

// RequestConsole requests a WebSocket VNC console.
func (c *ServerClient) RequestConsole(ctx context.Context, server *Server) (ServerRequestConsoleResult, *Response, error) {
	const opPath = "/servers/%d/actions/request_console"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	result := ServerRequestConsoleResult{}

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionRequestConsoleResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return result, resp, err
	}

	result.Action = ActionFromSchema(respBody.Action)
	result.WSSURL = respBody.WSSURL
	result.Password = respBody.Password

	return result, resp, nil
}

// ServerAttachToNetworkOpts specifies options for attaching a server to a network.
type ServerAttachToNetworkOpts struct {
	Network  *Network
	IP       net.IP
	AliasIPs []net.IP
}

// AttachToNetwork attaches a server to a network.
func (c *ServerClient) AttachToNetwork(ctx context.Context, server *Server, opts ServerAttachToNetworkOpts) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/attach_to_network"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionAttachToNetworkRequest{
		Network: opts.Network.ID,
	}
	if opts.IP != nil {
		reqBody.IP = Ptr(opts.IP.String())
	}
	for _, aliasIP := range opts.AliasIPs {
		reqBody.AliasIPs = append(reqBody.AliasIPs, Ptr(aliasIP.String()))
	}

	respBody, resp, err := postRequest[schema.ServerActionAttachToNetworkResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ServerDetachFromNetworkOpts specifies options for detaching a server from a network.
type ServerDetachFromNetworkOpts struct {
	Network *Network
}

// DetachFromNetwork detaches a server from a network.
func (c *ServerClient) DetachFromNetwork(ctx context.Context, server *Server, opts ServerDetachFromNetworkOpts) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/detach_from_network"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionDetachFromNetworkRequest{
		Network: opts.Network.ID,
	}

	respBody, resp, err := postRequest[schema.ServerActionDetachFromNetworkResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ServerChangeAliasIPsOpts specifies options for changing the alias ips of an already attached network.
type ServerChangeAliasIPsOpts struct {
	Network  *Network
	AliasIPs []net.IP
}

// ChangeAliasIPs changes a server's alias IPs in a network.
func (c *ServerClient) ChangeAliasIPs(ctx context.Context, server *Server, opts ServerChangeAliasIPsOpts) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/change_alias_ips"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionChangeAliasIPsRequest{
		Network:  opts.Network.ID,
		AliasIPs: []string{},
	}
	for _, aliasIP := range opts.AliasIPs {
		reqBody.AliasIPs = append(reqBody.AliasIPs, aliasIP.String())
	}

	respBody, resp, err := postRequest[schema.ServerActionChangeAliasIPsResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, err
}

// ServerMetricType is the type of available metrics for servers.
type ServerMetricType string

// Available types of server metrics. See Hetzner Cloud API documentation for
// details.
const (
	ServerMetricCPU     ServerMetricType = "cpu"
	ServerMetricDisk    ServerMetricType = "disk"
	ServerMetricNetwork ServerMetricType = "network"
)

// ServerGetMetricsOpts configures the call to get metrics for a Server.
type ServerGetMetricsOpts struct {
	Types []ServerMetricType
	Start time.Time
	End   time.Time
	Step  int
}

func (o ServerGetMetricsOpts) Validate() error {
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

func (o ServerGetMetricsOpts) values() url.Values {
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

// ServerMetrics contains the metrics requested for a Server.
type ServerMetrics struct {
	Start      time.Time
	End        time.Time
	Step       float64
	TimeSeries map[string][]ServerMetricsValue
}

// ServerMetricsValue represents a single value in a time series of metrics.
type ServerMetricsValue struct {
	Timestamp float64
	Value     string
}

// GetMetrics obtains metrics for Server.
func (c *ServerClient) GetMetrics(ctx context.Context, server *Server, opts ServerGetMetricsOpts) (*ServerMetrics, *Response, error) {
	const opPath = "/servers/%d/metrics?%s"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	if server == nil {
		return nil, nil, missingArgument("server", server)
	}

	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}

	reqPath := fmt.Sprintf(opPath, server.ID, opts.values().Encode())

	respBody, resp, err := getRequest[schema.ServerGetMetricsResponse](ctx, c.client, reqPath)
	if err != nil {
		return nil, resp, err
	}

	metrics, err := serverMetricsFromSchema(&respBody)
	if err != nil {
		return nil, nil, fmt.Errorf("convert response body: %w", err)
	}

	return metrics, resp, nil
}

func (c *ServerClient) AddToPlacementGroup(ctx context.Context, server *Server, placementGroup *PlacementGroup) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/add_to_placement_group"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	reqBody := schema.ServerActionAddToPlacementGroupRequest{
		PlacementGroup: placementGroup.ID,
	}

	respBody, resp, err := postRequest[schema.ServerActionAddToPlacementGroupResponse](ctx, c.client, reqPath, reqBody)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}

func (c *ServerClient) RemoveFromPlacementGroup(ctx context.Context, server *Server) (*Action, *Response, error) {
	const opPath = "/servers/%d/actions/remove_from_placement_group"
	ctx = ctxutil.SetOpPath(ctx, opPath)

	reqPath := fmt.Sprintf(opPath, server.ID)

	respBody, resp, err := postRequest[schema.ServerActionRemoveFromPlacementGroupResponse](ctx, c.client, reqPath, nil)
	if err != nil {
		return nil, resp, err
	}

	return ActionFromSchema(respBody.Action), resp, nil
}
