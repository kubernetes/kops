locals = {
  bastion_autoscaling_group_ids = ["${aws_autoscaling_group.bastion-private-shared-subnet-example-com.id}"]
  bastion_security_group_ids    = ["${aws_security_group.bastion-private-shared-subnet-example-com.id}"]
  bastions_role_arn             = "${aws_iam_role.bastions-private-shared-subnet-example-com.arn}"
  bastions_role_name            = "${aws_iam_role.bastions-private-shared-subnet-example-com.name}"
  cluster_name                  = "private-shared-subnet.example.com"
  master_autoscaling_group_ids  = ["${aws_autoscaling_group.master-us-test-1a-masters-private-shared-subnet-example-com.id}"]
  master_security_group_ids     = ["${aws_security_group.masters-private-shared-subnet-example-com.id}"]
  masters_role_arn              = "${aws_iam_role.masters-private-shared-subnet-example-com.arn}"
  masters_role_name             = "${aws_iam_role.masters-private-shared-subnet-example-com.name}"
  node_autoscaling_group_ids    = ["${aws_autoscaling_group.nodes-private-shared-subnet-example-com.id}"]
  node_security_group_ids       = ["${aws_security_group.nodes-private-shared-subnet-example-com.id}"]
  node_subnet_ids               = ["subnet-12345678"]
  nodes_role_arn                = "${aws_iam_role.nodes-private-shared-subnet-example-com.arn}"
  nodes_role_name               = "${aws_iam_role.nodes-private-shared-subnet-example-com.name}"
  region                        = "us-test-1"
  subnet_ids                    = ["subnet-12345678", "subnet-abcdef"]
  subnet_us-test-1a_id          = "subnet-12345678"
  subnet_utility-us-test-1a_id  = "subnet-abcdef"
  vpc_id                        = "vpc-12345678"
}

output "bastion_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.bastion-private-shared-subnet-example-com.id}"]
}

output "bastion_security_group_ids" {
  value = ["${aws_security_group.bastion-private-shared-subnet-example-com.id}"]
}

output "bastions_role_arn" {
  value = "${aws_iam_role.bastions-private-shared-subnet-example-com.arn}"
}

output "bastions_role_name" {
  value = "${aws_iam_role.bastions-private-shared-subnet-example-com.name}"
}

output "cluster_name" {
  value = "private-shared-subnet.example.com"
}

output "master_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.master-us-test-1a-masters-private-shared-subnet-example-com.id}"]
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-private-shared-subnet-example-com.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-private-shared-subnet-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-private-shared-subnet-example-com.name}"
}

output "node_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.nodes-private-shared-subnet-example-com.id}"]
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-private-shared-subnet-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["subnet-12345678"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-private-shared-subnet-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-private-shared-subnet-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

output "subnet_ids" {
  value = ["subnet-12345678", "subnet-abcdef"]
}

output "subnet_us-test-1a_id" {
  value = "subnet-12345678"
}

output "subnet_utility-us-test-1a_id" {
  value = "subnet-abcdef"
}

output "vpc_id" {
  value = "vpc-12345678"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_attachment" "bastion-private-shared-subnet-example-com" {
  elb                    = "${aws_elb.bastion-private-shared-subnet-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.bastion-private-shared-subnet-example-com.id}"
}

resource "aws_autoscaling_attachment" "master-us-test-1a-masters-private-shared-subnet-example-com" {
  elb                    = "${aws_elb.api-private-shared-subnet-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.master-us-test-1a-masters-private-shared-subnet-example-com.id}"
}

resource "aws_autoscaling_group" "bastion-private-shared-subnet-example-com" {
  name                 = "bastion.private-shared-subnet.example.com"
  launch_configuration = "${aws_launch_configuration.bastion-private-shared-subnet-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["subnet-abcdef"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "private-shared-subnet.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "bastion.private-shared-subnet.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/bastion"
    value               = "1"
    propagate_at_launch = true
  }

  tag = {
    key                 = "kops.k8s.io/instancegroup"
    value               = "bastion"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-private-shared-subnet-example-com" {
  name                 = "master-us-test-1a.masters.private-shared-subnet.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-private-shared-subnet-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["subnet-12345678"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "private-shared-subnet.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.private-shared-subnet.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  tag = {
    key                 = "kops.k8s.io/instancegroup"
    value               = "master-us-test-1a"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "nodes-private-shared-subnet-example-com" {
  name                 = "nodes.private-shared-subnet.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-private-shared-subnet-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["subnet-12345678"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "private-shared-subnet.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.private-shared-subnet.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }

  tag = {
    key                 = "kops.k8s.io/instancegroup"
    value               = "nodes"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-private-shared-subnet-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "us-test-1a.etcd-events.private-shared-subnet.example.com"
    "k8s.io/etcd/events"                                      = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                                      = "1"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-private-shared-subnet-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "us-test-1a.etcd-main.private-shared-subnet.example.com"
    "k8s.io/etcd/main"                                        = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                                      = "1"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_elb" "api-private-shared-subnet-example-com" {
  name = "api-private-shared-subnet-n2f8ak"

  listener = {
    instance_port     = 443
    instance_protocol = "TCP"
    lb_port           = 443
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.api-elb-private-shared-subnet-example-com.id}"]
  subnets         = ["subnet-abcdef"]

  health_check = {
    target              = "SSL:443"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  cross_zone_load_balancing = false
  idle_timeout              = 300

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "api.private-shared-subnet.example.com"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_elb" "bastion-private-shared-subnet-example-com" {
  name = "bastion-private-shared-su-5ol32q"

  listener = {
    instance_port     = 22
    instance_protocol = "TCP"
    lb_port           = 22
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.bastion-elb-private-shared-subnet-example-com.id}"]
  subnets         = ["subnet-abcdef"]

  health_check = {
    target              = "TCP:22"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "bastion.private-shared-subnet.example.com"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "bastions-private-shared-subnet-example-com" {
  name = "bastions.private-shared-subnet.example.com"
  role = "${aws_iam_role.bastions-private-shared-subnet-example-com.name}"
}

resource "aws_iam_instance_profile" "masters-private-shared-subnet-example-com" {
  name = "masters.private-shared-subnet.example.com"
  role = "${aws_iam_role.masters-private-shared-subnet-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-private-shared-subnet-example-com" {
  name = "nodes.private-shared-subnet.example.com"
  role = "${aws_iam_role.nodes-private-shared-subnet-example-com.name}"
}

resource "aws_iam_role" "bastions-private-shared-subnet-example-com" {
  name               = "bastions.private-shared-subnet.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_bastions.private-shared-subnet.example.com_policy")}"
}

resource "aws_iam_role" "masters-private-shared-subnet-example-com" {
  name               = "masters.private-shared-subnet.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.private-shared-subnet.example.com_policy")}"
}

resource "aws_iam_role" "nodes-private-shared-subnet-example-com" {
  name               = "nodes.private-shared-subnet.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.private-shared-subnet.example.com_policy")}"
}

resource "aws_iam_role_policy" "bastions-private-shared-subnet-example-com" {
  name   = "bastions.private-shared-subnet.example.com"
  role   = "${aws_iam_role.bastions-private-shared-subnet-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_bastions.private-shared-subnet.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-private-shared-subnet-example-com" {
  name   = "masters.private-shared-subnet.example.com"
  role   = "${aws_iam_role.masters-private-shared-subnet-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.private-shared-subnet.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-private-shared-subnet-example-com" {
  name   = "nodes.private-shared-subnet.example.com"
  role   = "${aws_iam_role.nodes-private-shared-subnet-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.private-shared-subnet.example.com_policy")}"
}

resource "aws_key_pair" "kubernetes-private-shared-subnet-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.private-shared-subnet.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.private-shared-subnet.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "bastion-private-shared-subnet-example-com" {
  name_prefix                 = "bastion.private-shared-subnet.example.com-"
  image_id                    = "ami-11400000"
  instance_type               = "t2.micro"
  key_name                    = "${aws_key_pair.kubernetes-private-shared-subnet-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.bastions-private-shared-subnet-example-com.id}"
  security_groups             = ["${aws_security_group.bastion-private-shared-subnet-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_bastion.private-shared-subnet.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 32
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }

  enable_monitoring = false
}

resource "aws_launch_configuration" "master-us-test-1a-masters-private-shared-subnet-example-com" {
  name_prefix                 = "master-us-test-1a.masters.private-shared-subnet.example.com-"
  image_id                    = "ami-11400000"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-private-shared-subnet-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-private-shared-subnet-example-com.id}"
  security_groups             = ["${aws_security_group.masters-private-shared-subnet-example-com.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.private-shared-subnet.example.com_user_data")}"

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

  enable_monitoring = false
}

resource "aws_launch_configuration" "nodes-private-shared-subnet-example-com" {
  name_prefix                 = "nodes.private-shared-subnet.example.com-"
  image_id                    = "ami-11400000"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-private-shared-subnet-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-private-shared-subnet-example-com.id}"
  security_groups             = ["${aws_security_group.nodes-private-shared-subnet-example-com.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.private-shared-subnet.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 128
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }

  enable_monitoring = false
}

resource "aws_route53_record" "api-private-shared-subnet-example-com" {
  name = "api.private-shared-subnet.example.com"
  type = "A"

  alias = {
    name                   = "${aws_elb.api-private-shared-subnet-example-com.dns_name}"
    zone_id                = "${aws_elb.api-private-shared-subnet-example-com.zone_id}"
    evaluate_target_health = false
  }

  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_security_group" "api-elb-private-shared-subnet-example-com" {
  name        = "api-elb.private-shared-subnet.example.com"
  vpc_id      = "vpc-12345678"
  description = "Security group for api ELB"

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "api-elb.private-shared-subnet.example.com"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_security_group" "bastion-elb-private-shared-subnet-example-com" {
  name        = "bastion-elb.private-shared-subnet.example.com"
  vpc_id      = "vpc-12345678"
  description = "Security group for bastion ELB"

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "bastion-elb.private-shared-subnet.example.com"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_security_group" "bastion-private-shared-subnet-example-com" {
  name        = "bastion.private-shared-subnet.example.com"
  vpc_id      = "vpc-12345678"
  description = "Security group for bastion"

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "bastion.private-shared-subnet.example.com"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_security_group" "masters-private-shared-subnet-example-com" {
  name        = "masters.private-shared-subnet.example.com"
  vpc_id      = "vpc-12345678"
  description = "Security group for masters"

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "masters.private-shared-subnet.example.com"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_security_group" "nodes-private-shared-subnet-example-com" {
  name        = "nodes.private-shared-subnet.example.com"
  vpc_id      = "vpc-12345678"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster                                         = "private-shared-subnet.example.com"
    Name                                                      = "nodes.private-shared-subnet.example.com"
    "kubernetes.io/cluster/private-shared-subnet.example.com" = "owned"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "api-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.api-elb-private-shared-subnet-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-private-shared-subnet-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-elb-private-shared-subnet-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-to-master-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-private-shared-subnet-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "bastion-to-node-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-private-shared-subnet-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.api-elb-private-shared-subnet-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "https-elb-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.api-elb-private-shared-subnet-example-com.id}"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "icmp-pmtu-api-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.api-elb-private-shared-subnet-example-com.id}"
  from_port         = 3
  to_port           = 4
  protocol          = "icmp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  from_port                = 1
  to_port                  = 2379
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-2382-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  from_port                = 2382
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-private-shared-subnet-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-elb-to-bastion" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.bastion-private-shared-subnet-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-elb-private-shared-subnet-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "ssh-external-to-bastion-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.bastion-elb-private-shared-subnet-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

terraform = {
  required_version = ">= 0.9.3"
}
