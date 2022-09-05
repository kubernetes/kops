locals {
  bastion_autoscaling_group_ids     = [aws_autoscaling_group.bastion-privatecanal-example-com.id]
  bastion_security_group_ids        = [aws_security_group.bastion-privatecanal-example-com.id]
  bastions_role_arn                 = aws_iam_role.bastions-privatecanal-example-com.arn
  bastions_role_name                = aws_iam_role.bastions-privatecanal-example-com.name
  cluster_name                      = "privatecanal.example.com"
  master_autoscaling_group_ids      = [aws_autoscaling_group.master-us-test-1a-masters-privatecanal-example-com.id]
  master_security_group_ids         = [aws_security_group.masters-privatecanal-example-com.id]
  masters_role_arn                  = aws_iam_role.masters-privatecanal-example-com.arn
  masters_role_name                 = aws_iam_role.masters-privatecanal-example-com.name
  node_autoscaling_group_ids        = [aws_autoscaling_group.nodes-privatecanal-example-com.id]
  node_security_group_ids           = [aws_security_group.nodes-privatecanal-example-com.id]
  node_subnet_ids                   = [aws_subnet.us-test-1a-privatecanal-example-com.id]
  nodes_role_arn                    = aws_iam_role.nodes-privatecanal-example-com.arn
  nodes_role_name                   = aws_iam_role.nodes-privatecanal-example-com.name
  region                            = "us-test-1"
  route_table_private-us-test-1a_id = aws_route_table.private-us-test-1a-privatecanal-example-com.id
  route_table_public_id             = aws_route_table.privatecanal-example-com.id
  subnet_us-test-1a_id              = aws_subnet.us-test-1a-privatecanal-example-com.id
  subnet_utility-us-test-1a_id      = aws_subnet.utility-us-test-1a-privatecanal-example-com.id
  vpc_cidr_block                    = aws_vpc.privatecanal-example-com.cidr_block
  vpc_id                            = aws_vpc.privatecanal-example-com.id
}

output "bastion_autoscaling_group_ids" {
  value = [aws_autoscaling_group.bastion-privatecanal-example-com.id]
}

output "bastion_security_group_ids" {
  value = [aws_security_group.bastion-privatecanal-example-com.id]
}

output "bastions_role_arn" {
  value = aws_iam_role.bastions-privatecanal-example-com.arn
}

output "bastions_role_name" {
  value = aws_iam_role.bastions-privatecanal-example-com.name
}

output "cluster_name" {
  value = "privatecanal.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-privatecanal-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-privatecanal-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-privatecanal-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-privatecanal-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-privatecanal-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-privatecanal-example-com.id]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-privatecanal-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-privatecanal-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-privatecanal-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "route_table_private-us-test-1a_id" {
  value = aws_route_table.private-us-test-1a-privatecanal-example-com.id
}

output "route_table_public_id" {
  value = aws_route_table.privatecanal-example-com.id
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-privatecanal-example-com.id
}

output "subnet_utility-us-test-1a_id" {
  value = aws_subnet.utility-us-test-1a-privatecanal-example-com.id
}

output "vpc_cidr_block" {
  value = aws_vpc.privatecanal-example-com.cidr_block
}

output "vpc_id" {
  value = aws_vpc.privatecanal-example-com.id
}

provider "aws" {
  region = "us-test-1"
}

provider "aws" {
  alias  = "files"
  region = "us-test-1"
}

resource "aws_autoscaling_group" "bastion-privatecanal-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.bastion-privatecanal-example-com.id
    version = aws_launch_template.bastion-privatecanal-example-com.latest_version
  }
  load_balancers        = [aws_elb.bastion-privatecanal-example-com.id]
  max_instance_lifetime = 0
  max_size              = 1
  metrics_granularity   = "1Minute"
  min_size              = 1
  name                  = "bastion.privatecanal.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privatecanal.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "bastion.privatecanal.example.com"
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
    key                 = "kubernetes.io/cluster/privatecanal.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.utility-us-test-1a-privatecanal-example-com.id]
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-privatecanal-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.master-us-test-1a-masters-privatecanal-example-com.id
    version = aws_launch_template.master-us-test-1a-masters-privatecanal-example-com.latest_version
  }
  load_balancers        = [aws_elb.api-privatecanal-example-com.id]
  max_instance_lifetime = 0
  max_size              = 1
  metrics_granularity   = "1Minute"
  min_size              = 1
  name                  = "master-us-test-1a.masters.privatecanal.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privatecanal.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.privatecanal.example.com"
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
    key                 = "kubernetes.io/cluster/privatecanal.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-privatecanal-example-com.id]
}

resource "aws_autoscaling_group" "nodes-privatecanal-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.nodes-privatecanal-example-com.id
    version = aws_launch_template.nodes-privatecanal-example-com.latest_version
  }
  max_instance_lifetime = 0
  max_size              = 2
  metrics_granularity   = "1Minute"
  min_size              = 2
  name                  = "nodes.privatecanal.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "privatecanal.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.privatecanal.example.com"
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
    key                 = "kubernetes.io/cluster/privatecanal.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-privatecanal-example-com.id]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-privatecanal-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "us-test-1a.etcd-events.privatecanal.example.com"
    "k8s.io/etcd/events"                             = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-privatecanal-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "us-test-1a.etcd-main.privatecanal.example.com"
    "k8s.io/etcd/main"                               = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_eip" "us-test-1a-privatecanal-example-com" {
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "us-test-1a.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  vpc = true
}

resource "aws_elb" "api-privatecanal-example-com" {
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
  name            = "api-privatecanal-example--6tql53"
  security_groups = [aws_security_group.api-elb-privatecanal-example-com.id]
  subnets         = [aws_subnet.utility-us-test-1a-privatecanal-example-com.id]
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "api.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_elb" "bastion-privatecanal-example-com" {
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
  name            = "bastion-privatecanal-exam-hmhsp5"
  security_groups = [aws_security_group.bastion-elb-privatecanal-example-com.id]
  subnets         = [aws_subnet.utility-us-test-1a-privatecanal-example-com.id]
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "bastion.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "bastions-privatecanal-example-com" {
  name = "bastions.privatecanal.example.com"
  role = aws_iam_role.bastions-privatecanal-example-com.name
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "bastions.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "masters-privatecanal-example-com" {
  name = "masters.privatecanal.example.com"
  role = aws_iam_role.masters-privatecanal-example-com.name
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "masters.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "nodes-privatecanal-example-com" {
  name = "nodes.privatecanal.example.com"
  role = aws_iam_role.nodes-privatecanal-example-com.name
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "nodes.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_iam_role" "bastions-privatecanal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_bastions.privatecanal.example.com_policy")
  name               = "bastions.privatecanal.example.com"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "bastions.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_iam_role" "masters-privatecanal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.privatecanal.example.com_policy")
  name               = "masters.privatecanal.example.com"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "masters.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_iam_role" "nodes-privatecanal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_nodes.privatecanal.example.com_policy")
  name               = "nodes.privatecanal.example.com"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "nodes.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_iam_role_policy" "bastions-privatecanal-example-com" {
  name   = "bastions.privatecanal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_bastions.privatecanal.example.com_policy")
  role   = aws_iam_role.bastions-privatecanal-example-com.name
}

resource "aws_iam_role_policy" "masters-privatecanal-example-com" {
  name   = "masters.privatecanal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.privatecanal.example.com_policy")
  role   = aws_iam_role.masters-privatecanal-example-com.name
}

resource "aws_iam_role_policy" "nodes-privatecanal-example-com" {
  name   = "nodes.privatecanal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.privatecanal.example.com_policy")
  role   = aws_iam_role.nodes-privatecanal-example-com.name
}

resource "aws_internet_gateway" "privatecanal-example-com" {
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_key_pair" "kubernetes-privatecanal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.privatecanal.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.privatecanal.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_launch_template" "bastion-privatecanal-example-com" {
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
    name = aws_iam_instance_profile.bastions-privatecanal-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.micro"
  key_name      = aws_key_pair.kubernetes-privatecanal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
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
  name = "bastion.privatecanal.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.bastion-privatecanal-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "privatecanal.example.com"
      "Name"                                                                       = "bastion.privatecanal.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/bastion"                                                        = "1"
      "kops.k8s.io/instancegroup"                                                  = "bastion"
      "kubernetes.io/cluster/privatecanal.example.com"                             = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "privatecanal.example.com"
      "Name"                                                                       = "bastion.privatecanal.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/bastion"                                                        = "1"
      "kops.k8s.io/instancegroup"                                                  = "bastion"
      "kubernetes.io/cluster/privatecanal.example.com"                             = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "privatecanal.example.com"
    "Name"                                                                       = "bastion.privatecanal.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/bastion"                                                        = "1"
    "kops.k8s.io/instancegroup"                                                  = "bastion"
    "kubernetes.io/cluster/privatecanal.example.com"                             = "owned"
  }
}

resource "aws_launch_template" "master-us-test-1a-masters-privatecanal-example-com" {
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
    name = aws_iam_instance_profile.masters-privatecanal-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "m3.medium"
  key_name      = aws_key_pair.kubernetes-privatecanal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
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
  name = "master-us-test-1a.masters.privatecanal.example.com"
  network_interfaces {
    associate_public_ip_address = false
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.masters-privatecanal-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                                                     = "privatecanal.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.privatecanal.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/privatecanal.example.com"                                                        = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                                                     = "privatecanal.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.privatecanal.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/privatecanal.example.com"                                                        = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                                                     = "privatecanal.example.com"
    "Name"                                                                                                  = "master-us-test-1a.masters.privatecanal.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
    "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
    "k8s.io/role/master"                                                                                    = "1"
    "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
    "kubernetes.io/cluster/privatecanal.example.com"                                                        = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_master-us-test-1a.masters.privatecanal.example.com_user_data")
}

resource "aws_launch_template" "nodes-privatecanal-example-com" {
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
    name = aws_iam_instance_profile.nodes-privatecanal-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.medium"
  key_name      = aws_key_pair.kubernetes-privatecanal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
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
  name = "nodes.privatecanal.example.com"
  network_interfaces {
    associate_public_ip_address = false
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.nodes-privatecanal-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "privatecanal.example.com"
      "Name"                                                                       = "nodes.privatecanal.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/privatecanal.example.com"                             = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "privatecanal.example.com"
      "Name"                                                                       = "nodes.privatecanal.example.com"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/privatecanal.example.com"                             = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "privatecanal.example.com"
    "Name"                                                                       = "nodes.privatecanal.example.com"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/node"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                  = "nodes"
    "kubernetes.io/cluster/privatecanal.example.com"                             = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_nodes.privatecanal.example.com_user_data")
}

resource "aws_nat_gateway" "us-test-1a-privatecanal-example-com" {
  allocation_id = aws_eip.us-test-1a-privatecanal-example-com.id
  subnet_id     = aws_subnet.utility-us-test-1a-privatecanal-example-com.id
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "us-test-1a.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.privatecanal-example-com.id
  route_table_id         = aws_route_table.privatecanal-example-com.id
}

resource "aws_route" "route-__--0" {
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.privatecanal-example-com.id
  route_table_id              = aws_route_table.privatecanal-example-com.id
}

resource "aws_route" "route-private-us-test-1a-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.us-test-1a-privatecanal-example-com.id
  route_table_id         = aws_route_table.private-us-test-1a-privatecanal-example-com.id
}

resource "aws_route53_record" "api-privatecanal-example-com" {
  alias {
    evaluate_target_health = false
    name                   = aws_elb.api-privatecanal-example-com.dns_name
    zone_id                = aws_elb.api-privatecanal-example-com.zone_id
  }
  name    = "api.privatecanal.example.com"
  type    = "A"
  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route_table" "private-us-test-1a-privatecanal-example-com" {
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "private-us-test-1a.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
    "kubernetes.io/kops/role"                        = "private-us-test-1a"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_route_table" "privatecanal-example-com" {
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
    "kubernetes.io/kops/role"                        = "public"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_route_table_association" "private-us-test-1a-privatecanal-example-com" {
  route_table_id = aws_route_table.private-us-test-1a-privatecanal-example-com.id
  subnet_id      = aws_subnet.us-test-1a-privatecanal-example-com.id
}

resource "aws_route_table_association" "utility-us-test-1a-privatecanal-example-com" {
  route_table_id = aws_route_table.privatecanal-example-com.id
  subnet_id      = aws_subnet.utility-us-test-1a-privatecanal-example-com.id
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "clusters.example.com/privatecanal.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "clusters.example.com/privatecanal.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "clusters.example.com/privatecanal.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "clusters.example.com/privatecanal.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-master-us-test-1a_content")
  key                    = "clusters.example.com/privatecanal.example.com/manifests/etcd/events-master-us-test-1a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-master-us-test-1a_content")
  key                    = "clusters.example.com/privatecanal.example.com/manifests/etcd/main-master-us-test-1a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "clusters.example.com/privatecanal.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test-1a_content")
  key                    = "clusters.example.com/privatecanal.example.com/igconfig/master/master-us-test-1a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "clusters.example.com/privatecanal.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-aws-cloud-controller-addons-k8s-io-k8s-1-18" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-aws-cloud-controller.addons.k8s.io-k8s-1.18_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/aws-cloud-controller.addons.k8s.io/k8s-1.18.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-aws-ebs-csi-driver-addons-k8s-io-k8s-1-17" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-aws-ebs-csi-driver.addons.k8s.io-k8s-1.17_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/aws-ebs-csi-driver.addons.k8s.io/k8s-1.17.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-bootstrap_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-leader-migration-rbac-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-leader-migration.rbac.addons.k8s.io-k8s-1.23_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/leader-migration.rbac.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-networking-projectcalico-org-canal-k8s-1-25" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-networking.projectcalico.org.canal-k8s-1.25_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/networking.projectcalico.org.canal/k8s-1.25.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "privatecanal-example-com-addons-storage-aws-addons-k8s-io-v1-15-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_privatecanal.example.com-addons-storage-aws.addons.k8s.io-v1.15.0_content")
  key                    = "clusters.example.com/privatecanal.example.com/addons/storage-aws.addons.k8s.io/v1.15.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_security_group" "api-elb-privatecanal-example-com" {
  description = "Security group for api ELB"
  name        = "api-elb.privatecanal.example.com"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "api-elb.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_security_group" "bastion-elb-privatecanal-example-com" {
  description = "Security group for bastion ELB"
  name        = "bastion-elb.privatecanal.example.com"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "bastion-elb.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_security_group" "bastion-privatecanal-example-com" {
  description = "Security group for bastion"
  name        = "bastion.privatecanal.example.com"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "bastion.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_security_group" "masters-privatecanal-example-com" {
  description = "Security group for masters"
  name        = "masters.privatecanal.example.com"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "masters.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_security_group" "nodes-privatecanal-example-com" {
  description = "Security group for nodes"
  name        = "nodes.privatecanal.example.com"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "nodes.privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-22to22-bastion-elb-privatecanal-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.bastion-elb-privatecanal-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-443to443-api-elb-privatecanal-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-privatecanal-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "from-api-elb-privatecanal-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-api-elb-privatecanal-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privatecanal-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-elb-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privatecanal-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-elb-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-elb-privatecanal-example-com-ingress-tcp-22to22-bastion-privatecanal-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.bastion-privatecanal-example-com.id
  source_security_group_id = aws_security_group.bastion-elb-privatecanal-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-bastion-privatecanal-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-privatecanal-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-bastion-privatecanal-example-com-ingress-tcp-22to22-masters-privatecanal-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privatecanal-example-com.id
  source_security_group_id = aws_security_group.bastion-privatecanal-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-bastion-privatecanal-example-com-ingress-tcp-22to22-nodes-privatecanal-example-com" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.nodes-privatecanal-example-com.id
  source_security_group_id = aws_security_group.bastion-privatecanal-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-masters-privatecanal-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-privatecanal-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.masters-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-privatecanal-example-com-ingress-all-0to0-masters-privatecanal-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-privatecanal-example-com.id
  source_security_group_id = aws_security_group.masters-privatecanal-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-masters-privatecanal-example-com-ingress-all-0to0-nodes-privatecanal-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-privatecanal-example-com.id
  source_security_group_id = aws_security_group.masters-privatecanal-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privatecanal-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-privatecanal-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-privatecanal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-privatecanal-example-com-ingress-all-0to0-nodes-privatecanal-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-privatecanal-example-com.id
  source_security_group_id = aws_security_group.nodes-privatecanal-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privatecanal-example-com-ingress-tcp-1to2379-masters-privatecanal-example-com" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privatecanal-example-com.id
  source_security_group_id = aws_security_group.nodes-privatecanal-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privatecanal-example-com-ingress-tcp-2382to4000-masters-privatecanal-example-com" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privatecanal-example-com.id
  source_security_group_id = aws_security_group.nodes-privatecanal-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privatecanal-example-com-ingress-tcp-4003to65535-masters-privatecanal-example-com" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privatecanal-example-com.id
  source_security_group_id = aws_security_group.nodes-privatecanal-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-privatecanal-example-com-ingress-udp-1to65535-masters-privatecanal-example-com" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-privatecanal-example-com.id
  source_security_group_id = aws_security_group.nodes-privatecanal-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "https-elb-to-master" {
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-privatecanal-example-com.id
  source_security_group_id = aws_security_group.api-elb-privatecanal-example-com.id
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.api-elb-privatecanal-example-com.id
  to_port           = 4
  type              = "ingress"
}

resource "aws_subnet" "us-test-1a-privatecanal-example-com" {
  availability_zone                           = "us-test-1a"
  cidr_block                                  = "172.20.32.0/19"
  enable_resource_name_dns_a_record_on_launch = true
  private_dns_hostname_type_on_launch         = "resource-name"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "us-test-1a.privatecanal.example.com"
    "SubnetType"                                     = "Private"
    "kops.k8s.io/instance-group/master-us-test-1a"   = "true"
    "kops.k8s.io/instance-group/nodes"               = "true"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
    "kubernetes.io/role/internal-elb"                = "1"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_subnet" "utility-us-test-1a-privatecanal-example-com" {
  availability_zone                           = "us-test-1a"
  cidr_block                                  = "172.20.4.0/22"
  enable_resource_name_dns_a_record_on_launch = true
  private_dns_hostname_type_on_launch         = "resource-name"
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "utility-us-test-1a.privatecanal.example.com"
    "SubnetType"                                     = "Utility"
    "kops.k8s.io/instance-group/bastion"             = "true"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
    "kubernetes.io/role/elb"                         = "1"
  }
  vpc_id = aws_vpc.privatecanal-example-com.id
}

resource "aws_vpc" "privatecanal-example-com" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "172.20.0.0/16"
  enable_dns_hostnames             = true
  enable_dns_support               = true
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "privatecanal-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags = {
    "KubernetesCluster"                              = "privatecanal.example.com"
    "Name"                                           = "privatecanal.example.com"
    "kubernetes.io/cluster/privatecanal.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options_association" "privatecanal-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.privatecanal-example-com.id
  vpc_id          = aws_vpc.privatecanal-example-com.id
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
