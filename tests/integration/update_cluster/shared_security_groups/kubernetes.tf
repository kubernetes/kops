output "cluster_name" {
  value = "sec-groups.example.com"
}

output "master_security_group_ids" {
  value = ["sg-c4d4a3b7", "sg-c4d4a3b8"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-sec-groups-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-sec-groups-example-com.name}"
}

output "node_security_group_ids" {
  value = ["sg-15eb9c66"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-sec-groups-example-com.id}"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-sec-groups-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-sec-groups-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

output "vpc_id" {
  value = "vpc-0301e07b"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-sec-groups-example-com" {
  name                 = "master-us-test-1a.masters.sec-groups.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-sec-groups-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-sec-groups-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "nodes-sec-groups-example-com" {
  name                 = "nodes.sec-groups.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-sec-groups-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-sec-groups-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_ebs_volume" "d-etcd-events-sec-groups-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "sec-groups.example.com"
    Name                 = "d.etcd-events.sec-groups.example.com"
    "k8s.io/etcd/events" = "d/d"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "d-etcd-main-sec-groups-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "sec-groups.example.com"
    Name                 = "d.etcd-main.sec-groups.example.com"
    "k8s.io/etcd/main"   = "d/d"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_iam_instance_profile" "masters-sec-groups-example-com" {
  name = "masters.sec-groups.example.com"
  role = "${aws_iam_role.masters-sec-groups-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-sec-groups-example-com" {
  name = "nodes.sec-groups.example.com"
  role = "${aws_iam_role.nodes-sec-groups-example-com.name}"
}

resource "aws_iam_role" "masters-sec-groups-example-com" {
  name               = "masters.sec-groups.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.sec-groups.example.com_policy")}"
}

resource "aws_iam_role" "nodes-sec-groups-example-com" {
  name               = "nodes.sec-groups.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.sec-groups.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-sec-groups-example-com" {
  name   = "masters.sec-groups.example.com"
  role   = "${aws_iam_role.masters-sec-groups-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.sec-groups.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-sec-groups-example-com" {
  name   = "nodes.sec-groups.example.com"
  role   = "${aws_iam_role.nodes-sec-groups-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.sec-groups.example.com_policy")}"
}

resource "aws_key_pair" "kubernetes-sec-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.sec-groups.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.sec-groups.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "master-us-test-1a-masters-sec-groups-example-com" {
  name_prefix                 = "master-us-test-1a.masters.sec-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m4.large"
  key_name                    = "${aws_key_pair.kubernetes-sec-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-sec-groups-example-com.id}"
  security_groups             = ["sg-c4d4a3b7", "sg-c4d4a3b8"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.sec-groups.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 64
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }

  spot_price = "0.4"
}

resource "aws_launch_configuration" "nodes-sec-groups-example-com" {
  name_prefix                 = "nodes.sec-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m4.large"
  key_name                    = "${aws_key_pair.kubernetes-sec-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-sec-groups-example-com.id}"
  security_groups             = ["sg-15eb9c66"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.sec-groups.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 128
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }

  spot_price = "0.4"
}

resource "aws_route" "0-0-0-0--0" {
  route_table_id         = "${aws_route_table.sec-groups-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "fake-ig"
}

resource "aws_route_table" "sec-groups-example-com" {
  vpc_id = "vpc-0301e07b"

  tags = {
    KubernetesCluster = "sec-groups.example.com"
    Name              = "sec-groups.example.com"
  }
}

resource "aws_route_table_association" "us-test-1a-sec-groups-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-sec-groups-example-com.id}"
  route_table_id = "${aws_route_table.sec-groups-example-com.id}"
}

resource "aws_subnet" "us-test-1a-sec-groups-example-com" {
  vpc_id            = "vpc-0301e07b"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                              = "sec-groups.example.com"
    Name                                           = "us-test-1a.sec-groups.example.com"
    "kubernetes.io/cluster/sec-groups.example.com" = "owned"
  }
}

terraform = {
  required_version = ">= 0.9.3"
}
