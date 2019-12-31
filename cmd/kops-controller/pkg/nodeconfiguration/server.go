package nodeconfiguration

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/kopscodecs"
	pb "k8s.io/kops/pkg/proto/nodeconfiguration"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type nodeConfigurationService struct {
	client     client.Client
	configBase vfs.Path
}

func NewNodeConfigurationService(mgr manager.Manager, opt *config.Options) (*nodeConfigurationService, error) {
	s := &nodeConfigurationService{}

	s.client = mgr.GetClient()

	configBase, err := vfs.Context.BuildVfsPath(opt.ConfigBase)
	if err != nil {
		return nil, fmt.Errorf("cannot parse ConfigBase %q: %v", opt.ConfigBase, err)
	}
	s.configBase = configBase

	return s, nil
}

var _ pb.NodeConfigurationServiceServer = &nodeConfigurationService{}

func (s *nodeConfigurationService) GetKeypair(ctx context.Context, request *pb.GetKeypairRequest) (*pb.GetKeypairResponse, error) {
	klog.V(2).Infof("grpc request: GetKeypair %v", request)

	errForbidden := status.New(codes.Unauthenticated, "forbidden").Err()

	c := &grpcContext{
		Context: ctx,
		Client:  s.client,
	}

	ig, err := c.FindInstanceGroup()
	if err != nil {
		klog.Warningf("failed to map to instancegroup: %v", err)
		return nil, errForbidden
	}
	if ig == nil {
		klog.Warningf("did not find instancegroup")
		return nil, errForbidden
	}

	switch request.Name {
	case "kubelet", "kube-proxy":
		// Nodes can load this key

	default:
		klog.Warningf("forbidden key %q", request.Name)
		return nil, errForbidden
	}

	// TODO: For now we always load from VFS
	cluster := &kops.Cluster{}
	{
		p := s.configBase.Join(registry.PathClusterCompleted)

		b, err := p.ReadFile()
		if err != nil {
			klog.Warningf("error loading Cluster %q: %v", p, err)
			return nil, errForbidden
		}

		if err := utils.YamlUnmarshal(b, cluster); err != nil {
			klog.Warningf("error parsing Cluster %q: %v", p, err)
			return nil, errForbidden
		}
	}

	var ks fi.CAStore
	{
		p, err := vfs.Context.BuildVfsPath(cluster.Spec.KeyStore)
		if err != nil {
			klog.Warningf("error building key store path: %v", err)
			return nil, errForbidden
		}

		allowList := false // We shouldn't need to list
		ks = fi.NewVFSCAStore(cluster, p, allowList)
	}

	cert, key, _, err := ks.FindKeypair(request.Name)
	if err != nil {
		klog.Warningf("error loading keypair %q: %v", request.Name, err)
		return nil, errForbidden
	}

	response := &pb.GetKeypairResponse{}

	if cert != nil {
		s, err := cert.AsString()
		if err != nil {
			klog.Warningf("error serializing cert: %v", err)
			return nil, errForbidden
		}
		response.Cert = s
	}

	if key != nil {
		s, err := key.AsString()
		if err != nil {
			klog.Warningf("error serializing key: %v", err)
			return nil, errForbidden
		}
		response.Key = s
	}

	klog.Infof("response: %v", response)

	return response, nil
}

func (s *nodeConfigurationService) GetConfiguration(ctx context.Context, request *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error) {
	klog.V(2).Infof("grpc request: GetConfiguration %v", request)

	errForbidden := status.New(codes.Unauthenticated, "forbidden").Err()

	c := &grpcContext{
		Context: ctx,
		Client:  s.client,
	}

	ig, err := c.FindInstanceGroup()
	if err != nil {
		klog.Warningf("failed to map to instancegroup: %v", err)
		return nil, errForbidden
	}
	if ig == nil {
		klog.Warningf("did not find instancegroup")
		return nil, errForbidden
	}

	response := &pb.GetConfigurationResponse{}

	// Note: For now, we're assuming there is only a single cluster, and it is ours.
	// We therefore use the configured base path

	// Today we load the full cluster config from the state store (e.g. S3) every time
	// TODO: we should generate it on the fly (to allow for cluster reconfiguration)
	{
		p := s.configBase.Join(registry.PathClusterCompleted)

		b, err := p.ReadFile()
		if err != nil {
			klog.Warningf("error loading Cluster %q: %v", p, err)
			return nil, errForbidden
		}
		response.ClusterFullConfig = string(b)
	}

	{
		b, err := kopscodecs.ToVersionedYaml(ig)
		if err != nil {
			klog.Warningf("error encoding instancegroup: %v", err)
			return nil, errForbidden
		}

		response.InstanceGroupConfig = string(b)
	}

	{
		cluster := &kops.Cluster{}

		if err := utils.YamlUnmarshal([]byte(response.ClusterFullConfig), cluster); err != nil {
			klog.Warningf("error parsing Cluster: %v", err)
			return nil, errForbidden
		}

		var ks fi.CAStore
		{
			p, err := vfs.Context.BuildVfsPath(cluster.Spec.KeyStore)
			if err != nil {
				klog.Warningf("error building key store path: %v", err)
				return nil, errForbidden
			}

			allowList := false // We shouldn't need to list
			ks = fi.NewVFSCAStore(cluster, p, allowList)
		}

		cert, err := ks.FindCert(fi.CertificateId_CA)
		if err != nil {
			klog.Warningf("error loading CA cert: %v", err)
			return nil, errForbidden
		}

		s, err := cert.AsString()
		if err != nil {
			klog.Warningf("error serializing cert: %v", err)
			return nil, errForbidden
		}

		response.CaCertificate = s
	}

	/*
		instanceGroupName := ig.Name

		// TODO: We also load the instancegroup from VFS, again meaning we aren't yet truly dynamic
		{
			p := s.configBase.Join("instancegroup", instanceGroupName)

			b, err := p.ReadFile()
			if err != nil {
				klog.Warningf("error loading InstanceGroup %q: %v", p, err)
				return nil, errForbidden
			}

			response.InstanceGroupConfig = string(b)
		}
	*/

	klog.Infof("response: %v", response)

	return response, nil
}
