output "bastion_security_group_ids" {
  value = ["${aws_security_group.bastion-privateweave-example-com.id}"]
}

output "bastions_role_arn" {
  value = "${aws_iam_role.bastions-privateweave-example-com.arn}"
}

output "bastions_role_name" {
  value = "${aws_iam_role.bastions-privateweave-example-com.name}"
}

output "cluster_name" {
  value = "privateweave.example.com"
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-privateweave-example-com.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-privateweave-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-privateweave-example-com.name}"
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-privateweave-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-privateweave-example-com.id}"]
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

output "vpc_id" {
  value = "${aws_vpc.privateweave-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "bastion-privateweave-example-com" {
  name                 = "bastion.privateweave.example.com"
  launch_configuration = "${aws_launch_configuration.bastion-privateweave-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.utility-us-test-1a-privateweave-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "privateweave.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "bastion.privateweave.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/bastion"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupMinSize", "GroupMaxSize", "GroupDesiredCapacity", "GroupInServiceInstances", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-privateweave-example-com" {
  name                 = "master-us-test-1a.masters.privateweave.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-privateweave-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-privateweave-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "privateweave.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.privateweave.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupMinSize", "GroupMaxSize", "GroupDesiredCapacity", "GroupInServiceInstances", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "nodes-privateweave-example-com" {
  name                 = "nodes.privateweave.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-privateweave-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-privateweave-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "privateweave.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.privateweave.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupMinSize", "GroupMaxSize", "GroupDesiredCapacity", "GroupInServiceInstances", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-privateweave-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "privateweave.example.com"
    Name                 = "us-test-1a.etcd-events.privateweave.example.com"
    "k8s.io/etcd/events" = "us-test-1a/us-test-1a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-privateweave-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "privateweave.example.com"
    Name                 = "us-test-1a.etcd-main.privateweave.example.com"
    "k8s.io/etcd/main"   = "us-test-1a/us-test-1a"
    "k8s.io/role/master" = "1"
  }
}
resource "aws_launch_configuration" "bastion-privateweave-example-com" {
  name_prefix                 = "bastion.privateweave.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.micro"
  key_name                    = "${aws_key_pair.kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.bastions-privateweave-example-com.id}"
  security_groups             = ["${aws_security_group.bastion-privateweave-example-com.id}"]
  associate_public_ip_address = true

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 32
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }
}

resource "aws_launch_configuration" "master-us-test-1a-masters-privateweave-example-com" {
  name_prefix                 = "master-us-test-1a.masters.privateweave.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-privateweave-example-com.id}"
  security_groups             = ["${aws_security_group.masters-privateweave-example-com.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.privateweave.example.com_user_data")}"

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
}

resource "aws_launch_configuration" "nodes-privateweave-example-com" {
  name_prefix                 = "nodes.privateweave.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-privateweave-example-com.id}"
  security_groups             = ["${aws_security_group.nodes-privateweave-example-com.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.privateweave.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 128
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }
}

terraform = {
  required_version = ">= 0.9.3"
}
