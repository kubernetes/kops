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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/registry/generic"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/apiserver"
	"k8s.io/kops/pkg/openapi"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const defaultEtcdPathPrefix = "/registry/kops.kubernetes.io"

type KopsServerOptions struct {
	Etcd          *genericoptions.EtcdOptions
	SecureServing *genericoptions.SecureServingOptions
	//InsecureServing *genericoptions.ServingOptions
	Authentication *genericoptions.DelegatingAuthenticationOptions
	Authorization  *genericoptions.DelegatingAuthorizationOptions

	StdOut io.Writer
	StdErr io.Writer

	PrintOpenapi bool
}

// NewCommandStartKopsServer provides a CLI handler for 'start master' command
func NewCommandStartKopsServer(out, err io.Writer) *cobra.Command {
	o := &KopsServerOptions{
		Etcd: genericoptions.NewEtcdOptions(&storagebackend.Config{
			Prefix: defaultEtcdPathPrefix,
			Copier: kops.Scheme,
			Codec:  nil,
		}),
		SecureServing: genericoptions.NewSecureServingOptions(),
		//InsecureServing: genericoptions.NewInsecureServingOptions(),
		Authentication: genericoptions.NewDelegatingAuthenticationOptions(),
		Authorization:  genericoptions.NewDelegatingAuthorizationOptions(),

		StdOut: out,
		StdErr: err,
	}
	o.Etcd.StorageConfig.Type = storagebackend.StorageTypeETCD2
	o.Etcd.StorageConfig.Codec = kops.Codecs.LegacyCodec(v1alpha2.SchemeGroupVersion)
	//o.SecureServing.ServingOptions.BindPort = 443

	cmd := &cobra.Command{
		Short: "Launch a kops API server",
		Long:  "Launch a kops API server",
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete())
			cmdutil.CheckErr(o.Validate(args))
			cmdutil.CheckErr(o.RunKopsServer())
		},
	}

	flags := cmd.Flags()
	o.Etcd.AddFlags(flags)
	o.SecureServing.AddFlags(flags)
	//o.InsecureServing.AddFlags(flags)
	o.Authentication.AddFlags(flags)
	o.Authorization.AddFlags(flags)

	flags.BoolVar(&o.PrintOpenapi, "print-openapi", false,
		"Print the openapi json and exit")

	return cmd
}

func (o KopsServerOptions) Validate(args []string) error {
	return nil
}

func (o *KopsServerOptions) Complete() error {
	return nil
}

func (o KopsServerOptions) RunKopsServer() error {
	// TODO have a "real" external address
	if err := o.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewConfig(kops.Codecs)
	// 1.6: serverConfig := genericapiserver.NewConfig().WithSerializer(kops.Codecs)
	//if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
	//	return nil, err
	//}

	serverConfig.CorsAllowedOriginList = []string{".*"}

	if err := o.Etcd.ApplyTo(serverConfig); err != nil {
		return err
	}

	if err := o.SecureServing.ApplyTo(serverConfig); err != nil {
		return err
	}
	//if err := o.InsecureServing.ApplyTo(serverConfig); err != nil {
	//      return err
	//}

	glog.Warningf("Authentication/Authorization disabled")

	//var err error
	//privilegedLoopbackToken := uuid.NewRandom().String()
	//if genericAPIServerConfig.LoopbackClientConfig, err = genericAPIServerConfig.SecureServingInfo.NewSelfClientConfig(privilegedLoopbackToken); err != nil {
	//               return err
	//       }

	config := apiserver.Config{
		GenericConfig:     serverConfig,
		RESTOptionsGetter: &restOptionsFactory{storageConfig: &o.Etcd.StorageConfig},
	}

	// Configure the openapi spec provided on /swagger.json
	// TODO: Come up with a better titlie and a meaningful version
	config.GenericConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(
		openapi.GetOpenAPIDefinitions, kops.Scheme)
	config.GenericConfig.OpenAPIConfig.Info.Title = "Kops API"
	config.GenericConfig.OpenAPIConfig.Info.Version = "0.1"

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	srv := server.GenericAPIServer.PrepareRun()

	// Just print the openapi spec and exit.  This is useful for
	// updating the published openapi and generating documentation.
	if o.PrintOpenapi {
		fmt.Printf("%s", readOpenapi(server.GenericAPIServer.Handler))
		os.Exit(0)
	}

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
