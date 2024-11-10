package aws

import (
	"context"
	"encoding/json"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
	"io/ioutil"
	"net/http"
	"time"
)

type ClusterNodes struct {
	InstanceId                   *string  `json:"instanceId,omitempty"`
	InstanceType                 *string  `json:"instanceType,omitempty"`
	AvailabilityZone             *string  `json:"availabilityZone,omitempty"`
	LaunchSpecId                 *string  `json:"launchSpecId,omitempty"`
	LaunchSpecName               *string  `json:"launchSpecName,omitempty"`
	LifeCycle                    *string  `json:"lifeCycle,omitempty"`
	PublicIp                     *string  `json:"publicIp,omitempty"`
	NodeName                     *string  `json:"nodeName,omitempty"`
	RegistrationStatus           *string  `json:"registrationStatus,omitempty"`
	WorkloadRequestedMilliCpu    *int     `json:"workloadRequestedMilliCpu,omitempty"`
	WorkloadRequestedMemoryInMiB *int     `json:"workloadRequestedMemoryInMiB,omitempty"`
	WorkloadRequestedGpu         *int     `json:"workloadRequestedGpu,omitempty"`
	HeadroomRequestedMilliCpu    *int     `json:"headroomRequestedMilliCpu,omitempty"`
	HeadroomRequestedMemoryInMiB *int     `json:"headroomRequestedMemoryInMiB,omitempty"`
	HeadroomRequestedGpu         *int     `json:"headroomRequestedGpu,omitempty"`
	AllocatableMilliCpu          *int     `json:"allocatableMilliCpu,omitempty"`
	AllocatableMemoryInMiB       *float64 `json:"allocatableMemoryInMiB,omitempty"`
	AllocatableGpu               *int     `json:"allocatableGpu,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`

	// forceSendFields is a list of field names (e.g. "Keys") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	forceSendFields []string

	// nullFields is a list of field names (e.g. "Keys") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	nullFields []string
}
type ReadClusterNodeInput struct {
	ClusterID    *string `json:"clusterId,omitempty"`
	LaunchSpecId *string `json:"launchSpecId,omitempty"`
	InstanceId   *string `json:"instanceId,omitempty"`
}

type ReadClusterNodeOutput struct {
	ClusterNode []*ClusterNodes `json:"clusterNode,omitempty"`
}

func (s *ServiceOp) ReadClusterNodes(ctx context.Context, input *ReadClusterNodeInput) (*ReadClusterNodeOutput, error) {
	path, err := uritemplates.Expand("/ocean/aws/k8s/cluster/{oceanClusterId}/nodes", uritemplates.Values{
		"oceanClusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}
	r := client.NewRequest(http.MethodGet, path)
	if input.LaunchSpecId != nil {
		r.Params["launchSpecId"] = []string{spotinst.StringValue(input.LaunchSpecId)}
	}
	if input.InstanceId != nil {
		r.Params["instanceId"] = []string{spotinst.StringValue(input.InstanceId)}
	}
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	gs, err := nodesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ReadClusterNodeOutput{ClusterNode: gs}, nil
}

func nodesFromHttpResponse(resp *http.Response) ([]*ClusterNodes, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nodesFromJSON(body)
}

func nodesFromJSON(in []byte) ([]*ClusterNodes, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*ClusterNodes, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := nodeFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func nodeFromJSON(in []byte) (*ClusterNodes, error) {
	b := new(ClusterNodes)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}
