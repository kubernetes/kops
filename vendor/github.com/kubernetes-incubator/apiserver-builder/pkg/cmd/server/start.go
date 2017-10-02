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
	"flag"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kubernetes-incubator/apiserver-builder/pkg/apiserver"
	"github.com/kubernetes-incubator/apiserver-builder/pkg/builders"
	"k8s.io/apimachinery/pkg/runtime/schema"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"

	"bytes"
	"net/http"
	"os"
    "path/filepath"
    "time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/apiserver-builder/pkg/validators"
	"k8s.io/apimachinery/pkg/openapi"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var GetOpenApiDefinition openapi.GetOpenAPIDefinitions

type ServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	APIBuilders        []*builders.APIGroupBuilder

	PrintBearerToken bool
	PrintOpenapi     bool
	RunDelegatedAuth bool
	BearerToken      string
	Kubeconfig       string
	PostStartHooks   []PostStartHook
}

type PostStartHook struct {
	Fn   genericapiserver.PostStartHookFunc
	Name string
}

// StartApiServer starts an apiserver hosting the provider apis and openapi definitions.
func StartApiServer(etcdPath string, apis []*builders.APIGroupBuilder, openapidefs openapi.GetOpenAPIDefinitions, title, version string) {
	logs.InitLogs()
	defer logs.FlushLogs()

	GetOpenApiDefinition = openapidefs

	// To disable providers, manually specify the list provided by getKnownProviders()
	cmd, _ := NewCommandStartServer(etcdPath, os.Stdout, os.Stderr, apis, wait.NeverStop, title, version)
	if logflag := flag.CommandLine.Lookup("v"); logflag != nil {
		level := logflag.Value.(*glog.Level)
		levelPtr := (*int32)(level)
		cmd.Flags().Int32Var(levelPtr, "loglevel", 0, "Set the level of log output")
	}
	cmd.Flags().AddFlagSet(pflag.CommandLine)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func NewServerOptions(etcdPath string, out, errOut io.Writer, b []*builders.APIGroupBuilder) *ServerOptions {
	versions := []schema.GroupVersion{}
	for _, b := range b {
		versions = append(versions, b.GetLegacyCodec()...)
	}

	o := &ServerOptions{
		RecommendedOptions: genericoptions.NewRecommendedOptions(etcdPath, builders.Scheme, builders.Codecs.LegacyCodec(versions...)),
		APIBuilders:        b,
		RunDelegatedAuth:   true,
	}
	o.RecommendedOptions.SecureServing.BindPort = 443

	return o
}

// NewCommandStartMaster provides a CLI handler for 'start master' command
func NewCommandStartServer(etcdPath string, out, errOut io.Writer, builders []*builders.APIGroupBuilder,
	stopCh <-chan struct{}, title, version string) (*cobra.Command, *ServerOptions) {
	o := NewServerOptions(etcdPath, out, errOut, builders)

	// Support overrides
	cmd := &cobra.Command{
		Short: "Launch an API server",
		Long:  "Launch an API server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.RunServer(stopCh, title, version); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&o.PrintBearerToken, "print-bearer-token", false,
		"Print a curl command with the bearer token to test the server")
	flags.BoolVar(&o.PrintOpenapi, "print-openapi", false,
		"Print the openapi json and exit")
	flags.BoolVar(&o.RunDelegatedAuth, "delegated-auth", true,
		"Setup delegated auth")
	flags.StringVar(&o.Kubeconfig, "kubeconfig", "", "Kubeconfig of apiserver to talk to.")
	o.RecommendedOptions.AddFlags(flags)
	return cmd, o
}

func (o ServerOptions) Validate(args []string) error {
	return nil
}

func (o *ServerOptions) Complete() error {
	return nil
}

func applyOptions(config *genericapiserver.Config, applyTo ...func(*genericapiserver.Config) error) error {
	for _, fn := range applyTo {
		if err := fn(config); err != nil {
			return err
		}
	}
	return nil
}

func (o ServerOptions) Config() (*apiserver.Config, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts(
		"localhost", nil, nil); err != nil {

		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewConfig(builders.Codecs)

	err := applyOptions(
		serverConfig,
		o.RecommendedOptions.Etcd.ApplyTo,
		o.RecommendedOptions.SecureServing.ApplyTo,
		o.RecommendedOptions.Audit.ApplyTo,
		o.RecommendedOptions.Features.ApplyTo,
	)
	if err != nil {
		return nil, err
	}

	if serverConfig.SharedInformerFactory == nil && len(o.Kubeconfig) > 0 {
		path, _ := filepath.Abs(o.Kubeconfig)
		glog.Infof("Creating shared informer factory from kubeconfig %s", path)
		config, err := clientcmd.BuildConfigFromFlags("", o.Kubeconfig)
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			glog.Errorf("Couldn't create clientset due to %v. SharedInformerFactory will not be set.", err)
			return nil, err
		}
		serverConfig.SharedInformerFactory = informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	}

	if o.RunDelegatedAuth {
		err := applyOptions(
			serverConfig,
			o.RecommendedOptions.Authentication.ApplyTo,
			o.RecommendedOptions.Authorization.ApplyTo,
		)
		if err != nil {
			return nil, err
		}
	}

	config := &apiserver.Config{GenericConfig: serverConfig}
	return config, nil
}

func (o *ServerOptions) RunServer(stopCh <-chan struct{}, title, version string) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	if o.PrintBearerToken {
		glog.Infof("Serving on loopback...")
		glog.Infof("\n\n********************************\nTo test the server run:\n"+
			"curl -k -H \"Authorization: Bearer %s\" %s\n********************************\n\n",
			config.GenericConfig.LoopbackClientConfig.BearerToken,
			config.GenericConfig.LoopbackClientConfig.Host)
	}
	o.BearerToken = config.GenericConfig.LoopbackClientConfig.BearerToken

	for _, provider := range o.APIBuilders {
		config.AddApi(provider)
	}

	config.Init()

	config.GenericConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(GetOpenApiDefinition, builders.Scheme)
	config.GenericConfig.OpenAPIConfig.Info.Title = title
	config.GenericConfig.OpenAPIConfig.Info.Version = version

	server, err := config.Complete().New()
	if err != nil {
		return err
	}

	for _, h := range o.PostStartHooks {
		server.GenericAPIServer.AddPostStartHook(h.Name, h.Fn)
	}

	s := server.GenericAPIServer.PrepareRun()
	err = validators.OpenAPI.SetSchema(readOpenapi(server.GenericAPIServer.Handler))
	if o.PrintOpenapi {
		fmt.Printf("%s", validators.OpenAPI.OpenApi)
		os.Exit(0)
	}
	if err != nil {
		return err
	}

	s.Run(stopCh)

	return nil
}

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
