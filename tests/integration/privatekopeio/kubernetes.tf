output "bastion_security_group_ids" {
  value = ["${aws_security_group.bastion-privatekopeio-example-com.id}"]
}

output "bastions_role_arn" {
  value = "${aws_iam_role.bastions-privatekopeio-example-com.arn}"
}

output "bastions_role_name" {
  value = "${aws_iam_role.bastions-privatekopeio-example-com.name}"
}

output "cluster_name" {
  value = "privatekopeio.example.com"
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-privatekopeio-example-com.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-privatekopeio-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-privatekopeio-example-com.name}"
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-privatekopeio-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-privatekopeio-example-com.id}"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-privatekopeio-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-privatekopeio-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

output "vpc_id" {
  value = "${aws_vpc.privatekopeio-example-com.id}"
}

resource "aws_autoscaling_attachment" "bastion-privatekopeio-example-com" {
  elb                    = "${aws_elb.bastion-privatekopeio-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.bastion-privatekopeio-example-com.id}"
}

resource "aws_autoscaling_attachment" "master-us-test-1a-masters-privatekopeio-example-com" {
  elb                    = "${aws_elb.api-privatekopeio-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.master-us-test-1a-masters-privatekopeio-example-com.id}"
}

resource "aws_autoscaling_group" "bastion-privatekopeio-example-com" {
  name                 = "bastion.privatekopeio.example.com"
  launch_configuration = "${aws_launch_configuration.bastion-privatekopeio-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.utility-us-test-1a-privatekopeio-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "privatekopeio.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "bastion.privatekopeio.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/bastion"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-privatekopeio-example-com" {
  name                 = "master-us-test-1a.masters.privatekopeio.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-privatekopeio-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-privatekopeio-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "privatekopeio.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.privatekopeio.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "nodes-privatekopeio-example-com" {
  name                 = "nodes.privatekopeio.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-privatekopeio-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-privatekopeio-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "privatekopeio.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.privatekopeio.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-privatekopeio-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "privatekopeio.example.com"
    Name                 = "us-test-1a.etcd-events.privatekopeio.example.com"
    "k8s.io/etcd/events" = "us-test-1a/us-test-1a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-privatekopeio-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "privatekopeio.example.com"
    Name                 = "us-test-1a.etcd-main.privatekopeio.example.com"
    "k8s.io/etcd/main"   = "us-test-1a/us-test-1a"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_elb" "api-privatekopeio-example-com" {
  name = "api-privatekopeio-example-tl2bv8"

  listener = {
    instance_port     = 443
    instance_protocol = "TCP"
    lb_port           = 443
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.api-elb-privatekopeio-example-com.id}"]
  subnets         = ["${aws_subnet.utility-us-test-1a-privatekopeio-example-com.id}"]

  health_check = {
    target              = "TCP:443"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "api.privatekopeio.example.com"
  }
}

resource "aws_elb" "bastion-privatekopeio-example-com" {
  name = "bastion-privatekopeio-exa-d8ef8e"

  listener = {
    instance_port     = 22
    instance_protocol = "TCP"
    lb_port           = 22
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.bastion-elb-privatekopeio-example-com.id}"]
  subnets         = ["${aws_subnet.utility-us-test-1a-privatekopeio-example-com.id}"]

  health_check = {
    target              = "TCP:22"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "bastion.privatekopeio.example.com"
  }
}

resource "aws_iam_instance_profile" "bastions-privatekopeio-example-com" {
  name = "bastions.privatekopeio.example.com"
  role = "${aws_iam_role.bastions-privatekopeio-example-com.name}"
}

resource "aws_iam_instance_profile" "masters-privatekopeio-example-com" {
  name = "masters.privatekopeio.example.com"
  role = "${aws_iam_role.masters-privatekopeio-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-privatekopeio-example-com" {
  name = "nodes.privatekopeio.example.com"
  role = "${aws_iam_role.nodes-privatekopeio-example-com.name}"
}

resource "aws_iam_role" "bastions-privatekopeio-example-com" {
  name               = "bastions.privatekopeio.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_bastions.privatekopeio.example.com_policy")}"
}

resource "aws_iam_role" "masters-privatekopeio-example-com" {
  name               = "masters.privatekopeio.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.privatekopeio.example.com_policy")}"
}

resource "aws_iam_role" "nodes-privatekopeio-example-com" {
  name               = "nodes.privatekopeio.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.privatekopeio.example.com_policy")}"
}

resource "aws_iam_role_policy" "bastions-privatekopeio-example-com" {
  name   = "bastions.privatekopeio.example.com"
  role   = "${aws_iam_role.bastions-privatekopeio-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_bastions.privatekopeio.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-privatekopeio-example-com" {
  name   = "masters.privatekopeio.example.com"
  role   = "${aws_iam_role.masters-privatekopeio-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.privatekopeio.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-privatekopeio-example-com" {
  name   = "nodes.privatekopeio.example.com"
  role   = "${aws_iam_role.nodes-privatekopeio-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.privatekopeio.example.com_policy")}"
}

resource "aws_internet_gateway" "privatekopeio-example-com" {
  vpc_id = "${aws_vpc.privatekopeio-example-com.id}"

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "privatekopeio.example.com"
  }
}

resource "aws_key_pair" "kubernetes-privatekopeio-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.privatekopeio.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.privatekopeio.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "bastion-privatekopeio-example-com" {
  name_prefix                 = "bastion.privatekopeio.example.com-"
  image_id                    = "ami-15000000"
  instance_type               = "t2.micro"
  key_name                    = "${aws_key_pair.kubernetes-privatekopeio-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.bastions-privatekopeio-example-com.id}"
  security_groups             = ["${aws_security_group.bastion-privatekopeio-example-com.id}"]
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

resource "aws_launch_configuration" "master-us-test-1a-masters-privatekopeio-example-com" {
  name_prefix                 = "master-us-test-1a.masters.privatekopeio.example.com-"
  image_id                    = "ami-15000000"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-privatekopeio-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-privatekopeio-example-com.id}"
  security_groups             = ["${aws_security_group.masters-privatekopeio-example-com.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.privatekopeio.example.com_user_data")}"

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

resource "aws_launch_configuration" "nodes-privatekopeio-example-com" {
  name_prefix                 = "nodes.privatekopeio.example.com-"
  image_id                    = "ami-15000000"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-privatekopeio-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-privatekopeio-example-com.id}"
  security_groups             = ["${aws_security_group.nodes-privatekopeio-example-com.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.privatekopeio.example.com_user_data")}"

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
  route_table_id         = "${aws_route_table.privatekopeio-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.privatekopeio-example-com.id}"
}

resource "aws_route" "private-us-test-1a-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-us-test-1a-privatekopeio-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "nat-12345678"
}

resource "aws_route53_record" "api-privatekopeio-example-com" {
  name = "api.privatekopeio.example.com"
  type = "A"

  alias = {
    name                   = "${aws_elb.api-privatekopeio-example-com.dns_name}"
    zone_id                = "${aws_elb.api-privatekopeio-example-com.zone_id}"
    evaluate_target_health = false
  }

  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route_table" "private-us-test-1a-privatekopeio-example-com" {
  vpc_id = "${aws_vpc.privatekopeio-example-com.id}"

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "private-us-test-1a.privatekopeio.example.com"
  }
}

resource "aws_route_table" "privatekopeio-example-com" {
  vpc_id = "${aws_vpc.privatekopeio-example-com.id}"

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "privatekopeio.example.com"
  }
}

resource "aws_route_table_association" "private-us-test-1a-privatekopeio-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-privatekopeio-example-com.id}"
  route_table_id = "${aws_route_table.private-us-test-1a-privatekopeio-example-com.id}"
}

resource "aws_route_table_association" "utility-us-test-1a-privatekopeio-example-com" {
  subnet_id      = "${aws_subnet.utility-us-test-1a-privatekopeio-example-com.id}"
  route_table_id = "${aws_route_table.privatekopeio-example-com.id}"
}

resource "aws_security_group" "api-elb-privatekopeio-example-com" {
  name        = "api-elb.privatekopeio.example.com"
  vpc_id      = "${aws_vpc.privatekopeio-example-com.id}"
  description = "Security group for api ELB"

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "api-elb.privatekopeio.example.com"
  }
}

resource "aws_security_group" "bastion-elb-privatekopeio-example-com" {
  name        = "bastion-elb.privatekopeio.example.com"
  vpc_id      = "${aws_vpc.privatekopeio-example-com.id}"
  description = "Security group for bastion ELB"

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "bastion-elb.privatekopeio.example.com"
  }
}

resource "aws_security_group" "bastion-privatekopeio-example-com" {
  name        = "bastion.privatekopeio.example.com"
  vpc_id      = "${aws_vpc.privatekopeio-example-com.id}"
  description = "Security group for bastion"

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "bastion.privatekopeio.example.com"
  }
}

resource "aws_security_group" "masters-privatekopeio-example-com" {
  name        = "masters.privatekopeio.example.com"
  vpc_id      = "${aws_vpc.privatekopeio-example-com.id}"
  description = "Security group for masters"

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "masters.privatekopeio.example.com"
  }
}

resource "aws_security_group" "nodes-privatekopeio-example-com" {
  name        = "nodes.privatekopeio.example.com"
  vpc_id      = "${aws_vpc.privatekopeio-example-com.id}"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "nodes.privatekopeio.example.com"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-privatekopeio-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-privatekopeio-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-privatekopeio-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "api-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.api-elb-privatekopeio-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-privatekopeio-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.bastion-elb-privatekopeio-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "bastion-to-master-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-privatekopeio-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "bastion-to-node-ssh" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-privatekopeio-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.api-elb-privatekopeio-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "https-elb-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.api-elb-privatekopeio-example-com.id}"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-privatekopeio-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-privatekopeio-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-privatekopeio-example-com.id}"
  from_port                = 1
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-privatekopeio-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-privatekopeio-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-elb-to-bastion" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.bastion-privatekopeio-example-com.id}"
  source_security_group_id = "${aws_security_group.bastion-elb-privatekopeio-example-com.id}"
  from_port                = 22
  to_port                  = 22
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "ssh-external-to-bastion-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.bastion-elb-privatekopeio-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_subnet" "us-test-1a-privatekopeio-example-com" {
  vpc_id            = "${aws_vpc.privatekopeio-example-com.id}"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                 = "privatekopeio.example.com"
    Name                                              = "us-test-1a.privatekopeio.example.com"
    "kubernetes.io/cluster/privatekopeio.example.com" = "owned"
  }
}

resource "aws_subnet" "utility-us-test-1a-privatekopeio-example-com" {
  vpc_id            = "${aws_vpc.privatekopeio-example-com.id}"
  cidr_block        = "172.20.4.0/22"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                 = "privatekopeio.example.com"
    Name                                              = "utility-us-test-1a.privatekopeio.example.com"
    "kubernetes.io/cluster/privatekopeio.example.com" = "owned"
  }
}

resource "aws_vpc" "privatekopeio-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    KubernetesCluster                                 = "privatekopeio.example.com"
    Name                                              = "privatekopeio.example.com"
    "kubernetes.io/cluster/privatekopeio.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "privatekopeio-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    KubernetesCluster = "privatekopeio.example.com"
    Name              = "privatekopeio.example.com"
  }
}

resource "aws_vpc_dhcp_options_association" "privatekopeio-example-com" {
  vpc_id          = "${aws_vpc.privatekopeio-example-com.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.privatekopeio-example-com.id}"
}

terraform = {
  required_version = ">= 0.9.3"
}
