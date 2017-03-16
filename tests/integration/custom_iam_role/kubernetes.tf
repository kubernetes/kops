output "cluster_name" {
  value = "custom-iam-role.example.com"
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-custom-iam-role-example-com.id}"]
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-custom-iam-role-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-custom-iam-role-example-com.id}"]
}

output "region" {
  value = "us-test-1"
}

output "vpc_id" {
  value = "${aws_vpc.custom-iam-role-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-custom-iam-role-example-com" {
  name                 = "master-us-test-1a.masters.custom-iam-role.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-custom-iam-role-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-custom-iam-role-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "custom-iam-role.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.custom-iam-role.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "nodes-custom-iam-role-example-com" {
  name                 = "nodes.custom-iam-role.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-custom-iam-role-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-custom-iam-role-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "custom-iam-role.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.custom-iam-role.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_ebs_volume" "a-etcd-events-custom-iam-role-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "custom-iam-role.example.com"
    Name                 = "a.etcd-events.custom-iam-role.example.com"
    "k8s.io/etcd/events" = "a/a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "a-etcd-main-custom-iam-role-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "custom-iam-role.example.com"
    Name                 = "a.etcd-main.custom-iam-role.example.com"
    "k8s.io/etcd/main"   = "a/a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_internet_gateway" "custom-iam-role-example-com" {
  vpc_id = "${aws_vpc.custom-iam-role-example-com.id}"

  tags = {
    KubernetesCluster = "custom-iam-role.example.com"
    Name              = "custom-iam-role.example.com"
  }
}

resource "aws_key_pair" "kubernetes-custom-iam-role-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.custom-iam-role.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.custom-iam-role.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "master-us-test-1a-masters-custom-iam-role-example-com" {
  name_prefix                 = "master-us-test-1a.masters.custom-iam-role.example.com-"
  image_id                    = "ami-15000000"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-custom-iam-role-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "arn:aws:iam::4222917490108:instance-profile/kops-custom-master-role"
  security_groups             = ["${aws_security_group.masters-custom-iam-role-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.custom-iam-role.example.com_user_data")}"

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
}

resource "aws_launch_configuration" "nodes-custom-iam-role-example-com" {
  name_prefix                 = "nodes.custom-iam-role.example.com-"
  image_id                    = "ami-15000000"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-custom-iam-role-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "arn:aws:iam::422917490108:instance-profile/kops-custom-node-role"
  security_groups             = ["${aws_security_group.nodes-custom-iam-role-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.custom-iam-role.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 128
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }
}

resource "aws_route" "0-0-0-0--0" {
  route_table_id         = "${aws_route_table.custom-iam-role-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.custom-iam-role-example-com.id}"
}

resource "aws_route_table" "custom-iam-role-example-com" {
  vpc_id = "${aws_vpc.custom-iam-role-example-com.id}"

  tags = {
    KubernetesCluster = "custom-iam-role.example.com"
    Name              = "custom-iam-role.example.com"
  }
}

resource "aws_route_table_association" "us-test-1a-custom-iam-role-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-custom-iam-role-example-com.id}"
  route_table_id = "${aws_route_table.custom-iam-role-example-com.id}"
}

resource "aws_security_group" "masters-custom-iam-role-example-com" {
  name        = "masters.custom-iam-role.example.com"
  vpc_id      = "${aws_vpc.custom-iam-role-example-com.id}"
  description = "Security group for masters"

  tags = {
    KubernetesCluster = "custom-iam-role.example.com"
    Name              = "masters.custom-iam-role.example.com"
  }
}

resource "aws_security_group" "nodes-custom-iam-role-example-com" {
  name        = "nodes.custom-iam-role.example.com"
  vpc_id      = "${aws_vpc.custom-iam-role-example-com.id}"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster = "custom-iam-role.example.com"
    Name              = "nodes.custom-iam-role.example.com"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-custom-iam-role-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-custom-iam-role-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-custom-iam-role-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "https-external-to-master-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-custom-iam-role-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-custom-iam-role-example-com.id}"
  from_port                = 1
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-custom-iam-role-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-custom-iam-role-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-external-to-master-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-custom-iam-role-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "ssh-external-to-node-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.nodes-custom-iam-role-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_subnet" "us-test-1a-custom-iam-role-example-com" {
  vpc_id            = "${aws_vpc.custom-iam-role-example-com.id}"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                   = "custom-iam-role.example.com"
    Name                                                = "us-test-1a.custom-iam-role.example.com"
    "kubernetes.io/cluster/custom-iam-role.example.com" = "owned"
  }
}

resource "aws_vpc" "custom-iam-role-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    KubernetesCluster                                   = "custom-iam-role.example.com"
    Name                                                = "custom-iam-role.example.com"
    "kubernetes.io/cluster/custom-iam-role.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "custom-iam-role-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    KubernetesCluster = "custom-iam-role.example.com"
    Name              = "custom-iam-role.example.com"
  }
}

resource "aws_vpc_dhcp_options_association" "custom-iam-role-example-com" {
  vpc_id          = "${aws_vpc.custom-iam-role-example-com.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.custom-iam-role-example-com.id}"
}

terraform = {
  required_version = ">= 0.9.3"
}
