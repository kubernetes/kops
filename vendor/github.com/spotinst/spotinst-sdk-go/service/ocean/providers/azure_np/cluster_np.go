package azure_np

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

type Cluster struct {
	ID                       *string                   `json:"id,omitempty"`
	Name                     *string                   `json:"name,omitempty"`
	ControllerClusterID      *string                   `json:"controllerClusterId,omitempty"`
	AKS                      *AKS                      `json:"aks,omitempty"`
	AutoScaler               *AutoScaler               `json:"autoScaler,omitempty"`
	Health                   *Health                   `json:"health,omitempty"`
	VirtualNodeGroupTemplate *VirtualNodeGroupTemplate `json:"virtualNodeGroupTemplate,omitempty"`
	Scheduling               *Scheduling               `json:"scheduling,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AKS struct {
	ClusterName                     *string `json:"clusterName,omitempty"`
	ResourceGroupName               *string `json:"resourceGroupName,omitempty"`
	Region                          *string `json:"region,omitempty"`
	InfrastructureResourceGroupName *string `json:"infrastructureResourceGroupName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AutoScaler struct {
	IsEnabled      *bool           `json:"isEnabled,omitempty"`
	ResourceLimits *ResourceLimits `json:"resourceLimits,omitempty"`
	Down           *Down           `json:"down,omitempty"`
	Headroom       *Headroom       `json:"headroom,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Health struct {
	GracePeriod *int `json:"gracePeriod,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VirtualNodeGroupTemplate struct {
	AvailabilityZones  []string            `json:"availabilityZones,omitempty"`
	NodePoolProperties *NodePoolProperties `json:"nodePoolProperties,omitempty"`
	NodeCountLimits    *NodeCountLimits    `json:"nodeCountLimits,omitempty"`
	Strategy           *Strategy           `json:"strategy,omitempty"`
	Labels             *map[string]string  `json:"labels,omitempty"`
	Tags               *map[string]string  `json:"tags,omitempty"`
	Taints             []*Taint            `json:"taints,omitempty"`
	AutoScale          *AutoScale          `json:"autoScale,omitempty"`
	VmSizes            *VmSizes            `json:"vmSizes,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ResourceLimits struct {
	MaxVCPU      *int `json:"maxVCpu,omitempty"`
	MaxMemoryGib *int `json:"maxMemoryGib,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Down struct {
	MaxScaleDownPercentage *int `json:"maxScaleDownPercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Headroom struct {
	Automatic *Automatic `json:"automatic,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Automatic struct {
	IsEnabled  *bool `json:"isEnabled,omitempty"`
	Percentage *int  `json:"percentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListClustersInput struct{}

type ListClustersOutput struct {
	Clusters []*Cluster `json:"clusters,omitempty"`
}

type CreateClusterInput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type CreateClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type ReadClusterInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type ReadClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type UpdateClusterInput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type UpdateClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type DeleteClusterInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type DeleteClusterOutput struct{}

// region Unmarshalls

func clusterFromJSON(in []byte) (*Cluster, error) {
	b := new(Cluster)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func clustersFromJSON(in []byte) ([]*Cluster, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Cluster, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := clusterFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func clustersFromHttpResponse(resp *http.Response) ([]*Cluster, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return clustersFromJSON(body)
}

func clusterImportFromJSON(in []byte) (*ImportClusterOutput, error) {
	b := new(ImportClusterOutput)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func clustersImportFromJSON(in []byte) ([]*ImportClusterOutput, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*ImportClusterOutput, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := clusterImportFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func clustersImportFromHttpResponse(resp *http.Response) ([]*ImportClusterOutput, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return clustersImportFromJSON(body)
}

// endregion

// region API requests

func (s *ServiceOp) ListClusters(ctx context.Context) (*ListClustersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/ocean/azure/np/cluster")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListClustersOutput{Clusters: gs}, nil
}

func (s *ServiceOp) CreateCluster(ctx context.Context, input *CreateClusterInput) (*CreateClusterOutput, error) {
	r := client.NewRequest(http.MethodPost, "/ocean/azure/np/cluster")
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

func (s *ServiceOp) ReadCluster(ctx context.Context, input *ReadClusterInput) (*ReadClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/azure/np/cluster/{clusterId}", uritemplates.Values{
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

	gs, err := clustersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateCluster(ctx context.Context, input *UpdateClusterInput) (*UpdateClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/azure/np/cluster/{clusterId}", uritemplates.Values{
		"clusterId": spotinst.StringValue(input.Cluster.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do NOT need the ID anymore, so let's drop it.
	input.Cluster.ID = nil

	r := client.NewRequest(http.MethodPut, path)
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

	output := new(UpdateClusterOutput)
	if len(gs) > 0 {
		output.Cluster = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteCluster(ctx context.Context, input *DeleteClusterInput) (*DeleteClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/azure/np/cluster/{clusterId}", uritemplates.Values{
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

func (s *ServiceOp) ImportCluster(ctx context.Context, input *ImportClusterInput) (*ImportClusterOutput, error) {
	path, err := uritemplates.Expand("/ocean/azure/np/cluster/aks/import/{acdIdentifier}", uritemplates.Values{
		"acdIdentifier": spotinst.StringValue(input.ACDIdentifier),
	})
	if err != nil {
		return nil, err
	}

	// We do NOT need the ID anymore, so let's drop it.
	input.ACDIdentifier = nil

	r := client.NewRequest(http.MethodPost, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := clustersImportFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ImportClusterOutput)
	if len(gs) > 0 {
		output = gs[0]
	}

	return output, nil
}

// endregion

// region Cluster

func (o Cluster) MarshalJSON() ([]byte, error) {
	type noMethod Cluster
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Cluster) SetId(v *string) *Cluster {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Cluster) SetName(v *string) *Cluster {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Cluster) SetControllerClusterId(v *string) *Cluster {
	if o.ControllerClusterID = v; o.ControllerClusterID == nil {
		o.nullFields = append(o.nullFields, "ControllerClusterID")
	}
	return o
}

func (o *Cluster) SetAKS(v *AKS) *Cluster {
	if o.AKS = v; o.AKS == nil {
		o.nullFields = append(o.nullFields, "AKS")
	}
	return o
}

func (o *Cluster) SetAutoScaler(v *AutoScaler) *Cluster {
	if o.AutoScaler = v; o.AutoScaler == nil {
		o.nullFields = append(o.nullFields, "AutoScaler")
	}
	return o
}

func (o *Cluster) SetHealth(v *Health) *Cluster {
	if o.Health = v; o.Health == nil {
		o.nullFields = append(o.nullFields, "Health")
	}
	return o
}

func (o *Cluster) SetVirtualNodeGroupTemplate(v *VirtualNodeGroupTemplate) *Cluster {
	if o.VirtualNodeGroupTemplate = v; o.VirtualNodeGroupTemplate == nil {
		o.nullFields = append(o.nullFields, "VirtualNodeGroupTemplate")
	}
	return o
}

func (o *Cluster) SetScheduling(v *Scheduling) *Cluster {
	if o.Scheduling = v; o.Scheduling == nil {
		o.nullFields = append(o.nullFields, "Scheduling")
	}
	return o
}

// endregion

// region AKS

func (o AKS) MarshalJSON() ([]byte, error) {
	type noMethod AKS
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AKS) SetClusterName(v *string) *AKS {
	if o.ClusterName = v; o.ClusterName == nil {
		o.nullFields = append(o.nullFields, "ClusterName")
	}
	return o
}

func (o *AKS) SetResourceGroupName(v *string) *AKS {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *AKS) SetRegion(v *string) *AKS {
	if o.Region = v; o.Region == nil {
		o.nullFields = append(o.nullFields, "Region")
	}
	return o
}

func (o *AKS) SetInfrastructureResourceGroupName(v *string) *AKS {
	if o.InfrastructureResourceGroupName = v; o.InfrastructureResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "InfrastructureResourceGroupName")
	}
	return o
}

// endregion

// region AutoScaler

func (o AutoScaler) MarshalJSON() ([]byte, error) {
	type noMethod AutoScaler
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScaler) SetIsEnabled(v *bool) *AutoScaler {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *AutoScaler) SetResourceLimits(v *ResourceLimits) *AutoScaler {
	if o.ResourceLimits = v; o.ResourceLimits == nil {
		o.nullFields = append(o.nullFields, "ResourceLimits")
	}
	return o
}

func (o *AutoScaler) SetDown(v *Down) *AutoScaler {
	if o.Down = v; o.Down == nil {
		o.nullFields = append(o.nullFields, "Down")
	}
	return o
}

func (o *AutoScaler) SetHeadroom(v *Headroom) *AutoScaler {
	if o.Headroom = v; o.Headroom == nil {
		o.nullFields = append(o.nullFields, "Headroom")
	}
	return o
}

// endregion

// region ResourceLimits

func (o ResourceLimits) MarshalJSON() ([]byte, error) {
	type noMethod ResourceLimits
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ResourceLimits) SetMaxVcpu(v *int) *ResourceLimits {
	if o.MaxVCPU = v; o.MaxVCPU == nil {
		o.nullFields = append(o.nullFields, "MaxVCPU")
	}
	return o
}

func (o *ResourceLimits) SetMaxMemoryGib(v *int) *ResourceLimits {
	if o.MaxMemoryGib = v; o.MaxMemoryGib == nil {
		o.nullFields = append(o.nullFields, "MaxMemoryGib")
	}
	return o
}

//end region

// region Down

func (o Down) MarshalJSON() ([]byte, error) {
	type noMethod Down
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Down) SetMaxScaleDownPercentage(v *int) *Down {
	if o.MaxScaleDownPercentage = v; o.MaxScaleDownPercentage == nil {
		o.nullFields = append(o.nullFields, "MaxScaleDownPercentage")
	}
	return o
}

//end region

// region Headroom

func (o Headroom) MarshalJSON() ([]byte, error) {
	type noMethod Headroom
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Headroom) SetAutomatic(v *Automatic) *Headroom {
	if o.Automatic = v; o.Automatic == nil {
		o.nullFields = append(o.nullFields, "Automatic")
	}
	return o
}

//end region

// region Automatic

func (o Automatic) MarshalJSON() ([]byte, error) {
	type noMethod Automatic
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

/*func (o *Automatic) SetIsEnabled(v *bool) *Automatic {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}*/

func (o *Automatic) SetPercentage(v *int) *Automatic {
	if o.Percentage = v; o.Percentage == nil {
		o.nullFields = append(o.nullFields, "Percentage")
	}
	return o
}

//end region

// region VirtualNodeGroupTemplate

func (o VirtualNodeGroupTemplate) MarshalJSON() ([]byte, error) {
	type noMethod VirtualNodeGroupTemplate
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VirtualNodeGroupTemplate) SetAvailabilityZones(v []string) *VirtualNodeGroupTemplate {
	if o.AvailabilityZones = v; o.AvailabilityZones == nil {
		o.nullFields = append(o.nullFields, "AvailabilityZones")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetNodePoolProperties(v *NodePoolProperties) *VirtualNodeGroupTemplate {
	if o.NodePoolProperties = v; o.NodePoolProperties == nil {
		o.nullFields = append(o.nullFields, "NodePoolProperties")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetNodeCountLimits(v *NodeCountLimits) *VirtualNodeGroupTemplate {
	if o.NodeCountLimits = v; o.NodeCountLimits == nil {
		o.nullFields = append(o.nullFields, "NodeCountLimits")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetStrategy(v *Strategy) *VirtualNodeGroupTemplate {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetLabels(v *map[string]string) *VirtualNodeGroupTemplate {
	if o.Labels = v; o.Labels == nil {
		o.nullFields = append(o.nullFields, "Labels")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetTaints(v []*Taint) *VirtualNodeGroupTemplate {
	if o.Taints = v; o.Taints == nil {
		o.nullFields = append(o.nullFields, "Taints")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetTags(v *map[string]string) *VirtualNodeGroupTemplate {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetAutoScale(v *AutoScale) *VirtualNodeGroupTemplate {
	if o.AutoScale = v; o.AutoScale == nil {
		o.nullFields = append(o.nullFields, "AutoScale")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetVmSizes(v *VmSizes) *VirtualNodeGroupTemplate {
	if o.VmSizes = v; o.VmSizes == nil {
		o.nullFields = append(o.nullFields, "VmSizes")
	}
	return o
}

// endregion

// region Health

func (o Health) MarshalJSON() ([]byte, error) {
	type noMethod Health
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Health) SetGracePeriod(v *int) *Health {
	if o.GracePeriod = v; o.GracePeriod == nil {
		o.nullFields = append(o.nullFields, "GracePeriod")
	}
	return o
}

// region Import

type ImportCluster struct {
	ControllerClusterID *string `json:"controllerClusterId,omitempty"`
	Name                *string `json:"name,omitempty"`
	AKS                 *AKS    `json:"aks,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ImportClusterInput struct {
	ACDIdentifier *string        `json:"acdIdentifier,omitempty"`
	Cluster       *ImportCluster `json:"cluster,omitempty"`
}

type ImportClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

// endregion
