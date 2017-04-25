package proxy

import (
	"net/http"
)

type inspectContainerInterceptor struct{ proxy *Proxy }

func (i *inspectContainerInterceptor) InterceptRequest(r *http.Request) error {
	return nil
}

func (i *inspectContainerInterceptor) InterceptResponse(r *http.Response) error {
	if !i.proxy.RewriteInspect || r.StatusCode != 200 {
		return nil
	}

	container := jsonObject{}
	if err := unmarshalResponseBody(r, &container); err != nil {
		return err
	}

	if err := i.proxy.updateContainerNetworkSettings(container); err != nil {
		Log.Warningf("Inspecting container %s failed: %s", container["Id"], err)
	}

	return marshalResponseBody(r, container)
}
