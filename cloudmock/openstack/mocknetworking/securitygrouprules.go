/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mocknetworking

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
)

type ruleListResponse struct {
	SecurityGroupRules []rules.SecGroupRule `json:"security_group_rules"`
}

type ruleGetResponse struct {
	SecurityGroupRule rules.SecGroupRule `json:"security_group_rule"`
}

type ruleCreateRequest struct {
	SecurityGroupRule rules.CreateOpts `json:"security_group_rule"`
}

func (m *MockClient) mockSecurityGroupRules() {
	re := regexp.MustCompile(`/security-group-rules/?`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		w.Header().Add("Content-Type", "application/json")

		sgrID := re.ReplaceAllString(r.URL.Path, "")
		switch r.Method {
		case http.MethodGet:
			if sgrID == "" {
				r.ParseForm()
				m.listSecurityGroupRules(w, r.Form)
			} else {
				m.getSecurityGroupRule(w, sgrID)
			}
		case http.MethodPost:
			m.createSecurityGroupRule(w, r)
		case http.MethodDelete:
			m.deleteSecurityGroupRule(w, sgrID)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	m.Mux.HandleFunc("/security-group-rules/", handler)
	m.Mux.HandleFunc("/security-group-rules", handler)
}

func (m *MockClient) listSecurityGroupRules(w http.ResponseWriter, vals url.Values) {

	w.WriteHeader(http.StatusOK)

	sgrs := filterRules(m.securityGroupRules, vals)

	resp := ruleListResponse{
		SecurityGroupRules: sgrs,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", resp))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}

func (m *MockClient) getSecurityGroupRule(w http.ResponseWriter, ruleID string) {
	if rule, ok := m.securityGroupRules[ruleID]; ok {
		resp := ruleGetResponse{
			SecurityGroupRule: rule,
		}
		respB, err := json.Marshal(resp)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal %+v", resp))
		}
		_, err = w.Write(respB)
		if err != nil {
			panic("failed to write body")
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) deleteSecurityGroupRule(w http.ResponseWriter, ruleID string) {
	if _, ok := m.securityGroupRules[ruleID]; ok {
		delete(m.securityGroupRules, ruleID)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockClient) createSecurityGroupRule(w http.ResponseWriter, r *http.Request) {
	var create ruleCreateRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		panic("error decoding create rule request")
	}

	w.WriteHeader(http.StatusAccepted)
	rule := rules.SecGroupRule{
		ID:             uuid.New().String(),
		PortRangeMax:   create.SecurityGroupRule.PortRangeMax,
		PortRangeMin:   create.SecurityGroupRule.PortRangeMin,
		Protocol:       string(create.SecurityGroupRule.Protocol),
		RemoteIPPrefix: create.SecurityGroupRule.RemoteIPPrefix,
		EtherType:      string(create.SecurityGroupRule.EtherType),
		RemoteGroupID:  create.SecurityGroupRule.RemoteGroupID,
		Direction:      string(create.SecurityGroupRule.Direction),
		SecGroupID:     create.SecurityGroupRule.SecGroupID,
	}
	m.securityGroupRules[rule.ID] = rule

	resp := ruleGetResponse{
		SecurityGroupRule: rule,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %+v", resp))
	}
	_, err = w.Write(respB)
	if err != nil {
		panic("failed to write body")
	}
}

func filterRules(allRules map[string]rules.SecGroupRule, vals url.Values) []rules.SecGroupRule {
	sgrs := make([]rules.SecGroupRule, 0)

	securityGroupIDFilter := vals.Get("security_group_id")
	directionFilter := vals.Get("direction")
	ethertypeFilter := vals.Get("ethertype")
	portRangeMaxFilter := vals.Get("port_range_max")
	portRangeMinFilter := vals.Get("port_range_min")
	protocolFilter := vals.Get("protocol")
	remoteGroupIDFilter := vals.Get("remote_group_id")

	// Example query string from cloudup
	// ?direction=ingress&ethertype=IPv4&port_range_max=53&port_range_min=53&protocol=udp&remote_group_id=3b39402b-320c-4e18-a2a1-11b5577a850f&security_group_id=df829d89-637c-4aff-8a46-716977f73464
	for _, s := range allRules {
		if securityGroupIDFilter != "" && s.SecGroupID != securityGroupIDFilter {
			continue
		}
		if directionFilter != "" && s.Direction != directionFilter {
			continue
		}
		if ethertypeFilter != "" && s.EtherType != ethertypeFilter {
			continue
		}
		if portRangeMaxFilter != "" {
			portRangeMax, err := strconv.ParseInt(portRangeMaxFilter, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("failed to parse port_range_max parameter %v", err))
			}
			if int64(s.PortRangeMax) != portRangeMax {
				continue
			}
		}
		if portRangeMinFilter != "" {
			portRangeMin, err := strconv.ParseInt(portRangeMinFilter, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("failed to parse port_range_max parameter %v", err))
			}
			if int64(s.PortRangeMin) != portRangeMin {
				continue
			}
		}
		if protocolFilter != "" && s.Protocol != protocolFilter {
			continue
		}
		// If a query doesn't provide remote_group_id this indicates we want to filter for rules
		// with an empty string value rather than not filter for remote_group_id
		if s.RemoteGroupID != remoteGroupIDFilter {
			continue
		}
		sgrs = append(sgrs, s)
	}
	return sgrs
}
