package federation

import (
	"k8s.io/kubernetes/federation/client/clientset_generated/federation_release_1_4"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/api/errors"
	"fmt"
	"k8s.io/kubernetes/pkg/api/v1"
)

func findNamespace(k8s federation_release_1_4.Interface, name string) (*v1.Namespace, error) {
	glog.V(2).Infof("querying k8s for federation Namespace %s", name)
	c, err := k8s.Core().Namespaces().Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error reading federation Namespace %s: %v", name, err)
		}
	}
	return c, nil
}

func mutateNamespace(k8s federation_release_1_4.Interface, name string, fn func(s *v1.Namespace) (*v1.Namespace, error)) (*v1.Namespace, error) {
	existing, err := findNamespace(k8s, name)
	if err != nil {
		return nil, err
	}
	createObject := existing == nil
	updated, err := fn(existing)
	if err != nil {
		return nil, err
	}

	updated.Name = name

	if createObject {
		glog.V(2).Infof("creating federation Namespace %s", name)
		created, err := k8s.Core().Namespaces().Create(updated)
		if err != nil {
			return nil, fmt.Errorf("error creating federation Namespace %s: %v", name, err)
		}
		return created, nil
	} else {
		glog.V(2).Infof("updating federation Namespace %s", name)
		created, err := k8s.Core().Namespaces().Update(updated)
		if err != nil {
			return nil, fmt.Errorf("error updating federation Namespace %s: %v", name, err)
		}
		return created, nil
	}
}


