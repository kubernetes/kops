locals {
  bastion_autoscaling_group_ids = [aws_autoscaling_group.bastion-unmanaged-example-com.id]
  bastion_security_group_ids    = [aws_security_group.bastion-unmanaged-example-com.id]
  bastions_role_arn             = aws_iam_role.bastions-unmanaged-example-com.arn
  bastions_role_name            = aws_iam_role.bastions-unmanaged-example-com.name
  cluster_name                  = "unmanaged.example.com"
  master_autoscaling_group_ids  = [aws_autoscaling_group.master-us-test-1a-masters-unmanaged-example-com.id]
  master_security_group_ids     = [aws_security_group.masters-unmanaged-example-com.id]
  masters_role_arn              = aws_iam_role.masters-unmanaged-example-com.arn
  masters_role_name             = aws_iam_role.masters-unmanaged-example-com.name
  node_autoscaling_group_ids    = [aws_autoscaling_group.nodes-unmanaged-example-com.id]
  node_security_group_ids       = [aws_security_group.nodes-unmanaged-example-com.id]
  node_subnet_ids               = [aws_subnet.us-test-1a-unmanaged-example-com.id, aws_subnet.us-test-1b-unmanaged-example-com.id]
  nodes_role_arn                = aws_iam_role.nodes-unmanaged-example-com.arn
  nodes_role_name               = aws_iam_role.nodes-unmanaged-example-com.name
  region                        = "us-test-1"
  subnet_us-test-1a_id          = aws_subnet.us-test-1a-unmanaged-example-com.id
  subnet_us-test-1b_id          = aws_subnet.us-test-1b-unmanaged-example-com.id
  subnet_utility-us-test-1a_id  = aws_subnet.utility-us-test-1a-unmanaged-example-com.id
  subnet_utility-us-test-1b_id  = aws_subnet.utility-us-test-1b-unmanaged-example-com.id
  vpc_id                        = "vpc-12345678"
}

output "bastion_autoscaling_group_ids" {
  value = [aws_autoscaling_group.bastion-unmanaged-example-com.id]
}

output "bastion_security_group_ids" {
  value = [aws_security_group.bastion-unmanaged-example-com.id]
}

output "bastions_role_arn" {
  value = aws_iam_role.bastions-unmanaged-example-com.arn
}

output "bastions_role_name" {
  value = aws_iam_role.bastions-unmanaged-example-com.name
}

output "cluster_name" {
  value = "unmanaged.example.com"
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-unmanaged-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-unmanaged-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-unmanaged-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-unmanaged-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-unmanaged-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-unmanaged-example-com.id]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-unmanaged-example-com.id, aws_subnet.us-test-1b-unmanaged-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-unmanaged-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-unmanaged-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-unmanaged-example-com.id
}

output "subnet_us-test-1b_id" {
  value = aws_subnet.us-test-1b-unmanaged-example-com.id
}

output "subnet_utility-us-test-1a_id" {
  value = aws_subnet.utility-us-test-1a-unmanaged-example-com.id
}

output "subnet_utility-us-test-1b_id" {
  value = aws_subnet.utility-us-test-1b-unmanaged-example-com.id
}

output "vpc_id" {
  value = "vpc-12345678"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_attachment" "bastion-unmanaged-example-com" {
  autoscaling_group_name = aws_autoscaling_group.bastion-unmanaged-example-com.id
  elb                    = aws_elb.bastion-unmanaged-example-com.id
}

resource "aws_autoscaling_attachment" "master-us-test-1a-masters-unmanaged-example-com" {
  autoscaling_group_name = aws_autoscaling_group.master-us-test-1a-masters-unmanaged-example-com.id
  elb                    = aws_elb.api-unmanaged-example-com.id
}

resource "aws_autoscaling_group" "bastion-unmanaged-example-com" {
  enabled_metrics      = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_configuration = aws_launch_configuration.bastion-unmanaged-example-com.id
  max_size             = 1
  metrics_granularity  = "1Minute"
  min_size             = 1
  name                 = "bastion.unmanaged.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "unmanaged.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "bastion.unmanaged.example.com"
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
    key                 = "kubernetes.io/cluster/unmanaged.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.utility-us-test-1a-unmanaged-example-com.id]
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-unmanaged-example-com" {
  enabled_metrics      = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_configuration = aws_launch_configuration.master-us-test-1a-masters-unmanaged-example-com.id
  max_size             = 1
  metrics_granularity  = "1Minute"
  min_size             = 1
  name                 = "master-us-test-1a.masters.unmanaged.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "unmanaged.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.unmanaged.example.com"
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
    key                 = "kubernetes.io/cluster/unmanaged.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-unmanaged-example-com.id]
}

resource "aws_autoscaling_group" "nodes-unmanaged-example-com" {
  enabled_metrics      = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_configuration = aws_launch_configuration.nodes-unmanaged-example-com.id
  max_size             = 2
  metrics_granularity  = "1Minute"
  min_size             = 2
  name                 = "nodes.unmanaged.example.com"
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "unmanaged.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.unmanaged.example.com"
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
    key                 = "kubernetes.io/cluster/unmanaged.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-unmanaged-example-com.id, aws_subnet.us-test-1b-unmanaged-example-com.id]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-unmanaged-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "us-test-1a.etcd-events.unmanaged.example.com"
    "k8s.io/etcd/events"                          = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                          = "1"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-unmanaged-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  size              = 20
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "us-test-1a.etcd-main.unmanaged.example.com"
    "k8s.io/etcd/main"                            = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                          = "1"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
  type = "gp2"
}

resource "aws_elb" "api-unmanaged-example-com" {
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
    instance_port      = 443
    instance_protocol  = "TCP"
    lb_port            = 443
    lb_protocol        = "TCP"
    ssl_certificate_id = ""
  }
  name            = "api-unmanaged-example-com-t82m6f"
  security_groups = [aws_security_group.api-elb-unmanaged-example-com.id]
  subnets         = [aws_subnet.utility-us-test-1a-unmanaged-example-com.id, aws_subnet.utility-us-test-1b-unmanaged-example-com.id]
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "api.unmanaged.example.com"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
}

resource "aws_elb" "bastion-unmanaged-example-com" {
  health_check {
    healthy_threshold   = 2
    interval            = 10
    target              = "TCP:22"
    timeout             = 5
    unhealthy_threshold = 2
  }
  idle_timeout = 300
  listener {
    instance_port      = 22
    instance_protocol  = "TCP"
    lb_port            = 22
    lb_protocol        = "TCP"
    ssl_certificate_id = ""
  }
  name            = "bastion-unmanaged-example-d7bn3d"
  security_groups = [aws_security_group.bastion-elb-unmanaged-example-com.id]
  subnets         = [aws_subnet.utility-us-test-1a-unmanaged-example-com.id]
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "bastion.unmanaged.example.com"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "bastions-unmanaged-example-com" {
  name = "bastions.unmanaged.example.com"
  role = aws_iam_role.bastions-unmanaged-example-com.name
}

resource "aws_iam_instance_profile" "masters-unmanaged-example-com" {
  name = "masters.unmanaged.example.com"
  role = aws_iam_role.masters-unmanaged-example-com.name
}

resource "aws_iam_instance_profile" "nodes-unmanaged-example-com" {
  name = "nodes.unmanaged.example.com"
  role = aws_iam_role.nodes-unmanaged-example-com.name
}

resource "aws_iam_role_policy" "bastions-unmanaged-example-com" {
  name   = "bastions.unmanaged.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_bastions.unmanaged.example.com_policy")
  role   = aws_iam_role.bastions-unmanaged-example-com.name
}

resource "aws_iam_role_policy" "masters-unmanaged-example-com" {
  name   = "masters.unmanaged.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.unmanaged.example.com_policy")
  role   = aws_iam_role.masters-unmanaged-example-com.name
}

resource "aws_iam_role_policy" "nodes-unmanaged-example-com" {
  name   = "nodes.unmanaged.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.unmanaged.example.com_policy")
  role   = aws_iam_role.nodes-unmanaged-example-com.name
}

resource "aws_iam_role" "bastions-unmanaged-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_bastions.unmanaged.example.com_policy")
  name               = "bastions.unmanaged.example.com"
}

resource "aws_iam_role" "masters-unmanaged-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.unmanaged.example.com_policy")
  name               = "masters.unmanaged.example.com"
}

resource "aws_iam_role" "nodes-unmanaged-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_nodes.unmanaged.example.com_policy")
  name               = "nodes.unmanaged.example.com"
}

resource "aws_key_pair" "kubernetes-unmanaged-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.unmanaged.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.unmanaged.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
}

resource "aws_launch_configuration" "bastion-unmanaged-example-com" {
  associate_public_ip_address = true
  enable_monitoring           = false
  iam_instance_profile        = aws_iam_instance_profile.bastions-unmanaged-example-com.id
  image_id                    = "ami-11400000"
  instance_type               = "t2.micro"
  key_name                    = aws_key_pair.kubernetes-unmanaged-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "bastion.unmanaged.example.com-"
  root_block_device {
    delete_on_termination = true
    volume_size           = 32
    volume_type           = "gp2"
  }
  security_groups = [aws_security_group.bastion-unmanaged-example-com.id]
}

resource "aws_launch_configuration" "master-us-test-1a-masters-unmanaged-example-com" {
  associate_public_ip_address = false
  enable_monitoring           = false
  ephemeral_block_device {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile = aws_iam_instance_profile.masters-unmanaged-example-com.id
  image_id             = "ami-11400000"
  instance_type        = "m3.medium"
  key_name             = aws_key_pair.kubernetes-unmanaged-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "master-us-test-1a.masters.unmanaged.example.com-"
  root_block_device {
    delete_on_termination = true
    volume_size           = 64
    volume_type           = "gp2"
  }
  security_groups = [aws_security_group.masters-unmanaged-example-com.id]
  user_data       = file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.unmanaged.example.com_user_data")
}

resource "aws_launch_configuration" "nodes-unmanaged-example-com" {
  associate_public_ip_address = false
  enable_monitoring           = false
  iam_instance_profile        = aws_iam_instance_profile.nodes-unmanaged-example-com.id
  image_id                    = "ami-11400000"
  instance_type               = "t2.medium"
  key_name                    = aws_key_pair.kubernetes-unmanaged-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  name_prefix = "nodes.unmanaged.example.com-"
  root_block_device {
    delete_on_termination = true
    volume_size           = 128
    volume_type           = "gp2"
  }
  security_groups = [aws_security_group.nodes-unmanaged-example-com.id]
  user_data       = file("${path.module}/data/aws_launch_configuration_nodes.unmanaged.example.com_user_data")
}

resource "aws_route53_record" "api-unmanaged-example-com" {
  alias {
    evaluate_target_health = false
    name                   = aws_elb.api-unmanaged-example-com.dns_name
    zone_id                = aws_elb.api-unmanaged-example-com.zone_id
  }
  name    = "api.unmanaged.example.com"
  type    = "A"
  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_security_group_rule" "all-master-to-master" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-unmanaged-example-com.id
  source_security_group_id = aws_security_group.masters-unmanaged-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "all-master-to-node" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-unmanaged-example-com.id
  source_security_group_id = aws_security_group.masters-unmanaged-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "all-node-to-node" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-unmanaged-example-com.id
  source_security_group_id = aws_security_group.nodes-unmanaged-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "api-elb-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.api-elb-unmanaged-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "bastion-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-unmanaged-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "bastion-elb-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.bastion-elb-unmanaged-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "bastion-to-master-ssh" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-unmanaged-example-com.id
  source_security_group_id = aws_security_group.bastion-unmanaged-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "bastion-to-node-ssh" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.nodes-unmanaged-example-com.id
  source_security_group_id = aws_security_group.bastion-unmanaged-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.api-elb-unmanaged-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "https-elb-to-master" {
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-unmanaged-example-com.id
  source_security_group_id = aws_security_group.api-elb-unmanaged-example-com.id
  to_port                  = 443
  type                     = "ingress"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 3
  protocol          = "icmp"
  security_group_id = aws_security_group.api-elb-unmanaged-example-com.id
  to_port           = 4
  type              = "ingress"
}

resource "aws_security_group_rule" "master-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-unmanaged-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "node-egress" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-unmanaged-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-unmanaged-example-com.id
  source_security_group_id = aws_security_group.nodes-unmanaged-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-tcp-2383-4000" {
  from_port                = 2383
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-unmanaged-example-com.id
  source_security_group_id = aws_security_group.nodes-unmanaged-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-unmanaged-example-com.id
  source_security_group_id = aws_security_group.nodes-unmanaged-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-unmanaged-example-com.id
  source_security_group_id = aws_security_group.nodes-unmanaged-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "ssh-elb-to-bastion" {
  from_port                = 22
  protocol                 = "tcp"
  security_group_id        = aws_security_group.bastion-unmanaged-example-com.id
  source_security_group_id = aws_security_group.bastion-elb-unmanaged-example-com.id
  to_port                  = 22
  type                     = "ingress"
}

resource "aws_security_group_rule" "ssh-external-to-bastion-elb-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.bastion-elb-unmanaged-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group" "api-elb-unmanaged-example-com" {
  description = "Security group for api ELB"
  name        = "api-elb.unmanaged.example.com"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "api-elb.unmanaged.example.com"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_security_group" "bastion-elb-unmanaged-example-com" {
  description = "Security group for bastion ELB"
  name        = "bastion-elb.unmanaged.example.com"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "bastion-elb.unmanaged.example.com"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_security_group" "bastion-unmanaged-example-com" {
  description = "Security group for bastion"
  name        = "bastion.unmanaged.example.com"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "bastion.unmanaged.example.com"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_security_group" "masters-unmanaged-example-com" {
  description = "Security group for masters"
  name        = "masters.unmanaged.example.com"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "masters.unmanaged.example.com"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_security_group" "nodes-unmanaged-example-com" {
  description = "Security group for nodes"
  name        = "nodes.unmanaged.example.com"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "nodes.unmanaged.example.com"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_subnet" "us-test-1a-unmanaged-example-com" {
  availability_zone = "us-test-1a"
  cidr_block        = "172.20.32.0/19"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "us-test-1a.unmanaged.example.com"
    "SubnetType"                                  = "Private"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
    "kubernetes.io/role/internal-elb"             = "1"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_subnet" "us-test-1b-unmanaged-example-com" {
  availability_zone = "us-test-1b"
  cidr_block        = "172.20.64.0/19"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "us-test-1b.unmanaged.example.com"
    "SubnetType"                                  = "Private"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
    "kubernetes.io/role/internal-elb"             = "1"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_subnet" "utility-us-test-1a-unmanaged-example-com" {
  availability_zone = "us-test-1a"
  cidr_block        = "172.20.4.0/22"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "utility-us-test-1a.unmanaged.example.com"
    "SubnetType"                                  = "Utility"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
    "kubernetes.io/role/elb"                      = "1"
  }
  vpc_id = "vpc-12345678"
}

resource "aws_subnet" "utility-us-test-1b-unmanaged-example-com" {
  availability_zone = "us-test-1b"
  cidr_block        = "172.20.8.0/22"
  tags = {
    "KubernetesCluster"                           = "unmanaged.example.com"
    "Name"                                        = "utility-us-test-1b.unmanaged.example.com"
    "SubnetType"                                  = "Utility"
    "kubernetes.io/cluster/unmanaged.example.com" = "owned"
    "kubernetes.io/role/elb"                      = "1"
  }
  vpc_id = "vpc-12345678"
}

terraform {
  required_version = ">= 0.12.0"
}
