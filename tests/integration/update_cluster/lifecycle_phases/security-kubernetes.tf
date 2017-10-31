output "bastions_role_arn" {
  value = "${aws_iam_role.bastions-privateweave-example-com.arn}"
}

output "bastions_role_name" {
  value = "${aws_iam_role.bastions-privateweave-example-com.name}"
}

output "cluster_name" {
  value = "privateweave.example.com"
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-privateweave-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-privateweave-example-com.name}"
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-privateweave-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-privateweave-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_iam_instance_profile" "bastions-privateweave-example-com" {
  name = "bastions.privateweave.example.com"
  role = "${aws_iam_role.bastions-privateweave-example-com.name}"
}

resource "aws_iam_instance_profile" "masters-privateweave-example-com" {
  name = "masters.privateweave.example.com"
  role = "${aws_iam_role.masters-privateweave-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-privateweave-example-com" {
  name = "nodes.privateweave.example.com"
  role = "${aws_iam_role.nodes-privateweave-example-com.name}"
}

resource "aws_iam_role" "bastions-privateweave-example-com" {
  name               = "bastions.privateweave.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_bastions.privateweave.example.com_policy")}"
}

resource "aws_iam_role" "masters-privateweave-example-com" {
  name               = "masters.privateweave.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.privateweave.example.com_policy")}"
}

resource "aws_iam_role" "nodes-privateweave-example-com" {
  name               = "nodes.privateweave.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.privateweave.example.com_policy")}"
}

resource "aws_iam_role_policy" "bastions-privateweave-example-com" {
  name   = "bastions.privateweave.example.com"
  role   = "${aws_iam_role.bastions-privateweave-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_bastions.privateweave.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-privateweave-example-com" {
  name   = "masters.privateweave.example.com"
  role   = "${aws_iam_role.masters-privateweave-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.privateweave.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-privateweave-example-com" {
  name   = "nodes.privateweave.example.com"
  role   = "${aws_iam_role.nodes-privateweave-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.privateweave.example.com_policy")}"
}

resource "aws_key_pair" "kubernetes-privateweave-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.privateweave.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.privateweave.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_security_group" "api-elb-privateweave-example-com" {
  name        = "api-elb.privateweave.example.com"
  vpc_id      = "${aws_vpc.privateweave-example-com.id}"
  description = "Security group for api ELB"

  tags = {
    KubernetesCluster = "privateweave.example.com"
    Name              = "api-elb.privateweave.example.com"
  }
}

resource "aws_security_group" "bastion-elb-privateweave-example-com" {
  name        = "bastion-elb.privateweave.example.com"
  vpc_id      = "${aws_vpc.privateweave-example-com.id}"
  description = "Security group for bastion ELB"

  tags = {
    KubernetesCluster = "privateweave.example.com"
    Name              = "bastion-elb.privateweave.example.com"
  }
}

resource "aws_security_group" "bastion-privateweave-example-com" {
  name        = "bastion.privateweave.example.com"
  vpc_id      = "${aws_vpc.privateweave-example-com.id}"
  description = "Security group for bastion"

  tags = {
    KubernetesCluster = "privateweave.example.com"
    Name              = "bastion.privateweave.example.com"
  }
}

resource "aws_security_group" "masters-privateweave-example-com" {
  name        = "masters.privateweave.example.com"
  vpc_id      = "${aws_vpc.privateweave-example-com.id}"
  description = "Security group for masters"

  tags = {
    KubernetesCluster = "privateweave.example.com"
    Name              = "masters.privateweave.example.com"
  }
}

resource "aws_security_group" "nodes-privateweave-example-com" {
  name        = "nodes.privateweave.example.com"
  vpc_id      = "${aws_vpc.privateweave-example-com.id}"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster = "privateweave.example.com"
    Name              = "nodes.privateweave.example.com"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-privateweave-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-privateweave-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-privateweave-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "api-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.api-elb-privateweave-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-privateweave-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-elb-privateweave-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-to-master-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-privateweave-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "bastion-to-node-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-privateweave-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.api-elb-privateweave-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "https-elb-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.api-elb-privateweave-example-com.id}"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-privateweave-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-privateweave-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-privateweave-example-com.id}"
  from_port                = 1
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-privateweave-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-privateweave-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-elb-to-bastion" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.bastion-privateweave-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-elb-privateweave-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "ssh-external-to-bastion-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.bastion-elb-privateweave-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

terraform = {
  required_version = ">= 0.9.3"
}
