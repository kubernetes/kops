package instance

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/scaleway/scaleway-sdk-go/errors"
	"github.com/scaleway/scaleway-sdk-go/logger"
)

var metadataRetryBindPort = 200

const (
	metadataAPIv4         = "http://169.254.42.42"
	metadataAPIv6         = "http://[fd00:42::42]"
	metadataTimeout       = time.Second * 3
	metadataFallbackDelay = time.Second * -1
)

var ErrUnexpectedStatus = errors.New("service returned an unexpected status")

// metadataHTTPClient is the default HTTP client used to reach the metadata
// service. Both metadata addresses (IPv4 link-local 169.254.42.42 and IPv6
// fd00:42::42) are non globally routable, so we bypass proxy resolution
// entirely by setting a transport with Proxy: nil.
var metadataHTTPClient = &http.Client{
	Timeout: metadataTimeout,
	Transport: &http.Transport{
		Proxy: nil,
	},
}

// MetadataAPIOption is a function which applies options to a MetadataAPI.
type MetadataAPIOption func(*MetadataAPI)

// WithMetadataHTTPClient allows passing a custom http.Client which will be used
// for all metadata requests. This is primarily useful for testing.
func WithMetadataHTTPClient(client *http.Client) MetadataAPIOption {
	return func(m *MetadataAPI) {
		m.httpClient = client
	}
}

// UserData represents the user data
type UserData struct {
	UserData []string `json:"user_data,omitempty"`
}

// MetadataAPI metadata API
type MetadataAPI struct {
	MetadataURL *string
	httpClient  *http.Client
}

// NewMetadataAPI returns a MetadataAPI object from a Scaleway client.
func NewMetadataAPI(opts ...MetadataAPIOption) *MetadataAPI {
	meta := &MetadataAPI{
		httpClient: metadataHTTPClient,
	}

	for _, opt := range opts {
		opt(meta)
	}

	return meta
}

// newUserDataHTTPClient returns an HTTP client configured to bind to the
// given local TCP address. The userdata API requires requests to come from
// a privileged source port (see ListUserData and friends).
func newUserDataHTTPClient(localAddr *net.TCPAddr) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: nil,
			DialContext: (&net.Dialer{
				LocalAddr:     localAddr,
				FallbackDelay: metadataFallbackDelay,
			}).DialContext,
		},
	}
}

func (meta *MetadataAPI) getMetadataURLWithContext(ctx context.Context) string {
	if meta.MetadataURL != nil {
		return *meta.MetadataURL
	}

	for _, url := range []string{metadataAPIv4, metadataAPIv6} {
		req, err := http.NewRequestWithContext(
			ctx, http.MethodGet, url, bytes.NewBufferString(""),
		)
		if err != nil {
			logger.Warningf("Failed to create metadata URL %s: %v", url, err)
			continue
		}

		resp, err := meta.httpClient.Do(req)
		if err != nil {
			logger.Warningf("Failed to reach metadata URL %s: %v", url, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			meta.MetadataURL = &url
			return url
		}

		resp.Body.Close()
	}

	return metadataAPIv4
}

// GetMetadataWithContext returns the metadata available from the server
func (meta *MetadataAPI) GetMetadataWithContext(ctx context.Context) (m *Metadata, err error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		meta.getMetadataURLWithContext(ctx)+"/conf?format=json",
		bytes.NewBufferString(""),
	)
	if err != nil {
		return nil, err
	}

	resp, err := meta.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting metadataURL")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(ErrUnexpectedStatus, "%d", resp.StatusCode)
	}

	metadata := &Metadata{}
	err = json.NewDecoder(resp.Body).Decode(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding metadata")
	}

	return metadata, nil
}

// MetadataIP represents all public IPs attached
type MetadataIP struct {
	ID               string   `json:"id"`
	Address          string   `json:"address"`
	Dynamic          bool     `json:"dynamic"`
	Gateway          string   `json:"gateway"`
	Netmask          string   `json:"netmask"`
	Family           string   `json:"family"`
	ProvisioningMode string   `json:"provisioning_mode"`
	Tags             []string `json:"tags"`
}

// Metadata represents the struct return by the metadata API
type Metadata struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	Organization   string `json:"organization,omitempty"`
	Project        string `json:"project,omitempty"`
	CommercialType string `json:"commercial_type,omitempty"`
	Image          Image  `json:"image,omitempty"`
	// PublicIP IPv4 only
	PublicIP struct {
		ID               string `json:"id"`
		Address          string `json:"address"`
		Dynamic          bool   `json:"dynamic"`
		Gateway          string `json:"gateway"`
		Netmask          string `json:"netmask"`
		Family           string `json:"family"`
		ProvisioningMode string `json:"provisioning_mode"`
	} `json:"public_ip,omitempty"`
	PublicIpsV4 []MetadataIP `json:"public_ips_v4,omitempty"`
	PublicIpsV6 []MetadataIP `json:"public_ips_v6,omitempty"`
	PrivateIP   string       `json:"private_ip,omitempty"`
	IPv6        struct {
		Netmask string `json:"netmask,omitempty"`
		Gateway string `json:"gateway,omitempty"`
		Address string `json:"address,omitempty"`
	} `json:"ipv6,omitempty"`
	Location struct {
		PlatformID   string `json:"platform_id,omitempty"`
		HypervisorID string `json:"hypervisor_id,omitempty"`
		NodeID       string `json:"node_id,omitempty"`
		ClusterID    string `json:"cluster_id,omitempty"`
		ZoneID       string `json:"zone_id,omitempty"`
	} `json:"location,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	StateDetail   string   `json:"state_detail,omitempty"`
	SSHPublicKeys []struct {
		Description      string `json:"description,omitempty"`
		ModificationDate string `json:"modification_date,omitempty"`
		IP               string `json:"ip,omitempty"`
		Key              string `json:"key,omitempty"`
		Fingerprint      string `json:"fingerprint,omitempty"`
		ID               string `json:"id,omitempty"`
		CreationDate     string `json:"creation_date,omitempty"`
	} `json:"ssh_public_keys,omitempty"`
	Timezone   string `json:"timezone,omitempty"`
	Bootscript struct {
		Kernel       string `json:"kernel,omitempty"`
		Title        string `json:"title,omitempty"`
		Default      bool   `json:"default,omitempty"`
		Dtb          string `json:"dtb,omitempty"`
		Public       bool   `json:"publc,omitempty"`
		Initrd       string `json:"initrd,omitempty"`
		Bootcmdargs  string `json:"bootcmdargs,omitempty"`
		Architecture string `json:"architecture,omitempty"`
		Organization string `json:"organization,omitempty"`
		Project      string `json:"project,omitempty"`
		ID           string `json:"id,omitempty"`
	} `json:"bootscript,omitempty"`
	Volumes map[string]struct {
		Name             string `json:"name,omitempty"`
		ModificationDate string `json:"modification_date,omitempty"`
		ExportURI        string `json:"export_uri,omitempty"`
		VolumeType       string `json:"volume_type,omitempty"`
		CreationDate     string `json:"creation_date,omitempty"`
		State            string `json:"state,omitempty"`
		Organization     string `json:"organization,omitempty"`
		Project          string `json:"project,omitempty"`
		Server           struct {
			ID   string `json:"id,omitempty"`
			Name string `json:"name,omitempty"`
		} `json:"server,omitempty"`
		ID   string `json:"id,omitempty"`
		Size int    `json:"size,omitempty"`
	} `json:"volumes,omitempty"`
	PrivateNICs []struct {
		ID               string `json:"id,omitempty"`
		PrivateNetworkID string `json:"private_network_id,omitempty"`
		ServerID         string `json:"server_id,omitempty"`
		MacAddress       string `json:"mac_address,omitempty"`
		CreationDate     string `json:"creation_date,omitempty"`
		Zone             string `json:"zone,omitempty"`
	} `json:"private_nics,omitempty"`
	MACAddress    string `json:"mac_address,omitempty"`
	EndOfService  bool   `json:"end_of_service,omitempty"`
	Protected     bool   `json:"protected,omitempty"`
	SecurityGroup struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"security_group,omitempty"`
	State          string   `json:"state,omitempty"`
	Arch           string   `json:"arch,omitempty"`
	AllowedActions []string `json:"allowed_actions,omitempty"`
	FileSystems    []struct {
		FileSystemID string `json:"filesystem_id,omitempty"`
		State        string `json:"state,omitempty"`
	}
	CreatedAt         string `json:"creation_date,omitempty"`
	UpdatedAt         string `json:"modification_date,omitempty"`
	Zone              string `json:"zone,omitempty"`
	EnableIPv6        bool   `json:"enable_ipv6,omitempty"`
	DynamicIPRequired bool   `json:"dynamic_ip_required,omitempty"`
	PlacementGroup    struct {
		ID              string   `json:"id,omitempty"`
		Name            string   `json:"name,omitempty"`
		Organization    string   `json:"organization,omitempty"`
		Project         string   `json:"project,omitempty"`
		PolicyMode      string   `json:"policy_mode,omitempty"`
		PolicyType      string   `json:"policy_type,omitempty"`
		PolicyRespected bool     `json:"policy_respected,omitempty"`
		Tags            []string `json:"tags,omitempty"`
		Zone            string   `json:"zone,omitempty"`
	} `json:"placement_group,omitempty"`
}

// ListUserDataWithContext returns the metadata available from the server
func (meta *MetadataAPI) ListUserDataWithContext(ctx context.Context) (res *UserData, err error) {
	retries := 0
	for retries <= metadataRetryBindPort {
		port := rand.Intn(1024)

		body, retry, err := meta.tryUserDataRequest(ctx, http.MethodGet, "", []byte(""), port)
		if err != nil {
			return nil, err
		}

		// Retriable error
		if retry {
			retries++
			continue
		}

		userdata := &UserData{}
		err = json.Unmarshal(body, userdata)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding userdata")
		}

		return userdata, nil
	}

	return nil, errors.New("too many bind port retries for ListUserData")
}

// GetUserDataWithContext returns the value for the given metadata key
func (meta *MetadataAPI) GetUserDataWithContext(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return make([]byte, 0), errors.New("key must not be empty in GetUserData")
	}

	retries := 0
	for retries <= metadataRetryBindPort {
		port := rand.Intn(1024)

		body, retry, err := meta.tryUserDataRequest(ctx, http.MethodGet, key, []byte(""), port)
		if err != nil {
			return nil, err
		}

		// Retriable error
		if retry {
			retries++
			continue
		}

		return body, nil
	}

	return make([]byte, 0), errors.New("too many bind port retries for GetUserData")
}

// SetUserDataWithContext sets the userdata key with the given value
func (meta *MetadataAPI) SetUserDataWithContext(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errors.New("key must not be empty in SetUserData")
	}

	retries := 0
	for retries <= metadataRetryBindPort {
		port := rand.Intn(1024)
		_, retry, err := meta.tryUserDataRequest(ctx, http.MethodPatch, key, value, port)
		if err != nil {
			return err
		}

		// Retriable error
		if retry {
			retries++
			continue
		}

		// Success
		return nil
	}

	return errors.New("too many bind port retries for SetUserData")
}

// DeleteUserDataWithContext deletes the userdata key and the associated value
func (meta *MetadataAPI) DeleteUserDataWithContext(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("key must not be empty in DeleteUserData")
	}

	retries := 0
	for retries <= metadataRetryBindPort {
		port := rand.Intn(1024)

		_, retry, err := meta.tryUserDataRequest(ctx, http.MethodDelete, key, []byte(""), port)
		if err != nil {
			return err
		}

		// Retriable error
		if retry {
			retries++
			continue
		}

		// Success
		return nil
	}

	return errors.New("too many bind port retries for DeleteUserData")
}

func (meta *MetadataAPI) tryUserDataRequest(
	ctx context.Context,
	method string,
	key string,
	value []byte,
	port int,
) (body []byte, retry bool, err error) {
	body = make([]byte, 0)

	localTCPAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return body, false, errors.Wrap(err, "error resolving tcp address")
	}

	userdataClient := newUserDataHTTPClient(localTCPAddr)

	url := meta.getMetadataURLWithContext(ctx)
	if key != "" {
		url += "/user_data/" + key
	} else {
		url += "/user_data?format=json"
	}

	request, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		bytes.NewBuffer(value),
	)
	if err != nil {
		return body, false, errors.Wrap(err, "error creating userdata request")
	}

	request.Header.Set("Content-Type", "text/plain")

	resp, err := userdataClient.Do(request)
	if err != nil {
		// Retriable error
		return body, true, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return body, false, errors.Wrap(ErrUnexpectedStatus, "%d", resp.StatusCode)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), false, errors.Wrap(err, "error reading userdata body")
	}

	// Success
	return body, false, nil
}

// DeleteUserData deletes the userdata key and the associated value
//
// Deprecated: use DeleteUserDataWithContext instead
//
//go:fix inline
func (meta *MetadataAPI) DeleteUserData(key string) error {
	return meta.DeleteUserDataWithContext(context.Background(), key)
}

// SetUserData sets the userdata key with the given value
//
// Deprecated: use SetUserDataWithContext instead
//
//go:fix inline
func (meta *MetadataAPI) SetUserData(key string, value []byte) error {
	return meta.SetUserDataWithContext(context.Background(), key, value)
}

// GetUserData returns the value for the given metadata key
//
// Deprecated: use GetUserDataWithContext instead
//
//go:fix inline
func (meta *MetadataAPI) GetUserData(key string) ([]byte, error) {
	return meta.GetUserDataWithContext(context.Background(), key)
}

// ListUserData returns the metadata available from the server
//
// Deprecated: use ListUserDataWithContext instead
//
//go:fix inline
func (meta *MetadataAPI) ListUserData() (res *UserData, err error) {
	return meta.ListUserDataWithContext(context.Background())
}

// GetMetadata returns the metadata available from the server
//
// Deprecated: use GetMetadataWithContext instead
//
//go:fix inline
func (meta *MetadataAPI) GetMetadata() (m *Metadata, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), metadataTimeout)
	defer cancel()

	return meta.GetMetadataWithContext(ctx)
}
