locals = {
  cluster_name                 = "existing-iam.example.com"
  master_autoscaling_group_ids = ["${aws_autoscaling_group.master-us-test-1a-masters-existing-iam-example-com.id}", "${aws_autoscaling_group.master-us-test-1b-masters-existing-iam-example-com.id}", "${aws_autoscaling_group.master-us-test-1c-masters-existing-iam-example-com.id}"]
  master_security_group_ids    = ["${aws_security_group.masters-existing-iam-example-com.id}"]
  node_autoscaling_group_ids   = ["${aws_autoscaling_group.nodes-existing-iam-example-com.id}"]
  node_security_group_ids      = ["${aws_security_group.nodes-existing-iam-example-com.id}"]
  node_subnet_ids              = ["${aws_subnet.us-test-1a-existing-iam-example-com.id}"]
  region                       = "us-test-1"
  route_table_public_id        = "${aws_route_table.existing-iam-example-com.id}"
  subnet_us-test-1a_id         = "${aws_subnet.us-test-1a-existing-iam-example-com.id}"
  subnet_us-test-1b_id         = "${aws_subnet.us-test-1b-existing-iam-example-com.id}"
  subnet_us-test-1c_id         = "${aws_subnet.us-test-1c-existing-iam-example-com.id}"
  vpc_cidr_block               = "${aws_vpc.existing-iam-example-com.cidr_block}"
  vpc_id                       = "${aws_vpc.existing-iam-example-com.id}"
}

output "cluster_name" {
  value = "existing-iam.example.com"
}

output "master_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.master-us-test-1a-masters-existing-iam-example-com.id}", "${aws_autoscaling_group.master-us-test-1b-masters-existing-iam-example-com.id}", "${aws_autoscaling_group.master-us-test-1c-masters-existing-iam-example-com.id}"]
}

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-existing-iam-example-com.id}"]
}

output "node_autoscaling_group_ids" {
  value = ["${aws_autoscaling_group.nodes-existing-iam-example-com.id}"]
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-existing-iam-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-existing-iam-example-com.id}"]
}

output "region" {
  value = "us-test-1"
}

output "route_table_public_id" {
  value = "${aws_route_table.existing-iam-example-com.id}"
}

output "subnet_us-test-1a_id" {
  value = "${aws_subnet.us-test-1a-existing-iam-example-com.id}"
}

output "subnet_us-test-1b_id" {
  value = "${aws_subnet.us-test-1b-existing-iam-example-com.id}"
}

output "subnet_us-test-1c_id" {
  value = "${aws_subnet.us-test-1c-existing-iam-example-com.id}"
}

output "vpc_cidr_block" {
  value = "${aws_vpc.existing-iam-example-com.cidr_block}"
}

output "vpc_id" {
  value = "${aws_vpc.existing-iam-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-existing-iam-example-com" {
  name                 = "master-us-test-1a.masters.existing-iam.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-existing-iam-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-existing-iam-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "existing-iam.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.existing-iam.example.com"
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
    key                 = "kubernetes.io/cluster/existing-iam.example.com"
    value               = "owned"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "master-us-test-1b-masters-existing-iam-example-com" {
  name                 = "master-us-test-1b.masters.existing-iam.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1b-masters-existing-iam-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1b-existing-iam-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "existing-iam.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1b.masters.existing-iam.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  tag = {
    key                 = "kops.k8s.io/instancegroup"
    value               = "master-us-test-1b"
    propagate_at_launch = true
  }

  tag = {
    key                 = "kubernetes.io/cluster/existing-iam.example.com"
    value               = "owned"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "master-us-test-1c-masters-existing-iam-example-com" {
  name                 = "master-us-test-1c.masters.existing-iam.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1c-masters-existing-iam-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.us-test-1c-existing-iam-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "existing-iam.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1c.masters.existing-iam.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  tag = {
    key                 = "kops.k8s.io/instancegroup"
    value               = "master-us-test-1c"
    propagate_at_launch = true
  }

  tag = {
    key                 = "kubernetes.io/cluster/existing-iam.example.com"
    value               = "owned"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "nodes-existing-iam-example-com" {
  name                 = "nodes.existing-iam.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-existing-iam-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["${aws_subnet.us-test-1a-existing-iam-example-com.id}"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "existing-iam.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.existing-iam.example.com"
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
    key                 = "kubernetes.io/cluster/existing-iam.example.com"
    value               = "owned"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_ebs_volume" "a-etcd-events-existing-iam-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "a.etcd-events.existing-iam.example.com"
    "k8s.io/etcd/events"                             = "a/a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
  }
}

resource "aws_ebs_volume" "a-etcd-main-existing-iam-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "a.etcd-main.existing-iam.example.com"
    "k8s.io/etcd/main"                               = "a/a"
    "k8s.io/role/master"                             = "1"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
  }
}

resource "aws_internet_gateway" "existing-iam-example-com" {
  vpc_id = "${aws_vpc.existing-iam-example-com.id}"

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "existing-iam.example.com"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
  }
}

resource "aws_key_pair" "kubernetes-existing-iam-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.existing-iam.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.existing-iam.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "master-us-test-1a-masters-existing-iam-example-com" {
  name_prefix                 = "master-us-test-1a.masters.existing-iam.example.com-"
  image_id                    = "ami-11400000"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-existing-iam-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "kops-custom-master-role"
  security_groups             = ["${aws_security_group.masters-existing-iam-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.existing-iam.example.com_user_data")}"

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

resource "aws_launch_configuration" "master-us-test-1b-masters-existing-iam-example-com" {
  name_prefix                 = "master-us-test-1b.masters.existing-iam.example.com-"
  image_id                    = "ami-11400000"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-existing-iam-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "kops-custom-master-role"
  security_groups             = ["${aws_security_group.masters-existing-iam-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1b.masters.existing-iam.example.com_user_data")}"

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

resource "aws_launch_configuration" "master-us-test-1c-masters-existing-iam-example-com" {
  name_prefix                 = "master-us-test-1c.masters.existing-iam.example.com-"
  image_id                    = "ami-11400000"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-existing-iam-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "kops-custom-master-role"
  security_groups             = ["${aws_security_group.masters-existing-iam-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1c.masters.existing-iam.example.com_user_data")}"

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

resource "aws_launch_configuration" "nodes-existing-iam-example-com" {
  name_prefix                 = "nodes.existing-iam.example.com-"
  image_id                    = "ami-11400000"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-existing-iam-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "kops-custom-node-role"
  security_groups             = ["${aws_security_group.nodes-existing-iam-example-com.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.existing-iam.example.com_user_data")}"

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
  route_table_id         = "${aws_route_table.existing-iam-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.existing-iam-example-com.id}"
}

resource "aws_route_table" "existing-iam-example-com" {
  vpc_id = "${aws_vpc.existing-iam-example-com.id}"

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "existing-iam.example.com"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
    "kubernetes.io/kops/role"                        = "public"
  }
}

resource "aws_route_table_association" "us-test-1a-existing-iam-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-existing-iam-example-com.id}"
  route_table_id = "${aws_route_table.existing-iam-example-com.id}"
}

resource "aws_route_table_association" "us-test-1b-existing-iam-example-com" {
  subnet_id      = "${aws_subnet.us-test-1b-existing-iam-example-com.id}"
  route_table_id = "${aws_route_table.existing-iam-example-com.id}"
}

resource "aws_route_table_association" "us-test-1c-existing-iam-example-com" {
  subnet_id      = "${aws_subnet.us-test-1c-existing-iam-example-com.id}"
  route_table_id = "${aws_route_table.existing-iam-example-com.id}"
}

resource "aws_security_group" "masters-existing-iam-example-com" {
  name        = "masters.existing-iam.example.com"
  vpc_id      = "${aws_vpc.existing-iam-example-com.id}"
  description = "Security group for masters"

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "masters.existing-iam.example.com"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
  }
}

resource "aws_security_group" "nodes-existing-iam-example-com" {
  name        = "nodes.existing-iam.example.com"
  vpc_id      = "${aws_vpc.existing-iam-example-com.id}"
  description = "Security group for nodes"

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "nodes.existing-iam.example.com"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
  }
}

resource "aws_security_group_rule" "all-master-to-master" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-existing-iam-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-existing-iam-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-master-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-existing-iam-example-com.id}"
  source_security_group_id = "${aws_security_group.masters-existing-iam-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "all-node-to-node" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.nodes-existing-iam-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-existing-iam-example-com.id}"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
}

resource "aws_security_group_rule" "https-external-to-master-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-existing-iam-example-com.id}"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "master-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.masters-existing-iam-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-egress" {
  type              = "egress"
  security_group_id = "${aws_security_group.nodes-existing-iam-example-com.id}"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "node-to-master-tcp-1-2379" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-existing-iam-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-existing-iam-example-com.id}"
  from_port                = 1
  to_port                  = 2379
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-2382-4000" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-existing-iam-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-existing-iam-example-com.id}"
  from_port                = 2382
  to_port                  = 4000
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-tcp-4003-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-existing-iam-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-existing-iam-example-com.id}"
  from_port                = 4003
  to_port                  = 65535
  protocol                 = "tcp"
}

resource "aws_security_group_rule" "node-to-master-udp-1-65535" {
  type                     = "ingress"
  security_group_id        = "${aws_security_group.masters-existing-iam-example-com.id}"
  source_security_group_id = "${aws_security_group.nodes-existing-iam-example-com.id}"
  from_port                = 1
  to_port                  = 65535
  protocol                 = "udp"
}

resource "aws_security_group_rule" "ssh-external-to-master-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.masters-existing-iam-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "ssh-external-to-node-0-0-0-0--0" {
  type              = "ingress"
  security_group_id = "${aws_security_group.nodes-existing-iam-example-com.id}"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_subnet" "us-test-1a-existing-iam-example-com" {
  vpc_id            = "${aws_vpc.existing-iam-example-com.id}"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "us-test-1a.existing-iam.example.com"
    SubnetType                                       = "Public"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
    "kubernetes.io/role/elb"                         = "1"
  }
}

resource "aws_subnet" "us-test-1b-existing-iam-example-com" {
  vpc_id            = "${aws_vpc.existing-iam-example-com.id}"
  cidr_block        = "172.20.64.0/19"
  availability_zone = "us-test-1b"

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "us-test-1b.existing-iam.example.com"
    SubnetType                                       = "Public"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
    "kubernetes.io/role/elb"                         = "1"
  }
}

resource "aws_subnet" "us-test-1c-existing-iam-example-com" {
  vpc_id            = "${aws_vpc.existing-iam-example-com.id}"
  cidr_block        = "172.20.96.0/19"
  availability_zone = "us-test-1c"

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "us-test-1c.existing-iam.example.com"
    SubnetType                                       = "Public"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
    "kubernetes.io/role/elb"                         = "1"
  }
}

resource "aws_vpc" "existing-iam-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "existing-iam.example.com"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "existing-iam-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    KubernetesCluster                                = "existing-iam.example.com"
    Name                                             = "existing-iam.example.com"
    "kubernetes.io/cluster/existing-iam.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options_association" "existing-iam-example-com" {
  vpc_id          = "${aws_vpc.existing-iam-example-com.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.existing-iam-example-com.id}"
}

terraform = {
  required_version = ">= 0.9.3"
}
