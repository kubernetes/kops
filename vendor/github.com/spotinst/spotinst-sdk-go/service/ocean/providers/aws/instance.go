package aws

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

type Instance struct {
	ID               *string    `json:"instanceId,omitempty"`
	SpotRequestID    *string    `json:"spotInstanceRequestId,omitempty"`
	InstanceType     *string    `json:"instanceType,omitempty"`
	Status           *string    `json:"status,omitempty"`
	Product          *string    `json:"product,omitempty"`
	AvailabilityZone *string    `json:"availabilityZone,omitempty"`
	PrivateIP        *string    `json:"privateIp,omitempty"`
	PublicIP         *string    `json:"publicIp,omitempty"`
	CreatedAt        *time.Time `json:"createdAt,omitempty"`
}

type ListClusterInstancesInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type ListClusterInstancesOutput struct {
	Instances []*Instance `json:"instances,omitempty"`
}

type DetachClusterInstancesInput struct {
	ClusterID                     *string  `json:"clusterId,omitempty"`
	InstanceIDs                   []string `json:"instancesToDetach,omitempty"`
	ShouldDecrementTargetCapacity *bool    `json:"shouldDecrementTargetCapacity,omitempty"`
	ShouldTerminateInstances      *bool    `json:"shouldTerminateInstances,omitempty"`
}

type DetachClusterInstancesOutput struct{}

func instanceFromJSON(in []byte) (*Instance, error) {
	b := new(Instance)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func instancesFromJSON(in []byte) ([]*Instance, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Instance, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := instanceFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func instancesFromHttpResponse(resp *http.Response) ([]*Instance, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return instancesFromJSON(body)
}

func (s *ServiceOp) ListClusterInstances(ctx context.Context, input *ListClusterInstancesInput) (*ListClusterInstancesOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}/instances", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	instances, err := instancesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListClusterInstancesOutput{Instances: instances}, nil
}

func (s *ServiceOp) DetachClusterInstances(ctx context.Context, input *DetachClusterInstancesInput) (*DetachClusterInstancesOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{clusterId}/detachInstances", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.ClusterID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DetachClusterInstancesOutput{}, nil
}
