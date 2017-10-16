output "cluster_name" {
  value = "privateweave.example.com"
}

output "region" {
  value = "us-test-1"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_autoscaling_attachment" "bastion-privateweave-example-com" {
  elb                    = "${aws_elb.bastion-privateweave-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.bastion-privateweave-example-com.id}"
}

resource "aws_autoscaling_attachment" "master-us-test-1a-masters-privateweave-example-com" {
  elb                    = "${aws_elb.api-privateweave-example-com.id}"
  autoscaling_group_name = "${aws_autoscaling_group.master-us-test-1a-masters-privateweave-example-com.id}"
}

resource "aws_elb" "api-privateweave-example-com" {
  name = "api-privateweave-example--l94cb4"

  listener = {
    instance_port     = 443
    instance_protocol = "TCP"
    lb_port           = 443
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.api-elb-privateweave-example-com.id}"]
  subnets         = ["${aws_subnet.utility-us-test-1a-privateweave-example-com.id}"]

  health_check = {
    target              = "SSL:443"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster = "privateweave.example.com"
    Name              = "api.privateweave.example.com"
  }
}

resource "aws_elb" "bastion-privateweave-example-com" {
  name = "bastion-privateweave-exam-fdb6ge"

  listener = {
    instance_port     = 22
    instance_protocol = "TCP"
    lb_port           = 22
    lb_protocol       = "TCP"
  }

  security_groups = ["${aws_security_group.bastion-elb-privateweave-example-com.id}"]
  subnets         = ["${aws_subnet.utility-us-test-1a-privateweave-example-com.id}"]

  health_check = {
    target              = "TCP:22"
    healthy_threshold   = 2
    unhealthy_threshold = 2
    interval            = 10
    timeout             = 5
  }

  idle_timeout = 300

  tags = {
    KubernetesCluster = "privateweave.example.com"
    Name              = "bastion.privateweave.example.com"
  }
}

resource "aws_route53_record" "api-privateweave-example-com" {
  name = "api.privateweave.example.com"
  type = "A"

  alias = {
    name                   = "${aws_elb.api-privateweave-example-com.dns_name}"
    zone_id                = "${aws_elb.api-privateweave-example-com.zone_id}"
    evaluate_target_health = false
  }

  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

resource "aws_route53_record" "bastion-privateweave-example-com" {
  name = "bastion.privateweave.example.com"
  type = "A"

  alias = {
    name                   = "${aws_elb.bastion-privateweave-example-com.dns_name}"
    zone_id                = "${aws_elb.bastion-privateweave-example-com.zone_id}"
    evaluate_target_health = false
  }

  zone_id = "/hostedzone/Z1AFAKE1ZON3YO"
}

terraform = {
  required_version = ">= 0.9.3"
}
