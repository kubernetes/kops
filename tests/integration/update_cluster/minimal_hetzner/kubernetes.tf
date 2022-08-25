locals {
  cluster_name = "minimal.k8s.local"
  region       = "eu-central"
}

output "cluster_name" {
  value = "minimal.k8s.local"
}

output "region" {
  value = "eu-central"
}

provider "hcloud" {
}

provider "aws" {
  alias  = "files"
  region = "us-test-1"
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "tests/minimal.k8s.local/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "tests/minimal.k8s.local/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "tests/minimal.k8s.local/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "tests/minimal.k8s.local/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-master-fsn1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-master-fsn1_content")
  key                    = "tests/minimal.k8s.local/manifests/etcd/events-master-fsn1.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-master-fsn1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-master-fsn1_content")
  key                    = "tests/minimal.k8s.local/manifests/etcd/main-master-fsn1.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "tests/minimal.k8s.local/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-bootstrap_content")
  key                    = "tests/minimal.k8s.local/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/minimal.k8s.local/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/minimal.k8s.local/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-hcloud-cloud-controller-addons-k8s-io-k8s-1-22" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-hcloud-cloud-controller.addons.k8s.io-k8s-1.22_content")
  key                    = "tests/minimal.k8s.local/addons/hcloud-cloud-controller.addons.k8s.io/k8s-1.22.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-hcloud-csi-driver-addons-k8s-io-k8s-1-22" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-hcloud-csi-driver.addons.k8s.io-k8s-1.22_content")
  key                    = "tests/minimal.k8s.local/addons/hcloud-csi-driver.addons.k8s.io/k8s-1.22.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "tests/minimal.k8s.local/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "tests/minimal.k8s.local/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-limit-range.addons.k8s.io_content")
  key                    = "tests/minimal.k8s.local/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-k8s-local-addons-rbac-addons-k8s-io-k8s-1-8" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.k8s.local-addons-rbac.addons.k8s.io-k8s-1.8_content")
  key                    = "tests/minimal.k8s.local/addons/rbac.addons.k8s.io/k8s-1.8.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-fsn1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-fsn1_content")
  key                    = "tests/minimal.k8s.local/igconfig/master/master-fsn1/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes-fsn1" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes-fsn1_content")
  key                    = "tests/minimal.k8s.local/igconfig/node/nodes-fsn1/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "hcloud_firewall" "control-plane-minimal-k8s-local" {
  apply_to {
    label_selector = "kops.k8s.io/cluster=minimal.k8s.local,kops.k8s.io/instance-role=Master"
  }
  labels = {
    "kops.k8s.io/cluster"       = "minimal.k8s.local"
    "kops.k8s.io/firewall-role" = "control-plane"
  }
  name = "control-plane.minimal.k8s.local"
  rule {
    direction  = "in"
    port       = "22"
    protocol   = "tcp"
    source_ips = ["0.0.0.0/0", "::/0"]
  }
}

resource "hcloud_firewall" "nodes-minimal-k8s-local" {
  apply_to {
    label_selector = "kops.k8s.io/cluster=minimal.k8s.local,kops.k8s.io/instance-role=Node"
  }
  labels = {
    "kops.k8s.io/cluster"       = "minimal.k8s.local"
    "kops.k8s.io/firewall-role" = "nodes"
  }
  name = "nodes.minimal.k8s.local"
  rule {
    direction  = "in"
    port       = "22"
    protocol   = "tcp"
    source_ips = ["0.0.0.0/0", "::/0"]
  }
}

resource "hcloud_load_balancer" "api-minimal-k8s-local" {
  labels = {
    "kops.k8s.io/cluster" = "minimal.k8s.local"
  }
  load_balancer_type = "lb11"
  location           = "fsn1"
  name               = "api.minimal.k8s.local"
}

resource "hcloud_load_balancer_network" "api-minimal-k8s-local" {
  load_balancer_id = hcloud_load_balancer.api-minimal-k8s-local.id
  network_id       = hcloud_network.minimal-k8s-local.id
}

resource "hcloud_load_balancer_service" "api-minimal-k8s-local" {
  destination_port = 443
  listen_port      = 443
  load_balancer_id = hcloud_load_balancer.api-minimal-k8s-local.id
  protocol         = "tcp"
}

resource "hcloud_load_balancer_target" "api-minimal-k8s-local" {
  label_selector   = "kops.k8s.io/cluster=minimal.k8s.local,kops.k8s.io/instance-role=Master"
  load_balancer_id = hcloud_load_balancer.api-minimal-k8s-local.id
  type             = "label_selector"
  use_private_ip   = true
}

resource "hcloud_network" "minimal-k8s-local" {
  ip_range = "10.0.0.0/16"
  labels = {
    "kops.k8s.io/cluster" = "minimal.k8s.local"
  }
  name = "minimal.k8s.local"
}

resource "hcloud_network_subnet" "minimal-k8s-local-10-0-0-0--16" {
  ip_range     = "10.0.0.0/16"
  network_id   = hcloud_network.minimal-k8s-local.id
  network_zone = "eu-central"
  type         = "cloud"
}

resource "hcloud_server" "master-fsn1" {
  count = 1
  image = "ubuntu-20.04"
  labels = {
    "kops.k8s.io/cluster"        = "minimal.k8s.local"
    "kops.k8s.io/instance-group" = "master-fsn1"
    "kops.k8s.io/instance-role"  = "Master"
  }
  location = "fsn1"
  name     = "master-fsn1-${count.index}"
  network {
    network_id = hcloud_network.minimal-k8s-local.id
  }
  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  server_type = "cx21"
  ssh_keys    = [hcloud_ssh_key.minimal-k8s-local-c4_a6_ed_9a_a8_89_b9_e2_c3_9c_d6_63_eb_9c_71_57.id]
  user_data   = filebase64("${path.module}/data/hcloud_server_master-fsn1_user_data")
}

resource "hcloud_server" "nodes-fsn1" {
  count = 1
  image = "ubuntu-20.04"
  labels = {
    "kops.k8s.io/cluster"        = "minimal.k8s.local"
    "kops.k8s.io/instance-group" = "nodes-fsn1"
    "kops.k8s.io/instance-role"  = "Node"
  }
  location = "fsn1"
  name     = "nodes-fsn1-${count.index}"
  network {
    network_id = hcloud_network.minimal-k8s-local.id
  }
  public_net {
    ipv4_enabled = true
    ipv6_enabled = false
  }
  server_type = "cx21"
  ssh_keys    = [hcloud_ssh_key.minimal-k8s-local-c4_a6_ed_9a_a8_89_b9_e2_c3_9c_d6_63_eb_9c_71_57.id]
  user_data   = filebase64("${path.module}/data/hcloud_server_nodes-fsn1_user_data")
}

resource "hcloud_ssh_key" "minimal-k8s-local-c4_a6_ed_9a_a8_89_b9_e2_c3_9c_d6_63_eb_9c_71_57" {
  labels = {
    "kops.k8s.io/cluster" = "minimal.k8s.local"
  }
  name       = "minimal.k8s.local-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
}

resource "hcloud_volume" "etcd-1-etcd-events-minimal-k8s-local" {
  labels = {
    "kops.k8s.io/cluster"        = "minimal.k8s.local"
    "kops.k8s.io/instance-group" = "master-fsn1"
    "kops.k8s.io/volume-role"    = "events"
  }
  location = "fsn1"
  name     = "etcd-1.etcd-events.minimal.k8s.local"
  size     = 20
}

resource "hcloud_volume" "etcd-1-etcd-main-minimal-k8s-local" {
  labels = {
    "kops.k8s.io/cluster"        = "minimal.k8s.local"
    "kops.k8s.io/instance-group" = "master-fsn1"
    "kops.k8s.io/volume-role"    = "main"
  }
  location = "fsn1"
  name     = "etcd-1.etcd-main.minimal.k8s.local"
  size     = 20
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    hcloud = {
      "source"  = "hetznercloud/hcloud"
      "version" = ">= 1.35.1"
    }
  }
}
