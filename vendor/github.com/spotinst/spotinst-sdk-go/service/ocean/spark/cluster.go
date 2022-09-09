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

	forceSendFields []string
	nullFields      []string
}

type LogCollectionConfig struct {
	CollectDriverLogs *bool `json:"collectDriverLogs,omitempty"`

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

type IngressConfig struct {
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	DeployIngress      *bool             `json:"deployIngress,omitempty"`

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
	ClusterID *string `json:"clusterId,omitempty"`
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

// endregion
