locals = {
  cluster_name                 = "sharedvpc.example.com"
  master_autoscaling_group_ids = ["${aws_autoscaling_group.master-us-test-1a-masters-sharedvpc-example-com.id}"]
  master_security_group_ids    = ["${aws_security_group.masters-sharedvpc-example-com.id}"]
  masters_role_arn             = "${aws_iam_role.masters-sharedvpc-example-com.arn}"
  masters_role_name            = "${aws_iam_role.masters-sharedvpc-example-com.name}"
  node_autoscaling_group_ids   = ["${aws_autoscaling_group.nodes-sharedvpc-example-com.id}"]
  node_security_group_ids      = ["${aws_security_group.nodes-sharedvpc-example-com.id}"]
  node_subnet_ids              = ["${aws_subnet.us-test-1a-sharedvpc-example-com.id}"]
  nodes_role_arn               = "${aws_iam_role.nodes-sharedvpc-example-com.arn}"
  nodes_role_name              = "${aws_iam_role.nodes-sharedvpc-example-com.name}"
  region                       = "us-test-1"
  route_table_public_id        = "${aws_route_table.sharedvpc-example-com.id}"
  subnet_us-test-1a_id         = "${aws_subnet.us-test-1a-sharedvpc-example-com.id}"
  vpc_id                       = "vpc-12345678"
}

output "cluster_name" {
  value = "sharedvpc.example.com"
}

output "master_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.master-us-test-1a-masters-sharedvpc-example-com.id}"]
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-sharedvpc-example-com.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-sharedvpc-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-sharedvpc-example-com.name}"
}

output "node_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.nodes-sharedvpc-example-com.id}"]
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-sharedvpc-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-sharedvpc-example-com.id}"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-sharedvpc-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-sharedvpc-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

output "route_table_public_id" {
  value = "${aws_route_table.sharedvpc-example-com.id}"
}

output "subnet_us-test-1a_id" {
  value = "${aws_subnet.us-test-1a-sharedvpc-example-com.id}"
}

output "vpc_id" {
  value = "vpc-12345678"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-sharedvpc-example-com" {
  name                 = "master-us-test-1a.masters.sharedvpc.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-sharedvpc-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-sharedvpc-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "sharedvpc.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.sharedvpc.example.com"
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

resource "aws_autoscaling_group" "nodes-sharedvpc-example-com" {
  name                 = "nodes.sharedvpc.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-sharedvpc-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-sharedvpc-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "sharedvpc.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.sharedvpc.example.com"
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

resource "aws_ebs_volume" "us-test-1a-etcd-events-sharedvpc-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster                             = "sharedvpc.example.com"
    Name                                          = "us-test-1a.etcd-events.sharedvpc.example.com"
    "k8s.io/etcd/events"                          = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                          = "1"
    "kubernetes.io/cluster/sharedvpc.example.com" = "owned"
  }
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-sharedvpc-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster                             = "sharedvpc.example.com"
    Name                                          = "us-test-1a.etcd-main.sharedvpc.example.com"
    "k8s.io/etcd/main"                            = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                          = "1"
    "kubernetes.io/cluster/sharedvpc.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "masters-sharedvpc-example-com" {
  name = "masters.sharedvpc.example.com"
  role = "${aws_iam_role.masters-sharedvpc-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-sharedvpc-example-com" {
  name = "nodes.sharedvpc.example.com"
  role = "${aws_iam_role.nodes-sharedvpc-example-com.name}"
}

resource "aws_iam_role" "masters-sharedvpc-example-com" {
  name               = "masters.sharedvpc.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.sharedvpc.example.com_policy")}"
}

resource "aws_iam_role" "nodes-sharedvpc-example-com" {
  name               = "nodes.sharedvpc.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.sharedvpc.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-sharedvpc-example-com" {
  name   = "masters.sharedvpc.example.com"
  role   = "${aws_iam_role.masters-sharedvpc-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.sharedvpc.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-sharedvpc-example-com" {
  name   = "nodes.sharedvpc.example.com"
  role   = "${aws_iam_role.nodes-sharedvpc-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.sharedvpc.example.com_policy")}"
}

resource "aws_key_pair" "kubernetes-sharedvpc-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.sharedvpc.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.sharedvpc.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "master-us-test-1a-masters-sharedvpc-example-com" {
  name_prefix                 = "master-us-test-1a.masters.sharedvpc.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-sharedvpc-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-sharedvpc-example-com.id}"
  security_groups             = ["${aws_security_group.masters-sharedvpc-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.sharedvpc.example.com_user_data")}"

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

resource "aws_launch_configuration" "nodes-sharedvpc-example-com" {
  name_prefix                 = "nodes.sharedvpc.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-sharedvpc-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-sharedvpc-example-com.id}"
  security_groups             = ["${aws_security_group.nodes-sharedvpc-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.sharedvpc.example.com_user_data")}"

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

resource "aws_route" "0-0-0-0--0" {
  route_table_id         = "${aws_route_table.sharedvpc-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "igw-1"
}

resource "aws_route_table" "sharedvpc-example-com" {
  vpc_id = "vpc-12345678"

  tags = {
    KubernetesCluster                             = "sharedvpc.example.com"
    Name                                          = "sharedvpc.example.com"
    "kubernetes.io/cluster/sharedvpc.example.com" = "owned"
    "kubernetes.io/kops/role"                     = "public"
  }
}

resource "aws_route_table_association" "us-test-1a-sharedvpc-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-sharedvpc-example-com.id}"
  route_table_id = "${aws_route_table.sharedvpc-example-com.id}"
}

resource "aws_security_group" "masters-sharedvpc-example-com" {
  name        = "masters.sharedvpc.example.com"
  vpc_id      = "vpc-12345678"
  description = "Security group for masters"

  tags = {
    KubernetesCluster                             = "sharedvpc.example.com"
    Name                                          = "masters.sharedvpc.example.com"
    "kubernetes.io/cluster/sharedvpc.example.com" = "owned"
  }
}

resource "aws_security_group" "nodes-sharedvpc-example-com" {
  name        = "nodes.sharedvpc.example.com"
  vpc_id      = "vpc-12345678"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster                             = "sharedvpc.example.com"
    Name                                          = "nodes.sharedvpc.example.com"
    "kubernetes.io/cluster/sharedvpc.example.com" = "owned"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-sharedvpc-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-sharedvpc-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-sharedvpc-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "https-external-to-master-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-sharedvpc-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-sharedvpc-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-sharedvpc-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  from_port                = 1
  to_port                  = 2379
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-2382-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-sharedvpc-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  from_port                = 2382
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-sharedvpc-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-sharedvpc-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-external-to-master-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-sharedvpc-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "ssh-external-to-node-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.nodes-sharedvpc-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_subnet" "us-test-1a-sharedvpc-example-com" {
  vpc_id            = "vpc-12345678"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                             = "sharedvpc.example.com"
    Name                                          = "us-test-1a.sharedvpc.example.com"
    SubnetType                                    = "Public"
    "kubernetes.io/cluster/sharedvpc.example.com" = "owned"
    "kubernetes.io/role/elb"                      = "1"
  }
}

terraform = {
  required_version = ">= 0.9.3"
}
