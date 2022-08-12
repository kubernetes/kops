package spark

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
	"io/ioutil"
	"net/http"
)

//region Cluster
func (s *ServiceOp) ListClusters(ctx context.Context, input *ListClustersInput) (*ListClustersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/spark/cluster")

	if input != nil {
		if input.ControllerClusterID != nil {
			r.Params.Set("controllerClusterId", spotinst.StringValue(input.ControllerClusterID))
		}

		if input.ClusterState != nil {
			r.Params.Set("state", spotinst.StringValue(input.ClusterState))
		}
	}

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	clusters, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListClustersOutput{Clusters: clusters}, nil
}

func (s *ServiceOp) ReadCluster(ctx context.Context, input *ReadClusterInput) (*ReadClusterOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	path, err := uritemplates.Expand("/ocean/spark/cluster/{clusterId}", uritemplates.Values{
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

	clusters, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadClusterOutput)
	if len(clusters) > 0 {
		output.Cluster = clusters[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteCluster(ctx context.Context, input *DeleteClusterInput) (*DeleteClusterOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	path, err := uritemplates.Expand("/ocean/spark/cluster/{clusterId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteClusterOutput{}, nil
}

func (s *ServiceOp) CreateCluster(ctx context.Context, input *CreateClusterInput) (*CreateClusterOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}
	r := client.NewRequest(http.MethodPost, "/ocean/spark/cluster")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateCluster(ctx context.Context, input *UpdateClusterInput) (*UpdateClusterOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	path, err := uritemplates.Expand("/ocean/spark/cluster/{clusterId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &UpdateClusterOutput{}, nil
}

func clustersFromHttpResponse(resp *http.Response) ([]*Cluster, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return clustersFromJSON(body)
}

func clustersFromJSON(in []byte) ([]*Cluster, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Cluster, len(rw.Response.Items))
	for i, rb := range rw.Response.Items {
		b, err := clusterFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func clusterFromJSON(in []byte) (*Cluster, error) {
	b := new(Cluster)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

//endregion

//region Virtual Node Group
func (s *ServiceOp) DetachVirtualNodeGroup(ctx context.Context, input *DetachVngInput) (*DetachVngOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	path, err := uritemplates.Expand("/ocean/spark/cluster/{clusterId}/virtualNodeGroup/{vngId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
		"vngId":     spotinst.StringValue(input.VngID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DetachVngOutput{}, nil
}

func (s *ServiceOp) AttachVirtualNodeGroup(ctx context.Context, input *AttachVngInput) (*AttachVngOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	path, err := uritemplates.Expand("/ocean/spark/cluster/{clusterId}/virtualNodeGroup", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.ClusterID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodPost, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := vngsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(AttachVngOutput)
	if len(gs) > 0 {
		output.VirtualNodeGroup = gs[0]
	}

	return output, nil
}

func vngsFromHttpResponse(resp *http.Response) ([]*DedicatedVirtualNodeGroup, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return vngsFromJSON(body)
}

func vngsFromJSON(in []byte) ([]*DedicatedVirtualNodeGroup, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*DedicatedVirtualNodeGroup, len(rw.Response.Items))
	for i, rb := range rw.Response.Items {
		b, err := vngFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func vngFromJSON(in []byte) (*DedicatedVirtualNodeGroup, error) {
	b := new(DedicatedVirtualNodeGroup)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

//endregion
