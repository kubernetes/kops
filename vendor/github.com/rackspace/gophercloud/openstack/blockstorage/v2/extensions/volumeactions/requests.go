package volumeactions

import (
	"github.com/rackspace/gophercloud"
)

// AttachOptsBuilder allows extensions to add additional parameters to the
// Attach request.
type AttachOptsBuilder interface {
	ToVolumeAttachMap() (map[string]interface{}, error)
}

// AttachMode describes the attachment mode for volumes.
type AttachMode string

// These constants determine how a volume is attached
const (
	ReadOnly  AttachMode = "ro"
	ReadWrite AttachMode = "rw"
)

// AttachOpts contains options for attaching a Volume.
type AttachOpts struct {
	// The mountpoint of this volume
	MountPoint string
	// The nova instance ID, can't set simultaneously with HostName
	InstanceUUID string
	// The hostname of baremetal host, can't set simultaneously with InstanceUUID
	HostName string
	// Mount mode of this volume
	Mode AttachMode
}

// ToVolumeAttachMap assembles a request body based on the contents of a
// AttachOpts.
func (opts AttachOpts) ToVolumeAttachMap() (map[string]interface{}, error) {
	v := make(map[string]interface{})

	if opts.MountPoint != "" {
		v["mountpoint"] = opts.MountPoint
	}
	if opts.Mode != "" {
		v["mode"] = opts.Mode
	}
	if opts.InstanceUUID != "" {
		v["instance_uuid"] = opts.InstanceUUID
	}
	if opts.HostName != "" {
		v["host_name"] = opts.HostName
	}

	return map[string]interface{}{"os-attach": v}, nil
}

// Attach will attach a volume based on the values in AttachOpts.
func Attach(client *gophercloud.ServiceClient, id string, opts AttachOptsBuilder) AttachResult {
	var res AttachResult

	reqBody, err := opts.ToVolumeAttachMap()
	if err != nil {
		res.Err = err
		return res
	}

	_, res.Err = client.Post(attachURL(client, id), reqBody, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})

	return res
}

// Attach will detach a volume based on volume id.
func Detach(client *gophercloud.ServiceClient, id string) DetachResult {
	var res DetachResult

	v := make(map[string]interface{})
	reqBody := map[string]interface{}{"os-detach": v}

	_, res.Err = client.Post(detachURL(client, id), reqBody, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})

	return res
}

// Reserve will reserve a volume based on volume id.
func Reserve(client *gophercloud.ServiceClient, id string) ReserveResult {
	var res ReserveResult

	v := make(map[string]interface{})
	reqBody := map[string]interface{}{"os-reserve": v}

	_, res.Err = client.Post(reserveURL(client, id), reqBody, nil, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201, 202},
	})

	return res
}

// Unreserve will unreserve a volume based on volume id.
func Unreserve(client *gophercloud.ServiceClient, id string) UnreserveResult {
	var res UnreserveResult

	v := make(map[string]interface{})
	reqBody := map[string]interface{}{"os-unreserve": v}

	_, res.Err = client.Post(unreserveURL(client, id), reqBody, nil, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201, 202},
	})

	return res
}

// ConnectorOptsBuilder allows extensions to add additional parameters to the
// InitializeConnection request.
type ConnectorOptsBuilder interface {
	ToConnectorMap() (map[string]interface{}, error)
}

// ConnectorOpts hosts options for InitializeConnection.
type ConnectorOpts struct {
	IP        string
	Host      string
	Initiator string
	Wwpns     []string
	Wwnns     string
	Multipath bool
	Platform  string
	OSType    string
}

// ToConnectorMap assembles a request body based on the contents of a
// ConnectorOpts.
func (opts ConnectorOpts) ToConnectorMap() (map[string]interface{}, error) {
	v := make(map[string]interface{})

	if opts.IP != "" {
		v["ip"] = opts.IP
	}
	if opts.Host != "" {
		v["host"] = opts.Host
	}
	if opts.Initiator != "" {
		v["initiator"] = opts.Initiator
	}
	if opts.Wwpns != nil {
		v["wwpns"] = opts.Wwpns
	}
	if opts.Wwnns != "" {
		v["wwnns"] = opts.Wwnns
	}

	v["multipath"] = opts.Multipath

	if opts.Platform != "" {
		v["platform"] = opts.Platform
	}
	if opts.OSType != "" {
		v["os_type"] = opts.OSType
	}

	return map[string]interface{}{"connector": v}, nil
}

// InitializeConnection initializes iscsi connection.
func InitializeConnection(client *gophercloud.ServiceClient, id string, opts *ConnectorOpts) InitializeConnectionResult {
	var res InitializeConnectionResult

	connctorMap, err := opts.ToConnectorMap()
	if err != nil {
		res.Err = err
		return res
	}

	reqBody := map[string]interface{}{"os-initialize_connection": connctorMap}

	_, res.Err = client.Post(initializeConnectionURL(client, id), reqBody, &res.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200, 201, 202},
	})

	return res
}

// TerminateConnection terminates iscsi connection.
func TerminateConnection(client *gophercloud.ServiceClient, id string, opts *ConnectorOpts) TerminateConnectionResult {
	var res TerminateConnectionResult

	connctorMap, err := opts.ToConnectorMap()
	if err != nil {
		res.Err = err
		return res
	}

	reqBody := map[string]interface{}{"os-terminate_connection": connctorMap}

	_, res.Err = client.Post(teminateConnectionURL(client, id), reqBody, nil, &gophercloud.RequestOpts{
		OkCodes: []int{202},
	})

	return res
}
