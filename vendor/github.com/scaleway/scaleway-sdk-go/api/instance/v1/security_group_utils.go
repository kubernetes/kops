package instance

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// UpdateSecurityGroupRequest contains the parameters to update a security group
type UpdateSecurityGroupRequest struct {
	Zone            scw.Zone `json:"-"`
	SecurityGroupID string   `json:"-"`

	Name                  *string              `json:"name,omitempty"`
	Description           *string              `json:"description,omitempty"`
	InboundDefaultPolicy  *SecurityGroupPolicy `json:"inbound_default_policy,omitempty"`
	OutboundDefaultPolicy *SecurityGroupPolicy `json:"outbound_default_policy,omitempty"`
	Stateful              *bool                `json:"stateful,omitempty"`
	OrganizationDefault   *bool                `json:"organization_default,omitempty"`
	ProjectDefault        *bool                `json:"project_default,omitempty"`
	EnableDefaultSecurity *bool                `json:"enable_default_security,omitempty"`
	Tags                  *[]string            `json:"tags,omitempty"`
}

type UpdateSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup
}

// UpdateSecurityGroup updates a security group.
func (s *API) UpdateSecurityGroup(req *UpdateSecurityGroupRequest, opts ...scw.RequestOption) (*UpdateSecurityGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	getSGResponse, err := s.GetSecurityGroup(&GetSecurityGroupRequest{
		Zone:            req.Zone,
		SecurityGroupID: req.SecurityGroupID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	setRequest := &setSecurityGroupRequest{
		ID:                    getSGResponse.SecurityGroup.ID,
		Name:                  getSGResponse.SecurityGroup.Name,
		Description:           getSGResponse.SecurityGroup.Description,
		Organization:          getSGResponse.SecurityGroup.Organization,
		Project:               getSGResponse.SecurityGroup.Project,
		OrganizationDefault:   getSGResponse.SecurityGroup.OrganizationDefault,
		ProjectDefault:        getSGResponse.SecurityGroup.ProjectDefault,
		OutboundDefaultPolicy: getSGResponse.SecurityGroup.OutboundDefaultPolicy,
		InboundDefaultPolicy:  getSGResponse.SecurityGroup.InboundDefaultPolicy,
		Stateful:              getSGResponse.SecurityGroup.Stateful,
		Zone:                  req.Zone,
		EnableDefaultSecurity: getSGResponse.SecurityGroup.EnableDefaultSecurity,
		CreationDate:          getSGResponse.SecurityGroup.CreationDate,
		ModificationDate:      getSGResponse.SecurityGroup.ModificationDate,
		Servers:               getSGResponse.SecurityGroup.Servers,
	}

	// Override the values that need to be updated
	if req.Name != nil {
		setRequest.Name = *req.Name
	}
	if req.Description != nil {
		setRequest.Description = *req.Description
	}
	if req.InboundDefaultPolicy != nil {
		setRequest.InboundDefaultPolicy = *req.InboundDefaultPolicy
	}
	if req.OutboundDefaultPolicy != nil {
		setRequest.OutboundDefaultPolicy = *req.OutboundDefaultPolicy
	}
	if req.Stateful != nil {
		setRequest.Stateful = *req.Stateful
	}
	if req.OrganizationDefault != nil {
		setRequest.OrganizationDefault = req.OrganizationDefault
	}
	if req.ProjectDefault != nil {
		setRequest.ProjectDefault = *req.ProjectDefault
	}
	if req.EnableDefaultSecurity != nil {
		setRequest.EnableDefaultSecurity = *req.EnableDefaultSecurity
	}
	if req.Tags != nil {
		setRequest.Tags = req.Tags
	}

	setRes, err := s.setSecurityGroup(setRequest, opts...)
	if err != nil {
		return nil, err
	}

	return &UpdateSecurityGroupResponse{
		SecurityGroup: setRes.SecurityGroup,
	}, nil
}

// UpdateSecurityGroupRuleRequest contains the parameters to update a security group rule
type UpdateSecurityGroupRuleRequest struct {
	Zone                scw.Zone `json:"-"`
	SecurityGroupID     string   `json:"-"`
	SecurityGroupRuleID string   `json:"-"`

	Protocol  *SecurityGroupRuleProtocol  `json:"protocol"`
	Direction *SecurityGroupRuleDirection `json:"direction"`
	Action    *SecurityGroupRuleAction    `json:"action"`
	IPRange   *scw.IPNet                  `json:"ip_range"`
	Position  *uint32                     `json:"position"`

	// If set to 0, DestPortFrom will be removed.
	// See SecurityGroupRule.DestPortFrom for more information
	DestPortFrom *uint32 `json:"dest_port_from"`

	// If set to 0, DestPortTo will be removed.
	// See SecurityGroupRule.DestPortTo for more information
	DestPortTo *uint32 `json:"dest_port_to"`
}

type UpdateSecurityGroupRuleResponse struct {
	Rule *SecurityGroupRule `json:"security_rule"`
}

// UpdateSecurityGroupRule updates a security group.
func (s *API) UpdateSecurityGroupRule(req *UpdateSecurityGroupRuleRequest, opts ...scw.RequestOption) (*UpdateSecurityGroupRuleResponse, error) {
	var err error

	if fmt.Sprint(req.Zone) == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	res, err := s.GetSecurityGroupRule(&GetSecurityGroupRuleRequest{
		SecurityGroupRuleID: req.SecurityGroupRuleID,
		SecurityGroupID:     req.SecurityGroupID,
		Zone:                req.Zone,
	})
	if err != nil {
		return nil, err
	}

	setRequest := &setSecurityGroupRuleRequest{
		Zone:                req.Zone,
		SecurityGroupID:     req.SecurityGroupID,
		SecurityGroupRuleID: req.SecurityGroupRuleID,
		ID:                  req.SecurityGroupRuleID,
		Direction:           res.Rule.Direction,
		Protocol:            res.Rule.Protocol,
		DestPortFrom:        res.Rule.DestPortFrom,
		DestPortTo:          res.Rule.DestPortTo,
		IPRange:             res.Rule.IPRange,
		Action:              res.Rule.Action,
		Position:            res.Rule.Position,
		Editable:            res.Rule.Editable,
	}

	// Override the values that need to be updated
	if req.Action != nil {
		setRequest.Action = *req.Action
	}
	if req.IPRange != nil {
		setRequest.IPRange = *req.IPRange
	}
	if req.DestPortTo != nil {
		if *req.DestPortTo > 0 {
			setRequest.DestPortTo = req.DestPortTo
		} else {
			setRequest.DestPortTo = nil
		}
	}
	if req.DestPortFrom != nil {
		if *req.DestPortFrom > 0 {
			setRequest.DestPortFrom = req.DestPortFrom
		} else {
			setRequest.DestPortFrom = nil
		}
	}
	if req.DestPortFrom != nil && req.DestPortTo != nil && *req.DestPortFrom == *req.DestPortTo {
		setRequest.DestPortTo = nil
	}
	if req.Protocol != nil {
		setRequest.Protocol = *req.Protocol
	}
	if req.Direction != nil {
		setRequest.Direction = *req.Direction
	}
	if req.Position != nil {
		setRequest.Position = *req.Position
	}

	// When we use ICMP protocol portFrom and portTo should be set to nil
	if req.Protocol != nil && *req.Protocol == SecurityGroupRuleProtocolICMP {
		setRequest.DestPortFrom = nil
		setRequest.DestPortTo = nil
	}

	resp, err := s.setSecurityGroupRule(setRequest)
	if err != nil {
		return nil, err
	}

	return &UpdateSecurityGroupRuleResponse{
		Rule: resp.Rule,
	}, nil
}
