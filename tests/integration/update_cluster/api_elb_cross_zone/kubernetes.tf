locals {
  cluster_name                 = "crosszone.example.com"
  master_autoscaling_group_ids = [aws_autoscaling_group.master-us-test-1a-masters-crosszone-example-com.id]
  master_security_group_ids    = [aws_security_group.masters-crosszone-example-com.id]
  masters_role_arn             = aws_iam_role.masters-crosszone-example-com.arn
  masters_role_name            = aws_iam_role.masters-crosszone-example-com.name
  node_autoscaling_group_ids   = [aws_autoscaling_group.nodes-crosszone-example-com.id]
  node_security_group_ids      = [aws_security_group.nodes-crosszone-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  node_subnet_ids              = [aws_subnet.us-test-1a-crosszone-example-com.id]
  nodes_role_arn               = aws_iam_role.nodes-crosszone-example-com.arn
  nodes_role_name              = aws_iam_role.nodes-crosszone-example-com.name
  region                       = "us-test-1"
  route_table_public_id        = aws_route_table.crosszone-example-com.id
  subnet_us-test-1a_id         = aws_subnet.us-test-1a-crosszone-example-com.id
  vpc_cidr_block               = aws_vpc.crosszone-example-com.cidr_block
  vpc_id                       = aws_vpc.crosszone-example-com.id
}

output "cluster_name" {
  value = "crosszone.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-crosszone-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-crosszone-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-crosszone-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-crosszone-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-crosszone-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-crosszone-example-com.id, "sg-exampleid3", "sg-exampleid4"]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-crosszone-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-crosszone-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-crosszone-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "route_table_public_id" {
  value = aws_route_table.crosszone-example-com.id
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-crosszone-example-com.id
}

output "vpc_cidr_block" {
  value = aws_vpc.crosszone-example-com.cidr_block
}

output "vpc_id" {
  value = aws_vpc.crosszone-example-com.id
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_attachment" "master-us-test-1a-masters-crosszone-example-com" {
  autoscaling_group_name = aws_autoscaling_group.master-us-test-1a-masters-crosszone-example-com.id
  elb                    = aws_elb.api-crosszone-example-com.id
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-crosszone-example-com" {
  enabled_metrics      = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_configuration = aws_launch_configuration.master-us-test-1a-masters-crosszone-example-com.id
  max_size             = 1
  metrics_granularity  = "1Minute"
  min_size             = 1
  name                 = "master-us-test-1a.masters.crosszone.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "crosszone.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.crosszone.example.com"
  }
  tag {
    key                 = "Owner"
    propagate_at_launch = true
    value               = "John Doe"
  }
  tag {
    key                 = "foo/bar"
    propagate_at_launch = true
    value               = "fib+baz"
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
    key                 = "kubernetes.io/cluster/crosszone.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-crosszone-example-com.id]
}

resource "aws_autoscaling_group" "nodes-crosszone-example-com" {
  enabled_metrics      = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_configuration = aws_launch_configuration.nodes-crosszone-example-com.id
  max_size             = 2
  metrics_granularity  = "1Minute"
  min_size             = 2
  name                 = "nodes.crosszone.example.com"
  suspended_processes  = ["AZRebalance"]
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "crosszone.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.crosszone.example.com"
  }
  tag {
    key                 = "Owner"
    propagate_at_launch = true
    value               = "John Doe"
  }
  tag {
    key                 = "foo/bar"
    propagate_at_launch = true
    value               = "fib+baz"
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
    key                 = "kubernetes.io/cluster/crosszone.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-crosszone-example-com.id]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-crosszone-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "us-test-1a.etcd-events.crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "k8s.io/etcd/events"                          = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                          = "1"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-crosszone-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "us-test-1a.etcd-main.crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "k8s.io/etcd/main"                            = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                          = "1"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_elb" "api-crosszone-example-com" {
  cross_zone_load_balancing = true
  health_check {
    healthy_threshold   = 2
    interval            = 10
    target              = "SSL:443"
    timeout             = 5
    unhealthy_threshold = 2
  }
  idle_timeout = 300
  listener {
    instance_port      = 443
    instance_protocol  = "TCP"
    lb_port            = 443
    lb_protocol        = "TCP"
    ssl_certificate_id = ""
  }
  name            = "api-crosszone-example-com-qhvtkl"
  security_groups = [aws_security_group.api-elb-crosszone-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  subnets         = [aws_subnet.us-test-1a-crosszone-example-com.id]
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "api.crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "masters-crosszone-example-com" {
  name = "masters.crosszone.example.com"
  role = aws_iam_role.masters-crosszone-example-com.name
}

resource "aws_iam_instance_profile" "nodes-crosszone-example-com" {
  name = "nodes.crosszone.example.com"
  role = aws_iam_role.nodes-crosszone-example-com.name
}

resource "aws_iam_role_policy" "masters-crosszone-example-com" {
  name   = "masters.crosszone.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.crosszone.example.com_policy")
  role   = aws_iam_role.masters-crosszone-example-com.name
}

resource "aws_iam_role_policy" "nodes-crosszone-example-com" {
  name   = "nodes.crosszone.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.crosszone.example.com_policy")
  role   = aws_iam_role.nodes-crosszone-example-com.name
}

resource "aws_iam_role" "masters-crosszone-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.crosszone.example.com_policy")
  name               = "masters.crosszone.example.com"
}

resource "aws_iam_role" "nodes-crosszone-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_nodes.crosszone.example.com_policy")
  name               = "nodes.crosszone.example.com"
}

resource "aws_internet_gateway" "crosszone-example-com" {
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
  vpc_id = aws_vpc.crosszone-example-com.id
}

resource "aws_key_pair" "kubernetes-crosszone-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.crosszone.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.crosszone.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
}

resource "aws_launch_configuration" "master-us-test-1a-masters-crosszone-example-com" {
  associate_public_ip_address = true
  enable_monitoring           = false
  ephemeral_block_device {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile = aws_iam_instance_profile.masters-crosszone-example-com.id
  image_id             = "ami-12345678"
  instance_type        = "m3.medium"
  key_name             = aws_key_pair.kubernetes-crosszone-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "master-us-test-1a.masters.crosszone.example.com-"
  root_block_device {
    delete_on_termination = true
    volume_size           = 64
    volume_type           = "gp2"
  }
  security_groups = [aws_security_group.masters-crosszone-example-com.id]
  user_data       = file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.crosszone.example.com_user_data")
}

resource "aws_launch_configuration" "nodes-crosszone-example-com" {
  associate_public_ip_address = true
  enable_monitoring           = true
  iam_instance_profile        = aws_iam_instance_profile.nodes-crosszone-example-com.id
  image_id                    = "ami-12345678"
  instance_type               = "t2.medium"
  key_name                    = aws_key_pair.kubernetes-crosszone-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "nodes.crosszone.example.com-"
  root_block_device {
    delete_on_termination = true
    volume_size           = 128
    volume_type           = "gp2"
  }
  security_groups = [aws_security_group.nodes-crosszone-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  user_data       = file("${path.module}/data/aws_launch_configuration_nodes.crosszone.example.com_user_data")
}

resource "aws_route53_record" "api-crosszone-example-com" {
  alias {
    evaluate_target_health = false
    name                   = aws_elb.api-crosszone-example-com.dns_name
    zone_id                = aws_elb.api-crosszone-example-com.zone_id
  }
  name    = "api.crosszone.example.com"
  type    = "A"
  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route_table_association" "us-test-1a-crosszone-example-com" {
  route_table_id = aws_route_table.crosszone-example-com.id
  subnet_id      = aws_subnet.us-test-1a-crosszone-example-com.id
}

resource "aws_route_table" "crosszone-example-com" {
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
    "kubernetes.io/kops/role"                     = "public"
  }
  vpc_id = aws_vpc.crosszone-example-com.id
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.crosszone-example-com.id
  route_table_id         = aws_route_table.crosszone-example-com.id
}

resource "aws_security_group_rule" "all-master-to-master" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-crosszone-example-com.id
  source_security_group_id = aws_security_group.masters-crosszone-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "all-master-to-node" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-crosszone-example-com.id
  source_security_group_id = aws_security_group.masters-crosszone-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "all-node-to-node" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-crosszone-example-com.id
  source_security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "api-elb-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-crosszone-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-crosszone-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "https-elb-to-master" {
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-crosszone-example-com.id
  source_security_group_id = aws_security_group.api-elb-crosszone-example-com.id
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.api-elb-crosszone-example-com.id
  to_port           = 4
  type              = "ingress"
}

resource "aws_security_group_rule" "master-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-crosszone-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "node-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-crosszone-example-com.id
  source_security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-tcp-2383-4000" {
  from_port                = 2383
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-crosszone-example-com.id
  source_security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-crosszone-example-com.id
  source_security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-crosszone-example-com.id
  source_security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodeport-tcp-external-to-node-1-2-3-4--32" {
  cidr_blocks       = ["1.2.3.4/32"]
  from_port         = 28000
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-tcp-external-to-node-10-20-30-0--24" {
  cidr_blocks       = ["10.20.30.0/24"]
  from_port         = 28000
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-udp-external-to-node-1-2-3-4--32" {
  cidr_blocks       = ["1.2.3.4/32"]
  from_port         = 28000
  protocol          = "udp"
  security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-udp-external-to-node-10-20-30-0--24" {
  cidr_blocks       = ["10.20.30.0/24"]
  from_port         = 28000
  protocol          = "udp"
  security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-master-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-crosszone-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-node-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-crosszone-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group" "api-elb-crosszone-example-com" {
  description = "Security group for api ELB"
  name        = "api-elb.crosszone.example.com"
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "api-elb.crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
  vpc_id = aws_vpc.crosszone-example-com.id
}

resource "aws_security_group" "masters-crosszone-example-com" {
  description = "Security group for masters"
  name        = "masters.crosszone.example.com"
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "masters.crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
  vpc_id = aws_vpc.crosszone-example-com.id
}

resource "aws_security_group" "nodes-crosszone-example-com" {
  description = "Security group for nodes"
  name        = "nodes.crosszone.example.com"
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "nodes.crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
  vpc_id = aws_vpc.crosszone-example-com.id
}

resource "aws_subnet" "us-test-1a-crosszone-example-com" {
  availability_zone = "us-test-1a"
  cidr_block        = "172.20.32.0/19"
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "us-test-1a.crosszone.example.com"
    "Owner"                                       = "John Doe"
    "SubnetType"                                  = "Public"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
    "kubernetes.io/role/elb"                      = "1"
  }
  vpc_id = aws_vpc.crosszone-example-com.id
}

resource "aws_vpc_dhcp_options_association" "crosszone-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.crosszone-example-com.id
  vpc_id          = aws_vpc.crosszone-example-com.id
}

resource "aws_vpc_dhcp_options" "crosszone-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
}

resource "aws_vpc" "crosszone-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    "KubernetesCluster"                           = "crosszone.example.com"
    "Name"                                        = "crosszone.example.com"
    "Owner"                                       = "John Doe"
    "foo/bar"                                     = "fib+baz"
    "kubernetes.io/cluster/crosszone.example.com" = "owned"
  }
}

terraform {
  required_version = ">= 0.12.0"
}
