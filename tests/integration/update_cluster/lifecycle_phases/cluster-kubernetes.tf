locals = {
  bastion_security_group_ids = ["${aws_security_group.bastion-lifecyclephases-example-com.id}"]
  bastions_role_arn          = "${aws_iam_role.bastions-lifecyclephases-example-com.arn}"
  bastions_role_name         = "${aws_iam_role.bastions-lifecyclephases-example-com.name}"
  cluster_name               = "lifecyclephases.example.com"
  master_security_group_ids  = ["${aws_security_group.masters-lifecyclephases-example-com.id}"]
  masters_role_arn           = "${aws_iam_role.masters-lifecyclephases-example-com.arn}"
  masters_role_name          = "${aws_iam_role.masters-lifecyclephases-example-com.name}"
  node_security_group_ids    = ["${aws_security_group.nodes-lifecyclephases-example-com.id}"]
  node_subnet_ids            = ["${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"]
  nodes_role_arn             = "${aws_iam_role.nodes-lifecyclephases-example-com.arn}"
  nodes_role_name            = "${aws_iam_role.nodes-lifecyclephases-example-com.name}"
  region                     = "us-test-1"
  vpc_id                     = "${aws_vpc.lifecyclephases-example-com.id}"
}

output "bastion_security_group_ids" {
  value = ["${aws_security_group.bastion-lifecyclephases-example-com.id}"]
}

output "bastions_role_arn" {
  value = "${aws_iam_role.bastions-lifecyclephases-example-com.arn}"
}

output "bastions_role_name" {
  value = "${aws_iam_role.bastions-lifecyclephases-example-com.name}"
}

output "cluster_name" {
  value = "lifecyclephases.example.com"
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-lifecyclephases-example-com.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-lifecyclephases-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-lifecyclephases-example-com.name}"
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-lifecyclephases-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-lifecyclephases-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-lifecyclephases-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

output "vpc_id" {
  value = "${aws_vpc.lifecyclephases-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "bastion-lifecyclephases-example-com" {
  name                 = "bastion.lifecyclephases.example.com"
  launch_configuration = "${aws_launch_configuration.bastion-lifecyclephases-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.utility-us-test-1a-lifecyclephases-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "lifecyclephases.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "bastion.lifecyclephases.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/bastion"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-lifecyclephases-example-com" {
  name                 = "master-us-test-1a.masters.lifecyclephases.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-lifecyclephases-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "lifecyclephases.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.lifecyclephases.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "nodes-lifecyclephases-example-com" {
  name                 = "nodes.lifecyclephases.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-lifecyclephases-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "lifecyclephases.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.lifecyclephases.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-lifecyclephases-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "lifecyclephases.example.com"
    Name                 = "us-test-1a.etcd-events.lifecyclephases.example.com"
    "k8s.io/etcd/events" = "us-test-1a/us-test-1a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-lifecyclephases-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "lifecyclephases.example.com"
    Name                 = "us-test-1a.etcd-main.lifecyclephases.example.com"
    "k8s.io/etcd/main"   = "us-test-1a/us-test-1a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_launch_configuration" "bastion-lifecyclephases-example-com" {
  name_prefix                 = "bastion.lifecyclephases.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.micro"
  key_name                    = "${aws_key_pair.kubernetes-lifecyclephases-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.bastions-lifecyclephases-example-com.id}"
  security_groups             = ["${aws_security_group.bastion-lifecyclephases-example-com.id}"]
  associate_public_ip_address = true

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 32
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }

  enable_monitoring = false
}

resource "aws_launch_configuration" "master-us-test-1a-masters-lifecyclephases-example-com" {
  name_prefix                 = "master-us-test-1a.masters.lifecyclephases.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-lifecyclephases-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-lifecyclephases-example-com.id}"
  security_groups             = ["${aws_security_group.masters-lifecyclephases-example-com.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.lifecyclephases.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 64
    delete_on_termination = true
  }

  ephemeral_block_device = {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }

  lifecycle = {
    create_before_destroy = true
  }

  enable_monitoring = false
}

resource "aws_launch_configuration" "nodes-lifecyclephases-example-com" {
  name_prefix                 = "nodes.lifecyclephases.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-lifecyclephases-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-lifecyclephases-example-com.id}"
  security_groups             = ["${aws_security_group.nodes-lifecyclephases-example-com.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.lifecyclephases.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 128
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }

  enable_monitoring = false
}

terraform = {
  required_version = ">= 0.9.3"
}
