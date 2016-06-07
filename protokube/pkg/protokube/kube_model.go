package protokube

import (
	"fmt"
	"github.com/golang/glog"
)

// ApplyModel applies the configuration as specified in the model
func (k *KubeBoot) ApplyModel() error {
	modelDir := "model/etcd"

	etcdClusters, err := k.BuildEtcdClusters(modelDir)
	if err != nil {
		return fmt.Errorf("error building etcd models: %v", err)
	}

	for _, etcdCluster := range etcdClusters {
		glog.Infof("configuring etcd cluster %s", etcdCluster.ClusterName)
		err := etcdCluster.configure(k)
		if err != nil {
			return fmt.Errorf("error applying etcd model: %v", err)
		}
	}

	return nil
}
