locals = {
  cluster_name                      = "lifecyclephases.example.com"
  region                            = "us-test-1"
  route_table_private-us-test-1a_id = "${aws_route_table.private-us-test-1a-lifecyclephases-example-com.id}"
  route_table_public_id             = "${aws_route_table.lifecyclephases-example-com.id}"
  subnet_us-test-1a_id              = "${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"
  subnet_utility-us-test-1a_id      = "${aws_subnet.utility-us-test-1a-lifecyclephases-example-com.id}"
  vpc_cidr_block                    = "${aws_vpc.lifecyclephases-example-com.cidr_block}"
  vpc_id                            = "${aws_vpc.lifecyclephases-example-com.id}"
}

output "cluster_name" {
  value = "lifecyclephases.example.com"
}

output "region" {
  value = "us-test-1"
}

output "route_table_private-us-test-1a_id" {
  value = "${aws_route_table.private-us-test-1a-lifecyclephases-example-com.id}"
}

output "route_table_public_id" {
  value = "${aws_route_table.lifecyclephases-example-com.id}"
}

output "subnet_us-test-1a_id" {
  value = "${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"
}

output "subnet_utility-us-test-1a_id" {
  value = "${aws_subnet.utility-us-test-1a-lifecyclephases-example-com.id}"
}

output "vpc_cidr_block" {
  value = "${aws_vpc.lifecyclephases-example-com.cidr_block}"
}

output "vpc_id" {
  value = "${aws_vpc.lifecyclephases-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_eip" "us-test-1a-lifecyclephases-example-com" {
  vpc = true

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "us-test-1a.lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_internet_gateway" "lifecyclephases-example-com" {
  vpc_id = "${aws_vpc.lifecyclephases-example-com.id}"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_nat_gateway" "us-test-1a-lifecyclephases-example-com" {
  allocation_id = "${aws_eip.us-test-1a-lifecyclephases-example-com.id}"
  subnet_id     = "${aws_subnet.utility-us-test-1a-lifecyclephases-example-com.id}"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "us-test-1a.lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_route" "0-0-0-0--0" {
  route_table_id         = "${aws_route_table.lifecyclephases-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.lifecyclephases-example-com.id}"
}

resource "aws_route" "private-us-test-1a-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-us-test-1a-lifecyclephases-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.us-test-1a-lifecyclephases-example-com.id}"
}

resource "aws_route_table" "lifecyclephases-example-com" {
  vpc_id = "${aws_vpc.lifecyclephases-example-com.id}"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
    "kubernetes.io/kops/role"                           = "public"
  }
}

resource "aws_route_table" "private-us-test-1a-lifecyclephases-example-com" {
  vpc_id = "${aws_vpc.lifecyclephases-example-com.id}"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "private-us-test-1a.lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
    "kubernetes.io/kops/role"                           = "private-us-test-1a"
  }
}

resource "aws_route_table_association" "private-us-test-1a-lifecyclephases-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-lifecyclephases-example-com.id}"
  route_table_id = "${aws_route_table.private-us-test-1a-lifecyclephases-example-com.id}"
}

resource "aws_route_table_association" "utility-us-test-1a-lifecyclephases-example-com" {
  subnet_id      = "${aws_subnet.utility-us-test-1a-lifecyclephases-example-com.id}"
  route_table_id = "${aws_route_table.lifecyclephases-example-com.id}"
}

resource "aws_subnet" "us-test-1a-lifecyclephases-example-com" {
  vpc_id            = "${aws_vpc.lifecyclephases-example-com.id}"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "us-test-1a.lifecyclephases.example.com"
    SubnetType                                          = "Private"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
    "kubernetes.io/role/internal-elb"                   = "1"
  }
}

resource "aws_subnet" "utility-us-test-1a-lifecyclephases-example-com" {
  vpc_id            = "${aws_vpc.lifecyclephases-example-com.id}"
  cidr_block        = "172.20.4.0/22"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "utility-us-test-1a.lifecyclephases.example.com"
    SubnetType                                          = "Utility"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
    "kubernetes.io/role/elb"                            = "1"
  }
}

resource "aws_vpc" "lifecyclephases-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "lifecyclephases-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    KubernetesCluster                                   = "lifecyclephases.example.com"
    Name                                                = "lifecyclephases.example.com"
    "kubernetes.io/cluster/lifecyclephases.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options_association" "lifecyclephases-example-com" {
  vpc_id          = "${aws_vpc.lifecyclephases-example-com.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.lifecyclephases-example-com.id}"
}

terraform = {
  required_version = ">= 0.9.3"
}
