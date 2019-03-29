locals = {
  bastion_security_group_ids = ["${aws_security_group.bastion-lifecyclephases-example-com.id}"]
  bastions_role_arn          = "${aws_iam_role.bastions-lifecyclephases-example-com.arn}"
  bastions_role_name         = "${aws_iam_role.bastions-lifecyclephases-example-com.name}"
  cluster_name               = "lifecyclephases.example.com"
  master_security_group_ids  = ["${aws_security_group.masters-lifecyclephases-example-com.id}"]
  masters_role_arn           = "${aws_iam_role.masters-lifecyclephases-example-com.arn}"
  masters_role_name          = "${aws_iam_role.masters-lifecyclephases-example-com.name}"
  node_security_group_ids    = ["${aws_security_group.nodes-lifecyclephases-example-com.id}"]
  node_subnet_ids            = ["${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"]
  nodes_role_arn             = "${aws_iam_role.nodes-lifecyclephases-example-com.arn}"
  nodes_role_name            = "${aws_iam_role.nodes-lifecyclephases-example-com.name}"
  region                     = "us-test-1"
  vpc_id                     = "${aws_vpc.lifecyclephases-example-com.id}"
}

output "bastion_security_group_ids" {
  value = ["${aws_security_group.bastion-lifecyclephases-example-com.id}"]
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

output "master_security_group_ids" {
  value = ["${aws_security_group.masters-lifecyclephases-example-com.id}"]
}

output "masters_role_arn" {
  value = "${aws_iam_role.masters-lifecyclephases-example-com.arn}"
}

output "masters_role_name" {
  value = "${aws_iam_role.masters-lifecyclephases-example-com.name}"
}

output "node_security_group_ids" {
  value = ["${aws_security_group.nodes-lifecyclephases-example-com.id}"]
}

output "node_subnet_ids" {
  value = ["${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"]
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

output "vpc_id" {
  value = "${aws_vpc.lifecyclephases-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_attachment" "bastion-lifecyclephases-example-com" {
  elb                    = "${aws_elb.bastion-lifecyclephases-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.bastion-lifecyclephases-example-com.id}"
}

resource "aws_autoscaling_attachment" "master-us-test-1a-masters-lifecyclephases-example-com" {
  elb                    = "${aws_elb.api-lifecyclephases-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.master-us-test-1a-masters-lifecyclephases-example-com.id}"
}

resource "aws_elb" "api-lifecyclephases-example-com" {
  name = "api-lifecyclephases-example--l94cb4"

  listener = {
    instance_port     = 443
    instance_protocol = "TCP"
    lb_port           = 443
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.api-elb-lifecyclephases-example-com.id}"]
  subnets         = ["${aws_subnet.utility-us-test-1a-lifecyclephases-example-com.id}"]

  health_check = {
    target              = "SSL:443"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster = "lifecyclephases.example.com"
    Name              = "api.lifecyclephases.example.com"
  }
}

resource "aws_elb" "bastion-lifecyclephases-example-com" {
  name = "bastion-lifecyclephases-exam-fdb6ge"

  listener = {
    instance_port     = 22
    instance_protocol = "TCP"
    lb_port           = 22
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.bastion-elb-lifecyclephases-example-com.id}"]
  subnets         = ["${aws_subnet.utility-us-test-1a-lifecyclephases-example-com.id}"]

  health_check = {
    target              = "TCP:22"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "bastion.lifecyclephases.example.com"
    "kubernetes.io/cluster/bastionuserdata.example.com" = "owned"
  }
}

resource "aws_route53_record" "api-lifecyclephases-example-com" {
  name = "api.lifecyclephases.example.com"
  type = "A"

  alias = {
    name                   = "${aws_elb.api-lifecyclephases-example-com.dns_name}"
    zone_id                = "${aws_elb.api-lifecyclephases-example-com.zone_id}"
    evaluate_target_health = false
  }

  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

terraform = {
  required_version = ">= 0.9.3"
}
