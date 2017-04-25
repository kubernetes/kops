package npc

import (
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"k8s.io/client-go/pkg/types"

	"github.com/weaveworks/weave/common"
)

type ruleSpec struct {
	key  string
	args []string
}

func newRuleSpec(proto *string, srcHost *selectorSpec, dstHost *selectorSpec, dstPort *string) *ruleSpec {
	args := []string{}
	if proto != nil {
		args = append(args, "-p", *proto)
	}
	if srcHost != nil {
		args = append(args, "-m", "set", "--match-set", string(srcHost.ipsetName), "src")
	}
	if dstHost != nil {
		args = append(args, "-m", "set", "--match-set", string(dstHost.ipsetName), "dst")
	}
	if dstPort != nil {
		args = append(args, "--dport", *dstPort)
	}
	args = append(args, "-j", "ACCEPT")
	key := strings.Join(args, " ")

	return &ruleSpec{key, args}
}

type ruleSet struct {
	ipt   *iptables.IPTables
	users map[string]map[types.UID]struct{}
}

func newRuleSet(ipt *iptables.IPTables) *ruleSet {
	return &ruleSet{ipt, make(map[string]map[types.UID]struct{})}
}

func (rs *ruleSet) deprovision(user types.UID, current, desired map[string]*ruleSpec) error {
	for key, spec := range current {
		if _, found := desired[key]; !found {
			delete(rs.users[key], user)
			if len(rs.users[key]) == 0 {
				common.Log.Infof("deleting rule: %v", spec.args)
				if err := rs.ipt.Delete(TableFilter, IngressChain, spec.args...); err != nil {
					return err
				}
				delete(rs.users, key)
			}
		}
	}

	return nil
}

func (rs *ruleSet) provision(user types.UID, current, desired map[string]*ruleSpec) error {
	for key, spec := range desired {
		if _, found := current[key]; !found {
			if _, found := rs.users[key]; !found {
				common.Log.Infof("adding rule: %v", spec.args)
				if err := rs.ipt.Append(TableFilter, IngressChain, spec.args...); err != nil {
					return err
				}
				rs.users[key] = make(map[types.UID]struct{})
			}
			rs.users[key][user] = struct{}{}
		}
	}

	return nil
}
