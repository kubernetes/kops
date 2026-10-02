locals {
  cluster_name = "minimal-gce-multirole.example.com"
  project      = "testproject"
  region       = "us-test1"
}

output "cluster_name" {
  value = "minimal-gce-multirole.example.com"
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
  key                    = "tests/minimal-gce-multirole.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "tests/minimal-gce-multirole.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "tests/minimal-gce-multirole.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "tests/minimal-gce-multirole.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-channels-kops-channels" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-channels-kops-channels_content")
  key                    = "tests/minimal-gce-multirole.example.com/manifests/channels/kops-channels.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-etcd-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-etcd-us-test1-a_content")
  key                    = "tests/minimal-gce-multirole.example.com/manifests/etcd/events-etcd-us-test1-a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-etcd-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-etcd-us-test1-a_content")
  key                    = "tests/minimal-gce-multirole.example.com/manifests/etcd/main-etcd-us-test1-a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "tests/minimal-gce-multirole.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-bootstrap_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-gcp-cloud-controller-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-gcp-cloud-controller.addons.k8s.io-k8s-1.23_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/gcp-cloud-controller.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-gcp-pd-csi-driver-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-gcp-pd-csi-driver.addons.k8s.io-k8s-1.23_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/gcp-pd-csi-driver.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-gce-multirole-example-com-addons-storage-gce-addons-k8s-io-v1-7-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal-gce-multirole.example.com-addons-storage-gce.addons.k8s.io-v1.7.0_content")
  key                    = "tests/minimal-gce-multirole.example.com/addons/storage-gce.addons.k8s.io/v1.7.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-etcd-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-etcd-us-test1-a_content")
  key                    = "tests/minimal-gce-multirole.example.com/igconfig/etcd/etcd-us-test1-a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-external-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-external-us-test1-a_content")
  key                    = "tests/minimal-gce-multirole.example.com/igconfig/apiserver/external-us-test1-a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-internal-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-internal-us-test1-a_content")
  key                    = "tests/minimal-gce-multirole.example.com/igconfig/apiserver/internal-us-test1-a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-kcm-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-kcm-us-test1-a_content")
  key                    = "tests/minimal-gce-multirole.example.com/igconfig/apiserver_kubecontrollermanager/kcm-us-test1-a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "tests/minimal-gce-multirole.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-scheduler-us-test1-a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-scheduler-us-test1-a_content")
  key                    = "tests/minimal-gce-multirole.example.com/igconfig/apiserver_scheduler/scheduler-us-test1-a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "google_compute_address" "api-minimal-gce-multirole-example-com" {
  name = "api-minimal-gce-multirole-example-com"
}

resource "google_compute_address" "api-us-test1-minimal-gce-multirole-example-com" {
  address_type = "INTERNAL"
  name         = "api-us-test1-minimal-gce-multirole-example-com"
  purpose      = "SHARED_LOADBALANCER_VIP"
  subnetwork   = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
}

resource "google_compute_address" "etcd-us-test1-minimal-gce-multirole-example-com" {
  address_type = "INTERNAL"
  name         = "etcd-us-test1-minimal-gce-multirole-example-com"
  purpose      = "SHARED_LOADBALANCER_VIP"
  subnetwork   = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
}

resource "google_compute_disk" "a-etcd-events-minimal-gce-multirole-example-com" {
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-multirole-example-com"
    "k8s-io-etcd-events"  = "a-2fa"
    "k8s-io-role-master"  = "master"
  }
  name = "a-etcd-events-minimal-gce-multirole-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_disk" "a-etcd-main-minimal-gce-multirole-example-com" {
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-multirole-example-com"
    "k8s-io-etcd-main"    = "a-2fa"
    "k8s-io-role-master"  = "master"
  }
  name = "a-etcd-main-minimal-gce-multirole-example-com"
  size = 20
  type = "pd-ssd"
  zone = "us-test1-a"
}

resource "google_compute_firewall" "https-api-ipv6-minimal-gce-multirole-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "https-api-ipv6-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver"]
}

resource "google_compute_firewall" "https-api-minimal-gce-multirole-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "https-api-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver"]
}

resource "google_compute_firewall" "lb-health-checks-minimal-gce-multirole-example-com" {
  allow {
    protocol = "tcp"
  }
  disabled      = false
  name          = "lb-health-checks-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["35.191.0.0/16", "130.211.0.0/22", "209.85.204.0/22", "209.85.152.0/22"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver", "minimal-gce-multirole-example-com-k8s-io-role-etcd"]
}

resource "google_compute_firewall" "master-to-master-minimal-gce-multirole-example-com" {
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
  name        = "master-to-master-minimal-gce-multirole-example-com"
  network     = google_compute_network.minimal-gce-multirole-example-com.name
  source_tags = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver", "minimal-gce-multirole-example-com-k8s-io-role-etcd", "minimal-gce-multirole-example-com-k8s-io-role-scheduler", "minimal-gce-multirole--oio03a-k8s-io-role-kubecontrollermanager", "minimal-gce-multirole-example-com-k8s-io-role-master"]
  target_tags = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver", "minimal-gce-multirole-example-com-k8s-io-role-etcd", "minimal-gce-multirole-example-com-k8s-io-role-scheduler", "minimal-gce-multirole--oio03a-k8s-io-role-kubecontrollermanager", "minimal-gce-multirole-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "master-to-node-minimal-gce-multirole-example-com" {
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
  name        = "master-to-node-minimal-gce-multirole-example-com"
  network     = google_compute_network.minimal-gce-multirole-example-com.name
  source_tags = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver", "minimal-gce-multirole-example-com-k8s-io-role-master"]
  target_tags = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "node-to-master-minimal-gce-multirole-example-com" {
  allow {
    ports    = ["443"]
    protocol = "tcp"
  }
  allow {
    ports    = ["10250"]
    protocol = "tcp"
  }
  allow {
    ports    = ["3988"]
    protocol = "tcp"
  }
  allow {
    ports    = ["10257"]
    protocol = "tcp"
  }
  allow {
    ports    = ["10259"]
    protocol = "tcp"
  }
  allow {
    ports    = ["10249"]
    protocol = "tcp"
  }
  allow {
    ports    = ["2382"]
    protocol = "tcp"
  }
  allow {
    ports    = ["2384"]
    protocol = "tcp"
  }
  allow {
    ports    = ["9100"]
    protocol = "tcp"
  }
  disabled    = false
  name        = "node-to-master-minimal-gce-multirole-example-com"
  network     = google_compute_network.minimal-gce-multirole-example-com.name
  source_tags = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
  target_tags = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver", "minimal-gce-multirole-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "node-to-node-minimal-gce-multirole-example-com" {
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
  name        = "node-to-node-minimal-gce-multirole-example-com"
  network     = google_compute_network.minimal-gce-multirole-example-com.name
  source_tags = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
  target_tags = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-ipv6-minimal-gce-multirole-exa-oio03a" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  disabled      = true
  name          = "nodeport-external-to-node-ipv6-minimal-gce-multirole-exa-oio03a"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "nodeport-external-to-node-minimal-gce-multirole-example-com" {
  allow {
    ports    = ["30000-32767"]
    protocol = "tcp"
  }
  allow {
    ports    = ["30000-32767"]
    protocol = "udp"
  }
  disabled      = true
  name          = "nodeport-external-to-node-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-master-ipv6-minimal-gce-multirole-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-master-ipv6-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver", "minimal-gce-multirole-example-com-k8s-io-role-etcd", "minimal-gce-multirole-example-com-k8s-io-role-scheduler", "minimal-gce-multirole--oio03a-k8s-io-role-kubecontrollermanager", "minimal-gce-multirole-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-master-minimal-gce-multirole-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-master-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-control-plane", "minimal-gce-multirole-example-com-k8s-io-role-apiserver", "minimal-gce-multirole-example-com-k8s-io-role-etcd", "minimal-gce-multirole-example-com-k8s-io-role-scheduler", "minimal-gce-multirole--oio03a-k8s-io-role-kubecontrollermanager", "minimal-gce-multirole-example-com-k8s-io-role-master"]
}

resource "google_compute_firewall" "ssh-external-to-node-ipv6-minimal-gce-multirole-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-node-ipv6-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["::/0"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
}

resource "google_compute_firewall" "ssh-external-to-node-minimal-gce-multirole-example-com" {
  allow {
    ports    = ["22"]
    protocol = "tcp"
  }
  disabled      = false
  name          = "ssh-external-to-node-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
}

resource "google_compute_forwarding_rule" "api-minimal-gce-multirole-example-com" {
  ip_address  = google_compute_address.api-minimal-gce-multirole-example-com.address
  ip_protocol = "TCP"
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-multirole-example-com"
    "name"                = "api"
  }
  load_balancing_scheme = "EXTERNAL"
  name                  = "api-minimal-gce-multirole-example-com"
  port_range            = "443-443"
  target                = google_compute_target_pool.api-minimal-gce-multirole-example-com.self_link
}

resource "google_compute_forwarding_rule" "api-us-test1-minimal-gce-multirole-example-com" {
  backend_service = google_compute_region_backend_service.api-minimal-gce-multirole-example-com.id
  ip_address      = google_compute_address.api-us-test1-minimal-gce-multirole-example-com.address
  ip_protocol     = "TCP"
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-multirole-example-com"
    "name"                = "api-us-test1"
  }
  load_balancing_scheme = "INTERNAL"
  name                  = "api-us-test1-minimal-gce-multirole-example-com"
  network               = google_compute_network.minimal-gce-multirole-example-com.name
  ports                 = ["443"]
  subnetwork            = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
}

resource "google_compute_forwarding_rule" "etcd-us-test1-minimal-gce-multirole-example-com" {
  backend_service = google_compute_region_backend_service.etcd-minimal-gce-multirole-example-com.id
  ip_address      = google_compute_address.etcd-us-test1-minimal-gce-multirole-example-com.address
  ip_protocol     = "TCP"
  labels = {
    "k8s-io-cluster-name" = "minimal-gce-multirole-example-com"
    "name"                = "etcd-us-test1"
  }
  load_balancing_scheme = "INTERNAL"
  name                  = "etcd-us-test1-minimal-gce-multirole-example-com"
  network               = google_compute_network.minimal-gce-multirole-example-com.name
  ports                 = ["4001", "4002"]
  subnetwork            = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
}

resource "google_compute_http_health_check" "api-minimal-gce-multirole-example-com" {
  name         = "api-minimal-gce-multirole-example-com"
  port         = 3990
  request_path = "/healthz"
}

resource "google_compute_instance_group_manager" "a-etcd-us-test1-a-minimal-gce-multirole-example-com" {
  base_instance_name = "etcd-us-test1-a"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "a-etcd-us-test1-a-minimal-gce-multirole-example-com"
  target_size                    = 1
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.etcd-us-test1-a-minimal-gce-multirole-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-external-us-test1-a-minimal-gce-multirole-example-com" {
  base_instance_name = "external-us-test1-a"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "a-external-us-test1-a-minimal-gce-multirole-example-com"
  target_pools                   = [google_compute_target_pool.api-minimal-gce-multirole-example-com.self_link]
  target_size                    = 2
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.external-us-test1-a-minimal-gce-multirole-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-internal-us-test1-a-minimal-gce-multirole-example-com" {
  base_instance_name = "internal-us-test1-a"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "a-internal-us-test1-a-minimal-gce-multirole-example-com"
  target_size                    = 1
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.internal-us-test1-a-minimal-gce-multirole-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-kcm-us-test1-a-minimal-gce-multirole-example-com" {
  base_instance_name = "kcm-us-test1-a"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "a-kcm-us-test1-a-minimal-gce-multirole-example-com"
  target_size                    = 1
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.kcm-us-test1-a-minimal-gce-multirole-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-nodes-minimal-gce-multirole-example-com" {
  base_instance_name = "nodes"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "a-nodes-minimal-gce-multirole-example-com"
  target_size                    = 1
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.nodes-minimal-gce-multirole-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "a-scheduler-us-test1-a-minimal-gce-multirole-example-com" {
  base_instance_name = "scheduler-us-test1-a"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "a-scheduler-us-test1-a-minimal-gce-multirole-example-com"
  target_size                    = 1
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.scheduler-us-test1-a-minimal-gce-multirole-example-com.self_link
  }
  zone = "us-test1-a"
}

resource "google_compute_instance_group_manager" "b-nodes-minimal-gce-multirole-example-com" {
  base_instance_name = "nodes"
  lifecycle {
    ignore_changes = [target_size]
  }
  list_managed_instances_results = "PAGINATED"
  name                           = "b-nodes-minimal-gce-multirole-example-com"
  target_size                    = 1
  update_policy {
    minimal_action = "REPLACE"
    type           = "OPPORTUNISTIC"
  }
  version {
    instance_template = google_compute_instance_template.nodes-minimal-gce-multirole-example-com.self_link
  }
  zone = "us-test1-b"
}

resource "google_compute_instance_template" "etcd-us-test1-a-minimal-gce-multirole-example-com" {
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
    source_image           = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2604-resolute-amd64-v20221018"
    type                   = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "minimal-gce-multirole-example-com"
    "k8s-io-instance-group" = "etcd-us-test1-a"
    "k8s-io-role-etcd"      = "etcd"
  }
  lifecycle {
    create_before_destroy = true
  }
  machine_type = "e2-standard-2"
  metadata = {
    "cluster-name"                    = "minimal-gce-multirole.example.com"
    "kops-k8s-io-instance-group-name" = "etcd-us-test1-a"
    "ssh-keys"                        = "ubuntu: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "user-data"                       = file("${path.module}/data/google_compute_instance_template_etcd-us-test1-a-minimal-gce-multirole-example-com_metadata_user-data")
  }
  name_prefix = "etcd-us-test1-a-minimal-g-vh8n1r-"
  network_interface {
    network    = google_compute_network.minimal-gce-multirole-example-com.name
    stack_type = "IPV4_ONLY"
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["minimal-gce-multirole-example-com-k8s-io-role-etcd"]
}

resource "google_compute_instance_template" "external-us-test1-a-minimal-gce-multirole-example-com" {
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
    source_image           = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2604-resolute-amd64-v20221018"
    type                   = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "minimal-gce-multirole-example-com"
    "k8s-io-instance-group" = "external-us-test1-a"
    "k8s-io-role-apiserver" = "apiserver"
  }
  lifecycle {
    create_before_destroy = true
  }
  machine_type = "e2-standard-2"
  metadata = {
    "cluster-name"                    = "minimal-gce-multirole.example.com"
    "kops-k8s-io-instance-group-name" = "external-us-test1-a"
    "ssh-keys"                        = "ubuntu: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "user-data"                       = file("${path.module}/data/google_compute_instance_template_external-us-test1-a-minimal-gce-multirole-example-com_metadata_user-data")
  }
  name_prefix = "external-us-test1-a-minim-3jn048-"
  network_interface {
    network    = google_compute_network.minimal-gce-multirole-example-com.name
    stack_type = "IPV4_ONLY"
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["minimal-gce-multirole-example-com-k8s-io-role-apiserver"]
}

resource "google_compute_instance_template" "internal-us-test1-a-minimal-gce-multirole-example-com" {
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
    source_image           = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2604-resolute-amd64-v20221018"
    type                   = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "minimal-gce-multirole-example-com"
    "k8s-io-instance-group" = "internal-us-test1-a"
    "k8s-io-role-apiserver" = "apiserver"
  }
  lifecycle {
    create_before_destroy = true
  }
  machine_type = "e2-standard-2"
  metadata = {
    "cluster-name"                    = "minimal-gce-multirole.example.com"
    "kops-k8s-io-instance-group-name" = "internal-us-test1-a"
    "ssh-keys"                        = "ubuntu: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "user-data"                       = file("${path.module}/data/google_compute_instance_template_internal-us-test1-a-minimal-gce-multirole-example-com_metadata_user-data")
  }
  name_prefix = "internal-us-test1-a-minim-6jfiqh-"
  network_interface {
    network    = google_compute_network.minimal-gce-multirole-example-com.name
    stack_type = "IPV4_ONLY"
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["minimal-gce-multirole-example-com-k8s-io-role-apiserver"]
}

resource "google_compute_instance_template" "kcm-us-test1-a-minimal-gce-multirole-example-com" {
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
    source_image           = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2604-resolute-amd64-v20221018"
    type                   = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"               = "minimal-gce-multirole-example-com"
    "k8s-io-instance-group"             = "kcm-us-test1-a"
    "k8s-io-role-apiserver"             = "apiserver"
    "k8s-io-role-kubecontrollermanager" = "kubecontrollermanager"
  }
  lifecycle {
    create_before_destroy = true
  }
  machine_type = "e2-standard-2"
  metadata = {
    "cluster-name"                    = "minimal-gce-multirole.example.com"
    "kops-k8s-io-instance-group-name" = "kcm-us-test1-a"
    "ssh-keys"                        = "ubuntu: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "user-data"                       = file("${path.module}/data/google_compute_instance_template_kcm-us-test1-a-minimal-gce-multirole-example-com_metadata_user-data")
  }
  name_prefix = "kcm-us-test1-a-minimal-gc-6dthce-"
  network_interface {
    network    = google_compute_network.minimal-gce-multirole-example-com.name
    stack_type = "IPV4_ONLY"
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/devstorage.read_write", "https://www.googleapis.com/auth/ndev.clouddns.readwrite"]
  }
  tags = ["minimal-gce-multirole--oio03a-k8s-io-role-kubecontrollermanager"]
}

resource "google_compute_instance_template" "nodes-minimal-gce-multirole-example-com" {
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
    source_image           = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2604-resolute-amd64-v20221018"
    type                   = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "minimal-gce-multirole-example-com"
    "k8s-io-instance-group" = "nodes"
    "k8s-io-role-node"      = "node"
  }
  lifecycle {
    create_before_destroy = true
  }
  machine_type = "e2-standard-4"
  metadata = {
    "cluster-name"                    = "minimal-gce-multirole.example.com"
    "kops-k8s-io-instance-group-name" = "nodes"
    "kube-env"                        = "AUTOSCALER_ENV_VARS: os_distribution=ubuntu;arch=amd64;os=linux"
    "ssh-keys"                        = "ubuntu: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "user-data"                       = file("${path.module}/data/google_compute_instance_template_nodes-minimal-gce-multirole-example-com_metadata_user-data")
  }
  name_prefix = "nodes-minimal-gce-multiro-oj8ive-"
  network_interface {
    network    = google_compute_network.minimal-gce-multirole-example-com.name
    stack_type = "IPV4_ONLY"
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/devstorage.read_only"]
  }
  tags = ["minimal-gce-multirole-example-com-k8s-io-role-node"]
}

resource "google_compute_instance_template" "scheduler-us-test1-a-minimal-gce-multirole-example-com" {
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
    source_image           = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2604-resolute-amd64-v20221018"
    type                   = "PERSISTENT"
  }
  labels = {
    "k8s-io-cluster-name"   = "minimal-gce-multirole-example-com"
    "k8s-io-instance-group" = "scheduler-us-test1-a"
    "k8s-io-role-apiserver" = "apiserver"
    "k8s-io-role-scheduler" = "scheduler"
  }
  lifecycle {
    create_before_destroy = true
  }
  machine_type = "e2-standard-2"
  metadata = {
    "cluster-name"                    = "minimal-gce-multirole.example.com"
    "kops-k8s-io-instance-group-name" = "scheduler-us-test1-a"
    "ssh-keys"                        = "ubuntu: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtWu40XQo8dczLsCq0OWV+hxm9uV3WxeH9Kgh4sMzQxNtoU1pvW0XdjpkBesRKGoolfWeCLXWxpyQb1IaiMkKoz7MdhQ/6UKjMjP66aFWWp3pwD0uj0HuJ7tq4gKHKRYGTaZIRWpzUiANBrjugVgA+Sd7E/mYwc/DMXkIyRZbvhQ=="
    "user-data"                       = file("${path.module}/data/google_compute_instance_template_scheduler-us-test1-a-minimal-gce-multirole-example-com_metadata_user-data")
  }
  name_prefix = "scheduler-us-test1-a-mini-opijmc-"
  network_interface {
    network    = google_compute_network.minimal-gce-multirole-example-com.name
    stack_type = "IPV4_ONLY"
    subnetwork = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }
  service_account {
    email  = "default"
    scopes = ["https://www.googleapis.com/auth/compute", "https://www.googleapis.com/auth/monitoring", "https://www.googleapis.com/auth/logging.write", "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/devstorage.read_write"]
  }
  tags = ["minimal-gce-multirole-example-com-k8s-io-role-scheduler"]
}

resource "google_compute_network" "minimal-gce-multirole-example-com" {
  auto_create_subnetworks = false
  name                    = "minimal-gce-multirole-example-com"
}

resource "google_compute_region_backend_service" "api-minimal-gce-multirole-example-com" {
  backend {
    balancing_mode = "CONNECTION"
    group          = google_compute_instance_group_manager.a-internal-us-test1-a-minimal-gce-multirole-example-com.instance_group
  }
  health_checks         = [google_compute_region_health_check.api-minimal-gce-multirole-example-com.id]
  load_balancing_scheme = "INTERNAL"
  name                  = "api-minimal-gce-multirole-example-com"
  protocol              = "TCP"
}

resource "google_compute_region_backend_service" "etcd-minimal-gce-multirole-example-com" {
  backend {
    balancing_mode = "CONNECTION"
    group          = google_compute_instance_group_manager.a-etcd-us-test1-a-minimal-gce-multirole-example-com.instance_group
  }
  health_checks         = [google_compute_region_health_check.etcd-main-minimal-gce-multirole-example-com.id]
  load_balancing_scheme = "INTERNAL"
  name                  = "etcd-minimal-gce-multirole-example-com"
  protocol              = "TCP"
}

resource "google_compute_region_backend_service" "kops-controller-minimal-gce-multirole-example-com" {
  backend {
    balancing_mode = "CONNECTION"
    group          = google_compute_instance_group_manager.a-internal-us-test1-a-minimal-gce-multirole-example-com.instance_group
  }
  health_checks         = [google_compute_region_health_check.kops-controller-minimal-gce-multirole-example-com.id]
  load_balancing_scheme = "INTERNAL"
  name                  = "kops-controller-minimal-gce-multirole-example-com"
  protocol              = "TCP"
}

resource "google_compute_region_health_check" "api-minimal-gce-multirole-example-com" {
  name = "api-minimal-gce-multirole-example-com"
  tcp_health_check {
    port = 443
  }
}

resource "google_compute_region_health_check" "etcd-main-minimal-gce-multirole-example-com" {
  name = "etcd-main-minimal-gce-multirole-example-com"
  tcp_health_check {
    port = 4001
  }
}

resource "google_compute_region_health_check" "kops-controller-minimal-gce-multirole-example-com" {
  name = "kops-controller-minimal-gce-multirole-example-com"
  ssl_health_check {
    port = 3988
  }
}

resource "google_compute_router" "nat-minimal-gce-multirole-example-com" {
  name    = "nat-minimal-gce-multirole-example-com"
  network = google_compute_network.minimal-gce-multirole-example-com.name
}

resource "google_compute_router_nat" "nat-minimal-gce-multirole-example-com" {
  name                               = "nat-minimal-gce-multirole-example-com"
  nat_ip_allocate_option             = "AUTO_ONLY"
  region                             = "us-test1"
  router                             = google_compute_router.nat-minimal-gce-multirole-example-com.name
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"
  subnetwork {
    name                    = google_compute_subnetwork.us-test1-minimal-gce-multirole-example-com.name
    source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
  }
}

resource "google_compute_subnetwork" "us-test1-minimal-gce-multirole-example-com" {
  ip_cidr_range = "10.0.16.0/20"
  name          = "us-test1-minimal-gce-multirole-example-com"
  network       = google_compute_network.minimal-gce-multirole-example-com.name
  region        = "us-test1"
  stack_type    = "IPV4_ONLY"
}

resource "google_compute_target_pool" "api-minimal-gce-multirole-example-com" {
  health_checks = [google_compute_http_health_check.api-minimal-gce-multirole-example-com.self_link]
  name          = "api-minimal-gce-multirole-example-com"
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 6.57.1"
    }
    google = {
      "source"  = "hashicorp/google"
      "version" = ">= 6.8.0"
    }
  }
}
