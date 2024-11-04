package azure_np

import "github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"

// NodePoolProperties region
type NodePoolProperties struct {
	MaxPodsPerNode     *int           `json:"maxPodsPerNode,omitempty"`
	EnableNodePublicIP *bool          `json:"enableNodePublicIP,omitempty"`
	OsDiskSizeGB       *int           `json:"osDiskSizeGB,omitempty"`
	OsDiskType         *string        `json:"osDiskType,omitempty"`
	OsType             *string        `json:"osType,omitempty"`
	OsSKU              *string        `json:"osSKU,omitempty"`
	KubernetesVersion  *string        `json:"kubernetesVersion,omitempty"`
	PodSubnetIDs       []string       `json:"podSubnetIDs,omitempty"`
	VnetSubnetIDs      []string       `json:"vnetSubnetIDs,omitempty"`
	LinuxOSConfig      *LinuxOSConfig `json:"linuxOSConfig,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o NodePoolProperties) MarshalJSON() ([]byte, error) {
	type noMethod NodePoolProperties
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *NodePoolProperties) SetLinuxOSConfig(v *LinuxOSConfig) *NodePoolProperties {
	if o.LinuxOSConfig = v; o.LinuxOSConfig == nil {
		o.nullFields = append(o.nullFields, "LinuxOSConfig")
	}
	return o
}

func (o *NodePoolProperties) SetMaxPodsPerNode(v *int) *NodePoolProperties {
	if o.MaxPodsPerNode = v; o.MaxPodsPerNode == nil {
		o.nullFields = append(o.nullFields, "MaxPodsPerNode")
	}
	return o
}

func (o *NodePoolProperties) SetEnableNodePublicIP(v *bool) *NodePoolProperties {
	if o.EnableNodePublicIP = v; o.EnableNodePublicIP == nil {
		o.nullFields = append(o.nullFields, "EnableNodePublicIP")
	}
	return o
}

func (o *NodePoolProperties) SetOsDiskSizeGB(v *int) *NodePoolProperties {
	if o.OsDiskSizeGB = v; o.OsDiskSizeGB == nil {
		o.nullFields = append(o.nullFields, "OsDiskSizeGB")
	}
	return o
}

func (o *NodePoolProperties) SetOsDiskType(v *string) *NodePoolProperties {
	if o.OsDiskType = v; o.OsDiskType == nil {
		o.nullFields = append(o.nullFields, "OsDiskType")
	}
	return o
}

func (o *NodePoolProperties) SetOsType(v *string) *NodePoolProperties {
	if o.OsType = v; o.OsType == nil {
		o.nullFields = append(o.nullFields, "OsType")
	}
	return o
}

func (o *NodePoolProperties) SetOsSKU(v *string) *NodePoolProperties {
	if o.OsSKU = v; o.OsSKU == nil {
		o.nullFields = append(o.nullFields, "OsSKU")
	}
	return o
}

func (o *NodePoolProperties) SetKubernetesVersion(v *string) *NodePoolProperties {
	if o.KubernetesVersion = v; o.KubernetesVersion == nil {
		o.nullFields = append(o.nullFields, "KubernetesVersion")
	}
	return o
}

func (o *NodePoolProperties) SetPodSubnetIDs(v []string) *NodePoolProperties {
	if o.PodSubnetIDs = v; o.PodSubnetIDs == nil {
		o.nullFields = append(o.nullFields, "PodSubnetIDs")
	}
	return o
}

func (o *NodePoolProperties) SetVnetSubnetIDs(v []string) *NodePoolProperties {
	if o.VnetSubnetIDs = v; o.VnetSubnetIDs == nil {
		o.nullFields = append(o.nullFields, "VnetSubnetIDs")
	}
	return o
}

// endregion

// LinuxOSConfig region
type LinuxOSConfig struct {
	Sysctls *Sysctls `json:"sysctls,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o *LinuxOSConfig) SetSysctls(v *Sysctls) *LinuxOSConfig {
	if o.Sysctls = v; o.Sysctls == nil {
		o.nullFields = append(o.nullFields, "Sysctls")
	}
	return o
}

// Sysctls region
type Sysctls struct {
	VmMaxMapCount *int `json:"vmMaxMapCount,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o *Sysctls) SetVmMaxMapCount(v *int) *Sysctls {
	if o.VmMaxMapCount = v; o.VmMaxMapCount == nil {
		o.nullFields = append(o.nullFields, "VmMaxMapCount")
	}
	return o
}

// NodeCountLimits region
type NodeCountLimits struct {
	MinCount *int `json:"minCount,omitempty"`
	MaxCount *int `json:"maxCount,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o NodeCountLimits) MarshalJSON() ([]byte, error) {
	type noMethod NodeCountLimits
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *NodeCountLimits) SetMinCount(v *int) *NodeCountLimits {
	if o.MinCount = v; o.MinCount == nil {
		o.nullFields = append(o.nullFields, "MinCount")
	}
	return o
}

func (o *NodeCountLimits) SetMaxCount(v *int) *NodeCountLimits {
	if o.MaxCount = v; o.MaxCount == nil {
		o.nullFields = append(o.nullFields, "MaxCount")
	}
	return o
}

// endregion

// Strategy region
type Strategy struct {
	SpotPercentage *int  `json:"spotPercentage,omitempty"`
	FallbackToOD   *bool `json:"fallbackToOd,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o Strategy) MarshalJSON() ([]byte, error) {
	type noMethod Strategy
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Strategy) SetSpotPercentage(v *int) *Strategy {
	if o.SpotPercentage = v; o.SpotPercentage == nil {
		o.nullFields = append(o.nullFields, "SpotPercentage")
	}
	return o
}

func (o *Strategy) SetFallbackToOD(v *bool) *Strategy {
	if o.FallbackToOD = v; o.FallbackToOD == nil {
		o.nullFields = append(o.nullFields, "FallbackToOD")
	}
	return o
}

// endregion

// region Taint

type Taint struct {
	Key    *string `json:"key,omitempty"`
	Value  *string `json:"value,omitempty"`
	Effect *string `json:"effect,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o Taint) MarshalJSON() ([]byte, error) {
	type noMethod Taint
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Taint) SetKey(v *string) *Taint {
	if o.Key = v; o.Key == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Taint) SetValue(v *string) *Taint {
	if o.Value = v; o.Value == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

func (o *Taint) SetEffect(v *string) *Taint {
	if o.Effect = v; o.Effect == nil {
		o.nullFields = append(o.nullFields, "Effect")
	}
	return o
}

// endregion

//region AutoScale

type AutoScale struct {
	Headrooms []*Headrooms `json:"headrooms,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Headrooms struct {
	CpuPerUnit    *int `json:"cpuPerUnit,omitempty"`
	MemoryPerUnit *int `json:"memoryPerUnit,omitempty"`
	GpuPerUnit    *int `json:"gpuPerUnit,omitempty"`
	NumberOfUnits *int `json:"numOfUnits,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o AutoScale) MarshalJSON() ([]byte, error) {
	type noMethod AutoScale
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AutoScale) SetHeadrooms(v []*Headrooms) *AutoScale {
	if o.Headrooms = v; o.Headrooms == nil {
		o.nullFields = append(o.nullFields, "Headrooms")
	}
	return o
}

//end region

//region Headrooms

func (o Headrooms) MarshalJSON() ([]byte, error) {
	type noMethod Headrooms
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Headrooms) SetCpuPerUnit(v *int) *Headrooms {
	if o.CpuPerUnit = v; o.CpuPerUnit == nil {
		o.nullFields = append(o.nullFields, "CpuPerUnit")
	}
	return o
}

func (o *Headrooms) SetMemoryPerUnit(v *int) *Headrooms {
	if o.MemoryPerUnit = v; o.MemoryPerUnit == nil {
		o.nullFields = append(o.nullFields, "MemoryPerUnit")
	}
	return o
}

func (o *Headrooms) SetGpuPerUnit(v *int) *Headrooms {
	if o.GpuPerUnit = v; o.GpuPerUnit == nil {
		o.nullFields = append(o.nullFields, "GpuPerUnit")
	}
	return o
}

func (o *Headrooms) SetNumOfUnits(v *int) *Headrooms {
	if o.NumberOfUnits = v; o.NumberOfUnits == nil {
		o.nullFields = append(o.nullFields, "NumberOfUnits")
	}
	return o
}

// endregion

//region Scheduling

type Scheduling struct {
	ShutdownHours *ShutdownHours `json:"shutdownHours,omitempty"`
	Tasks         []*Tasks       `json:"tasks,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ShutdownHours struct {
	TimeWindows []string `json:"timeWindows,omitempty"`
	IsEnabled   *bool    `json:"isEnabled,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Tasks struct {
	IsEnabled      *bool       `json:"isEnabled,omitempty"`
	TaskType       *string     `json:"taskType,omitempty"`
	CronExpression *string     `json:"cronExpression,omitempty"`
	Parameters     *Parameters `json:"parameters,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Parameters struct {
	ClusterRoll *ParameterClusterRoll `json:"clusterRoll,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ParameterClusterRoll struct {
	BatchSizePercentage       *int     `json:"batchSizePercentage,omitempty"`
	Comment                   *string  `json:"comment,omitempty"`
	RespectPdb                *bool    `json:"respectPdb,omitempty"`
	RespectRestrictScaleDown  *bool    `json:"respectRestrictScaleDown,omitempty"`
	BatchMinHealthyPercentage *int     `json:"batchMinHealthyPercentage,omitempty"`
	VngIds                    []string `json:"vngIds,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o Scheduling) MarshalJSON() ([]byte, error) {
	type noMethod Scheduling
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scheduling) SetShutdownHours(v *ShutdownHours) *Scheduling {
	if o.ShutdownHours = v; o.ShutdownHours == nil {
		o.nullFields = append(o.nullFields, "ShutdownHours")
	}
	return o
}

func (o ShutdownHours) MarshalJSON() ([]byte, error) {
	type noMethod ShutdownHours
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ShutdownHours) SetTimeWindows(v []string) *ShutdownHours {
	if o.TimeWindows = v; o.TimeWindows == nil {
		o.nullFields = append(o.nullFields, "TimeWindows")
	}
	return o
}

func (o *ShutdownHours) SetIsEnabled(v *bool) *ShutdownHours {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

// end region

// region vmSizes

type VmSizes struct {
	Filters *Filters `json:"filters,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Filters struct {
	MinVcpu               *int     `json:"minVCpu,omitempty"`
	MaxVcpu               *int     `json:"maxVCpu,omitempty"`
	MinMemoryGiB          *float64 `json:"minMemoryGiB,omitempty"`
	MaxMemoryGiB          *float64 `json:"maxMemoryGiB,omitempty"`
	Series                []string `json:"series,omitempty"`
	Architectures         []string `json:"architectures,omitempty"`
	ExcludeSeries         []string `json:"excludeSeries,omitempty"`
	AcceleratedNetworking *string  `json:"acceleratedNetworking,omitempty"`
	DiskPerformance       *string  `json:"diskPerformance,omitempty"`
	MinGpu                *float64 `json:"minGpu,omitempty"`
	MaxGpu                *float64 `json:"maxGpu,omitempty"`
	MinNICs               *int     `json:"minNICs,omitempty"`
	VmTypes               []string `json:"vmTypes,omitempty"`
	MinDisk               *int     `json:"minDisk,omitempty"`
	GpuTypes              []string `json:"gpuTypes,omitempty"`

	forceSendFields []string
	nullFields      []string
}

func (o VmSizes) MarshalJSON() ([]byte, error) {
	type noMethod VmSizes
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VmSizes) SetFilters(v *Filters) *VmSizes {
	if o.Filters = v; o.Filters == nil {
		o.nullFields = append(o.nullFields, "Filters")
	}
	return o
}

// end region

//region filters

func (o Filters) MarshalJSON() ([]byte, error) {
	type noMethod Filters
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Filters) SetMinVcpu(v *int) *Filters {
	if o.MinVcpu = v; o.MinVcpu == nil {
		o.nullFields = append(o.nullFields, "MinVcpu")
	}
	return o
}

func (o *Filters) SetMaxVcpu(v *int) *Filters {
	if o.MaxVcpu = v; o.MaxVcpu == nil {
		o.nullFields = append(o.nullFields, "MaxVcpu")
	}
	return o
}

func (o *Filters) SetMinMemoryGiB(v *float64) *Filters {
	if o.MinMemoryGiB = v; o.MinMemoryGiB == nil {
		o.nullFields = append(o.nullFields, "MinMemoryGiB")
	}
	return o
}

func (o *Filters) SetMaxMemoryGiB(v *float64) *Filters {
	if o.MaxMemoryGiB = v; o.MaxMemoryGiB == nil {
		o.nullFields = append(o.nullFields, "MaxMemoryGiB")
	}
	return o
}

func (o *Filters) SetSeries(v []string) *Filters {
	if o.Series = v; o.Series == nil {
		o.nullFields = append(o.nullFields, "Series")
	}
	return o
}

func (o *Filters) SetArchitectures(v []string) *Filters {
	if o.Architectures = v; o.Architectures == nil {
		o.nullFields = append(o.nullFields, "Architectures")
	}
	return o
}

func (o *Filters) SetExcludeSeries(v []string) *Filters {
	if o.ExcludeSeries = v; o.ExcludeSeries == nil {
		o.nullFields = append(o.nullFields, "ExcludeSeries")
	}
	return o
}

func (o *Filters) SetAcceleratedNetworking(v *string) *Filters {
	if o.AcceleratedNetworking = v; o.AcceleratedNetworking == nil {
		o.nullFields = append(o.nullFields, "AcceleratedNetworking")
	}
	return o
}

func (o *Filters) SetDiskPerformance(v *string) *Filters {
	if o.DiskPerformance = v; o.DiskPerformance == nil {
		o.nullFields = append(o.nullFields, "DiskPerformance")
	}
	return o
}

func (o *Filters) SetMinGpu(v *float64) *Filters {
	if o.MinGpu = v; o.MinGpu == nil {
		o.nullFields = append(o.nullFields, "MinGpu")
	}
	return o
}

func (o *Filters) SetMaxGpu(v *float64) *Filters {
	if o.MaxGpu = v; o.MaxGpu == nil {
		o.nullFields = append(o.nullFields, "MaxGpu")
	}
	return o
}

func (o *Filters) SetMinNICs(v *int) *Filters {
	if o.MinNICs = v; o.MinNICs == nil {
		o.nullFields = append(o.nullFields, "MinNICs")
	}
	return o
}

func (o *Filters) SetVmTypes(v []string) *Filters {
	if o.VmTypes = v; o.VmTypes == nil {
		o.nullFields = append(o.nullFields, "VmTypes")
	}
	return o
}

func (o *Filters) SetMinDisk(v *int) *Filters {
	if o.MinDisk = v; o.MinDisk == nil {
		o.nullFields = append(o.nullFields, "MinDisk")
	}
	return o
}

func (o *Filters) SetGpuTypes(v []string) *Filters {
	if o.GpuTypes = v; o.GpuTypes == nil {
		o.nullFields = append(o.nullFields, "GpuTypes")
	}
	return o
}

//end region

// region Tasks

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

func (o *Tasks) SetIsEnabled(v *bool) *Tasks {
	if o.IsEnabled = v; o.IsEnabled == nil {
		o.nullFields = append(o.nullFields, "IsEnabled")
	}
	return o
}

func (o *Tasks) SetTaskType(v *string) *Tasks {
	if o.TaskType = v; o.TaskType == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Tasks) SetCronExpression(v *string) *Tasks {
	if o.CronExpression = v; o.CronExpression == nil {
		o.nullFields = append(o.nullFields, "CronExpression")
	}
	return o
}

// endregion

// region Parameters
func (o Parameters) MarshalJSON() ([]byte, error) {
	type noMethod Parameters
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Tasks) SetParameters(v *Parameters) *Tasks {
	if o.Parameters = v; o.Parameters == nil {
		o.nullFields = append(o.nullFields, "Parameter")
	}
	return o
}

// endregion

// region ClusterRoll
func (o *Parameters) SetClusterRoll(v *ParameterClusterRoll) *Parameters {
	if o.ClusterRoll = v; o.ClusterRoll == nil {
		o.nullFields = append(o.nullFields, "ClusterRoll")
	}
	return o
}

func (o ParameterClusterRoll) MarshalJSON() ([]byte, error) {
	type noMethod ParameterClusterRoll
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

// endregion

// region ParameterClusterRoll
func (o *ParameterClusterRoll) SetBatchSizePercentage(v *int) *ParameterClusterRoll {
	if o.BatchSizePercentage = v; o.BatchSizePercentage == nil {
		o.nullFields = append(o.nullFields, "BatchSizePercentage")
	}
	return o
}

func (o *ParameterClusterRoll) SetComment(v *string) *ParameterClusterRoll {
	if o.Comment = v; o.Comment == nil {
		o.nullFields = append(o.nullFields, "Comment")
	}
	return o
}

func (o *ParameterClusterRoll) SetRespectPdb(v *bool) *ParameterClusterRoll {
	if o.RespectPdb = v; o.RespectPdb == nil {
		o.nullFields = append(o.nullFields, "RespectPdb")
	}
	return o
}

func (o *ParameterClusterRoll) SetRespectRestrictScaleDown(v *bool) *ParameterClusterRoll {
	if o.RespectRestrictScaleDown = v; o.RespectRestrictScaleDown == nil {
		o.nullFields = append(o.nullFields, "RespectRestrictScaleDown")
	}
	return o
}

func (o *ParameterClusterRoll) SetBatchMinHealthyPercentage(v *int) *ParameterClusterRoll {
	if o.BatchMinHealthyPercentage = v; o.BatchMinHealthyPercentage == nil {
		o.nullFields = append(o.nullFields, "BatchMinHealthyPercentage")
	}
	return o
}

func (o *ParameterClusterRoll) SetVngIds(v []string) *ParameterClusterRoll {
	if o.VngIds = v; o.VngIds == nil {
		o.nullFields = append(o.nullFields, "VngIds")
	}
	return o
}

// endregion
