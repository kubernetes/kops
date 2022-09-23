locals {
  cluster_name                      = "complex.example.com"
  master_autoscaling_group_ids      = [aws_autoscaling_group.master-us-test-1a-masters-complex-example-com.id]
  master_security_group_ids         = [aws_security_group.masters-complex-example-com.id, "sg-exampleid5", "sg-exampleid6"]
  masters_role_arn                  = aws_iam_role.masters-complex-example-com.arn
  masters_role_name                 = aws_iam_role.masters-complex-example-com.name
  node_autoscaling_group_ids        = [aws_autoscaling_group.nodes-complex-example-com.id]
  node_security_group_ids           = [aws_security_group.nodes-complex-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  node_subnet_ids                   = [aws_subnet.us-test-1a-complex-example-com.id]
  nodes_role_arn                    = aws_iam_role.nodes-complex-example-com.arn
  nodes_role_name                   = aws_iam_role.nodes-complex-example-com.name
  region                            = "us-test-1"
  route_table_private-us-test-1a_id = aws_route_table.private-us-test-1a-complex-example-com.id
  route_table_public_id             = aws_route_table.complex-example-com.id
  subnet_us-east-1a-private_id      = aws_subnet.us-east-1a-private-complex-example-com.id
  subnet_us-east-1a-utility_id      = aws_subnet.us-east-1a-utility-complex-example-com.id
  subnet_us-test-1a_id              = aws_subnet.us-test-1a-complex-example-com.id
  vpc_cidr_block                    = aws_vpc.complex-example-com.cidr_block
  vpc_id                            = aws_vpc.complex-example-com.id
}

output "cluster_name" {
  value = "complex.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-complex-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-complex-example-com.id, "sg-exampleid5", "sg-exampleid6"]
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

output "route_table_private-us-test-1a_id" {
  value = aws_route_table.private-us-test-1a-complex-example-com.id
}

output "route_table_public_id" {
  value = aws_route_table.complex-example-com.id
}

output "subnet_us-east-1a-private_id" {
  value = aws_subnet.us-east-1a-private-complex-example-com.id
}

output "subnet_us-east-1a-utility_id" {
  value = aws_subnet.us-east-1a-utility-complex-example-com.id
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
  region      = "us-test-1"
  max_retries = "10"
}

provider "aws" {
  alias   = "files"
  region  = "us-test-1"
  profile = "foo"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-complex-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.master-us-test-1a-masters-complex-example-com.id
    version = aws_launch_template.master-us-test-1a-masters-complex-example-com.latest_version
  }
  load_balancers        = ["my-external-lb-1"]
  max_instance_lifetime = 0
  max_size              = 1
  metrics_granularity   = "1Minute"
  min_size              = 1
  name                  = "master-us-test-1a.masters.complex.example.com"
  protect_from_scale_in = false
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
    key                 = "kubernetes.io/cluster/complex.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  target_group_arns   = [aws_lb_target_group.tcp-complex-example-com-vpjolq.id, aws_lb_target_group.tls-complex-example-com-5nursn.id]
  vpc_zone_identifier = [aws_subnet.us-test-1a-complex-example-com.id]
}

resource "aws_autoscaling_group" "nodes-complex-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.nodes-complex-example-com.id
    version = aws_launch_template.nodes-complex-example-com.latest_version
  }
  load_balancers        = ["my-external-lb-1"]
  max_instance_lifetime = 0
  max_size              = 2
  metrics_granularity   = "1Minute"
  min_size              = 2
  name                  = "nodes.complex.example.com"
  protect_from_scale_in = false
  suspended_processes   = ["AZRebalance"]
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

resource "aws_ebs_volume" "a-etcd-events-complex-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "a.etcd-events.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "k8s.io/etcd/events"                        = "a/a"
    "k8s.io/role/master"                        = "1"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_ebs_volume" "a-etcd-main-complex-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "a.etcd-main.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "k8s.io/etcd/main"                          = "a/a"
    "k8s.io/role/master"                        = "1"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_iam_instance_profile" "masters-complex-example-com" {
  name = "masters.complex.example.com"
  role = aws_iam_role.masters-complex-example-com.name
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "masters.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "nodes-complex-example-com" {
  name = "nodes.complex.example.com"
  role = aws_iam_role.nodes-complex-example-com.name
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "nodes.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
}

resource "aws_iam_role" "masters-complex-example-com" {
  assume_role_policy   = file("${path.module}/data/aws_iam_role_masters.complex.example.com_policy")
  name                 = "masters.complex.example.com"
  permissions_boundary = "arn:aws-test:iam::000000000000:policy/boundaries"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "masters.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
}

resource "aws_iam_role" "nodes-complex-example-com" {
  assume_role_policy   = file("${path.module}/data/aws_iam_role_nodes.complex.example.com_policy")
  name                 = "nodes.complex.example.com"
  permissions_boundary = "arn:aws-test:iam::000000000000:policy/boundaries"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "nodes.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
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
      iops                  = 3000
      kms_key_id            = "arn:aws-test:kms:us-test-1:000000000000:key/1234abcd-12ab-34cd-56ef-1234567890ab"
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
    name = aws_iam_instance_profile.masters-complex-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "m3.medium"
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "required"
  }
  monitoring {
    enabled = false
  }
  name = "master-us-test-1a.masters.complex.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.masters-complex-example-com.id, "sg-exampleid5", "sg-exampleid6"]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                                                     = "complex.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.complex.example.com"
      "Owner"                                                                                                 = "John Doe"
      "foo/bar"                                                                                               = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/complex.example.com"                                                             = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                                                     = "complex.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.complex.example.com"
      "Owner"                                                                                                 = "John Doe"
      "foo/bar"                                                                                               = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/complex.example.com"                                                             = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                                                     = "complex.example.com"
    "Name"                                                                                                  = "master-us-test-1a.masters.complex.example.com"
    "Owner"                                                                                                 = "John Doe"
    "foo/bar"                                                                                               = "fib+baz"
    "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
    "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
    "k8s.io/role/master"                                                                                    = "1"
    "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
    "kubernetes.io/cluster/complex.example.com"                                                             = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_master-us-test-1a.masters.complex.example.com_user_data")
}

resource "aws_launch_template" "nodes-complex-example-com" {
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
  block_device_mappings {
    device_name = "/dev/xvdd"
    ebs {
      delete_on_termination = true
      encrypted             = true
      kms_key_id            = "arn:aws-test:kms:us-test-1:000000000000:key/1234abcd-12ab-34cd-56ef-1234567890ab"
      volume_size           = 20
      volume_type           = "gp2"
    }
  }
  credit_specification {
    cpu_credits = "standard"
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes-complex-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.medium"
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
    enabled = true
  }
  name = "nodes.complex.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.nodes-complex-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "complex.example.com"
      "Name"                                                                       = "nodes.complex.example.com"
      "Owner"                                                                      = "John Doe"
      "foo/bar"                                                                    = "fib+baz"
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
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/node"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                  = "nodes"
    "kubernetes.io/cluster/complex.example.com"                                  = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_nodes.complex.example.com_user_data")
}

resource "aws_lb" "api-complex-example-com" {
  access_logs {
    bucket  = "access-log-example"
    enabled = true
  }
  enable_cross_zone_load_balancing = true
  internal                         = false
  load_balancer_type               = "network"
  name                             = "api-complex-example-com-vd3t5n"
  subnet_mapping {
    allocation_id = "eipalloc-012345a678b9cdefa"
    subnet_id     = aws_subnet.us-test-1a-complex-example-com.id
  }
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "api.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
}

resource "aws_lb_listener" "api-complex-example-com-443" {
  certificate_arn = "arn:aws-test:acm:us-test-1:000000000000:certificate/123456789012-1234-1234-1234-12345678"
  default_action {
    target_group_arn = aws_lb_target_group.tls-complex-example-com-5nursn.id
    type             = "forward"
  }
  load_balancer_arn = aws_lb.api-complex-example-com.id
  port              = 443
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
}

resource "aws_lb_listener" "api-complex-example-com-8443" {
  default_action {
    target_group_arn = aws_lb_target_group.tcp-complex-example-com-vpjolq.id
    type             = "forward"
  }
  load_balancer_arn = aws_lb.api-complex-example-com.id
  port              = 8443
  protocol          = "TCP"
}

resource "aws_lb_target_group" "tcp-complex-example-com-vpjolq" {
  health_check {
    healthy_threshold   = 2
    protocol            = "TCP"
    unhealthy_threshold = 2
  }
  name     = "tcp-complex-example-com-vpjolq"
  port     = 443
  protocol = "TCP"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "tcp-complex-example-com-vpjolq"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_lb_target_group" "tls-complex-example-com-5nursn" {
  health_check {
    healthy_threshold   = 2
    protocol            = "TCP"
    unhealthy_threshold = 2
  }
  name     = "tls-complex-example-com-5nursn"
  port     = 443
  protocol = "TLS"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "tls-complex-example-com-5nursn"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.complex-example-com.id
  route_table_id         = aws_route_table.complex-example-com.id
}

resource "aws_route" "route-__--0" {
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.complex-example-com.id
  route_table_id              = aws_route_table.complex-example-com.id
}

resource "aws_route" "route-private-us-test-1a-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  route_table_id         = aws_route_table.private-us-test-1a-complex-example-com.id
  transit_gateway_id     = "tgw-123456"
}

resource "aws_route" "route-us-east-1a-private-192-168-1-10--32" {
  destination_cidr_block = "192.168.1.10/32"
  route_table_id         = aws_route_table.private-us-test-1a-complex-example-com.id
  transit_gateway_id     = "tgw-0123456"
}

resource "aws_route53_record" "api-complex-example-com" {
  alias {
    evaluate_target_health = false
    name                   = aws_lb.api-complex-example-com.dns_name
    zone_id                = aws_lb.api-complex-example-com.zone_id
  }
  name    = "api.complex.example.com"
  type    = "A"
  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
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

resource "aws_route_table" "private-us-test-1a-complex-example-com" {
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "private-us-test-1a.complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
    "kubernetes.io/kops/role"                   = "private-us-test-1a"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_route_table_association" "private-us-east-1a-private-complex-example-com" {
  route_table_id = aws_route_table.private-us-test-1a-complex-example-com.id
  subnet_id      = aws_subnet.us-east-1a-private-complex-example-com.id
}

resource "aws_route_table_association" "us-east-1a-utility-complex-example-com" {
  route_table_id = aws_route_table.complex-example-com.id
  subnet_id      = aws_subnet.us-east-1a-utility-complex-example-com.id
}

resource "aws_route_table_association" "us-test-1a-complex-example-com" {
  route_table_id = aws_route_table.complex-example-com.id
  subnet_id      = aws_subnet.us-test-1a-complex-example-com.id
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "clusters.example.com/complex.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-authentication-aws-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-authentication.aws-k8s-1.12_content")
  key                    = "clusters.example.com/complex.example.com/addons/authentication.aws/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-aws-cloud-controller-addons-k8s-io-k8s-1-18" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-aws-cloud-controller.addons.k8s.io-k8s-1.18_content")
  key                    = "clusters.example.com/complex.example.com/addons/aws-cloud-controller.addons.k8s.io/k8s-1.18.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-aws-ebs-csi-driver-addons-k8s-io-k8s-1-17" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-aws-ebs-csi-driver.addons.k8s.io-k8s-1.17_content")
  key                    = "clusters.example.com/complex.example.com/addons/aws-ebs-csi-driver.addons.k8s.io/k8s-1.17.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-bootstrap_content")
  key                    = "clusters.example.com/complex.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/complex.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/complex.example.com/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "clusters.example.com/complex.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "clusters.example.com/complex.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-leader-migration-rbac-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-leader-migration.rbac.addons.k8s.io-k8s-1.23_content")
  key                    = "clusters.example.com/complex.example.com/addons/leader-migration.rbac.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "clusters.example.com/complex.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "complex-example-com-addons-storage-aws-addons-k8s-io-v1-15-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_complex.example.com-addons-storage-aws.addons.k8s.io-v1.15.0_content")
  key                    = "clusters.example.com/complex.example.com/addons/storage-aws.addons.k8s.io/v1.15.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "clusters.example.com/complex.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "clusters.example.com/complex.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "clusters.example.com/complex.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-master-us-test-1a_content")
  key                    = "clusters.example.com/complex.example.com/manifests/etcd/events-master-us-test-1a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-master-us-test-1a_content")
  key                    = "clusters.example.com/complex.example.com/manifests/etcd/main-master-us-test-1a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "clusters.example.com/complex.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test-1a_content")
  key                    = "clusters.example.com/complex.example.com/igconfig/master/master-us-test-1a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "clusters.example.com/complex.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
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

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-22to22-masters-complex-example-com" {
  from_port         = 22
  prefix_list_ids   = ["pl-66666666"]
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-22to22-nodes-complex-example-com" {
  from_port         = 22
  prefix_list_ids   = ["pl-66666666"]
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-443to443-masters-complex-example-com" {
  from_port         = 443
  prefix_list_ids   = ["pl-44444444"]
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "from-1-1-1-0--24-ingress-tcp-443to443-masters-complex-example-com" {
  cidr_blocks       = ["1.1.1.0/24"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "from-1-1-1-1--32-ingress-tcp-22to22-masters-complex-example-com" {
  cidr_blocks       = ["1.1.1.1/32"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-1-1-1-1--32-ingress-tcp-22to22-nodes-complex-example-com" {
  cidr_blocks       = ["1.1.1.1/32"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-masters-complex-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-complex-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-complex-example-com-ingress-all-0to0-masters-complex-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.masters-complex-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-masters-complex-example-com-ingress-all-0to0-nodes-complex-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-complex-example-com.id
  source_security_group_id = aws_security_group.masters-complex-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-complex-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-complex-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-complex-example-com-ingress-all-0to0-nodes-complex-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-complex-example-com-ingress-tcp-1to2379-masters-complex-example-com" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-complex-example-com-ingress-tcp-2382to4000-masters-complex-example-com" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-complex-example-com-ingress-tcp-4003to65535-masters-complex-example-com" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-complex-example-com-ingress-udp-1to65535-masters-complex-example-com" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-complex-example-com.id
  source_security_group_id = aws_security_group.nodes-complex-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "https-elb-to-master" {
  cidr_blocks       = ["172.20.0.0/16"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "https-lb-to-master-10-1-0-0--16" {
  cidr_blocks       = ["10.1.0.0/16"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "https-lb-to-master-10-2-0-0--16" {
  cidr_blocks       = ["10.2.0.0/16"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-1-1-1-0--24" {
  cidr_blocks       = ["1.1.1.0/24"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 4
  type              = "ingress"
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

resource "aws_security_group_rule" "tcp-api-1-1-1-0--24" {
  cidr_blocks       = ["1.1.1.0/24"]
  from_port         = 8443
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 8443
  type              = "ingress"
}

resource "aws_security_group_rule" "tcp-api-pl-44444444" {
  from_port         = 8443
  prefix_list_ids   = ["pl-44444444"]
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-complex-example-com.id
  to_port           = 8443
  type              = "ingress"
}

resource "aws_subnet" "us-east-1a-private-complex-example-com" {
  availability_zone                           = "us-test-1a"
  cidr_block                                  = "172.20.64.0/19"
  enable_resource_name_dns_a_record_on_launch = true
  private_dns_hostname_type_on_launch         = "resource-name"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "us-east-1a-private.complex.example.com"
    "Owner"                                     = "John Doe"
    "SubnetType"                                = "Private"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
    "kubernetes.io/role/internal-elb"           = "1"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_subnet" "us-east-1a-utility-complex-example-com" {
  availability_zone                           = "us-test-1a"
  cidr_block                                  = "172.20.96.0/19"
  enable_resource_name_dns_a_record_on_launch = true
  private_dns_hostname_type_on_launch         = "resource-name"
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "us-east-1a-utility.complex.example.com"
    "Owner"                                     = "John Doe"
    "SubnetType"                                = "Utility"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
    "kubernetes.io/role/elb"                    = "1"
    "kubernetes.io/role/internal-elb"           = "1"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_subnet" "us-test-1a-complex-example-com" {
  availability_zone                           = "us-test-1a"
  cidr_block                                  = "172.20.32.0/19"
  enable_resource_name_dns_a_record_on_launch = true
  private_dns_hostname_type_on_launch         = "resource-name"
  tags = {
    "KubernetesCluster"                            = "complex.example.com"
    "Name"                                         = "us-test-1a.complex.example.com"
    "Owner"                                        = "John Doe"
    "SubnetType"                                   = "Public"
    "foo/bar"                                      = "fib+baz"
    "kops.k8s.io/instance-group/master-us-test-1a" = "true"
    "kops.k8s.io/instance-group/nodes"             = "true"
    "kubernetes.io/cluster/complex.example.com"    = "owned"
    "kubernetes.io/role/elb"                       = "1"
    "kubernetes.io/role/internal-elb"              = "1"
  }
  vpc_id = aws_vpc.complex-example-com.id
}

resource "aws_vpc" "complex-example-com" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "172.20.0.0/16"
  enable_dns_hostnames             = true
  enable_dns_support               = true
  tags = {
    "KubernetesCluster"                         = "complex.example.com"
    "Name"                                      = "complex.example.com"
    "Owner"                                     = "John Doe"
    "foo/bar"                                   = "fib+baz"
    "kubernetes.io/cluster/complex.example.com" = "owned"
  }
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

resource "aws_vpc_dhcp_options_association" "complex-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.complex-example-com.id
  vpc_id          = aws_vpc.complex-example-com.id
}

resource "aws_vpc_ipv4_cidr_block_association" "cidr-10-1-0-0--16" {
  cidr_block = "10.1.0.0/16"
  vpc_id     = aws_vpc.complex-example-com.id
}

resource "aws_vpc_ipv4_cidr_block_association" "cidr-10-2-0-0--16" {
  cidr_block = "10.2.0.0/16"
  vpc_id     = aws_vpc.complex-example-com.id
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
