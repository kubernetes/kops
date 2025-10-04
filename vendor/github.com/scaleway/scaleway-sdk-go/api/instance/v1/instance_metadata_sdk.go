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
	metadataAPIv4 = "http://169.254.42.42"
	metadataAPIv6 = "http://[fd00:42::42]"
)

// MetadataAPI metadata API
type MetadataAPI struct {
	MetadataURL *string
}

// NewMetadataAPI returns a MetadataAPI object from a Scaleway client.
func NewMetadataAPI() *MetadataAPI {
	return &MetadataAPI{}
}

func (meta *MetadataAPI) getMetadataURL() string {
	if meta.MetadataURL != nil {
		return *meta.MetadataURL
	}

	ctx := context.Background()
	for _, url := range []string{metadataAPIv4, metadataAPIv6} {
		http.DefaultClient.Timeout = 3 * time.Second
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, bytes.NewBufferString(""))
		if err != nil {
			logger.Warningf("Failed to create metadata URL %s: %v", url, err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			meta.MetadataURL = &url
			return url
		}
		defer resp.Body.Close()
	}
	return metadataAPIv4
}

// GetMetadata returns the metadata available from the server
func (meta *MetadataAPI) GetMetadata() (m *Metadata, err error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, meta.getMetadataURL()+"/conf?format=json", bytes.NewBufferString(""))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error getting metadataURL")
	}
	defer resp.Body.Close()

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
}

// ListUserData returns the metadata available from the server
func (meta *MetadataAPI) ListUserData() (res *UserData, err error) {
	retries := 0
	ctx := context.Background()
	for retries <= metadataRetryBindPort {
		port := rand.Intn(1024)
		localTCPAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return nil, errors.Wrap(err, "error resolving tcp address")
		}

		userdataClient := &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					LocalAddr:     localTCPAddr,
					DualStack:     false,
					FallbackDelay: time.Second * -1,
				}).DialContext,
			},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, meta.getMetadataURL()+"/user_data?format=json", bytes.NewBufferString(""))
		if err != nil {
			return nil, err
		}
		resp, err := userdataClient.Do(req)
		if err != nil {
			retries++ // retry with a different source port
			continue
		}
		defer resp.Body.Close()

		userdata := &UserData{}
		err = json.NewDecoder(resp.Body).Decode(userdata)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding userdata")
		}
		return userdata, nil
	}
	return nil, errors.New("too many bind port retries for ListUserData")
}

// GetUserData returns the value for the given metadata key
func (meta *MetadataAPI) GetUserData(key string) ([]byte, error) {
	if key == "" {
		return make([]byte, 0), errors.New("key must not be empty in GetUserData")
	}

	ctx := context.Background()
	retries := 0
	for retries <= metadataRetryBindPort {
		port := rand.Intn(1024)
		localTCPAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return make([]byte, 0), errors.Wrap(err, "error resolving tcp address")
		}

		userdataClient := &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					LocalAddr:     localTCPAddr,
					DualStack:     false,
					FallbackDelay: time.Second * -1,
				}).DialContext,
			},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, meta.getMetadataURL()+"/user_data/"+key, bytes.NewBufferString(""))
		if err != nil {
			return nil, err
		}

		resp, err := userdataClient.Do(req)
		if err != nil {
			retries++ // retry with a different source port
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return make([]byte, 0), errors.Wrap(err, "error reading userdata body")
		}

		return body, nil
	}
	return make([]byte, 0), errors.New("too may bind port retries for GetUserData")
}

// SetUserData sets the userdata key with the given value
func (meta *MetadataAPI) SetUserData(key string, value []byte) error {
	if key == "" {
		return errors.New("key must not be empty in SetUserData")
	}

	ctx := context.Background()
	retries := 0
	for retries <= metadataRetryBindPort {
		port := rand.Intn(1024)
		localTCPAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return errors.Wrap(err, "error resolving tcp address")
		}

		userdataClient := &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					LocalAddr:     localTCPAddr,
					DualStack:     false,
					FallbackDelay: time.Second * -1,
				}).DialContext,
			},
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPatch, meta.getMetadataURL()+"/user_data/"+key, bytes.NewBuffer(value))
		if err != nil {
			return errors.Wrap(err, "error creating patch userdata request")
		}
		request.Header.Set("Content-Type", "text/plain")
		resp, err := userdataClient.Do(request)
		if err != nil {
			retries++ // retry with a different source port
			continue
		}
		defer resp.Body.Close()

		return nil
	}
	return errors.New("too may bind port retries for SetUserData")
}

// DeleteUserData deletes the userdata key and the associated value
func (meta *MetadataAPI) DeleteUserData(key string) error {
	if key == "" {
		return errors.New("key must not be empty in DeleteUserData")
	}

	ctx := context.Background()
	retries := 0
	for retries <= metadataRetryBindPort {
		port := rand.Intn(1024)
		localTCPAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return errors.Wrap(err, "error resolving tcp address")
		}

		userdataClient := &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					LocalAddr:     localTCPAddr,
					DualStack:     false,
					FallbackDelay: time.Second * -1,
				}).DialContext,
			},
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodDelete, meta.getMetadataURL()+"/user_data/"+key, bytes.NewBufferString(""))
		if err != nil {
			return errors.Wrap(err, "error creating delete userdata request")
		}
		resp, err := userdataClient.Do(request)
		if err != nil {
			retries++ // retry with a different source port
			continue
		}
		defer resp.Body.Close()

		return nil
	}
	return errors.New("too may bind port retries for DeleteUserData")
}

// UserData represents the user data
type UserData struct {
	UserData []string `json:"user_data,omitempty"`
}
