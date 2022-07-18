package instance

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
)

var (
	metadataURL           = "http://169.254.42.42"
	metadataRetryBindPort = 200
)

// MetadataAPI metadata API
type MetadataAPI struct {
}

// NewMetadataAPI returns a MetadataAPI object from a Scaleway client.
func NewMetadataAPI() *MetadataAPI {
	return &MetadataAPI{}
}

// GetMetadata returns the metadata available from the server
func (*MetadataAPI) GetMetadata() (m *Metadata, err error) {
	resp, err := http.Get(metadataURL + "/conf?format=json")
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

// Metadata represents the struct return by the metadata API
type Metadata struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	Organization   string `json:"organization,omitempty"`
	Project        string `json:"project,omitempty"`
	CommercialType string `json:"commercial_type,omitempty"`
	PublicIP       struct {
		Dynamic bool   `json:"dynamic,omitempty"`
		ID      string `json:"id,omitempty"`
		Address string `json:"address,omitempty"`
	} `json:"public_ip,omitempty"`
	PrivateIP string `json:"private_ip,omitempty"`
	IPv6      struct {
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
func (*MetadataAPI) ListUserData() (res *UserData, err error) {
	retries := 0
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

		resp, err := userdataClient.Get(metadataURL + "/user_data?format=json")
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
func (*MetadataAPI) GetUserData(key string) ([]byte, error) {
	if key == "" {
		return make([]byte, 0), errors.New("key must not be empty in GetUserData")
	}

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

		resp, err := userdataClient.Get(metadataURL + "/user_data/" + key)
		if err != nil {
			retries++ // retry with a different source port
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return make([]byte, 0), errors.Wrap(err, "error reading userdata body")
		}

		return body, nil
	}
	return make([]byte, 0), errors.New("too may bind port retries for GetUserData")
}

// SetUserData sets the userdata key with the given value
func (*MetadataAPI) SetUserData(key string, value []byte) error {
	if key == "" {
		return errors.New("key must not be empty in SetUserData")
	}

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
		request, err := http.NewRequest("PATCH", metadataURL+"/user_data/"+key, bytes.NewBuffer(value))
		if err != nil {
			return errors.Wrap(err, "error creating patch userdata request")
		}
		request.Header.Set("Content-Type", "text/plain")
		_, err = userdataClient.Do(request)
		if err != nil {
			retries++ // retry with a different source port
			continue
		}

		return nil
	}
	return errors.New("too may bind port retries for SetUserData")
}

// DeleteUserData deletes the userdata key and the associated value
func (*MetadataAPI) DeleteUserData(key string) error {
	if key == "" {
		return errors.New("key must not be empty in DeleteUserData")
	}

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
		request, err := http.NewRequest("DELETE", metadataURL+"/user_data/"+key, bytes.NewBuffer([]byte("")))
		if err != nil {
			return errors.Wrap(err, "error creating delete userdata request")
		}
		_, err = userdataClient.Do(request)
		if err != nil {
			retries++ // retry with a different source port
			continue
		}

		return nil
	}
	return errors.New("too may bind port retries for DeleteUserData")
}

// UserData represents the user data
type UserData struct {
	UserData []string `json:"user_data,omitempty"`
}
