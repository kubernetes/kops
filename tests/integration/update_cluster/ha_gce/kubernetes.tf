locals {
  cluster_name = "ha-gce.example.com"
  project      = "us-test1"
  region       = "us-test1"
}

output "cluster_name" {
  value = "ha-gce.example.com"
}

output "project" {
  value = "us-test1"
}

output "region" {
  value = "us-test1"
}

provider "google" {
  region  = "us-test1"
  version = ">= 3.0.0"
}

resource "google_compute_disk" "d1-etcd-events-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-events"  = "1-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d1-etcd-events-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_disk" "d1-etcd-main-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-main"    = "1-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d1-etcd-main-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_disk" "d2-etcd-events-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-events"  = "2-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d2-etcd-events-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-b"
}

resource "google_compute_disk" "d2-etcd-main-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-main"    = "2-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d2-etcd-main-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-b"
}

resource "google_compute_disk" "d3-etcd-events-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-events"  = "3-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d3-etcd-events-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-c"
}

resource "google_compute_disk" "d3-etcd-main-ha-gce-example-com" {
  labels = {
    "k8s-io-cluster-name" = "ha-gce-example-com"
    "k8s-io-etcd-main"    = "3-2f1-2c2-2c3"
    "k8s-io-role-master"  = "master"
  }
  name = "d3-etcd-main-ha-gce-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-c"
}

resource "google_compute_firewall" "cidr-to-master-ha-gce-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  allow {
    ports    = ["4194"]
    protocol = "tcp"
  }
  name          = "cidr-to-master-ha-gce-example-com"
  network       = google_compute_network.default.name
  source_ranges = ["100.64.0.0/10"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "cidr-to-node-ha-gce-example-com" {
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "sctp"
  }
  name          = "cidr-to-node-ha-gce-example-com"
  network       = google_compute_network.default.name
  source_ranges = ["100.64.0.0/10"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "kubernetes-master-https-ha-gce-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  name          = "kubernetes-master-https-ha-gce-example-com"
  network       = google_compute_network.default.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-master-ha-gce-example-com" {
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "sctp"
  }
  name        = "master-to-master-ha-gce-example-com"
  network     = google_compute_network.default.name
  source_tags = ["ha-gce-example-com-k8s-io-role-master"]
  target_tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-node-ha-gce-example-com" {
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "sctp"
  }
  name        = "master-to-node-ha-gce-example-com"
  network     = google_compute_network.default.name
  source_tags = ["ha-gce-example-com-k8s-io-role-master"]
  target_tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "node-to-master-ha-gce-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  allow {
    ports    = ["4194"]
    protocol = "tcp"
  }
  name        = "node-to-master-ha-gce-example-com"
  network     = google_compute_network.default.name
  source_tags = ["ha-gce-example-com-k8s-io-role-node"]
  target_tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "node-to-node-ha-gce-example-com" {
  allow {
    protocol = "tcp"
  }
  allow {
    protocol = "udp"
  }
  allow {
    protocol = "icmp"
  }
  allow {
    protocol = "esp"
  }
  allow {
    protocol = "ah"
  }
  allow {
    protocol = "sctp"
  }
  name        = "node-to-node-ha-gce-example-com"
  network     = google_compute_network.default.name
  source_tags = ["ha-gce-example-com-k8s-io-role-node"]
  target_tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-ha-gce-example-com" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  name        = "nodeport-external-to-node-ha-gce-example-com"
  network     = google_compute_network.default.name
  source_tags = ["ha-gce-example-com-k8s-io-role-node"]
  target_tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-master-ha-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  name          = "ssh-external-to-master-ha-gce-example-com"
  network       = google_compute_network.default.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-node-ha-gce-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  name          = "ssh-external-to-node-ha-gce-example-com"
  network       = google_compute_network.default.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_instance_group_manager" "a-master-us-test1-a-ha-gce-example-com" {
  base_instance_name = "master-us-test1-a"
  name               = "a-master-us-test1-a-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.master-us-test1-a-ha-gce-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-nodes-ha-gce-example-com" {
  base_instance_name = "nodes"
  name               = "a-nodes-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.nodes-ha-gce-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "b-master-us-test1-b-ha-gce-example-com" {
  base_instance_name = "master-us-test1-b"
  name               = "b-master-us-test1-b-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.master-us-test1-b-ha-gce-example-com.self_link
  }
  zone = "us-test1-b"
}

resource "google_compute_instance_group_manager" "b-nodes-ha-gce-example-com" {
  base_instance_name = "nodes"
  name               = "b-nodes-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.nodes-ha-gce-example-com.self_link
  }
  zone = "us-test1-b"
}

resource "google_compute_instance_group_manager" "c-master-us-test1-c-ha-gce-example-com" {
  base_instance_name = "master-us-test1-c"
  name               = "c-master-us-test1-c-ha-gce-example-com"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.master-us-test1-c-ha-gce-example-com.self_link
  }
  zone = "us-test1-c"
}

resource "google_compute_instance_group_manager" "c-nodes-ha-gce-example-com" {
  base_instance_name = "nodes"
  name               = "c-nodes-ha-gce-example-com"
  target_size        = 0
  version {
    instance_template = google_compute_instance_template.nodes-ha-gce-example-com.self_link
  }
  zone = "us-test1-c"
}

resource "google_compute_instance_template" "master-us-test1-a-ha-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete  = true
    boot         = true
    device_name  = "persistent-disks-0"
    disk_name    = ""
    disk_size_gb = 64
    disk_type    = "pd-standard"
    interface    = ""
    mode         = "READ_WRITE"
    source       = ""
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    type         = "PERSISTENT"
  }
  machine_type = "n1-standard-1"
  metadata = {
    "cluster-name"                    = file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_cluster-name")
    "kops-k8s-io-instance-group-name" = file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_kops-k8s-io-instance-group-name")
    "ssh-keys"                        = file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_ssh-keys")
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_startup-script")
  }
  name_prefix = "master-us-test1-a-ha-gce--ke5ah6-"
  network_interface {
    access_config {
    }
    network = google_compute_network.default.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_instance_template" "master-us-test1-b-ha-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete  = true
    boot         = true
    device_name  = "persistent-disks-0"
    disk_name    = ""
    disk_size_gb = 64
    disk_type    = "pd-standard"
    interface    = ""
    mode         = "READ_WRITE"
    source       = ""
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    type         = "PERSISTENT"
  }
  machine_type = "n1-standard-1"
  metadata = {
    "cluster-name"                    = file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_cluster-name")
    "kops-k8s-io-instance-group-name" = file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_kops-k8s-io-instance-group-name")
    "ssh-keys"                        = file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_ssh-keys")
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_startup-script")
  }
  name_prefix = "master-us-test1-b-ha-gce--c8u7qq-"
  network_interface {
    access_config {
    }
    network = google_compute_network.default.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_instance_template" "master-us-test1-c-ha-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete  = true
    boot         = true
    device_name  = "persistent-disks-0"
    disk_name    = ""
    disk_size_gb = 64
    disk_type    = "pd-standard"
    interface    = ""
    mode         = "READ_WRITE"
    source       = ""
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    type         = "PERSISTENT"
  }
  machine_type = "n1-standard-1"
  metadata = {
    "cluster-name"                    = file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_cluster-name")
    "kops-k8s-io-instance-group-name" = file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_kops-k8s-io-instance-group-name")
    "ssh-keys"                        = file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_ssh-keys")
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_startup-script")
  }
  name_prefix = "master-us-test1-c-ha-gce--3unp7l-"
  network_interface {
    access_config {
    }
    network = google_compute_network.default.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_instance_template" "nodes-ha-gce-example-com" {
  can_ip_forward = true
  disk {
    auto_delete  = true
    boot         = true
    device_name  = "persistent-disks-0"
    disk_name    = ""
    disk_size_gb = 128
    disk_type    = "pd-standard"
    interface    = ""
    mode         = "READ_WRITE"
    source       = ""
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    type         = "PERSISTENT"
  }
  machine_type = "n1-standard-2"
  metadata = {
    "cluster-name"                    = file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_cluster-name")
    "kops-k8s-io-instance-group-name" = file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_kops-k8s-io-instance-group-name")
    "ssh-keys"                        = file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_ssh-keys")
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_startup-script")
  }
  name_prefix = "nodes-ha-gce-example-com-"
  network_interface {
    access_config {
    }
    network = google_compute_network.default.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_only"]
  }
  tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_network" "default" {
  auto_create_subnetworks = true
  name                    = "default"
}

terraform {
  required_version = ">= 0.12.0"
}
