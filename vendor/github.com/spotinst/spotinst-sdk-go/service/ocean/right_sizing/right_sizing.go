package right_sizing

import (
	"context"
	"encoding/json"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
	"io/ioutil"
	"net/http"
)

type RightsizingRule struct {
	RuleName                                *string                                  `json:"ruleName,omitempty"`
	OceanId                                 *string                                  `json:"oceanId,omitempty"`
	RestartReplicas                         *string                                  `json:"restartReplicas,omitempty"`
	ExcludePreliminaryRecommendations       *bool                                    `json:"excludePreliminaryRecommendations,omitempty"`
	RecommendationApplicationIntervals      []*RecommendationApplicationIntervals    `json:"recommendationApplicationIntervals,omitempty"`
	RecommendationApplicationMinThreshold   *RecommendationApplicationMinThreshold   `json:"recommendationApplicationMinThreshold,omitempty"`
	RecommendationApplicationBoundaries     *RecommendationApplicationBoundaries     `json:"recommendationApplicationBoundaries,omitempty"`
	RecommendationApplicationOverheadValues *RecommendationApplicationOverheadValues `json:"recommendationApplicationOverheadValues,omitempty"`
	RecommendationApplicationHPA            *RecommendationApplicationHPA            `json:"recommendationApplicationHPA,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type RecommendationApplicationIntervals struct {
	RepetitionBasis        *string                 `json:"repetitionBasis,omitempty"`
	WeeklyRepetitionBasis  *WeeklyRepetitionBasis  `json:"weeklyRepetitionBasis,omitempty"`
	MonthlyRepetitionBasis *MonthlyRepetitionBasis `json:"monthlyRepetitionBasis,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type WeeklyRepetitionBasis struct {
	IntervalDays  []string       `json:"intervalDays,omitempty"`
	IntervalHours *IntervalHours `json:"intervalHours,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type MonthlyRepetitionBasis struct {
	IntervalMonths        []int                  `json:"intervalMonths,omitempty"`
	WeekOfTheMonth        []string               `json:"weekOfTheMonth,omitempty"`
	WeeklyRepetitionBasis *WeeklyRepetitionBasis `json:"weeklyRepetitionBasis,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type IntervalHours struct {
	StartTime *string `json:"startTime,omitempty"`
	EndTime   *string `json:"endTime,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type RecommendationApplicationMinThreshold struct {
	CpuPercentage    *float64 `json:"cpuPercentage,omitempty"`
	MemoryPercentage *float64 `json:"memoryPercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type RecommendationApplicationBoundaries struct {
	Cpu    *Cpu    `json:"cpu,omitempty"`
	Memory *Memory `json:"memory,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Cpu struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Memory struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`

	forceSendFields []string
	nullFields      []string
}
type Label struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Namespace struct {
	NamespaceName *string     `json:"namespaceName,omitempty"`
	Workloads     []*Workload `json:"workloads,omitempty"`
	Labels        []*Label    `json:"labels,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type Workload struct {
	Name         *string `json:"name,omitempty"`
	WorkloadType *string `json:"workloadType,omitempty"`
	RegexName    *string `json:"regexName,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type RecommendationApplicationOverheadValues struct {
	CpuPercentage    *float64 `json:"cpuPercentage,omitempty"`
	MemoryPercentage *float64 `json:"memoryPercentage,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type RecommendationApplicationHPA struct {
	AllowHPARecommendations *bool `json:"allowHPARecommendations,omitempty"`

	forceSendFields []string
	nullFields      []string
}

type ListRightsizingRulesInput struct {
	OceanId *string `json:"oceanId,omitempty"`
}

type ListRightsizingRulesOutput struct {
	RightsizingRules []*RightsizingRule `json:"rightsizingRule,omitempty"`
}

type RightSizingAttachDetachInput struct {
	RuleName   *string      `json:"ruleName,omitempty"`
	OceanId    *string      `json:"oceanId,omitempty"`
	Namespaces []*Namespace `json:"namespaces,omitempty"`
}

type RightSizingAttachDetachOutput struct{}

type ReadRightsizingRuleInput struct {
	RuleName *string `json:"ruleName,omitempty"`
	OceanId  *string `json:"oceanId,omitempty"`
}

type ReadRightsizingRuleOutput struct {
	RightsizingRule *RightsizingRule `json:"rightsizingRule,omitempty"`
}

type UpdateRightsizingRuleInput struct {
	RuleName        *string          `json:"ruleName,omitempty"`
	RightsizingRule *RightsizingRule `json:"rightsizingRule,omitempty"`
}

type UpdateRightsizingRuleOutput struct {
	RightsizingRule *RightsizingRule `json:"rightsizingRule,omitempty"`
}

type DeleteRightsizingRuleInput struct {
	RuleNames []string `json:"ruleNames,omitempty"`
	OceanId   *string  `json:"oceanId,omitempty"`
}

type DeleteRightsizingRuleOutput struct{}

type CreateRightsizingRuleInput struct {
	RightsizingRule *RightsizingRule `json:"rightsizingRule,omitempty"`
}

type CreateRightsizingRuleOutput struct {
	RightsizingRule *RightsizingRule `json:"rightsizingRule,omitempty"`
}

func rightsizingRuleFromJSON(in []byte) (*RightsizingRule, error) {
	b := new(RightsizingRule)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func rightsizingRulesFromJSON(in []byte) ([]*RightsizingRule, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*RightsizingRule, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := rightsizingRuleFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func rightsizingRulesFromHttpResponse(resp *http.Response) ([]*RightsizingRule, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return rightsizingRulesFromJSON(body)
}

func (s *ServiceOp) ListRightsizingRules(ctx context.Context, input *ListRightsizingRulesInput) (*ListRightsizingRulesOutput, error) {
	path, err := uritemplates.Expand("/ocean/{oceanId}/rightSizing/rule", uritemplates.Values{
		"oceanId": spotinst.StringValue(input.OceanId),
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

	gs, err := rightsizingRulesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListRightsizingRulesOutput{RightsizingRules: gs}, nil
}

func (s *ServiceOp) CreateRightsizingRule(ctx context.Context, input *CreateRightsizingRuleInput) (*CreateRightsizingRuleOutput, error) {
	path, err := uritemplates.Expand("/ocean/{oceanId}/rightSizing/rule", uritemplates.Values{
		"oceanId": spotinst.StringValue(input.RightsizingRule.OceanId),
	})
	if err != nil {
		return nil, err
	}

	// We do NOT need the ID anymore, so let's drop it.
	input.RightsizingRule.OceanId = nil
	r := client.NewRequest(http.MethodPost, path)
	r.Obj = input.RightsizingRule

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := rightsizingRulesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateRightsizingRuleOutput)
	if len(gs) > 0 {
		output.RightsizingRule = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) ReadRightsizingRule(ctx context.Context, input *ReadRightsizingRuleInput) (*ReadRightsizingRuleOutput, error) {
	path, err := uritemplates.Expand("/ocean/{oceanId}/rightSizing/rule/{ruleName}", uritemplates.Values{
		"oceanId":  spotinst.StringValue(input.OceanId),
		"ruleName": spotinst.StringValue(input.RuleName),
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

	gs, err := rightsizingRulesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadRightsizingRuleOutput)
	if len(gs) > 0 {
		output.RightsizingRule = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) UpdateRightsizingRule(ctx context.Context, input *UpdateRightsizingRuleInput) (*UpdateRightsizingRuleOutput, error) {
	path, err := uritemplates.Expand("/ocean/{oceanId}/rightSizing/rule/{ruleName}", uritemplates.Values{
		"oceanId":  spotinst.StringValue(input.RightsizingRule.OceanId),
		"ruleName": spotinst.StringValue(input.RuleName),
	})
	if err != nil {
		return nil, err
	}

	input.RightsizingRule.OceanId = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input.RightsizingRule
	input.RuleName = nil

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := rightsizingRulesFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateRightsizingRuleOutput)
	if len(gs) > 0 {
		output.RightsizingRule = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) DeleteRightsizingRules(ctx context.Context, input *DeleteRightsizingRuleInput) (*DeleteRightsizingRuleOutput, error) {
	path, err := uritemplates.Expand("/ocean/{oceanId}/rightSizing/rule", uritemplates.Values{
		"oceanId": spotinst.StringValue(input.OceanId),
	})
	if err != nil {
		return nil, err
	}

	// We do NOT need the ID anymore, so let's drop it.
	input.OceanId = nil

	r := client.NewRequest(http.MethodDelete, path)
	r.Obj = input
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteRightsizingRuleOutput{}, nil
}

func (s *ServiceOp) AttachRightSizingRule(ctx context.Context, input *RightSizingAttachDetachInput) (*RightSizingAttachDetachOutput, error) {
	path, err := uritemplates.Expand("/ocean/{oceanId}/rightSizing/rule/{ruleName}/attachment", uritemplates.Values{
		"oceanId":  spotinst.StringValue(input.OceanId),
		"ruleName": spotinst.StringValue(input.RuleName),
	})

	r := client.NewRequest(http.MethodPost, path)

	input.OceanId = nil
	input.RuleName = nil

	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &RightSizingAttachDetachOutput{}, nil
}

func (s *ServiceOp) DetachRightSizingRule(ctx context.Context, input *RightSizingAttachDetachInput) (*RightSizingAttachDetachOutput, error) {
	path, err := uritemplates.Expand("/ocean/{oceanId}/rightSizing/rule/{ruleName}/detachment", uritemplates.Values{
		"oceanId":  spotinst.StringValue(input.OceanId),
		"ruleName": spotinst.StringValue(input.RuleName),
	})

	r := client.NewRequest(http.MethodPost, path)

	input.OceanId = nil
	input.RuleName = nil

	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &RightSizingAttachDetachOutput{}, nil
}

// region RightsizingRule

func (o RightsizingRule) MarshalJSON() ([]byte, error) {
	type noMethod RightsizingRule
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *RightsizingRule) SetRuleName(v *string) *RightsizingRule {
	if o.RuleName = v; o.RuleName == nil {
		o.nullFields = append(o.nullFields, "RuleName")
	}
	return o
}

func (o *RightsizingRule) SetOceanId(v *string) *RightsizingRule {
	if o.OceanId = v; o.OceanId == nil {
		o.nullFields = append(o.nullFields, "oceanId")
	}
	return o
}

func (o *RightsizingRule) SetRestartReplicas(v *string) *RightsizingRule {
	if o.RestartReplicas = v; o.RestartReplicas == nil {
		o.nullFields = append(o.nullFields, "RestartReplicas")
	}
	return o
}

func (o *RightsizingRule) SetExcludePreliminaryRecommendations(v *bool) *RightsizingRule {
	if o.ExcludePreliminaryRecommendations = v; o.ExcludePreliminaryRecommendations == nil {
		o.nullFields = append(o.nullFields, "ExcludePreliminaryRecommendations")
	}
	return o
}

func (o *RightsizingRule) SetRecommendationApplicationIntervals(v []*RecommendationApplicationIntervals) *RightsizingRule {
	if o.RecommendationApplicationIntervals = v; o.RecommendationApplicationIntervals == nil {
		o.nullFields = append(o.nullFields, "RecommendationApplicationIntervals")
	}
	return o
}

func (o *RightsizingRule) SetRecommendationApplicationBoundaries(v *RecommendationApplicationBoundaries) *RightsizingRule {
	if o.RecommendationApplicationBoundaries = v; o.RecommendationApplicationBoundaries == nil {
		o.nullFields = append(o.nullFields, "RecommendationApplicationBoundaries")
	}
	return o
}

func (o *RightsizingRule) SetRecommendationApplicationMinThreshold(v *RecommendationApplicationMinThreshold) *RightsizingRule {
	if o.RecommendationApplicationMinThreshold = v; o.RecommendationApplicationMinThreshold == nil {
		o.nullFields = append(o.nullFields, "RecommendationApplicationMinThreshold")
	}
	return o
}

func (o *RightsizingRule) SetRecommendationApplicationOverheadValues(v *RecommendationApplicationOverheadValues) *RightsizingRule {
	if o.RecommendationApplicationOverheadValues = v; o.RecommendationApplicationOverheadValues == nil {
		o.nullFields = append(o.nullFields, "RecommendationApplicationOverheadValues")
	}
	return o
}

func (o *RightsizingRule) SetRecommendationApplicationHPA(v *RecommendationApplicationHPA) *RightsizingRule {
	if o.RecommendationApplicationHPA = v; o.RecommendationApplicationHPA == nil {
		o.nullFields = append(o.nullFields, "RecommendationApplicationHPA")
	}
	return o
}

// region RecommendationApplicationInterval

func (o RecommendationApplicationIntervals) MarshalJSON() ([]byte, error) {
	type noMethod RecommendationApplicationIntervals
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *RecommendationApplicationIntervals) SetRepetitionBasis(v *string) *RecommendationApplicationIntervals {
	if o.RepetitionBasis = v; o.RepetitionBasis == nil {
		o.nullFields = append(o.nullFields, "RepetitionBasis")
	}
	return o
}

func (o *RecommendationApplicationIntervals) SetWeeklyRepetitionBasis(v *WeeklyRepetitionBasis) *RecommendationApplicationIntervals {
	if o.WeeklyRepetitionBasis = v; o.WeeklyRepetitionBasis == nil {
		o.nullFields = append(o.nullFields, "WeeklyRepetitionBasis")
	}
	return o
}

func (o *RecommendationApplicationIntervals) SetMonthlyRepetitionBasis(v *MonthlyRepetitionBasis) *RecommendationApplicationIntervals {
	if o.MonthlyRepetitionBasis = v; o.MonthlyRepetitionBasis == nil {
		o.nullFields = append(o.nullFields, "MonthlyRepetitionBasis")
	}
	return o
}

// region WeeklyRepetitionBasis

func (o *WeeklyRepetitionBasis) SetIntervalDays(v []string) *WeeklyRepetitionBasis {
	if o.IntervalDays = v; o.IntervalDays == nil {
		o.nullFields = append(o.nullFields, "IntervalDays")
	}
	return o
}

func (o *WeeklyRepetitionBasis) SetIntervalHours(v *IntervalHours) *WeeklyRepetitionBasis {
	if o.IntervalHours = v; o.IntervalHours == nil {
		o.nullFields = append(o.nullFields, "IntervalHours")
	}
	return o
}

// region IntervalHours

func (o *IntervalHours) SetStartTime(v *string) *IntervalHours {
	if o.StartTime = v; o.StartTime == nil {
		o.nullFields = append(o.nullFields, "StartTime")
	}
	return o
}

func (o *IntervalHours) SetEndTime(v *string) *IntervalHours {
	if o.EndTime = v; o.EndTime == nil {
		o.nullFields = append(o.nullFields, "EndTime")
	}
	return o
}

// region MonthlyRepetitionBasis

func (o *MonthlyRepetitionBasis) SetIntervalMonths(v []int) *MonthlyRepetitionBasis {
	if o.IntervalMonths = v; o.IntervalMonths == nil {
		o.nullFields = append(o.nullFields, "IntervalMonths")
	}
	return o
}

func (o *MonthlyRepetitionBasis) SetWeekOfTheMonth(v []string) *MonthlyRepetitionBasis {
	if o.WeekOfTheMonth = v; o.WeekOfTheMonth == nil {
		o.nullFields = append(o.nullFields, "WeekOfTheMonth")
	}
	return o
}

func (o *MonthlyRepetitionBasis) SetMonthlyWeeklyRepetitionBasis(v *WeeklyRepetitionBasis) *MonthlyRepetitionBasis {
	if o.WeeklyRepetitionBasis = v; o.WeeklyRepetitionBasis == nil {
		o.nullFields = append(o.nullFields, "WeeklyRepetitionBasis")
	}
	return o
}

// region RecommendationApplicationBoundaries
func (o RecommendationApplicationBoundaries) MarshalJSON() ([]byte, error) {
	type noMethod RecommendationApplicationBoundaries
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}
func (o *RecommendationApplicationBoundaries) SetCpu(v *Cpu) *RecommendationApplicationBoundaries {
	if o.Cpu = v; o.Cpu == nil {
		o.nullFields = append(o.nullFields, "Cpu")
	}
	return o
}

func (o *RecommendationApplicationBoundaries) SetMemory(v *Memory) *RecommendationApplicationBoundaries {
	if o.Memory = v; o.Memory == nil {
		o.nullFields = append(o.nullFields, "Memory")
	}
	return o
}

// region Cpu
func (o Cpu) MarshalJSON() ([]byte, error) {
	type noMethod Cpu
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Cpu) SetMin(v *float64) *Cpu {
	if o.Min = v; o.Min == nil {
		o.nullFields = append(o.nullFields, "Cpu")
	}
	return o
}

func (o *Cpu) SetMax(v *float64) *Cpu {
	if o.Max = v; o.Min == nil {
		o.nullFields = append(o.nullFields, "Cpu")
	}
	return o
}

// region Memory
func (o Memory) MarshalJSON() ([]byte, error) {
	type noMethod Memory
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Memory) SetMin(v *int) *Memory {
	if o.Min = v; o.Min == nil {
		o.nullFields = append(o.nullFields, "Memory")
	}
	return o
}

func (o *Memory) SetMax(v *int) *Memory {
	if o.Max = v; o.Max == nil {
		o.nullFields = append(o.nullFields, "Memory")
	}
	return o
}

// region RecommendationApplicationMinThreshold
func (o RecommendationApplicationMinThreshold) MarshalJSON() ([]byte, error) {
	type noMethod RecommendationApplicationMinThreshold
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *RecommendationApplicationMinThreshold) SetCpuPercentage(v *float64) *RecommendationApplicationMinThreshold {
	if o.CpuPercentage = v; o.CpuPercentage == nil {
		o.nullFields = append(o.nullFields, "CpuPercentage")
	}
	return o
}

func (o *RecommendationApplicationMinThreshold) SetMemoryPercentage(v *float64) *RecommendationApplicationMinThreshold {
	if o.MemoryPercentage = v; o.MemoryPercentage == nil {
		o.nullFields = append(o.nullFields, "MemoryPercentage")
	}
	return o
}

// region RecommendationApplicationOverheadValues
func (o RecommendationApplicationOverheadValues) MarshalJSON() ([]byte, error) {
	type noMethod RecommendationApplicationOverheadValues
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *RecommendationApplicationOverheadValues) SetOverheadCpuPercentage(v *float64) *RecommendationApplicationOverheadValues {
	if o.CpuPercentage = v; o.CpuPercentage == nil {
		o.nullFields = append(o.nullFields, "CpuPercentage")
	}
	return o
}

func (o *RecommendationApplicationOverheadValues) SetOverheadMemoryPercentage(v *float64) *RecommendationApplicationOverheadValues {
	if o.MemoryPercentage = v; o.MemoryPercentage == nil {
		o.nullFields = append(o.nullFields, "MemoryPercentage")
	}
	return o
}

// region RecommendationApplicationHPA
func (o RecommendationApplicationHPA) MarshalJSON() ([]byte, error) {
	type noMethod RecommendationApplicationHPA
	raw := noMethod(o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *RecommendationApplicationHPA) SetAllowHPARecommendations(v *bool) *RecommendationApplicationHPA {
	if o.AllowHPARecommendations = v; o.AllowHPARecommendations == nil {
		o.nullFields = append(o.nullFields, "AllowHPARecommendations")
	}
	return o
}
