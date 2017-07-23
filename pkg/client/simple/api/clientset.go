package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/api/errors"
)

// RESTClientset is an implementation of clientset that uses a "real" generated REST client
type RESTClientset struct {
	BaseURL    *url.URL
	KopsClient kopsinternalversion.KopsInterface
}

func (c *RESTClientset) ClustersFor(cluster *kops.Cluster) kopsinternalversion.ClusterInterface {
	namespace := restNamespaceForClusterName(cluster.Name)
	return c.KopsClient.Clusters(namespace)
}

func (c *RESTClientset) GetCluster(name string) (*kops.Cluster, error) {
	namespace := restNamespaceForClusterName(name)
	return c.KopsClient.Clusters(namespace).Get(name, metav1.GetOptions{})
}

func (c *RESTClientset) ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error) {
	if cluster.Spec.ConfigBase != "" {
		return vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	}
	// URL for clusters looks like  https://<server>/apis/kops/v1alpha2/namespaces/<cluster>/clusters/<cluster>
	// We probably want to add a subresource for full resources
	return vfs.Context.BuildVfsPath(c.BaseURL.String())
}

func (c *RESTClientset) ListClusters(options metav1.ListOptions) (*kops.ClusterList, error) {
	return c.KopsClient.Clusters(metav1.NamespaceAll).List(options)
}

func (c *RESTClientset) InstanceGroupsFor(cluster *kops.Cluster) kopsinternalversion.InstanceGroupInterface {
	namespace := restNamespaceForClusterName(cluster.Name)
	return c.KopsClient.InstanceGroups(namespace)
}

func (c *RESTClientset) FederationsFor(federation *kops.Federation) kopsinternalversion.FederationInterface {
	// Unsure if this should be namespaced or not - probably, so that we can RBAC it...
	panic("Federations are curently not supported by the server API")
	//namespace := restNamespaceForFederationName(federation.Name)
	//return c.KopsClient.Federations(namespace)
}

func (c *RESTClientset) ListFederations(options metav1.ListOptions) (*kops.FederationList, error) {
	return c.KopsClient.Federations(metav1.NamespaceAll).List(options)
}

func (c *RESTClientset) GetFederation(name string) (*kops.Federation, error) {
	namespace := restNamespaceForFederationName(name)
	return c.KopsClient.Federations(namespace).Get(name, metav1.GetOptions{})
}

func (c *RESTClientset) SecretStore(cluster *kops.Cluster) (fi.SecretStore, error) {
	namespace := restNamespaceForClusterName(cluster.Name)
	return secrets.NewClientsetSecretStore(c.KopsClient, namespace), nil
}

func (c *RESTClientset) KeyStore(cluster *kops.Cluster) (fi.CAStore, error) {
	namespace := restNamespaceForClusterName(cluster.Name)
	return fi.NewClientsetCAStore(c.KopsClient, namespace), nil
}

func (c *RESTClientset) DeleteCluster(cluster *kops.Cluster) error {
	configBase, err := registry.ConfigBase(cluster)
	if err != nil {
		return err
	}

	err = vfsclientset.DeleteAllClusterState(configBase)
	if err != nil {
		return err
	}

	name := cluster.Name
	namespace := restNamespaceForClusterName(name)

	{
		keysets, err := c.KopsClient.Keysets(namespace).List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing Keysets: %v", err)
		}

		for i := range keysets.Items {
			keyset := &keysets.Items[i]
			err = c.KopsClient.Keysets(namespace).Delete(keyset.Name, &metav1.DeleteOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					// Unlikely...
					glog.Warningf("Keyset was concurrently deleted")
				} else {
					return fmt.Errorf("error deleting Keyset %q: %v", keyset.Name, err)
				}
			}
		}
	}

	{
		igs, err := c.KopsClient.InstanceGroups(namespace).List(metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing instance groups: %v", err)
		}

		for i := range igs.Items {
			ig := &igs.Items[i]
			err = c.KopsClient.InstanceGroups(namespace).Delete(ig.Name, &metav1.DeleteOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					// Unlikely...
					glog.Warningf("instance group was concurrently deleted")
				} else {
					return fmt.Errorf("error deleting instance group %q: %v", ig.Name, err)
				}
			}
		}
	}

	err = c.KopsClient.Clusters(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Unlikely...
			glog.Warningf("cluster %q was concurrently deleted", name)
		} else {
			return fmt.Errorf("error deleting cluster%q: %v", name, err)
		}
	}

	return nil
}

func restNamespaceForClusterName(clusterName string) string {
	// We are not allowed dots, so we map them to dashes
	// This can conflict, but this will simply be a limitation that we pass on to the user
	// i.e. it will not be possible to create a.b.example.com and a-b.example.com
	namespace := strings.Replace(clusterName, ".", "-", -1)
	return namespace
}

func restNamespaceForFederationName(clusterName string) string {
	namespace := strings.Replace(clusterName, ".", "-", -1)
	return namespace
}
