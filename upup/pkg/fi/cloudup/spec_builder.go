package cloudup

import (
	"github.com/golang/glog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
)

type SpecBuilder struct {
	OptionsLoader *loader.OptionsLoader

	Tags map[string]struct{}
}

func (l *SpecBuilder) BuildCompleteSpec(clusterSpec *api.ClusterSpec, modelStore vfs.Path, models []string) (*api.ClusterSpec, error) {
	// First pass over models: load options
	tw := &loader.TreeWalker{
		DefaultHandler: ignoreHandler,
		Contexts: map[string]loader.Handler{
			"resources": ignoreHandler,
		},
		Extensions: map[string]loader.Handler{
			".options": l.OptionsLoader.HandleOptions,
		},
		Tags: l.Tags,
	}
	for _, model := range models {
		modelDir := modelStore.Join(model)
		err := tw.Walk(modelDir)
		if err != nil {
			return nil, err
		}
	}

	loaded, err := l.OptionsLoader.Build(clusterSpec)
	if err != nil {
		return nil, err
	}
	completed := &api.ClusterSpec{}
	*completed = *(loaded.(*api.ClusterSpec))

	// Master kubelet config = (base kubelet config + master kubelet config)
	masterKubelet := &api.KubeletConfigSpec{}
	utils.JsonMergeStruct(masterKubelet, completed.Kubelet)
	utils.JsonMergeStruct(masterKubelet, completed.MasterKubelet)
	completed.MasterKubelet = masterKubelet

	glog.V(1).Infof("options: %s", fi.DebugAsJsonStringIndent(completed))
	return completed, nil
}
