package linodego

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

// StreamType represents the type of ACLP logs stream.
type StreamType string

const (
	// StreamTypeAuditLogs configures a stream for ACLP audit logs.
	StreamTypeAuditLogs StreamType = "audit_logs"
	// StreamTypeLKEAuditLogs configures a stream for LKE enterprise cluster audit logs.
	StreamTypeLKEAuditLogs StreamType = "lke_audit_logs"
)

// StreamStatus represents the availability state of an ACLP logs stream.
type StreamStatus string

const (
	// StreamStatusActive means the stream is actively delivering logs.
	StreamStatusActive StreamStatus = "active"
	// StreamStatusInactive means the stream is paused.
	StreamStatusInactive StreamStatus = "inactive"
	// StreamStatusProvisioning means the stream is being set up.
	StreamStatusProvisioning StreamStatus = "provisioning"
	// StreamStatusDeactivating means the stream is being deactivated.
	StreamStatusDeactivating StreamStatus = "deactivating"
)

// StreamDestinationType represents the destination type for ACLP logs streams.
type StreamDestinationType string

const (
	// StreamDestinationTypeAkamaiObjectStorage sends logs to Akamai Object Storage.
	StreamDestinationTypeAkamaiObjectStorage StreamDestinationType = "akamai_object_storage"
	// StreamDestinationTypeCustomHTTPS sends logs to a custom HTTPS endpoint.
	StreamDestinationTypeCustomHTTPS StreamDestinationType = "custom_https"
)

// StreamDetails contains additional details for a logs stream.
// This only applies to streams with a Type of StreamTypeLKEAuditLogs.
type StreamDetails struct {
	ClusterIDs                  []int `json:"cluster_ids,omitzero"`
	IsAutoAddAllClustersEnabled bool  `json:"is_auto_add_all_clusters_enabled"`
}

// StreamDestination is a destination configured on an ACLP logs stream.
type StreamDestination struct {
	ID      int                    `json:"id"`
	Label   string                 `json:"label"`
	Type    StreamDestinationType  `json:"type"`
	Details LogsDestinationDetails `json:"details"`
}

// Stream represents an ACLP logs stream.
type Stream struct {
	ID           int                 `json:"id"`
	Label        string              `json:"label"`
	Type         StreamType          `json:"type"`
	Status       StreamStatus        `json:"status"`
	Version      int                 `json:"version"`
	Destinations []StreamDestination `json:"destinations"`
	Details      *StreamDetails      `json:"details,omitzero"`
	Created      *time.Time          `json:"-"`
	Updated      *time.Time          `json:"-"`
	CreatedBy    string              `json:"created_by"`
	UpdatedBy    string              `json:"updated_by"`
}

// StreamCreateOptions are the fields used to create an ACLP logs stream.
type StreamCreateOptions struct {
	Destinations []int          `json:"destinations"`
	Label        string         `json:"label"`
	Type         StreamType     `json:"type"`
	Status       *StreamStatus  `json:"status,omitzero"`
	Details      *StreamDetails `json:"details,omitzero"`
}

// StreamUpdateOptions are the fields used to update an ACLP logs stream.
type StreamUpdateOptions struct {
	Destinations []int          `json:"destinations,omitzero"`
	Label        *string        `json:"label,omitzero"`
	Status       *StreamStatus  `json:"status,omitzero"`
	Details      *StreamDetails `json:"details,omitzero"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Stream) UnmarshalJSON(b []byte) error {
	type mask Stream

	p := struct {
		*mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		mask: (*mask)(s),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	s.Created = (*time.Time)(p.Created)
	s.Updated = (*time.Time)(p.Updated)

	return nil
}

// ListLogStreams returns all ACLP logs streams under the account.
func (c *Client) ListLogStreams(ctx context.Context, opts *ListOptions) ([]Stream, error) {
	return getPaginatedResults[Stream](ctx, c, "monitor/streams", opts)
}

// GetLogStream returns a single ACLP logs stream by ID.
func (c *Client) GetLogStream(ctx context.Context, streamID int) (*Stream, error) {
	e := formatAPIPath("monitor/streams/%d", streamID)
	return doGETRequest[Stream](ctx, c, e)
}

// CreateLogStream creates a new ACLP logs stream.
func (c *Client) CreateLogStream(ctx context.Context, opts StreamCreateOptions) (*Stream, error) {
	e := formatAPIPath("monitor/streams")
	return doPOSTRequest[Stream](ctx, c, e, opts)
}

// UpdateLogStream updates an ACLP logs stream by ID.
func (c *Client) UpdateLogStream(ctx context.Context, streamID int, opts StreamUpdateOptions) (*Stream, error) {
	e := formatAPIPath("monitor/streams/%d", streamID)
	return doPUTRequest[Stream](ctx, c, e, opts)
}

// ListLogStreamHistory returns all versions of an ACLP logs stream.
func (c *Client) ListLogStreamHistory(ctx context.Context, streamID int, opts *ListOptions) ([]Stream, error) {
	e := formatAPIPath("monitor/streams/%d/history", streamID)
	return getPaginatedResults[Stream](ctx, c, e, opts)
}

// DeleteLogStream deletes an ACLP logs stream by ID.
func (c *Client) DeleteLogStream(ctx context.Context, streamID int) error {
	e := formatAPIPath("monitor/streams/%d", streamID)
	return doDELETERequest(ctx, c, e)
}
