package proxy

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

const MaxDockerHostname = 64

var (
	ErrNoCommandSpecified = errors.New("No command specified")
)

type createContainerInterceptor struct{ proxy *Proxy }

// ErrNoSuchImage replaces docker.NoSuchImage, which does not contain the image
// name, which in turn breaks docker clients post 1.7.0 since they expect the
// image name to be present in errors.
type ErrNoSuchImage struct {
	Name string
}

func (err *ErrNoSuchImage) Error() string {
	return "No such image: " + err.Name
}

func (i *createContainerInterceptor) InterceptRequest(r *http.Request) error {
	container := jsonObject{}
	if err := unmarshalRequestBody(r, &container); err != nil {
		return err
	}

	hostConfig, err := container.Object("HostConfig")
	if err != nil {
		return err
	}

	networkMode, err := hostConfig.String("NetworkMode")
	if err != nil {
		return err
	}

	env, err := container.StringArray("Env")
	if err != nil {
		return err
	}

	if cidrs, err := i.proxy.weaveCIDRs(networkMode, env); err != nil {
		Log.Infof("Leaving container alone because %s", err)
	} else {
		Log.Infof("Creating container with WEAVE_CIDR \"%s\"", strings.Join(cidrs, " "))
		if i.proxy.NoMulticastRoute {
			if err := addVolume(hostConfig, i.proxy.weaveWaitNomcastVolume, "/w", "ro"); err != nil {
				return err
			}
		} else {
			if err := addVolume(hostConfig, i.proxy.weaveWaitVolume, "/w", "ro"); err != nil {
				return err
			}
		}
		if err := i.setWeaveWaitEntrypoint(container); err != nil {
			return err
		}
		hostname, err := i.containerHostname(r, container)
		if err != nil {
			return err
		}
		if dnsDomain := i.proxy.getDNSDomain(); dnsDomain != "" {
			if err := i.setHostname(container, hostname, dnsDomain); err != nil {
				return err
			}
			if err := i.proxy.setWeaveDNS(hostConfig, hostname, dnsDomain); err != nil {
				return err
			}
		}

		return marshalRequestBody(r, container)
	}

	return nil
}

func (i *createContainerInterceptor) setWeaveWaitEntrypoint(container jsonObject) error {
	var entrypoint []string
	entrypoint, err := container.StringArray("Entrypoint")
	if err != nil {
		return err
	}

	cmd, err := container.StringArray("Cmd")
	if err != nil {
		return err
	}

	if len(entrypoint) == 0 {
		containerImage, err := container.String("Image")
		if err != nil {
			return err
		}

		image, err := i.proxy.client.InspectImage(containerImage)
		if err == docker.ErrNoSuchImage {
			return &ErrNoSuchImage{containerImage}
		} else if err != nil {
			return err
		}

		if len(cmd) == 0 {
			cmd = image.Config.Cmd
			container["Cmd"] = cmd
		}

		if entrypoint == nil {
			entrypoint = image.Config.Entrypoint
			container["Entrypoint"] = entrypoint
		}
	}

	if len(entrypoint) == 0 && len(cmd) == 0 {
		return ErrNoCommandSpecified
	}

	if len(entrypoint) == 0 || entrypoint[0] != weaveWaitEntrypoint[0] {
		container["Entrypoint"] = append(weaveWaitEntrypoint, entrypoint...)
	}

	return nil
}

func (i *createContainerInterceptor) setHostname(container jsonObject, name, dnsDomain string) error {
	hostname, err := container.String("Hostname")
	if err != nil {
		return err
	}
	if hostname == "" && name != "" {
		// Strip trailing period because it's unusual to see it used on the end of a host name
		trimmedDNSDomain := strings.TrimSuffix(dnsDomain, ".")
		if len(name)+1+len(trimmedDNSDomain) > MaxDockerHostname {
			Log.Warningf("Container name [%s] too long to be used as hostname", name)
		} else {
			container["Hostname"] = name
			container["Domainname"] = trimmedDNSDomain
		}
	}

	return nil
}

func (i *createContainerInterceptor) InterceptResponse(r *http.Response) error {
	return nil
}

func (i *createContainerInterceptor) containerHostname(r *http.Request, container jsonObject) (hostname string, err error) {
	hostname = r.URL.Query().Get("name")
	if i.proxy.Config.HostnameFromLabel != "" {
		hostname, err = i.hostnameFromLabel(hostname, container)
	}
	hostname = i.proxy.hostnameMatchRegexp.ReplaceAllString(hostname, i.proxy.HostnameReplacement)
	return
}

func (i *createContainerInterceptor) hostnameFromLabel(hostname string, container jsonObject) (string, error) {
	labels, err := container.Object("Labels")
	if err != nil {
		return "", err
	}
	label, err := labels.String(i.proxy.Config.HostnameFromLabel)
	if err != nil {
		return "", err
	}
	if label == "" {
		return hostname, nil
	}

	return label, nil
}
