locals = {
  cluster_name                 = "restrictaccess.example.com"
  master_autoscaling_group_ids = ["${aws_autoscaling_group.master-us-test-1a-masters-restrictaccess-example-com.id}"]
  master_security_group_ids    = ["${aws_security_group.masters-restrictaccess-example-com.id}"]
  masters_role_arn             = "${aws_iam_role.masters-restrictaccess-example-com.arn}"
  masters_role_name            = "${aws_iam_role.masters-restrictaccess-example-com.name}"
  node_autoscaling_group_ids   = ["${aws_autoscaling_group.nodes-restrictaccess-example-com.id}"]
  node_security_group_ids      = ["${aws_security_group.nodes-restrictaccess-example-com.id}"]
  node_subnet_ids              = ["${aws_subnet.us-test-1a-restrictaccess-example-com.id}"]
  nodes_role_arn               = "${aws_iam_role.nodes-restrictaccess-example-com.arn}"
  nodes_role_name              = "${aws_iam_role.nodes-restrictaccess-example-com.name}"
  region                       = "us-test-1"
  route_table_public_id        = "${aws_route_table.restrictaccess-example-com.id}"
  subnet_us-test-1a_id         = "${aws_subnet.us-test-1a-restrictaccess-example-com.id}"
  vpc_cidr_block               = "${aws_vpc.restrictaccess-example-com.cidr_block}"
  vpc_id                       = "${aws_vpc.restrictaccess-example-com.id}"
}

output "cluster_name" {
  value = "restrictaccess.example.com"
}

output "master_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.master-us-test-1a-masters-restrictaccess-example-com.id}"]
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-restrictaccess-example-com.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-restrictaccess-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-restrictaccess-example-com.name}"
}

output "node_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.nodes-restrictaccess-example-com.id}"]
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-restrictaccess-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-restrictaccess-example-com.id}"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-restrictaccess-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-restrictaccess-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

output "route_table_public_id" {
  value = "${aws_route_table.restrictaccess-example-com.id}"
}

output "subnet_us-test-1a_id" {
  value = "${aws_subnet.us-test-1a-restrictaccess-example-com.id}"
}

output "vpc_cidr_block" {
  value = "${aws_vpc.restrictaccess-example-com.cidr_block}"
}

output "vpc_id" {
  value = "${aws_vpc.restrictaccess-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-restrictaccess-example-com" {
  name                 = "master-us-test-1a.masters.restrictaccess.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-restrictaccess-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-restrictaccess-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "restrictaccess.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.restrictaccess.example.com"
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

  tag = {
    key                 = "kubernetes.io/cluster/restrictaccess.example.com"
    value               = "owned"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "nodes-restrictaccess-example-com" {
  name                 = "nodes.restrictaccess.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-restrictaccess-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-restrictaccess-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "restrictaccess.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.restrictaccess.example.com"
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

  tag = {
    key                 = "kubernetes.io/cluster/restrictaccess.example.com"
    value               = "owned"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-restrictaccess-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "us-test-1a.etcd-events.restrictaccess.example.com"
    "k8s.io/etcd/events"                               = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                               = "1"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
  }
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-restrictaccess-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "us-test-1a.etcd-main.restrictaccess.example.com"
    "k8s.io/etcd/main"                                 = "us-test-1a/us-test-1a"
    "k8s.io/role/master"                               = "1"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "masters-restrictaccess-example-com" {
  name = "masters.restrictaccess.example.com"
  role = "${aws_iam_role.masters-restrictaccess-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-restrictaccess-example-com" {
  name = "nodes.restrictaccess.example.com"
  role = "${aws_iam_role.nodes-restrictaccess-example-com.name}"
}

resource "aws_iam_role" "masters-restrictaccess-example-com" {
  name               = "masters.restrictaccess.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.restrictaccess.example.com_policy")}"
}

resource "aws_iam_role" "nodes-restrictaccess-example-com" {
  name               = "nodes.restrictaccess.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.restrictaccess.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-restrictaccess-example-com" {
  name   = "masters.restrictaccess.example.com"
  role   = "${aws_iam_role.masters-restrictaccess-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.restrictaccess.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-restrictaccess-example-com" {
  name   = "nodes.restrictaccess.example.com"
  role   = "${aws_iam_role.nodes-restrictaccess-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.restrictaccess.example.com_policy")}"
}

resource "aws_internet_gateway" "restrictaccess-example-com" {
  vpc_id = "${aws_vpc.restrictaccess-example-com.id}"

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "restrictaccess.example.com"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
  }
}

resource "aws_key_pair" "kubernetes-restrictaccess-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.restrictaccess.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.restrictaccess.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "master-us-test-1a-masters-restrictaccess-example-com" {
  name_prefix                 = "master-us-test-1a.masters.restrictaccess.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-restrictaccess-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-restrictaccess-example-com.id}"
  security_groups             = ["${aws_security_group.masters-restrictaccess-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.restrictaccess.example.com_user_data")}"

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

resource "aws_launch_configuration" "nodes-restrictaccess-example-com" {
  name_prefix                 = "nodes.restrictaccess.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-restrictaccess-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-restrictaccess-example-com.id}"
  security_groups             = ["${aws_security_group.nodes-restrictaccess-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.restrictaccess.example.com_user_data")}"

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

resource "aws_route" "route-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.restrictaccess-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.restrictaccess-example-com.id}"
}

resource "aws_route_table" "restrictaccess-example-com" {
  vpc_id = "${aws_vpc.restrictaccess-example-com.id}"

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "restrictaccess.example.com"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
    "kubernetes.io/kops/role"                          = "public"
  }
}

resource "aws_route_table_association" "us-test-1a-restrictaccess-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-restrictaccess-example-com.id}"
  route_table_id = "${aws_route_table.restrictaccess-example-com.id}"
}

resource "aws_security_group" "masters-restrictaccess-example-com" {
  name        = "masters.restrictaccess.example.com"
  vpc_id      = "${aws_vpc.restrictaccess-example-com.id}"
  description = "Security group for masters"

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "masters.restrictaccess.example.com"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
  }
}

resource "aws_security_group" "nodes-restrictaccess-example-com" {
  name        = "nodes.restrictaccess.example.com"
  vpc_id      = "${aws_vpc.restrictaccess-example-com.id}"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "nodes.restrictaccess.example.com"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-restrictaccess-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-restrictaccess-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-restrictaccess-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "https-external-to-master-1-1-1-0--24" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-restrictaccess-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["1.1.1.0/24"]
}

resource "aws_security_group_rule" "https-external-to-master-2001_0_8500__--40" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-restrictaccess-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["2001:0:8500::/40"]
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-restrictaccess-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-restrictaccess-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  from_port                = 1
  to_port                  = 2379
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-2382-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-restrictaccess-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  from_port                = 2382
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-restrictaccess-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-restrictaccess-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-external-to-master-1-1-1-1--32" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-restrictaccess-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["1.1.1.1/32"]
}

resource "aws_security_group_rule" "ssh-external-to-master-2001_0_85a3__--48" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-restrictaccess-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["2001:0:85a3::/48"]
}

resource "aws_security_group_rule" "ssh-external-to-node-1-1-1-1--32" {
  type              = "ingress"
  security_group_id = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["1.1.1.1/32"]
}

resource "aws_security_group_rule" "ssh-external-to-node-2001_0_85a3__--48" {
  type              = "ingress"
  security_group_id = "${aws_security_group.nodes-restrictaccess-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["2001:0:85a3::/48"]
}

resource "aws_subnet" "us-test-1a-restrictaccess-example-com" {
  vpc_id            = "${aws_vpc.restrictaccess-example-com.id}"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "us-test-1a.restrictaccess.example.com"
    SubnetType                                         = "Public"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
    "kubernetes.io/role/elb"                           = "1"
  }
}

resource "aws_vpc" "restrictaccess-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "restrictaccess.example.com"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "restrictaccess-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    KubernetesCluster                                  = "restrictaccess.example.com"
    Name                                               = "restrictaccess.example.com"
    "kubernetes.io/cluster/restrictaccess.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options_association" "restrictaccess-example-com" {
  vpc_id          = "${aws_vpc.restrictaccess-example-com.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.restrictaccess-example-com.id}"
}

terraform = {
  required_version = ">= 0.9.3"
}
