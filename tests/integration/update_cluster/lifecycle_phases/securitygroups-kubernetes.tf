output "bastion_security_group_ids" {
  value = ["${aws_security_group.bastion-privateweave-example-com.id}"]
}

output "cluster_name" {
  value = "privateweave.example.com"
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-privateweave-example-com.id}"]
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-privateweave-example-com.id}"]
}

output "region" {
  value = "us-test-1"
}

provider "aws" {
  region = "us-test-1"
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
