package azure

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
	ControllerClusterID      *string                   `json:"controllerClusterId,omitempty"`
	Name                     *string                   `json:"name,omitempty"`
	AKS                      *AKS                      `json:"aks,omitempty"`
	AutoScaler               *AutoScaler               `json:"autoScaler,omitempty"`
	Strategy                 *Strategy                 `json:"strategy,omitempty"`
	Health                   *Health                   `json:"health,omitempty"`
	VirtualNodeGroupTemplate *VirtualNodeGroupTemplate `json:"virtualNodeGroupTemplate,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AKS struct {
	Name              *string `json:"name,omitempty"`
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`

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

type Strategy struct {
	SpotPercentage *int  `json:"spotPercentage,omitempty"`
	FallbackToOD   *bool `json:"fallbackToOd,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Health struct {
	GracePeriod *int `json:"gracePeriod,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VirtualNodeGroupTemplate struct {
	VMSizes             *VMSizes             `json:"vmSizes,omitempty"`
	LaunchSpecification *LaunchSpecification `json:"launchSpecification,omitempty"`
	Zones               []string             `json:"zones,omitempty"`

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
	MaxScaleDownPercentage *float64 `json:"maxScaleDownPercentage,omitempty"`
	FallbackToOD           *bool    `json:"fallbackToOd,omitempty"`

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

type VMSizes struct {
	Whitelist []string `json:"whitelist,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LaunchSpecification struct {
	ResourceGroupName        *string                   `json:"resourceGroupName,omitempty"`
	CustomData               *string                   `json:"customData,omitempty"`
	Image                    *Image                    `json:"image,omitempty"`
	Network                  *Network                  `json:"network,omitempty"`
	OSDisk                   *OSDisk                   `json:"osDisk,omitempty"`
	Login                    *Login                    `json:"login,omitempty"`
	LoadBalancersConfig      *LoadBalancersConfig      `json:"loadBalancersConfig,omitempty"`
	ManagedServiceIdentities []*ManagedServiceIdentity `json:"managedServiceIdentities,omitempty"`
	Extensions               []*Extension              `json:"extensions,omitempty"`
	Tags                     []*Tag                    `json:"tags,omitempty"`
	MaxPods                  *int                      `json:"maxPods,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Extension struct {
	APIVersion              *string     `json:"apiVersion,omitempty"`
	MinorVersionAutoUpgrade *bool       `json:"minorVersionAutoUpgrade,omitempty"`
	Name                    *string     `json:"name,omitempty"`
	Publisher               *string     `json:"publisher,omitempty"`
	Type                    *string     `json:"type,omitempty"`
	ProtectedSettings       interface{} `json:"protectedSettings,omitempty"`
	PublicSettings          interface{} `json:"publicSettings,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Image struct {
	MarketplaceImage *MarketplaceImage `json:"marketplace,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LoadBalancersConfig struct {
	LoadBalancers []*LoadBalancer `json:"loadBalancers,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Login struct {
	SSHPublicKey *string `json:"sshPublicKey,omitempty"`
	UserName     *string `json:"userName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Network struct {
	NetworkInterfaces  []*NetworkInterface `json:"networkInterfaces,omitempty"`
	ResourceGroupName  *string             `json:"resourceGroupName,omitempty"`
	VirtualNetworkName *string             `json:"virtualNetworkName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type MarketplaceImage struct {
	Publisher *string `json:"publisher,omitempty"`
	Offer     *string `json:"offer,omitempty"`
	SKU       *string `json:"sku,omitempty"`
	Version   *string `json:"version,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LoadBalancer struct {
	BackendPoolNames  []string `json:"backendPoolNames,omitempty"`
	LoadBalancerSKU   *string  `json:"loadBalancerSku,omitempty"`
	Name              *string  `json:"name,omitempty"`
	ResourceGroupName *string  `json:"resourceGroupName,omitempty"`
	Type              *string  `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type NetworkInterface struct {
	SubnetName          *string               `json:"subnetName,omitempty"`
	AssignPublicIP      *bool                 `json:"assignPublicIp,omitempty"`
	IsPrimary           *bool                 `json:"isPrimary,omitempty"`
	EnableIPForwarding  *bool                 `json:"enableIPForwarding,omitempty"`
	PublicIPSKU         *string               `json:"publicIpSku,omitempty"`
	SecurityGroup       *SecurityGroup        `json:"securityGroup,omitempty"`
	AdditionalIPConfigs []*AdditionalIPConfig `json:"additionalIpConfigurations,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AdditionalIPConfig struct {
	Name                    *string `json:"name,omitempty"`
	PrivateIPAddressVersion *string `json:"privateIpAddressVersion,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type SecurityGroup struct {
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`
	Name              *string `json:"name,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ManagedServiceIdentity struct {
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`
	Name              *string `json:"name,omitempty"`

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
	r := client.NewRequest(http.MethodGet, "/ocean/azure/k8s/cluster")
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
	r := client.NewRequest(http.MethodPost, "/ocean/azure/k8s/cluster")
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
	path, err := uritemplates.Expand("/ocean/azure/k8s/cluster/{clusterId}", uritemplates.Values{
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
	path, err := uritemplates.Expand("/ocean/azure/k8s/cluster/{clusterId}", uritemplates.Values{
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
	path, err := uritemplates.Expand("/ocean/azure/k8s/cluster/{clusterId}", uritemplates.Values{
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
	path, err := uritemplates.Expand("/ocean/azure/k8s/cluster/aks/import/{acdIdentifier}", uritemplates.Values{
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

func (o *Cluster) SetControllerClusterId(v *string) *Cluster {
	if o.ControllerClusterID = v; o.ControllerClusterID == nil {
		o.nullFields = append(o.nullFields, "ControllerClusterID")
	}
	return o
}

func (o *Cluster) SetName(v *string) *Cluster {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Cluster) SetAKS(v *AKS) *Cluster {
	if o.AKS = v; o.AKS == nil {
		o.nullFields = append(o.nullFields, "AKS")
	}
	return o
}

func (o *Cluster) SetStrategy(v *Strategy) *Cluster {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *Cluster) SetHealth(v *Health) *Cluster {
	if o.Health = v; o.Health == nil {
		o.nullFields = append(o.nullFields, "Health")
	}
	return o
}

func (o *Cluster) SetAutoScaler(v *AutoScaler) *Cluster {
	if o.AutoScaler = v; o.AutoScaler == nil {
		o.nullFields = append(o.nullFields, "AutoScaler")
	}
	return o
}

func (o *Cluster) SetVirtualNodeGroupTemplate(v *VirtualNodeGroupTemplate) *Cluster {
	if o.VirtualNodeGroupTemplate = v; o.VirtualNodeGroupTemplate == nil {
		o.nullFields = append(o.nullFields, "VirtualNodeGroupTemplate")
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

func (o *AKS) SetName(v *string) *AKS {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *AKS) SetResourceGroupName(v *string) *AKS {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

// endregion

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

// region Strategy

func (o Strategy) MarshalJSON() ([]byte, error) {
	type noMethod Strategy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Strategy) SetFallbackToOD(v *bool) *Strategy {
	if o.FallbackToOD = v; o.FallbackToOD == nil {
		o.nullFields = append(o.nullFields, "FallbackToOD")
	}
	return o
}

func (o *Strategy) SetSpotPercentage(v *int) *Strategy {
	if o.SpotPercentage = v; o.SpotPercentage == nil {
		o.nullFields = append(o.nullFields, "SpotPercentage")
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

// endregion

// region VirtualNodeGroupTemplate

func (o VirtualNodeGroupTemplate) MarshalJSON() ([]byte, error) {
	type noMethod VirtualNodeGroupTemplate
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VirtualNodeGroupTemplate) SetVMSizes(v *VMSizes) *VirtualNodeGroupTemplate {
	if o.VMSizes = v; o.VMSizes == nil {
		o.nullFields = append(o.nullFields, "VMSizes")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetLaunchSpecification(v *LaunchSpecification) *VirtualNodeGroupTemplate {
	if o.LaunchSpecification = v; o.LaunchSpecification == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecification")
	}
	return o
}

func (o *VirtualNodeGroupTemplate) SetZones(v []string) *VirtualNodeGroupTemplate {
	if o.Zones = v; o.Zones == nil {
		o.nullFields = append(o.nullFields, "Zones")
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

func (o *ResourceLimits) SetMaxVCPU(v *int) *ResourceLimits {
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

// endregion

// region Down

func (o Down) MarshalJSON() ([]byte, error) {
	type noMethod Down
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Down) SetMaxScaleDownPercentage(v *float64) *Down {
	if o.MaxScaleDownPercentage = v; o.MaxScaleDownPercentage == nil {
		o.nullFields = append(o.nullFields, "MaxScaleDownPercentage")
	}
	return o
}

func (o *Down) SetFallbackToOD(v *bool) *Down {
	if o.FallbackToOD = v; o.FallbackToOD == nil {
		o.nullFields = append(o.nullFields, "FallbackToOD")
	}
	return o
}

// endregion

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

// endregion

// region Automatic

func (o Automatic) MarshalJSON() ([]byte, error) {
	type noMethod Automatic
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Automatic) SetIsEnabled(v *bool) *Automatic {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *Automatic) SetPercentage(v *int) *Automatic {
	if o.Percentage = v; o.Percentage == nil {
		o.nullFields = append(o.nullFields, "Percentage")
	}
	return o
}

// endregion

// region VMSizes

func (o VMSizes) MarshalJSON() ([]byte, error) {
	type noMethod VMSizes
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VMSizes) SetWhitelist(v []string) *VMSizes {
	if o.Whitelist = v; o.Whitelist == nil {
		o.nullFields = append(o.nullFields, "Whitelist")
	}
	return o
}

// endregion

// region LaunchSpecification

func (o LaunchSpecification) MarshalJSON() ([]byte, error) {
	type noMethod LaunchSpecification
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LaunchSpecification) SetResourceGroupName(v *string) *LaunchSpecification {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *LaunchSpecification) SetCustomData(v *string) *LaunchSpecification {
	if o.CustomData = v; o.CustomData == nil {
		o.nullFields = append(o.nullFields, "CustomData")
	}
	return o
}

func (o *LaunchSpecification) SetImage(v *Image) *LaunchSpecification {
	if o.Image = v; o.Image == nil {
		o.nullFields = append(o.nullFields, "Image")
	}
	return o
}

func (o *LaunchSpecification) SetNetwork(v *Network) *LaunchSpecification {
	if o.Network = v; o.Network == nil {
		o.nullFields = append(o.nullFields, "Network")
	}
	return o
}

func (o *LaunchSpecification) SetLogin(v *Login) *LaunchSpecification {
	if o.Login = v; o.Login == nil {
		o.nullFields = append(o.nullFields, "Login")
	}
	return o
}

func (o *LaunchSpecification) SetManagedServiceIdentities(v []*ManagedServiceIdentity) *LaunchSpecification {
	if o.ManagedServiceIdentities = v; o.ManagedServiceIdentities == nil {
		o.nullFields = append(o.nullFields, "ManagedServiceIdentities")
	}
	return o
}

func (o *LaunchSpecification) SetExtensions(v []*Extension) *LaunchSpecification {
	if o.Extensions = v; o.Extensions == nil {
		o.nullFields = append(o.nullFields, "Extensions")
	}
	return o
}

func (o *LaunchSpecification) SetLoadBalancersConfig(v *LoadBalancersConfig) *LaunchSpecification {
	if o.LoadBalancersConfig = v; o.LoadBalancersConfig == nil {
		o.nullFields = append(o.nullFields, "LoadBalancersConfig")
	}
	return o
}

func (o *LaunchSpecification) SetOSDisk(v *OSDisk) *LaunchSpecification {
	if o.OSDisk = v; o.OSDisk == nil {
		o.nullFields = append(o.nullFields, "OSDisk")
	}
	return o
}

func (o *LaunchSpecification) SetTags(v []*Tag) *LaunchSpecification {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

func (o *LaunchSpecification) SetMaxPods(v *int) *LaunchSpecification {
	if o.MaxPods = v; o.MaxPods == nil {
		o.nullFields = append(o.nullFields, "MaxPods")
	}
	return o
}

// endregion

// region Image

func (o Image) MarshalJSON() ([]byte, error) {
	type noMethod Image
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Image) SetMarketplaceImage(v *MarketplaceImage) *Image {
	if o.MarketplaceImage = v; o.MarketplaceImage == nil {
		o.nullFields = append(o.nullFields, "MarketplaceImage")
	}
	return o
}

// endregion

// region MarketplaceImage

func (o MarketplaceImage) MarshalJSON() ([]byte, error) {
	type noMethod MarketplaceImage
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *MarketplaceImage) SetPublisher(v *string) *MarketplaceImage {
	if o.Publisher = v; o.Publisher == nil {
		o.nullFields = append(o.nullFields, "Publisher")
	}
	return o
}

func (o *MarketplaceImage) SetOffer(v *string) *MarketplaceImage {
	if o.Offer = v; o.Offer == nil {
		o.nullFields = append(o.nullFields, "Offer")
	}
	return o
}

func (o *MarketplaceImage) SetSKU(v *string) *MarketplaceImage {
	if o.SKU = v; o.SKU == nil {
		o.nullFields = append(o.nullFields, "SKU")
	}
	return o
}

func (o *MarketplaceImage) SetVersion(v *string) *MarketplaceImage {
	if o.Version = v; o.Version == nil {
		o.nullFields = append(o.nullFields, "Version")
	}
	return o
}

// endregion

// region Network

func (o Network) MarshalJSON() ([]byte, error) {
	type noMethod Network
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Network) SetVirtualNetworkName(v *string) *Network {
	if o.VirtualNetworkName = v; o.VirtualNetworkName == nil {
		o.nullFields = append(o.nullFields, "VirtualNetworkName")
	}
	return o
}

func (o *Network) SetResourceGroupName(v *string) *Network {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *Network) SetNetworkInterfaces(v []*NetworkInterface) *Network {
	if o.NetworkInterfaces = v; o.NetworkInterfaces == nil {
		o.nullFields = append(o.nullFields, "NetworkInterfaces")
	}
	return o
}

// endregion

// region NetworkInterface

func (o NetworkInterface) MarshalJSON() ([]byte, error) {
	type noMethod NetworkInterface
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *NetworkInterface) SetSubnetName(v *string) *NetworkInterface {
	if o.SubnetName = v; o.SubnetName == nil {
		o.nullFields = append(o.nullFields, "SubnetName")
	}
	return o
}

func (o *NetworkInterface) SetAdditionalIPConfigs(v []*AdditionalIPConfig) *NetworkInterface {
	if o.AdditionalIPConfigs = v; o.AdditionalIPConfigs == nil {
		o.nullFields = append(o.nullFields, "AdditionalIPConfigs")
	}
	return o
}

func (o *NetworkInterface) SetAssignPublicIP(v *bool) *NetworkInterface {
	if o.AssignPublicIP = v; o.AssignPublicIP == nil {
		o.nullFields = append(o.nullFields, "AssignPublicIP")
	}
	return o
}

func (o *NetworkInterface) SetIsPrimary(v *bool) *NetworkInterface {
	if o.IsPrimary = v; o.IsPrimary == nil {
		o.nullFields = append(o.nullFields, "IsPrimary")
	}
	return o
}

func (o *NetworkInterface) SetEnableIPForwarding(v *bool) *NetworkInterface {
	if o.EnableIPForwarding = v; o.EnableIPForwarding == nil {
		o.nullFields = append(o.nullFields, "EnableIPForwarding")
	}
	return o
}

func (o *NetworkInterface) SetPublicIPSKU(v *string) *NetworkInterface {
	if o.PublicIPSKU = v; o.PublicIPSKU == nil {
		o.nullFields = append(o.nullFields, "PublicIPSKU")
	}
	return o
}

func (o *NetworkInterface) SetSecurityGroup(v *SecurityGroup) *NetworkInterface {
	if o.SecurityGroup = v; o.SecurityGroup == nil {
		o.nullFields = append(o.nullFields, "SecurityGroup")
	}
	return o
}

// endregion

// region AdditionalIPConfig

func (o AdditionalIPConfig) MarshalJSON() ([]byte, error) {
	type noMethod AdditionalIPConfig
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// SetName sets the name
func (o *AdditionalIPConfig) SetName(v *string) *AdditionalIPConfig {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// SetPrivateIPAddressVersion sets the ip address version
func (o *AdditionalIPConfig) SetPrivateIPAddressVersion(v *string) *AdditionalIPConfig {
	if o.PrivateIPAddressVersion = v; o.PrivateIPAddressVersion == nil {
		o.nullFields = append(o.nullFields, "PrivateIPAddressVersion")
	}
	return o
}

// endregion

// region Login

func (o Login) MarshalJSON() ([]byte, error) {
	type noMethod Login
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Login) SetUserName(v *string) *Login {
	if o.UserName = v; o.UserName == nil {
		o.nullFields = append(o.nullFields, "UserName")
	}
	return o
}

func (o *Login) SetSSHPublicKey(v *string) *Login {
	if o.SSHPublicKey = v; o.SSHPublicKey == nil {
		o.nullFields = append(o.nullFields, "SSHPublicKey")
	}
	return o
}

// endregion

// region Extension

func (o Extension) MarshalJSON() ([]byte, error) {
	type noMethod Extension
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Extension) SetAPIVersion(v *string) *Extension {
	if o.APIVersion = v; o.APIVersion == nil {
		o.nullFields = append(o.nullFields, "APIVersion")
	}
	return o
}

func (o *Extension) SetName(v *string) *Extension {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Extension) SetPublisher(v *string) *Extension {
	if o.Publisher = v; o.Publisher == nil {
		o.nullFields = append(o.nullFields, "Publisher")
	}
	return o
}

func (o *Extension) SetType(v *string) *Extension {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Extension) SetMinorVersionAutoUpgrade(v *bool) *Extension {
	if o.MinorVersionAutoUpgrade = v; o.MinorVersionAutoUpgrade == nil {
		o.nullFields = append(o.nullFields, "MinorVersionAutoUpgrade")
	}
	return o
}

func (o *Extension) SetProtectedSettings(v interface{}) *Extension {
	if o.ProtectedSettings = v; o.ProtectedSettings == nil {
		o.nullFields = append(o.nullFields, "ProtectedSettings")
	}
	return o
}

func (o *Extension) SetPublicSettings(v interface{}) *Extension {
	if o.PublicSettings = v; o.PublicSettings == nil {
		o.nullFields = append(o.nullFields, "PublicSettings")
	}
	return o
}

// endregion

// region LoadBalancersConfig

func (o LoadBalancersConfig) MarshalJSON() ([]byte, error) {
	type noMethod LoadBalancersConfig
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LoadBalancersConfig) SetLoadBalancers(v []*LoadBalancer) *LoadBalancersConfig {
	if o.LoadBalancers = v; o.LoadBalancers == nil {
		o.nullFields = append(o.nullFields, "LoadBalancers")
	}
	return o
}

// endregion

// region LoadBalancer

func (o LoadBalancer) MarshalJSON() ([]byte, error) {
	type noMethod LoadBalancer
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *LoadBalancer) SetLoadBalancerSKU(v *string) *LoadBalancer {
	if o.LoadBalancerSKU = v; o.LoadBalancerSKU == nil {
		o.nullFields = append(o.nullFields, "LoadBalancerSKU")
	}
	return o
}

func (o *LoadBalancer) SetName(v *string) *LoadBalancer {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *LoadBalancer) SetResourceGroupName(v *string) *LoadBalancer {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *LoadBalancer) SetType(v *string) *LoadBalancer {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *LoadBalancer) SeBackendPoolNames(v []string) *LoadBalancer {
	if o.BackendPoolNames = v; o.BackendPoolNames == nil {
		o.nullFields = append(o.nullFields, "BackendPoolNames")
	}
	return o
}

// endregion

// region SecurityGroup

func (o SecurityGroup) MarshalJSON() ([]byte, error) {
	type noMethod SecurityGroup
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *SecurityGroup) SetResourceGroupName(v *string) *SecurityGroup {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *SecurityGroup) SetName(v *string) *SecurityGroup {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// endregion

// region ManagedServiceIdentity

func (o ManagedServiceIdentity) MarshalJSON() ([]byte, error) {
	type noMethod ManagedServiceIdentity
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ManagedServiceIdentity) SetResourceGroupName(v *string) *ManagedServiceIdentity {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *ManagedServiceIdentity) SetName(v *string) *ManagedServiceIdentity {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// endregion
