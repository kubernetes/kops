package metadata

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/hetznercloud/hcloud-go/hcloud/internal/instrumentation"
	"github.com/prometheus/client_golang/prometheus"
)

const Endpoint = "http://169.254.169.254/hetzner/v1/metadata"

// Client is a client for the Hetzner Cloud Server Metadata Endpoints.
type Client struct {
	endpoint string

	httpClient              *http.Client
	instrumentationRegistry *prometheus.Registry
}

// A ClientOption is used to configure a Client.
type ClientOption func(*Client)

// WithEndpoint configures a Client to use the specified Metadata API endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(client *Client) {
		client.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithHTTPClient configures a Client to perform HTTP requests with httpClient.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithInstrumentation configures a Client to collect metrics about the performed HTTP requests.
func WithInstrumentation(registry *prometheus.Registry) ClientOption {
	return func(client *Client) {
		client.instrumentationRegistry = registry
	}
}

// NewClient creates a new client.
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		endpoint:   Endpoint,
		httpClient: &http.Client{},
	}

	for _, option := range options {
		option(client)
	}

	if client.instrumentationRegistry != nil {
		i := instrumentation.New("metadata", client.instrumentationRegistry)
		client.httpClient.Transport = i.InstrumentedRoundTripper()
	}
	return client
}

// NewRequest creates an HTTP request against the API. The returned request
// is assigned with ctx and has all necessary headers set (auth, user agent, etc.).
func (c *Client) get(path string) (string, error) {
	url := c.endpoint + path
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf("response status was %d", resp.StatusCode)
	}
	return body, nil
}

// IsHcloudServer checks if the currently called server is a hcloud server by calling a metadata endpoint
// if the endpoint answers with a non-empty value this method returns true, otherwise false
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

// Hostname returns the hostname of the server that did the request to the Metadata server
func (c *Client) Hostname() (string, error) {
	return c.get("/hostname")
}

// InstanceID returns the ID of the server that did the request to the Metadata server
func (c *Client) InstanceID() (int, error) {
	resp, err := c.get("/instance-id")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(resp)
}

// PublicIPv4 returns the Public IPv4 of the server that did the request to the Metadata server
func (c *Client) PublicIPv4() (net.IP, error) {
	resp, err := c.get("/public-ipv4")
	if err != nil {
		return nil, err
	}
	return net.ParseIP(resp), nil
}

// Region returns the Network Zone of the server that did the request to the Metadata server
func (c *Client) Region() (string, error) {
	return c.get("/region")
}

// AvailabilityZone returns the datacenter of the server that did the request to the Metadata server
func (c *Client) AvailabilityZone() (string, error) {
	return c.get("/availability-zone")
}

// PrivateNetworks returns details about the private networks the server is attached to
// Returns YAML (unparsed)
func (c *Client) PrivateNetworks() (string, error) {
	return c.get("/private-networks")
}
