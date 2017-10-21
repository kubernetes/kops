output "bastions_role_arn" {
  value = "${aws_iam_role.bastions-privateweave-example-com.arn}"
}

output "bastions_role_name" {
  value = "${aws_iam_role.bastions-privateweave-example-com.name}"
}

output "cluster_name" {
  value = "privateweave.example.com"
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-privateweave-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-privateweave-example-com.name}"
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-privateweave-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-privateweave-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_iam_instance_profile" "bastions-privateweave-example-com" {
  name = "bastions.privateweave.example.com"
  role = "${aws_iam_role.bastions-privateweave-example-com.name}"
}

resource "aws_iam_instance_profile" "masters-privateweave-example-com" {
  name = "masters.privateweave.example.com"
  role = "${aws_iam_role.masters-privateweave-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-privateweave-example-com" {
  name = "nodes.privateweave.example.com"
  role = "${aws_iam_role.nodes-privateweave-example-com.name}"
}

resource "aws_iam_role" "bastions-privateweave-example-com" {
  name               = "bastions.privateweave.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_bastions.privateweave.example.com_policy")}"
}

resource "aws_iam_role" "masters-privateweave-example-com" {
  name               = "masters.privateweave.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.privateweave.example.com_policy")}"
}

resource "aws_iam_role" "nodes-privateweave-example-com" {
  name               = "nodes.privateweave.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.privateweave.example.com_policy")}"
}

resource "aws_iam_role_policy" "bastions-privateweave-example-com" {
  name   = "bastions.privateweave.example.com"
  role   = "${aws_iam_role.bastions-privateweave-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_bastions.privateweave.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-privateweave-example-com" {
  name   = "masters.privateweave.example.com"
  role   = "${aws_iam_role.masters-privateweave-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.privateweave.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-privateweave-example-com" {
  name   = "nodes.privateweave.example.com"
  role   = "${aws_iam_role.nodes-privateweave-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.privateweave.example.com_policy")}"
}

resource "aws_key_pair" "kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.privateweave.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.privateweave.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

terraform = {
  required_version = ">= 0.9.3"
}
