locals = {
  bastions_role_arn  = "${aws_iam_role.bastions-lifecyclephases-example-com.arn}"
  bastions_role_name = "${aws_iam_role.bastions-lifecyclephases-example-com.name}"
  cluster_name       = "lifecyclephases.example.com"
  masters_role_arn   = "${aws_iam_role.masters-lifecyclephases-example-com.arn}"
  masters_role_name  = "${aws_iam_role.masters-lifecyclephases-example-com.name}"
  nodes_role_arn     = "${aws_iam_role.nodes-lifecyclephases-example-com.arn}"
  nodes_role_name    = "${aws_iam_role.nodes-lifecyclephases-example-com.name}"
  region             = "us-test-1"
}

output "bastions_role_arn" {
  value = "${aws_iam_role.bastions-lifecyclephases-example-com.arn}"
}

output "bastions_role_name" {
  value = "${aws_iam_role.bastions-lifecyclephases-example-com.name}"
}

output "cluster_name" {
  value = "lifecyclephases.example.com"
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-lifecyclephases-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-lifecyclephases-example-com.name}"
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-lifecyclephases-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-lifecyclephases-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_iam_instance_profile" "bastions-lifecyclephases-example-com" {
  name = "bastions.lifecyclephases.example.com"
  role = "${aws_iam_role.bastions-lifecyclephases-example-com.name}"
}

resource "aws_iam_instance_profile" "masters-lifecyclephases-example-com" {
  name = "masters.lifecyclephases.example.com"
  role = "${aws_iam_role.masters-lifecyclephases-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-lifecyclephases-example-com" {
  name = "nodes.lifecyclephases.example.com"
  role = "${aws_iam_role.nodes-lifecyclephases-example-com.name}"
}

resource "aws_iam_role" "bastions-lifecyclephases-example-com" {
  name               = "bastions.lifecyclephases.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_bastions.lifecyclephases.example.com_policy")}"
}

resource "aws_iam_role" "masters-lifecyclephases-example-com" {
  name               = "masters.lifecyclephases.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.lifecyclephases.example.com_policy")}"
}

resource "aws_iam_role" "nodes-lifecyclephases-example-com" {
  name               = "nodes.lifecyclephases.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.lifecyclephases.example.com_policy")}"
}

resource "aws_iam_role_policy" "bastions-lifecyclephases-example-com" {
  name   = "bastions.lifecyclephases.example.com"
  role   = "${aws_iam_role.bastions-lifecyclephases-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_bastions.lifecyclephases.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-lifecyclephases-example-com" {
  name   = "masters.lifecyclephases.example.com"
  role   = "${aws_iam_role.masters-lifecyclephases-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.lifecyclephases.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-lifecyclephases-example-com" {
  name   = "nodes.lifecyclephases.example.com"
  role   = "${aws_iam_role.nodes-lifecyclephases-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.lifecyclephases.example.com_policy")}"
}

resource "aws_key_pair" "kubernetes-lifecyclephases-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.lifecyclephases.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.lifecyclephases.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_security_group" "api-elb-lifecyclephases-example-com" {
  name        = "api-elb.lifecyclephases.example.com"
  vpc_id      = "${aws_vpc.lifecyclephases-example-com.id}"
  description = "Security group for api ELB"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "api-elb.lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_security_group" "bastion-elb-lifecyclephases-example-com" {
  name        = "bastion-elb.lifecyclephases.example.com"
  vpc_id      = "${aws_vpc.lifecyclephases-example-com.id}"
  description = "Security group for bastion ELB"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "bastion-elb.lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_security_group" "bastion-lifecyclephases-example-com" {
  name        = "bastion.lifecyclephases.example.com"
  vpc_id      = "${aws_vpc.lifecyclephases-example-com.id}"
  description = "Security group for bastion"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "bastion.lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_security_group" "masters-lifecyclephases-example-com" {
  name        = "masters.lifecyclephases.example.com"
  vpc_id      = "${aws_vpc.lifecyclephases-example-com.id}"
  description = "Security group for masters"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "masters.lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_security_group" "nodes-lifecyclephases-example-com" {
  name        = "nodes.lifecyclephases.example.com"
  vpc_id      = "${aws_vpc.lifecyclephases-example-com.id}"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "nodes.lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "api-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.api-elb-lifecyclephases-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-lifecyclephases-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-elb-lifecyclephases-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-to-master-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-lifecyclephases-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "bastion-to-node-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-lifecyclephases-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.api-elb-lifecyclephases-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "https-elb-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.api-elb-lifecyclephases-example-com.id}"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  from_port                = 1
  to_port                  = 2379
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-2382-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  from_port                = 2382
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-lifecyclephases-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-elb-to-bastion" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.bastion-lifecyclephases-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-elb-lifecyclephases-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "ssh-external-to-bastion-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.bastion-elb-lifecyclephases-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

terraform = {
  required_version = ">= 0.9.3"
}
