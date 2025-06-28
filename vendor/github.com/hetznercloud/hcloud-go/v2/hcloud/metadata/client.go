package metadata

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/internal/instrumentation"
)

const Endpoint = "http://169.254.169.254/hetzner/v1/metadata"

// Client is a client for the Hetzner Cloud Server Metadata Endpoints.
type Client struct {
	endpoint string
	timeout  time.Duration

	httpClient              *http.Client
	instrumentationRegistry prometheus.Registerer
}

// A ClientOption is used to configure a [Client].
type ClientOption func(*Client)

// WithEndpoint configures a [Client] to use the specified Metadata API endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(client *Client) {
		client.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithHTTPClient configures a [Client] to perform HTTP requests with httpClient.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithInstrumentation configures a [Client] to collect metrics about the performed HTTP requests.
func WithInstrumentation(registry prometheus.Registerer) ClientOption {
	return func(client *Client) {
		client.instrumentationRegistry = registry
	}
}

// WithTimeout specifies a time limit for requests made by this [Client]. Defaults to 5 seconds.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.timeout = timeout
	}
}

// NewClient creates a new [Client] with the options applied.
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		endpoint:   Endpoint,
		httpClient: &http.Client{},
		timeout:    5 * time.Second,
	}

	for _, option := range options {
		option(client)
	}

	client.httpClient.Timeout = client.timeout

	if client.instrumentationRegistry != nil {
		i := instrumentation.New("metadata", client.instrumentationRegistry)
		client.httpClient.Transport = i.InstrumentedRoundTripper(client.httpClient.Transport)
	}
	return client
}

// get executes an HTTP request against the API.
func (c *Client) get(path string) (string, error) {
	ctx := ctxutil.SetOpPath(context.Background(), path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+path, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body := string(bytes.TrimSpace(bodyBytes))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf("response status was %d", resp.StatusCode)
	}
	return body, nil
}

// IsHcloudServer checks if the currently called server is a hcloud server by calling a metadata endpoint
// if the endpoint answers with a non-empty value this method returns true, otherwise false.
func (c *Client) IsHcloudServer() bool {
	hostname, err := c.Hostname()
	if err != nil {
		return false
	}
	if len(hostname) > 0 {
		return true
	}
	return false
}

// Hostname returns the hostname of the server that did the request to the Metadata server.
func (c *Client) Hostname() (string, error) {
	return c.get("/hostname")
}

// InstanceID returns the ID of the server that did the request to the Metadata server.
func (c *Client) InstanceID() (int64, error) {
	resp, err := c.get("/instance-id")
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(resp, 10, 64)
}

// PublicIPv4 returns the Public IPv4 of the server that did the request to the Metadata server.
func (c *Client) PublicIPv4() (net.IP, error) {
	resp, err := c.get("/public-ipv4")
	if err != nil {
		return nil, err
	}
	return net.ParseIP(resp), nil
}

// Region returns the Network Zone of the server that did the request to the Metadata server.
func (c *Client) Region() (string, error) {
	return c.get("/region")
}

// AvailabilityZone returns the datacenter of the server that did the request to the Metadata server.
func (c *Client) AvailabilityZone() (string, error) {
	return c.get("/availability-zone")
}

// PrivateNetworks returns details about the private networks the server is attached to.
// Returns YAML (unparsed).
func (c *Client) PrivateNetworks() (string, error) {
	return c.get("/private-networks")
}
