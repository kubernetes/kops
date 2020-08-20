locals {
  cluster_name                 = "complex.example.com"
  master_autoscaling_group_ids = [aws_autoscaling_group.master-us-test-1a-masters-complex-example-com.id]
  master_security_group_ids    = [aws_security_group.masters-complex-example-com.id]
  masters_role_arn             = aws_iam_role.masters-complex-example-com.arn
  masters_role_name            = aws_iam_role.masters-complex-example-com.name
  node_autoscaling_group_ids   = [aws_autoscaling_group.nodes-complex-example-com.id]
  node_security_group_ids      = [aws_security_group.nodes-complex-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  node_subnet_ids              = [aws_subnet.us-test-1a-complex-example-com.id]
  nodes_role_arn               = aws_iam_role.nodes-complex-example-com.arn
  nodes_role_name              = aws_iam_role.nodes-complex-example-com.name
  region                       = "us-test-1"
  route_table_public_id        = aws_route_table.complex-example-com.id
  subnet_us-test-1a_id         = aws_subnet.us-test-1a-complex-example-com.id
  vpc_cidr_block               = aws_vpc.complex-example-com.cidr_block
  vpc_id                       = aws_vpc.complex-example-com.id
}

output "cluster_name" {
  value = "complex.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-complex-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-complex-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-complex-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-complex-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-complex-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-complex-example-com.id, "sg-exampleid3", "sg-exampleid4"]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-complex-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-complex-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-complex-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "route_table_public_id" {
  value = aws_route_table.complex-example-com.id
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-complex-example-com.id
}

output "vpc_cidr_block" {
  value = aws_vpc.complex-example-com.cidr_block
}

output "vpc_id" {
  value = aws_vpc.complex-example-com.id
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_attachment" "master-us-test-1a-masters-complex-example-com" {
  autoscaling_group_name = aws_autoscaling_group.master-us-test-1a-masters-complex-example-com.id
  elb                    = aws_elb.api-complex-example-com.id
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-complex-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.master-us-test-1a-masters-complex-example-com.id
    version = aws_launch_template.master-us-test-1a-masters-complex-example-com.latest_version
  }
  max_size            = 1
  metrics_granularity = "1Minute"
  min_size            = 1
  name                = "master-us-test-1a.masters.complex.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "complex.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.complex.example.com"
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
    key                 = "kubernetes.io/cluster/complex.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-complex-example-com.id]
}

resource "aws_autoscaling_group" "nodes-complex-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.nodes-complex-example-com.id
    version = aws_launch_template.nodes-complex-example-com.latest_version
  }
  max_size            = 2
  metrics_granularity = "1Minute"
  min_size            = 2
  name                = "nodes.complex.example.com"
  suspended_processes = ["AZRebalance"]
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "complex.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.complex.example.com"
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
    key                 = "kubernetes.io/cluster/complex.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-complex-example-com.id]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-complex-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "us-test-1a.etcd-events.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "k8s.io/etcd/events"                        = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                        = "1"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-complex-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "us-test-1a.etcd-main.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "k8s.io/etcd/main"                          = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                        = "1"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_elb" "api-complex-example-com" {
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
    instance_protocol  = "SSL"
    lb_port            = 443
    lb_protocol        = "SSL"
    ssl_certificate_id = "arn:aws:acm:us-test-1:000000000000:certificate/123456789012-1234-1234-1234-12345678"
  }
  listener {
    instance_port     = 8443
    instance_protocol = "TCP"
    lb_port           = 8443
    lb_protocol       = "TCP"
  }
  name            = "api-complex-example-com-vd3t5n"
  security_groups = [aws_security_group.api-elb-complex-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  subnets         = [aws_subnet.us-test-1a-complex-example-com.id]
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "api.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "masters-complex-example-com" {
  name = "masters.complex.example.com"
  role = aws_iam_role.masters-complex-example-com.name
}

resource "aws_iam_instance_profile" "nodes-complex-example-com" {
  name = "nodes.complex.example.com"
  role = aws_iam_role.nodes-complex-example-com.name
}

resource "aws_iam_role_policy" "masters-complex-example-com" {
  name   = "masters.complex.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.complex.example.com_policy")
  role   = aws_iam_role.masters-complex-example-com.name
}

resource "aws_iam_role_policy" "nodes-complex-example-com" {
  name   = "nodes.complex.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.complex.example.com_policy")
  role   = aws_iam_role.nodes-complex-example-com.name
}

resource "aws_iam_role" "masters-complex-example-com" {
  assume_role_policy   = file("${path.module}/data/aws_iam_role_masters.complex.example.com_policy")
  name                 = "masters.complex.example.com"
  permissions_boundary = "arn:aws:iam:00000000000:policy/boundaries"
}

resource "aws_iam_role" "nodes-complex-example-com" {
  assume_role_policy   = file("${path.module}/data/aws_iam_role_nodes.complex.example.com_policy")
  name                 = "nodes.complex.example.com"
  permissions_boundary = "arn:aws:iam:00000000000:policy/boundaries"
}

resource "aws_internet_gateway" "complex-example-com" {
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_launch_template" "master-us-test-1a-masters-complex-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      volume_size           = 64
      volume_type           = "gp2"
    }
  }
  block_device_mappings {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.masters-complex-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "m3.medium"
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "master-us-test-1a.masters.complex.example.com-"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    security_groups             = [aws_security_group.masters-complex-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                            = "complex.example.com"
      "Name"                                                                         = "master-us-test-1a.masters.complex.example.com"
      "Owner"                                                                        = "John Doe"
      "foo/bar"                                                                      = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
      "k8s.io/role/master"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
      "kubernetes.io/cluster/complex.example.com"                                    = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                            = "complex.example.com"
      "Name"                                                                         = "master-us-test-1a.masters.complex.example.com"
      "Owner"                                                                        = "John Doe"
      "foo/bar"                                                                      = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
      "k8s.io/role/master"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
      "kubernetes.io/cluster/complex.example.com"                                    = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                            = "complex.example.com"
    "Name"                                                                         = "master-us-test-1a.masters.complex.example.com"
    "Owner"                                                                        = "John Doe"
    "foo/bar"                                                                      = "fib+baz"
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
    "k8s.io/role/master"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
    "kubernetes.io/cluster/complex.example.com"                                    = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_master-us-test-1a.masters.complex.example.com_user_data")
}

resource "aws_launch_template" "nodes-complex-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      volume_size           = 128
      volume_type           = "gp2"
    }
  }
  block_device_mappings {
    device_name = "/dev/xvdd"
    ebs {
      delete_on_termination = true
      volume_size           = 20
      volume_type           = "gp2"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes-complex-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.medium"
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "nodes.complex.example.com-"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    security_groups             = [aws_security_group.nodes-complex-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "complex.example.com"
      "Name"                                                                       = "nodes.complex.example.com"
      "Owner"                                                                      = "John Doe"
      "foo/bar"                                                                    = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/complex.example.com"                                  = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "complex.example.com"
      "Name"                                                                       = "nodes.complex.example.com"
      "Owner"                                                                      = "John Doe"
      "foo/bar"                                                                    = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/complex.example.com"                                  = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "complex.example.com"
    "Name"                                                                       = "nodes.complex.example.com"
    "Owner"                                                                      = "John Doe"
    "foo/bar"                                                                    = "fib+baz"
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/node"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                  = "nodes"
    "kubernetes.io/cluster/complex.example.com"                                  = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_nodes.complex.example.com_user_data")
}

resource "aws_route53_record" "api-complex-example-com" {
  alias {
    evaluate_target_health = false
    name                   = aws_elb.api-complex-example-com.dns_name
    zone_id                = aws_elb.api-complex-example-com.zone_id
  }
  name    = "api.complex.example.com"
  type    = "A"
  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route_table_association" "us-test-1a-complex-example-com" {
  route_table_id = aws_route_table.complex-example-com.id
  subnet_id      = aws_subnet.us-test-1a-complex-example-com.id
}

resource "aws_route_table" "complex-example-com" {
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
    "kubernetes.io/kops/role"                   = "public"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.complex-example-com.id
  route_table_id         = aws_route_table.complex-example-com.id
}

resource "aws_security_group_rule" "all-master-to-master" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.masters-complex-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "all-master-to-node" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-complex-example-com.id
  source_security_group_id = aws_security_group.masters-complex-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "all-node-to-node" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "api-elb-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "https-api-elb-1-1-1-0--24" {
  cidr_blocks       = ["1.1.1.0/24"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "https-api-elb-2001_0_8500__--40" {
  cidr_blocks       = ["2001:0:8500::/40"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "https-elb-to-master" {
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-1-1-1-0--24" {
  cidr_blocks       = ["1.1.1.0/24"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port           = 4
  type              = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-2001_0_8500__--40" {
  cidr_blocks       = ["2001:0:8500::/40"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port           = 4
  type              = "ingress"
}

resource "aws_security_group_rule" "master-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "node-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-tcp-2382-4000" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodeport-tcp-external-to-node-1-2-3-4--32" {
  cidr_blocks       = ["1.2.3.4/32"]
  from_port         = 28000
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-tcp-external-to-node-10-20-30-0--24" {
  cidr_blocks       = ["10.20.30.0/24"]
  from_port         = 28000
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-udp-external-to-node-1-2-3-4--32" {
  cidr_blocks       = ["1.2.3.4/32"]
  from_port         = 28000
  protocol          = "udp"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-udp-external-to-node-10-20-30-0--24" {
  cidr_blocks       = ["10.20.30.0/24"]
  from_port         = 28000
  protocol          = "udp"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-master-1-1-1-1--32" {
  cidr_blocks       = ["1.1.1.1/32"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-master-2001_0_85a3__--48" {
  cidr_blocks       = ["2001:0:85a3::/48"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-node-1-1-1-1--32" {
  cidr_blocks       = ["1.1.1.1/32"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-node-2001_0_85a3__--48" {
  cidr_blocks       = ["2001:0:85a3::/48"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "tcp-api-elb-1-1-1-0--24" {
  cidr_blocks       = ["1.1.1.0/24"]
  from_port         = 8443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port           = 8443
  type              = "ingress"
}

resource "aws_security_group_rule" "tcp-api-elb-2001_0_8500__--40" {
  cidr_blocks       = ["2001:0:8500::/40"]
  from_port         = 8443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port           = 8443
  type              = "ingress"
}

resource "aws_security_group_rule" "tcp-elb-to-master" {
  from_port                = 8443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.api-elb-complex-example-com.id
  to_port                  = 8443
  type                     = "ingress"
}

resource "aws_security_group" "api-elb-complex-example-com" {
  description = "Security group for api ELB"
  name        = "api-elb.complex.example.com"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "api-elb.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_security_group" "masters-complex-example-com" {
  description = "Security group for masters"
  name        = "masters.complex.example.com"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "masters.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_security_group" "nodes-complex-example-com" {
  description = "Security group for nodes"
  name        = "nodes.complex.example.com"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "nodes.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_subnet" "us-test-1a-complex-example-com" {
  availability_zone = "us-test-1a"
  cidr_block        = "172.20.32.0/19"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "us-test-1a.complex.example.com"
    "Owner"                                     = "John Doe"
    "SubnetType"                                = "Public"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
    "kubernetes.io/role/elb"                    = "1"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_vpc_dhcp_options_association" "complex-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.complex-example-com.id
  vpc_id          = aws_vpc.complex-example-com.id
}

resource "aws_vpc_dhcp_options" "complex-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "cidr-10-1-0-0--16" {
  cidr_block = "10.1.0.0/16"
  vpc_id     = aws_vpc.complex-example-com.id
}

resource "aws_vpc_ipv4_cidr_block_association" "cidr-10-2-0-0--16" {
  cidr_block = "10.2.0.0/16"
  vpc_id     = aws_vpc.complex-example-com.id
}

resource "aws_vpc" "complex-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
}

terraform {
  required_version = ">= 0.12.0"
}
