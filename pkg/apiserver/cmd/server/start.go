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
	"fmt"
	"io"
	"net"

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
	//      return nil, err
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

	server, err := config.Complete().New()
	if err != nil {
		return err
	}
	return server.GenericAPIServer.PrepareRun().Run(wait.NeverStop)

}

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
