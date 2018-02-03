package google

import (
	"fmt"
	"bytes"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/kube-deploy/cluster-api-gcp/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Long term, we should retrieve the current status by asking k8s, gce etc. for all the needed info.
// For now, it is stored in the matching CRD under an annotation. This is similar to
// the spec and status concept where the machine CRD is the instance spec and the annotation is the instance status.

const InstanceStatusAnnotationKey = "instance-status"

type instanceStatus *clusterv1.Machine

// Get the status of the instance identified by the given machine
func (gce *GCEClient) instanceStatus(machine *clusterv1.Machine) (instanceStatus, error) {
	currentMachine, err := util.GetCurrentMachineIfExists(gce.machineClient, machine)
	if err != nil {
		return nil, err
	}


	if currentMachine == nil {
		// The current status no longer exists because the matching CRD has been deleted (or does not exist yet ie. bootstrapping)
		return nil, nil
	}
	return gce.machineInstanceStatus(currentMachine)
}

// Sets the status of the instance identified by the given machine to the given machine
func (gce *GCEClient) updateInstanceStatus(machine *clusterv1.Machine) error {
	status := instanceStatus(machine)
	currentMachine, err := util.GetCurrentMachineIfExists(gce.machineClient, machine)
	if err != nil {
		return err
	}

	if currentMachine == nil {
		// The current status no longer exists because the matching CRD has been deleted.
		return fmt.Errorf("Machine has already been deleted. Cannot update current instance status for machine %v", machine.ObjectMeta.Name)
	}

	m, err := gce.setMachineInstanceStatus(currentMachine, status)
	if err != nil {
		return err
	}

	_, err = gce.machineClient.Update(m)
	return err
}

// Gets the state of the instance stored on the given machine CRD
func (gce *GCEClient) machineInstanceStatus(machine *clusterv1.Machine) (instanceStatus, error) {
	if machine.ObjectMeta.Annotations == nil {
		// No state
		return nil, nil
	}

	a := machine.ObjectMeta.Annotations[InstanceStatusAnnotationKey]
	if a == "" {
		// No state
		return nil, nil
	}

	serializer := json.NewSerializer(json.DefaultMetaFactory, gce.scheme, gce.scheme, false)
	var status clusterv1.Machine
	_, _, err := serializer.Decode([]byte(a), &schema.GroupVersionKind{Group:"", Version:"cluster.k8s.io/v1alpha1", Kind:"Machine"}, &status)
	if err != nil {
		return nil, fmt.Errorf("decoding failure: %v", err)
	}

	return instanceStatus(&status), nil
}

// Applies the state of an instance onto a given machine CRD
func (gce *GCEClient) setMachineInstanceStatus(machine *clusterv1.Machine, status instanceStatus)  (*clusterv1.Machine, error)  {
	// Avoid status within status within status ...
	status.ObjectMeta.Annotations[InstanceStatusAnnotationKey] = ""

	serializer := json.NewSerializer(json.DefaultMetaFactory, gce.scheme, gce.scheme, false)
	b := []byte{}
	buff := bytes.NewBuffer(b)
	err := serializer.Encode((*clusterv1.Machine)(status), buff)
	if err != nil {
		return nil, fmt.Errorf("encoding failure: %v", err)
	}

	if machine.ObjectMeta.Annotations == nil {
		machine.ObjectMeta.Annotations = make(map[string]string)
	}
	machine.ObjectMeta.Annotations[InstanceStatusAnnotationKey] = buff.String()
	return machine, nil
}
