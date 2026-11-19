package scw

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/auth"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

// ClientOption is a function which applies options to a settings object.
type ClientOption func(*settings)

// httpClient wraps the net/http Client Do method
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// WithHTTPClient client option allows passing a custom http.Client which will be used for all requests.
func WithHTTPClient(httpClient httpClient) ClientOption {
	return func(s *settings) {
		s.httpClient = httpClient
	}
}

// WithoutAuth client option sets the client token to an empty token.
func WithoutAuth() ClientOption {
	return func(s *settings) {
		s.token = auth.NewNoAuth()
	}
}

// WithAuth client option sets the client access key and secret key.
func WithAuth(accessKey, secretKey string) ClientOption {
	return func(s *settings) {
		s.token = auth.NewToken(accessKey, secretKey)
	}
}

// WithJWT client option sets the client session token.
func WithJWT(token string) ClientOption {
	return func(s *settings) {
		s.token = auth.NewJWT(token)
	}
}

// WithAPIURL client option overrides the API URL of the Scaleway API to the given URL.
func WithAPIURL(apiURL string) ClientOption {
	return func(s *settings) {
		s.apiURL = apiURL
	}
}

// WithS3Endpoint client option overrides the S3 URL of the Scaleway Object Bucket API to the given URL.
func WithS3Endpoint(s3Endpoint string) ClientOption {
	return func(s *settings) {
		s.s3Endpoint = s3Endpoint
	}
}

// WithS3UsePathStyle client option overrides the UsePathStyle option when
// accessing the S3 API.
func WithS3UsePathStyle(s3UsePathStyle bool) ClientOption {
	return func(s *settings) {
		s.s3UsePathStyle = s3UsePathStyle
	}
}

// WithInsecure client option enables insecure transport on the client.
func WithInsecure() ClientOption {
	return func(s *settings) {
		s.insecure = true
	}
}

// WithUserAgent client option append a user agent to the default user agent of the SDK.
func WithUserAgent(ua string) ClientOption {
	return func(s *settings) {
		if s.userAgent != "" && ua != "" {
			s.userAgent += " "
		}
		s.userAgent += ua
	}
}

// WithDefaultUserAgent client option overrides the default user agent of the SDK.
func WithDefaultUserAgent(ua string) ClientOption {
	return func(s *settings) {
		s.userAgent = ua
	}
}

// WithProfile client option configures a client from the given profile.
func WithProfile(p *Profile) ClientOption {
	return func(s *settings) {
		accessKey := ""
		if p.AccessKey != nil {
			accessKey = *p.AccessKey
			s.token = auth.NewAccessKeyOnly(accessKey)
		}

		if p.SecretKey != nil {
			s.token = auth.NewToken(accessKey, *p.SecretKey)
		}

		if p.APIURL != nil {
			s.apiURL = *p.APIURL
		}

		if p.S3Endpoint != nil {
			s.s3Endpoint = *p.S3Endpoint
		}

		if p.S3UsePathStyle != nil {
			s.s3UsePathStyle = *p.S3UsePathStyle
		}

		if p.Insecure != nil {
			s.insecure = *p.Insecure
		}

		if p.DefaultOrganizationID != nil {
			organizationID := *p.DefaultOrganizationID
			s.defaultOrganizationID = &organizationID
		}

		if p.DefaultProjectID != nil {
			projectID := *p.DefaultProjectID
			s.defaultProjectID = &projectID
		}

		if p.DefaultRegion != nil {
			defaultRegion := Region(*p.DefaultRegion)
			s.defaultRegion = &defaultRegion
		}

		if p.DefaultZone != nil {
			defaultZone := Zone(*p.DefaultZone)
			s.defaultZone = &defaultZone
		}

		if p.UserAgent != nil {
			s.userAgent = *p.UserAgent
		}
	}
}

// WithEnv client option configures a client from the environment variables.
func WithEnv() ClientOption {
	return WithProfile(LoadEnvProfile())
}

// WithDefaultOrganizationID client option sets the client default organization ID.
//
// It will be used as the default value of the organization_id field in all requests made with this client.
func WithDefaultOrganizationID(organizationID string) ClientOption {
	return func(s *settings) {
		s.defaultOrganizationID = &organizationID
	}
}

// WithDefaultProjectID client option sets the client default project ID.
//
// It will be used as the default value of the projectID field in all requests made with this client.
func WithDefaultProjectID(projectID string) ClientOption {
	return func(s *settings) {
		s.defaultProjectID = &projectID
	}
}

// WithDefaultRegion client option sets the client default region.
//
// It will be used as the default value of the region field in all requests made with this client.
func WithDefaultRegion(region Region) ClientOption {
	return func(s *settings) {
		s.defaultRegion = &region
	}
}

// WithDefaultZone client option sets the client default zone.
//
// It will be used as the default value of the zone field in all requests made with this client.
func WithDefaultZone(zone Zone) ClientOption {
	return func(s *settings) {
		s.defaultZone = &zone
	}
}

// WithDefaultPageSize client option overrides the default page size of the SDK.
//
// It will be used as the default value of the page_size field in all requests made with this client.
func WithDefaultPageSize(pageSize uint32) ClientOption {
	return func(s *settings) {
		s.defaultPageSize = &pageSize
	}
}

// settings hold the values of all client options
type settings struct {
	apiURL                string
	s3Endpoint            string
	s3UsePathStyle        bool
	token                 auth.Auth
	userAgent             string
	httpClient            httpClient
	insecure              bool
	defaultOrganizationID *string
	defaultProjectID      *string
	defaultRegion         *Region
	defaultZone           *Zone
	defaultPageSize       *uint32
}

func newSettings() *settings {
	return &settings{}
}

func (s *settings) apply(opts []ClientOption) {
	for _, opt := range opts {
		opt(s)
	}
}

func (s *settings) validate() error {
	if err := s.validateAuth(); err != nil {
		return err
	}

	if err := s.validateOrgID(); err != nil {
		return err
	}

	if err := s.validateProjectID(); err != nil {
		return err
	}

	if err := s.validateRegion(); err != nil {
		return err
	}

	if err := s.validateZone(); err != nil {
		return err
	}

	if err := s.validateUserAgent(); err != nil {
		return err
	}

	if err := s.validateAPIURL(); err != nil {
		return err
	}

	if err := s.validateS3Endpoint(); err != nil {
		return err
	}

	// TODO: check for max s.defaultPageSize

	return nil
}

func (s *settings) validateAuth() error {
	if s.token == nil {
		// It should not happen, WithoutAuth option is used by default.
		panic(errors.New("no credential option provided"))
	}

	if token, isToken := s.token.(*auth.Token); isToken {
		if token.AccessKey == "" {
			return NewInvalidClientOptionError("access key cannot be empty")
		}
		if !validation.IsAccessKey(token.AccessKey) {
			return NewInvalidClientOptionError("invalid access key format '%s', expected SCWXXXXXXXXXXXXXXXXX format", token.AccessKey)
		}
		if token.SecretKey == "" {
			return NewInvalidClientOptionError("secret key cannot be empty")
		}
		if !validation.IsSecretKey(token.SecretKey) {
			return NewInvalidClientOptionError("invalid secret key format '%s', expected a UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", token.SecretKey)
		}
	}

	return nil
}

func (s *settings) validateOrgID() error {
	if s.defaultOrganizationID != nil {
		if *s.defaultOrganizationID == "" {
			return NewInvalidClientOptionError("default organization ID cannot be empty")
		}
		if !validation.IsOrganizationID(*s.defaultOrganizationID) {
			return NewInvalidClientOptionError("invalid organization ID format '%s', expected a UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", *s.defaultOrganizationID)
		}
	}

	return nil
}

func (s *settings) validateProjectID() error {
	if s.defaultProjectID != nil {
		if *s.defaultProjectID == "" {
			return NewInvalidClientOptionError("default project ID cannot be empty")
		}
		if !validation.IsProjectID(*s.defaultProjectID) {
			return NewInvalidClientOptionError("invalid project ID format '%s', expected a UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", *s.defaultProjectID)
		}
	}

	return nil
}

func (s *settings) validateRegion() error {
	if s.defaultRegion != nil {
		if *s.defaultRegion == "" {
			return NewInvalidClientOptionError("default region cannot be empty")
		}
		if !validation.IsRegion(string(*s.defaultRegion)) {
			regions := []string(nil)
			for _, r := range AllRegions {
				regions = append(regions, string(r))
			}
			return NewInvalidClientOptionError("invalid default region format '%s', available regions are: %s", *s.defaultRegion, strings.Join(regions, ", "))
		}
	}

	return nil
}

func (s *settings) validateZone() error {
	if s.defaultZone != nil {
		if *s.defaultZone == "" {
			return NewInvalidClientOptionError("default zone cannot be empty")
		}
		if !validation.IsZone(string(*s.defaultZone)) {
			zones := []string(nil)
			for _, z := range AllZones {
				zones = append(zones, string(z))
			}
			return NewInvalidClientOptionError("invalid default zone format '%s', available zones are: %s", *s.defaultZone, strings.Join(zones, ", "))
		}
	}

	return nil
}

func (s *settings) validateUserAgent() error {
	if s.userAgent != "" && s.userAgent != defaultUserAgent {
		safeUA := SanitizeForLogging(s.userAgent)

		if len(s.userAgent) > UserAgentMaxLength {
			return NewInvalidClientOptionError(
				"invalid user agent '%s': length should not be over %d",
				safeUA, UserAgentMaxLength,
			)
		}

		if !utf8.ValidString(s.userAgent) {
			return NewInvalidClientOptionError(
				"invalid user agent '%s': is not a valid UTF-8 string",
				safeUA,
			)
		}

		// Reject control characters (including CRLF \r\n, tabs, and null bytes)
		if !printableASCIIRegex.MatchString(s.userAgent) {
			return NewInvalidClientOptionError(
				"invalid user agent '%s': contains prohibited characters",
				safeUA,
			)
		}
	}

	return nil
}

func (s *settings) validateAPIURL() error {
	if !validation.IsURL(s.apiURL) {
		return NewInvalidClientOptionError("invalid API url '%s'", s.apiURL)
	}

	if s.apiURL[len(s.apiURL)-1:] == "/" {
		return NewInvalidClientOptionError("invalid API url '%s' it should not have a trailing slash", s.apiURL)
	}

	return nil
}

func (s *settings) validateS3Endpoint() error {
	if s.s3Endpoint != "" {
		if !validation.IsURL(s.s3Endpoint) {
			return NewInvalidClientOptionError("invalid S3 endpoint '%s'", s.s3Endpoint)
		}
		if s.s3Endpoint[len(s.s3Endpoint)-1:] == "/" {
			return NewInvalidClientOptionError("invalid S3 endpoint '%s': trailing slash is not allowed", s.s3Endpoint)
		}
	}

	return nil
}
