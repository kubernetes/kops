locals {
  cluster_name = "scw-minimal.k8s.local"
  region       = "fr-par"
}

output "cluster_name" {
  value = "scw-minimal.k8s.local"
}

output "region" {
  value = "fr-par"
}

provider "scaleway" {
}

provider "aws" {
  alias  = "files"
  region = "us-test-1"
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "tests/scw-minimal.k8s.local/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "tests/scw-minimal.k8s.local/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "tests/scw-minimal.k8s.local/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "tests/scw-minimal.k8s.local/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-control-plane-fr-par-1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-control-plane-fr-par-1_content")
  key                    = "tests/scw-minimal.k8s.local/manifests/etcd/events-control-plane-fr-par-1.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-control-plane-fr-par-1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-control-plane-fr-par-1_content")
  key                    = "tests/scw-minimal.k8s.local/manifests/etcd/main-control-plane-fr-par-1.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "tests/scw-minimal.k8s.local/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-bootstrap_content")
  key                    = "tests/scw-minimal.k8s.local/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/scw-minimal.k8s.local/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-scaleway-cloud-controller-addons-k8s-io-k8s-1-22" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-scaleway-cloud-controller.addons.k8s.io-k8s-1.22_content")
  key                    = "tests/scw-minimal.k8s.local/addons/scaleway-cloud-controller.addons.k8s.io/k8s-1.22.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-scaleway-csi-driver-addons-k8s-io-k8s-1-22" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-scaleway-csi-driver.addons.k8s.io-k8s-1.22_content")
  key                    = "tests/scw-minimal.k8s.local/addons/scaleway-csi-driver.addons.k8s.io/k8s-1.22.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "tests/scw-minimal.k8s.local/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "tests/scw-minimal.k8s.local/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-limit-range.addons.k8s.io_content")
  key                    = "tests/scw-minimal.k8s.local/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-control-plane-fr-par-1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-control-plane-fr-par-1_content")
  key                    = "tests/scw-minimal.k8s.local/igconfig/control-plane/control-plane-fr-par-1/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes-fr-par-1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes-fr-par-1_content")
  key                    = "tests/scw-minimal.k8s.local/igconfig/node/nodes-fr-par-1/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

###############################
###     LOAD - BALANCER     ###
###############################

resource "scaleway_lb_ip" "api-scw-minimal-k8s-local" {
#  zone = "fr-par-1"
}

resource "scaleway_lb" "api-scw-minimal-k8s-local" {
  ip_id = scaleway_lb_ip.api-scw-minimal-k8s-local.id
#  zone  = scaleway_lb_ip.api-scw-minimal-k8s-local.zone
  type  = "LB-S"
  name  = "api.scw-minimal.k8s.local"
  tags  = [
    "kops.k8s.io/cluster=scw-minimal.k8s.local",
    "kops.k8s.io/role=load-balancer"
  ]
}

resource "scaleway_lb_backend" "api-scw-minimal-k8s-local" {
  lb_id            = scaleway_lb.api-scw-minimal-k8s-local.id
  name             = "lb-backend"
  forward_protocol = "tcp"
  forward_port     = "443"
}

resource "scaleway_lb_frontend" "api-scw-minimal-k8s-local" {
  lb_id        = scaleway_lb.api-scw-minimal-k8s-local.id
  backend_id   = scaleway_lb_backend.api-scw-minimal-k8s-local.id
  name         = "lb-frontend"
  inbound_port = "443"
}

###############################
###        INSTANCES        ###
###############################

## CONTROL - PLANE

resource "scaleway_instance_ip" "control-plane-fr-par-1" {}

resource "scaleway_instance_server" "control-plane-fr-par-1" {
#  zone = "fr-par-1"
  type = "DEV1-M"
  count  = 1
  image  = "ubuntu_jammy"
  ip_id = scaleway_instance_ip.control-plane-fr-par-1.id
  tags = [
    "kops.k8s.io/cluster=scw-minimal.k8s.local",
    "kops.k8s.io/instance-group=control-plane-fr-par-1",
    "kops.k8s.io/role=ControlPlane",
  ]
  name     = "control-plane-fr-par-1-${count.index}"
#  ssh_keys    = [scaleway_ssh_key.scw-minimal-k8s-local-c4_a6_ed_9a_a8_89_b9_e2_c3_9c_d6_63_eb_9c_71_57.id]
  user_data   = {
    cloud-init = filebase64("${path.module}/data/scaleway_server_control-plane-fr-par-1_user_data")
  }
}

## NODE

resource "scaleway_instance_ip" "nodes-fr-par-1" {}

resource "scaleway_instance_server" "nodes-fr-par-1" {
#  zone = "fr-par-1"
  type = "DEV1-M"
  count  = 1
  image  = "ubuntu_jammy"
  ip_id = scaleway_instance_ip.nodes-fr-par-1.id
  tags = [
    "kops.k8s.io/cluster=scw-minimal.k8s.local",
    "kops.k8s.io/instance-group=control-plane-fr-par-1",
    "kops.k8s.io/role=ControlPlane",
  ]
  name     = "nodes-fr-par-1-${count.index}"
  #  ssh_keys    = [scaleway_ssh_key.scw-minimal-k8s-local-c4_a6_ed_9a_a8_89_b9_e2_c3_9c_d6_63_eb_9c_71_57.id]
  user_data   = {
    cloud-init = filebase64("${path.module}/data/scaleway_server_control-plane-fr-par-1_user_data")
  }
}

###############################
###        SSH - KEY        ###
###############################

resource "scaleway_iam_ssh_key" "scw-minimal-k8s-local" {
  name       = "main"
  public_key = "kubernetes.scw-minimal-k8s-local-be:9e:c3:eb:cb:0c:c0:50:ea:bd:b4:5a:15:e3:40:2a"
}

###############################
###         VOLUMES         ###
###############################

resource "scaleway_volume" "etcd-1-etcd-events-scw-minimal-k8s-local" {
  type = "b_ssd"
  name     = "etcd-1.etcd-events.scw-minimal.k8s.local"
  size_in_gb     = 20
  tags = [
    "kops.k8s.io/cluster=scw-minimal.k8s.local",
    "kops.k8s.io/etcd=events",
    "kops.k8s.io/role=ControlPlane",
    "kops.k8s.io/instance-group=control-plane-fr-par-1",
  ]
}

resource "scaleway_volume" "etcd-1-etcd-main-scw-minimal-k8s-local" {
  type = "b_ssd"
  name     = "etcd-1.etcd-main.scw-minimal.k8s.local"
  size_in_gb     = 20
  tags = [
    "kops.k8s.io/cluster=scw-minimal.k8s.local",
    "kops.k8s.io/etcd=main",
    "kops.k8s.io/role=ControlPlane",
    "kops.k8s.io/instance-group=control-plane-fr-par-1",
  ]
}

###############################
###         PROVIDER        ###
###############################

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 4.0.0"
    }
    scaleway = {
      "source"  = "scaleway/scaleway"
      "version" = ">= 2.2.1"
      "zone" = "fr-par-1"
    }
  }
}
