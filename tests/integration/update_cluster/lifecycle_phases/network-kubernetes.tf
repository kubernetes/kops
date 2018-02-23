output "cluster_name" {
  value = "privateweave.example.com"
}

output "region" {
  value = "us-test-1"
}

output "vpc_id" {
  value = "${aws_vpc.privateweave-example-com.id}"
}

provider "aws" {
  region = "us-test-1"
}

resource "aws_eip" "us-test-1a-privateweave-example-com" {
  vpc = true
}

resource "aws_internet_gateway" "privateweave-example-com" {
  vpc_id = "${aws_vpc.privateweave-example-com.id}"

  tags = {
    KubernetesCluster                                = "privateweave.example.com"
    Name                                             = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_nat_gateway" "us-test-1a-privateweave-example-com" {
  allocation_id = "${aws_eip.us-test-1a-privateweave-example-com.id}"
  subnet_id     = "${aws_subnet.utility-us-test-1a-privateweave-example-com.id}"
}

resource "aws_route" "0-0-0-0--0" {
  route_table_id         = "${aws_route_table.privateweave-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.privateweave-example-com.id}"
}

resource "aws_route" "private-us-test-1a-0-0-0-0--0" {
  route_table_id         = "${aws_route_table.private-us-test-1a-privateweave-example-com.id}"
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = "${aws_nat_gateway.us-test-1a-privateweave-example-com.id}"
}

resource "aws_route_table" "private-us-test-1a-privateweave-example-com" {
  vpc_id = "${aws_vpc.privateweave-example-com.id}"

  tags = {
    KubernetesCluster                                = "privateweave.example.com"
    Name                                             = "private-us-test-1a.privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_route_table" "privateweave-example-com" {
  vpc_id = "${aws_vpc.privateweave-example-com.id}"

  tags = {
    KubernetesCluster                                = "privateweave.example.com"
    Name                                             = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_route_table_association" "private-us-test-1a-privateweave-example-com" {
  subnet_id      = "${aws_subnet.us-test-1a-privateweave-example-com.id}"
  route_table_id = "${aws_route_table.private-us-test-1a-privateweave-example-com.id}"
}

resource "aws_route_table_association" "utility-us-test-1a-privateweave-example-com" {
  subnet_id      = "${aws_subnet.utility-us-test-1a-privateweave-example-com.id}"
  route_table_id = "${aws_route_table.privateweave-example-com.id}"
}

resource "aws_subnet" "us-test-1a-privateweave-example-com" {
  vpc_id            = "${aws_vpc.privateweave-example-com.id}"
  cidr_block        = "172.20.32.0/19"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                = "privateweave.example.com"
    Name                                             = "us-test-1a.privateweave.example.com"
    SubnetType                                       = "Private"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
    "kubernetes.io/role/internal-elb"                = "1"
  }
}

resource "aws_subnet" "utility-us-test-1a-privateweave-example-com" {
  vpc_id            = "${aws_vpc.privateweave-example-com.id}"
  cidr_block        = "172.20.4.0/22"
  availability_zone = "us-test-1a"

  tags = {
    KubernetesCluster                                = "privateweave.example.com"
    Name                                             = "utility-us-test-1a.privateweave.example.com"
    SubnetType                                       = "Utility"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
    "kubernetes.io/role/elb"                         = "1"
  }
}

resource "aws_vpc" "privateweave-example-com" {
  cidr_block           = "172.20.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    KubernetesCluster                                = "privateweave.example.com"
    Name                                             = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "privateweave-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]

  tags = {
    KubernetesCluster                                = "privateweave.example.com"
    Name                                             = "privateweave.example.com"
    "kubernetes.io/cluster/privateweave.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options_association" "privateweave-example-com" {
  vpc_id          = "${aws_vpc.privateweave-example-com.id}"
  dhcp_options_id = "${aws_vpc_dhcp_options.privateweave-example-com.id}"
}

terraform = {
  required_version = ">= 0.9.3"
}
