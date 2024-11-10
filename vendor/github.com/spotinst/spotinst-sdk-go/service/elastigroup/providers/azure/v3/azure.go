package v3

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

type Group struct {
	ID                *string     `json:"id,omitempty"`
	Name              *string     `json:"name,omitempty"`
	ResourceGroupName *string     `json:"resourceGroupName,omitempty"`
	Region            *string     `json:"region,omitempty"`
	Capacity          *Capacity   `json:"capacity,omitempty"`
	Compute           *Compute    `json:"compute,omitempty"`
	Strategy          *Strategy   `json:"strategy,omitempty"`
	Scaling           *Scaling    `json:"scaling,omitempty"`
	Description       *string     `json:"description,omitempty"`
	Health            *Health     `json:"health,omitempty"`
	Scheduling        *Scheduling `json:"scheduling,omitempty"`

	// Read-only fields.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`

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

type Strategy struct {
	OnDemandCount       *int                 `json:"onDemandCount,omitempty"`
	DrainingTimeout     *int                 `json:"drainingTimeout,omitempty"`
	SpotPercentage      *int                 `json:"spotPercentage,omitempty"`
	FallbackToOnDemand  *bool                `json:"fallbackToOd,omitempty"`
	AvailabilityVsCost  *int                 `json:"availabilityVsCost,omitempty"`
	CapacityReservation *CapacityReservation `json:"capacityReservation,omitempty"`
	OptimizationWindows []string             `json:"optimizationWindows,omitempty"`
	RevertToSpot        *RevertToSpot        `json:"revertToSpot,omitempty"`
	Signals             []*Signals           `json:"signals,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Capacity struct {
	Minimum *int `json:"minimum,omitempty"`
	Maximum *int `json:"maximum,omitempty"`
	Target  *int `json:"target,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Compute struct {
	VMSizes             *VMSizes             `json:"vmSizes,omitempty"`
	OS                  *string              `json:"os,omitempty"`
	LaunchSpecification *LaunchSpecification `json:"launchSpecification,omitempty"`
	PreferredZones      []string             `json:"preferredZones,omitempty"`
	Zones               []string             `json:"zones,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VMSizes struct {
	OnDemandSizes      []string `json:"odSizes,omitempty"`
	SpotSizes          []string `json:"spotSizes,omitempty"`
	PreferredSpotSizes []string `json:"preferredSpotSizes,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LaunchSpecification struct {
	Image                    *Image                      `json:"image,omitempty"`
	Network                  *Network                    `json:"network,omitempty"`
	Login                    *Login                      `json:"login,omitempty"`
	CustomData               *string                     `json:"customData,omitempty"`
	ManagedServiceIdentities []*ManagedServiceIdentity   `json:"managedServiceIdentities,omitempty"`
	Tags                     []*Tags                     `json:"tags,omitempty"`
	LoadBalancersConfig      *LoadBalancersConfig        `json:"loadBalancersConfig,omitempty"`
	ShutdownScript           *string                     `json:"shutdownScript,omitempty"`
	Extensions               []*Extensions               `json:"extensions,omitempty"`
	BootDiagnostics          *BootDiagnostics            `json:"bootDiagnostics,omitempty"`
	DataDisks                []*DataDisks                `json:"dataDisks,omitempty"`
	OsDisk                   *OsDisk                     `json:"osDisk,omitempty"`
	ProximityPlacementGroups []*ProximityPlacementGroups `json:"proximityPlacementGroups,omitempty"`
	Secrets                  []*Secrets                  `json:"secrets,omitempty"`
	Security                 *Security                   `json:"security,omitempty"`
	UserData                 *string                     `json:"userData,omitempty"`
	VmNamePrefix             *string                     `json:"vmNamePrefix,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LoadBalancersConfig struct {
	LoadBalancers []*LoadBalancer `json:"loadBalancers,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LoadBalancer struct {
	Type              *string  `json:"type,omitempty"`
	ResourceGroupName *string  `json:"resourceGroupName,omitempty"`
	Name              *string  `json:"name,omitempty"`
	SKU               *string  `json:"sku,omitempty"`
	BackendPoolNames  []string `json:"backendPoolNames,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Image struct {
	MarketPlace  *MarketPlaceImage `json:"marketplace,omitempty"`
	Custom       *CustomImage      `json:"custom,omitempty"`
	GalleryImage *GalleryImage     `json:"gallery,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type MarketPlaceImage struct {
	Publisher *string `json:"publisher,omitempty"`
	Offer     *string `json:"offer,omitempty"`
	SKU       *string `json:"sku,omitempty"`
	Version   *string `json:"version,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type CustomImage struct {
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`
	Name              *string `json:"name,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type GalleryImage struct {
	GalleryName       *string `json:"galleryName,omitempty"`
	ImageName         *string `json:"imageName,omitempty"`
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`
	SpotAccountId     *string `json:"spotAccountId,omitempty"`
	Version           *string `json:"version,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Network struct {
	VirtualNetworkName *string             `json:"virtualNetworkName,omitempty"`
	ResourceGroupName  *string             `json:"resourceGroupName,omitempty"`
	NetworkInterfaces  []*NetworkInterface `json:"networkInterfaces,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type NetworkInterface struct {
	SubnetName                *string                     `json:"subnetName,omitempty"`
	AssignPublicIP            *bool                       `json:"assignPublicIp,omitempty"`
	IsPrimary                 *bool                       `json:"isPrimary,omitempty"`
	AdditionalIPConfigs       []*AdditionalIPConfig       `json:"additionalIpConfigurations,omitempty"`
	ApplicationSecurityGroups []*ApplicationSecurityGroup `json:"applicationSecurityGroups,omitempty"`
	EnableIPForwarding        *bool                       `json:"enableIPForwarding,omitempty"`
	PrivateIpAddresses        []string                    `json:"privateIpAddresses,omitempty"`
	PublicIps                 []*PublicIps                `json:"publicIps,omitempty"`
	PublicIpSku               *string                     `json:"publicIpSku,omitempty"`
	SecurityGroup             *SecurityGroup              `json:"securityGroup,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type AdditionalIPConfig struct {
	Name                    *string `json:"name,omitempty"`
	PrivateIPAddressVersion *string `json:"privateIpAddressVersion,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Login struct {
	UserName     *string `json:"userName,omitempty"`
	SSHPublicKey *string `json:"sshPublicKey,omitempty"`
	Password     *string `json:"password,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ApplicationSecurityGroup struct {
	Name              *string `json:"name,omitempty"`
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ManagedServiceIdentity struct {
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`
	Name              *string `json:"name,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Tags struct {
	TagKey   *string `json:"tagKey,omitempty"`
	TagValue *string `json:"tagValue,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Scaling struct {
	Up   []*ScalingPolicy `json:"up,omitempty"`
	Down []*ScalingPolicy `json:"down,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ScalingPolicy struct {
	PolicyName        *string       `json:"policyName,omitempty"`
	MetricName        *string       `json:"metricName,omitempty"`
	Namespace         *string       `json:"namespace,omitempty"`
	Statistic         *string       `json:"statistic,omitempty"`
	Unit              *string       `json:"unit,omitempty"`
	Threshold         *float64      `json:"threshold,omitempty"`
	EvaluationPeriods *int          `json:"evaluationPeriods,omitempty"`
	Period            *int          `json:"period,omitempty"`
	Cooldown          *int          `json:"cooldown,omitempty"`
	Operator          *string       `json:"operator,omitempty"`
	Dimensions        []*Dimensions `json:"dimensions,omitempty"`
	Action            *Action       `json:"action,omitempty"`
	Source            *string       `json:"source,omitempty"`
	IsEnabled         *bool         `json:"isEnabled,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Action struct {
	Type       *string `json:"type,omitempty"`
	Adjustment *string `json:"adjustment,omitempty"`
	Maximum    *string `json:"maximum,omitempty"`
	Minimum    *string `json:"minimum,omitempty"`
	Target     *string `json:"target,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Dimensions struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Extensions struct {
	APIVersion                    *string                        `json:"apiVersion,omitempty"`
	EnableAutomaticUpgrade        *bool                          `json:"enableAutomaticUpgrade,omitempty"`
	MinorVersionAutoUpgrade       *bool                          `json:"minorVersionAutoUpgrade,omitempty"`
	Name                          *string                        `json:"name,omitempty"`
	ProtectedSettingsFromKeyVault *ProtectedSettingsFromKeyVault `json:"protectedSettingsFromKeyVault,omitempty"`
	ProtectedSettings             map[string]interface{}         `json:"protectedSettings,omitempty"`
	PublicSettings                map[string]interface{}         `json:"publicSettings,omitempty"`
	Publisher                     *string                        `json:"publisher,omitempty"`
	Type                          *string                        `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}
type ProtectedSettingsFromKeyVault struct {
	SecretUrl   *string `json:"secretUrl,omitempty"`
	SourceVault *string `json:"sourceVault,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Health struct {
	AutoHealing       *bool    `json:"autoHealing,omitempty"`
	GracePeriod       *int     `json:"gracePeriod,omitempty"`
	HealthCheckTypes  []string `json:"healthCheckTypes,omitempty"`
	UnhealthyDuration *int     `json:"unhealthyDuration,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Scheduling struct {
	Tasks []*Tasks `json:"tasks,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Tasks struct {
	CronExpression       *string `json:"cronExpression,omitempty"`
	IsEnabled            *bool   `json:"isEnabled,omitempty"`
	ScaleMaxCapacity     *int    `json:"scaleMaxCapacity,omitempty"`
	ScaleMinCapacity     *int    `json:"scaleMinCapacity,omitempty"`
	ScaleTargetCapacity  *int    `json:"scaleTargetCapacity,omitempty"`
	Type                 *string `json:"type,omitempty"`
	Adjustment           *int    `json:"adjustment,omitempty"`
	AdjustmentPercentage *int    `json:"adjustmentPercentage,omitempty"`
	BatchSizePercentage  *int    `json:"batchSizePercentage,omitempty"`
	GracePeriod          *int    `json:"gracePeriod,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type DataDisks struct {
	Lun    *int    `json:"lun,omitempty"`
	SizeGB *int    `json:"sizeGB,omitempty"`
	Type   *string `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type OsDisk struct {
	SizeGB *int    `json:"sizeGB,omitempty"`
	Type   *string `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type BootDiagnostics struct {
	IsEnabled  *bool   `json:"isEnabled,omitempty"`
	StorageUri *string `json:"storageUri,omitempty"`
	Type       *string `json:"type,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ProximityPlacementGroups struct {
	Name              *string `json:"name,omitempty"`
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Secrets struct {
	SourceVault       *SourceVault         `json:"sourceVault,omitempty"`
	VaultCertificates []*VaultCertificates `json:"vaultCertificates,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Security struct {
	ConfidentialOsDiskEncryption *bool   `json:"confidentialOsDiskEncryption,omitempty"`
	SecureBootEnabled            *bool   `json:"secureBootEnabled,omitempty"`
	SecurityType                 *string `json:"securityType,omitempty"`
	VTpmEnabled                  *bool   `json:"vTpmEnabled,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type SourceVault struct {
	Name              *string `json:"name,omitempty"`
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type VaultCertificates struct {
	CertificateStore *string `json:"certificateStore,omitempty"`
	CertificateUrl   *string `json:"certificateUrl,omitepmty"`

	forceSendFields []string
	nullFields      []string
}

type CapacityReservation struct {
	CapacityReservationGroups []*CapacityReservationGroups `json:"capacityReservationGroups,omitempty"`
	ShouldUtilize             *bool                        `json:"shouldUtilize,omitempty"`
	UtilizationStrategy       *string                      `json:"utilizationStrategy,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type CapacityReservationGroups struct {
	Name              *string `json:"name,omitempty"`
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`
	ShouldPrioritize  *bool   `json:"shouldPrioritize,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type RevertToSpot struct {
	PerformAt *string `json:"performAt,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Signals struct {
	Type    *string `json:"type,omitempty"`
	Timeout *int    `json:"timeout,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type PublicIps struct {
	Name              *string `json:"name,omitempty"`
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type SecurityGroup struct {
	Name              *string `json:"name,omitempty"`
	ResourceGroupName *string `json:"resourceGroupName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type CreateGroupInput struct {
	Group *Group `json:"group,omitempty"`
}

type CreateGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

type ReadGroupInput struct {
	GroupID *string `json:"groupId,omitempty"`
}

type ReadGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

type UpdateGroupInput struct {
	Group *Group `json:"group,omitempty"`
}

type UpdateGroupOutput struct {
	Group *Group `json:"group,omitempty"`
}

type DeleteGroupInput struct {
	GroupID *string `json:"groupId,omitempty"`
}

type DeleteGroupOutput struct{}

type ListGroupsInput struct{}

type ListGroupsOutput struct {
	Groups []*Group `json:"groups,omitempty"`
}

// region Unmarshallers

func groupFromJSON(in []byte) (*Group, error) {
	b := new(Group)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func groupsFromJSON(in []byte) ([]*Group, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Group, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := groupFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func groupsFromHttpResponse(resp *http.Response) ([]*Group, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return groupsFromJSON(body)
}

// endregion

// region API requests

func (s *ServiceOp) List(ctx context.Context, input *ListGroupsInput) (*ListGroupsOutput, error) {
	r := client.NewRequest(http.MethodGet, "/azure/compute/group")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListGroupsOutput{Groups: gs}, nil
}

func (s *ServiceOp) Create(ctx context.Context, input *CreateGroupInput) (*CreateGroupOutput, error) {
	r := client.NewRequest(http.MethodPost, "/azure/compute/group")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateGroupOutput)
	if len(gs) > 0 {
		output.Group = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) Read(ctx context.Context, input *ReadGroupInput) (*ReadGroupOutput, error) {
	path, err := uritemplates.Expand("/azure/compute/group/{groupId}", uritemplates.Values{
		"groupId": spotinst.StringValue(input.GroupID),
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

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadGroupOutput)
	if len(gs) > 0 {
		output.Group = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) Update(ctx context.Context, input *UpdateGroupInput) (*UpdateGroupOutput, error) {
	path, err := uritemplates.Expand("/azure/compute/group/{groupId}", uritemplates.Values{
		"groupId": spotinst.StringValue(input.Group.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do NOT need the ID anymore, so let's drop it.
	input.Group.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := groupsFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateGroupOutput)
	if len(gs) > 0 {
		output.Group = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) Delete(ctx context.Context, input *DeleteGroupInput) (*DeleteGroupOutput, error) {
	path, err := uritemplates.Expand("/azure/compute/group/{groupId}", uritemplates.Values{
		"groupId": spotinst.StringValue(input.GroupID),
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

	return &DeleteGroupOutput{}, nil
}

// endregion

// region Group

func (o Group) MarshalJSON() ([]byte, error) {
	type noMethod Group
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Group) SetId(v *string) *Group {
	if o.ID = v; o.ID == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Group) SetName(v *string) *Group {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Group) SetResourceGroupName(v *string) *Group {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *Group) SetCapacity(v *Capacity) *Group {
	if o.Capacity = v; o.Capacity == nil {
		o.nullFields = append(o.nullFields, "Capacity")
	}
	return o
}

func (o *Group) SetCompute(v *Compute) *Group {
	if o.Compute = v; o.Compute == nil {
		o.nullFields = append(o.nullFields, "Compute")
	}
	return o
}

func (o *Group) SetStrategy(v *Strategy) *Group {
	if o.Strategy = v; o.Strategy == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *Group) SetRegion(v *string) *Group {
	if o.Region = v; o.Region == nil {
		o.nullFields = append(o.nullFields, "Region")
	}
	return o
}

func (o *Group) SetScaling(v *Scaling) *Group {
	if o.Scaling = v; o.Scaling == nil {
		o.nullFields = append(o.nullFields, "Scaling")
	}
	return o
}

func (o *Group) SetDescription(v *string) *Group {
	if o.Description = v; o.Description == nil {
		o.nullFields = append(o.nullFields, "Description")
	}
	return o
}

func (o *Group) SetHealth(v *Health) *Group {
	if o.Health = v; o.Health == nil {
		o.nullFields = append(o.nullFields, "Health")
	}
	return o
}

func (o *Group) SetScheduling(v *Scheduling) *Group {
	if o.Scheduling = v; o.Scheduling == nil {
		o.nullFields = append(o.nullFields, "Scheduling")
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

func (o *Strategy) SetOnDemandCount(v *int) *Strategy {
	if o.OnDemandCount = v; o.OnDemandCount == nil {
		o.nullFields = append(o.nullFields, "OnDemandCount")
	}
	return o
}

func (o *Strategy) SetDrainingTimeout(v *int) *Strategy {
	if o.DrainingTimeout = v; o.DrainingTimeout == nil {
		o.nullFields = append(o.nullFields, "DrainingTimeout")
	}
	return o
}

func (o *Strategy) SetSpotPercentage(v *int) *Strategy {
	if o.SpotPercentage = v; o.SpotPercentage == nil {
		o.nullFields = append(o.nullFields, "SpotPercentage")
	}
	return o
}

func (o *Strategy) SetFallbackToOnDemand(v *bool) *Strategy {
	if o.FallbackToOnDemand = v; o.FallbackToOnDemand == nil {
		o.nullFields = append(o.nullFields, "FallbackToOnDemand")
	}
	return o
}

func (o *Strategy) SetAvailabilityVsCost(v *int) *Strategy {
	if o.AvailabilityVsCost = v; o.AvailabilityVsCost == nil {
		o.nullFields = append(o.nullFields, "AvailabilityVsCost")
	}
	return o
}

func (o *Strategy) SetCapacityReservation(v *CapacityReservation) *Strategy {
	if o.CapacityReservation = v; o.CapacityReservation == nil {
		o.nullFields = append(o.nullFields, "CapacityReservation")
	}
	return o
}

func (o *Strategy) SetOptimizationWindows(v []string) *Strategy {
	if o.OptimizationWindows = v; o.OptimizationWindows == nil {
		o.nullFields = append(o.nullFields, "OptimizationWindows")
	}
	return o
}

func (o *Strategy) SetRevertToSpot(v *RevertToSpot) *Strategy {
	if o.RevertToSpot = v; o.RevertToSpot == nil {
		o.nullFields = append(o.nullFields, "RevertToSpot")
	}
	return o
}

func (o *Strategy) SetSignals(v []*Signals) *Strategy {
	if o.Signals = v; o.Signals == nil {
		o.nullFields = append(o.nullFields, "Signals")
	}
	return o
}

// endregion

// region Capacity

func (o Capacity) MarshalJSON() ([]byte, error) {
	type noMethod Capacity
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Capacity) SetMinimum(v *int) *Capacity {
	if o.Minimum = v; o.Minimum == nil {
		o.nullFields = append(o.nullFields, "Minimum")
	}
	return o
}

func (o *Capacity) SetMaximum(v *int) *Capacity {
	if o.Maximum = v; o.Maximum == nil {
		o.nullFields = append(o.nullFields, "Maximum")
	}
	return o
}

func (o *Capacity) SetTarget(v *int) *Capacity {
	if o.Target = v; o.Target == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

// endregion

// region Compute

func (o Compute) MarshalJSON() ([]byte, error) {
	type noMethod Compute
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Compute) SetVMSizes(v *VMSizes) *Compute {
	if o.VMSizes = v; o.VMSizes == nil {
		o.nullFields = append(o.nullFields, "VMSizes")
	}
	return o
}

func (o *Compute) SetOS(v *string) *Compute {
	if o.OS = v; o.OS == nil {
		o.nullFields = append(o.nullFields, "OS")
	}
	return o
}

func (o *Compute) SetLaunchSpecification(v *LaunchSpecification) *Compute {
	if o.LaunchSpecification = v; o.LaunchSpecification == nil {
		o.nullFields = append(o.nullFields, "LaunchSpecification")
	}
	return o
}

func (o *Compute) SetPreferredZones(v []string) *Compute {
	if o.PreferredZones = v; o.PreferredZones == nil {
		o.nullFields = append(o.nullFields, "PreferredZones")
	}
	return o
}

func (o *Compute) SetZones(v []string) *Compute {
	if o.Zones = v; o.Zones == nil {
		o.nullFields = append(o.nullFields, "Zones")
	}
	return o
}

// endregion

// region VMSize

func (o VMSizes) MarshalJSON() ([]byte, error) {
	type noMethod VMSizes
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VMSizes) SetOnDemandSizes(v []string) *VMSizes {
	if o.OnDemandSizes = v; o.OnDemandSizes == nil {
		o.nullFields = append(o.nullFields, "OnDemandSizes")
	}
	return o
}

func (o *VMSizes) SetSpotSizes(v []string) *VMSizes {
	if o.SpotSizes = v; o.SpotSizes == nil {
		o.nullFields = append(o.nullFields, "SpotSizes")
	}
	return o
}

func (o *VMSizes) SetPreferredSpotSizes(v []string) *VMSizes {
	if o.PreferredSpotSizes = v; o.PreferredSpotSizes == nil {
		o.nullFields = append(o.nullFields, "PreferredSpotSizes")
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

func (o *LaunchSpecification) SetCustomData(v *string) *LaunchSpecification {
	if o.CustomData = v; o.CustomData == nil {
		o.nullFields = append(o.nullFields, "CustomData")
	}
	return o
}

func (o *LaunchSpecification) SetManagedServiceIdentities(v []*ManagedServiceIdentity) *LaunchSpecification {
	if o.ManagedServiceIdentities = v; o.ManagedServiceIdentities == nil {
		o.nullFields = append(o.nullFields, "ManagedServiceIdentities")
	}
	return o
}

func (o *LaunchSpecification) SetLoadBalancersConfig(v *LoadBalancersConfig) *LaunchSpecification {
	if o.LoadBalancersConfig = v; o.LoadBalancersConfig == nil {
		o.nullFields = append(o.nullFields, "LoadBalancersConfig")
	}
	return o
}

// SetShutdownScript sets the shutdown script used when draining instances
func (o *LaunchSpecification) SetShutdownScript(v *string) *LaunchSpecification {
	if o.ShutdownScript = v; o.ShutdownScript == nil {
		o.nullFields = append(o.nullFields, "ShutdownScript")
	}
	return o
}

func (o *LaunchSpecification) SetTags(v []*Tags) *LaunchSpecification {
	if o.Tags = v; o.Tags == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

func (o *LaunchSpecification) SetExtensions(v []*Extensions) *LaunchSpecification {
	if o.Extensions = v; o.Extensions == nil {
		o.nullFields = append(o.nullFields, "Extensions")
	}
	return o
}

func (o *LaunchSpecification) SetBootDiagnostics(v *BootDiagnostics) *LaunchSpecification {
	if o.BootDiagnostics = v; o.BootDiagnostics == nil {
		o.nullFields = append(o.nullFields, "BootDiagnostics")
	}
	return o
}

func (o *LaunchSpecification) SetDataDisks(v []*DataDisks) *LaunchSpecification {
	if o.DataDisks = v; o.DataDisks == nil {
		o.nullFields = append(o.nullFields, "DataDisks")
	}
	return o
}

func (o *LaunchSpecification) SetOsDisk(v *OsDisk) *LaunchSpecification {
	if o.OsDisk = v; o.OsDisk == nil {
		o.nullFields = append(o.nullFields, "OsDisk")
	}
	return o
}

func (o *LaunchSpecification) SetProximityPlacementGroups(v []*ProximityPlacementGroups) *LaunchSpecification {
	if o.ProximityPlacementGroups = v; o.ProximityPlacementGroups == nil {
		o.nullFields = append(o.nullFields, "ProximityPlacementGroups")
	}
	return o
}

func (o *LaunchSpecification) SetSecrets(v []*Secrets) *LaunchSpecification {
	if o.Secrets = v; o.Secrets == nil {
		o.nullFields = append(o.nullFields, "Secrets")
	}
	return o
}

func (o *LaunchSpecification) SetSecurity(v *Security) *LaunchSpecification {
	if o.Security = v; o.Security == nil {
		o.nullFields = append(o.nullFields, "Security")
	}
	return o
}

func (o *LaunchSpecification) SetUserData(v *string) *LaunchSpecification {
	if o.UserData = v; o.UserData == nil {
		o.nullFields = append(o.nullFields, "UserData")
	}
	return o
}

func (o *LaunchSpecification) SetVmNamePrefix(v *string) *LaunchSpecification {
	if o.VmNamePrefix = v; o.VmNamePrefix == nil {
		o.nullFields = append(o.nullFields, "VmNamePrefix")
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

func (o *Image) SetMarketPlaceImage(v *MarketPlaceImage) *Image {
	if o.MarketPlace = v; o.MarketPlace == nil {
		o.nullFields = append(o.nullFields, "MarketPlace")
	}
	return o
}

func (o *Image) SetCustom(v *CustomImage) *Image {
	if o.Custom = v; o.Custom == nil {
		o.nullFields = append(o.nullFields, "Custom")
	}
	return o
}

func (o *Image) SetGalleryImage(v *GalleryImage) *Image {
	if o.GalleryImage = v; o.GalleryImage == nil {
		o.nullFields = append(o.nullFields, "GalleryImage")
	}
	return o
}

// endregion

// region MarketPlaceImage

func (o MarketPlaceImage) MarshalJSON() ([]byte, error) {
	type noMethod MarketPlaceImage
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *MarketPlaceImage) SetPublisher(v *string) *MarketPlaceImage {
	if o.Publisher = v; o.Publisher == nil {
		o.nullFields = append(o.nullFields, "Publisher")
	}
	return o
}

func (o *MarketPlaceImage) SetOffer(v *string) *MarketPlaceImage {
	if o.Offer = v; o.Offer == nil {
		o.nullFields = append(o.nullFields, "Offer")
	}
	return o
}

func (o *MarketPlaceImage) SetSKU(v *string) *MarketPlaceImage {
	if o.SKU = v; o.SKU == nil {
		o.nullFields = append(o.nullFields, "SKU")
	}
	return o
}

func (o *MarketPlaceImage) SetVersion(v *string) *MarketPlaceImage {
	if o.Version = v; o.Version == nil {
		o.nullFields = append(o.nullFields, "Version")
	}
	return o
}

// endregion

// region Tags

func (o Tags) MarshalJSON() ([]byte, error) {
	type noMethod Tags
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Tags) SetTagKey(v *string) *Tags {
	if o.TagKey = v; o.TagKey == nil {
		o.nullFields = append(o.nullFields, "TagKey")
	}
	return o
}

func (o *Tags) SetTagValue(v *string) *Tags {
	if o.TagValue = v; o.TagValue == nil {
		o.nullFields = append(o.nullFields, "TagValue")
	}
	return o
}

// endregion

// region CustomImage

func (o CustomImage) MarshalJSON() ([]byte, error) {
	type noMethod CustomImage
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *CustomImage) SetResourceGroupName(v *string) *CustomImage {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *CustomImage) SetName(v *string) *CustomImage {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

// endregion

func (o GalleryImage) MarshalJSON() ([]byte, error) {
	type noMethod GalleryImage
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *GalleryImage) SetGalleryName(v *string) *GalleryImage {
	if o.GalleryName = v; o.GalleryName == nil {
		o.nullFields = append(o.nullFields, "GalleryName")
	}
	return o
}

func (o *GalleryImage) SetImageName(v *string) *GalleryImage {
	if o.ImageName = v; o.ImageName == nil {
		o.nullFields = append(o.nullFields, "ImageName")
	}
	return o
}

func (o *GalleryImage) SetResourceGroupName(v *string) *GalleryImage {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *GalleryImage) SetSpotAccountId(v *string) *GalleryImage {
	if o.SpotAccountId = v; o.SpotAccountId == nil {
		o.nullFields = append(o.nullFields, "SpotAccountId")
	}
	return o
}

func (o *GalleryImage) SetVersion(v *string) *GalleryImage {
	if o.Version = v; o.Version == nil {
		o.nullFields = append(o.nullFields, "Version")
	}
	return o
}

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

func (o *NetworkInterface) SetApplicationSecurityGroups(v []*ApplicationSecurityGroup) *NetworkInterface {
	if o.ApplicationSecurityGroups = v; o.ApplicationSecurityGroups == nil {
		o.nullFields = append(o.nullFields, "ApplicationSecurityGroups")
	}
	return o
}

func (o *NetworkInterface) SetEnableIPForwarding(v *bool) *NetworkInterface {
	if o.EnableIPForwarding = v; o.EnableIPForwarding == nil {
		o.nullFields = append(o.nullFields, "EnableIPForwarding")
	}
	return o
}

func (o *NetworkInterface) SetPrivateIpAddresses(v []string) *NetworkInterface {
	if o.PrivateIpAddresses = v; o.PrivateIpAddresses == nil {
		o.nullFields = append(o.nullFields, "PrivateIpAddresses")
	}
	return o
}

func (o *NetworkInterface) SetPublicIps(v []*PublicIps) *NetworkInterface {
	if o.PublicIps = v; o.PublicIps == nil {
		o.nullFields = append(o.nullFields, "PublicIps")
	}
	return o
}

func (o *NetworkInterface) SetPublicIpSku(v *string) *NetworkInterface {
	if o.PublicIpSku = v; o.PublicIpSku == nil {
		o.nullFields = append(o.nullFields, "PublicIpSku")
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

func (o *AdditionalIPConfig) SetName(v *string) *AdditionalIPConfig {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

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

func (o *Login) SetPassword(v *string) *Login {
	if o.Password = v; o.Password == nil {
		o.nullFields = append(o.nullFields, "Password")
	}
	return o
}

// endregion

// region ApplicationSecurityGroup

func (o ApplicationSecurityGroup) MarshalJSON() ([]byte, error) {
	type noMethod ApplicationSecurityGroup
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ApplicationSecurityGroup) SetName(v *string) *ApplicationSecurityGroup {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *ApplicationSecurityGroup) SetResourceGroupName(v *string) *ApplicationSecurityGroup {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
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

func (o *LoadBalancer) SetType(v *string) *LoadBalancer {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *LoadBalancer) SetResourceGroupName(v *string) *LoadBalancer {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *LoadBalancer) SetName(v *string) *LoadBalancer {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *LoadBalancer) SetSKU(v *string) *LoadBalancer {
	if o.SKU = v; o.SKU == nil {
		o.nullFields = append(o.nullFields, "SKU")
	}
	return o
}

func (o *LoadBalancer) SetBackendPoolNames(v []string) *LoadBalancer {
	if o.BackendPoolNames = v; o.BackendPoolNames == nil {
		o.nullFields = append(o.nullFields, "BackendPoolNames")
	}
	return o
}

// endregion

func (o Scaling) MarshalJSON() ([]byte, error) {
	type noMethod Scaling
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scaling) SetUp(v []*ScalingPolicy) *Scaling {
	if o.Up = v; o.Up == nil {
		o.nullFields = append(o.nullFields, "Up")
	}
	return o
}

func (o *Scaling) SetDown(v []*ScalingPolicy) *Scaling {
	if o.Down = v; o.Down == nil {
		o.nullFields = append(o.nullFields, "Down")
	}
	return o
}

func (o ScalingPolicy) MarshalJSON() ([]byte, error) {
	type noMethod ScalingPolicy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ScalingPolicy) SetPolicyName(v *string) *ScalingPolicy {
	if o.PolicyName = v; o.PolicyName == nil {
		o.nullFields = append(o.nullFields, "PolicyName")
	}
	return o
}

func (o *ScalingPolicy) SetMetricName(v *string) *ScalingPolicy {
	if o.MetricName = v; o.MetricName == nil {
		o.nullFields = append(o.nullFields, "MetricName")
	}
	return o
}

func (o *ScalingPolicy) SetNamespace(v *string) *ScalingPolicy {
	if o.Namespace = v; o.Namespace == nil {
		o.nullFields = append(o.nullFields, "Namespace")
	}
	return o
}

func (o *ScalingPolicy) SetStatistic(v *string) *ScalingPolicy {
	if o.Statistic = v; o.Statistic == nil {
		o.nullFields = append(o.nullFields, "Statistic")
	}
	return o
}

func (o *ScalingPolicy) SetUnit(v *string) *ScalingPolicy {
	if o.Unit = v; o.Unit == nil {
		o.nullFields = append(o.nullFields, "Unit")
	}
	return o
}

func (o *ScalingPolicy) SetThreshold(v *float64) *ScalingPolicy {
	if o.Threshold = v; o.Threshold == nil {
		o.nullFields = append(o.nullFields, "Threshold")
	}
	return o
}

func (o *ScalingPolicy) SetEvaluationPeriods(v *int) *ScalingPolicy {
	if o.EvaluationPeriods = v; o.EvaluationPeriods == nil {
		o.nullFields = append(o.nullFields, "EvaluationPeriods")
	}
	return o
}

func (o *ScalingPolicy) SetPeriod(v *int) *ScalingPolicy {
	if o.Period = v; o.Period == nil {
		o.nullFields = append(o.nullFields, "Period")
	}
	return o
}

func (o *ScalingPolicy) SetCooldown(v *int) *ScalingPolicy {
	if o.Cooldown = v; o.Cooldown == nil {
		o.nullFields = append(o.nullFields, "Cooldown")
	}
	return o
}

func (o *ScalingPolicy) SetOperator(v *string) *ScalingPolicy {
	if o.Operator = v; o.Operator == nil {
		o.nullFields = append(o.nullFields, "Operator")
	}
	return o
}

func (o *ScalingPolicy) SetDimensions(v []*Dimensions) *ScalingPolicy {
	if o.Dimensions = v; o.Dimensions == nil {
		o.nullFields = append(o.nullFields, "Dimensions")
	}
	return o
}

func (o *ScalingPolicy) SetAction(v *Action) *ScalingPolicy {
	if o.Action = v; o.Action == nil {
		o.nullFields = append(o.nullFields, "Action")
	}
	return o
}

func (o *ScalingPolicy) SetIsEnabled(v *bool) *ScalingPolicy {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *ScalingPolicy) SetSource(v *string) *ScalingPolicy {
	if o.Source = v; o.Source == nil {
		o.nullFields = append(o.nullFields, "Source")
	}
	return o
}

func (o Action) MarshalJSON() ([]byte, error) {
	type noMethod Action
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Action) SetType(v *string) *Action {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Action) SetAdjustment(v *string) *Action {
	if o.Adjustment = v; o.Adjustment == nil {
		o.nullFields = append(o.nullFields, "Adjustment")
	}
	return o
}

func (o *Action) SetMaximum(v *string) *Action {
	if o.Maximum = v; o.Maximum == nil {
		o.nullFields = append(o.nullFields, "Maximum")
	}
	return o
}

func (o *Action) SetMinimum(v *string) *Action {
	if o.Minimum = v; o.Minimum == nil {
		o.nullFields = append(o.nullFields, "Minimum")
	}
	return o
}

func (o *Action) SetTarget(v *string) *Action {
	if o.Target = v; o.Target == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

// endregion

// region Dimension

func (o Dimensions) MarshalJSON() ([]byte, error) {
	type noMethod Dimensions
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Dimensions) SetName(v *string) *Dimensions {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Dimensions) SetValue(v *string) *Dimensions {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

// endregion

// region Extension

func (o Extensions) MarshalJSON() ([]byte, error) {
	type noMethod Extensions
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Extensions) SetName(v *string) *Extensions {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Extensions) SetType(v *string) *Extensions {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Extensions) SetPublisher(v *string) *Extensions {
	if o.Publisher = v; o.Publisher == nil {
		o.nullFields = append(o.nullFields, "Publisher")
	}
	return o
}

func (o *Extensions) SetAPIVersion(v *string) *Extensions {
	if o.APIVersion = v; o.APIVersion == nil {
		o.nullFields = append(o.nullFields, "APIVersion")
	}
	return o
}

func (o *Extensions) SetMinorVersionAutoUpgrade(v *bool) *Extensions {
	if o.MinorVersionAutoUpgrade = v; o.MinorVersionAutoUpgrade == nil {
		o.nullFields = append(o.nullFields, "MinorVersionAutoUpgrade")
	}
	return o
}

func (o *Extensions) SetEnableAutomaticUpgrade(v *bool) *Extensions {
	if o.EnableAutomaticUpgrade = v; o.EnableAutomaticUpgrade == nil {
		o.nullFields = append(o.nullFields, "EnableAutomaticUpgrade")
	}
	return o
}

func (o *Extensions) SetProtectedSettings(v map[string]interface{}) *Extensions {
	if o.ProtectedSettings = v; o.ProtectedSettings == nil {
		o.nullFields = append(o.nullFields, "ProtectedSettings")
	}
	return o
}

func (o *Extensions) SetProtectedSettingsFromKeyVault(v *ProtectedSettingsFromKeyVault) *Extensions {
	if o.ProtectedSettingsFromKeyVault = v; o.ProtectedSettingsFromKeyVault == nil {
		o.nullFields = append(o.nullFields, "ProtectedSettingsFromKeyVault")
	}
	return o
}

func (o *Extensions) SetPublicSettings(v map[string]interface{}) *Extensions {
	if o.PublicSettings = v; o.PublicSettings == nil {
		o.nullFields = append(o.nullFields, "PublicSettings")
	}
	return o
}

// endregion

func (o ProtectedSettingsFromKeyVault) MarshalJSON() ([]byte, error) {
	type noMethod ProtectedSettingsFromKeyVault
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ProtectedSettingsFromKeyVault) SetSecretUrl(v *string) *ProtectedSettingsFromKeyVault {
	if o.SecretUrl = v; o.SecretUrl == nil {
		o.nullFields = append(o.nullFields, "SecretUrl")
	}
	return o
}

func (o *ProtectedSettingsFromKeyVault) SetSourceVault(v *string) *ProtectedSettingsFromKeyVault {
	if o.SourceVault = v; o.SourceVault == nil {
		o.nullFields = append(o.nullFields, "SourceVault")
	}
	return o
}

// region Signal

func (o Signals) MarshalJSON() ([]byte, error) {
	type noMethod Signals
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Signals) SetType(v *string) *Signals {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Signals) SetTimeout(v *int) *Signals {
	if o.Timeout = v; o.Timeout == nil {
		o.nullFields = append(o.nullFields, "Timeout")
	}
	return o
}

// endregion

// region RevertToSpot

func (o RevertToSpot) MarshalJSON() ([]byte, error) {
	type noMethod RevertToSpot
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *RevertToSpot) SetPerformAt(v *string) *RevertToSpot {
	if o.PerformAt = v; o.PerformAt == nil {
		o.nullFields = append(o.nullFields, "PerformAt")
	}
	return o
}

// endregion

// region CapacityReservation

func (o CapacityReservation) MarshalJSON() ([]byte, error) {
	type noMethod CapacityReservation
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *CapacityReservation) SetShouldUtilize(v *bool) *CapacityReservation {
	if o.ShouldUtilize = v; o.ShouldUtilize == nil {
		o.nullFields = append(o.nullFields, "ShouldUtilize")
	}
	return o
}

func (o *CapacityReservation) SetUtilizationStrategy(v *string) *CapacityReservation {
	if o.UtilizationStrategy = v; o.UtilizationStrategy == nil {
		o.nullFields = append(o.nullFields, "UtilizationStrategy")
	}
	return o
}

func (o *CapacityReservation) SetCapacityReservationGroups(v []*CapacityReservationGroups) *CapacityReservation {
	if o.CapacityReservationGroups = v; o.CapacityReservationGroups == nil {
		o.nullFields = append(o.nullFields, "CapacityReservationGroups")
	}
	return o
}

// endregion

// region CapacityReservationGroup

func (o CapacityReservationGroups) MarshalJSON() ([]byte, error) {
	type noMethod CapacityReservationGroups
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *CapacityReservationGroups) SetName(v *string) *CapacityReservationGroups {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *CapacityReservationGroups) SetResourceGroupName(v *string) *CapacityReservationGroups {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o *CapacityReservationGroups) SetShouldPrioritize(v *bool) *CapacityReservationGroups {
	if o.ShouldPrioritize = v; o.ShouldPrioritize == nil {
		o.nullFields = append(o.nullFields, "ShouldPrioritize")
	}
	return o
}

// endregion

func (o Health) MarshalJSON() ([]byte, error) {
	type noMethod Health
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Health) SetAutoHealing(v *bool) *Health {
	if o.AutoHealing = v; o.AutoHealing == nil {
		o.nullFields = append(o.nullFields, "AutoHealing")
	}
	return o
}

func (o *Health) SetGracePeriod(v *int) *Health {
	if o.GracePeriod = v; o.GracePeriod == nil {
		o.nullFields = append(o.nullFields, "GracePeriod")
	}
	return o
}

func (o *Health) SetHealthCheckTypes(v []string) *Health {
	if o.HealthCheckTypes = v; o.HealthCheckTypes == nil {
		o.nullFields = append(o.nullFields, "HealthCheckTypes")
	}
	return o
}

func (o *Health) SetUnhealthyDuration(v *int) *Health {
	if o.UnhealthyDuration = v; o.UnhealthyDuration == nil {
		o.nullFields = append(o.nullFields, "UnhealthyDuration")
	}
	return o
}

func (o Scheduling) MarshalJSON() ([]byte, error) {
	type noMethod Scheduling
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scheduling) SetTasks(v []*Tasks) *Scheduling {
	if o.Tasks = v; o.Tasks == nil {
		o.nullFields = append(o.nullFields, "Tasks")
	}
	return o
}

func (o Tasks) MarshalJSON() ([]byte, error) {
	type noMethod Tasks
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Tasks) SetCronExpression(v *string) *Tasks {
	if o.CronExpression = v; o.CronExpression == nil {
		o.nullFields = append(o.nullFields, "CronExpression")
	}
	return o
}

func (o *Tasks) SetIsEnabled(v *bool) *Tasks {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *Tasks) SetScaleMaxCapacity(v *int) *Tasks {
	if o.ScaleMaxCapacity = v; o.ScaleMaxCapacity == nil {
		o.nullFields = append(o.nullFields, "ScaleMaxCapacity")
	}
	return o
}

func (o *Tasks) SetScaleMinCapacity(v *int) *Tasks {
	if o.ScaleMinCapacity = v; o.ScaleMinCapacity == nil {
		o.nullFields = append(o.nullFields, "ScaleMinCapacity")
	}
	return o
}

func (o *Tasks) SetScaleTargetCapacity(v *int) *Tasks {
	if o.ScaleTargetCapacity = v; o.ScaleTargetCapacity == nil {
		o.nullFields = append(o.nullFields, "ScaleTargetCapacity")
	}
	return o
}

func (o *Tasks) SetType(v *string) *Tasks {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Tasks) SetAdjustment(v *int) *Tasks {
	if o.Adjustment = v; o.Adjustment == nil {
		o.nullFields = append(o.nullFields, "Adjustment")
	}
	return o
}

func (o *Tasks) SetAdjustmentPercentage(v *int) *Tasks {
	if o.AdjustmentPercentage = v; o.AdjustmentPercentage == nil {
		o.nullFields = append(o.nullFields, "AdjustmentPercentage")
	}
	return o
}

func (o *Tasks) SetBatchSizePercentage(v *int) *Tasks {
	if o.BatchSizePercentage = v; o.BatchSizePercentage == nil {
		o.nullFields = append(o.nullFields, "BatchSizePercentage")
	}
	return o
}

func (o *Tasks) SetGracePeriod(v *int) *Tasks {
	if o.GracePeriod = v; o.GracePeriod == nil {
		o.nullFields = append(o.nullFields, "GracePeriod")
	}
	return o
}

func (o DataDisks) MarshalJSON() ([]byte, error) {
	type noMethod DataDisks
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *DataDisks) SetLun(v *int) *DataDisks {
	if o.Lun = v; o.Lun == nil {
		o.nullFields = append(o.nullFields, "Lun")
	}
	return o
}

func (o *DataDisks) SetSizeGB(v *int) *DataDisks {
	if o.SizeGB = v; o.SizeGB == nil {
		o.nullFields = append(o.nullFields, "SizeGB")
	}
	return o
}

func (o *DataDisks) SetType(v *string) *DataDisks {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o OsDisk) MarshalJSON() ([]byte, error) {
	type noMethod OsDisk
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *OsDisk) SetSizeGB(v *int) *OsDisk {
	if o.SizeGB = v; o.SizeGB == nil {
		o.nullFields = append(o.nullFields, "SizeGB")
	}
	return o
}

func (o *OsDisk) SetType(v *string) *OsDisk {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o BootDiagnostics) MarshalJSON() ([]byte, error) {
	type noMethod BootDiagnostics
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *BootDiagnostics) SetStorageUri(v *string) *BootDiagnostics {
	if o.StorageUri = v; o.StorageUri == nil {
		o.nullFields = append(o.nullFields, "StorageUri")
	}
	return o
}

func (o *BootDiagnostics) SetIsEnabled(v *bool) *BootDiagnostics {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *BootDiagnostics) SetType(v *string) *BootDiagnostics {
	if o.Type = v; o.Type == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o ProximityPlacementGroups) MarshalJSON() ([]byte, error) {
	type noMethod ProximityPlacementGroups
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ProximityPlacementGroups) SetName(v *string) *ProximityPlacementGroups {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *ProximityPlacementGroups) SetResourceGroupName(v *string) *ProximityPlacementGroups {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o Secrets) MarshalJSON() ([]byte, error) {
	type noMethod Secrets
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Secrets) SetSourceVault(v *SourceVault) *Secrets {
	if o.SourceVault = v; o.SourceVault == nil {
		o.nullFields = append(o.nullFields, "SourceVault")
	}
	return o
}

func (o *Secrets) SetVaultCertificates(v []*VaultCertificates) *Secrets {
	if o.VaultCertificates = v; o.VaultCertificates == nil {
		o.nullFields = append(o.nullFields, "VaultCertificates")
	}
	return o
}

func (o Security) MarshalJSON() ([]byte, error) {
	type noMethod Security
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Security) SetConfidentialOsDiskEncryption(v *bool) *Security {
	if o.ConfidentialOsDiskEncryption = v; o.ConfidentialOsDiskEncryption == nil {
		o.nullFields = append(o.nullFields, "ConfidentialOsDiskEncryption")
	}
	return o
}

func (o *Security) SetSecureBootEnabled(v *bool) *Security {
	if o.SecureBootEnabled = v; o.SecureBootEnabled == nil {
		o.nullFields = append(o.nullFields, "SecureBootEnabled")
	}
	return o
}

func (o *Security) SetSecurityType(v *string) *Security {
	if o.SecurityType = v; o.SecurityType == nil {
		o.nullFields = append(o.nullFields, "SecurityType")
	}
	return o
}

func (o *Security) SetVTpmEnabled(v *bool) *Security {
	if o.VTpmEnabled = v; o.VTpmEnabled == nil {
		o.nullFields = append(o.nullFields, "VTpmEnabled")
	}
	return o
}

func (o SourceVault) MarshalJSON() ([]byte, error) {
	type noMethod SourceVault
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *SourceVault) SetName(v *string) *SourceVault {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *SourceVault) SetResourceGroupName(v *string) *SourceVault {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o VaultCertificates) MarshalJSON() ([]byte, error) {
	type noMethod VaultCertificates
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VaultCertificates) SetCertificateStore(v *string) *VaultCertificates {
	if o.CertificateStore = v; o.CertificateStore == nil {
		o.nullFields = append(o.nullFields, "CertificateStore")
	}
	return o
}

func (o *VaultCertificates) SetCertificateUrl(v *string) *VaultCertificates {
	if o.CertificateUrl = v; o.CertificateUrl == nil {
		o.nullFields = append(o.nullFields, "CertificateUrl")
	}
	return o
}

func (o PublicIps) MarshalJSON() ([]byte, error) {
	type noMethod PublicIps
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *PublicIps) SetName(v *string) *PublicIps {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *PublicIps) SetResourceGroupName(v *string) *PublicIps {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}

func (o SecurityGroup) MarshalJSON() ([]byte, error) {
	type noMethod SecurityGroup
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *SecurityGroup) SetName(v *string) *SecurityGroup {
	if o.Name = v; o.Name == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *SecurityGroup) SetResourceGroupName(v *string) *SecurityGroup {
	if o.ResourceGroupName = v; o.ResourceGroupName == nil {
		o.nullFields = append(o.nullFields, "ResourceGroupName")
	}
	return o
}
