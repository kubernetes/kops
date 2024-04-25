locals {
  cluster_name = "scw-minimal.example.com"
  region       = "fr-par"
  zone         = "fr-par-1"
}

output "cluster_name" {
  value = "scw-minimal.example.com"
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
  key                    = "tests/scw-minimal.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "tests/scw-minimal.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "tests/scw-minimal.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "tests/scw-minimal.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-control-plane-fr-par-1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-control-plane-fr-par-1_content")
  key                    = "tests/scw-minimal.example.com/manifests/etcd/events-control-plane-fr-par-1.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-control-plane-fr-par-1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-control-plane-fr-par-1_content")
  key                    = "tests/scw-minimal.example.com/manifests/etcd/main-control-plane-fr-par-1.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "tests/scw-minimal.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-control-plane-fr-par-1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-control-plane-fr-par-1_content")
  key                    = "tests/scw-minimal.example.com/igconfig/control-plane/control-plane-fr-par-1/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes-fr-par-1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes-fr-par-1_content")
  key                    = "tests/scw-minimal.example.com/igconfig/node/nodes-fr-par-1/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.example.com-addons-bootstrap_content")
  key                    = "tests/scw-minimal.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/scw-minimal.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "tests/scw-minimal.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "tests/scw-minimal.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "tests/scw-minimal.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-example-com-addons-networking-cilium-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.example.com-addons-networking.cilium.io-k8s-1.16_content")
  key                    = "tests/scw-minimal.example.com/addons/networking.cilium.io/k8s-1.16-v1.15.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-example-com-addons-scaleway-cloud-controller-addons-k8s-io-k8s-1-24" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.example.com-addons-scaleway-cloud-controller.addons.k8s.io-k8s-1.24_content")
  key                    = "tests/scw-minimal.example.com/addons/scaleway-cloud-controller.addons.k8s.io/k8s-1.24.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "scw-minimal-example-com-addons-scaleway-csi-driver-addons-k8s-io-k8s-1-24" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_scw-minimal.example.com-addons-scaleway-csi-driver.addons.k8s.io-k8s-1.24_content")
  key                    = "tests/scw-minimal.example.com/addons/scaleway-csi-driver.addons.k8s.io/k8s-1.24.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "scaleway_iam_ssh_key" "kubernetes-scw-minimal-example-com-be_9e_c3_eb_cb_0c_c0_50_ea_bd_b4_5a_15_e3_40_2a" {
  name       = "kubernetes.scw-minimal.example.com-be:9e:c3:eb:cb:0c:c0:50:ea:bd:b4:5a:15:e3:40:2a"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDKqbVEozfAqng0gx8HTUu69EppcE5SWet6MpwrGShqMVUC4wkoiuVtJDPhMmWmdt7B7Ttc5pvnAZAZaQ6TKMguyBoAyS7qOTLU9/hM803XtSiwQUftOXiJfmsqAXEc8yDyb7UnrF8X7aA3gQJsnQBGJGdp+C88dPHNZenw4PnQc8BNYTCXG9d8F5vJ3xQ5qbiG4HVNoQ2CZh2ht+GedZJ3hl9lMJ24kE/cbMCLKxabMP4ROetECG6PU251jnm84NA8rm0Av/JMmn/c9CFAe0D0D1dGDlHWPsk4mbhGKJ0yU0YliatmPfmgSasismbYzIFf7VPq91ARzRUbavd1fYMBmkMsce0YR/5FdtrpzRhqDzuvwQgQRsoTcttdvp0puFcrtNefMfk8NCbBedIlkzOFxfGiBbe6jde4wqsqEnSrNHwZ2b+Er8z7vjcDPBqYk3gubmMBCrYxg6o1lOS6tTN0kJDUlyKO2AN1ZDr3mpkbhkvZV/N7gLglcClM0X5X7iM= leila@leila-ThinkPad-T14s-Gen-2i"
}

resource "scaleway_instance_ip" "control-plane-fr-par-1-0" {
  tags = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com"]
}

resource "scaleway_instance_ip" "nodes-fr-par-1-0" {
  tags = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com"]
}

resource "scaleway_instance_private_nic" "control-plane-fr-par-1-0" {
  private_network_id = scaleway_vpc_private_network.scw-minimal-example-com.id
  server_id          = scaleway_instance_server.control-plane-fr-par-1-0.id
  tags               = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com", "noprefix=kops.k8s.io/instance-group=control-plane-fr-par-1"]
}

resource "scaleway_instance_private_nic" "nodes-fr-par-1-0" {
  private_network_id = scaleway_vpc_private_network.scw-minimal-example-com.id
  server_id          = scaleway_instance_server.nodes-fr-par-1-0.id
  tags               = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com", "noprefix=kops.k8s.io/instance-group=nodes-fr-par-1"]
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
  tags                   = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com", "noprefix=kops.k8s.io/instance-group=control-plane-fr-par-1", "noprefix=kops.k8s.io/role=ControlPlane"]
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
  tags                   = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com", "noprefix=kops.k8s.io/instance-group=nodes-fr-par-1", "noprefix=kops.k8s.io/role=Node"]
  type                   = "DEV1-M"
  user_data = {
    "cloud-init" = file("${path.module}/data/scaleway_instance_server_nodes-fr-par-1-0_user_data")
  }
}

resource "scaleway_instance_volume" "etcd-1-etcd-events-scw-minimal-example-com" {
  name       = "etcd-1.etcd-events.scw-minimal.example.com"
  size_in_gb = 20
  tags       = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com", "noprefix=kops.k8s.io/etcd=events", "noprefix=kops.k8s.io/role=ControlPlane", "noprefix=kops.k8s.io/instance-group=control-plane-fr-par-1"]
  type       = "b_ssd"
}

resource "scaleway_instance_volume" "etcd-1-etcd-main-scw-minimal-example-com" {
  name       = "etcd-1.etcd-main.scw-minimal.example.com"
  size_in_gb = 20
  tags       = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com", "noprefix=kops.k8s.io/etcd=main", "noprefix=kops.k8s.io/role=ControlPlane", "noprefix=kops.k8s.io/instance-group=control-plane-fr-par-1"]
  type       = "b_ssd"
}

resource "scaleway_lb" "api-scw-minimal-example-com" {
  description = "Load-balancer for kops cluster scw-minimal.example.com"
  ip_id       = scaleway_lb_ip.api-scw-minimal-example-com.id
  name        = "api.scw-minimal.example.com"
  tags        = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com", "noprefix=kops.k8s.io/role=ControlPlane"]
  type        = "LB-S"
}

resource "scaleway_lb_backend" "lb-backend-https" {
  forward_port     = 443
  forward_protocol = "tcp"
  lb_id            = scaleway_lb.api-scw-minimal-example-com.id
  name             = "lb-backend-https"
  proxy_protocol   = "none"
  server_ips       = [scaleway_instance_server.control-plane-fr-par-1-0.private_ip]
}

resource "scaleway_lb_backend" "lb-backend-kops-controller" {
  forward_port     = 3988
  forward_protocol = "tcp"
  lb_id            = scaleway_lb.api-scw-minimal-example-com.id
  name             = "lb-backend-kops-controller"
  proxy_protocol   = "none"
  server_ips       = [scaleway_instance_server.control-plane-fr-par-1-0.private_ip]
}

resource "scaleway_lb_frontend" "lb-frontend-https" {
  backend_id   = scaleway_lb_backend.lb-backend-https.id
  inbound_port = 443
  lb_id        = scaleway_lb.api-scw-minimal-example-com.id
  name         = "lb-frontend-https"
}

resource "scaleway_lb_frontend" "lb-frontend-kops-controller" {
  backend_id   = scaleway_lb_backend.lb-backend-kops-controller.id
  inbound_port = 3988
  lb_id        = scaleway_lb.api-scw-minimal-example-com.id
  name         = "lb-frontend-kops-controller"
}

resource "scaleway_lb_ip" "api-scw-minimal-example-com" {
}

resource "scaleway_vpc" "scw-minimal-example-com" {
  name = "scw-minimal.example.com"
  tags = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com"]
}

resource "scaleway_vpc_gateway_network" "scw-minimal-example-com" {
  enable_dhcp       = true
  enable_masquerade = true
  gateway_id        = scaleway_vpc_public_gateway.scw-minimal-example-com.id
  ipam_config {
    push_default_route = true
  }
  private_network_id = scaleway_vpc_private_network.scw-minimal-example-com.id
}

resource "scaleway_vpc_private_network" "scw-minimal-example-com" {
  ipv4_subnet {
    subnet = "192.168.1.0/24"
  }
  name   = "scw-minimal.example.com"
  tags   = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com"]
  vpc_id = scaleway_vpc.scw-minimal-example-com.id
}

resource "scaleway_vpc_public_gateway" "scw-minimal-example-com" {
  name = "scw-minimal.example.com"
  tags = ["noprefix=kops.k8s.io/cluster=scw-minimal.example.com"]
  type = "VPC-GW-S"
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
