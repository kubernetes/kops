/*
Copyright 2019 The Kubernetes Authors.

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

package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops/util"
	"k8s.io/kops/pkg/apis/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands/commandutils"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/util/pkg/text"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	replaceLong = templates.LongDesc(i18n.T(`
		Replace a resource desired configuration by filename or stdin.`))

	replaceExample = templates.Examples(i18n.T(`
		# Replace a cluster desired configuration using a YAML file
		kops replace -f my-cluster.yaml

		# Replace an instancegroup using YAML passed into stdin.
		cat instancegroup.yaml | kops replace -f -

		# Note, if the resource does not exist the command will error, use --force to provision resource
		kops replace -f my-cluster.yaml --force
		`))

	replaceShort = i18n.T(`Replace cluster resources.`)
)

// ReplaceOptions is the options for the command
type ReplaceOptions struct {
	// Filenames is a list of files containing resources to replace.
	Filenames []string
	// Force causes any missing rescources to be created.
	Force bool

	// ClusterName can be specified.  If specified all objects must match the given cluster.
	ClusterName string
}

// NewCmdReplace returns a new replace command
func NewCmdReplace(f *util.Factory, out io.Writer) *cobra.Command {
	options := &ReplaceOptions{}

	cmd := &cobra.Command{
		Use:               "replace {-f FILENAME}...",
		Short:             replaceShort,
		Long:              replaceLong,
		Example:           replaceExample,
		Args:              rootCommand.clusterNameArgs(&options.ClusterName),
		ValidArgsFunction: commandutils.CompleteClusterName(f, true, false),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flag("name").Changed {
				// Don't infer the cluster name from the kubeconfig,
				// that both breaks compatibility and is pretty dangerous.
				options.ClusterName = ""
			}
			return RunReplace(cmd.Context(), f, out, options)
		},
	}
	cmd.Flags().StringSliceVarP(&options.Filenames, "filename", "f", options.Filenames, "A list of one or more files separated by a comma.")
	cmd.MarkFlagRequired("filename")
	cmd.Flags().BoolVarP(&options.Force, "force", "", false, "Force any changes, which will also create any non-existing resource")

	return cmd
}

// RunReplace processes the replace command
func RunReplace(ctx context.Context, f *util.Factory, out io.Writer, opt *ReplaceOptions) error {
	clientset, err := f.KopsClient()
	if err != nil {
		return err
	}

	var objects []runtime.Object
	for _, filename := range opt.Filenames {
		var contents []byte
		if filename == "-" {
			contents, err = ConsumeStdin()
			if err != nil {
				return err
			}
		} else {
			contents, err = vfs.Context.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("error reading file %q: %v", filename, err)
			}
		}
		sections := text.SplitContentToSections(contents)

		for _, section := range sections {
			o, _, err := kopscodecs.Decode(section, nil)
			if err != nil {
				return fmt.Errorf("error parsing file %q: %v", filename, err)
			}

			objects = append(objects, o)
		}
	}

	for _, obj := range objects {
		gvk := obj.GetObjectKind().GroupVersionKind()

		clusterName := ""

		switch v := obj.(type) {
		case *kopsapi.Cluster:
			clusterName = v.Name
		case *kopsapi.InstanceGroup:
			clusterName = v.ObjectMeta.Labels[kopsapi.LabelClusterName]
		case *kopsapi.SSHCredential:
			clusterName = v.ObjectMeta.Labels[kopsapi.LabelClusterName]
		case *unstructured.Unstructured:
			if !featureflag.ClusterAddons.Enabled() {
				klog.V(2).Infof("Type of object was %T", v)
				return fmt.Errorf("unhandled kind %v", gvk)
			}
			// To encourage use of the cluster-name flag, we don't recognize the label on additional objects.
			// This will let us move towards something better, likely spec.clusterName or spec.clusterRef.
			// clusterName = v.GetLabels()[kopsapi.LabelClusterName]

		default:
			klog.V(2).Infof("Type of object was %T", v)
			return fmt.Errorf("unhandled kind %v", gvk)
		}
		if clusterName == "" {
			clusterName = opt.ClusterName
		}
		if opt.ClusterName != "" && clusterName != opt.ClusterName {
			return fmt.Errorf("mismatch on cluster name: found %q but %q was specified", clusterName, opt.ClusterName)
		}
		if clusterName == "" {
			return fmt.Errorf("must specify the name of the cluster (or use the %q label)", kopsapi.LabelClusterName)
		}
	}

	changeset := clusterChangeSet{
		clientset: clientset,
	}
	for _, obj := range objects {
		switch v := obj.(type) {
		case *kopsapi.Cluster:
			{
				// Retrieve the current status of the cluster.  This will eventually be part of the cluster object.
				cloud, err := cloudup.BuildCloud(v)
				if err != nil {
					return err
				}
				status, err := cloud.FindClusterStatus(v)
				if err != nil {
					return err
				}

				// Check if the cluster exists already
				clusterName := v.Name
				if clusterName == "" {
					clusterName = opt.ClusterName
				}
				cluster, err := clientset.GetCluster(ctx, clusterName)
				if err != nil {
					if errors.IsNotFound(err) {
						cluster = nil
					} else {
						return fmt.Errorf("error fetching cluster %q: %v", clusterName, err)
					}
				}
				if cluster == nil {
					if !opt.Force {
						return fmt.Errorf("cluster %v does not exist (try adding --force flag)", clusterName)
					}
					_, err = clientset.CreateCluster(ctx, v)
					if err != nil {
						return fmt.Errorf("error creating cluster: %v", err)
					}
				} else {
					_, err = clientset.UpdateCluster(ctx, v, status)
					if err != nil {
						return fmt.Errorf("error replacing cluster: %v", err)
					}
				}
			}

		case *kopsapi.InstanceGroup:
			clusterName := v.ObjectMeta.Labels[kopsapi.LabelClusterName]
			if clusterName == "" {
				clusterName = opt.ClusterName
			}
			cluster, err := clientset.GetCluster(ctx, clusterName)
			if err != nil {
				if errors.IsNotFound(err) {
					return fmt.Errorf("cluster %q not found", clusterName)
				}
				return fmt.Errorf("error fetching cluster %q: %v", clusterName, err)
			}
			// check if the instancegroup exists already
			igName := v.ObjectMeta.Name
			ig, err := clientset.InstanceGroupsFor(cluster).Get(ctx, igName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					if !opt.Force {
						return fmt.Errorf("instanceGroup: %v does not exist (try adding --force flag)", igName)
					}
				} else {
					return fmt.Errorf("unable to check for instanceGroup: %v", err)
				}
			}
			switch ig {
			case nil:
				klog.Infof("instanceGroup: %v was not found, creating resource now", igName)
				_, err = clientset.InstanceGroupsFor(cluster).Create(ctx, v, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("error creating instanceGroup: %v", err)
				}
			default:
				_, err = clientset.InstanceGroupsFor(cluster).Update(ctx, v, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("error replacing instanceGroup: %v", err)
				}
			}
		case *kopsapi.SSHCredential:
			clusterName := v.ObjectMeta.Labels[kopsapi.LabelClusterName]
			if clusterName == "" {
				clusterName = opt.ClusterName
			}
			if v.Spec.PublicKey == "" {
				return fmt.Errorf("spec.PublicKey is required")
			}

			cluster, err := clientset.GetCluster(ctx, clusterName)
			if err != nil {
				return err
			}

			sshCredentialStore, err := clientset.SSHCredentialStore(cluster)
			if err != nil {
				return err
			}

			sshKeyArr := []byte(v.Spec.PublicKey)
			err = sshCredentialStore.AddSSHPublicKey(ctx, sshKeyArr)
			if err != nil {
				return fmt.Errorf("error replacing SSHCredential: %v", err)
			}
		case *unstructured.Unstructured:
			clusterName := v.GetLabels()[kopsapi.LabelClusterName]
			if clusterName == "" {
				clusterName = opt.ClusterName
			}
			if err := changeset.changesForCluster(clusterName).CreateOrUpdateAdditionalObject(ctx, v); err != nil {
				return err
			}
		default:
			gvk := v.GetObjectKind().GroupVersionKind()
			return fmt.Errorf("unhandled kind %v", gvk)
		}
	}

	if err := changeset.FlushAll(ctx); err != nil {
		return err
	}

	return nil
}

// clusterChangeSet buffers a set of changes to a set of clusters.
type clusterChangeSet struct {
	clientset simple.Clientset
	clusters  map[string]*clusterChange
}

// FlushAll writes all changes to the backing stores.
func (c *clusterChangeSet) FlushAll(ctx context.Context) error {
	for _, change := range c.clusters {
		if err := change.FlushAll(ctx); err != nil {
			return err
		}
	}
	return nil
}

// changesForCluster gets the changes that pertain to a particular cluster.
func (c *clusterChangeSet) changesForCluster(clusterName string) *clusterChange {
	if c.clusters == nil {
		c.clusters = make(map[string]*clusterChange)
	}

	change := c.clusters[clusterName]
	if change == nil {
		change = &clusterChange{
			clientset:   c.clientset,
			clusterName: clusterName,
		}
		c.clusters[clusterName] = change
	}
	return change
}

// clusterChange buffers changes to a single cluster
type clusterChange struct {
	clusterName string
	clientset   simple.Clientset

	cachedCluster *kops.Cluster

	addons        kubemanifest.ObjectList
	addonsChanged bool
}

// cluster gets the cluster object, caching it for future calls.
func (c *clusterChange) cluster(ctx context.Context) (*kops.Cluster, error) {
	if c.cachedCluster != nil {
		return c.cachedCluster, nil
	}
	cluster, err := c.clientset.GetCluster(ctx, c.clusterName)
	if err != nil {
		return nil, err
	}
	c.cachedCluster = cluster
	return cluster, nil
}

// FlushAll writes all changes to the backing stores.
func (c *clusterChange) FlushAll(ctx context.Context) error {
	if c.addonsChanged {
		cluster, err := c.cluster(ctx)
		if err != nil {
			return err
		}

		if err := c.clientset.AddonsFor(cluster).Replace(c.addons); err != nil {
			return err
		}
	}
	return nil
}

// CreateOrUpdateAdditionalObject updates the matching addon object if found, or adds it if not.
func (c *clusterChange) CreateOrUpdateAdditionalObject(ctx context.Context, u *unstructured.Unstructured) error {
	if c.addons == nil {
		cluster, err := c.cluster(ctx)
		if err != nil {
			return err
		}
		addons, err := c.clientset.AddonsFor(cluster).List()
		if err != nil {
			return err
		}
		c.addons = addons
	}

	gvk := u.GroupVersionKind()

	apiVersion := gvk.GroupVersion().Identifier()
	kind := gvk.Kind
	name := u.GetName()
	namespace := u.GetNamespace()

	id := name
	if namespace != "" {
		id = namespace + "/" + id
	}

	obj := kubemanifest.NewObject(u.Object)

	found := false
	for i, addon := range c.addons {
		if addon.APIVersion() == apiVersion && addon.Kind() == kind && addon.GetName() == name && addon.GetNamespace() == namespace {
			klog.Infof("replacing object %v %s", gvk, id)
			found = true
			c.addons[i] = obj
			c.addonsChanged = true
			break
		}
	}
	if !found {
		klog.Infof("adding object %v %s", gvk, id)
		c.addons = append(c.addons, obj)
		c.addonsChanged = true
	}

	return nil
}
