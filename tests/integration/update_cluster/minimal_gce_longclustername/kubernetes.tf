locals {
  cluster_name = "minimal-gce-with-a-very-very-very-very-very-long-name.example.com"
  project      = "testproject"
  region       = "us-test1"
}

output "cluster_name" {
  value = "minimal-gce-with-a-very-very-very-very-very-long-name.example.com"
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
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-master-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-master-us-test1-a_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/manifests/etcd/events-master-us-test1-a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-master-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-master-us-test1-a_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/manifests/etcd/main-master-us-test1-a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-bootstrap_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-gcp-pd-csi-driver-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-gcp-pd-csi-driver.addons.k8s.io-k8s-1.23_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/gcp-pd-csi-driver.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-leader-migration-rbac-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-leader-migration.rbac.addons.k8s.io-k8s-1.23_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/leader-migration.rbac.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-metadata-proxy-addons-k8s-io-v0-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-metadata-proxy.addons.k8s.io-v0.1.12_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/metadata-proxy.addons.k8s.io/v0.1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-with-a-very-very-very-very-very-long-name-example-com-addons-storage-gce-addons-k8s-io-v1-7-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-with-a-very-very-very-very-very-long-name.example.com-addons-storage-gce.addons.k8s.io-v1.7.0_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/addons/storage-gce.addons.k8s.io/v1.7.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test1-a_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/igconfig/master/master-us-test1-a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "tests/minimal-gce-with-a-very-very-very-very-very-long-name.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "google_compute_disk" "d1-etcd-events-minimal-gce-with-a-very-very-very-very-ve-96dqvi" {
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-with-a-very-very-very-very-very-long-name-example-com"
    "k8s-io-etcd-events"  = "1-2f1"
    "k8s-io-role-master"  = "master"
  }
  name = "d1-etcd-events-minimal-gce-with-a-very-very-very-very-ve-96dqvi"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_disk" "d1-etcd-main-minimal-gce-with-a-very-very-very-very-very-96dqvi" {
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-with-a-very-very-very-very-very-long-name-example-com"
    "k8s-io-etcd-main"    = "1-2f1"
    "k8s-io-role-master"  = "master"
  }
  name = "d1-etcd-main-minimal-gce-with-a-very-very-very-very-very-96dqvi"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_firewall" "kubernetes-master-https-ipv6-minimal-gce-with-a-very-ver-96dqvi" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled      = true
  name          = "kubernetes-master-https-ipv6-minimal-gce-with-a-very-ver-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
}

resource "google_compute_firewall" "kubernetes-master-https-minimal-gce-with-a-very-very-ver-96dqvi" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "kubernetes-master-https-minimal-gce-with-a-very-very-ver-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-master-minimal-gce-with-a-very-very-very-very--96dqvi" {
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
  disabled    = false
  name        = "master-to-master-minimal-gce-with-a-very-very-very-very--96dqvi"
  network     = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_tags = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
  target_tags = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-node-minimal-gce-with-a-very-very-very-very-ve-96dqvi" {
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
  disabled    = false
  name        = "master-to-node-minimal-gce-with-a-very-very-very-very-ve-96dqvi"
  network     = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_tags = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
  target_tags = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
}

resource "google_compute_firewall" "node-to-master-minimal-gce-with-a-very-very-very-very-ve-96dqvi" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  allow {
    ports    = ["3988"]
    protocol = "tcp"
  }
  disabled    = false
  name        = "node-to-master-minimal-gce-with-a-very-very-very-very-ve-96dqvi"
  network     = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_tags = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
  target_tags = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
}

resource "google_compute_firewall" "node-to-node-minimal-gce-with-a-very-very-very-very-very-96dqvi" {
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
  disabled    = false
  name        = "node-to-node-minimal-gce-with-a-very-very-very-very-very-96dqvi"
  network     = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_tags = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
  target_tags = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-ipv6-minimal-gce-with-a-very-v-96dqvi" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  disabled      = true
  name          = "nodeport-external-to-node-ipv6-minimal-gce-with-a-very-v-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-minimal-gce-with-a-very-very-v-96dqvi" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  disabled      = true
  name          = "nodeport-external-to-node-minimal-gce-with-a-very-very-v-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-master-ipv6-minimal-gce-with-a-very-very-96dqvi" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = true
  name          = "ssh-external-to-master-ipv6-minimal-gce-with-a-very-very-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-master-minimal-gce-with-a-very-very-very-96dqvi" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-master-minimal-gce-with-a-very-very-very-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-node-ipv6-minimal-gce-with-a-very-very-v-96dqvi" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = true
  name          = "ssh-external-to-node-ipv6-minimal-gce-with-a-very-very-v-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-node-minimal-gce-with-a-very-very-very-v-96dqvi" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-node-minimal-gce-with-a-very-very-very-v-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
}

resource "google_compute_instance_group_manager" "a-master-us-test1-a-minimal-gce-with-a-very-very-very-ve-j0fh8f" {
  base_instance_name = "master-us-test1-a"
  name               = "a-master-us-test1-a-minimal-gce-with-a-very-very-very-ve-j0fh8f"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.master-us-test1-a-minimal-gce-with-a-very-very-very-very-very-long-name-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-nodes-minimal-gce-with-a-very-very-very-very-very-long-qk78uj" {
  base_instance_name = "nodes"
  name               = "a-nodes-minimal-gce-with-a-very-very-very-very-very-long-qk78uj"
  target_size        = 2
  version {
    instance_template = google_compute_instance_template.nodes-minimal-gce-with-a-very-very-very-very-very-long-name-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_template" "master-us-test1-a-minimal-gce-with-a-very-very-very-very-very-long-name-example-com" {
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
  labels = {
    "k8s-io-cluster-name"   = "minimal-gce-with-a-very-very-very-very-very-long-name-example-com"
    "k8s-io-instance-group" = "master-us-test1-a"
    "k8s-io-role-master"    = ""
  }
  machine_type = "n1-standard-1"
  metadata = {
    "cluster-name"                    = "minimal-gce-with-a-very-very-very-very-very-long-name.example.com"
    "kops-k8s-io-instance-group-name" = "master-us-test1-a"
    "ssh-keys"                        = "admin: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_master-us-test1-a-minimal-gce-with-a-very-very-very-very-very-long-name-example-com_metadata_startup-script")
  }
  name_prefix = "master-us-test1-a-minimal-ivl9ll-"
  network_interface {
    access_config {
    }
    network    = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-with-a-very-very-very-very-very-lon-96dqvi.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = google_service_account.control-plane.email
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["minimal-gce-with-a-very-very-very-ver-96dqvi-k8s-io-role-master"]
}

resource "google_compute_instance_template" "nodes-minimal-gce-with-a-very-very-very-very-very-long-name-example-com" {
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
  labels = {
    "k8s-io-cluster-name"   = "minimal-gce-with-a-very-very-very-very-very-long-name-example-com"
    "k8s-io-instance-group" = "nodes"
    "k8s-io-role-node"      = ""
  }
  machine_type = "n1-standard-2"
  metadata = {
    "cluster-name"                    = "minimal-gce-with-a-very-very-very-very-very-long-name.example.com"
    "kops-k8s-io-instance-group-name" = "nodes"
    "ssh-keys"                        = "admin: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "startup-script"                  = file("${path.module}/data/google_compute_instance_template_nodes-minimal-gce-with-a-very-very-very-very-very-long-name-example-com_metadata_startup-script")
  }
  name_prefix = "nodes-minimal-gce-with-a--k0ql96-"
  network_interface {
    access_config {
    }
    network    = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-with-a-very-very-very-very-very-lon-96dqvi.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = google_service_account.node.email
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/devstorage.read_only"]
  }
  tags = ["minimal-gce-with-a-very-very-very-very--96dqvi-k8s-io-role-node"]
}

resource "google_compute_network" "minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi" {
  auto_create_subnetworks = false
  name                    = "minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi"
}

resource "google_compute_subnetwork" "us-test1-minimal-gce-with-a-very-very-very-very-very-lon-96dqvi" {
  ip_cidr_range = "10.0.16.0/20"
  name          = "us-test1-minimal-gce-with-a-very-very-very-very-very-lon-96dqvi"
  network       = google_compute_network.minimal-gce-with-a-very-very-very-very-very-long-name-ex-96dqvi.name
  region        = "us-test1"
}

resource "google_project_iam_binding" "serviceaccount-control-plane" {
  members = ["serviceAccount:control-plane-minimal-g-96dqvi@testproject.iam.gserviceaccount.com"]
  project = "testproject"
  role    = "roles/container.serviceAgent"
}

resource "google_project_iam_binding" "serviceaccount-nodes" {
  members = ["serviceAccount:node-minimal-gce-with-a-96dqvi@testproject.iam.gserviceaccount.com"]
  project = "testproject"
  role    = "roles/compute.viewer"
}

resource "google_service_account" "control-plane" {
  account_id   = "control-plane-minimal-g-96dqvi"
  description  = "kubernetes control-plane instances"
  display_name = "control-plane"
  project      = "testproject"
}

resource "google_service_account" "node" {
  account_id   = "node-minimal-gce-with-a-96dqvi"
  description  = "kubernetes worker nodes"
  display_name = "node"
  project      = "testproject"
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    google = {
      "source"  = "hashicorp/google"
      "version" = ">= 2.19.0"
    }
  }
}
