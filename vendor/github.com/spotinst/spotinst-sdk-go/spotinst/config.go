package spotinst

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	"github.com/spotinst/spotinst-sdk-go/spotinst/log"
)

const (
	// defaultBaseURL is the default base URL of the Spotinst API.
	// It is used e.g. when initializing a new Client without a specific address.
	defaultBaseURL = "https://api.spotinst.io"

	// defaultContentType is the default content type to use when making HTTP
	// calls.
	defaultContentType = "application/json"

	// defaultUserAgent is the default user agent to use when making HTTP
	// calls.
	defaultUserAgent = SDKName + "/" + SDKVersion

	// defaultMaxRetries is the number of retries for a single request after
	// the client will give up and return an error. It is zero by default, so
	// retry is disabled by default.
	defaultMaxRetries = 0

	// defaultGzipEnabled specifies if gzip compression is enabled by default.
	defaultGzipEnabled = false
)

// A Config provides Configuration to a service client instance.
type Config struct {
	BaseURL     *url.URL
	HTTPClient  *http.Client
	Credentials *credentials.Credentials
	Logger      log.Logger
	UserAgent   string
	ContentType string
}

func DefaultBaseURL() *url.URL {
	baseURL, _ := url.Parse(defaultBaseURL)
	return baseURL
}

// DefaultTransport returns a new http.Transport with similar default
// values to http.DefaultTransport. Do not use this for transient transports as
// it can leak file descriptors over time. Only use this for transports that
// will be re-used for the same host(s).
func DefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 1,
	}
}

// DefaultHTTPClient returns a new http.Client with similar default values to
// http.Client, but with a non-shared Transport, idle connections disabled, and
// KeepAlives disabled.
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport(),
	}
}

// DefaultConfig returns a default configuration for the client. By default this
// will pool and reuse idle connections to API. If you have a long-lived
// client object, this is the desired behavior and should make the most efficient
// use of the connections to API.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:     DefaultBaseURL(),
		HTTPClient:  DefaultHTTPClient(),
		UserAgent:   defaultUserAgent,
		ContentType: defaultContentType,
		Credentials: credentials.NewChainCredentials(
			new(credentials.EnvProvider),
			new(credentials.FileProvider),
		),
	}
}

// WithBaseURL defines the base URL of the Spotinst API.
func (c *Config) WithBaseURL(rawurl string) *Config {
	baseURL, _ := url.Parse(rawurl)
	c.BaseURL = baseURL
	return c
}

// WithHTTPClient defines the HTTP client.
func (c *Config) WithHTTPClient(client *http.Client) *Config {
	c.HTTPClient = client
	return c
}

// WithCredentials defines the credentials.
func (c *Config) WithCredentials(creds *credentials.Credentials) *Config {
	c.Credentials = creds
	return c
}

// WithUserAgent defines the user agent.
func (c *Config) WithUserAgent(ua string) *Config {
	c.UserAgent = fmt.Sprintf("%s+%s", ua, c.UserAgent)
	return c
}

// WithContentType defines the content type.
func (c *Config) WithContentType(ct string) *Config {
	c.ContentType = ct
	return c
}

// WithLogger defines the logger for informational messages, e.g. requests
// and their response times. It is nil by default.
func (c *Config) WithLogger(logger log.Logger) *Config {
	c.Logger = logger
	return c
}

// Merge merges the passed in configs into the existing config object.
func (c *Config) Merge(cfgs ...*Config) {
	for _, other := range cfgs {
		mergeConfig(c, other)
	}
}

func mergeConfig(dst *Config, other *Config) {
	if other == nil {
		return
	}
	if other.BaseURL != nil {
		dst.BaseURL = other.BaseURL
	}
	if other.Credentials != nil {
		dst.Credentials = other.Credentials
	}
	if other.HTTPClient != nil {
		dst.HTTPClient = other.HTTPClient
	}
	if other.UserAgent != "" {
		dst.UserAgent = other.UserAgent
	}
	if other.ContentType != "" {
		dst.ContentType = other.ContentType
	}
	if other.Logger != nil {
		dst.Logger = other.Logger
	}
}
