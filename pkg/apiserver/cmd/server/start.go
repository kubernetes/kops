/*
Copyright 2016 The Kubernetes Authors.

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

package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	openapinamer "k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/registry/generic"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/apiserver"
	"k8s.io/kops/pkg/openapi"
)

const defaultEtcdPathPrefix = "/registry/kops.kubernetes.io"

var processInfo genericoptions.ProcessInfo

type KopsServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions

	StdOut io.Writer
	StdErr io.Writer

	PrintOpenapi bool
}

// NewCommandStartKopsServer provides a CLI handler for 'start master' command
func NewCommandStartKopsServer(out, err io.Writer) *cobra.Command {
	o := &KopsServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix,
			apiserver.Codecs.LegacyCodec(v1alpha2.SchemeGroupVersion), &processInfo),

		StdOut: out,
		StdErr: err,
	}
	//o.RecommendedOptions.Etcd.StorageConfig.Type = storagebackend.StorageTypeETCD2
	o.RecommendedOptions.Etcd.StorageConfig.Codec = apiserver.Codecs.LegacyCodec(v1alpha2.SchemeGroupVersion)
	//o.SecureServing.ServingOptions.BindPort = 443

	cmd := &cobra.Command{
		Short: "Launch a kops API server",
		Long:  "Launch a kops API server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunKopsServer(); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	o.RecommendedOptions.AddFlags(flags)

	flags.BoolVar(&o.PrintOpenapi, "print-openapi", false,
		"Print the openapi json and exit")

	return cmd
}

func (o KopsServerOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

func (o *KopsServerOptions) Complete() error {
	return nil
}

func (o KopsServerOptions) Config() (*apiserver.Config, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	scheme := apiserver.Scheme
	config := genericapiserver.NewRecommendedConfig(apiserver.Codecs)

	// We have to skip some of these to get docs to work...
	// if err := o.RecommendedOptions.ApplyTo(config, scheme); err != nil {
	// 	return nil, err
	// }

	if err := o.RecommendedOptions.Etcd.ApplyTo(&config.Config); err != nil {
		return nil, err
	}
	if err := o.RecommendedOptions.SecureServing.ApplyTo(&config.Config.SecureServing, &config.Config.LoopbackClientConfig); err != nil {
		return nil, err
	}

	if !o.PrintOpenapi {
		if err := o.RecommendedOptions.Authentication.ApplyTo(&config.Config.Authentication, config.SecureServing, config.OpenAPIConfig); err != nil {
			return nil, err
		}
		if err := o.RecommendedOptions.Authorization.ApplyTo(&config.Config.Authorization); err != nil {
			return nil, err
		}

		klog.Warningf("Authentication/Authorization disabled")
	}

	//if err := o.RecommendedOptions.Audit.ApplyTo(&config.Config); err != nil {
	//	return nil, err
	//}
	if err := o.RecommendedOptions.Features.ApplyTo(&config.Config); err != nil {
		return nil, err
	}

	if !o.PrintOpenapi {
		if err := o.RecommendedOptions.CoreAPI.ApplyTo(config); err != nil {
			return nil, err
		}

		if initializers, err := o.RecommendedOptions.ExtraAdmissionInitializers(config); err != nil {
			return nil, err
		} else if err := o.RecommendedOptions.Admission.ApplyTo(&config.Config, config.SharedInformerFactory, config.ClientConfig, scheme, initializers...); err != nil {
			return nil, err
		}
	}

	// serverConfig.CorsAllowedOriginList = []string{".*"}

	return &apiserver.Config{
		GenericConfig: config,
		ExtraConfig:   apiserver.ExtraConfig{},
	}, nil
}

func (o KopsServerOptions) RunKopsServer() error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	// Configure the openapi spec provided on /swagger.json
	// TODO: Come up with a better title and a meaningful version

	config.GenericConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(openapi.GetOpenAPIDefinitions, openapinamer.NewDefinitionNamer(apiserver.Scheme))
	config.GenericConfig.OpenAPIConfig.Info.Title = "Kops API"
	config.GenericConfig.OpenAPIConfig.Info.Version = "0.1"

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	//server.GenericAPIServer.AddPostStartHook("start-sample-server-informers", func(context genericapiserver.PostStartHookContext) error {
	//	config.GenericConfig.SharedInformerFactory.Start(context.StopCh)
	//	return nil
	//})

	// Just print the openapi spec and exit.  This is useful for
	// updating the published openapi and generating documentation.
	if o.PrintOpenapi {
		fmt.Printf("%s", readOpenapi(server.GenericAPIServer.Handler))
		os.Exit(0)
	}

	srv := server.GenericAPIServer.PrepareRun()
	return srv.Run(wait.NeverStop)
}

// Read the openapi spec from the http request handler.
func readOpenapi(handler *genericapiserver.APIServerHandler) string {
	req, err := http.NewRequest("GET", "/swagger.json", nil)
	if err != nil {
		panic(fmt.Errorf("Could not create openapi request %v", err))
	}
	resp := &BufferedResponse{}
	handler.ServeHTTP(resp, req)
	return resp.String()
}

type BufferedResponse struct {
	bytes.Buffer
}

func (BufferedResponse) Header() http.Header { return http.Header{} }
func (BufferedResponse) WriteHeader(int)     {}

type restOptionsFactory struct {
	storageConfig *storagebackend.Config
}

func (f *restOptionsFactory) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	ro := generic.RESTOptions{
		StorageConfig:           f.storageConfig,
		Decorator:               generic.UndecoratedStorage,
		DeleteCollectionWorkers: 1,
		EnableGarbageCollection: false,
		ResourcePrefix:          f.storageConfig.Prefix + "/" + resource.Group + "/" + resource.Resource,
	}

	//if f.Options.EnableWatchCache {
	//	ro.Decorator = registry.StorageWithCacher(f.Options.DefaultWatchCacheSize)
	//}

	return ro, nil
}
