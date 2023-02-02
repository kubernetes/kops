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
  region = "fr-par"
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

resource "aws_s3_object" "scw-minimal-k8s-local-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/scw-minimal.k8s.local/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
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

resource "aws_s3_object" "scw-minimal-k8s-local-addons-networking-cilium-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-networking.cilium.io-k8s-1.16_content")
  key                    = "tests/scw-minimal.k8s.local/addons/networking.cilium.io/k8s-1.16-v1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-rbac-addons-k8s-io-k8s-1-8" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-rbac.addons.k8s.io-k8s-1.8_content")
  key                    = "tests/scw-minimal.k8s.local/addons/rbac.addons.k8s.io/k8s-1.8.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-scaleway-cloud-controller-addons-k8s-io-k8s-1-24" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-scaleway-cloud-controller.addons.k8s.io-k8s-1.24_content")
  key                    = "tests/scw-minimal.k8s.local/addons/scaleway-cloud-controller.addons.k8s.io/k8s-1.24.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-k8s-local-addons-scaleway-csi-driver-addons-k8s-io-k8s-1-24" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.k8s.local-addons-scaleway-csi-driver.addons.k8s.io-k8s-1.24_content")
  key                    = "tests/scw-minimal.k8s.local/addons/scaleway-csi-driver.addons.k8s.io/k8s-1.24.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "scaleway_iam_ssh_key" "kubernetes-scw-minimal-k8s-local-be_9e_c3_eb_cb_0c_c0_50_ea_bd_b4_5a_15_e3_40_2a" {
  name       = "kubernetes-scw-minimal-k8s.local-be:9e:c3:eb:cb:0c:c0:50:ea:bd:b4:5a:15:e3:40:2a"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDKqbVEozfAqng0gx8HTUu69EppcE5SWet6MpwrGShqMVUC4wkoiuVtJDPhMmWmdt7B7Ttc5pvnAZAZaQ6TKMguyBoAyS7qOTLU9/hM803XtSiwQUftOXiJfmsqAXEc8yDyb7UnrF8X7aA3gQJsnQBGJGdp+C88dPHNZenw4PnQc8BNYTCXG9d8F5vJ3xQ5qbiG4HVNoQ2CZh2ht+GedZJ3hl9lMJ24kE/cbMCLKxabMP4ROetECG6PU251jnm84NA8rm0Av/JMmn/c9CFAe0D0D1dGDlHWPsk4mbhGKJ0yU0YliatmPfmgSasismbYzIFf7VPq91ARzRUbavd1fYMBmkMsce0YR/5FdtrpzRhqDzuvwQgQRsoTcttdvp0puFcrtNefMfk8NCbBedIlkzOFxfGiBbe6jde4wqsqEnSrNHwZ2b+Er8z7vjcDPBqYk3gubmMBCrYxg6o1lOS6tTN0kJDUlyKO2AN1ZDr3mpkbhkvZV/N7gLglcClM0X5X7iM= leila@leila-ThinkPad-T14s-Gen-2i"
}

resource "scaleway_instance_ip" "control-plane-fr-par-1" {
}

resource "scaleway_instance_ip" "nodes-fr-par-1" {
}

resource "scaleway_instance_server" "control-plane-fr-par-1" {
  image = "ubuntu_focal"
  ip_id = scaleway_instance_ip.control-plane-fr-par-1.id
  name  = "control-plane-fr-par-1"
  tags  = ["kops.k8s.io/instance-group=control-plane-fr-par-1", "kops.k8s.io/cluster=scw-minimal.k8s.local", "kops.k8s.io/role=ControlPlane"]
  type  = "DEV1-M"
  user_data = {
    "cloud-init" = filebase64("${path.module}/data/scaleway_instance_server_control-plane-fr-par-1_user_data")
  }
}

resource "scaleway_instance_server" "nodes-fr-par-1" {
  image = "ubuntu_focal"
  ip_id = scaleway_instance_ip.nodes-fr-par-1.id
  name  = "nodes-fr-par-1"
  tags  = ["kops.k8s.io/instance-group=nodes-fr-par-1", "kops.k8s.io/cluster=scw-minimal.k8s.local"]
  type  = "DEV1-M"
  user_data = {
    "cloud-init" = filebase64("${path.module}/data/scaleway_instance_server_nodes-fr-par-1_user_data")
  }
}

resource "scaleway_instance_volume" "etcd-1-etcd-events-scw-minimal-k8s-local" {
  name       = "etcd-1.etcd-events.scw-minimal.k8s.local"
  size_in_gb = 20
  tags       = ["kops.k8s.io/cluster=scw-minimal.k8s.local", "kops.k8s.io/etcd=events", "kops.k8s.io/role=ControlPlane", "kops.k8s.io/instance-group=control-plane-fr-par-1"]
  type       = "b_ssd"
}

resource "scaleway_instance_volume" "etcd-1-etcd-main-scw-minimal-k8s-local" {
  name       = "etcd-1.etcd-main.scw-minimal.k8s.local"
  size_in_gb = 20
  tags       = ["kops.k8s.io/cluster=scw-minimal.k8s.local", "kops.k8s.io/etcd=main", "kops.k8s.io/role=ControlPlane", "kops.k8s.io/instance-group=control-plane-fr-par-1"]
  type       = "b_ssd"
}

resource "scaleway_lb" "api-scw-minimal-k8s-local" {
  ip_id = scaleway_lb_ip.api-scw-minimal-k8s-local.id
  name  = "api.scw-minimal.k8s.local"
  tags  = ["kops.k8s.io/cluster=scw-minimal.k8s.local", "kops.k8s.io/role=ControlPlane"]
  type  = "LB-S"
}

resource "scaleway_lb_backend" "api-scw-minimal-k8s-local" {
  forward_port     = 443
  forward_protocol = "tcp"
  lb_id            = scaleway_lb.api-scw-minimal-k8s-local.id
  name             = "lb-backend"
}

resource "scaleway_lb_frontend" "api-scw-minimal-k8s-local" {
  backend_id   = scaleway_lb_backend.api-scw-minimal-k8s-local.id
  inbound_port = 443
  lb_id        = scaleway_lb.api-scw-minimal-k8s-local.id
  name         = "lb-frontend"
}

resource "scaleway_lb_ip" "api-scw-minimal-k8s-local" {
}

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
    }
  }
}
