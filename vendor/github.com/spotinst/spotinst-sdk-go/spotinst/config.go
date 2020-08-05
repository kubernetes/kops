package spotinst

import (
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	"github.com/spotinst/spotinst-sdk-go/spotinst/log"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/useragent"
)

const (
	// defaultBaseURL is the default base URL of the Spotinst API.
	// It is used e.g. when initializing a new Client without a specific address.
	defaultBaseURL = "https://api.spotinst.io"

	// defaultContentType is the default content type to use when making HTTP calls.
	defaultContentType = "application/json"
)

// A Config provides Configuration to a service client instance.
type Config struct {
	// The base URL the SDK's HTTP client will use when invoking HTTP requests.
	BaseURL *url.URL

	// The HTTP Client the SDK's API clients will use to invoke HTTP requests.
	//
	// Defaults to a DefaultHTTPClient allowing API clients to create copies of
	// the HTTP client for service specific customizations.
	HTTPClient *http.Client

	// The credentials object to use when signing requests.
	//
	// Defaults to a chain of credential providers to search for credentials in
	// environment variables and shared credential file.
	Credentials *credentials.Credentials

	// The logger writer interface to write logging messages to.
	//
	// Defaults to standard out.
	Logger log.Logger

	// The User-Agent and Content-Type HTTP headers to set when invoking HTTP
	// requests.
	UserAgent, ContentType string
}

// DefaultBaseURL returns the default base URL.
func DefaultBaseURL() *url.URL {
	baseURL, _ := url.Parse(defaultBaseURL)
	return baseURL
}

// DefaultUserAgent returns the default User-Agent header.
func DefaultUserAgent() string {
	return useragent.New(
		SDKName,
		SDKVersion,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH).String()
}

// DefaultContentType returns the default Content-Type header.
func DefaultContentType() string {
	return defaultContentType
}

// DefaultTransport returns a new http.Transport with similar default values to
// http.DefaultTransport. Do not use this for transient transports as it can
// leak file descriptors over time. Only use this for transports that will be
// re-used for the same host(s).
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
// will pool and reuse idle connections to API. If you have a long-lived client
// object, this is the desired behavior and should make the most efficient use
// of the connections to API.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:     DefaultBaseURL(),
		HTTPClient:  DefaultHTTPClient(),
		UserAgent:   DefaultUserAgent(),
		ContentType: DefaultContentType(),
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
	c.UserAgent = strings.TrimSpace(strings.Join([]string{ua, c.UserAgent}, " "))
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
	for _, cfg := range cfgs {
		mergeConfigs(c, cfg)
	}
}

func mergeConfigs(c1, c2 *Config) {
	if c2 == nil {
		return
	}
	if c2.BaseURL != nil {
		c1.BaseURL = c2.BaseURL
	}
	if c2.Credentials != nil {
		c1.Credentials = c2.Credentials
	}
	if c2.HTTPClient != nil {
		c1.HTTPClient = c2.HTTPClient
	}
	if c2.UserAgent != "" {
		c1.UserAgent = c2.UserAgent
	}
	if c2.ContentType != "" {
		c1.ContentType = c2.ContentType
	}
	if c2.Logger != nil {
		c1.Logger = c2.Logger
	}
}
