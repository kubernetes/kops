package channels

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3"
	"k8s.io/kubernetes/pkg/util/validation/field"
	"net/url"
)

type Addon struct {
	Name            string
	ChannelName     string
	ChannelLocation url.URL
	Spec            *api.AddonSpec
}

type AddonUpdate struct {
	Name            string
	ExistingVersion *ChannelVersion
	NewVersion      *ChannelVersion
}

func (a *Addon) ChannelVersion() *ChannelVersion {
	return &ChannelVersion{
		Channel: &a.ChannelName,
		Version: a.Spec.Version,
	}
}

func (a *Addon) buildChannel() *Channel {
	namespace := "kube-system"
	if a.Spec.Namespace != nil {
		namespace = *a.Spec.Namespace
	}

	channel := &Channel{
		Namespace: namespace,
		Name:      a.Name,
	}
	return channel
}
func (a *Addon) GetRequiredUpdates(k8sClient *release_1_3.Clientset) (*AddonUpdate, error) {
	newVersion := a.ChannelVersion()

	channel := a.buildChannel()

	existingVersion, err := channel.GetInstalledVersion(k8sClient)
	if err != nil {
		return nil, err
	}

	if existingVersion != nil && !newVersion.Replaces(existingVersion) {
		return nil, nil
	}

	return &AddonUpdate{
		Name:            a.Name,
		ExistingVersion: existingVersion,
		NewVersion:      newVersion,
	}, nil
}

func (a *Addon) EnsureUpdated(k8sClient *release_1_3.Clientset) (*AddonUpdate, error) {
	required, err := a.GetRequiredUpdates(k8sClient)
	if err != nil {
		return nil, err
	}
	if required == nil {
		return nil, nil
	}

	if a.Spec.Manifest == nil || *a.Spec.Manifest == "" {
		return nil, field.Required(field.NewPath("Spec", "Manifest"), "")
	}

	manifest := *a.Spec.Manifest
	manifestURL, err := url.Parse(manifest)
	if err != nil {
		return nil, field.Invalid(field.NewPath("Spec", "Manifest"), manifest, "Not a valid URL")
	}
	if !manifestURL.IsAbs() {
		manifestURL = a.ChannelLocation.ResolveReference(manifestURL)
	}
	glog.Infof("Applying update from %q", manifestURL)

	err = Apply(manifestURL.String())
	if err != nil {
		return nil, fmt.Errorf("error applying update from %q: %v", manifest, err)
	}

	channel := a.buildChannel()
	err = channel.SetInstalledVersion(k8sClient, a.ChannelVersion())
	if err != nil {
		return nil, fmt.Errorf("error applying annotation to to record addon installation: %v", err)
	}

	return required, nil
}
