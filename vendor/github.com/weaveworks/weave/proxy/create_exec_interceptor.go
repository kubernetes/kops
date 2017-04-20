package proxy

import (
	"net/http"
	"strings"
)

type createExecInterceptor struct{ proxy *Proxy }

func (i *createExecInterceptor) InterceptRequest(r *http.Request) error {
	options := jsonObject{}
	if err := unmarshalRequestBody(r, &options); err != nil {
		return err
	}

	container, err := inspectContainerInPath(i.proxy.client, r.URL.Path)
	if err != nil {
		return err
	}

	if _, hasWeaveWait := container.Volumes["/w"]; !hasWeaveWait {
		return nil
	}

	cidrs, err := i.proxy.weaveCIDRs(container.HostConfig.NetworkMode, container.Config.Env)
	if err != nil {
		Log.Infof("Leaving container %s alone because %s", container.ID, err)
		return nil
	}

	cmd, err := options.StringArray("Cmd")
	if err != nil {
		return err
	}

	Log.Infof("Exec in container %s with WEAVE_CIDR \"%s\"", container.ID, strings.Join(cidrs, " "))
	options["Cmd"] = append(weaveWaitEntrypoint, cmd...)

	return marshalRequestBody(r, options)
}

func (i *createExecInterceptor) InterceptResponse(r *http.Response) error {
	return nil
}
