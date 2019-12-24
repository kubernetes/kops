package nodeconfiguration

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"
)

type genericWrapper struct {
	unstructured.Unstructured
}

func (o *genericWrapper) Namespace() string {
	return o.Unstructured.GetNamespace()
}

func (o *genericWrapper) Name() string {
	return o.Unstructured.GetName()
}

func (o *genericWrapper) FindOwnerReference(group string, kind string) (*metav1.OwnerReference, error) {
	var refs []metav1.OwnerReference
	for _, ref := range o.GetOwnerReferences() {
		if !strings.HasPrefix(ref.APIVersion, group+"/") {
			continue
		}
		if ref.Kind != kind {
			continue
		}
		refs = append(refs, ref)
	}

	if len(refs) == 0 {
		return nil, nil
	}

	if len(refs) != 1 {
		return nil, fmt.Errorf("found multiple matching ownerRefs")
	}

	ref := refs[0]
	return &ref, nil
}

type capiMachine struct {
	genericWrapper
}

type capiMachineSet struct {
	genericWrapper
}

type capiMachineDeployment struct {
	genericWrapper
}

func (m *capiMachine) Addresses() []*capiMachineAddress {
	o := m.Unstructured.Object
	status, ok := o["status"]
	if !ok {
		return nil
	}
	statusMap, ok := status.(map[string]interface{})
	if !ok {
		klog.Warningf("status was not of expected type: was %T", status)
		return nil
	}

	addresses, ok := statusMap["addresses"]
	if !ok {
		return nil
	}
	addressesList, ok := addresses.([]interface{})
	if !ok {
		klog.Warningf("addresses was not of expected type: was %T", addresses)
		return nil
	}

	var ret []*capiMachineAddress
	for _, a := range addressesList {
		address, ok := a.(map[string]interface{})
		if !ok {
			klog.Warningf("address was not of expected type: was %T", address)
			continue
		}

		ret = append(ret, &capiMachineAddress{address})
	}
	return ret
}

type capiMachineAddress struct {
	data map[string]interface{}
}

func (m *capiMachineAddress) Address() string {
	v, ok := m.data["address"]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		klog.Warningf("address was not of expected type: was %T", v)
		return ""
	}
	return s
}
