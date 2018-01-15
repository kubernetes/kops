package cs

import (
	"net/http"
	"net/url"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/util"
	"math"
	"time"
)

type ClusterState string

const (
	Initial      = ClusterState("initial")
	Running      = ClusterState("running")
	Updating     = ClusterState("updating")
	Scaling      = ClusterState("scaling")
	Failed       = ClusterState("failed")
	Deleting     = ClusterState("deleting")
	DeleteFailed = ClusterState("deleteFailed")
	Deleted      = ClusterState("deleted")
	InActive     = ClusterState("inactive")
)

type NodeStatus struct {
	Health   int64 `json:"health"`
	Unhealth int64 `json:"unhealth"`
}

type NetworkModeType string

const (
	ClassicNetwork = NetworkModeType("classic")
	VPCNetwork     = NetworkModeType("vpc")
)

// https://help.aliyun.com/document_detail/26053.html
type ClusterType struct {
	AgentVersion           string           `json:"agent_version"`
	ClusterID              string           `json:"cluster_id"`
	Name                   string           `json:"name"`
	Created                util.ISO6801Time `json:"created"`
	ExternalLoadbalancerID string           `json:"external_loadbalancer_id"`
	MasterURL              string           `json:"master_url"`
	NetworkMode            NetworkModeType  `json:"network_mode"`
	RegionID               common.Region    `json:"region_id"`
	SecurityGroupID        string           `json:"security_group_id"`
	Size                   int64            `json:"size"`
	State                  ClusterState     `json:"state"`
	Updated                util.ISO6801Time `json:"updated"`
	VPCID                  string           `json:"vpc_id"`
	VSwitchID              string           `json:"vswitch_id"`
	NodeStatus             string           `json:"node_status"`
	DockerVersion          string           `json:"docker_version"`
}

func (client *Client) DescribeClusters(nameFilter string) (clusters []ClusterType, err error) {
	query := make(url.Values)

	if nameFilter != "" {
		query.Add("name", nameFilter)
	}

	err = client.Invoke("", http.MethodGet, "/clusters", query, nil, &clusters)
	return
}

func (client *Client) DescribeCluster(id string) (cluster ClusterType, err error) {
	err = client.Invoke("", http.MethodGet, "/clusters/"+id, nil, nil, &cluster)
	return
}

type ClusterCreationArgs struct {
	Name             string           `json:"name"`
	Size             int64            `json:"size"`
	NetworkMode      NetworkModeType  `json:"network_mode"`
	SubnetCIDR       string           `json:"subnet_cidr,omitempty"`
	InstanceType     string           `json:"instance_type"`
	VPCID            string           `json:"vpc_id,omitempty"`
	VSwitchID        string           `json:"vswitch_id,omitempty"`
	Password         string           `json:"password"`
	DataDiskSize     int64            `json:"data_disk_size"`
	DataDiskCategory ecs.DiskCategory `json:"data_disk_category"`
	ECSImageID       string           `json:"ecs_image_id,omitempty"`
	IOOptimized      ecs.IoOptimized  `json:"io_optimized"`
}

type ClusterCreationResponse struct {
	Response
	ClusterID string `json:"cluster_id"`
}

func (client *Client) CreateCluster(region common.Region, args *ClusterCreationArgs) (cluster ClusterCreationResponse, err error) {
	err = client.Invoke(region, http.MethodPost, "/clusters", nil, args, &cluster)
	return
}

type ClusterResizeArgs struct {
	Size             int64            `json:"size"`
	InstanceType     string           `json:"instance_type"`
	Password         string           `json:"password"`
	DataDiskSize     int64            `json:"data_disk_size"`
	DataDiskCategory ecs.DiskCategory `json:"data_disk_category"`
	ECSImageID       string           `json:"ecs_image_id,omitempty"`
	IOOptimized      ecs.IoOptimized  `json:"io_optimized"`
}

func (client *Client) ResizeCluster(clusterID string, args *ClusterResizeArgs) error {
	return client.Invoke("", http.MethodPut, "/clusters/"+clusterID, nil, args, nil)
}

func (client *Client) DeleteCluster(clusterID string) error {
	return client.Invoke("", http.MethodDelete, "/clusters/"+clusterID, nil, nil, nil)
}

type ClusterCerts struct {
	CA   string `json:"ca,omitempty"`
	Key  string `json:"key,omitempty"`
	Cert string `json:"cert,omitempty"`
}

func (client *Client) GetClusterCerts(id string) (certs ClusterCerts, err error) {
	err = client.Invoke("", http.MethodGet, "/clusters/"+id+"/certs", nil, nil, &certs)
	return
}

const ClusterDefaultTimeout = 300
const DefaultWaitForInterval = 10
const DefaultPreSleepTime = 240

// WaitForCluster waits for instance to given status
// when instance.NotFound wait until timeout
func (client *Client) WaitForClusterAsyn(clusterId string, status ClusterState, timeout int) error {
	if timeout <= 0 {
		timeout = ClusterDefaultTimeout
	}
	cluster, err := client.DescribeCluster(clusterId)
	if err != nil {
		return err
	} else if cluster.State == status {
		//TODO
		return nil
	}
	// Create or Reset cluster usually cost at least 4 min, so there will sleep a long time before polling
	sleep := math.Min(float64(timeout), float64(DefaultPreSleepTime))
	time.Sleep(time.Duration(sleep) * time.Second)

	for {
		cluster, err := client.DescribeCluster(clusterId)
		if err != nil {
			return err
		} else if cluster.State == status {
			//TODO
			break
		}
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return common.GetClientErrorFromString("Timeout")
		}
		time.Sleep(DefaultWaitForInterval * time.Second)
	}
	return nil
}

func (client *Client) GetProjectClient(clusterId string) (projectClient *ProjectClient, err error) {
	cluster, err := client.DescribeCluster(clusterId)
	if err != nil {
		return
	}

	certs, err := client.GetClusterCerts(clusterId)
	if err != nil {
		return
	}

	projectClient, err = NewProjectClient(clusterId, cluster.MasterURL, certs)

	if err != nil {
		return
	}

	projectClient.SetDebug(client.debug)

	return
}
