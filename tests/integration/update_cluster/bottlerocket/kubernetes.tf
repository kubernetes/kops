locals {
  cluster_name                 = "bottlerocket.example.com"
  master_autoscaling_group_ids = [aws_autoscaling_group.master-us-test-1a-masters-bottlerocket-example-com.id]
  master_security_group_ids    = [aws_security_group.masters-bottlerocket-example-com.id]
  masters_role_arn             = aws_iam_role.masters-bottlerocket-example-com.arn
  masters_role_name            = aws_iam_role.masters-bottlerocket-example-com.name
  node_autoscaling_group_ids   = [aws_autoscaling_group.nodes-bottlerocket-example-com.id]
  node_security_group_ids      = [aws_security_group.nodes-bottlerocket-example-com.id]
  node_subnet_ids              = [aws_subnet.us-test-1a-bottlerocket-example-com.id]
  nodes_role_arn               = aws_iam_role.nodes-bottlerocket-example-com.arn
  nodes_role_name              = aws_iam_role.nodes-bottlerocket-example-com.name
  region                       = "us-test-1"
  route_table_public_id        = aws_route_table.bottlerocket-example-com.id
  subnet_us-test-1a_id         = aws_subnet.us-test-1a-bottlerocket-example-com.id
  vpc_cidr_block               = aws_vpc.bottlerocket-example-com.cidr_block
  vpc_id                       = aws_vpc.bottlerocket-example-com.id
}

output "cluster_name" {
  value = "bottlerocket.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-bottlerocket-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-bottlerocket-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-bottlerocket-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-bottlerocket-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-bottlerocket-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-bottlerocket-example-com.id]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-bottlerocket-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-bottlerocket-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-bottlerocket-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "route_table_public_id" {
  value = aws_route_table.bottlerocket-example-com.id
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-bottlerocket-example-com.id
}

output "vpc_cidr_block" {
  value = aws_vpc.bottlerocket-example-com.cidr_block
}

output "vpc_id" {
  value = aws_vpc.bottlerocket-example-com.id
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-bottlerocket-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.master-us-test-1a-masters-bottlerocket-example-com.id
    version = aws_launch_template.master-us-test-1a-masters-bottlerocket-example-com.latest_version
  }
  max_size            = 1
  metrics_granularity = "1Minute"
  min_size            = 1
  name                = "master-us-test-1a.masters.bottlerocket.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "bottlerocket.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.bottlerocket.example.com"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"
    propagate_at_launch = true
    value               = "master"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master"
    propagate_at_launch = true
    value               = ""
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
    key                 = "kubernetes.io/cluster/bottlerocket.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-bottlerocket-example-com.id]
}

resource "aws_autoscaling_group" "nodes-bottlerocket-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.nodes-bottlerocket-example-com.id
    version = aws_launch_template.nodes-bottlerocket-example-com.latest_version
  }
  max_size            = 2
  metrics_granularity = "1Minute"
  min_size            = 2
  name                = "nodes.bottlerocket.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "bottlerocket.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.bottlerocket.example.com"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"
    propagate_at_launch = true
    value               = "node"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/taint/dedicated"
    propagate_at_launch = true
    value               = "gpu:NoSchedule"
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
    key                 = "kubernetes.io/cluster/bottlerocket.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-bottlerocket-example-com.id]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-bottlerocket-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "us-test-1a.etcd-events.bottlerocket.example.com"
    "k8s.io/etcd/events"                             = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-bottlerocket-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "us-test-1a.etcd-main.bottlerocket.example.com"
    "k8s.io/etcd/main"                               = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_iam_instance_profile" "masters-bottlerocket-example-com" {
  name = "masters.bottlerocket.example.com"
  role = aws_iam_role.masters-bottlerocket-example-com.name
}

resource "aws_iam_instance_profile" "nodes-bottlerocket-example-com" {
  name = "nodes.bottlerocket.example.com"
  role = aws_iam_role.nodes-bottlerocket-example-com.name
}

resource "aws_iam_role_policy" "masters-bottlerocket-example-com" {
  name   = "masters.bottlerocket.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.bottlerocket.example.com_policy")
  role   = aws_iam_role.masters-bottlerocket-example-com.name
}

resource "aws_iam_role_policy" "nodes-bottlerocket-example-com" {
  name   = "nodes.bottlerocket.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.bottlerocket.example.com_policy")
  role   = aws_iam_role.nodes-bottlerocket-example-com.name
}

resource "aws_iam_role" "masters-bottlerocket-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.bottlerocket.example.com_policy")
  name               = "masters.bottlerocket.example.com"
}

resource "aws_iam_role" "nodes-bottlerocket-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_nodes.bottlerocket.example.com_policy")
  name               = "nodes.bottlerocket.example.com"
}

resource "aws_internet_gateway" "bottlerocket-example-com" {
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "bottlerocket.example.com"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
  }
  vpc_id = aws_vpc.bottlerocket-example-com.id
}

resource "aws_key_pair" "kubernetes-bottlerocket-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.bottlerocket.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.bottlerocket.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "bottlerocket.example.com"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
  }
}

resource "aws_launch_template" "master-us-test-1a-masters-bottlerocket-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = false
      volume_size           = 64
      volume_type           = "gp2"
    }
  }
  block_device_mappings {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.masters-bottlerocket-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "m3.medium"
  key_name      = aws_key_pair.kubernetes-bottlerocket-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  name = "master-us-test-1a.masters.bottlerocket.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    security_groups             = [aws_security_group.masters-bottlerocket-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                            = "bottlerocket.example.com"
      "Name"                                                                         = "master-us-test-1a.masters.bottlerocket.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
      "k8s.io/role/master"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
      "kubernetes.io/cluster/bottlerocket.example.com"                               = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                            = "bottlerocket.example.com"
      "Name"                                                                         = "master-us-test-1a.masters.bottlerocket.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
      "k8s.io/role/master"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
      "kubernetes.io/cluster/bottlerocket.example.com"                               = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                            = "bottlerocket.example.com"
    "Name"                                                                         = "master-us-test-1a.masters.bottlerocket.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
    "k8s.io/role/master"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
    "kubernetes.io/cluster/bottlerocket.example.com"                               = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_master-us-test-1a.masters.bottlerocket.example.com_user_data")
}

resource "aws_launch_template" "nodes-bottlerocket-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = false
      volume_size           = 128
      volume_type           = "gp2"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes-bottlerocket-example-com.id
  }
  image_id      = "ami-010968555174f9163"
  instance_type = "t2.medium"
  key_name      = aws_key_pair.kubernetes-bottlerocket-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  name = "nodes.bottlerocket.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    security_groups             = [aws_security_group.nodes-bottlerocket-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "bottlerocket.example.com"
      "Name"                                                                       = "nodes.bottlerocket.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/cluster-autoscaler/node-template/taint/dedicated"                    = "gpu:NoSchedule"
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/bottlerocket.example.com"                             = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "bottlerocket.example.com"
      "Name"                                                                       = "nodes.bottlerocket.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/cluster-autoscaler/node-template/taint/dedicated"                    = "gpu:NoSchedule"
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/bottlerocket.example.com"                             = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "bottlerocket.example.com"
    "Name"                                                                       = "nodes.bottlerocket.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/cluster-autoscaler/node-template/taint/dedicated"                    = "gpu:NoSchedule"
    "k8s.io/role/node"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                  = "nodes"
    "kubernetes.io/cluster/bottlerocket.example.com"                             = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_nodes.bottlerocket.example.com_user_data")
}

resource "aws_route_table_association" "us-test-1a-bottlerocket-example-com" {
  route_table_id = aws_route_table.bottlerocket-example-com.id
  subnet_id      = aws_subnet.us-test-1a-bottlerocket-example-com.id
}

resource "aws_route_table" "bottlerocket-example-com" {
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "bottlerocket.example.com"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
    "kubernetes.io/kops/role"                        = "public"
  }
  vpc_id = aws_vpc.bottlerocket-example-com.id
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.bottlerocket-example-com.id
  route_table_id         = aws_route_table.bottlerocket-example-com.id
}

resource "aws_security_group_rule" "https-external-to-master-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-bottlerocket-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "masters-bottlerocket-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-bottlerocket-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "masters-bottlerocket-example-com-ingress-all-0to0-masters-bottlerocket-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-bottlerocket-example-com.id
  source_security_group_id = aws_security_group.masters-bottlerocket-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "masters-bottlerocket-example-com-ingress-all-0to0-nodes-bottlerocket-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-bottlerocket-example-com.id
  source_security_group_id = aws_security_group.masters-bottlerocket-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-bottlerocket-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-bottlerocket-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "nodes-bottlerocket-example-com-ingress-all-0to0-nodes-bottlerocket-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-bottlerocket-example-com.id
  source_security_group_id = aws_security_group.nodes-bottlerocket-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-bottlerocket-example-com-ingress-tcp-1to2379-masters-bottlerocket-example-com" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-bottlerocket-example-com.id
  source_security_group_id = aws_security_group.nodes-bottlerocket-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-bottlerocket-example-com-ingress-tcp-2382to4000-masters-bottlerocket-example-com" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-bottlerocket-example-com.id
  source_security_group_id = aws_security_group.nodes-bottlerocket-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-bottlerocket-example-com-ingress-tcp-4003to65535-masters-bottlerocket-example-com" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-bottlerocket-example-com.id
  source_security_group_id = aws_security_group.nodes-bottlerocket-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-bottlerocket-example-com-ingress-udp-1to65535-masters-bottlerocket-example-com" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-bottlerocket-example-com.id
  source_security_group_id = aws_security_group.nodes-bottlerocket-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-master-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-bottlerocket-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-node-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-bottlerocket-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group" "masters-bottlerocket-example-com" {
  description = "Security group for masters"
  name        = "masters.bottlerocket.example.com"
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "masters.bottlerocket.example.com"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
  }
  vpc_id = aws_vpc.bottlerocket-example-com.id
}

resource "aws_security_group" "nodes-bottlerocket-example-com" {
  description = "Security group for nodes"
  name        = "nodes.bottlerocket.example.com"
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "nodes.bottlerocket.example.com"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
  }
  vpc_id = aws_vpc.bottlerocket-example-com.id
}

resource "aws_subnet" "us-test-1a-bottlerocket-example-com" {
  availability_zone = "us-test-1a"
  cidr_block        = "172.20.32.0/19"
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "us-test-1a.bottlerocket.example.com"
    "SubnetType"                                     = "Public"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
    "kubernetes.io/role/elb"                         = "1"
  }
  vpc_id = aws_vpc.bottlerocket-example-com.id
}

resource "aws_vpc_dhcp_options_association" "bottlerocket-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.bottlerocket-example-com.id
  vpc_id          = aws_vpc.bottlerocket-example-com.id
}

resource "aws_vpc_dhcp_options" "bottlerocket-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "bottlerocket.example.com"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
  }
}

resource "aws_vpc" "bottlerocket-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    "KubernetesCluster"                              = "bottlerocket.example.com"
    "Name"                                           = "bottlerocket.example.com"
    "kubernetes.io/cluster/bottlerocket.example.com" = "owned"
  }
}

terraform {
  required_version = ">= 0.12.26"
  required_providers {
    aws = {
      "source"  = "hashicorp/aws"
      "version" = ">= 2.46.0"
    }
  }
}
