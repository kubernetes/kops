package cs

import (
	"testing"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
)

func TestClient_DescribeClusters(t *testing.T) {

	client := NewTestDebugAussumeRoleClient()

	clusters, err := client.DescribeClusters("")
	if err != nil {
		t.Fatalf("Failed to DescribeClusters: %v", err)
	}

	for _, cluster := range clusters {
		t.Logf("Cluster: %++v", cluster)
		c, err := client.DescribeCluster(cluster.ClusterID)
		if err != nil {
			t.Errorf("Failed to DescribeCluster: %v", err)
		}
		t.Logf("Cluster Describe: %++v", c)
		certs, err := client.GetClusterCerts(cluster.ClusterID)
		if err != nil {
			t.Errorf("Failed to GetClusterCerts: %v", err)
		}
		t.Logf("Cluster certs: %++v", certs)

	}
}

func TestListClusters(t *testing.T) {

	client := NewTestClientForDebug()

	clusters, err := client.DescribeClusters("")
	if err != nil {
		t.Fatalf("Failed to DescribeClusters: %v", err)
	}

	for _, cluster := range clusters {
		t.Logf("Cluster: %++v", cluster)
		c, err := client.DescribeCluster(cluster.ClusterID)
		if err != nil {
			t.Errorf("Failed to DescribeCluster: %v", err)
		}
		t.Logf("Cluster Describe: %++v", c)
		certs, err := client.GetClusterCerts(cluster.ClusterID)
		if err != nil {
			t.Errorf("Failed to GetClusterCerts: %v", err)
		}
		t.Logf("Cluster certs: %++v", certs)

	}
}

func _TestCreateClusters(t *testing.T) {

	client := NewTestClientForDebug()

	args := ClusterCreationArgs{
		Name:             "test",
		Size:             1,
		NetworkMode:      ClassicNetwork,
		DataDiskCategory: ecs.DiskCategoryCloud,
		InstanceType:     "ecs.s2.small",
		Password:         "just$test",
	}
	cluster, err := client.CreateCluster(common.Beijing, &args)
	if err != nil {
		t.Fatalf("Failed to CreateCluster: %v", err)
	}

	t.Logf("Cluster: %++v", cluster)
}

func _TestDeleteClusters(t *testing.T) {

	client := NewTestClientForDebug()
	clusterId := "c14601b7676204f73b838329685704902"
	err := client.DeleteCluster(clusterId)
	if err != nil {
		t.Fatalf("Failed to CreateCluster: %v", err)
	}
	t.Logf("Cluster %s is deleting", clusterId)
}
