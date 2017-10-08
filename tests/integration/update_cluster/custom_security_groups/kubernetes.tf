output "bastion_security_group_ids" {
  value = ["${aws_security_group.bastion-custom-security-groups-example-com.id}"]
}

output "bastions_role_arn" {
  value = "${aws_iam_role.bastions-custom-security-groups-example-com.arn}"
}

output "bastions_role_name" {
  value = "${aws_iam_role.bastions-custom-security-groups-example-com.name}"
}

output "cluster_name" {
  value = "custom-security-groups.example.com"
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-custom-security-groups-example-com.id}", "sg-exampleid3", "sg-exampleid4"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-custom-security-groups-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-custom-security-groups-example-com.name}"
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-custom-security-groups-example-com.id}", "sg-exampleid", "sg-exampleid2"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-custom-security-groups-example-com.id}"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-custom-security-groups-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-custom-security-groups-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

output "vpc_id" {
  value = "${aws_vpc.custom-security-groups-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_attachment" "bastion-custom-security-groups-example-com" {
  elb                    = "${aws_elb.bastion-custom-security-groups-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.bastion-custom-security-groups-example-com.id}"
}

resource "aws_autoscaling_attachment" "master-us-test-1a-masters-custom-security-groups-example-com" {
  elb                    = "${aws_elb.api-custom-security-groups-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.master-us-test-1a-masters-custom-security-groups-example-com.id}"
}

resource "aws_autoscaling_group" "bastion-custom-security-groups-example-com" {
  name                 = "bastion.custom-security-groups.example.com"
  launch_configuration = "${aws_launch_configuration.bastion-custom-security-groups-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.utility-us-test-1a-custom-security-groups-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "custom-security-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "bastion.custom-security-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/bastion"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-custom-security-groups-example-com" {
  name                 = "master-us-test-1a.masters.custom-security-groups.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-custom-security-groups-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-custom-security-groups-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "custom-security-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.custom-security-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "nodes-custom-security-groups-example-com" {
  name                 = "nodes.custom-security-groups.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-custom-security-groups-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-custom-security-groups-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "custom-security-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.custom-security-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_ebs_volume" "a-etcd-events-custom-security-groups-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "custom-security-groups.example.com"
    Name                 = "a.etcd-events.custom-security-groups.example.com"
    "k8s.io/etcd/events" = "a/a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "a-etcd-main-custom-security-groups-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "custom-security-groups.example.com"
    Name                 = "a.etcd-main.custom-security-groups.example.com"
    "k8s.io/etcd/main"   = "a/a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_eip" "us-test-1a-custom-security-groups-example-com" {
  vpc = true
}

resource "aws_elb" "api-custom-security-groups-example-com" {
  name = "api-custom-security-group-ueqqth"

  listener = {
    instance_port     = 443
    instance_protocol = "TCP"
    lb_port           = 443
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.api-elb-custom-security-groups-example-com.id}"]
  subnets         = ["${aws_subnet.utility-us-test-1a-custom-security-groups-example-com.id}"]

  health_check = {
    target              = "SSL:443"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "api.custom-security-groups.example.com"
  }
}

resource "aws_elb" "bastion-custom-security-groups-example-com" {
  name = "bastion-custom-security-g-lufgs3"

  listener = {
    instance_port     = 22
    instance_protocol = "TCP"
    lb_port           = 22
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.bastion-elb-custom-security-groups-example-com.id}"]
  subnets         = ["${aws_subnet.utility-us-test-1a-custom-security-groups-example-com.id}"]

  health_check = {
    target              = "TCP:22"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "bastion.custom-security-groups.example.com"
  }
}

resource "aws_iam_instance_profile" "bastions-custom-security-groups-example-com" {
  name = "bastions.custom-security-groups.example.com"
  role = "${aws_iam_role.bastions-custom-security-groups-example-com.name}"
}

resource "aws_iam_instance_profile" "masters-custom-security-groups-example-com" {
  name = "masters.custom-security-groups.example.com"
  role = "${aws_iam_role.masters-custom-security-groups-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-custom-security-groups-example-com" {
  name = "nodes.custom-security-groups.example.com"
  role = "${aws_iam_role.nodes-custom-security-groups-example-com.name}"
}

resource "aws_iam_role" "bastions-custom-security-groups-example-com" {
  name               = "bastions.custom-security-groups.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_bastions.custom-security-groups.example.com_policy")}"
}

resource "aws_iam_role" "masters-custom-security-groups-example-com" {
  name               = "masters.custom-security-groups.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.custom-security-groups.example.com_policy")}"
}

resource "aws_iam_role" "nodes-custom-security-groups-example-com" {
  name               = "nodes.custom-security-groups.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.custom-security-groups.example.com_policy")}"
}

resource "aws_iam_role_policy" "bastions-custom-security-groups-example-com" {
  name   = "bastions.custom-security-groups.example.com"
  role   = "${aws_iam_role.bastions-custom-security-groups-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_bastions.custom-security-groups.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-custom-security-groups-example-com" {
  name   = "masters.custom-security-groups.example.com"
  role   = "${aws_iam_role.masters-custom-security-groups-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.custom-security-groups.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-custom-security-groups-example-com" {
  name   = "nodes.custom-security-groups.example.com"
  role   = "${aws_iam_role.nodes-custom-security-groups-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.custom-security-groups.example.com_policy")}"
}

resource "aws_internet_gateway" "custom-security-groups-example-com" {
  vpc_id = "${aws_vpc.custom-security-groups-example-com.id}"

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "custom-security-groups.example.com"
  }
}

resource "aws_key_pair" "kubernetes-custom-security-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.custom-security-groups.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.custom-security-groups.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "bastion-custom-security-groups-example-com" {
  name_prefix                 = "bastion.custom-security-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.micro"
  key_name                    = "${aws_key_pair.kubernetes-custom-security-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.bastions-custom-security-groups-example-com.id}"
  security_groups             = ["${aws_security_group.bastion-custom-security-groups-example-com.id}"]
  associate_public_ip_address = true

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 32
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }
}

resource "aws_launch_configuration" "master-us-test-1a-masters-custom-security-groups-example-com" {
  name_prefix                 = "master-us-test-1a.masters.custom-security-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-custom-security-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-custom-security-groups-example-com.id}"
  security_groups             = ["${aws_security_group.masters-custom-security-groups-example-com.id}", "sg-exampleid3", "sg-exampleid4"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.custom-security-groups.example.com_user_data")}"

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

resource "aws_launch_configuration" "nodes-custom-security-groups-example-com" {
  name_prefix                 = "nodes.custom-security-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-custom-security-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-custom-security-groups-example-com.id}"
  security_groups             = ["${aws_security_group.nodes-custom-security-groups-example-com.id}", "sg-exampleid", "sg-exampleid2"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.custom-security-groups.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 128
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }
}

resource "aws_nat_gateway" "us-test-1a-custom-security-groups-example-com" {
  allocation_id = "${aws_eip.us-test-1a-custom-security-groups-example-com.id}"
  subnet_id     = "${aws_subnet.utility-us-test-1a-custom-security-groups-example-com.id}"
}

resource "aws_route" "0-0-0-0--0" {
  route_table_id         = "${aws_route_table.custom-security-groups-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.custom-security-groups-example-com.id}"
}

resource "aws_route" "private-us-test-1a-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-us-test-1a-custom-security-groups-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.us-test-1a-custom-security-groups-example-com.id}"
}

resource "aws_route53_record" "api-custom-security-groups-example-com" {
  name = "api.custom-security-groups.example.com"
  type = "A"

  alias = {
    name                   = "${aws_elb.api-custom-security-groups-example-com.dns_name}"
    zone_id                = "${aws_elb.api-custom-security-groups-example-com.zone_id}"
    evaluate_target_health = false
  }

  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route_table" "custom-security-groups-example-com" {
  vpc_id = "${aws_vpc.custom-security-groups-example-com.id}"

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "custom-security-groups.example.com"
  }
}

resource "aws_route_table" "private-us-test-1a-custom-security-groups-example-com" {
  vpc_id = "${aws_vpc.custom-security-groups-example-com.id}"

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "private-us-test-1a.custom-security-groups.example.com"
  }
}

resource "aws_route_table_association" "private-us-test-1a-custom-security-groups-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-custom-security-groups-example-com.id}"
  route_table_id = "${aws_route_table.private-us-test-1a-custom-security-groups-example-com.id}"
}

resource "aws_route_table_association" "utility-us-test-1a-custom-security-groups-example-com" {
  subnet_id      = "${aws_subnet.utility-us-test-1a-custom-security-groups-example-com.id}"
  route_table_id = "${aws_route_table.custom-security-groups-example-com.id}"
}

resource "aws_security_group" "api-elb-custom-security-groups-example-com" {
  name        = "api-elb.custom-security-groups.example.com"
  vpc_id      = "${aws_vpc.custom-security-groups-example-com.id}"
  description = "Security group for api ELB"

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "api-elb.custom-security-groups.example.com"
  }
}

resource "aws_security_group" "bastion-custom-security-groups-example-com" {
  name        = "bastion.custom-security-groups.example.com"
  vpc_id      = "${aws_vpc.custom-security-groups-example-com.id}"
  description = "Security group for bastion"

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "bastion.custom-security-groups.example.com"
  }
}

resource "aws_security_group" "bastion-elb-custom-security-groups-example-com" {
  name        = "bastion-elb.custom-security-groups.example.com"
  vpc_id      = "${aws_vpc.custom-security-groups-example-com.id}"
  description = "Security group for bastion ELB"

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "bastion-elb.custom-security-groups.example.com"
  }
}

resource "aws_security_group" "masters-custom-security-groups-example-com" {
  name        = "masters.custom-security-groups.example.com"
  vpc_id      = "${aws_vpc.custom-security-groups-example-com.id}"
  description = "Security group for masters"

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "masters.custom-security-groups.example.com"
  }
}

resource "aws_security_group" "nodes-custom-security-groups-example-com" {
  name        = "nodes.custom-security-groups.example.com"
  vpc_id      = "${aws_vpc.custom-security-groups-example-com.id}"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "nodes.custom-security-groups.example.com"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-custom-security-groups-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "api-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.api-elb-custom-security-groups-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-custom-security-groups-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-elb-custom-security-groups-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-to-master-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-custom-security-groups-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "bastion-to-node-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-custom-security-groups-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.api-elb-custom-security-groups-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "https-elb-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.api-elb-custom-security-groups-example-com.id}"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-custom-security-groups-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-custom-security-groups-example-com.id}"
  from_port                = 1
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-custom-security-groups-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-custom-security-groups-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-elb-to-bastion" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.bastion-custom-security-groups-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-elb-custom-security-groups-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "ssh-external-to-bastion-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.bastion-elb-custom-security-groups-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_subnet" "us-test-1a-custom-security-groups-example-com" {
  vpc_id            = "${aws_vpc.custom-security-groups-example-com.id}"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                          = "custom-security-groups.example.com"
    Name                                                       = "us-test-1a.custom-security-groups.example.com"
    "kubernetes.io/cluster/custom-security-groups.example.com" = "owned"
  }
}

resource "aws_subnet" "utility-us-test-1a-custom-security-groups-example-com" {
  vpc_id            = "${aws_vpc.custom-security-groups-example-com.id}"
  cidr_block        = "172.20.4.0/22"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                          = "custom-security-groups.example.com"
    Name                                                       = "utility-us-test-1a.custom-security-groups.example.com"
    "kubernetes.io/cluster/custom-security-groups.example.com" = "owned"
  }
}

resource "aws_vpc" "custom-security-groups-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    KubernetesCluster                                          = "custom-security-groups.example.com"
    Name                                                       = "custom-security-groups.example.com"
    "kubernetes.io/cluster/custom-security-groups.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "custom-security-groups-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    KubernetesCluster = "custom-security-groups.example.com"
    Name              = "custom-security-groups.example.com"
  }
}

resource "aws_vpc_dhcp_options_association" "custom-security-groups-example-com" {
  vpc_id          = "${aws_vpc.custom-security-groups-example-com.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.custom-security-groups-example-com.id}"
}

terraform = {
  required_version = ">= 0.9.3"
}
