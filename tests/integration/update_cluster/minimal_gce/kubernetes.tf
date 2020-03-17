locals = {
  cluster_name = "minimal-gce.example.com"
  project      = "us-test1"
  region       = "us-test1"
}

output "cluster_name" {
  value = "minimal-gce.example.com"
}

output "project" {
  value = "us-test1"
}

output "region" {
  value = "us-test1"
}

provider "google" {
  project = "us-test1"
  region  = "us-test1"
  version = ">= 3.0.0"
}

resource "google_compute_disk" "d1-etcd-events-minimal-gce-example-com" {
  name = "d1-etcd-events-minimal-gce-example-com"
  type = "pd-ssd"
  size = 20
  zone = "us-test1-a"

  labels = {
    k8s-io-cluster-name = "minimal-gce-example-com"
    k8s-io-etcd-events  = "1-2f1"
    k8s-io-role-master  = "master"
  }
}

resource "google_compute_disk" "d1-etcd-main-minimal-gce-example-com" {
  name = "d1-etcd-main-minimal-gce-example-com"
  type = "pd-ssd"
  size = 20
  zone = "us-test1-a"

  labels = {
    k8s-io-cluster-name = "minimal-gce-example-com"
    k8s-io-etcd-main    = "1-2f1"
    k8s-io-role-master  = "master"
  }
}

resource "google_compute_firewall" "cidr-to-master-minimal-gce-example-com" {
  name    = "cidr-to-master-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["443"]
  }

  allow = {
    protocol = "tcp"
    ports    = ["4194"]
  }

  source_ranges = ["100.64.0.0/10"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "cidr-to-node-minimal-gce-example-com" {
  name    = "cidr-to-node-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
  }

  allow = {
    protocol = "udp"
  }

  allow = {
    protocol = "icmp"
  }

  allow = {
    protocol = "esp"
  }

  allow = {
    protocol = "ah"
  }

  allow = {
    protocol = "sctp"
  }

  source_ranges = ["100.64.0.0/10"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "kubernetes-master-https-minimal-gce-example-com" {
  name    = "kubernetes-master-https-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-master-minimal-gce-example-com" {
  name    = "master-to-master-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
  }

  allow = {
    protocol = "udp"
  }

  allow = {
    protocol = "icmp"
  }

  allow = {
    protocol = "esp"
  }

  allow = {
    protocol = "ah"
  }

  allow = {
    protocol = "sctp"
  }

  source_tags = ["minimal-gce-example-com-k8s-io-role-master"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-node-minimal-gce-example-com" {
  name    = "master-to-node-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
  }

  allow = {
    protocol = "udp"
  }

  allow = {
    protocol = "icmp"
  }

  allow = {
    protocol = "esp"
  }

  allow = {
    protocol = "ah"
  }

  allow = {
    protocol = "sctp"
  }

  source_tags = ["minimal-gce-example-com-k8s-io-role-master"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "node-to-master-minimal-gce-example-com" {
  name    = "node-to-master-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["443"]
  }

  allow = {
    protocol = "tcp"
    ports    = ["4194"]
  }

  source_tags = ["minimal-gce-example-com-k8s-io-role-node"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "node-to-node-minimal-gce-example-com" {
  name    = "node-to-node-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
  }

  allow = {
    protocol = "udp"
  }

  allow = {
    protocol = "icmp"
  }

  allow = {
    protocol = "esp"
  }

  allow = {
    protocol = "ah"
  }

  allow = {
    protocol = "sctp"
  }

  source_tags = ["minimal-gce-example-com-k8s-io-role-node"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-minimal-gce-example-com" {
  name    = "nodeport-external-to-node-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["30000-32767"]
  }

  allow = {
    protocol = "udp"
    ports    = ["30000-32767"]
  }

  source_tags = ["minimal-gce-example-com-k8s-io-role-node"]
  target_tags = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-master-minimal-gce-example-com" {
  name    = "ssh-external-to-master-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-node-minimal-gce-example-com" {
  name    = "ssh-external-to-node-minimal-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_instance_group_manager" "a-master-us-test1-a-minimal-gce-example-com" {
  name               = "a-master-us-test1-a-minimal-gce-example-com"
  zone               = "us-test1-a"
  base_instance_name = "master-us-test1-a"

  version = {
    instance_template = "${google_compute_instance_template.master-us-test1-a-minimal-gce-example-com.self_link}"
  }

  target_size = 1
}

resource "google_compute_instance_group_manager" "a-nodes-minimal-gce-example-com" {
  name               = "a-nodes-minimal-gce-example-com"
  zone               = "us-test1-a"
  base_instance_name = "nodes"

  version = {
    instance_template = "${google_compute_instance_template.nodes-minimal-gce-example-com.self_link}"
  }

  target_size = 2
}

resource "google_compute_instance_template" "master-us-test1-a-minimal-gce-example-com" {
  can_ip_forward = true
  machine_type   = "n1-standard-1"

  service_account = {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }

  scheduling = {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }

  disk = {
    auto_delete  = true
    device_name  = "persistent-disks-0"
    type         = "PERSISTENT"
    boot         = true
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    mode         = "READ_WRITE"
    disk_type    = "pd-standard"
    disk_size_gb = 64
  }

  network_interface = {
    network       = "${google_compute_network.default.name}"
    access_config = {}
  }

  metadata = {
    cluster-name                    = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-a-minimal-gce-example-com_metadata_cluster-name")}"
    kops-k8s-io-instance-group-name = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-a-minimal-gce-example-com_metadata_kops-k8s-io-instance-group-name")}"
    ssh-keys                        = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-a-minimal-gce-example-com_metadata_ssh-keys")}"
    startup-script                  = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-a-minimal-gce-example-com_metadata_startup-script")}"
  }

  tags        = ["minimal-gce-example-com-k8s-io-role-master"]
  name_prefix = "master-us-test1-a-minimal-do16cp-"
}

resource "google_compute_instance_template" "nodes-minimal-gce-example-com" {
  can_ip_forward = true
  machine_type   = "n1-standard-2"

  service_account = {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_only"]
  }

  scheduling = {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
  }

  disk = {
    auto_delete  = true
    device_name  = "persistent-disks-0"
    type         = "PERSISTENT"
    boot         = true
    source_image = "https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/cos-stable-57-9202-64-0"
    mode         = "READ_WRITE"
    disk_type    = "pd-standard"
    disk_size_gb = 128
  }

  network_interface = {
    network       = "${google_compute_network.default.name}"
    access_config = {}
  }

  metadata = {
    cluster-name                    = "${file("${path.module}/data/google_compute_instance_template_nodes-minimal-gce-example-com_metadata_cluster-name")}"
    kops-k8s-io-instance-group-name = "${file("${path.module}/data/google_compute_instance_template_nodes-minimal-gce-example-com_metadata_kops-k8s-io-instance-group-name")}"
    ssh-keys                        = "${file("${path.module}/data/google_compute_instance_template_nodes-minimal-gce-example-com_metadata_ssh-keys")}"
    startup-script                  = "${file("${path.module}/data/google_compute_instance_template_nodes-minimal-gce-example-com_metadata_startup-script")}"
  }

  tags        = ["minimal-gce-example-com-k8s-io-role-node"]
  name_prefix = "nodes-minimal-gce-example-com-"
}

resource "google_compute_network" "default" {
  name                    = "default"
  auto_create_subnetworks = true
}

terraform = {
  required_version = ">= 0.9.3"
}
