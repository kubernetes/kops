locals = {
  cluster_name              = "k8s-iam.us-west-2.td.priv"
  master_security_group_ids = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  masters_role_arn          = "${aws_iam_role.masters-k8s-iam-us-west-2-td-priv.arn}"
  masters_role_name         = "${aws_iam_role.masters-k8s-iam-us-west-2-td-priv.name}"
  node_security_group_ids   = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  node_subnet_ids           = ["${aws_subnet.us-west-2a-k8s-iam-us-west-2-td-priv.id}", "${aws_subnet.us-west-2b-k8s-iam-us-west-2-td-priv.id}", "${aws_subnet.us-west-2c-k8s-iam-us-west-2-td-priv.id}"]
  nodes_role_arn            = "${aws_iam_role.nodes-k8s-iam-us-west-2-td-priv.arn}"
  nodes_role_name           = "${aws_iam_role.nodes-k8s-iam-us-west-2-td-priv.name}"
  region                    = "us-west-2"
  vpc_id                    = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
}

output "cluster_name" {
  value = "k8s-iam.us-west-2.td.priv"
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-k8s-iam-us-west-2-td-priv.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-k8s-iam-us-west-2-td-priv.name}"
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-west-2a-k8s-iam-us-west-2-td-priv.id}", "${aws_subnet.us-west-2b-k8s-iam-us-west-2-td-priv.id}", "${aws_subnet.us-west-2c-k8s-iam-us-west-2-td-priv.id}"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-k8s-iam-us-west-2-td-priv.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-k8s-iam-us-west-2-td-priv.name}"
}

output "region" {
  value = "us-west-2"
}

output "vpc_id" {
  value = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_autoscaling_attachment" "master-us-west-2a-masters-k8s-iam-us-west-2-td-priv" {
  elb                    = "${aws_elb.api-k8s-iam-us-west-2-td-priv.id}"
  autoscaling_group_name = "${aws_autoscaling_group.master-us-west-2a-masters-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_autoscaling_group" "master-us-west-2a-masters-k8s-iam-us-west-2-td-priv" {
  name                 = "master-us-west-2a.masters.k8s-iam.us-west-2.td.priv"
  launch_configuration = "${aws_launch_configuration.master-us-west-2a-masters-k8s-iam-us-west-2-td-priv.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-west-2a-k8s-iam-us-west-2-td-priv.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "k8s-iam.us-west-2.td.priv"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-west-2a.masters.k8s-iam.us-west-2.td.priv"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/instancegroup"
    value               = "master-us-west-2a"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  tag = {
    key                 = "terraform"
    value               = "true"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "nodes-k8s-iam-us-west-2-td-priv" {
  name                 = "nodes.k8s-iam.us-west-2.td.priv"
  launch_configuration = "${aws_launch_configuration.nodes-k8s-iam-us-west-2-td-priv.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-west-2a-k8s-iam-us-west-2-td-priv.id}", "${aws_subnet.us-west-2b-k8s-iam-us-west-2-td-priv.id}", "${aws_subnet.us-west-2c-k8s-iam-us-west-2-td-priv.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "k8s-iam.us-west-2.td.priv"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.k8s-iam.us-west-2.td.priv"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/instancegroup"
    value               = "nodes"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }

  tag = {
    key                 = "terraform"
    value               = "true"
    propagate_at_launch = true
  }
}

resource "aws_ebs_volume" "a-etcd-events-k8s-iam-us-west-2-td-priv" {
  availability_zone = "us-west-2a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "k8s-iam.us-west-2.td.priv"
    Name                 = "a.etcd-events.k8s-iam.us-west-2.td.priv"
    "k8s.io/etcd/events" = "a/a"
    "k8s.io/role/master" = "1"
    terraform            = "true"
  }
}

resource "aws_ebs_volume" "a-etcd-main-k8s-iam-us-west-2-td-priv" {
  availability_zone = "us-west-2a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "k8s-iam.us-west-2.td.priv"
    Name                 = "a.etcd-main.k8s-iam.us-west-2.td.priv"
    "k8s.io/etcd/main"   = "a/a"
    "k8s.io/role/master" = "1"
    terraform            = "true"
  }
}

resource "aws_eip" "us-west-2a-k8s-iam-us-west-2-td-priv" {
  vpc = true
}

resource "aws_eip" "us-west-2b-k8s-iam-us-west-2-td-priv" {
  vpc = true
}

resource "aws_eip" "us-west-2c-k8s-iam-us-west-2-td-priv" {
  vpc = true
}

resource "aws_elb" "api-k8s-iam-us-west-2-td-priv" {
  name = "api-k8s-iam-us-west-2-td--a7fd54"

  listener = {
    instance_port     = 443
    instance_protocol = "TCP"
    lb_port           = 443
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.api-elb-k8s-iam-us-west-2-td-priv.id}"]
  subnets         = ["${aws_subnet.us-west-2a-k8s-iam-us-west-2-td-priv.id}", "${aws_subnet.us-west-2b-k8s-iam-us-west-2-td-priv.id}", "${aws_subnet.us-west-2c-k8s-iam-us-west-2-td-priv.id}"]
  internal        = true

  health_check = {
    target              = "SSL:443"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "api.k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_iam_instance_profile" "masters-k8s-iam-us-west-2-td-priv" {
  name = "masters.k8s-iam.us-west-2.td.priv"
  role = "${aws_iam_role.masters-k8s-iam-us-west-2-td-priv.name}"
}

resource "aws_iam_instance_profile" "nodes-k8s-iam-us-west-2-td-priv" {
  name = "nodes.k8s-iam.us-west-2.td.priv"
  role = "${aws_iam_role.nodes-k8s-iam-us-west-2-td-priv.name}"
}

resource "aws_iam_role" "masters-k8s-iam-us-west-2-td-priv" {
  name               = "masters.k8s-iam.us-west-2.td.priv"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.k8s-iam.us-west-2.td.priv_policy")}"
}

resource "aws_iam_role" "nodes-k8s-iam-us-west-2-td-priv" {
  name               = "nodes.k8s-iam.us-west-2.td.priv"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.k8s-iam.us-west-2.td.priv_policy")}"
}

resource "aws_iam_role_policy" "masters-k8s-iam-us-west-2-td-priv" {
  name   = "masters.k8s-iam.us-west-2.td.priv"
  role   = "${aws_iam_role.masters-k8s-iam-us-west-2-td-priv.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.k8s-iam.us-west-2.td.priv_policy")}"
}

resource "aws_iam_role_policy" "nodes-k8s-iam-us-west-2-td-priv" {
  name   = "nodes.k8s-iam.us-west-2.td.priv"
  role   = "${aws_iam_role.nodes-k8s-iam-us-west-2-td-priv.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.k8s-iam.us-west-2.td.priv_policy")}"
}

resource "aws_internet_gateway" "k8s-iam-us-west-2-td-priv" {
  vpc_id = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_key_pair" "kubernetes-k8s-iam-us-west-2-td-priv-ad4e821eea9c965ed12a95b3bde99ed3" {
  key_name   = "kubernetes.k8s-iam.us-west-2.td.priv-ad:4e:82:1e:ea:9c:96:5e:d1:2a:95:b3:bd:e9:9e:d3"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.k8s-iam.us-west-2.td.priv-ad4e821eea9c965ed12a95b3bde99ed3_public_key")}"
}

resource "aws_launch_configuration" "master-us-west-2a-masters-k8s-iam-us-west-2-td-priv" {
  name_prefix                 = "master-us-west-2a.masters.k8s-iam.us-west-2.td.priv-"
  image_id                    = "ami-06a57e7e"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-k8s-iam-us-west-2-td-priv-ad4e821eea9c965ed12a95b3bde99ed3.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-k8s-iam-us-west-2-td-priv.id}"
  security_groups             = ["${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-west-2a.masters.k8s-iam.us-west-2.td.priv_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 64
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }
}

resource "aws_launch_configuration" "nodes-k8s-iam-us-west-2-td-priv" {
  name_prefix                 = "nodes.k8s-iam.us-west-2.td.priv-"
  image_id                    = "ami-06a57e7e"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-k8s-iam-us-west-2-td-priv-ad4e821eea9c965ed12a95b3bde99ed3.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-k8s-iam-us-west-2-td-priv.id}"
  security_groups             = ["${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"]
  associate_public_ip_address = false
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.k8s-iam.us-west-2.td.priv_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 128
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }
}

resource "aws_nat_gateway" "us-west-2a-k8s-iam-us-west-2-td-priv" {
  allocation_id = "${aws_eip.us-west-2a-k8s-iam-us-west-2-td-priv.id}"
  subnet_id     = "${aws_subnet.utility-us-west-2a-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_nat_gateway" "us-west-2b-k8s-iam-us-west-2-td-priv" {
  allocation_id = "${aws_eip.us-west-2b-k8s-iam-us-west-2-td-priv.id}"
  subnet_id     = "${aws_subnet.utility-us-west-2b-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_nat_gateway" "us-west-2c-k8s-iam-us-west-2-td-priv" {
  allocation_id = "${aws_eip.us-west-2c-k8s-iam-us-west-2-td-priv.id}"
  subnet_id     = "${aws_subnet.utility-us-west-2c-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route" "route-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.k8s-iam-us-west-2-td-priv.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route" "route-private-us-west-2a-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-us-west-2a-k8s-iam-us-west-2-td-priv.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.us-west-2a-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route" "route-private-us-west-2b-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-us-west-2b-k8s-iam-us-west-2-td-priv.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.us-west-2b-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route" "route-private-us-west-2c-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-us-west-2c-k8s-iam-us-west-2-td-priv.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.us-west-2c-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route53_record" "api-k8s-iam-us-west-2-td-priv" {
  name = "api.k8s-iam.us-west-2.td.priv"
  type = "A"

  alias = {
    name                   = "${aws_elb.api-k8s-iam-us-west-2-td-priv.dns_name}"
    zone_id                = "${aws_elb.api-k8s-iam-us-west-2-td-priv.zone_id}"
    evaluate_target_health = false
  }

  zone_id = "/hostedzone/Z1WJ08IMPUI44S"
}

resource "aws_route53_zone_association" "us-west-2-td-priv" {
  zone_id = "/hostedzone/Z1WJ08IMPUI44S"
  vpc_id  = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route_table" "k8s-iam-us-west-2-td-priv" {
  vpc_id = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_route_table" "private-us-west-2a-k8s-iam-us-west-2-td-priv" {
  vpc_id = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "private-us-west-2a.k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_route_table" "private-us-west-2b-k8s-iam-us-west-2-td-priv" {
  vpc_id = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "private-us-west-2b.k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_route_table" "private-us-west-2c-k8s-iam-us-west-2-td-priv" {
  vpc_id = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "private-us-west-2c.k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_route_table_association" "private-us-west-2a-k8s-iam-us-west-2-td-priv" {
  subnet_id      = "${aws_subnet.us-west-2a-k8s-iam-us-west-2-td-priv.id}"
  route_table_id = "${aws_route_table.private-us-west-2a-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route_table_association" "private-us-west-2b-k8s-iam-us-west-2-td-priv" {
  subnet_id      = "${aws_subnet.us-west-2b-k8s-iam-us-west-2-td-priv.id}"
  route_table_id = "${aws_route_table.private-us-west-2b-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route_table_association" "private-us-west-2c-k8s-iam-us-west-2-td-priv" {
  subnet_id      = "${aws_subnet.us-west-2c-k8s-iam-us-west-2-td-priv.id}"
  route_table_id = "${aws_route_table.private-us-west-2c-k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route_table_association" "utility-us-west-2a-k8s-iam-us-west-2-td-priv" {
  subnet_id      = "${aws_subnet.utility-us-west-2a-k8s-iam-us-west-2-td-priv.id}"
  route_table_id = "${aws_route_table.k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route_table_association" "utility-us-west-2b-k8s-iam-us-west-2-td-priv" {
  subnet_id      = "${aws_subnet.utility-us-west-2b-k8s-iam-us-west-2-td-priv.id}"
  route_table_id = "${aws_route_table.k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_route_table_association" "utility-us-west-2c-k8s-iam-us-west-2-td-priv" {
  subnet_id      = "${aws_subnet.utility-us-west-2c-k8s-iam-us-west-2-td-priv.id}"
  route_table_id = "${aws_route_table.k8s-iam-us-west-2-td-priv.id}"
}

resource "aws_security_group" "api-elb-k8s-iam-us-west-2-td-priv" {
  name        = "api-elb.k8s-iam.us-west-2.td.priv"
  vpc_id      = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  description = "Security group for api ELB"

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "api-elb.k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_security_group" "masters-k8s-iam-us-west-2-td-priv" {
  name        = "masters.k8s-iam.us-west-2.td.priv"
  vpc_id      = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  description = "Security group for masters"

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "masters.k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_security_group" "nodes-k8s-iam-us-west-2-td-priv" {
  name        = "nodes.k8s-iam.us-west-2.td.priv"
  vpc_id      = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "nodes.k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "api-elb-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.api-elb-k8s-iam-us-west-2-td-priv.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "https-api-elb-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.api-elb-k8s-iam-us-west-2-td-priv.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "https-elb-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.api-elb-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-protocol-ipip" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 0
  to_port                  = 65535
  protocol                 = "4"
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 1
  to_port                  = 2379
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-2382-4001" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 2382
  to_port                  = 4001
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  source_security_group_id = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-external-to-master-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-k8s-iam-us-west-2-td-priv.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "ssh-external-to-node-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.nodes-k8s-iam-us-west-2-td-priv.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_subnet" "us-west-2a-k8s-iam-us-west-2-td-priv" {
  vpc_id            = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  cidr_block        = "10.203.32.0/19"
  availability_zone = "us-west-2a"

  tags = {
    KubernetesCluster                                 = "k8s-iam.us-west-2.td.priv"
    Name                                              = "us-west-2a.k8s-iam.us-west-2.td.priv"
    "kubernetes.io/cluster/k8s-iam.us-west-2.td.priv" = "owned"
    "kubernetes.io/role/internal-elb"                 = "1"
  }
}

resource "aws_subnet" "us-west-2b-k8s-iam-us-west-2-td-priv" {
  vpc_id            = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  cidr_block        = "10.203.64.0/19"
  availability_zone = "us-west-2b"

  tags = {
    KubernetesCluster                                 = "k8s-iam.us-west-2.td.priv"
    Name                                              = "us-west-2b.k8s-iam.us-west-2.td.priv"
    "kubernetes.io/cluster/k8s-iam.us-west-2.td.priv" = "owned"
    "kubernetes.io/role/internal-elb"                 = "1"
  }
}

resource "aws_subnet" "us-west-2c-k8s-iam-us-west-2-td-priv" {
  vpc_id            = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  cidr_block        = "10.203.96.0/19"
  availability_zone = "us-west-2c"

  tags = {
    KubernetesCluster                                 = "k8s-iam.us-west-2.td.priv"
    Name                                              = "us-west-2c.k8s-iam.us-west-2.td.priv"
    "kubernetes.io/cluster/k8s-iam.us-west-2.td.priv" = "owned"
    "kubernetes.io/role/internal-elb"                 = "1"
  }
}

resource "aws_subnet" "utility-us-west-2a-k8s-iam-us-west-2-td-priv" {
  vpc_id            = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  cidr_block        = "10.203.0.0/22"
  availability_zone = "us-west-2a"

  tags = {
    KubernetesCluster                                 = "k8s-iam.us-west-2.td.priv"
    Name                                              = "utility-us-west-2a.k8s-iam.us-west-2.td.priv"
    "kubernetes.io/cluster/k8s-iam.us-west-2.td.priv" = "owned"
    "kubernetes.io/role/elb"                          = "1"
  }
}

resource "aws_subnet" "utility-us-west-2b-k8s-iam-us-west-2-td-priv" {
  vpc_id            = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  cidr_block        = "10.203.4.0/22"
  availability_zone = "us-west-2b"

  tags = {
    KubernetesCluster                                 = "k8s-iam.us-west-2.td.priv"
    Name                                              = "utility-us-west-2b.k8s-iam.us-west-2.td.priv"
    "kubernetes.io/cluster/k8s-iam.us-west-2.td.priv" = "owned"
    "kubernetes.io/role/elb"                          = "1"
  }
}

resource "aws_subnet" "utility-us-west-2c-k8s-iam-us-west-2-td-priv" {
  vpc_id            = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  cidr_block        = "10.203.8.0/22"
  availability_zone = "us-west-2c"

  tags = {
    KubernetesCluster                                 = "k8s-iam.us-west-2.td.priv"
    Name                                              = "utility-us-west-2c.k8s-iam.us-west-2.td.priv"
    "kubernetes.io/cluster/k8s-iam.us-west-2.td.priv" = "owned"
    "kubernetes.io/role/elb"                          = "1"
  }
}

resource "aws_vpc" "k8s-iam-us-west-2-td-priv" {
  cidr_block           = "10.203.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    KubernetesCluster                                 = "k8s-iam.us-west-2.td.priv"
    Name                                              = "k8s-iam.us-west-2.td.priv"
    "kubernetes.io/cluster/k8s-iam.us-west-2.td.priv" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "k8s-iam-us-west-2-td-priv" {
  domain_name         = "us-west-2.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    KubernetesCluster = "k8s-iam.us-west-2.td.priv"
    Name              = "k8s-iam.us-west-2.td.priv"
  }
}

resource "aws_vpc_dhcp_options_association" "k8s-iam-us-west-2-td-priv" {
  vpc_id          = "${aws_vpc.k8s-iam-us-west-2-td-priv.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.k8s-iam-us-west-2-td-priv.id}"
}

terraform = {
  required_version = ">= 0.9.3"
}
