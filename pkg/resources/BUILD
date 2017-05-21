load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "cluster_resources.go",
        "delete_cluster.go",
        "delete_cluster_gce.go",
        "delete_cluster_vsphere.go",
        "utils.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/dns:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/cloudup/awsup:go_default_library",
        "//upup/pkg/fi/cloudup/gce:go_default_library",
        "//upup/pkg/fi/cloudup/vsphere:go_default_library",
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//service/autoscaling:go_default_library",
        "@com_github_aws_aws_sdk_go//service/cloudformation:go_default_library",
        "@com_github_aws_aws_sdk_go//service/ec2:go_default_library",
        "@com_github_aws_aws_sdk_go//service/elb:go_default_library",
        "@com_github_aws_aws_sdk_go//service/iam:go_default_library",
        "@com_github_aws_aws_sdk_go//service/route53:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@com_github_vmware_govmomi//find:go_default_library",
        "@com_github_vmware_govmomi//object:go_default_library",
        "@io_k8s_apimachinery//pkg/util/sets:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider:go_default_library",
        "@org_golang_google_api//compute/v0.beta:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["deletecluster_test.go"],
    library = ":go_default_library",
    deps = [
        "//cloudmock/aws/mockec2:go_default_library",
        "//upup/pkg/fi/cloudup/awsup:go_default_library",
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//service/ec2:go_default_library",
    ],
)
