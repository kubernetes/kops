output "cluster_name" {
  value = "sec-groups.example.com"
}

output "master_security_group_ids" {
  value = ["${aws_security_group.sg-c4d4a3b7.id}", "${aws_security_group.sg-c4d4a3b8.id}", "${aws_security_group.sg-c4d4a3b9.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-sec-groups-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-sec-groups-example-com.name}"
}

output "node_security_group_ids" {
  value = ["${aws_security_group.sg-15eb9c66.id}"]
}

output "node_subnet_ids" {
  value = ["subnet-12345671", "subnet-12345678", "subnet-12345679"]
}

output "nodes_role_arn" {
  value = "${aws_iam_role.nodes-sec-groups-example-com.arn}"
}

output "nodes_role_name" {
  value = "${aws_iam_role.nodes-sec-groups-example-com.name}"
}

output "region" {
  value = "us-test-1"
}

output "subnet_ids" {
  value = ["subnet-12345671", "subnet-12345678", "subnet-12345679"]
}

output "vpc_id" {
  value = "vpc-0301e07b"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-sec-groups-example-com" {
  name                 = "master-us-test-1a.masters.sec-groups.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1a-masters-sec-groups-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["subnet-12345678"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1a.masters.sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "master-us-test-1b-masters-sec-groups-example-com" {
  name                 = "master-us-test-1b.masters.sec-groups.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1b-masters-sec-groups-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["subnet-12345679"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1b.masters.sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "master-us-test-1c-masters-sec-groups-example-com" {
  name                 = "master-us-test-1c.masters.sec-groups.example.com"
  launch_configuration = "${aws_launch_configuration.master-us-test-1c-masters-sec-groups-example-com.id}"
  max_size             = 1
  min_size             = 1
  vpc_zone_identifier  = ["subnet-12345671"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "master-us-test-1c.masters.sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/master"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_autoscaling_group" "nodes-sec-groups-example-com" {
  name                 = "nodes.sec-groups.example.com"
  launch_configuration = "${aws_launch_configuration.nodes-sec-groups-example-com.id}"
  max_size             = 2
  min_size             = 2
  vpc_zone_identifier  = ["subnet-12345678", "subnet-12345679", "subnet-12345671"]

  tag = {
    key                 = "KubernetesCluster"
    value               = "sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "Name"
    value               = "nodes.sec-groups.example.com"
    propagate_at_launch = true
  }

  tag = {
    key                 = "k8s.io/role/node"
    value               = "1"
    propagate_at_launch = true
  }

  metrics_granularity = "1Minute"
  enabled_metrics     = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
}

resource "aws_ebs_volume" "a-etcd-events-sec-groups-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "sec-groups.example.com"
    Name                 = "a.etcd-events.sec-groups.example.com"
    "k8s.io/etcd/events" = "a/a,b,c"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "a-etcd-main-sec-groups-example-com" {
  availability_zone = "us-test-1a"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "sec-groups.example.com"
    Name                 = "a.etcd-main.sec-groups.example.com"
    "k8s.io/etcd/main"   = "a/a,b,c"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "b-etcd-events-sec-groups-example-com" {
  availability_zone = "us-test-1b"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "sec-groups.example.com"
    Name                 = "b.etcd-events.sec-groups.example.com"
    "k8s.io/etcd/events" = "b/a,b,c"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "b-etcd-main-sec-groups-example-com" {
  availability_zone = "us-test-1b"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "sec-groups.example.com"
    Name                 = "b.etcd-main.sec-groups.example.com"
    "k8s.io/etcd/main"   = "b/a,b,c"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "c-etcd-events-sec-groups-example-com" {
  availability_zone = "us-test-1c"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "sec-groups.example.com"
    Name                 = "c.etcd-events.sec-groups.example.com"
    "k8s.io/etcd/events" = "c/a,b,c"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_ebs_volume" "c-etcd-main-sec-groups-example-com" {
  availability_zone = "us-test-1c"
  size              = 20
  type              = "gp2"
  encrypted         = false

  tags = {
    KubernetesCluster    = "sec-groups.example.com"
    Name                 = "c.etcd-main.sec-groups.example.com"
    "k8s.io/etcd/main"   = "c/a,b,c"
    "k8s.io/role/master" = "1"
  }
}

resource "aws_iam_instance_profile" "masters-sec-groups-example-com" {
  name = "masters.sec-groups.example.com"
  role = "${aws_iam_role.masters-sec-groups-example-com.name}"
}

resource "aws_iam_instance_profile" "nodes-sec-groups-example-com" {
  name = "nodes.sec-groups.example.com"
  role = "${aws_iam_role.nodes-sec-groups-example-com.name}"
}

resource "aws_iam_role" "masters-sec-groups-example-com" {
  name               = "masters.sec-groups.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_masters.sec-groups.example.com_policy")}"
}

resource "aws_iam_role" "nodes-sec-groups-example-com" {
  name               = "nodes.sec-groups.example.com"
  assume_role_policy = "${file("${path.module}/data/aws_iam_role_nodes.sec-groups.example.com_policy")}"
}

resource "aws_iam_role_policy" "masters-sec-groups-example-com" {
  name   = "masters.sec-groups.example.com"
  role   = "${aws_iam_role.masters-sec-groups-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_masters.sec-groups.example.com_policy")}"
}

resource "aws_iam_role_policy" "nodes-sec-groups-example-com" {
  name   = "nodes.sec-groups.example.com"
  role   = "${aws_iam_role.nodes-sec-groups-example-com.name}"
  policy = "${file("${path.module}/data/aws_iam_role_policy_nodes.sec-groups.example.com_policy")}"
}

resource "aws_key_pair" "kubernetes-sec-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.sec-groups.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = "${file("${path.module}/data/aws_key_pair_kubernetes.sec-groups.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")}"
}

resource "aws_launch_configuration" "master-us-test-1a-masters-sec-groups-example-com" {
  name_prefix                 = "master-us-test-1a.masters.sec-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-sec-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-sec-groups-example-com.id}"
  security_groups             = ["${aws_security_group.sg-c4d4a3b7.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1a.masters.sec-groups.example.com_user_data")}"

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

resource "aws_launch_configuration" "master-us-test-1b-masters-sec-groups-example-com" {
  name_prefix                 = "master-us-test-1b.masters.sec-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-sec-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-sec-groups-example-com.id}"
  security_groups             = ["${aws_security_group.sg-c4d4a3b8.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1b.masters.sec-groups.example.com_user_data")}"

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

resource "aws_launch_configuration" "master-us-test-1c-masters-sec-groups-example-com" {
  name_prefix                 = "master-us-test-1c.masters.sec-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "m3.medium"
  key_name                    = "${aws_key_pair.kubernetes-sec-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.masters-sec-groups-example-com.id}"
  security_groups             = ["${aws_security_group.sg-c4d4a3b9.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_master-us-test-1c.masters.sec-groups.example.com_user_data")}"

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

resource "aws_launch_configuration" "nodes-sec-groups-example-com" {
  name_prefix                 = "nodes.sec-groups.example.com-"
  image_id                    = "ami-12345678"
  instance_type               = "t2.medium"
  key_name                    = "${aws_key_pair.kubernetes-sec-groups-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id}"
  iam_instance_profile        = "${aws_iam_instance_profile.nodes-sec-groups-example-com.id}"
  security_groups             = ["${aws_security_group.sg-15eb9c66.id}"]
  associate_public_ip_address = true
  user_data                   = "${file("${path.module}/data/aws_launch_configuration_nodes.sec-groups.example.com_user_data")}"

  root_block_device = {
    volume_type           = "gp2"
    volume_size           = 128
    delete_on_termination = true
  }

  lifecycle = {
    create_before_destroy = true
  }
}

terraform = {
  required_version = ">= 0.9.3"
}
