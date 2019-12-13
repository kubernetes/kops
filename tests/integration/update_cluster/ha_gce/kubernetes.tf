locals = {
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
  project = "us-test1"
  region  = "us-test1"
}

resource "google_compute_disk" "d1-etcd-events-ha-gce-example-com" {
  name = "d1-etcd-events-ha-gce-example-com"
  type = "pd-ssd"
  size = 20
  zone = "us-test1-a"

  labels = {
    k8s-io-cluster-name = "ha-gce-example-com"
    k8s-io-etcd-events  = "1-2f1-2c2-2c3"
    k8s-io-role-master  = "master"
  }
}

resource "google_compute_disk" "d1-etcd-main-ha-gce-example-com" {
  name = "d1-etcd-main-ha-gce-example-com"
  type = "pd-ssd"
  size = 20
  zone = "us-test1-a"

  labels = {
    k8s-io-cluster-name = "ha-gce-example-com"
    k8s-io-etcd-main    = "1-2f1-2c2-2c3"
    k8s-io-role-master  = "master"
  }
}

resource "google_compute_disk" "d2-etcd-events-ha-gce-example-com" {
  name = "d2-etcd-events-ha-gce-example-com"
  type = "pd-ssd"
  size = 20
  zone = "us-test1-b"

  labels = {
    k8s-io-cluster-name = "ha-gce-example-com"
    k8s-io-etcd-events  = "2-2f1-2c2-2c3"
    k8s-io-role-master  = "master"
  }
}

resource "google_compute_disk" "d2-etcd-main-ha-gce-example-com" {
  name = "d2-etcd-main-ha-gce-example-com"
  type = "pd-ssd"
  size = 20
  zone = "us-test1-b"

  labels = {
    k8s-io-cluster-name = "ha-gce-example-com"
    k8s-io-etcd-main    = "2-2f1-2c2-2c3"
    k8s-io-role-master  = "master"
  }
}

resource "google_compute_disk" "d3-etcd-events-ha-gce-example-com" {
  name = "d3-etcd-events-ha-gce-example-com"
  type = "pd-ssd"
  size = 20
  zone = "us-test1-c"

  labels = {
    k8s-io-cluster-name = "ha-gce-example-com"
    k8s-io-etcd-events  = "3-2f1-2c2-2c3"
    k8s-io-role-master  = "master"
  }
}

resource "google_compute_disk" "d3-etcd-main-ha-gce-example-com" {
  name = "d3-etcd-main-ha-gce-example-com"
  type = "pd-ssd"
  size = 20
  zone = "us-test1-c"

  labels = {
    k8s-io-cluster-name = "ha-gce-example-com"
    k8s-io-etcd-main    = "3-2f1-2c2-2c3"
    k8s-io-role-master  = "master"
  }
}

resource "google_compute_firewall" "cidr-to-master-ha-gce-example-com" {
  name    = "cidr-to-master-ha-gce-example-com"
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
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "cidr-to-node-ha-gce-example-com" {
  name    = "cidr-to-node-ha-gce-example-com"
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
  target_tags   = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "kubernetes-master-https-ha-gce-example-com" {
  name    = "kubernetes-master-https-ha-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-master-ha-gce-example-com" {
  name    = "master-to-master-ha-gce-example-com"
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

  source_tags = ["ha-gce-example-com-k8s-io-role-master"]
  target_tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-node-ha-gce-example-com" {
  name    = "master-to-node-ha-gce-example-com"
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

  source_tags = ["ha-gce-example-com-k8s-io-role-master"]
  target_tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "node-to-master-ha-gce-example-com" {
  name    = "node-to-master-ha-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["443"]
  }

  allow = {
    protocol = "tcp"
    ports    = ["4194"]
  }

  source_tags = ["ha-gce-example-com-k8s-io-role-node"]
  target_tags = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "node-to-node-ha-gce-example-com" {
  name    = "node-to-node-ha-gce-example-com"
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

  source_tags = ["ha-gce-example-com-k8s-io-role-node"]
  target_tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-ha-gce-example-com" {
  name    = "nodeport-external-to-node-ha-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["30000-32767"]
  }

  allow = {
    protocol = "udp"
    ports    = ["30000-32767"]
  }

  source_tags = ["ha-gce-example-com-k8s-io-role-node"]
  target_tags = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-master-ha-gce-example-com" {
  name    = "ssh-external-to-master-ha-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-node-ha-gce-example-com" {
  name    = "ssh-external-to-node-ha-gce-example-com"
  network = "${google_compute_network.default.name}"

  allow = {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ha-gce-example-com-k8s-io-role-node"]
}

resource "google_compute_instance_group_manager" "a-master-us-test1-a-ha-gce-example-com" {
  name               = "a-master-us-test1-a-ha-gce-example-com"
  zone               = "us-test1-a"
  base_instance_name = "master-us-test1-a"

  version = {
    instance_template = "${google_compute_instance_template.master-us-test1-a-ha-gce-example-com.self_link}"
  }

  target_size = 1
}

resource "google_compute_instance_group_manager" "a-nodes-ha-gce-example-com" {
  name               = "a-nodes-ha-gce-example-com"
  zone               = "us-test1-a"
  base_instance_name = "nodes"

  version = {
    instance_template = "${google_compute_instance_template.nodes-ha-gce-example-com.self_link}"
  }

  target_size = 1
}

resource "google_compute_instance_group_manager" "b-master-us-test1-b-ha-gce-example-com" {
  name               = "b-master-us-test1-b-ha-gce-example-com"
  zone               = "us-test1-b"
  base_instance_name = "master-us-test1-b"

  version = {
    instance_template = "${google_compute_instance_template.master-us-test1-b-ha-gce-example-com.self_link}"
  }

  target_size = 1
}

resource "google_compute_instance_group_manager" "b-nodes-ha-gce-example-com" {
  name               = "b-nodes-ha-gce-example-com"
  zone               = "us-test1-b"
  base_instance_name = "nodes"

  version = {
    instance_template = "${google_compute_instance_template.nodes-ha-gce-example-com.self_link}"
  }

  target_size = 1
}

resource "google_compute_instance_group_manager" "c-master-us-test1-c-ha-gce-example-com" {
  name               = "c-master-us-test1-c-ha-gce-example-com"
  zone               = "us-test1-c"
  base_instance_name = "master-us-test1-c"

  version = {
    instance_template = "${google_compute_instance_template.master-us-test1-c-ha-gce-example-com.self_link}"
  }

  target_size = 1
}

resource "google_compute_instance_group_manager" "c-nodes-ha-gce-example-com" {
  name               = "c-nodes-ha-gce-example-com"
  zone               = "us-test1-c"
  base_instance_name = "nodes"

  version = {
    instance_template = "${google_compute_instance_template.nodes-ha-gce-example-com.self_link}"
  }

  target_size = 0
}

resource "google_compute_instance_template" "master-us-test1-a-ha-gce-example-com" {
  can_ip_forward = true
  machine_type   = "n1-standard-1"

  service_account = {
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
    cluster-name                    = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_cluster-name")}"
    kops-k8s-io-instance-group-name = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_kops-k8s-io-instance-group-name")}"
    ssh-keys                        = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_ssh-keys")}"
    startup-script                  = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-a-ha-gce-example-com_metadata_startup-script")}"
  }

  tags        = ["ha-gce-example-com-k8s-io-role-master"]
  name_prefix = "master-us-test1-a-ha-gce--ke5ah6-"
}

resource "google_compute_instance_template" "master-us-test1-b-ha-gce-example-com" {
  can_ip_forward = true
  machine_type   = "n1-standard-1"

  service_account = {
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
    cluster-name                    = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_cluster-name")}"
    kops-k8s-io-instance-group-name = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_kops-k8s-io-instance-group-name")}"
    ssh-keys                        = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_ssh-keys")}"
    startup-script                  = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-b-ha-gce-example-com_metadata_startup-script")}"
  }

  tags        = ["ha-gce-example-com-k8s-io-role-master"]
  name_prefix = "master-us-test1-b-ha-gce--c8u7qq-"
}

resource "google_compute_instance_template" "master-us-test1-c-ha-gce-example-com" {
  can_ip_forward = true
  machine_type   = "n1-standard-1"

  service_account = {
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
    cluster-name                    = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_cluster-name")}"
    kops-k8s-io-instance-group-name = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_kops-k8s-io-instance-group-name")}"
    ssh-keys                        = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_ssh-keys")}"
    startup-script                  = "${file("${path.module}/data/google_compute_instance_template_master-us-test1-c-ha-gce-example-com_metadata_startup-script")}"
  }

  tags        = ["ha-gce-example-com-k8s-io-role-master"]
  name_prefix = "master-us-test1-c-ha-gce--3unp7l-"
}

resource "google_compute_instance_template" "nodes-ha-gce-example-com" {
  can_ip_forward = true
  machine_type   = "n1-standard-2"

  service_account = {
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
    cluster-name                    = "${file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_cluster-name")}"
    kops-k8s-io-instance-group-name = "${file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_kops-k8s-io-instance-group-name")}"
    ssh-keys                        = "${file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_ssh-keys")}"
    startup-script                  = "${file("${path.module}/data/google_compute_instance_template_nodes-ha-gce-example-com_metadata_startup-script")}"
  }

  tags        = ["ha-gce-example-com-k8s-io-role-node"]
  name_prefix = "nodes-ha-gce-example-com-"
}

resource "google_compute_network" "default" {
  name                    = "default"
  auto_create_subnetworks = true
}

terraform = {
  required_version = ">= 0.9.3"
}
