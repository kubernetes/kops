package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// LogsDestinationType represents the type of a logs destination.
type LogsDestinationType string

const (
	LogsDestinationTypeAkamaiObjectStorage LogsDestinationType = "akamai_object_storage"
	LogsDestinationTypeCustomHTTPS         LogsDestinationType = "custom_https"
)

// LogsDestinationStatus represents the status of a logs destination.
type LogsDestinationStatus string

const (
	LogsDestinationStatusActive   LogsDestinationStatus = "active"
	LogsDestinationStatusInactive LogsDestinationStatus = "inactive"
)

// LogsDestinationDetails represents the details block returned in a LogsDestination response.
// Fields are populated based on the destination type.
type LogsDestinationDetails struct {
	// akamai_object_storage fields
	AccessKeyID string `json:"access_key_id,omitzero"`
	BucketName  string `json:"bucket_name,omitzero"`
	Host        string `json:"host,omitzero"`
	Path        string `json:"path,omitzero"`

	// custom_https fields
	EndpointURL              string                                   `json:"endpoint_url,omitzero"`
	Authentication           *LogsDestinationCustomHTTPSAuthDetails   `json:"authentication,omitzero"`
	ClientCertificateDetails *LogsDestinationClientCertificateDetails `json:"client_certificate_details,omitzero"`
	ContentType              string                                   `json:"content_type,omitzero"`
	CustomHeaders            []LogsDestinationCustomHTTPSHeader       `json:"custom_headers,omitzero"`
	DataCompression          string                                   `json:"data_compression,omitzero"`
}

// LogsDestinationDetailsCreateOptions represents the details block used when creating
// an akamai_object_storage LogsDestination.
type LogsDestinationDetailsCreateOptions struct {
	AccessKeyID     string  `json:"access_key_id"`
	AccessKeySecret string  `json:"access_key_secret"`
	BucketName      string  `json:"bucket_name"`
	Host            string  `json:"host"`
	Path            *string `json:"path,omitzero"`
}

// LogsDestinationCustomHTTPSAuthType represents the authentication type for a custom_https destination.
type LogsDestinationCustomHTTPSAuthType string

const (
	LogsDestinationCustomHTTPSAuthTypeBasic LogsDestinationCustomHTTPSAuthType = "basic"
	LogsDestinationCustomHTTPSAuthTypeNone  LogsDestinationCustomHTTPSAuthType = "none"
)

// LogsDestinationCustomHTTPSBasicAuthDetails holds credentials for basic authentication.
// Both fields are required when authentication type is "basic".
type LogsDestinationCustomHTTPSBasicAuthDetails struct {
	Username string `json:"basic_authentication_user"`
	Password string `json:"basic_authentication_password"`
}

// LogsDestinationCustomHTTPSAuthDetails holds authentication configuration for a custom_https destination.
type LogsDestinationCustomHTTPSAuthDetails struct {
	Type    LogsDestinationCustomHTTPSAuthType          `json:"type"`
	Details *LogsDestinationCustomHTTPSBasicAuthDetails `json:"details,omitzero"`
}

// LogsDestinationCustomHTTPSHeader represents a single custom HTTP header.
type LogsDestinationCustomHTTPSHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// LogsDestinationClientCertificateDetails contains TLS client certificate information
type LogsDestinationClientCertificateDetails struct {
	ClientCACertificate string `json:"client_ca_certificate"`
	ClientCertificate   string `json:"client_certificate"`
	ClientPrivateKey    string `json:"client_private_key"`
	TLSHostname         string `json:"tls_hostname"`
}

// LogsDestinationCustomHTTPSDetailsCreateOptions represents the details block used when
// creating a custom_https LogsDestination.
type LogsDestinationCustomHTTPSDetailsCreateOptions struct {
	EndpointURL              string                                   `json:"endpoint_url"`
	Authentication           *LogsDestinationCustomHTTPSAuthDetails   `json:"authentication"`
	ClientCertificateDetails *LogsDestinationClientCertificateDetails `json:"client_certificate_details,omitzero"`
	ContentType              string                                   `json:"content_type,omitzero"`
	CustomHeaders            []LogsDestinationCustomHTTPSHeader       `json:"custom_headers,omitzero"`
	DataCompression          string                                   `json:"data_compression,omitzero"`
}

// LogsDestination represents a logs destination object.
type LogsDestination struct {
	Created   *time.Time             `json:"-"`
	CreatedBy string                 `json:"created_by"`
	Details   LogsDestinationDetails `json:"details"`
	ID        int                    `json:"id"`
	Label     string                 `json:"label"`
	Status    LogsDestinationStatus  `json:"status"`
	Type      LogsDestinationType    `json:"type"`
	Updated   *time.Time             `json:"-"`
	UpdatedBy string                 `json:"updated_by"`
	Version   int                    `json:"version"`
}

// UnmarshalJSON implements the json.Unmarshaler interface for LogsDestination.
func (i *LogsDestination) UnmarshalJSON(b []byte) error {
	type Mask LogsDestination

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.Updated = (*time.Time)(p.Updated)

	return nil
}

const logsDestinationBaseEndpoint = "monitor/streams/destinations"

// LogsDestinationCreateOptions are the options used to create a new logs destination.
type LogsDestinationCreateOptions struct {
	Label   string              `json:"label"`
	Type    LogsDestinationType `json:"type"`
	Details any                 `json:"details"`
}

// LogsDestinationDetailsUpdateOptions represents the details block used when updating
// an akamai_object_storage LogsDestination.
type LogsDestinationDetailsUpdateOptions struct {
	AccessKeyID     string  `json:"access_key_id,omitzero"`
	AccessKeySecret string  `json:"access_key_secret,omitzero"`
	BucketName      string  `json:"bucket_name,omitzero"`
	Host            string  `json:"host,omitzero"`
	Path            *string `json:"path,omitzero"`
}

// LogsDestinationCustomHTTPSDetailsUpdateOptions represents the details block used when
// updating a custom_https LogsDestination.
type LogsDestinationCustomHTTPSDetailsUpdateOptions struct {
	EndpointURL              string                                   `json:"endpoint_url,omitzero"`
	Authentication           *LogsDestinationCustomHTTPSAuthDetails   `json:"authentication,omitzero"`
	ClientCertificateDetails *LogsDestinationClientCertificateDetails `json:"client_certificate_details,omitzero"`
	ContentType              string                                   `json:"content_type,omitzero"`
	CustomHeaders            []LogsDestinationCustomHTTPSHeader       `json:"custom_headers,omitzero"`
	DataCompression          string                                   `json:"data_compression,omitzero"`
}

// LogsDestinationUpdateOptions are the options used to update a LogsDestination.
// Set Details to *LogsDestinationDetailsUpdateOptions for akamai_object_storage,
// or *LogsDestinationCustomHTTPSDetailsUpdateOptions for custom_https.
type LogsDestinationUpdateOptions struct {
	Label   string `json:"label,omitzero"`
	Details any    `json:"details,omitzero"`
}

// ListLogsDestinations returns a paginated list of logs destinations.
func (c *Client) ListLogsDestinations(ctx context.Context, opts *ListOptions) ([]LogsDestination, error) {
	return getPaginatedResults[LogsDestination](ctx, c, logsDestinationBaseEndpoint, opts)
}

// GetLogsDestination gets a single logs destination by ID.
func (c *Client) GetLogsDestination(ctx context.Context, destinationID int) (*LogsDestination, error) {
	e := formatAPIPath(logsDestinationBaseEndpoint+"/%d", destinationID)
	return doGETRequest[LogsDestination](ctx, c, e)
}

// CreateLogsDestination creates a new logs destination.
func (c *Client) CreateLogsDestination(ctx context.Context, opts LogsDestinationCreateOptions) (*LogsDestination, error) {
	return doPOSTRequest[LogsDestination](ctx, c, logsDestinationBaseEndpoint, opts)
}

// UpdateLogsDestination updates a logs destination.
func (c *Client) UpdateLogsDestination(ctx context.Context, destinationID int, opts LogsDestinationUpdateOptions) (*LogsDestination, error) {
	e := formatAPIPath(logsDestinationBaseEndpoint+"/%d", destinationID)
	return doPUTRequest[LogsDestination](ctx, c, e, opts)
}

// DeleteLogsDestination deletes a logs destination.
func (c *Client) DeleteLogsDestination(ctx context.Context, destinationID int) error {
	e := formatAPIPath(logsDestinationBaseEndpoint+"/%d", destinationID)
	return doDELETERequest(ctx, c, e)
}

// ListLogsDestinationHistory returns the version history for a logs destination.
func (c *Client) ListLogsDestinationHistory(ctx context.Context, destinationID int, opts *ListOptions) ([]LogsDestination, error) {
	e := formatAPIPath(logsDestinationBaseEndpoint+"/%d/history", destinationID)
	return getPaginatedResults[LogsDestination](ctx, c, e, opts)
}
