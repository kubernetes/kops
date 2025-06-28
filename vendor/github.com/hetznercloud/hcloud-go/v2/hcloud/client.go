package hcloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/http/httpguts"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/internal/instrumentation"
)

// Endpoint is the base URL of the API.
const Endpoint = "https://api.hetzner.cloud/v1"

// UserAgent is the value for the library part of the User-Agent header
// that is sent with each request.
const UserAgent = "hcloud-go/" + Version

// A BackoffFunc returns the duration to wait before performing the
// next retry. The retries argument specifies how many retries have
// already been performed. When called for the first time, retries is 0.
type BackoffFunc func(retries int) time.Duration

// ConstantBackoff returns a BackoffFunc which backs off for
// constant duration d.
func ConstantBackoff(d time.Duration) BackoffFunc {
	return func(_ int) time.Duration {
		return d
	}
}

// ExponentialBackoff returns a BackoffFunc which implements an exponential
// backoff, truncated to 60 seconds.
// See [ExponentialBackoffWithOpts] for more details.
func ExponentialBackoff(multiplier float64, base time.Duration) BackoffFunc {
	return ExponentialBackoffWithOpts(ExponentialBackoffOpts{
		Base:       base,
		Multiplier: multiplier,
		Cap:        time.Minute,
	})
}

// ExponentialBackoffOpts defines the options used by [ExponentialBackoffWithOpts].
type ExponentialBackoffOpts struct {
	Base       time.Duration
	Multiplier float64
	Cap        time.Duration
	Jitter     bool
}

// ExponentialBackoffWithOpts returns a BackoffFunc which implements an exponential
// backoff, truncated to a maximum, and an optional full jitter.
//
// See https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func ExponentialBackoffWithOpts(opts ExponentialBackoffOpts) BackoffFunc {
	baseSeconds := opts.Base.Seconds()
	capSeconds := opts.Cap.Seconds()

	return func(retries int) time.Duration {
		// Exponential backoff
		backoff := baseSeconds * math.Pow(opts.Multiplier, float64(retries))
		// Cap backoff
		backoff = math.Min(capSeconds, backoff)
		// Add jitter
		if opts.Jitter {
			backoff = ((backoff - baseSeconds) * rand.Float64()) + baseSeconds // #nosec G404
		}

		return time.Duration(backoff * float64(time.Second))
	}
}

// Client is a client for the Hetzner Cloud API.
type Client struct {
	endpoint                string
	token                   string
	tokenValid              bool
	retryBackoffFunc        BackoffFunc
	retryMaxRetries         int
	pollBackoffFunc         BackoffFunc
	httpClient              *http.Client
	applicationName         string
	applicationVersion      string
	userAgent               string
	debugWriter             io.Writer
	instrumentationRegistry prometheus.Registerer
	handler                 handler

	Action           ActionClient
	Certificate      CertificateClient
	Datacenter       DatacenterClient
	Firewall         FirewallClient
	FloatingIP       FloatingIPClient
	Image            ImageClient
	ISO              ISOClient
	LoadBalancer     LoadBalancerClient
	LoadBalancerType LoadBalancerTypeClient
	Location         LocationClient
	Network          NetworkClient
	Pricing          PricingClient
	Server           ServerClient
	ServerType       ServerTypeClient
	SSHKey           SSHKeyClient
	Volume           VolumeClient
	PlacementGroup   PlacementGroupClient
	RDNS             RDNSClient
	PrimaryIP        PrimaryIPClient
}

// A ClientOption is used to configure a Client.
type ClientOption func(*Client)

// WithEndpoint configures a Client to use the specified API endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(client *Client) {
		client.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithToken configures a Client to use the specified token for authentication.
func WithToken(token string) ClientOption {
	return func(client *Client) {
		client.token = token
		client.tokenValid = httpguts.ValidHeaderFieldValue(token)
	}
}

// WithPollInterval configures a Client to use the specified interval when
// polling from the API.
//
// Deprecated: Setting the poll interval is deprecated, you can now configure
// [WithPollOpts] with a [ConstantBackoff] to get the same results. To
// migrate your code, replace your usage like this:
//
//	// before
//	hcloud.WithPollInterval(2 * time.Second)
//	// now
//	hcloud.WithPollOpts(hcloud.PollOpts{
//		BackoffFunc: hcloud.ConstantBackoff(2 * time.Second),
//	})
func WithPollInterval(pollInterval time.Duration) ClientOption {
	return WithPollOpts(PollOpts{
		BackoffFunc: ConstantBackoff(pollInterval),
	})
}

// WithPollBackoffFunc configures a Client to use the specified backoff
// function when polling from the API.
//
// Deprecated: WithPollBackoffFunc is deprecated, use [WithPollOpts] instead.
func WithPollBackoffFunc(f BackoffFunc) ClientOption {
	return WithPollOpts(PollOpts{
		BackoffFunc: f,
	})
}

// PollOpts defines the options used by [WithPollOpts].
type PollOpts struct {
	BackoffFunc BackoffFunc
}

// WithPollOpts configures a Client to use the specified options when polling from the API.
//
// If [PollOpts.BackoffFunc] is nil, the existing backoff function will be preserved.
func WithPollOpts(opts PollOpts) ClientOption {
	return func(client *Client) {
		if opts.BackoffFunc != nil {
			client.pollBackoffFunc = opts.BackoffFunc
		}
	}
}

// WithBackoffFunc configures a Client to use the specified backoff function.
// The backoff function is used for retrying HTTP requests.
//
// Deprecated: WithBackoffFunc is deprecated, use [WithRetryOpts] instead.
func WithBackoffFunc(f BackoffFunc) ClientOption {
	return func(client *Client) {
		client.retryBackoffFunc = f
	}
}

// RetryOpts defines the options used by [WithRetryOpts].
type RetryOpts struct {
	BackoffFunc BackoffFunc
	MaxRetries  int
}

// WithRetryOpts configures a Client to use the specified options when retrying API
// requests.
//
// If [RetryOpts.BackoffFunc] is nil, the existing backoff function will be preserved.
func WithRetryOpts(opts RetryOpts) ClientOption {
	return func(client *Client) {
		if opts.BackoffFunc != nil {
			client.retryBackoffFunc = opts.BackoffFunc
		}
		client.retryMaxRetries = opts.MaxRetries
	}
}

// WithApplication configures a Client with the given application name and
// application version. The version may be blank. Programs are encouraged
// to at least set an application name.
func WithApplication(name, version string) ClientOption {
	return func(client *Client) {
		client.applicationName = name
		client.applicationVersion = version
	}
}

// WithDebugWriter configures a Client to print debug information to the given
// writer. To, for example, print debug information on stderr, set it to os.Stderr.
func WithDebugWriter(debugWriter io.Writer) ClientOption {
	return func(client *Client) {
		client.debugWriter = debugWriter
	}
}

// WithHTTPClient configures a Client to perform HTTP requests with httpClient.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithInstrumentation configures a Client to collect metrics about the performed HTTP requests.
func WithInstrumentation(registry prometheus.Registerer) ClientOption {
	return func(client *Client) {
		client.instrumentationRegistry = registry
	}
}

// NewClient creates a new client.
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		endpoint:   Endpoint,
		tokenValid: true,
		httpClient: &http.Client{},

		retryBackoffFunc: ExponentialBackoffWithOpts(ExponentialBackoffOpts{
			Base:       time.Second,
			Multiplier: 2,
			Cap:        time.Minute,
			Jitter:     true,
		}),
		retryMaxRetries: 5,

		pollBackoffFunc: ConstantBackoff(500 * time.Millisecond),
	}

	for _, option := range options {
		option(client)
	}

	client.buildUserAgent()
	if client.instrumentationRegistry != nil {
		i := instrumentation.New("api", client.instrumentationRegistry)
		client.httpClient.Transport = i.InstrumentedRoundTripper(client.httpClient.Transport)
	}

	client.handler = assembleHandlerChain(client)

	client.Action = ActionClient{action: &ResourceActionClient{client: client}}
	client.Datacenter = DatacenterClient{client: client}
	client.FloatingIP = FloatingIPClient{client: client, Action: &ResourceActionClient{client: client, resource: "floating_ips"}}
	client.Image = ImageClient{client: client, Action: &ResourceActionClient{client: client, resource: "images"}}
	client.ISO = ISOClient{client: client}
	client.Location = LocationClient{client: client}
	client.Network = NetworkClient{client: client, Action: &ResourceActionClient{client: client, resource: "networks"}}
	client.Pricing = PricingClient{client: client}
	client.Server = ServerClient{client: client, Action: &ResourceActionClient{client: client, resource: "servers"}}
	client.ServerType = ServerTypeClient{client: client}
	client.SSHKey = SSHKeyClient{client: client}
	client.Volume = VolumeClient{client: client, Action: &ResourceActionClient{client: client, resource: "volumes"}}
	client.LoadBalancer = LoadBalancerClient{client: client, Action: &ResourceActionClient{client: client, resource: "load_balancers"}}
	client.LoadBalancerType = LoadBalancerTypeClient{client: client}
	client.Certificate = CertificateClient{client: client, Action: &ResourceActionClient{client: client, resource: "certificates"}}
	client.Firewall = FirewallClient{client: client, Action: &ResourceActionClient{client: client, resource: "firewalls"}}
	client.PlacementGroup = PlacementGroupClient{client: client}
	client.RDNS = RDNSClient{client: client}
	client.PrimaryIP = PrimaryIPClient{client: client, Action: &ResourceActionClient{client: client, resource: "primary_ips"}}

	return client
}

// NewRequest creates an HTTP request against the API. The returned request
// is assigned with ctx and has all necessary headers set (auth, user agent, etc.).
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := c.endpoint + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	if !c.tokenValid {
		return nil, errors.New("Authorization token contains invalid characters")
	} else if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req = req.WithContext(ctx)
	return req, nil
}

// Do performs an HTTP request against the API.
// v can be nil, an io.Writer to write the response body to or a pointer to
// a struct to json.Unmarshal the response to.
func (c *Client) Do(req *http.Request, v any) (*Response, error) {
	return c.handler.Do(req, v)
}

func (c *Client) buildUserAgent() {
	switch {
	case c.applicationName != "" && c.applicationVersion != "":
		c.userAgent = c.applicationName + "/" + c.applicationVersion + " " + UserAgent
	case c.applicationName != "" && c.applicationVersion == "":
		c.userAgent = c.applicationName + " " + UserAgent
	default:
		c.userAgent = UserAgent
	}
}

const (
	headerCorrelationID = "X-Correlation-Id"
)

// Response represents a response from the API. It embeds http.Response.
type Response struct {
	*http.Response
	Meta Meta

	// body holds a copy of the http.Response body that must be used within the handler
	// chain. The http.Response.Body is reserved for external users.
	body []byte
}

// populateBody copies the original [http.Response] body into the internal [Response] body
// property, and restore the original [http.Response] body as if it was untouched.
func (r *Response) populateBody() error {
	// Read full response body and save it for later use
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	r.body = body

	// Restore the body as if it was untouched, as it might be read by external users
	r.Body = io.NopCloser(bytes.NewReader(body))

	return nil
}

// hasJSONBody returns whether the response has a JSON body.
func (r *Response) hasJSONBody() bool {
	return len(r.body) > 0 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")
}

// internalCorrelationID returns the unique ID of the request as set by the API. This ID can help with support requests,
// as it allows the people working on identify this request in particular.
func (r *Response) internalCorrelationID() string {
	return r.Header.Get(headerCorrelationID)
}

// Meta represents meta information included in an API response.
type Meta struct {
	Pagination *Pagination
	Ratelimit  Ratelimit
}

// Pagination represents pagination meta information.
type Pagination struct {
	Page         int
	PerPage      int
	PreviousPage int
	NextPage     int
	LastPage     int
	TotalEntries int
}

// Ratelimit represents ratelimit information.
type Ratelimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// ListOpts specifies options for listing resources.
type ListOpts struct {
	Page          int    // Page (starting at 1)
	PerPage       int    // Items per page (0 means default)
	LabelSelector string // Label selector for filtering by labels
}

// Values returns the ListOpts as URL values.
func (l ListOpts) Values() url.Values {
	vals := url.Values{}
	if l.Page > 0 {
		vals.Add("page", strconv.Itoa(l.Page))
	}
	if l.PerPage > 0 {
		vals.Add("per_page", strconv.Itoa(l.PerPage))
	}
	if len(l.LabelSelector) > 0 {
		vals.Add("label_selector", l.LabelSelector)
	}
	return vals
}
