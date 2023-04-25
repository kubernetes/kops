locals {
  cluster_name = "scw-minimal.k8s.local"
  region       = "fr-par"
  zone         = "fr-par-1"
}

output "cluster_name" {
  value = "scw-minimal.k8s.local"
}

output "region" {
  value = "fr-par"
}

output "zone" {
  value = "fr-par-1"
}

provider "scaleway" {
  region = "fr-par"
  zone   = "fr-par-1"
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
  key                    = "tests/scw-minimal.k8s.local/addons/networking.cilium.io/k8s-1.16-v1.15.yaml"
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

resource "scaleway_iam_ssh_key" "kubernetes-scw-minimal-k8s-local-ae_ea_e9_42_75_4b_cd_2a_0d_68_c8_5a_af_7c_b1_c4" {
  name       = "kubernetes.scw-minimal.k8s.local-ae:ea:e9:42:75:4b:cd:2a:0d:68:c8:5a:af:7c:b1:c4"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDDJ3lhaLGWT/YYhVMQFNo9YH1q4MGgO3qOX0dCU6bnTTchOHs0wByRuWhq7OnLt4H2JXSn3FEnjtnFDV9NvrTpg9fkovUdWJJCt7JYcEUcDSy5+G1LyO/PEDVyxuHmBiyek92JU+Kl9QCARaZRjvhYuvH4hCI0HaMYap/livITnUDyI+OM6hEbkXXjzHr6s+5Fwy/ztH4rgq8f+ZmejmjczpDD/6LQzsVtcrAF5SMaUvjuwfOhBpyWtwtwfi35k8ac0+CPBI/s6bnAGgtg6ylhC5Er2YFSKpg2S0hdR42pzjdlB3eXB1eKlU7IRAaEdrR56PSzUcrefOJ9DRsQ39PKaS5j/IJvrEUL9TOvF5PmIuWBW1fWh9BWAveivE/rhC9i96QiU7sFxVq2p6sR5tu9o0UEY9xLHaYbG2DfRDb7TXlPS3uBZ/yDRv7cJtH6cEEaAiGhCXBJ108xCpM0Dbva/jfzwW7rBRqJDRZsN1Kas2S2saTzPNHj4ANG6wKDWEc="
}

resource "scaleway_instance_ip" "control-plane-fr-par-1-0" {
  tags = ["noprefix=kops.k8s.io/cluster=scw-minimal.k8s.local"]
}

resource "scaleway_instance_ip" "nodes-fr-par-1-0" {
  tags = ["noprefix=kops.k8s.io/cluster=scw-minimal.k8s.local"]
}

resource "scaleway_instance_server" "control-plane-fr-par-1-0" {
  enable_dynamic_ip = true
  image             = "ubuntu_focal"
  ip_id             = scaleway_instance_ip.control-plane-fr-par-1-0.id
  lifecycle {
    ignore_changes = [additional_volume_ids]
  }
  name                   = "control-plane-fr-par-1-0"
  replace_on_type_change = false
  tags                   = ["noprefix=kops.k8s.io/cluster=scw-minimal.k8s.local", "noprefix=kops.k8s.io/instance-group=control-plane-fr-par-1", "noprefix=kops.k8s.io/role=ControlPlane"]
  type                   = "DEV1-M"
  user_data = {
    "cloud-init" = file("${path.module}/data/scaleway_instance_server_control-plane-fr-par-1-0_user_data")
  }
}

resource "scaleway_instance_server" "nodes-fr-par-1-0" {
  enable_dynamic_ip      = true
  image                  = "ubuntu_focal"
  ip_id                  = scaleway_instance_ip.nodes-fr-par-1-0.id
  name                   = "nodes-fr-par-1-0"
  replace_on_type_change = false
  tags                   = ["noprefix=kops.k8s.io/cluster=scw-minimal.k8s.local", "noprefix=kops.k8s.io/instance-group=nodes-fr-par-1"]
  type                   = "DEV1-M"
  user_data = {
    "cloud-init" = file("${path.module}/data/scaleway_instance_server_nodes-fr-par-1-0_user_data")
  }
}

resource "scaleway_instance_volume" "etcd-1-etcd-events-scw-minimal-k8s-local" {
  name       = "etcd-1.etcd-events.scw-minimal.k8s.local"
  size_in_gb = 20
  tags       = ["noprefix=kops.k8s.io/cluster=scw-minimal.k8s.local", "noprefix=kops.k8s.io/etcd=events", "noprefix=kops.k8s.io/role=ControlPlane", "noprefix=kops.k8s.io/instance-group=control-plane-fr-par-1"]
  type       = "b_ssd"
}

resource "scaleway_instance_volume" "etcd-1-etcd-main-scw-minimal-k8s-local" {
  name       = "etcd-1.etcd-main.scw-minimal.k8s.local"
  size_in_gb = 20
  tags       = ["noprefix=kops.k8s.io/cluster=scw-minimal.k8s.local", "noprefix=kops.k8s.io/etcd=main", "noprefix=kops.k8s.io/role=ControlPlane", "noprefix=kops.k8s.io/instance-group=control-plane-fr-par-1"]
  type       = "b_ssd"
}

resource "scaleway_lb" "api-scw-minimal-k8s-local" {
  description = "Load-balancer for kops cluster scw-minimal.k8s.local"
  ip_id       = scaleway_lb_ip.api-scw-minimal-k8s-local.id
  name        = "api.scw-minimal.k8s.local"
  tags        = ["noprefix=kops.k8s.io/cluster=scw-minimal.k8s.local", "noprefix=kops.k8s.io/role=ControlPlane"]
  type        = "LB-S"
}

resource "scaleway_lb_backend" "lb-backend-https" {
  forward_port     = 443
  forward_protocol = "tcp"
  lb_id            = scaleway_lb.api-scw-minimal-k8s-local.id
  name             = "lb-backend-https"
  proxy_protocol   = "none"
  server_ips       = [scaleway_instance_server.control-plane-fr-par-1-0.private_ip]
}

resource "scaleway_lb_backend" "lb-backend-kops-controller" {
  forward_port     = 3988
  forward_protocol = "tcp"
  lb_id            = scaleway_lb.api-scw-minimal-k8s-local.id
  name             = "lb-backend-kops-controller"
  proxy_protocol   = "none"
  server_ips       = [scaleway_instance_server.control-plane-fr-par-1-0.private_ip]
}

resource "scaleway_lb_frontend" "lb-frontend-https" {
  backend_id   = scaleway_lb_backend.lb-backend-https.id
  inbound_port = 443
  lb_id        = scaleway_lb.api-scw-minimal-k8s-local.id
  name         = "lb-frontend-https"
}

resource "scaleway_lb_frontend" "lb-frontend-kops-controller" {
  backend_id   = scaleway_lb_backend.lb-backend-kops-controller.id
  inbound_port = 3988
  lb_id        = scaleway_lb.api-scw-minimal-k8s-local.id
  name         = "lb-frontend-kops-controller"
}

resource "scaleway_lb_ip" "api-scw-minimal-k8s-local" {
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 5.0.0"
    }
    scaleway = {
      "source"  = "scaleway/scaleway"
      "version" = ">= 2.2.1"
    }
  }
}
