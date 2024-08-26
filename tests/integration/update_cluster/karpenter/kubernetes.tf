locals {
  cluster_name                                       = "minimal.example.com"
  iam_openid_connect_provider_arn                    = aws_iam_openid_connect_provider.minimal-example-com.arn
  iam_openid_connect_provider_issuer                 = "discovery.example.com/minimal.example.com"
  kube-system-aws-cloud-controller-manager_role_arn  = aws_iam_role.aws-cloud-controller-manager-kube-system-sa-minimal-example-com.arn
  kube-system-aws-cloud-controller-manager_role_name = aws_iam_role.aws-cloud-controller-manager-kube-system-sa-minimal-example-com.name
  kube-system-aws-node-termination-handler_role_arn  = aws_iam_role.aws-node-termination-handler-kube-system-sa-minimal-example-com.arn
  kube-system-aws-node-termination-handler_role_name = aws_iam_role.aws-node-termination-handler-kube-system-sa-minimal-example-com.name
  kube-system-dns-controller_role_arn                = aws_iam_role.dns-controller-kube-system-sa-minimal-example-com.arn
  kube-system-dns-controller_role_name               = aws_iam_role.dns-controller-kube-system-sa-minimal-example-com.name
  kube-system-ebs-csi-controller-sa_role_arn         = aws_iam_role.ebs-csi-controller-sa-kube-system-sa-minimal-example-com.arn
  kube-system-ebs-csi-controller-sa_role_name        = aws_iam_role.ebs-csi-controller-sa-kube-system-sa-minimal-example-com.name
  kube-system-karpenter_role_arn                     = aws_iam_role.karpenter-kube-system-sa-minimal-example-com.arn
  kube-system-karpenter_role_name                    = aws_iam_role.karpenter-kube-system-sa-minimal-example-com.name
  master_autoscaling_group_ids                       = [aws_autoscaling_group.master-us-test-1a-masters-minimal-example-com.id]
  master_security_group_ids                          = [aws_security_group.masters-minimal-example-com.id]
  masters_role_arn                                   = aws_iam_role.masters-minimal-example-com.arn
  masters_role_name                                  = aws_iam_role.masters-minimal-example-com.name
  node_autoscaling_group_ids                         = [aws_autoscaling_group.nodes-minimal-example-com.id]
  node_security_group_ids                            = [aws_security_group.nodes-minimal-example-com.id]
  node_subnet_ids                                    = [aws_subnet.us-test-1a-minimal-example-com.id]
  nodes_role_arn                                     = aws_iam_role.nodes-minimal-example-com.arn
  nodes_role_name                                    = aws_iam_role.nodes-minimal-example-com.name
  region                                             = "us-test-1"
  route_table_public_id                              = aws_route_table.minimal-example-com.id
  subnet_us-test-1a_id                               = aws_subnet.us-test-1a-minimal-example-com.id
  vpc_cidr_block                                     = aws_vpc.minimal-example-com.cidr_block
  vpc_id                                             = aws_vpc.minimal-example-com.id
  vpc_ipv6_cidr_block                                = aws_vpc.minimal-example-com.ipv6_cidr_block
  vpc_ipv6_cidr_length                               = local.vpc_ipv6_cidr_block == "" ? null : tonumber(regex(".*/(\\d+)", local.vpc_ipv6_cidr_block)[0])
}

output "cluster_name" {
  value = "minimal.example.com"
}

output "iam_openid_connect_provider_arn" {
  value = aws_iam_openid_connect_provider.minimal-example-com.arn
}

output "iam_openid_connect_provider_issuer" {
  value = "discovery.example.com/minimal.example.com"
}

output "kube-system-aws-cloud-controller-manager_role_arn" {
  value = aws_iam_role.aws-cloud-controller-manager-kube-system-sa-minimal-example-com.arn
}

output "kube-system-aws-cloud-controller-manager_role_name" {
  value = aws_iam_role.aws-cloud-controller-manager-kube-system-sa-minimal-example-com.name
}

output "kube-system-aws-node-termination-handler_role_arn" {
  value = aws_iam_role.aws-node-termination-handler-kube-system-sa-minimal-example-com.arn
}

output "kube-system-aws-node-termination-handler_role_name" {
  value = aws_iam_role.aws-node-termination-handler-kube-system-sa-minimal-example-com.name
}

output "kube-system-dns-controller_role_arn" {
  value = aws_iam_role.dns-controller-kube-system-sa-minimal-example-com.arn
}

output "kube-system-dns-controller_role_name" {
  value = aws_iam_role.dns-controller-kube-system-sa-minimal-example-com.name
}

output "kube-system-ebs-csi-controller-sa_role_arn" {
  value = aws_iam_role.ebs-csi-controller-sa-kube-system-sa-minimal-example-com.arn
}

output "kube-system-ebs-csi-controller-sa_role_name" {
  value = aws_iam_role.ebs-csi-controller-sa-kube-system-sa-minimal-example-com.name
}

output "kube-system-karpenter_role_arn" {
  value = aws_iam_role.karpenter-kube-system-sa-minimal-example-com.arn
}

output "kube-system-karpenter_role_name" {
  value = aws_iam_role.karpenter-kube-system-sa-minimal-example-com.name
}

output "master_autoscaling_group_ids" {
  value = [aws_autoscaling_group.master-us-test-1a-masters-minimal-example-com.id]
}

output "master_security_group_ids" {
  value = [aws_security_group.masters-minimal-example-com.id]
}

output "masters_role_arn" {
  value = aws_iam_role.masters-minimal-example-com.arn
}

output "masters_role_name" {
  value = aws_iam_role.masters-minimal-example-com.name
}

output "node_autoscaling_group_ids" {
  value = [aws_autoscaling_group.nodes-minimal-example-com.id]
}

output "node_security_group_ids" {
  value = [aws_security_group.nodes-minimal-example-com.id]
}

output "node_subnet_ids" {
  value = [aws_subnet.us-test-1a-minimal-example-com.id]
}

output "nodes_role_arn" {
  value = aws_iam_role.nodes-minimal-example-com.arn
}

output "nodes_role_name" {
  value = aws_iam_role.nodes-minimal-example-com.name
}

output "region" {
  value = "us-test-1"
}

output "route_table_public_id" {
  value = aws_route_table.minimal-example-com.id
}

output "subnet_us-test-1a_id" {
  value = aws_subnet.us-test-1a-minimal-example-com.id
}

output "vpc_cidr_block" {
  value = aws_vpc.minimal-example-com.cidr_block
}

output "vpc_id" {
  value = aws_vpc.minimal-example-com.id
}

output "vpc_ipv6_cidr_block" {
  value = aws_vpc.minimal-example-com.ipv6_cidr_block
}

output "vpc_ipv6_cidr_length" {
  value = local.vpc_ipv6_cidr_block == "" ? null : tonumber(regex(".*/(\\d+)", local.vpc_ipv6_cidr_block)[0])
}

provider "aws" {
  region = "us-test-1"
}

provider "aws" {
  alias  = "files"
  region = "us-test-1"
}

resource "aws_autoscaling_group" "master-us-test-1a-masters-minimal-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.master-us-test-1a-masters-minimal-example-com.id
    version = aws_launch_template.master-us-test-1a-masters-minimal-example-com.latest_version
  }
  max_instance_lifetime = 0
  max_size              = 1
  metrics_granularity   = "1Minute"
  min_size              = 1
  name                  = "master-us-test-1a.masters.minimal.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "minimal.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "master-us-test-1a.masters.minimal.example.com"
  }
  tag {
    key                 = "aws-node-termination-handler/managed"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/role/control-plane"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "k8s.io/role/master"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "kops.k8s.io/instancegroup"
    propagate_at_launch = true
    value               = "master-us-test-1a"
  }
  tag {
    key                 = "kubernetes.io/cluster/minimal.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-minimal-example-com.id]
}

resource "aws_autoscaling_group" "nodes-minimal-example-com" {
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupMaxSize", "GroupMinSize", "GroupPendingInstances", "GroupStandbyInstances", "GroupTerminatingInstances", "GroupTotalInstances"]
  launch_template {
    id      = aws_launch_template.nodes-minimal-example-com.id
    version = aws_launch_template.nodes-minimal-example-com.latest_version
  }
  max_instance_lifetime = 0
  max_size              = 2
  metrics_granularity   = "1Minute"
  min_size              = 2
  name                  = "nodes.minimal.example.com"
  protect_from_scale_in = false
  tag {
    key                 = "KubernetesCluster"
    propagate_at_launch = true
    value               = "minimal.example.com"
  }
  tag {
    key                 = "Name"
    propagate_at_launch = true
    value               = "nodes.minimal.example.com"
  }
  tag {
    key                 = "aws-node-termination-handler/managed"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"
    propagate_at_launch = true
    value               = ""
  }
  tag {
    key                 = "k8s.io/role/node"
    propagate_at_launch = true
    value               = "1"
  }
  tag {
    key                 = "kops.k8s.io/instancegroup"
    propagate_at_launch = true
    value               = "nodes"
  }
  tag {
    key                 = "kubernetes.io/cluster/minimal.example.com"
    propagate_at_launch = true
    value               = "owned"
  }
  vpc_zone_identifier = [aws_subnet.us-test-1a-minimal-example-com.id]
}

resource "aws_autoscaling_lifecycle_hook" "master-us-test-1a-NTHLifecycleHook" {
  autoscaling_group_name = aws_autoscaling_group.master-us-test-1a-masters-minimal-example-com.id
  default_result         = "CONTINUE"
  heartbeat_timeout      = 300
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_TERMINATING"
  name                   = "master-us-test-1a-NTHLifecycleHook"
}

resource "aws_autoscaling_lifecycle_hook" "nodes-NTHLifecycleHook" {
  autoscaling_group_name = aws_autoscaling_group.nodes-minimal-example-com.id
  default_result         = "CONTINUE"
  heartbeat_timeout      = 300
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_TERMINATING"
  name                   = "nodes-NTHLifecycleHook"
}

resource "aws_cloudwatch_event_rule" "minimal-example-com-ASGLifecycle" {
  event_pattern = file("${path.module}/data/aws_cloudwatch_event_rule_minimal.example.com-ASGLifecycle_event_pattern")
  name          = "minimal.example.com-ASGLifecycle"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com-ASGLifecycle"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_cloudwatch_event_rule" "minimal-example-com-InstanceScheduledChange" {
  event_pattern = file("${path.module}/data/aws_cloudwatch_event_rule_minimal.example.com-InstanceScheduledChange_event_pattern")
  name          = "minimal.example.com-InstanceScheduledChange"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com-InstanceScheduledChange"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_cloudwatch_event_rule" "minimal-example-com-InstanceStateChange" {
  event_pattern = file("${path.module}/data/aws_cloudwatch_event_rule_minimal.example.com-InstanceStateChange_event_pattern")
  name          = "minimal.example.com-InstanceStateChange"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com-InstanceStateChange"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_cloudwatch_event_rule" "minimal-example-com-SpotInterruption" {
  event_pattern = file("${path.module}/data/aws_cloudwatch_event_rule_minimal.example.com-SpotInterruption_event_pattern")
  name          = "minimal.example.com-SpotInterruption"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com-SpotInterruption"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_cloudwatch_event_target" "minimal-example-com-ASGLifecycle-Target" {
  arn  = aws_sqs_queue.minimal-example-com-nth.arn
  rule = aws_cloudwatch_event_rule.minimal-example-com-ASGLifecycle.id
}

resource "aws_cloudwatch_event_target" "minimal-example-com-InstanceScheduledChange-Target" {
  arn  = aws_sqs_queue.minimal-example-com-nth.arn
  rule = aws_cloudwatch_event_rule.minimal-example-com-InstanceScheduledChange.id
}

resource "aws_cloudwatch_event_target" "minimal-example-com-InstanceStateChange-Target" {
  arn  = aws_sqs_queue.minimal-example-com-nth.arn
  rule = aws_cloudwatch_event_rule.minimal-example-com-InstanceStateChange.id
}

resource "aws_cloudwatch_event_target" "minimal-example-com-SpotInterruption-Target" {
  arn  = aws_sqs_queue.minimal-example-com-nth.arn
  rule = aws_cloudwatch_event_rule.minimal-example-com-SpotInterruption.id
}

resource "aws_ebs_volume" "us-test-1a-etcd-events-minimal-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "us-test-1a.etcd-events.minimal.example.com"
    "k8s.io/etcd/events"                        = "us-test-1a/us-test-1a"
    "k8s.io/role/control-plane"                 = "1"
    "k8s.io/role/master"                        = "1"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_ebs_volume" "us-test-1a-etcd-main-minimal-example-com" {
  availability_zone = "us-test-1a"
  encrypted         = false
  iops              = 3000
  size              = 20
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "us-test-1a.etcd-main.minimal.example.com"
    "k8s.io/etcd/main"                          = "us-test-1a/us-test-1a"
    "k8s.io/role/control-plane"                 = "1"
    "k8s.io/role/master"                        = "1"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
  throughput = 125
  type       = "gp3"
}

resource "aws_iam_instance_profile" "masters-minimal-example-com" {
  name = "masters.minimal.example.com"
  role = aws_iam_role.masters-minimal-example-com.name
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "masters.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_iam_instance_profile" "nodes-minimal-example-com" {
  name = "nodes.minimal.example.com"
  role = aws_iam_role.nodes-minimal-example-com.name
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "nodes.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_iam_openid_connect_provider" "minimal-example-com" {
  client_id_list = ["amazonaws.com"]
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
  thumbprint_list = ["9e99a48a9960b14926bb7f3b02e22da2b0ab7280", "a9d53002e97e00e043244f3d170d6f4c414104fd"]
  url             = "https://discovery.example.com/minimal.example.com"
}

resource "aws_iam_role" "aws-cloud-controller-manager-kube-system-sa-minimal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_aws-cloud-controller-manager.kube-system.sa.minimal.example.com_policy")
  name               = "aws-cloud-controller-manager.kube-system.sa.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "aws-cloud-controller-manager.kube-system.sa.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
    "service-account.kops.k8s.io/name"          = "aws-cloud-controller-manager"
    "service-account.kops.k8s.io/namespace"     = "kube-system"
  }
}

resource "aws_iam_role" "aws-node-termination-handler-kube-system-sa-minimal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_aws-node-termination-handler.kube-system.sa.minimal.example.com_policy")
  name               = "aws-node-termination-handler.kube-system.sa.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "aws-node-termination-handler.kube-system.sa.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
    "service-account.kops.k8s.io/name"          = "aws-node-termination-handler"
    "service-account.kops.k8s.io/namespace"     = "kube-system"
  }
}

resource "aws_iam_role" "dns-controller-kube-system-sa-minimal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_dns-controller.kube-system.sa.minimal.example.com_policy")
  name               = "dns-controller.kube-system.sa.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "dns-controller.kube-system.sa.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
    "service-account.kops.k8s.io/name"          = "dns-controller"
    "service-account.kops.k8s.io/namespace"     = "kube-system"
  }
}

resource "aws_iam_role" "ebs-csi-controller-sa-kube-system-sa-minimal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_ebs-csi-controller-sa.kube-system.sa.minimal.example.com_policy")
  name               = "ebs-csi-controller-sa.kube-system.sa.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "ebs-csi-controller-sa.kube-system.sa.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
    "service-account.kops.k8s.io/name"          = "ebs-csi-controller-sa"
    "service-account.kops.k8s.io/namespace"     = "kube-system"
  }
}

resource "aws_iam_role" "karpenter-kube-system-sa-minimal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_karpenter.kube-system.sa.minimal.example.com_policy")
  name               = "karpenter.kube-system.sa.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "karpenter.kube-system.sa.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
    "service-account.kops.k8s.io/name"          = "karpenter"
    "service-account.kops.k8s.io/namespace"     = "kube-system"
  }
}

resource "aws_iam_role" "masters-minimal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_masters.minimal.example.com_policy")
  name               = "masters.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "masters.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_iam_role" "nodes-minimal-example-com" {
  assume_role_policy = file("${path.module}/data/aws_iam_role_nodes.minimal.example.com_policy")
  name               = "nodes.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "nodes.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_iam_role_policy" "aws-cloud-controller-manager-kube-system-sa-minimal-example-com" {
  name   = "aws-cloud-controller-manager.kube-system.sa.minimal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_aws-cloud-controller-manager.kube-system.sa.minimal.example.com_policy")
  role   = aws_iam_role.aws-cloud-controller-manager-kube-system-sa-minimal-example-com.name
}

resource "aws_iam_role_policy" "aws-node-termination-handler-kube-system-sa-minimal-example-com" {
  name   = "aws-node-termination-handler.kube-system.sa.minimal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_aws-node-termination-handler.kube-system.sa.minimal.example.com_policy")
  role   = aws_iam_role.aws-node-termination-handler-kube-system-sa-minimal-example-com.name
}

resource "aws_iam_role_policy" "dns-controller-kube-system-sa-minimal-example-com" {
  name   = "dns-controller.kube-system.sa.minimal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_dns-controller.kube-system.sa.minimal.example.com_policy")
  role   = aws_iam_role.dns-controller-kube-system-sa-minimal-example-com.name
}

resource "aws_iam_role_policy" "ebs-csi-controller-sa-kube-system-sa-minimal-example-com" {
  name   = "ebs-csi-controller-sa.kube-system.sa.minimal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_ebs-csi-controller-sa.kube-system.sa.minimal.example.com_policy")
  role   = aws_iam_role.ebs-csi-controller-sa-kube-system-sa-minimal-example-com.name
}

resource "aws_iam_role_policy" "karpenter-kube-system-sa-minimal-example-com" {
  name   = "karpenter.kube-system.sa.minimal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_karpenter.kube-system.sa.minimal.example.com_policy")
  role   = aws_iam_role.karpenter-kube-system-sa-minimal-example-com.name
}

resource "aws_iam_role_policy" "masters-minimal-example-com" {
  name   = "masters.minimal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_masters.minimal.example.com_policy")
  role   = aws_iam_role.masters-minimal-example-com.name
}

resource "aws_iam_role_policy" "nodes-minimal-example-com" {
  name   = "nodes.minimal.example.com"
  policy = file("${path.module}/data/aws_iam_role_policy_nodes.minimal.example.com_policy")
  role   = aws_iam_role.nodes-minimal-example-com.name
}

resource "aws_internet_gateway" "minimal-example-com" {
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
  vpc_id = aws_vpc.minimal-example-com.id
}

resource "aws_key_pair" "kubernetes-minimal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157" {
  key_name   = "kubernetes.minimal.example.com-c4:a6:ed:9a:a8:89:b9:e2:c3:9c:d6:63:eb:9c:71:57"
  public_key = file("${path.module}/data/aws_key_pair_kubernetes.minimal.example.com-c4a6ed9aa889b9e2c39cd663eb9c7157_public_key")
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_launch_template" "karpenter-nodes-default-minimal-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 128
      volume_type           = "gp3"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes-minimal-example-com.id
  }
  image_id = "ami-12345678"
  key_name = aws_key_pair.kubernetes-minimal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "karpenter-nodes-default.minimal.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.nodes-minimal-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                           = "minimal.example.com"
      "Name"                                                                        = "karpenter-nodes-default.minimal.example.com"
      "aws-node-termination-handler/managed"                                        = ""
      "k8s.io/cluster-autoscaler/node-template/label/foo"                           = "baz"
      "k8s.io/cluster-autoscaler/node-template/label/karpenter.sh/provisioner-name" = "karpenter-nodes-default"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"  = ""
      "k8s.io/role/node"                                                            = "1"
      "kops.k8s.io/instancegroup"                                                   = "karpenter-nodes-default"
      "kubernetes.io/cluster/minimal.example.com"                                   = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                           = "minimal.example.com"
      "Name"                                                                        = "karpenter-nodes-default.minimal.example.com"
      "aws-node-termination-handler/managed"                                        = ""
      "k8s.io/cluster-autoscaler/node-template/label/foo"                           = "baz"
      "k8s.io/cluster-autoscaler/node-template/label/karpenter.sh/provisioner-name" = "karpenter-nodes-default"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"  = ""
      "k8s.io/role/node"                                                            = "1"
      "kops.k8s.io/instancegroup"                                                   = "karpenter-nodes-default"
      "kubernetes.io/cluster/minimal.example.com"                                   = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                           = "minimal.example.com"
    "Name"                                                                        = "karpenter-nodes-default.minimal.example.com"
    "aws-node-termination-handler/managed"                                        = ""
    "k8s.io/cluster-autoscaler/node-template/label/foo"                           = "baz"
    "k8s.io/cluster-autoscaler/node-template/label/karpenter.sh/provisioner-name" = "karpenter-nodes-default"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"  = ""
    "k8s.io/role/node"                                                            = "1"
    "kops.k8s.io/instancegroup"                                                   = "karpenter-nodes-default"
    "kubernetes.io/cluster/minimal.example.com"                                   = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_karpenter-nodes-default.minimal.example.com_user_data")
}

resource "aws_launch_template" "karpenter-nodes-single-machinetype-minimal-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 4000
      kms_key_id            = "arn:aws:kms:us-east-1:012345678910:key/1234abcd-12ab-34cd-56ef-1234567890ab"
      throughput            = 200
      volume_size           = 200
      volume_type           = "gp3"
    }
  }
  block_device_mappings {
    device_name = "/dev/xvdd"
    ebs {
      delete_on_termination = true
      encrypted             = true
      kms_key_id            = "arn:aws:kms:us-east-1:012345678910:key/1234abcd-12ab-34cd-56ef-1234567890ab"
      volume_size           = 20
      volume_type           = "gp2"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes-minimal-example-com.id
  }
  image_id = "ami-12345678"
  key_name = aws_key_pair.kubernetes-minimal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "karpenter-nodes-single-machinetype.minimal.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.nodes-minimal-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                           = "minimal.example.com"
      "Name"                                                                        = "karpenter-nodes-single-machinetype.minimal.example.com"
      "aws-node-termination-handler/managed"                                        = ""
      "k8s.io/cluster-autoscaler/node-template/label/foo"                           = "bar"
      "k8s.io/cluster-autoscaler/node-template/label/karpenter.sh/provisioner-name" = "karpenter-nodes-single-machinetype"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"  = ""
      "k8s.io/role/node"                                                            = "1"
      "kops.k8s.io/instancegroup"                                                   = "karpenter-nodes-single-machinetype"
      "kubernetes.io/cluster/minimal.example.com"                                   = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                           = "minimal.example.com"
      "Name"                                                                        = "karpenter-nodes-single-machinetype.minimal.example.com"
      "aws-node-termination-handler/managed"                                        = ""
      "k8s.io/cluster-autoscaler/node-template/label/foo"                           = "bar"
      "k8s.io/cluster-autoscaler/node-template/label/karpenter.sh/provisioner-name" = "karpenter-nodes-single-machinetype"
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"  = ""
      "k8s.io/role/node"                                                            = "1"
      "kops.k8s.io/instancegroup"                                                   = "karpenter-nodes-single-machinetype"
      "kubernetes.io/cluster/minimal.example.com"                                   = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                           = "minimal.example.com"
    "Name"                                                                        = "karpenter-nodes-single-machinetype.minimal.example.com"
    "aws-node-termination-handler/managed"                                        = ""
    "k8s.io/cluster-autoscaler/node-template/label/foo"                           = "bar"
    "k8s.io/cluster-autoscaler/node-template/label/karpenter.sh/provisioner-name" = "karpenter-nodes-single-machinetype"
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node"  = ""
    "k8s.io/role/node"                                                            = "1"
    "kops.k8s.io/instancegroup"                                                   = "karpenter-nodes-single-machinetype"
    "kubernetes.io/cluster/minimal.example.com"                                   = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_karpenter-nodes-single-machinetype.minimal.example.com_user_data")
}

resource "aws_launch_template" "master-us-test-1a-masters-minimal-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 64
      volume_type           = "gp3"
    }
  }
  block_device_mappings {
    device_name  = "/dev/sdc"
    virtual_name = "ephemeral0"
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.masters-minimal-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "m3.medium"
  key_name      = aws_key_pair.kubernetes-minimal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "master-us-test-1a.masters.minimal.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.masters-minimal-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                                                     = "minimal.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.minimal.example.com"
      "aws-node-termination-handler/managed"                                                                  = ""
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/control-plane"                                                                             = "1"
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/minimal.example.com"                                                             = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                                                     = "minimal.example.com"
      "Name"                                                                                                  = "master-us-test-1a.masters.minimal.example.com"
      "aws-node-termination-handler/managed"                                                                  = ""
      "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
      "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
      "k8s.io/role/control-plane"                                                                             = "1"
      "k8s.io/role/master"                                                                                    = "1"
      "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
      "kubernetes.io/cluster/minimal.example.com"                                                             = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                                                     = "minimal.example.com"
    "Name"                                                                                                  = "master-us-test-1a.masters.minimal.example.com"
    "aws-node-termination-handler/managed"                                                                  = ""
    "k8s.io/cluster-autoscaler/node-template/label/kops.k8s.io/kops-controller-pki"                         = ""
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/control-plane"                   = ""
    "k8s.io/cluster-autoscaler/node-template/label/node.kubernetes.io/exclude-from-external-load-balancers" = ""
    "k8s.io/role/control-plane"                                                                             = "1"
    "k8s.io/role/master"                                                                                    = "1"
    "kops.k8s.io/instancegroup"                                                                             = "master-us-test-1a"
    "kubernetes.io/cluster/minimal.example.com"                                                             = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_master-us-test-1a.masters.minimal.example.com_user_data")
}

resource "aws_launch_template" "nodes-minimal-example-com" {
  block_device_mappings {
    device_name = "/dev/xvda"
    ebs {
      delete_on_termination = true
      encrypted             = true
      iops                  = 3000
      throughput            = 125
      volume_size           = 128
      volume_type           = "gp3"
    }
  }
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes-minimal-example-com.id
  }
  image_id      = "ami-12345678"
  instance_type = "t2.medium"
  key_name      = aws_key_pair.kubernetes-minimal-example-com-c4a6ed9aa889b9e2c39cd663eb9c7157.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_protocol_ipv6          = "disabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = false
  }
  name = "nodes.minimal.example.com"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    ipv6_address_count          = 0
    security_groups             = [aws_security_group.nodes-minimal-example-com.id]
  }
  tag_specifications {
    resource_type = "instance"
    tags = {
      "KubernetesCluster"                                                          = "minimal.example.com"
      "Name"                                                                       = "nodes.minimal.example.com"
      "aws-node-termination-handler/managed"                                       = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/minimal.example.com"                                  = "owned"
    }
  }
  tag_specifications {
    resource_type = "volume"
    tags = {
      "KubernetesCluster"                                                          = "minimal.example.com"
      "Name"                                                                       = "nodes.minimal.example.com"
      "aws-node-termination-handler/managed"                                       = ""
      "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
      "k8s.io/role/node"                                                           = "1"
      "kops.k8s.io/instancegroup"                                                  = "nodes"
      "kubernetes.io/cluster/minimal.example.com"                                  = "owned"
    }
  }
  tags = {
    "KubernetesCluster"                                                          = "minimal.example.com"
    "Name"                                                                       = "nodes.minimal.example.com"
    "aws-node-termination-handler/managed"                                       = ""
    "k8s.io/cluster-autoscaler/node-template/label/node-role.kubernetes.io/node" = ""
    "k8s.io/role/node"                                                           = "1"
    "kops.k8s.io/instancegroup"                                                  = "nodes"
    "kubernetes.io/cluster/minimal.example.com"                                  = "owned"
  }
  user_data = filebase64("${path.module}/data/aws_launch_template_nodes.minimal.example.com_user_data")
}

resource "aws_route" "route-0-0-0-0--0" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.minimal-example-com.id
  route_table_id         = aws_route_table.minimal-example-com.id
}

resource "aws_route" "route-__--0" {
  destination_ipv6_cidr_block = "::/0"
  gateway_id                  = aws_internet_gateway.minimal-example-com.id
  route_table_id              = aws_route_table.minimal-example-com.id
}

resource "aws_route_table" "minimal-example-com" {
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
    "kubernetes.io/kops/role"                   = "public"
  }
  vpc_id = aws_vpc.minimal-example-com.id
}

resource "aws_route_table_association" "us-test-1a-minimal-example-com" {
  route_table_id = aws_route_table.minimal-example-com.id
  subnet_id      = aws_subnet.us-test-1a-minimal-example-com.id
}

resource "aws_s3_object" "cluster-completed-spec" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_cluster-completed.spec_content")
  key                    = "clusters.example.com/minimal.example.com/cluster-completed.spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "discovery-json" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_discovery.json_content")
  key                    = "discovery.example.com/minimal.example.com/.well-known/openid-configuration"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-events" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-events_content")
  key                    = "clusters.example.com/minimal.example.com/backups/etcd/events/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "etcd-cluster-spec-main" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_etcd-cluster-spec-main_content")
  key                    = "clusters.example.com/minimal.example.com/backups/etcd/main/control/etcd-cluster-spec"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "keys-json" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_keys.json_content")
  key                    = "discovery.example.com/minimal.example.com/openid/v1/jwks"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "kops-version-txt" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_kops-version.txt_content")
  key                    = "clusters.example.com/minimal.example.com/kops-version.txt"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-events-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-events-master-us-test-1a_content")
  key                    = "clusters.example.com/minimal.example.com/manifests/etcd/events-master-us-test-1a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-etcdmanager-main-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-etcdmanager-main-master-us-test-1a_content")
  key                    = "clusters.example.com/minimal.example.com/manifests/etcd/main-master-us-test-1a.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "manifests-static-kube-apiserver-healthcheck" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_manifests-static-kube-apiserver-healthcheck_content")
  key                    = "clusters.example.com/minimal.example.com/manifests/static/kube-apiserver-healthcheck.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-aws-cloud-controller-addons-k8s-io-k8s-1-18" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-aws-cloud-controller.addons.k8s.io-k8s-1.18_content")
  key                    = "clusters.example.com/minimal.example.com/addons/aws-cloud-controller.addons.k8s.io/k8s-1.18.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-aws-ebs-csi-driver-addons-k8s-io-k8s-1-17" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-aws-ebs-csi-driver.addons.k8s.io-k8s-1.17_content")
  key                    = "clusters.example.com/minimal.example.com/addons/aws-ebs-csi-driver.addons.k8s.io/k8s-1.17.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-bootstrap" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-bootstrap_content")
  key                    = "clusters.example.com/minimal.example.com/addons/bootstrap-channel.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-coredns-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-coredns.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/minimal.example.com/addons/coredns.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-dns-controller-addons-k8s-io-k8s-1-12" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-dns-controller.addons.k8s.io-k8s-1.12_content")
  key                    = "clusters.example.com/minimal.example.com/addons/dns-controller.addons.k8s.io/k8s-1.12.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-karpenter-sh-k8s-1-19" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-karpenter.sh-k8s-1.19_content")
  key                    = "clusters.example.com/minimal.example.com/addons/karpenter.sh/k8s-1.19.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-kops-controller-addons-k8s-io-k8s-1-16" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-kops-controller.addons.k8s.io-k8s-1.16_content")
  key                    = "clusters.example.com/minimal.example.com/addons/kops-controller.addons.k8s.io/k8s-1.16.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-kubelet-api-rbac-addons-k8s-io-k8s-1-9" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-kubelet-api.rbac.addons.k8s.io-k8s-1.9_content")
  key                    = "clusters.example.com/minimal.example.com/addons/kubelet-api.rbac.addons.k8s.io/k8s-1.9.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-leader-migration-rbac-addons-k8s-io-k8s-1-23" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-leader-migration.rbac.addons.k8s.io-k8s-1.23_content")
  key                    = "clusters.example.com/minimal.example.com/addons/leader-migration.rbac.addons.k8s.io/k8s-1.23.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-limit-range-addons-k8s-io" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-limit-range.addons.k8s.io_content")
  key                    = "clusters.example.com/minimal.example.com/addons/limit-range.addons.k8s.io/v1.5.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-node-termination-handler-aws-k8s-1-11" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-node-termination-handler.aws-k8s-1.11_content")
  key                    = "clusters.example.com/minimal.example.com/addons/node-termination-handler.aws/k8s-1.11.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "minimal-example-com-addons-storage-aws-addons-k8s-io-v1-15-0" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_minimal.example.com-addons-storage-aws.addons.k8s.io-v1.15.0_content")
  key                    = "clusters.example.com/minimal.example.com/addons/storage-aws.addons.k8s.io/v1.15.0.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-karpenter-nodes-default" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-karpenter-nodes-default_content")
  key                    = "clusters.example.com/minimal.example.com/igconfig/node/karpenter-nodes-default/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-karpenter-nodes-single-machinetype" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-karpenter-nodes-single-machinetype_content")
  key                    = "clusters.example.com/minimal.example.com/igconfig/node/karpenter-nodes-single-machinetype/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-master-us-test-1a" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-master-us-test-1a_content")
  key                    = "clusters.example.com/minimal.example.com/igconfig/control-plane/master-us-test-1a/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_s3_object" "nodeupconfig-nodes" {
  bucket                 = "testingBucket"
  content                = file("${path.module}/data/aws_s3_object_nodeupconfig-nodes_content")
  key                    = "clusters.example.com/minimal.example.com/igconfig/node/nodes/nodeupconfig.yaml"
  provider               = aws.files
  server_side_encryption = "AES256"
}

resource "aws_security_group" "masters-minimal-example-com" {
  description = "Security group for masters"
  name        = "masters.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "masters.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
  vpc_id = aws_vpc.minimal-example-com.id
}

resource "aws_security_group" "nodes-minimal-example-com" {
  description = "Security group for nodes"
  name        = "nodes.minimal.example.com"
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "nodes.minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
  vpc_id = aws_vpc.minimal-example-com.id
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-22to22-masters-minimal-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-minimal-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-22to22-nodes-minimal-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 22
  protocol          = "tcp"
  security_group_id = aws_security_group.nodes-minimal-example-com.id
  to_port           = 22
  type              = "ingress"
}

resource "aws_security_group_rule" "from-0-0-0-0--0-ingress-tcp-443to443-masters-minimal-example-com" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.masters-minimal-example-com.id
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group_rule" "from-masters-minimal-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.masters-minimal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-minimal-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.masters-minimal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-masters-minimal-example-com-ingress-all-0to0-masters-minimal-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.masters-minimal-example-com.id
  source_security_group_id = aws_security_group.masters-minimal-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-masters-minimal-example-com-ingress-all-0to0-nodes-minimal-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-minimal-example-com.id
  source_security_group_id = aws_security_group.masters-minimal-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-minimal-example-com-egress-all-0to0-0-0-0-0--0" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = 0
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-minimal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-minimal-example-com-egress-all-0to0-__--0" {
  from_port         = 0
  ipv6_cidr_blocks  = ["::/0"]
  protocol          = "-1"
  security_group_id = aws_security_group.nodes-minimal-example-com.id
  to_port           = 0
  type              = "egress"
}

resource "aws_security_group_rule" "from-nodes-minimal-example-com-ingress-all-0to0-nodes-minimal-example-com" {
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.nodes-minimal-example-com.id
  source_security_group_id = aws_security_group.nodes-minimal-example-com.id
  to_port                  = 0
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-minimal-example-com-ingress-tcp-1to2379-masters-minimal-example-com" {
  from_port                = 1
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-minimal-example-com.id
  source_security_group_id = aws_security_group.nodes-minimal-example-com.id
  to_port                  = 2379
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-minimal-example-com-ingress-tcp-2382to4000-masters-minimal-example-com" {
  from_port                = 2382
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-minimal-example-com.id
  source_security_group_id = aws_security_group.nodes-minimal-example-com.id
  to_port                  = 4000
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-minimal-example-com-ingress-tcp-4003to65535-masters-minimal-example-com" {
  from_port                = 4003
  protocol                 = "tcp"
  security_group_id        = aws_security_group.masters-minimal-example-com.id
  source_security_group_id = aws_security_group.nodes-minimal-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "from-nodes-minimal-example-com-ingress-udp-1to65535-masters-minimal-example-com" {
  from_port                = 1
  protocol                 = "udp"
  security_group_id        = aws_security_group.masters-minimal-example-com.id
  source_security_group_id = aws_security_group.nodes-minimal-example-com.id
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_sqs_queue" "minimal-example-com-nth" {
  message_retention_seconds = 300
  name                      = "minimal-example-com-nth"
  policy                    = file("${path.module}/data/aws_sqs_queue_minimal-example-com-nth_policy")
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal-example-com-nth"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_subnet" "us-test-1a-minimal-example-com" {
  availability_zone                           = "us-test-1a"
  cidr_block                                  = "172.20.32.0/19"
  enable_resource_name_dns_a_record_on_launch = true
  private_dns_hostname_type_on_launch         = "resource-name"
  tags = {
    "KubernetesCluster"                                             = "minimal.example.com"
    "Name"                                                          = "us-test-1a.minimal.example.com"
    "SubnetType"                                                    = "Public"
    "kops.k8s.io/instance-group/karpenter-nodes-default"            = "true"
    "kops.k8s.io/instance-group/karpenter-nodes-single-machinetype" = "true"
    "kubernetes.io/cluster/minimal.example.com"                     = "owned"
    "kubernetes.io/role/elb"                                        = "1"
    "kubernetes.io/role/internal-elb"                               = "1"
  }
  vpc_id = aws_vpc.minimal-example-com.id
}

resource "aws_vpc" "minimal-example-com" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "172.20.0.0/16"
  enable_dns_hostnames             = true
  enable_dns_support               = true
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options" "minimal-example-com" {
  domain_name         = "us-test-1.compute.internal"
  domain_name_servers = ["AmazonProvidedDNS"]
  tags = {
    "KubernetesCluster"                         = "minimal.example.com"
    "Name"                                      = "minimal.example.com"
    "kubernetes.io/cluster/minimal.example.com" = "owned"
  }
}

resource "aws_vpc_dhcp_options_association" "minimal-example-com" {
  dhcp_options_id = aws_vpc_dhcp_options.minimal-example-com.id
  vpc_id          = aws_vpc.minimal-example-com.id
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 5.0.0"
    }
  }
}
