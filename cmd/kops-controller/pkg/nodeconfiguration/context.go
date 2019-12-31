package nodeconfiguration

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc/peer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	kopsv1alpha2 "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type grpcContext struct {
	Context context.Context
	Client  client.Client

	machine           *capiMachine
	machineSet        *capiMachineSet
	machineDeployment *capiMachineDeployment

	instanceGroup *kopsv1alpha2.InstanceGroup
}

func (c *grpcContext) FindMachine() (*capiMachine, error) {
	if c.machine != nil {
		return c.machine, nil
	}

	// TODO: Should the node prove its identify / help us identify it?
	peerInfo, ok := peer.FromContext(c.Context)
	if !ok {
		return nil, fmt.Errorf("failed to get peer info from client")
	}

	ns := "kube-system"

	peerAddress, _, err := net.SplitHostPort(peerInfo.Addr.String())
	if err != nil {
		klog.Warningf("failed to parse peer address %q", peerInfo.Addr.String())
		return nil, fmt.Errorf("failed to identify node")
	}

	// TODO: Use an index to make this fast
	machineList := &unstructured.UnstructuredList{}
	machineList.SetAPIVersion("cluster.x-k8s.io/v1alpha3")
	machineList.SetKind("MachineList")
	if err := c.Client.List(c.Context, machineList, client.InNamespace(ns)); err != nil {
		return nil, fmt.Errorf("failed to list machines: %v", err)
	}
	var matches []*capiMachine
	for _, u := range machineList.Items {
		m := &capiMachine{genericWrapper{u}}
		for _, address := range m.Addresses() {
			// TODO: Check type?
			if address.Address() == peerAddress {
				matches = append(matches, m)
			}
		}
	}

	if len(matches) == 0 {
		return nil, nil
	}

	if len(matches) == 1 {
		// TODO: Additional checks?  Check that this is the right network etc?
		c.machine = matches[0]

		return matches[0], nil
	}

	return nil, fmt.Errorf("found multiple matching machines")
}

func (c *grpcContext) FindMachineSet() (*capiMachineSet, error) {
	m, err := c.FindMachine()
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	ref, err := m.FindOwnerReference("cluster.x-k8s.io", "MachineSet")
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, nil
	}

	machineSet := &unstructured.Unstructured{}
	machineSet.SetAPIVersion("cluster.x-k8s.io/v1alpha3")
	machineSet.SetKind("MachineSet")

	key := types.NamespacedName{
		Namespace: m.Namespace(),
		Name:      ref.Name,
	}
	if err := c.Client.Get(c.Context, key, machineSet); err != nil {
		return nil, fmt.Errorf("failed to get machineset %v: %v", key, err)
	}

	return &capiMachineSet{genericWrapper{*machineSet}}, nil
}

func (c *grpcContext) FindMachineDeployment() (*capiMachineDeployment, error) {
	ms, err := c.FindMachineSet()
	if err != nil {
		return nil, err
	}
	if ms == nil {
		return nil, nil
	}

	ref, err := ms.FindOwnerReference("cluster.x-k8s.io", "MachineDeployment")
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, nil
	}

	machineDeployment := &unstructured.Unstructured{}
	machineDeployment.SetAPIVersion("cluster.x-k8s.io/v1alpha3")
	machineDeployment.SetKind("MachineDeployment")

	key := types.NamespacedName{
		Namespace: ms.Namespace(),
		Name:      ref.Name,
	}
	if err := c.Client.Get(c.Context, key, machineDeployment); err != nil {
		return nil, fmt.Errorf("failed to get machinedeployment %v: %v", key, err)
	}

	return &capiMachineDeployment{genericWrapper{*machineDeployment}}, nil
}

func (c *grpcContext) FindInstanceGroup() (*kopsv1alpha2.InstanceGroup, error) {
	md, err := c.FindMachineDeployment()
	if err != nil {
		return nil, err
	}
	if md == nil {
		return nil, nil
	}

	ref, err := md.FindOwnerReference("kops.k8s.io", "InstanceGroup")
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, nil
	}

	ig := &kopsv1alpha2.InstanceGroup{}

	key := types.NamespacedName{
		Namespace: md.Namespace(),
		Name:      ref.Name,
	}
	if err := c.Client.Get(c.Context, key, ig); err != nil {
		return nil, fmt.Errorf("failed to get instanceGroup %v: %v", key, err)
	}

	return ig, nil
}
