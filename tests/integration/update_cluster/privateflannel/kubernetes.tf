locals {
  bastion_autoscaling_group_ids     = [aws_autoscaling_group.bastion-privateflannel-example-com.id]
  bastion_security_group_ids        = [aws_security_group.bastion-privateflannel-example-com.id]
  bastions_role_arn                 = aws_iam_role.bastions-privateflannel-example-com.arn
  bastions_role_name                = aws_iam_role.bastions-privateflannel-example-com.name
  cluster_name                      = "privateflannel.example.com"
  master_autoscaling_group_ids      = [aws_autoscaling_group.master-us-test-1a-masters-privateflannel-example-com.id]
  master_security_group_ids         = [aws_security_group.masters-privateflannel-example-com.id]
  masters_role_arn                  = aws_iam_role.masters-privateflannel-example-com.arn
  masters_role_name                 = aws_iam_role.masters-privateflannel-example-com.name
  node_autoscaling_group_ids        = [aws_autoscaling_group.nodes-privateflannel-example-com.id]
  node_security_group_ids           = [aws_security_group.nodes-privateflannel-example-com.id]
  node_subnet_ids                   = [aws_subnet.us-test-1a-privateflannel-example-com.id]
  nodes_role_arn                    = aws_iam_role.nodes-privateflannel-example-com.arn
  nodes_role_name                   = aws_iam_role.nodes-privateflannel-example-com.name
  region                            = "us-test-1"
  route_table_private-us-test-1a_id = aws_route_table.private-us-test-1a-privateflannel-example-com.id
  route_table_public_id             = aws_route_table.privateflannel-example-com.id
  subnet_us-test-1a_id              = aws_subnet.us-test-1a-privateflannel-example-com.id
  subnet_utility-us-test-1a_id      = aws_subnet.utility-us-test-1a-privateflannel-example-com.id
  vpc_cidr_block                    = aws_vpc.privateflannel-example-com.cidr_block
  vpc_id                            = aws_vpc.privateflannel-example-com.id
}

output "bastion_autoscaling_group_ids" {
  value = [aws_autoscaling_group.bastion-privateflannel-example-com.id]
}

output "bastion_security_group_ids" {
  value = [aws_security_group.bastion-privateflannel-example-com.id]
}

output "bastions_role_arn" {
  value = aws_iam_role.bastions-privateflannel-example-com.arn
}

output "bastions_role_name" {
  value = aws_iam_role.bastions-privateflannel-example-com.name
}

output "cluster_name" {
  value = "privateflannel.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-privateflannel-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-privateflannel-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-privateflannel-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-privateflannel-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-privateflannel-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-privateflannel-example-com.id]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-privateflannel-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-privateflannel-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-privateflannel-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "route_table_private-us-test-1a_id" {
  value = aws_route_table.private-us-test-1a-privateflannel-example-com.id
}

output "route_table_public_id" {
  value = aws_route_table.privateflannel-example-com.id
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-privateflannel-example-com.id
}

output "subnet_utility-us-test-1a_id" {
  value = aws_subnet.utility-us-test-1a-privateflannel-example-com.id
}

output "vpc_cidr_block" {
  value = aws_vpc.privateflannel-example-com.cidr_block
}

output "vpc_id" {
  value = aws_vpc.privateflannel-example-com.id
}

provider "aws" {
  region = "us-test-1"
}

provider "aws" {
  alias  = "files"
  region = "us-test-1"
}

resource "aws_autoscaling_group" "bastion-privateflannel-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.bastion-privateflannel-example-com.id
    version = aws_launch_template.bastion-privateflannel-example-com.latest_version
  }
  load_balancers        = [aws_elb.bastion-privateflannel-example-com.id]
  max_instance_lifetime = 0
  max_size              = 1
  metrics_granularity   = "1Minute"
  min_size              = 1
  name                  = "bastion.privateflannel.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privateflannel.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "bastion.privateflannel.example.com"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/role/bastion"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "kops.k8s.io/instancegroup"
    propagate_at_launch = true
    value               = "bastion"
  }
  tag {
    key                 = "kubernetes.io/cluster/privateflannel.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.utility-us-test-1a-privateflannel-example-com.id]
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-privateflannel-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.master-us-test-1a-masters-privateflannel-example-com.id
    version = aws_launch_template.master-us-test-1a-masters-privateflannel-example-com.latest_version
  }
  load_balancers        = [aws_elb.api-privateflannel-example-com.id]
  max_instance_lifetime = 0
  max_size              = 1
  metrics_granularity   = "1Minute"
  min_size              = 1
  name                  = "master-us-test-1a.masters.privateflannel.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privateflannel.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.privateflannel.example.com"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers"
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
    key                 = "kubernetes.io/cluster/privateflannel.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-privateflannel-example-com.id]
}

resource "aws_autoscaling_group" "nodes-privateflannel-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.nodes-privateflannel-example-com.id
    version = aws_launch_template.nodes-privateflannel-example-com.latest_version
  }
  max_instance_lifetime = 0
  max_size              = 2
  metrics_granularity   = "1Minute"
  min_size              = 2
  name                  = "nodes.privateflannel.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privateflannel.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.privateflannel.example.com"
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"
    propagate_at_launch = true
    value               = ""
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
    key                 = "kubernetes.io/cluster/privateflannel.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-privateflannel-example-com.id]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-privateflannel-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "us-test-1a.etcd-events.privateflannel.example.com"
    "k8s.io/etcd/events"                               = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                               = "1"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-privateflannel-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "us-test-1a.etcd-main.privateflannel.example.com"
    "k8s.io/etcd/main"                                 = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                               = "1"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_eip" "us-test-1a-privateflannel-example-com" {
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "us-test-1a.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  vpc = true
}

resource "aws_elb" "api-privateflannel-example-com" {
  connection_draining         = true
  connection_draining_timeout = 300
  cross_zone_load_balancing   = false
  health_check {
    healthy_threshold   = 2
    interval            = 10
    target              = "SSL:443"
    timeout             = 5
    unhealthy_threshold = 2
  }
  idle_timeout = 300
  listener {
    instance_port     = 443
    instance_protocol = "TCP"
    lb_port           = 443
    lb_protocol       = "TCP"
  }
  name            = "api-privateflannel-exampl-hsu11v"
  security_groups = [aws_security_group.api-elb-privateflannel-example-com.id]
  subnets         = [aws_subnet.utility-us-test-1a-privateflannel-example-com.id]
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "api.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_elb" "bastion-privateflannel-example-com" {
  health_check {
    healthy_threshold   = 2
    interval            = 10
    target              = "TCP:22"
    timeout             = 5
    unhealthy_threshold = 2
  }
  idle_timeout = 300
  listener {
    instance_port     = 22
    instance_protocol = "TCP"
    lb_port           = 22
    lb_protocol       = "TCP"
  }
  name            = "bastion-privateflannel-ex-753531"
  security_groups = [aws_security_group.bastion-elb-privateflannel-example-com.id]
  subnets         = [aws_subnet.utility-us-test-1a-privateflannel-example-com.id]
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "bastion.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "bastions-privateflannel-example-com" {
  name = "bastions.privateflannel.example.com"
  role = aws_iam_role.bastions-privateflannel-example-com.name
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "bastions.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "masters-privateflannel-example-com" {
  name = "masters.privateflannel.example.com"
  role = aws_iam_role.masters-privateflannel-example-com.name
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "masters.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "nodes-privateflannel-example-com" {
  name = "nodes.privateflannel.example.com"
  role = aws_iam_role.nodes-privateflannel-example-com.name
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "nodes.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_iam_role" "bastions-privateflannel-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_bastions.privateflannel.example.com_policy")
  name               = "bastions.privateflannel.example.com"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "bastions.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_iam_role" "masters-privateflannel-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.privateflannel.example.com_policy")
  name               = "masters.privateflannel.example.com"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "masters.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_iam_role" "nodes-privateflannel-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_nodes.privateflannel.example.com_policy")
  name               = "nodes.privateflannel.example.com"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "nodes.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_iam_role_policy" "bastions-privateflannel-example-com" {
  name   = "bastions.privateflannel.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_bastions.privateflannel.example.com_policy")
  role   = aws_iam_role.bastions-privateflannel-example-com.name
}

resource "aws_iam_role_policy" "masters-privateflannel-example-com" {
  name   = "masters.privateflannel.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.privateflannel.example.com_policy")
  role   = aws_iam_role.masters-privateflannel-example-com.name
}

resource "aws_iam_role_policy" "nodes-privateflannel-example-com" {
  name   = "nodes.privateflannel.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.privateflannel.example.com_policy")
  role   = aws_iam_role.nodes-privateflannel-example-com.name
}

resource "aws_internet_gateway" "privateflannel-example-com" {
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_key_pair" "kubernetes-privateflannel-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.privateflannel.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.privateflannel.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_launch_template" "bastion-privateflannel-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 32
      volume_type           = "gp3"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.bastions-privateflannel-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.micro"
  key_name      = aws_key_pair.kubernetes-privateflannel-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "bastion.privateflannel.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.bastion-privateflannel-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "privateflannel.example.com"
      "Name"                                                                       = "bastion.privateflannel.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/bastion"                                                        = "1"
      "kops.k8s.io/instancegroup"                                                  = "bastion"
      "kubernetes.io/cluster/privateflannel.example.com"                           = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "privateflannel.example.com"
      "Name"                                                                       = "bastion.privateflannel.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/bastion"                                                        = "1"
      "kops.k8s.io/instancegroup"                                                  = "bastion"
      "kubernetes.io/cluster/privateflannel.example.com"                           = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "privateflannel.example.com"
    "Name"                                                                       = "bastion.privateflannel.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/bastion"                                                        = "1"
    "kops.k8s.io/instancegroup"                                                  = "bastion"
    "kubernetes.io/cluster/privateflannel.example.com"                           = "owned"
  }
}

resource "aws_launch_template" "master-us-test-1a-masters-privateflannel-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 64
      volume_type           = "gp3"
    }
  }
  block_device_mappings {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.masters-privateflannel-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "m3.medium"
  key_name      = aws_key_pair.kubernetes-privateflannel-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "master-us-test-1a.masters.privateflannel.example.com"
  network_interfaces {
    associate_public_ip_address = false
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.masters-privateflannel-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                                                     = "privateflannel.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.privateflannel.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/privateflannel.example.com"                                                      = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                                                     = "privateflannel.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.privateflannel.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/privateflannel.example.com"                                                      = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                                                     = "privateflannel.example.com"
    "Name"                                                                                                  = "master-us-test-1a.masters.privateflannel.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
    "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
    "k8s.io/role/master"                                                                                    = "1"
    "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
    "kubernetes.io/cluster/privateflannel.example.com"                                                      = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_master-us-test-1a.masters.privateflannel.example.com_user_data")
}

resource "aws_launch_template" "nodes-privateflannel-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 128
      volume_type           = "gp3"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes-privateflannel-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.medium"
  key_name      = aws_key_pair.kubernetes-privateflannel-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "nodes.privateflannel.example.com"
  network_interfaces {
    associate_public_ip_address = false
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.nodes-privateflannel-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "privateflannel.example.com"
      "Name"                                                                       = "nodes.privateflannel.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/privateflannel.example.com"                           = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "privateflannel.example.com"
      "Name"                                                                       = "nodes.privateflannel.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/privateflannel.example.com"                           = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "privateflannel.example.com"
    "Name"                                                                       = "nodes.privateflannel.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/node"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                  = "nodes"
    "kubernetes.io/cluster/privateflannel.example.com"                           = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_nodes.privateflannel.example.com_user_data")
}

resource "aws_nat_gateway" "us-test-1a-privateflannel-example-com" {
  allocation_id = aws_eip.us-test-1a-privateflannel-example-com.id
  subnet_id     = aws_subnet.utility-us-test-1a-privateflannel-example-com.id
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "us-test-1a.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.privateflannel-example-com.id
  route_table_id         = aws_route_table.privateflannel-example-com.id
}

resource "aws_route" "route-__--0" {
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.privateflannel-example-com.id
  route_table_id              = aws_route_table.privateflannel-example-com.id
}

resource "aws_route" "route-private-us-test-1a-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.us-test-1a-privateflannel-example-com.id
  route_table_id         = aws_route_table.private-us-test-1a-privateflannel-example-com.id
}

resource "aws_route53_record" "api-privateflannel-example-com" {
  alias {
    evaluate_target_health = false
    name                   = aws_elb.api-privateflannel-example-com.dns_name
    zone_id                = aws_elb.api-privateflannel-example-com.zone_id
  }
  name    = "api.privateflannel.example.com"
  type    = "A"
  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route_table" "private-us-test-1a-privateflannel-example-com" {
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "private-us-test-1a.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
    "kubernetes.io/kops/role"                          = "private-us-test-1a"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_route_table" "privateflannel-example-com" {
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
    "kubernetes.io/kops/role"                          = "public"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_route_table_association" "private-us-test-1a-privateflannel-example-com" {
  route_table_id = aws_route_table.private-us-test-1a-privateflannel-example-com.id
  subnet_id      = aws_subnet.us-test-1a-privateflannel-example-com.id
}

resource "aws_route_table_association" "utility-us-test-1a-privateflannel-example-com" {
  route_table_id = aws_route_table.privateflannel-example-com.id
  subnet_id      = aws_subnet.utility-us-test-1a-privateflannel-example-com.id
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "clusters.example.com/privateflannel.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "clusters.example.com/privateflannel.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "clusters.example.com/privateflannel.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "clusters.example.com/privateflannel.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-master-us-test-1a_content")
  key                    = "clusters.example.com/privateflannel.example.com/manifests/etcd/events-master-us-test-1a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-master-us-test-1a_content")
  key                    = "clusters.example.com/privateflannel.example.com/manifests/etcd/main-master-us-test-1a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "clusters.example.com/privateflannel.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test-1a_content")
  key                    = "clusters.example.com/privateflannel.example.com/igconfig/master/master-us-test-1a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "clusters.example.com/privateflannel.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-aws-cloud-controller-addons-k8s-io-k8s-1-18" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-aws-cloud-controller.addons.k8s.io-k8s-1.18_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/aws-cloud-controller.addons.k8s.io/k8s-1.18.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-aws-ebs-csi-driver-addons-k8s-io-k8s-1-17" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-aws-ebs-csi-driver.addons.k8s.io-k8s-1.17_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/aws-ebs-csi-driver.addons.k8s.io/k8s-1.17.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-bootstrap_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-leader-migration-rbac-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-leader-migration.rbac.addons.k8s.io-k8s-1.23_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/leader-migration.rbac.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-networking-flannel-k8s-1-25" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-networking.flannel-k8s-1.25_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/networking.flannel/k8s-1.25.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privateflannel-example-com-addons-storage-aws-addons-k8s-io-v1-15-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privateflannel.example.com-addons-storage-aws.addons.k8s.io-v1.15.0_content")
  key                    = "clusters.example.com/privateflannel.example.com/addons/storage-aws.addons.k8s.io/v1.15.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_security_group" "api-elb-privateflannel-example-com" {
  description = "Security group for api ELB"
  name        = "api-elb.privateflannel.example.com"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "api-elb.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_security_group" "bastion-elb-privateflannel-example-com" {
  description = "Security group for bastion ELB"
  name        = "bastion-elb.privateflannel.example.com"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "bastion-elb.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_security_group" "bastion-privateflannel-example-com" {
  description = "Security group for bastion"
  name        = "bastion.privateflannel.example.com"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "bastion.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_security_group" "masters-privateflannel-example-com" {
  description = "Security group for masters"
  name        = "masters.privateflannel.example.com"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "masters.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_security_group" "nodes-privateflannel-example-com" {
  description = "Security group for nodes"
  name        = "nodes.privateflannel.example.com"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "nodes.privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-22to22-bastion-elb-privateflannel-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.bastion-elb-privateflannel-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-443to443-api-elb-privateflannel-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-privateflannel-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "from-api-elb-privateflannel-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-api-elb-privateflannel-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privateflannel-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-elb-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privateflannel-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-elb-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privateflannel-example-com-ingress-tcp-22to22-bastion-privateflannel-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.bastion-privateflannel-example-com.id
  source_security_group_id = aws_security_group.bastion-elb-privateflannel-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-bastion-privateflannel-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-privateflannel-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-privateflannel-example-com-ingress-tcp-22to22-masters-privateflannel-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateflannel-example-com.id
  source_security_group_id = aws_security_group.bastion-privateflannel-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-bastion-privateflannel-example-com-ingress-tcp-22to22-nodes-privateflannel-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.nodes-privateflannel-example-com.id
  source_security_group_id = aws_security_group.bastion-privateflannel-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-masters-privateflannel-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-privateflannel-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.masters-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-privateflannel-example-com-ingress-all-0to0-masters-privateflannel-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-privateflannel-example-com.id
  source_security_group_id = aws_security_group.masters-privateflannel-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-masters-privateflannel-example-com-ingress-all-0to0-nodes-privateflannel-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-privateflannel-example-com.id
  source_security_group_id = aws_security_group.masters-privateflannel-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateflannel-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-privateflannel-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-privateflannel-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-privateflannel-example-com-ingress-all-0to0-nodes-privateflannel-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-privateflannel-example-com.id
  source_security_group_id = aws_security_group.nodes-privateflannel-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateflannel-example-com-ingress-tcp-1to2379-masters-privateflannel-example-com" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateflannel-example-com.id
  source_security_group_id = aws_security_group.nodes-privateflannel-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateflannel-example-com-ingress-tcp-2382to4000-masters-privateflannel-example-com" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateflannel-example-com.id
  source_security_group_id = aws_security_group.nodes-privateflannel-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateflannel-example-com-ingress-tcp-4003to65535-masters-privateflannel-example-com" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateflannel-example-com.id
  source_security_group_id = aws_security_group.nodes-privateflannel-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privateflannel-example-com-ingress-udp-1to65535-masters-privateflannel-example-com" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-privateflannel-example-com.id
  source_security_group_id = aws_security_group.nodes-privateflannel-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "https-elb-to-master" {
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privateflannel-example-com.id
  source_security_group_id = aws_security_group.api-elb-privateflannel-example-com.id
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.api-elb-privateflannel-example-com.id
  to_port           = 4
  type              = "ingress"
}

resource "aws_subnet" "us-test-1a-privateflannel-example-com" {
  availability_zone                           = "us-test-1a"
  cidr_block                                  = "172.20.32.0/19"
  enable_resource_name_dns_a_record_on_launch = true
  private_dns_hostname_type_on_launch         = "resource-name"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "us-test-1a.privateflannel.example.com"
    "SubnetType"                                       = "Private"
    "kops.k8s.io/instance-group/master-us-test-1a"     = "true"
    "kops.k8s.io/instance-group/nodes"                 = "true"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
    "kubernetes.io/role/internal-elb"                  = "1"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_subnet" "utility-us-test-1a-privateflannel-example-com" {
  availability_zone                           = "us-test-1a"
  cidr_block                                  = "172.20.4.0/22"
  enable_resource_name_dns_a_record_on_launch = true
  private_dns_hostname_type_on_launch         = "resource-name"
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "utility-us-test-1a.privateflannel.example.com"
    "SubnetType"                                       = "Utility"
    "kops.k8s.io/instance-group/bastion"               = "true"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
    "kubernetes.io/role/elb"                           = "1"
  }
  vpc_id = aws_vpc.privateflannel-example-com.id
}

resource "aws_vpc" "privateflannel-example-com" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "172.20.0.0/16"
  enable_dns_hostnames             = true
  enable_dns_support               = true
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "privateflannel-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags = {
    "KubernetesCluster"                                = "privateflannel.example.com"
    "Name"                                             = "privateflannel.example.com"
    "kubernetes.io/cluster/privateflannel.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options_association" "privateflannel-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.privateflannel-example-com.id
  vpc_id          = aws_vpc.privateflannel-example-com.id
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 4.0.0"
    }
  }
}
