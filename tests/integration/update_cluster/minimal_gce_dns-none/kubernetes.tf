locals {
  cluster_name = "minimal-gce.example.com"
  project      = "testproject"
  region       = "us-test1"
}

output "cluster_name" {
  value = "minimal-gce.example.com"
}

output "project" {
  value = "testproject"
}

output "region" {
  value = "us-test1"
}

provider "google" {
  project = "testproject"
  region  = "us-test1"
}

provider "aws" {
  alias  = "files"
  region = "us-test-1"
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "tests/minimal-gce.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "tests/minimal-gce.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "tests/minimal-gce.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "tests/minimal-gce.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-master-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-master-us-test1-a_content")
  key                    = "tests/minimal-gce.example.com/manifests/etcd/events-master-us-test1-a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-master-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-master-us-test1-a_content")
  key                    = "tests/minimal-gce.example.com/manifests/etcd/main-master-us-test1-a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "tests/minimal-gce.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce.example.com-addons-bootstrap_content")
  key                    = "tests/minimal-gce.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/minimal-gce.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-example-com-addons-gcp-cloud-controller-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce.example.com-addons-gcp-cloud-controller.addons.k8s.io-k8s-1.23_content")
  key                    = "tests/minimal-gce.example.com/addons/gcp-cloud-controller.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-example-com-addons-gcp-pd-csi-driver-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce.example.com-addons-gcp-pd-csi-driver.addons.k8s.io-k8s-1.23_content")
  key                    = "tests/minimal-gce.example.com/addons/gcp-pd-csi-driver.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "tests/minimal-gce.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "tests/minimal-gce.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "tests/minimal-gce.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-example-com-addons-storage-gce-addons-k8s-io-v1-7-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce.example.com-addons-storage-gce.addons.k8s.io-v1.7.0_content")
  key                    = "tests/minimal-gce.example.com/addons/storage-gce.addons.k8s.io/v1.7.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test1-a_content")
  key                    = "tests/minimal-gce.example.com/igconfig/control-plane/master-us-test1-a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "tests/minimal-gce.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "google_compute_address" "api-us-test1-minimal-gce-example-com" {
  address_type = "INTERNAL"
  name         = "api-us-test1-minimal-gce-example-com"
  purpose      = "SHARED_LOADBALANCER_VIP"
  subnetwork   = google_compute_subnetwork.us-test1-minimal-gce-example-com.name
}

resource "google_compute_disk" "a-etcd-events-minimal-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-example-com"
    "k8s-io-etcd-events"  = "a-2fa"
    "k8s-io-role-master"  = "master"
  }
  name = "a-etcd-events-minimal-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_disk" "a-etcd-main-minimal-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-example-com"
    "k8s-io-etcd-main"    = "a-2fa"
    "k8s-io-role-master"  = "master"
  }
  name = "a-etcd-main-minimal-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_firewall" "https-api-ipv6-minimal-gce-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "https-api-ipv6-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-control-plane"]
}

resource "google_compute_firewall" "https-api-minimal-gce-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "https-api-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-control-plane"]
}

resource "google_compute_firewall" "kops-controller-ipv6-minimal-gce-example-com" {
  allow {
    ports    = ["3988"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "kops-controller-ipv6-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-control-plane"]
}

resource "google_compute_firewall" "kops-controller-minimal-gce-example-com" {
  allow {
    ports    = ["3988"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "kops-controller-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-control-plane"]
}

resource "google_compute_firewall" "lb-health-checks-minimal-gce-example-com" {
  allow {
    protocol = "tcp"
  }
  disabled      = false
  name          = "lb-health-checks-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["130.211.0.0/22", "209.85.152.0/22", "209.85.204.0/22", "35.191.0.0/16"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-control-plane"]
}

resource "google_compute_firewall" "master-to-master-minimal-gce-example-com" {
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "sctp"
  }
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  disabled    = false
  name        = "master-to-master-minimal-gce-example-com"
  network     = google_compute_network.minimal-gce-example-com.name
  source_tags = ["minimal-gce-example-com-k8s-io-role-control-plane", "minimal-gce-example-com-k8s-io-role-master"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-control-plane", "minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-node-minimal-gce-example-com" {
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "sctp"
  }
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  disabled    = false
  name        = "master-to-node-minimal-gce-example-com"
  network     = google_compute_network.minimal-gce-example-com.name
  source_tags = ["minimal-gce-example-com-k8s-io-role-control-plane", "minimal-gce-example-com-k8s-io-role-master"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "node-to-master-minimal-gce-example-com" {
  allow {
    ports    = ["10250"]
    protocol = "tcp"
  }
  allow {
    ports    = ["3988"]
    protocol = "tcp"
  }
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled    = false
  name        = "node-to-master-minimal-gce-example-com"
  network     = google_compute_network.minimal-gce-example-com.name
  source_tags = ["minimal-gce-example-com-k8s-io-role-node"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-control-plane", "minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "node-to-node-minimal-gce-example-com" {
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "sctp"
  }
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  disabled    = false
  name        = "node-to-node-minimal-gce-example-com"
  network     = google_compute_network.minimal-gce-example-com.name
  source_tags = ["minimal-gce-example-com-k8s-io-role-node"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-ipv6-minimal-gce-example-com" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  disabled      = true
  name          = "nodeport-external-to-node-ipv6-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-minimal-gce-example-com" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  disabled      = true
  name          = "nodeport-external-to-node-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-master-ipv6-minimal-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-master-ipv6-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-control-plane", "minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-master-minimal-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-master-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-control-plane", "minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-node-ipv6-minimal-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-node-ipv6-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-node-minimal-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-node-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_forwarding_rule" "api-us-test1-minimal-gce-example-com" {
  backend_service = google_compute_region_backend_service.api-minimal-gce-example-com.id
  ip_address      = google_compute_address.api-us-test1-minimal-gce-example-com.address
  ip_protocol     = "TCP"
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-example-com"
    "name"                = "api-us-test1"
  }
  load_balancing_scheme = "INTERNAL"
  name                  = "api-us-test1-minimal-gce-example-com"
  network               = google_compute_network.minimal-gce-example-com.name
  ports                 = ["443"]
  subnetwork            = google_compute_subnetwork.us-test1-minimal-gce-example-com.name
}

resource "google_compute_forwarding_rule" "kops-controller-us-test1-minimal-gce-example-com" {
  backend_service = google_compute_region_backend_service.api-minimal-gce-example-com.id
  ip_address      = google_compute_address.api-us-test1-minimal-gce-example-com.address
  ip_protocol     = "TCP"
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-example-com"
    "name"                = "kops-controller-us-test1"
  }
  load_balancing_scheme = "INTERNAL"
  name                  = "kops-controller-us-test1-minimal-gce-example-com"
  network               = google_compute_network.minimal-gce-example-com.name
  ports                 = ["3988"]
  subnetwork            = google_compute_subnetwork.us-test1-minimal-gce-example-com.name
}

resource "google_compute_instance_group_manager" "a-master-us-test1-a-minimal-gce-example-com" {
  base_instance_name = "master-us-test1-a"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "a-master-us-test1-a-minimal-gce-example-com"
  target_size                    = 1
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.master-us-test1-a-minimal-gce-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-nodes-minimal-gce-example-com" {
  base_instance_name = "nodes"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "a-nodes-minimal-gce-example-com"
  target_size                    = 2
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.nodes-minimal-gce-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_template" "master-us-test1-a-minimal-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete            = true
    boot                   = true
    device_name            = "persistent-disks-0"
    disk_name              = ""
    disk_size_gb           = 64
    disk_type              = "pd-standard"
    interface              = ""
    mode                   = "READ_WRITE"
    provisioned_iops       = 0
    provisioned_throughput = 0
    source                 = ""
    source_image           = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2004-focal-v20221018"
    type                   = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"       = "minimal-gce-example-com"
    "k8s-io-instance-group"     = "master-us-test1-a"
    "k8s-io-role-control-plane" = "control-plane"
    "k8s-io-role-master"        = "master"
  }
  lifecycle {
    create_before_destroy = true
  }
  machine_type = "e2-medium"
  metadata = {
    "cluster-name"                    = "minimal-gce.example.com"
    "kops-k8s-io-instance-group-name" = "master-us-test1-a"
    "ssh-keys"                        = "admin: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "user-data"                       = file("${path.module}/data/google_compute_instance_template_master-us-test1-a-minimal-gce-example-com_metadata_user-data")
  }
  name_prefix = "master-us-test1-a-minimal-do16cp-"
  network_interface {
    network    = google_compute_network.minimal-gce-example-com.name
    stack_type = "IPV4_ONLY"
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = google_service_account.control-plane.email
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["minimal-gce-example-com-k8s-io-role-control-plane", "minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_instance_template" "nodes-minimal-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete            = true
    boot                   = true
    device_name            = "persistent-disks-0"
    disk_name              = ""
    disk_size_gb           = 128
    disk_type              = "pd-standard"
    interface              = ""
    mode                   = "READ_WRITE"
    provisioned_iops       = 0
    provisioned_throughput = 0
    source                 = ""
    source_image           = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2004-focal-v20221018"
    type                   = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "minimal-gce-example-com"
    "k8s-io-instance-group" = "nodes"
    "k8s-io-role-node"      = "node"
  }
  lifecycle {
    create_before_destroy = true
  }
  machine_type = "e2-medium"
  metadata = {
    "cluster-name"                    = "minimal-gce.example.com"
    "kops-k8s-io-instance-group-name" = "nodes"
    "kube-env"                        = "AUTOSCALER_ENV_VARS: os_distribution=ubuntu;arch=amd64;os=linux"
    "ssh-keys"                        = "admin: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "user-data"                       = file("${path.module}/data/google_compute_instance_template_nodes-minimal-gce-example-com_metadata_user-data")
  }
  name_prefix = "nodes-minimal-gce-example-com-"
  network_interface {
    network    = google_compute_network.minimal-gce-example-com.name
    stack_type = "IPV4_ONLY"
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = google_service_account.node.email
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/devstorage.read_only"]
  }
  tags = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_network" "minimal-gce-example-com" {
  auto_create_subnetworks = false
  name                    = "minimal-gce-example-com"
}

resource "google_compute_region_backend_service" "api-minimal-gce-example-com" {
  backend {
    balancing_mode = "CONNECTION"
    group          = google_compute_instance_group_manager.a-master-us-test1-a-minimal-gce-example-com.instance_group
  }
  health_checks         = [google_compute_region_health_check.api-minimal-gce-example-com.id]
  load_balancing_scheme = "INTERNAL"
  name                  = "api-minimal-gce-example-com"
  protocol              = "TCP"
}

resource "google_compute_region_health_check" "api-minimal-gce-example-com" {
  name = "api-minimal-gce-example-com"
  tcp_health_check {
    port = 443
  }
}

resource "google_compute_router" "nat-minimal-gce-example-com" {
  name    = "nat-minimal-gce-example-com"
  network = google_compute_network.minimal-gce-example-com.name
}

resource "google_compute_router_nat" "nat-minimal-gce-example-com" {
  name                               = "nat-minimal-gce-example-com"
  nat_ip_allocate_option             = "AUTO_ONLY"
  region                             = "us-test1"
  router                             = google_compute_router.nat-minimal-gce-example-com.name
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
  subnetwork {
    name                    = google_compute_subnetwork.us-test1-minimal-gce-example-com.name
    source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
  }
}

resource "google_compute_subnetwork" "us-test1-minimal-gce-example-com" {
  ip_cidr_range = "10.0.16.0/20"
  name          = "us-test1-minimal-gce-example-com"
  network       = google_compute_network.minimal-gce-example-com.name
  region        = "us-test1"
  stack_type    = "IPV4_ONLY"
}

resource "google_project_iam_binding" "serviceaccount-control-plane" {
  members = [format("serviceAccount:%s", google_service_account.control-plane.email)]
  project = "testproject"
  role    = "roles/container.serviceAgent"
}

resource "google_project_iam_binding" "serviceaccount-nodes" {
  members = [format("serviceAccount:%s", google_service_account.node.email)]
  project = "testproject"
  role    = "roles/compute.viewer"
}

resource "google_service_account" "control-plane" {
  account_id   = "control-plane-minimal-g-fu1mg6"
  description  = "kubernetes control-plane instances"
  display_name = "control-plane"
  project      = "testproject"
}

resource "google_service_account" "node" {
  account_id   = "node-minimal-gce-example-com"
  description  = "kubernetes worker nodes"
  display_name = "node"
  project      = "testproject"
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 5.0.0"
    }
    google = {
      "source"  = "hashicorp/google"
      "version" = ">= 5.11.0"
    }
  }
}
