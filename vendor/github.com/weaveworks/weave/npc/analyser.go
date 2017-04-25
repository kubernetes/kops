package npc

import (
	"fmt"

	"k8s.io/client-go/pkg/api"
	extnapi "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"

	"github.com/weaveworks/weave/npc/ipset"
)

func (ns *ns) analysePolicy(policy *extnapi.NetworkPolicy) (
	rules map[string]*ruleSpec,
	nsSelectors, podSelectors map[string]*selectorSpec,
	err error) {

	nsSelectors = make(map[string]*selectorSpec)
	podSelectors = make(map[string]*selectorSpec)
	rules = make(map[string]*ruleSpec)

	dstSelector, err := newSelectorSpec(&policy.Spec.PodSelector, ns.name, ipset.HashIP)
	if err != nil {
		return nil, nil, nil, err
	}
	podSelectors[dstSelector.key] = dstSelector

	for _, ingressRule := range policy.Spec.Ingress {
		if ingressRule.Ports != nil && len(ingressRule.Ports) == 0 {
			// Ports is empty, this rule matches no ports (no traffic matches).
			continue
		}

		if ingressRule.From != nil && len(ingressRule.From) == 0 {
			// From is empty, this rule matches no sources (no traffic matches).
			continue
		}

		if ingressRule.From == nil {
			// From is not provided, this rule matches all sources (traffic not restricted by source).
			if ingressRule.Ports == nil {
				// Ports is not provided, this rule matches all ports (traffic not restricted by port).
				rule := newRuleSpec(nil, nil, dstSelector, nil)
				rules[rule.key] = rule
			} else {
				// Ports is present and contains at least one item, then this rule allows traffic
				// only if the traffic matches at least one port in the ports list.
				withNormalisedProtoAndPort(ingressRule.Ports, func(proto, port string) {
					rule := newRuleSpec(&proto, nil, dstSelector, &port)
					rules[rule.key] = rule
				})
			}
		} else {
			// From is present and contains at least on item, this rule allows traffic only if the
			// traffic matches at least one item in the from list.
			for _, peer := range ingressRule.From {
				var srcSelector *selectorSpec
				if peer.PodSelector != nil {
					srcSelector, err = newSelectorSpec(peer.PodSelector, ns.name, ipset.HashIP)
					if err != nil {
						return nil, nil, nil, err
					}
					podSelectors[srcSelector.key] = srcSelector
				}
				if peer.NamespaceSelector != nil {
					srcSelector, err = newSelectorSpec(peer.NamespaceSelector, "", ipset.ListSet)
					if err != nil {
						return nil, nil, nil, err
					}
					nsSelectors[srcSelector.key] = srcSelector
				}

				if ingressRule.Ports == nil {
					// Ports is not provided, this rule matches all ports (traffic not restricted by port).
					rule := newRuleSpec(nil, srcSelector, dstSelector, nil)
					rules[rule.key] = rule
				} else {
					// Ports is present and contains at least one item, then this rule allows traffic
					// only if the traffic matches at least one port in the ports list.
					withNormalisedProtoAndPort(ingressRule.Ports, func(proto, port string) {
						rule := newRuleSpec(&proto, srcSelector, dstSelector, &port)
						rules[rule.key] = rule
					})
				}
			}
		}
	}

	return rules, nsSelectors, podSelectors, nil
}

func withNormalisedProtoAndPort(npps []extnapi.NetworkPolicyPort, f func(proto, port string)) {
	for _, npp := range npps {
		// If no proto is specified, default to TCP
		proto := string(api.ProtocolTCP)
		if npp.Protocol != nil {
			proto = string(*npp.Protocol)
		}

		// If no port is specified, match any port. Let iptables executable handle
		// service name resolution
		port := "0:65535"
		if npp.Port != nil {
			switch npp.Port.Type {
			case intstr.Int:
				port = fmt.Sprintf("%d", npp.Port.IntVal)
			case intstr.String:
				port = npp.Port.StrVal
			}
		}

		f(proto, port)
	}
}
