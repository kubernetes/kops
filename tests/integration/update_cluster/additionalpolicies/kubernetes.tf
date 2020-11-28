locals {
  cluster_name                 = "additionalpolicies.example.com"
  master-us-test-1a_role_arn   = aws_iam_role.masters-additionalpolicies-example-com.arn
  master-us-test-1a_role_name  = aws_iam_role.masters-additionalpolicies-example-com.name
  master_autoscaling_group_ids = [aws_autoscaling_group.master-us-test-1a-masters-additionalpolicies-example-com.id]
  master_security_group_ids    = [aws_security_group.masters-additionalpolicies-example-com.id]
  node_autoscaling_group_ids   = [aws_autoscaling_group.nodes-additionalpolicies-example-com.id]
  node_security_group_ids      = [aws_security_group.nodes-additionalpolicies-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  node_subnet_ids              = [aws_subnet.us-test-1a-additionalpolicies-example-com.id]
  nodes_role_arn               = aws_iam_role.ig-nodes-additionalpolicies-example-com.arn
  nodes_role_name              = aws_iam_role.ig-nodes-additionalpolicies-example-com.name
  region                       = "us-test-1"
  route_table_public_id        = aws_route_table.additionalpolicies-example-com.id
  subnet_us-test-1a_id         = aws_subnet.us-test-1a-additionalpolicies-example-com.id
  vpc_cidr_block               = aws_vpc.additionalpolicies-example-com.cidr_block
  vpc_id                       = aws_vpc.additionalpolicies-example-com.id
}

output "cluster_name" {
  value = "additionalpolicies.example.com"
}

output "master-us-test-1a_role_arn" {
  value = aws_iam_role.masters-additionalpolicies-example-com.arn
}

output "master-us-test-1a_role_name" {
  value = aws_iam_role.masters-additionalpolicies-example-com.name
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-additionalpolicies-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-additionalpolicies-example-com.id]
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-additionalpolicies-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-additionalpolicies-example-com.id, "sg-exampleid3", "sg-exampleid4"]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-additionalpolicies-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.ig-nodes-additionalpolicies-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.ig-nodes-additionalpolicies-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "route_table_public_id" {
  value = aws_route_table.additionalpolicies-example-com.id
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-additionalpolicies-example-com.id
}

output "vpc_cidr_block" {
  value = aws_vpc.additionalpolicies-example-com.cidr_block
}

output "vpc_id" {
  value = aws_vpc.additionalpolicies-example-com.id
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-additionalpolicies-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.master-us-test-1a-masters-additionalpolicies-example-com.id
    version = aws_launch_template.master-us-test-1a-masters-additionalpolicies-example-com.latest_version
  }
  load_balancers      = [aws_elb.api-additionalpolicies-example-com.id]
  max_size            = 1
  metrics_granularity = "1Minute"
  min_size            = 1
  name                = "master-us-test-1a.masters.additionalpolicies.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "additionalpolicies.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.additionalpolicies.example.com"
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
    key                 = "kubernetes.io/cluster/additionalpolicies.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-additionalpolicies-example-com.id]
}

resource "aws_autoscaling_group" "nodes-additionalpolicies-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.nodes-additionalpolicies-example-com.id
    version = aws_launch_template.nodes-additionalpolicies-example-com.latest_version
  }
  max_size            = 2
  metrics_granularity = "1Minute"
  min_size            = 2
  name                = "nodes.additionalpolicies.example.com"
  suspended_processes = ["AZRebalance"]
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "additionalpolicies.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.additionalpolicies.example.com"
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
    key                 = "kubernetes.io/cluster/additionalpolicies.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-additionalpolicies-example-com.id]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-additionalpolicies-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "us-test-1a.etcd-events.additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "k8s.io/etcd/events"                                   = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                                   = "1"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-additionalpolicies-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "us-test-1a.etcd-main.additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "k8s.io/etcd/main"                                     = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                                   = "1"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_elb" "api-additionalpolicies-example-com" {
  cross_zone_load_balancing = false
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
  name            = "api-additionalpolicies-ex-bjl7m1"
  security_groups = [aws_security_group.api-elb-additionalpolicies-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  subnets         = [aws_subnet.us-test-1a-additionalpolicies-example-com.id]
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "api.additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "ig-nodes-additionalpolicies-example-com" {
  name = "ig-nodes.additionalpolicies.example.com"
  role = aws_iam_role.ig-nodes-additionalpolicies-example-com.name
}

resource "aws_iam_instance_profile" "masters-additionalpolicies-example-com" {
  name = "masters.additionalpolicies.example.com"
  role = aws_iam_role.masters-additionalpolicies-example-com.name
}

resource "aws_iam_role_policy" "additional-ig-ig-nodes-additionalpolicies-example-com" {
  name   = "additional-ig.ig-nodes.additionalpolicies.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_additional-ig.ig-nodes.additionalpolicies.example.com_policy")
  role   = aws_iam_role.ig-nodes-additionalpolicies-example-com.name
}

resource "aws_iam_role_policy" "additional-ig-nodes-additionalpolicies-example-com" {
  name   = "additional.ig-nodes.additionalpolicies.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_additional.ig-nodes.additionalpolicies.example.com_policy")
  role   = aws_iam_role.ig-nodes-additionalpolicies-example-com.name
}

resource "aws_iam_role_policy" "additional-masters-additionalpolicies-example-com" {
  name   = "additional.masters.additionalpolicies.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_additional.masters.additionalpolicies.example.com_policy")
  role   = aws_iam_role.masters-additionalpolicies-example-com.name
}

resource "aws_iam_role_policy" "ig-nodes-additionalpolicies-example-com" {
  name   = "ig-nodes.additionalpolicies.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_ig-nodes.additionalpolicies.example.com_policy")
  role   = aws_iam_role.ig-nodes-additionalpolicies-example-com.name
}

resource "aws_iam_role_policy" "masters-additionalpolicies-example-com" {
  name   = "masters.additionalpolicies.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.additionalpolicies.example.com_policy")
  role   = aws_iam_role.masters-additionalpolicies-example-com.name
}

resource "aws_iam_role" "ig-nodes-additionalpolicies-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_ig-nodes.additionalpolicies.example.com_policy")
  name               = "ig-nodes.additionalpolicies.example.com"
}

resource "aws_iam_role" "masters-additionalpolicies-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.additionalpolicies.example.com_policy")
  name               = "masters.additionalpolicies.example.com"
}

resource "aws_internet_gateway" "additionalpolicies-example-com" {
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
  vpc_id = aws_vpc.additionalpolicies-example-com.id
}

resource "aws_key_pair" "kubernetes-additionalpolicies-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.additionalpolicies.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.additionalpolicies.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
}

resource "aws_launch_template" "master-us-test-1a-masters-additionalpolicies-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      volume_size           = 64
      volume_type           = "gp2"
    }
  }
  block_device_mappings {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.masters-additionalpolicies-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "m3.medium"
  key_name      = aws_key_pair.kubernetes-additionalpolicies-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  name = "master-us-test-1a.masters.additionalpolicies.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    security_groups             = [aws_security_group.masters-additionalpolicies-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                            = "additionalpolicies.example.com"
      "Name"                                                                         = "master-us-test-1a.masters.additionalpolicies.example.com"
      "Owner"                                                                        = "John Doe"
      "foo/bar"                                                                      = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
      "k8s.io/role/master"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
      "kubernetes.io/cluster/additionalpolicies.example.com"                         = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                            = "additionalpolicies.example.com"
      "Name"                                                                         = "master-us-test-1a.masters.additionalpolicies.example.com"
      "Owner"                                                                        = "John Doe"
      "foo/bar"                                                                      = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
      "k8s.io/role/master"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
      "kubernetes.io/cluster/additionalpolicies.example.com"                         = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                            = "additionalpolicies.example.com"
    "Name"                                                                         = "master-us-test-1a.masters.additionalpolicies.example.com"
    "Owner"                                                                        = "John Doe"
    "foo/bar"                                                                      = "fib+baz"
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"             = "master"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/master" = ""
    "k8s.io/role/master"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                    = "master-us-test-1a"
    "kubernetes.io/cluster/additionalpolicies.example.com"                         = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_master-us-test-1a.masters.additionalpolicies.example.com_user_data")
}

resource "aws_launch_template" "nodes-additionalpolicies-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      volume_size           = 128
      volume_type           = "gp2"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.ig-nodes-additionalpolicies-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.medium"
  key_name      = aws_key_pair.kubernetes-additionalpolicies-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  monitoring {
    enabled = true
  }
  name = "nodes.additionalpolicies.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    security_groups             = [aws_security_group.nodes-additionalpolicies-example-com.id, "sg-exampleid3", "sg-exampleid4"]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "additionalpolicies.example.com"
      "Name"                                                                       = "nodes.additionalpolicies.example.com"
      "Owner"                                                                      = "John Doe"
      "foo/bar"                                                                    = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/additionalpolicies.example.com"                       = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "additionalpolicies.example.com"
      "Name"                                                                       = "nodes.additionalpolicies.example.com"
      "Owner"                                                                      = "John Doe"
      "foo/bar"                                                                    = "fib+baz"
      "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/additionalpolicies.example.com"                       = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "additionalpolicies.example.com"
    "Name"                                                                       = "nodes.additionalpolicies.example.com"
    "Owner"                                                                      = "John Doe"
    "foo/bar"                                                                    = "fib+baz"
    "k8s.io/cluster-autoscaler/node-template/label/kubernetes.io/role"           = "node"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/node"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                  = "nodes"
    "kubernetes.io/cluster/additionalpolicies.example.com"                       = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_nodes.additionalpolicies.example.com_user_data")
}

resource "aws_route53_record" "api-additionalpolicies-example-com" {
  alias {
    evaluate_target_health = false
    name                   = aws_elb.api-additionalpolicies-example-com.dns_name
    zone_id                = aws_elb.api-additionalpolicies-example-com.zone_id
  }
  name    = "api.additionalpolicies.example.com"
  type    = "A"
  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route_table_association" "us-test-1a-additionalpolicies-example-com" {
  route_table_id = aws_route_table.additionalpolicies-example-com.id
  subnet_id      = aws_subnet.us-test-1a-additionalpolicies-example-com.id
}

resource "aws_route_table" "additionalpolicies-example-com" {
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
    "kubernetes.io/kops/role"                              = "public"
  }
  vpc_id = aws_vpc.additionalpolicies-example-com.id
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.additionalpolicies-example-com.id
  route_table_id         = aws_route_table.additionalpolicies-example-com.id
}

resource "aws_security_group_rule" "api-elb-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-additionalpolicies-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-additionalpolicies-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "https-elb-to-master" {
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-additionalpolicies-example-com.id
  source_security_group_id = aws_security_group.api-elb-additionalpolicies-example-com.id
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.api-elb-additionalpolicies-example-com.id
  to_port           = 4
  type              = "ingress"
}

resource "aws_security_group_rule" "masters-additionalpolicies-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-additionalpolicies-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "masters-additionalpolicies-example-com-ingress-all-0to0-masters-additionalpolicies-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-additionalpolicies-example-com.id
  source_security_group_id = aws_security_group.masters-additionalpolicies-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "masters-additionalpolicies-example-com-ingress-all-0to0-nodes-additionalpolicies-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-additionalpolicies-example-com.id
  source_security_group_id = aws_security_group.masters-additionalpolicies-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodeport-tcp-external-to-node-1-2-3-4--32" {
  cidr_blocks       = ["1.2.3.4/32"]
  from_port         = 28000
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-tcp-external-to-node-10-20-30-0--24" {
  cidr_blocks       = ["10.20.30.0/24"]
  from_port         = 28000
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-udp-external-to-node-1-2-3-4--32" {
  cidr_blocks       = ["1.2.3.4/32"]
  from_port         = 28000
  protocol          = "udp"
  security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodeport-udp-external-to-node-10-20-30-0--24" {
  cidr_blocks       = ["10.20.30.0/24"]
  from_port         = 28000
  protocol          = "udp"
  security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port           = 32767
  type              = "ingress"
}

resource "aws_security_group_rule" "nodes-additionalpolicies-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "nodes-additionalpolicies-example-com-ingress-all-0to0-nodes-additionalpolicies-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-additionalpolicies-example-com.id
  source_security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-additionalpolicies-example-com-ingress-tcp-1to2379-masters-additionalpolicies-example-com" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-additionalpolicies-example-com.id
  source_security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-additionalpolicies-example-com-ingress-tcp-2382to4000-masters-additionalpolicies-example-com" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-additionalpolicies-example-com.id
  source_security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-additionalpolicies-example-com-ingress-tcp-4003to65535-masters-additionalpolicies-example-com" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-additionalpolicies-example-com.id
  source_security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "nodes-additionalpolicies-example-com-ingress-udp-1to65535-masters-additionalpolicies-example-com" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-additionalpolicies-example-com.id
  source_security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-master-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-additionalpolicies-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-node-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-additionalpolicies-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group" "api-elb-additionalpolicies-example-com" {
  description = "Security group for api ELB"
  name        = "api-elb.additionalpolicies.example.com"
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "api-elb.additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
  vpc_id = aws_vpc.additionalpolicies-example-com.id
}

resource "aws_security_group" "masters-additionalpolicies-example-com" {
  description = "Security group for masters"
  name        = "masters.additionalpolicies.example.com"
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "masters.additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
  vpc_id = aws_vpc.additionalpolicies-example-com.id
}

resource "aws_security_group" "nodes-additionalpolicies-example-com" {
  description = "Security group for nodes"
  name        = "nodes.additionalpolicies.example.com"
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "nodes.additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
  vpc_id = aws_vpc.additionalpolicies-example-com.id
}

resource "aws_subnet" "us-test-1a-additionalpolicies-example-com" {
  availability_zone = "us-test-1a"
  cidr_block        = "172.20.32.0/19"
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "us-test-1a.additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "SubnetType"                                           = "Public"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
    "kubernetes.io/role/elb"                               = "1"
  }
  vpc_id = aws_vpc.additionalpolicies-example-com.id
}

resource "aws_vpc_dhcp_options_association" "additionalpolicies-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.additionalpolicies-example-com.id
  vpc_id          = aws_vpc.additionalpolicies-example-com.id
}

resource "aws_vpc_dhcp_options" "additionalpolicies-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
  }
}

resource "aws_vpc" "additionalpolicies-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    "KubernetesCluster"                                    = "additionalpolicies.example.com"
    "Name"                                                 = "additionalpolicies.example.com"
    "Owner"                                                = "John Doe"
    "foo/bar"                                              = "fib+baz"
    "kubernetes.io/cluster/additionalpolicies.example.com" = "owned"
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
