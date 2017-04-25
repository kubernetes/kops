package npc

import (
	"sync"

	"github.com/coreos/go-iptables/iptables"
	"github.com/pkg/errors"
	coreapi "k8s.io/client-go/pkg/api/v1"
	extnapi "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/npc/ipset"
)

type NetworkPolicyController interface {
	AddNamespace(ns *coreapi.Namespace) error
	UpdateNamespace(oldObj, newObj *coreapi.Namespace) error
	DeleteNamespace(ns *coreapi.Namespace) error

	AddPod(obj *coreapi.Pod) error
	UpdatePod(oldObj, newObj *coreapi.Pod) error
	DeletePod(obj *coreapi.Pod) error

	AddNetworkPolicy(obj *extnapi.NetworkPolicy) error
	UpdateNetworkPolicy(oldObj, newObj *extnapi.NetworkPolicy) error
	DeleteNetworkPolicy(obj *extnapi.NetworkPolicy) error
}

type controller struct {
	sync.Mutex

	ipt *iptables.IPTables
	ips ipset.Interface

	nss         map[string]*ns // ns name -> ns struct
	nsSelectors *selectorSet   // selector string -> nsSelector
}

func New(ipt *iptables.IPTables, ips ipset.Interface) NetworkPolicyController {
	c := &controller{
		ipt: ipt,
		ips: ips,
		nss: make(map[string]*ns)}

	c.nsSelectors = newSelectorSet(ips, c.onNewNsSelector)

	return c
}

func (npc *controller) onNewNsSelector(selector *selector) error {
	for _, ns := range npc.nss {
		if ns.namespace != nil {
			if selector.matches(ns.namespace.ObjectMeta.Labels) {
				if err := ns.ips.AddEntry(selector.spec.ipsetName, string(ns.allPods.ipsetName)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (npc *controller) withNS(name string, f func(ns *ns) error) error {
	ns, found := npc.nss[name]
	if !found {
		newNs, err := newNS(name, npc.ipt, npc.ips, npc.nsSelectors)
		if err != nil {
			return err
		}
		npc.nss[name] = newNs
		ns = newNs
	}
	if err := f(ns); err != nil {
		return err
	}
	if ns.empty() {
		if err := ns.destroy(); err != nil {
			return err
		}
		delete(npc.nss, name)
	}
	return nil
}

func (npc *controller) AddPod(obj *coreapi.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Debugf("EVENT AddPod %s", js(obj))
	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.addPod(obj), "add pod")
	})
}

func (npc *controller) UpdatePod(oldObj, newObj *coreapi.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Debugf("EVENT UpdatePod %s %s", js(oldObj), js(newObj))
	return npc.withNS(oldObj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.updatePod(oldObj, newObj), "update pod")
	})
}

func (npc *controller) DeletePod(obj *coreapi.Pod) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Debugf("EVENT DeletePod %s", js(obj))
	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.deletePod(obj), "delete pod")
	})
}

func (npc *controller) AddNetworkPolicy(obj *extnapi.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Infof("EVENT AddNetworkPolicy %s", js(obj))
	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.addNetworkPolicy(obj), "add network policy")
	})
}

func (npc *controller) UpdateNetworkPolicy(oldObj, newObj *extnapi.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Infof("EVENT UpdateNetworkPolicy %s %s", js(oldObj), js(newObj))
	return npc.withNS(oldObj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.updateNetworkPolicy(oldObj, newObj), "update network policy")
	})
}

func (npc *controller) DeleteNetworkPolicy(obj *extnapi.NetworkPolicy) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Infof("EVENT DeleteNetworkPolicy %s", js(obj))
	return npc.withNS(obj.ObjectMeta.Namespace, func(ns *ns) error {
		return errors.Wrap(ns.deleteNetworkPolicy(obj), "delete network policy")
	})
}

func (npc *controller) AddNamespace(obj *coreapi.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Infof("EVENT AddNamespace %s", js(obj))
	return npc.withNS(obj.ObjectMeta.Name, func(ns *ns) error {
		return errors.Wrap(ns.addNamespace(obj), "add namespace")
	})
}

func (npc *controller) UpdateNamespace(oldObj, newObj *coreapi.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Infof("EVENT UpdateNamespace %s %s", js(oldObj), js(newObj))
	return npc.withNS(oldObj.ObjectMeta.Name, func(ns *ns) error {
		return errors.Wrap(ns.updateNamespace(oldObj, newObj), "update namespace")
	})
}

func (npc *controller) DeleteNamespace(obj *coreapi.Namespace) error {
	npc.Lock()
	defer npc.Unlock()

	common.Log.Infof("EVENT DeleteNamespace %s", js(obj))
	return npc.withNS(obj.ObjectMeta.Name, func(ns *ns) error {
		return errors.Wrap(ns.deleteNamespace(obj), "delete namespace")
	})
}
