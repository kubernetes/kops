/*
Copyright 2021 The Kubernetes Authors.

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

package metrics

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/cmd/kops-controller/pkg/metrics/collector"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
)

type Server struct {
	server *http.Server
}

func NewServer(opt *config.Options) (*Server, error) {
	if opt.Server.MetricsListen == "" {
		return nil, nil
	}
	server := &http.Server{
		Addr: opt.Server.MetricsListen,
	}

	s := &Server{
		server: server,
	}
	r := http.NewServeMux()
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	cluster, err := getCluster(opt.ConfigBase)
	if err != nil {
		return nil, err
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	client, err := getClient(opt.RegistryPath)
	if err != nil {
		return nil, err
	}

	c, err := collector.NewCollector(cluster, cloud, client, k8sClient)
	if err != nil {
		return nil, err
	}
	prometheus.MustRegister(c)
	r.Handle("/metrics", promhttp.Handler())

	server.Handler = recovery(r)

	return s, nil
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// recovery is responsible for ensuring we don't exit on a panic.
func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				klog.Errorf("failed to handle request: threw exception: %v: %s", err, debug.Stack())
			}
		}()

		next.ServeHTTP(w, req)
	})
}

func getCluster(configPath string) (*kops.Cluster, error) {
	configBase, err := vfs.Context.BuildVfsPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot parse ConfigBase %q: %v", configPath, err)
	}
	clusterPath := configBase.Join(registry.PathClusterCompleted)
	klog.Infof("cluster Path is :%#v", clusterPath)
	cluster, err := loadCluster(clusterPath)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func loadCluster(p vfs.Path) (*kops.Cluster, error) {
	b, err := p.ReadFile()
	if err != nil {
		return nil, fmt.Errorf("error loading Cluster %q: %v", p, err)
	}

	cluster := &kops.Cluster{}
	if err := utils.YamlUnmarshal(b, cluster); err != nil {
		return nil, fmt.Errorf("error parsing Cluster %q: %v", p, err)
	}

	return cluster, nil
}

func getClient(path string) (simple.Clientset, error) {
	basePath, err := vfs.Context.BuildVfsPath(path)
	if err != nil {
		return nil, err
	}
	return vfsclientset.NewVFSClientset(basePath), nil
}
