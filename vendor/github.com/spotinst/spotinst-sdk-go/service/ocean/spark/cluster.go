package spark

import (
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
)

type Cluster struct {
	Config *Config `json:"config,omitempty"`

	// Read-only fields.
	ID                    *string    `json:"id,omitempty"`
	ControllerClusterID   *string    `json:"controllerClusterId,omitempty"`
	OceanClusterID        *string    `json:"oceanClusterId,omitempty"`
	Region                *string    `json:"region,omitempty"`
	State                 *string    `json:"state,omitempty"`
	K8sVersion            *string    `json:"k8sVersion,omitempty"`
	OperatorVersion       *string    `json:"operatorVersion,omitempty"`
	OperatorLastHeartbeat *time.Time `json:"operatorLastHeartbeat,omitempty"`
	CreatedAt             *time.Time `json:"createdAt,omitempty"`
	UpdatedAt             *time.Time `json:"updatedAt,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Config struct {
	Ingress       *IngressConfig       `json:"ingress,omitempty"`
	Webhook       *WebhookConfig       `json:"webhook,omitempty"`
	Compute       *ComputeConfig       `json:"compute,omitempty"`
	LogCollection *LogCollectionConfig `json:"logCollection,omitempty"`
	Spark         *SparkConfig         `json:"spark,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type LogCollectionConfig struct {
	// Deprecated: Use CollectAppLogs instead.
	CollectDriverLogs *bool `json:"collectDriverLogs,omitempty"`
	CollectAppLogs    *bool `json:"collectAppLogs,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ComputeConfig struct {
	UseTaints  *bool `json:"useTaints,omitempty"`
	CreateVngs *bool `json:"createVngs,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type WebhookConfig struct {
	UseHostNetwork   *bool  `json:"useHostNetwork,omitempty"`
	HostNetworkPorts []*int `json:"hostNetworkPorts,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type SparkConfig struct {
	AppNamespaces []*string `json:"appNamespaces,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type IngressConfig struct {
	// Deprecated: Use LoadBalancer.ServiceAnnotations instead.
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	// Deprecated: Has no effect.
	DeployIngress *bool `json:"deployIngress,omitempty"`

	Controller     *IngressConfigController     `json:"controller,omitempty"`
	CustomEndpoint *IngressConfigCustomEndpoint `json:"customEndpoint,omitempty"`
	LoadBalancer   *IngressConfigLoadBalancer   `json:"loadBalancer,omitempty"`
	PrivateLink    *IngressConfigPrivateLink    `json:"privateLink,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type IngressConfigController struct {
	Managed *bool `json:"managed,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type IngressConfigCustomEndpoint struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Address *string `json:"address,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type IngressConfigLoadBalancer struct {
	Managed            *bool             `json:"managed,omitempty"`
	TargetGroupARN     *string           `json:"targetGroupArn,omitempty"`
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type IngressConfigPrivateLink struct {
	Enabled            *bool   `json:"enabled,omitempty"`
	VPCEndpointService *string `json:"vpcEndpointService,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListClustersInput struct {
	ControllerClusterID *string `json:"controllerClusterId,omitempty"`
	ClusterState        *string `json:"clusterState,omitempty"`
}

type ListClustersOutput struct {
	Clusters []*Cluster `json:"clusters,omitempty"`
}

type ReadClusterInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
}

type ReadClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type CreateClusterInput struct {
	Cluster *CreateClusterRequest `json:"cluster,omitempty"`
}

type CreateClusterRequest struct {
	OceanClusterID *string `json:"oceanClusterId,omitempty"`
	Config         *Config `json:"config,omitempty"`
}

type CreateClusterOutput struct {
	Cluster *Cluster `json:"cluster,omitempty"`
}

type UpdateClusterInput struct {
	ClusterID *string               `json:"-"`
	Cluster   *UpdateClusterRequest `json:"cluster,omitempty"`
}

type UpdateClusterRequest struct {
	Config *Config `json:"config,omitempty"`
}

type UpdateClusterOutput struct{}

type DeleteClusterInput struct {
	ForceDelete *bool   `json:"-"`
	ClusterID   *string `json:"clusterId,omitempty"`
}

type DeleteClusterOutput struct{}

// region Cluster

func (c Cluster) MarshalJSON() ([]byte, error) {
	type noMethod Cluster
	raw := noMethod(c)
	return jsonutil.MarshalJSON(raw, c.forceSendFields, c.nullFields)
}

func (c *Cluster) SetConfig(v *Config) *Cluster {
	if c.Config = v; c.Config == nil {
		c.nullFields = append(c.nullFields, "Config")
	}
	return c
}

// endregion

// region Config

func (c Config) MarshalJSON() ([]byte, error) {
	type noMethod Config
	raw := noMethod(c)
	return jsonutil.MarshalJSON(raw, c.forceSendFields, c.nullFields)
}

func (c *Config) SetIngress(v *IngressConfig) *Config {
	if c.Ingress = v; c.Ingress == nil {
		c.nullFields = append(c.nullFields, "Ingress")
	}
	return c
}

func (c *Config) SetWebhook(v *WebhookConfig) *Config {
	if c.Webhook = v; c.Webhook == nil {
		c.nullFields = append(c.nullFields, "Webhook")
	}
	return c
}

func (c *Config) SetCompute(v *ComputeConfig) *Config {
	if c.Compute = v; c.Compute == nil {
		c.nullFields = append(c.nullFields, "Compute")
	}
	return c
}

func (c *Config) SetLogCollection(v *LogCollectionConfig) *Config {
	if c.LogCollection = v; c.LogCollection == nil {
		c.nullFields = append(c.nullFields, "LogCollection")
	}
	return c
}

func (c *Config) SetSpark(v *SparkConfig) *Config {
	if c.Spark = v; c.Spark == nil {
		c.nullFields = append(c.nullFields, "Spark")
	}
	return c
}

// endregion

// region Ingress

func (i IngressConfig) MarshalJSON() ([]byte, error) {
	type noMethod IngressConfig
	raw := noMethod(i)
	return jsonutil.MarshalJSON(raw, i.forceSendFields, i.nullFields)
}

func (i *IngressConfig) SetServiceAnnotations(v map[string]string) *IngressConfig {
	if i.ServiceAnnotations = v; i.ServiceAnnotations == nil {
		i.nullFields = append(i.nullFields, "ServiceAnnotations")
	}
	return i
}

func (i *IngressConfig) SetDeployIngress(v *bool) *IngressConfig {
	if i.DeployIngress = v; i.DeployIngress == nil {
		i.nullFields = append(i.nullFields, "DeployIngress")
	}
	return i
}

func (i *IngressConfig) SetController(c *IngressConfigController) *IngressConfig {
	if i.Controller = c; i.Controller == nil {
		i.nullFields = append(i.nullFields, "Controller")
	}
	return i
}

func (i *IngressConfig) SetLoadBalancer(lb *IngressConfigLoadBalancer) *IngressConfig {
	if i.LoadBalancer = lb; i.LoadBalancer == nil {
		i.nullFields = append(i.nullFields, "LoadBalancer")
	}
	return i
}

func (i *IngressConfig) SetCustomEndpoint(ce *IngressConfigCustomEndpoint) *IngressConfig {
	if i.CustomEndpoint = ce; i.CustomEndpoint == nil {
		i.nullFields = append(i.nullFields, "CustomEndpoint")
	}
	return i
}

func (i *IngressConfig) SetPrivateLink(pl *IngressConfigPrivateLink) *IngressConfig {
	if i.PrivateLink = pl; i.PrivateLink == nil {
		i.nullFields = append(i.nullFields, "PrivateLink")
	}
	return i
}

// region Ingress controller

func (c IngressConfigController) MarshalJSON() ([]byte, error) {
	type noMethod IngressConfigController
	raw := noMethod(c)
	return jsonutil.MarshalJSON(raw, c.forceSendFields, c.nullFields)
}

func (c *IngressConfigController) SetManaged(v *bool) *IngressConfigController {
	if c.Managed = v; c.Managed == nil {
		c.nullFields = append(c.nullFields, "Managed")
	}
	return c
}

// endregion

// region Ingress load balancer

func (lb IngressConfigLoadBalancer) MarshalJSON() ([]byte, error) {
	type noMethod IngressConfigLoadBalancer
	raw := noMethod(lb)
	return jsonutil.MarshalJSON(raw, lb.forceSendFields, lb.nullFields)
}

func (lb *IngressConfigLoadBalancer) SetManaged(v *bool) *IngressConfigLoadBalancer {
	if lb.Managed = v; lb.Managed == nil {
		lb.nullFields = append(lb.nullFields, "Managed")
	}
	return lb
}

func (lb *IngressConfigLoadBalancer) SetTargetGroupARN(v *string) *IngressConfigLoadBalancer {
	if lb.TargetGroupARN = v; lb.TargetGroupARN == nil {
		lb.nullFields = append(lb.nullFields, "TargetGroupARN")
	}
	return lb
}

func (lb *IngressConfigLoadBalancer) SetServiceAnnotations(v map[string]string) *IngressConfigLoadBalancer {
	if lb.ServiceAnnotations = v; lb.ServiceAnnotations == nil {
		lb.nullFields = append(lb.nullFields, "ServiceAnnotations")
	}
	return lb
}

// endregion

// region Ingress custom endpoint

func (ce IngressConfigCustomEndpoint) MarshalJSON() ([]byte, error) {
	type noMethod IngressConfigCustomEndpoint
	raw := noMethod(ce)
	return jsonutil.MarshalJSON(raw, ce.forceSendFields, ce.nullFields)
}

func (ce *IngressConfigCustomEndpoint) SetEnabled(v *bool) *IngressConfigCustomEndpoint {
	if ce.Enabled = v; ce.Enabled == nil {
		ce.nullFields = append(ce.nullFields, "Enabled")
	}
	return ce
}

func (ce *IngressConfigCustomEndpoint) SetAddress(v *string) *IngressConfigCustomEndpoint {
	if ce.Address = v; ce.Address == nil {
		ce.nullFields = append(ce.nullFields, "Address")
	}
	return ce
}

// endregion

// region Ingress private link

func (pl IngressConfigPrivateLink) MarshalJSON() ([]byte, error) {
	type noMethod IngressConfigPrivateLink
	raw := noMethod(pl)
	return jsonutil.MarshalJSON(raw, pl.forceSendFields, pl.nullFields)
}

func (pl *IngressConfigPrivateLink) SetEnabled(v *bool) *IngressConfigPrivateLink {
	if pl.Enabled = v; pl.Enabled == nil {
		pl.nullFields = append(pl.nullFields, "Enabled")
	}
	return pl
}

func (pl *IngressConfigPrivateLink) SetVPCEndpointService(v *string) *IngressConfigPrivateLink {
	if pl.VPCEndpointService = v; pl.VPCEndpointService == nil {
		pl.nullFields = append(pl.nullFields, "VPCEndpointService")
	}
	return pl
}

// endregion

// endregion

// region Webhook

func (w WebhookConfig) MarshalJSON() ([]byte, error) {
	type noMethod WebhookConfig
	raw := noMethod(w)
	return jsonutil.MarshalJSON(raw, w.forceSendFields, w.nullFields)
}

func (w *WebhookConfig) SetUseHostNetwork(v *bool) *WebhookConfig {
	if w.UseHostNetwork = v; w.UseHostNetwork == nil {
		w.nullFields = append(w.nullFields, "UseHostNetwork")
	}
	return w
}

func (w *WebhookConfig) SetHostNetworkPorts(v []*int) *WebhookConfig {
	if w.HostNetworkPorts = v; w.HostNetworkPorts == nil {
		w.nullFields = append(w.nullFields, "HostNetworkPorts")
	}
	return w
}

// endregion

// region Compute

func (c ComputeConfig) MarshalJSON() ([]byte, error) {
	type noMethod ComputeConfig
	raw := noMethod(c)
	return jsonutil.MarshalJSON(raw, c.forceSendFields, c.nullFields)
}

func (c *ComputeConfig) SetUseTaints(v *bool) *ComputeConfig {
	if c.UseTaints = v; c.UseTaints == nil {
		c.nullFields = append(c.nullFields, "UseTaints")
	}
	return c
}

func (c *ComputeConfig) SetCreateVNGs(v *bool) *ComputeConfig {
	if c.CreateVngs = v; c.CreateVngs == nil {
		c.nullFields = append(c.nullFields, "CreateVngs")
	}
	return c
}

// endregion

// region Log collection

func (l LogCollectionConfig) MarshalJSON() ([]byte, error) {
	type noMethod LogCollectionConfig
	raw := noMethod(l)
	return jsonutil.MarshalJSON(raw, l.forceSendFields, l.nullFields)
}

func (l *LogCollectionConfig) SetCollectDriverLogs(v *bool) *LogCollectionConfig {
	if l.CollectDriverLogs = v; l.CollectDriverLogs == nil {
		l.nullFields = append(l.nullFields, "CollectDriverLogs")
	}
	return l
}

func (l *LogCollectionConfig) SetCollectAppLogs(v *bool) *LogCollectionConfig {
	if l.CollectAppLogs = v; l.CollectAppLogs == nil {
		l.nullFields = append(l.nullFields, "CollectAppLogs")
	}
	return l
}

// endregion

// region Spark

func (s SparkConfig) MarshalJSON() ([]byte, error) {
	type noMethod SparkConfig
	raw := noMethod(s)
	return jsonutil.MarshalJSON(raw, s.forceSendFields, s.nullFields)
}

func (s *SparkConfig) SetAppNamespaces(v []*string) *SparkConfig {
	if s.AppNamespaces = v; s.AppNamespaces == nil {
		s.nullFields = append(s.nullFields, "AppNamespaces")
	}
	return s
}

// endregion
