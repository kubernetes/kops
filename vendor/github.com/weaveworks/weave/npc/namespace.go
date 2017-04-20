package npc

import (
	"encoding/json"

	"github.com/coreos/go-iptables/iptables"
	"k8s.io/client-go/pkg/api/unversioned"
	coreapi "k8s.io/client-go/pkg/api/v1"
	extnapi "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/types"
	"k8s.io/client-go/pkg/util/uuid"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/npc/ipset"
)

type ns struct {
	ipt *iptables.IPTables // interface to iptables
	ips ipset.Interface    // interface to ipset

	name      string                               // k8s Namespace name
	namespace *coreapi.Namespace                   // k8s Namespace object
	pods      map[types.UID]*coreapi.Pod           // k8s Pod objects by UID
	policies  map[types.UID]*extnapi.NetworkPolicy // k8s NetworkPolicy objects by UID

	uid     types.UID     // surrogate UID to own allPods selector
	allPods *selectorSpec // hash:ip ipset of all pod IPs in this namespace

	nsSelectors  *selectorSet
	podSelectors *selectorSet
	rules        *ruleSet
}

func newNS(name string, ipt *iptables.IPTables, ips ipset.Interface, nsSelectors *selectorSet) (*ns, error) {
	allPods, err := newSelectorSpec(&unversioned.LabelSelector{}, name, ipset.HashIP)
	if err != nil {
		return nil, err
	}

	ns := &ns{
		ipt:         ipt,
		ips:         ips,
		name:        name,
		pods:        make(map[types.UID]*coreapi.Pod),
		policies:    make(map[types.UID]*extnapi.NetworkPolicy),
		uid:         uuid.NewUUID(),
		allPods:     allPods,
		nsSelectors: nsSelectors,
		rules:       newRuleSet(ipt)}

	ns.podSelectors = newSelectorSet(ips, ns.onNewPodSelector)

	if err := ns.podSelectors.provision(ns.uid, nil, map[string]*selectorSpec{ns.allPods.key: ns.allPods}); err != nil {
		return nil, err
	}

	return ns, nil
}

func (ns *ns) empty() bool {
	return len(ns.pods) == 0 && len(ns.policies) == 0 && ns.namespace == nil
}

func (ns *ns) destroy() error {
	if err := ns.podSelectors.deprovision(ns.uid, map[string]*selectorSpec{ns.allPods.key: ns.allPods}, nil); err != nil {
		return err
	}
	return nil
}

func (ns *ns) onNewPodSelector(selector *selector) error {
	for _, pod := range ns.pods {
		if hasIP(pod) {
			if selector.matches(pod.ObjectMeta.Labels) {
				if err := selector.addEntry(pod.Status.PodIP); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ns *ns) addPod(obj *coreapi.Pod) error {
	ns.pods[obj.ObjectMeta.UID] = obj

	if !hasIP(obj) {
		return nil
	}

	return ns.podSelectors.addToMatching(obj.ObjectMeta.Labels, obj.Status.PodIP)
}

func (ns *ns) updatePod(oldObj, newObj *coreapi.Pod) error {
	delete(ns.pods, oldObj.ObjectMeta.UID)
	ns.pods[newObj.ObjectMeta.UID] = newObj

	if !hasIP(oldObj) && !hasIP(newObj) {
		return nil
	}

	if hasIP(oldObj) && !hasIP(newObj) {
		return ns.podSelectors.delFromMatching(oldObj.ObjectMeta.Labels, oldObj.Status.PodIP)
	}

	if !hasIP(oldObj) && hasIP(newObj) {
		return ns.podSelectors.addToMatching(newObj.ObjectMeta.Labels, newObj.Status.PodIP)
	}

	if !equals(oldObj.ObjectMeta.Labels, newObj.ObjectMeta.Labels) ||
		oldObj.Status.PodIP != newObj.Status.PodIP {

		for _, ps := range ns.podSelectors.entries {
			oldMatch := ps.matches(oldObj.ObjectMeta.Labels)
			newMatch := ps.matches(newObj.ObjectMeta.Labels)
			if oldMatch == newMatch && oldObj.Status.PodIP == newObj.Status.PodIP {
				continue
			}
			if oldMatch {
				if err := ps.delEntry(oldObj.Status.PodIP); err != nil {
					return err
				}
			}
			if newMatch {
				if err := ps.addEntry(newObj.Status.PodIP); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (ns *ns) deletePod(obj *coreapi.Pod) error {
	delete(ns.pods, obj.ObjectMeta.UID)

	if !hasIP(obj) {
		return nil
	}

	return ns.podSelectors.delFromMatching(obj.ObjectMeta.Labels, obj.Status.PodIP)
}

func (ns *ns) addNetworkPolicy(obj *extnapi.NetworkPolicy) error {
	ns.policies[obj.ObjectMeta.UID] = obj

	// Analyse policy, determine which rules and ipsets are required
	rules, nsSelectors, podSelectors, err := ns.analysePolicy(obj)
	if err != nil {
		return err
	}

	// Provision required resources in dependency order
	if err := ns.nsSelectors.provision(obj.ObjectMeta.UID, nil, nsSelectors); err != nil {
		return err
	}
	if err := ns.podSelectors.provision(obj.ObjectMeta.UID, nil, podSelectors); err != nil {
		return err
	}
	if err := ns.rules.provision(obj.ObjectMeta.UID, nil, rules); err != nil {
		return err
	}

	return nil
}

func (ns *ns) updateNetworkPolicy(oldObj, newObj *extnapi.NetworkPolicy) error {
	delete(ns.policies, oldObj.ObjectMeta.UID)
	ns.policies[newObj.ObjectMeta.UID] = newObj

	// Analyse the old and the new policy so we can determine differences
	oldRules, oldNsSelectors, oldPodSelectors, err := ns.analysePolicy(oldObj)
	if err != nil {
		return err
	}
	newRules, newNsSelectors, newPodSelectors, err := ns.analysePolicy(newObj)
	if err != nil {
		return err
	}

	// Deprovision unused and provision newly required resources in dependency order
	if err := ns.rules.deprovision(oldObj.ObjectMeta.UID, oldRules, newRules); err != nil {
		return err
	}
	if err := ns.nsSelectors.deprovision(oldObj.ObjectMeta.UID, oldNsSelectors, newNsSelectors); err != nil {
		return err
	}
	if err := ns.podSelectors.deprovision(oldObj.ObjectMeta.UID, oldPodSelectors, newPodSelectors); err != nil {
		return err
	}
	if err := ns.nsSelectors.provision(oldObj.ObjectMeta.UID, oldNsSelectors, newNsSelectors); err != nil {
		return err
	}
	if err := ns.podSelectors.provision(oldObj.ObjectMeta.UID, oldPodSelectors, newPodSelectors); err != nil {
		return err
	}
	if err := ns.rules.provision(oldObj.ObjectMeta.UID, oldRules, newRules); err != nil {
		return err
	}

	return nil
}

func (ns *ns) deleteNetworkPolicy(obj *extnapi.NetworkPolicy) error {
	delete(ns.policies, obj.ObjectMeta.UID)

	// Analyse network policy to free resources
	rules, nsSelectors, podSelectors, err := ns.analysePolicy(obj)
	if err != nil {
		return err
	}

	// Deprovision unused resources in dependency order
	if err := ns.rules.deprovision(obj.ObjectMeta.UID, rules, nil); err != nil {
		return err
	}
	if err := ns.nsSelectors.deprovision(obj.ObjectMeta.UID, nsSelectors, nil); err != nil {
		return err
	}
	if err := ns.podSelectors.deprovision(obj.ObjectMeta.UID, podSelectors, nil); err != nil {
		return err
	}

	return nil
}

func bypassRule(nsIpsetName ipset.Name) []string {
	return []string{"-m", "set", "--match-set", string(nsIpsetName), "dst", "-j", "ACCEPT"}
}

func (ns *ns) ensureBypassRule(nsIpsetName ipset.Name) error {
	return ns.ipt.Append(TableFilter, DefaultChain, bypassRule(ns.allPods.ipsetName)...)
}

func (ns *ns) deleteBypassRule(nsIpsetName ipset.Name) error {
	return ns.ipt.Delete(TableFilter, DefaultChain, bypassRule(ns.allPods.ipsetName)...)
}

func (ns *ns) addNamespace(obj *coreapi.Namespace) error {
	ns.namespace = obj

	// Insert a rule to bypass policies if namespace is DefaultAllow
	if !isDefaultDeny(obj) {
		return ns.ensureBypassRule(ns.allPods.ipsetName)
	}

	// Add namespace ipset to matching namespace selectors
	return ns.nsSelectors.addToMatching(obj.ObjectMeta.Labels, string(ns.allPods.ipsetName))
}

func (ns *ns) updateNamespace(oldObj, newObj *coreapi.Namespace) error {
	ns.namespace = newObj

	// Update bypass rule if ingress default has changed
	oldDefaultDeny := isDefaultDeny(oldObj)
	newDefaultDeny := isDefaultDeny(newObj)

	if oldDefaultDeny != newDefaultDeny {
		if oldDefaultDeny {
			return ns.ensureBypassRule(ns.allPods.ipsetName)
		}
		if newDefaultDeny {
			return ns.deleteBypassRule(ns.allPods.ipsetName)
		}
	}

	// Re-evaluate namespace selector membership if labels have changed
	if !equals(oldObj.ObjectMeta.Labels, newObj.ObjectMeta.Labels) {
		for _, selector := range ns.nsSelectors.entries {
			oldMatch := selector.matches(oldObj.ObjectMeta.Labels)
			newMatch := selector.matches(newObj.ObjectMeta.Labels)
			if oldMatch == newMatch {
				continue
			}
			if oldMatch {
				if err := selector.delEntry(string(ns.allPods.ipsetName)); err != nil {
					return err
				}
			}
			if newMatch {
				if err := selector.addEntry(string(ns.allPods.ipsetName)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (ns *ns) deleteNamespace(obj *coreapi.Namespace) error {
	ns.namespace = nil

	// Remove bypass rule
	if !isDefaultDeny(obj) {
		return ns.deleteBypassRule(ns.allPods.ipsetName)
	}

	// Remove namespace ipset from any matching namespace selectors
	return ns.nsSelectors.delFromMatching(obj.ObjectMeta.Labels, string(ns.allPods.ipsetName))
}

func hasIP(pod *coreapi.Pod) bool {
	// Ensure pod isn't dead, has an IP address and isn't sharing the host network namespace
	return pod.Status.Phase != "Succeeded" && pod.Status.Phase != "Failed" &&
		len(pod.Status.PodIP) > 0 && !pod.Spec.HostNetwork
}

func equals(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for ak, av := range a {
		if b[ak] != av {
			return false
		}
	}
	return true
}

func isDefaultDeny(namespace *coreapi.Namespace) bool {
	nnpJSON, found := namespace.ObjectMeta.Annotations["net.beta.kubernetes.io/network-policy"]
	if !found {
		return false
	}

	var nnp NamespaceNetworkPolicy
	if err := json.Unmarshal([]byte(nnpJSON), &nnp); err != nil {
		common.Log.Warn("Ignoring network policy annotation: unmarshal failed:", err)
		// If we can't understand the annotation, behave as if it isn't present
		return false
	}

	return nnp.Ingress != nil &&
		nnp.Ingress.Isolation != nil &&
		*(nnp.Ingress.Isolation) == DefaultDeny
}
