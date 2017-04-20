package proxy

import (
	"net/http"
	"strings"
)

type startContainerInterceptor struct{ proxy *Proxy }

func (i *startContainerInterceptor) InterceptRequest(r *http.Request) error {
	container, err := inspectContainerInPath(i.proxy.client, r.URL.Path)
	if err != nil {
		return err
	}

	// If the client has sent some JSON which might be a HostConfig, add our
	// parameters back into it, otherwise Docker will consider them overwritten
	if containerShouldAttach(container) && r.Header.Get("Content-Type") == "application/json" && r.ContentLength > 0 {
		params := map[string]interface{}{}
		if err := unmarshalRequestBody(r, &params); err != nil {
			return err
		}
		// HostConfig can be sent either as a struct named HostConfig, or unnamed at top level
		var hostConfig map[string]interface{}
		if subParam, found := params["HostConfig"]; found {
			if typecast, ok := subParam.(map[string]interface{}); ok {
				hostConfig = typecast
			}
			// We can't reliably detect what Docker will see as a top-level HostConfig,
			// so just assume any parameter at top level indicates there is one.
		} else if len(params) > 0 {
			hostConfig = params
		}
		if hostConfig != nil {
			networkMode, err := jsonObject(hostConfig).String("NetworkMode")
			if err != nil {
				return err
			}
			if strings.HasPrefix(networkMode, "container:") || networkMode == "host" {
				if err := addVolume(hostConfig, i.proxy.weaveWaitNoopVolume, "/w", "ro"); err != nil {
					return err
				}
			} else {
				if i.proxy.NoMulticastRoute {
					if err := addVolume(hostConfig, i.proxy.weaveWaitNomcastVolume, "/w", "ro"); err != nil {
						return err
					}
				} else {
					if err := addVolume(hostConfig, i.proxy.weaveWaitVolume, "/w", "ro"); err != nil {
						return err
					}
				}
				if dnsDomain := i.proxy.getDNSDomain(); dnsDomain != "" {
					if err := i.proxy.setWeaveDNS(hostConfig, container.Config.Hostname, dnsDomain); err != nil {
						return err
					}
				}
			}

			// Note we marshal the original top-level dictionary to avoid disturbing anything else
			if err := marshalRequestBody(r, params); err != nil {
				return err
			}
		}
	}
	i.proxy.createWait(r, container.ID)
	return nil
}

func (i *startContainerInterceptor) InterceptResponse(r *http.Response) error {
	defer i.proxy.removeWait(r.Request)
	if r.StatusCode < 200 || r.StatusCode >= 300 { // Docker didn't do the start
		return nil
	}
	return i.proxy.waitForStart(r.Request)
}
