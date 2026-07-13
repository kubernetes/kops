package linodego

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/linode/linodego/v2/internal/parseabletime"
)

type AlertDefinitionStatus string

const (
	AlertDefinitionStatusProvisioning AlertDefinitionStatus = "provisioning"
	AlertDefinitionStatusEnabling     AlertDefinitionStatus = "enabling"
	AlertDefinitionStatusDisabling    AlertDefinitionStatus = "disabling"
	AlertDefinitionStatusEnabled      AlertDefinitionStatus = "enabled"
	AlertDefinitionStatusDisabled     AlertDefinitionStatus = "disabled"
	AlertDefinitionStatusFailed       AlertDefinitionStatus = "failed"
)

// AlertDefinitionScope represents the scope of an alert definition: "account", "entity", or "region". Defaults to "entity".
type AlertDefinitionScope string

const (
	AlertDefinitionScopeAccount AlertDefinitionScope = "account"
	AlertDefinitionScopeEntity  AlertDefinitionScope = "entity"
	AlertDefinitionScopeRegion  AlertDefinitionScope = "region"
)

// AlertDefinitionEntities represents entity metadata for an alert definition.
// For entity scoped alerts, entities contains the URL to list entities, a count, and a has_more_resources flag.
// For region/account scoped alerts, the entities are returned as an empty object.
type AlertDefinitionEntities struct {
	URL              string `json:"url"`
	Count            int    `json:"count"`
	HasMoreResources bool   `json:"has_more_resources"`
}

// AlertDefinitionEntity represents a single entity associated with an alert definition.
type AlertDefinitionEntity struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

// AlertDefinition represents an ACLP Alert Definition object
type AlertDefinition struct {
	ID                int                     `json:"id"`
	Label             string                  `json:"label"`
	Severity          int                     `json:"severity"`
	Type              string                  `json:"type"`
	ServiceType       string                  `json:"service_type"`
	Status            AlertDefinitionStatus   `json:"status"`
	RuleCriteria      RuleCriteria            `json:"rule_criteria"`
	TriggerConditions TriggerConditions       `json:"trigger_conditions"`
	AlertChannels     []AlertChannelEnvelope  `json:"alert_channels"`
	Created           *time.Time              `json:"-"`
	Updated           *time.Time              `json:"-"`
	UpdatedBy         string                  `json:"updated_by"`
	CreatedBy         string                  `json:"created_by"`
	Description       string                  `json:"description"`
	Class             string                  `json:"class"`
	Scope             AlertDefinitionScope    `json:"scope"`
	Regions           []string                `json:"regions"`
	Entities          AlertDefinitionEntities `json:"entities"`
	GroupBy           []string                `json:"group_by"`
}

// TriggerConditions represents the trigger conditions for an alert.
type TriggerConditions struct {
	CriteriaCondition       string `json:"criteria_condition,omitzero"`
	EvaluationPeriodSeconds int    `json:"evaluation_period_seconds,omitzero"`
	PollingIntervalSeconds  int    `json:"polling_interval_seconds,omitzero"`
	TriggerOccurrences      int    `json:"trigger_occurrences,omitzero"`
}

// RuleCriteria represents the rule criteria for an alert.
type RuleCriteria struct {
	Rules []Rule `json:"rules,omitzero"`
}

// Rule represents a single rule for an alert.
type Rule struct {
	AggregateFunction string            `json:"aggregate_function"`
	DimensionFilters  []DimensionFilter `json:"dimension_filters"`
	Label             string            `json:"label"`
	Metric            string            `json:"metric"`
	Operator          string            `json:"operator"`
	Threshold         float64           `json:"threshold"`
	Unit              string            `json:"unit"`
}

// DimensionFilter represents a single dimension filter used inside a Rule.
type DimensionFilter struct {
	DimensionLabel string `json:"dimension_label"`
	Label          string `json:"label"`
	Operator       string `json:"operator"`
	Value          string `json:"value"`
}

// RuleCriteriaOptions represents the rule criteria options for an alert.
type RuleCriteriaOptions struct {
	Rules []RuleOptions `json:"rules,omitzero"`
}

// RuleOptions represents a single rule option for an alert.
type RuleOptions struct {
	AggregateFunction string                   `json:"aggregate_function,omitzero"`
	DimensionFilters  []DimensionFilterOptions `json:"dimension_filters,omitzero"`
	Metric            string                   `json:"metric,omitzero"`
	Operator          string                   `json:"operator,omitzero"`
	Threshold         float64                  `json:"threshold,omitzero"`
}

// DimensionFilterOptions represents a single dimension filter option used inside a Rule.
type DimensionFilterOptions struct {
	DimensionLabel string `json:"dimension_label,omitzero"`
	Operator       string `json:"operator,omitzero"`
	Value          string `json:"value,omitzero"`
}

// AlertChannelEnvelope represents a single alert channel entry returned inside alert definition
type AlertChannelEnvelope struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

// AlertType represents the type of alert: "user" or "system"
type AlertType string

const (
	AlertTypeUser   AlertType = "user"
	AlertTypeSystem AlertType = "system"
)

// Severity represents the severity level of an alert.
// 0 = Severe, 1 = Medium, 2 = Low, 3 = Info
type Severity int

const (
	SeveritySevere Severity = 0
	SeverityMedium Severity = 1
	SeverityLow    Severity = 2
	SeverityInfo   Severity = 3
)

// CriteriaCondition represents supported criteria conditions
type CriteriaCondition string

const (
	CriteriaConditionAll CriteriaCondition = "ALL"
)

// AlertDefinitionCreateOptions are the options used to create a new alert definition.
type AlertDefinitionCreateOptions struct {
	Label             string               `json:"label"`
	Severity          int                  `json:"severity"`
	ChannelIDs        []int                `json:"channel_ids"`
	RuleCriteria      *RuleCriteriaOptions `json:"rule_criteria,omitzero"`
	TriggerConditions *TriggerConditions   `json:"trigger_conditions,omitzero"`
	EntityIDs         []string             `json:"entity_ids,omitzero"`
	Description       *string              `json:"description,omitzero"`
	Scope             AlertDefinitionScope `json:"scope,omitzero"`
	Regions           []string             `json:"regions,omitzero"`
	GroupBy           []string             `json:"group_by,omitzero"`
}

// AlertDefinitionUpdateOptions are the options used to update an alert definition.
type AlertDefinitionUpdateOptions struct {
	Label             string                 `json:"label"`
	Severity          int                    `json:"severity"`
	ChannelIDs        []int                  `json:"channel_ids"`
	RuleCriteria      *RuleCriteriaOptions   `json:"rule_criteria,omitzero"`
	TriggerConditions *TriggerConditions     `json:"trigger_conditions,omitzero"`
	EntityIDs         []string               `json:"entity_ids,omitzero"`
	Description       *string                `json:"description,omitzero"`
	Status            *AlertDefinitionStatus `json:"status,omitzero"`
	Regions           []string               `json:"regions,omitzero"`
	GroupBy           []string               `json:"group_by,omitzero"`
}

// AlertDefinitionCloneOptions are the options used to clone an existing alert definition.
type AlertDefinitionCloneOptions struct {
	Label             string               `json:"label"`
	Severity          *int                 `json:"severity,omitzero"`
	ChannelIDs        []int                `json:"channel_ids,omitzero"`
	RuleCriteria      *RuleCriteriaOptions `json:"rule_criteria,omitzero"`
	TriggerConditions *TriggerConditions   `json:"trigger_conditions,omitzero"`
	EntityIDs         []string             `json:"entity_ids,omitzero"`
	Description       *string              `json:"description,omitzero"`
	Scope             AlertDefinitionScope `json:"scope,omitzero"`
	Regions           []string             `json:"regions,omitzero"`
	GroupBy           []string             `json:"group_by,omitzero"`
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (i *AlertDefinition) UnmarshalJSON(b []byte) error {
	type Mask AlertDefinition

	p := struct {
		*Mask

		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.Updated = (*time.Time)(p.Updated)

	return nil
}

// ListMonitorAlertDefinitions returns a paginated list of ACLP Monitor Alert Definitions by service type.
func (c *Client) ListMonitorAlertDefinitions(
	ctx context.Context,
	serviceType string,
	opts *ListOptions,
) ([]AlertDefinition, error) {
	endpoint := formatAPIPath("monitor/services/%s/alert-definitions", serviceType)
	return getPaginatedResults[AlertDefinition](ctx, c, endpoint, opts)
}

// ListAllMonitorAlertDefinitions returns a paginated list of all ACLP Monitor Alert Definitions under this account.
func (c *Client) ListAllMonitorAlertDefinitions(
	ctx context.Context,
	opts *ListOptions,
) ([]AlertDefinition, error) {
	endpoint := formatAPIPath("monitor/alert-definitions")
	return getPaginatedResults[AlertDefinition](ctx, c, endpoint, opts)
}

// GetMonitorAlertDefinition gets an ACLP Monitor Alert Definition.
func (c *Client) GetMonitorAlertDefinition(
	ctx context.Context,
	serviceType string,
	alertID int,
) (*AlertDefinition, error) {
	e := formatAPIPath("monitor/services/%s/alert-definitions/%d", serviceType, alertID)
	return doGETRequest[AlertDefinition](ctx, c, e)
}

// CreateMonitorAlertDefinition creates an ACLP Monitor Alert Definition.
func (c *Client) CreateMonitorAlertDefinition(
	ctx context.Context,
	serviceType string,
	opts AlertDefinitionCreateOptions,
) (*AlertDefinition, error) {
	e := formatAPIPath("monitor/services/%s/alert-definitions", serviceType)
	return doPOSTRequest[AlertDefinition](ctx, c, e, opts)
}

// CreateMonitorAlertDefinitionWithIdempotency creates an ACLP Monitor Alert Definition
// and optionally sends an Idempotency-Key header to make the request idempotent.
func (c *Client) CreateMonitorAlertDefinitionWithIdempotency(
	ctx context.Context,
	serviceType string,
	opts AlertDefinitionCreateOptions,
	idempotencyKey string,
) (*AlertDefinition, error) {
	e := formatAPIPath("monitor/services/%s/alert-definitions", serviceType)

	var result AlertDefinition

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	params := requestParams{
		Response: &result,
		Body:     bytes.NewReader(body),
	}

	if idempotencyKey != "" {
		params.Headers = http.Header{
			"Idempotency-Key": {idempotencyKey},
		}
	}

	err = c.doRequest(ctx, http.MethodPost, e, params, nil)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateMonitorAlertDefinition updates an ACLP Monitor Alert Definition.
func (c *Client) UpdateMonitorAlertDefinition(
	ctx context.Context,
	serviceType string,
	alertID int,
	opts AlertDefinitionUpdateOptions,
) (*AlertDefinition, error) {
	e := formatAPIPath("monitor/services/%s/alert-definitions/%d", serviceType, alertID)
	return doPUTRequest[AlertDefinition](ctx, c, e, opts)
}

// DeleteMonitorAlertDefinition deletes an ACLP Monitor Alert Definition.
func (c *Client) DeleteMonitorAlertDefinition(ctx context.Context, serviceType string, alertID int) error {
	e := formatAPIPath("monitor/services/%s/alert-definitions/%d", serviceType, alertID)
	return doDELETERequest(ctx, c, e)
}

// ListMonitorAlertDefinitionEntities gets the entities associated with an ACLP Monitor Alert Definition.
func (c *Client) ListMonitorAlertDefinitionEntities(
	ctx context.Context,
	serviceType string,
	alertID int,
	opts *ListOptions,
) ([]AlertDefinitionEntity, error) {
	e := formatAPIPath("monitor/services/%s/alert-definitions/%d/entities", serviceType, alertID)
	return getPaginatedResults[AlertDefinitionEntity](ctx, c, e, opts)
}

// CloneMonitorAlertDefinition clones an ACLP Monitor Alert Definition.
func (c *Client) CloneMonitorAlertDefinition(
	ctx context.Context,
	serviceType string,
	alertID int,
	opts AlertDefinitionCloneOptions,
) (*AlertDefinition, error) {
	e := formatAPIPath("monitor/services/%s/alert-definitions/%d/clone", serviceType, alertID)
	return doPOSTRequest[AlertDefinition](ctx, c, e, opts)
}
