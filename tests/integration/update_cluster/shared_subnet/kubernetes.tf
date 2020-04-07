locals {
  cluster_name                 = "sharedsubnet.example.com"
  master_autoscaling_group_ids = [aws_autoscaling_group.master-us-test-1a-masters-sharedsubnet-example-com.id]
  master_security_group_ids    = [aws_security_group.masters-sharedsubnet-example-com.id]
  masters_role_arn             = aws_iam_role.masters-sharedsubnet-example-com.arn
  masters_role_name            = aws_iam_role.masters-sharedsubnet-example-com.name
  node_autoscaling_group_ids   = [aws_autoscaling_group.nodes-sharedsubnet-example-com.id]
  node_security_group_ids      = [aws_security_group.nodes-sharedsubnet-example-com.id]
  node_subnet_ids              = ["subnet-12345678"]
  nodes_role_arn               = aws_iam_role.nodes-sharedsubnet-example-com.arn
  nodes_role_name              = aws_iam_role.nodes-sharedsubnet-example-com.name
  region                       = "us-test-1"
  subnet_ids                   = ["subnet-12345678"]
  subnet_us-test-1a_id         = "subnet-12345678"
  vpc_id                       = "vpc-12345678"
}

output "cluster_name" {
  value = "sharedsubnet.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-sharedsubnet-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-sharedsubnet-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-sharedsubnet-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-sharedsubnet-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-sharedsubnet-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-sharedsubnet-example-com.id]
}

output "node_subnet_ids" {
  value = ["subnet-12345678"]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-sharedsubnet-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-sharedsubnet-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "subnet_ids" {
  value = ["subnet-12345678"]
}

output "subnet_us-test-1a_id" {
  value = "subnet-12345678"
}

output "vpc_id" {
  value = "vpc-12345678"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-sharedsubnet-example-com" {
  enabled_metrics      = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_configuration = aws_launch_configuration.master-us-test-1a-masters-sharedsubnet-example-com.id
  max_size             = 1
  metrics_granularity  = "1Minute"
  min_size             = 1
  name                 = "master-us-test-1a.masters.sharedsubnet.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "sharedsubnet.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.sharedsubnet.example.com"
  }
  tag {
    key                 = "k8s.io/role/master"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "kops.k8s.io/instancegroup"
    propagate_at_launch = true
    value               = "master-us-test-1a"
  }
  tag {
    key                 = "kubernetes.io/cluster/sharedsubnet.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = ["subnet-12345678"]
}

resource "aws_autoscaling_group" "nodes-sharedsubnet-example-com" {
  enabled_metrics      = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_configuration = aws_launch_configuration.nodes-sharedsubnet-example-com.id
  max_size             = 2
  metrics_granularity  = "1Minute"
  min_size             = 2
  name                 = "nodes.sharedsubnet.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "sharedsubnet.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.sharedsubnet.example.com"
  }
  tag {
    key                 = "k8s.io/role/node"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "kops.k8s.io/instancegroup"
    propagate_at_launch = true
    value               = "nodes"
  }
  tag {
    key                 = "kubernetes.io/cluster/sharedsubnet.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = ["subnet-12345678"]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-sharedsubnet-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                              = "sharedsubnet.example.com"
    "Name"                                           = "us-test-1a.etcd-events.sharedsubnet.example.com"
    "k8s.io/etcd/events"                             = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/sharedsubnet.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-sharedsubnet-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                              = "sharedsubnet.example.com"
    "Name"                                           = "us-test-1a.etcd-main.sharedsubnet.example.com"
    "k8s.io/etcd/main"                               = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/sharedsubnet.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_iam_instance_profile" "masters-sharedsubnet-example-com" {
  name = "masters.sharedsubnet.example.com"
  role = aws_iam_role.masters-sharedsubnet-example-com.name
}

resource "aws_iam_instance_profile" "nodes-sharedsubnet-example-com" {
  name = "nodes.sharedsubnet.example.com"
  role = aws_iam_role.nodes-sharedsubnet-example-com.name
}

resource "aws_iam_role_policy" "masters-sharedsubnet-example-com" {
  name   = "masters.sharedsubnet.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.sharedsubnet.example.com_policy")
  role   = aws_iam_role.masters-sharedsubnet-example-com.name
}

resource "aws_iam_role_policy" "nodes-sharedsubnet-example-com" {
  name   = "nodes.sharedsubnet.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.sharedsubnet.example.com_policy")
  role   = aws_iam_role.nodes-sharedsubnet-example-com.name
}

resource "aws_iam_role" "masters-sharedsubnet-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.sharedsubnet.example.com_policy")
  name               = "masters.sharedsubnet.example.com"
}

resource "aws_iam_role" "nodes-sharedsubnet-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_nodes.sharedsubnet.example.com_policy")
  name               = "nodes.sharedsubnet.example.com"
}

resource "aws_key_pair" "kubernetes-sharedsubnet-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.sharedsubnet.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.sharedsubnet.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
}

resource "aws_launch_configuration" "master-us-test-1a-masters-sharedsubnet-example-com" {
  associate_public_ip_address = true
  enable_monitoring           = false
  ephemeral_block_device {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile = aws_iam_instance_profile.masters-sharedsubnet-example-com.id
  image_id             = "ami-12345678"
  instance_type        = "m3.medium"
  key_name             = aws_key_pair.kubernetes-sharedsubnet-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "master-us-test-1a.masters.sharedsubnet.example.com-"
  root_block_device {
    delete_on_termination = true
    volume_size           = 64
    volume_type           = "gp2"
  }
  security_groups = [aws_security_group.masters-sharedsubnet-example-com.id]
  user_data       = file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.sharedsubnet.example.com_user_data")
}

resource "aws_launch_configuration" "nodes-sharedsubnet-example-com" {
  associate_public_ip_address = true
  enable_monitoring           = false
  iam_instance_profile        = aws_iam_instance_profile.nodes-sharedsubnet-example-com.id
  image_id                    = "ami-12345678"
  instance_type               = "t2.medium"
  key_name                    = aws_key_pair.kubernetes-sharedsubnet-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "nodes.sharedsubnet.example.com-"
  root_block_device {
    delete_on_termination = true
    volume_size           = 128
    volume_type           = "gp2"
  }
  security_groups = [aws_security_group.nodes-sharedsubnet-example-com.id]
  user_data       = file("${path.module}/data/aws_launch_configuration_nodes.sharedsubnet.example.com_user_data")
}

resource "aws_security_group_rule" "all-master-to-master" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-sharedsubnet-example-com.id
  source_security_group_id = aws_security_group.masters-sharedsubnet-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "all-master-to-node" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-sharedsubnet-example-com.id
  source_security_group_id = aws_security_group.masters-sharedsubnet-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "all-node-to-node" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-sharedsubnet-example-com.id
  source_security_group_id = aws_security_group.nodes-sharedsubnet-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "https-external-to-master-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-sharedsubnet-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "master-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-sharedsubnet-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "node-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-sharedsubnet-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-sharedsubnet-example-com.id
  source_security_group_id = aws_security_group.nodes-sharedsubnet-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-tcp-2382-4000" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-sharedsubnet-example-com.id
  source_security_group_id = aws_security_group.nodes-sharedsubnet-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-sharedsubnet-example-com.id
  source_security_group_id = aws_security_group.nodes-sharedsubnet-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-sharedsubnet-example-com.id
  source_security_group_id = aws_security_group.nodes-sharedsubnet-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-master-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-sharedsubnet-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-node-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-sharedsubnet-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group" "masters-sharedsubnet-example-com" {
  description = "Security group for masters"
  name        = "masters.sharedsubnet.example.com"
  tags = {
    "KubernetesCluster"                              = "sharedsubnet.example.com"
    "Name"                                           = "masters.sharedsubnet.example.com"
    "kubernetes.io/cluster/sharedsubnet.example.com" = "owned"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_security_group" "nodes-sharedsubnet-example-com" {
  description = "Security group for nodes"
  name        = "nodes.sharedsubnet.example.com"
  tags = {
    "KubernetesCluster"                              = "sharedsubnet.example.com"
    "Name"                                           = "nodes.sharedsubnet.example.com"
    "kubernetes.io/cluster/sharedsubnet.example.com" = "owned"
  }
  vpc_id = "vpc-12345678"
}

terraform {
  required_version = ">= 0.12.0"
}
