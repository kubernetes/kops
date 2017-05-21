load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "aws_apitarget.go",
        "aws_cloud.go",
        "aws_utils.go",
        "logging_retryer.go",
        "machine_types.go",
        "mock_aws_cloud.go",
        "request_logger.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/awserr:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/client:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/request:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/session:go_default_library",
        "@com_github_aws_aws_sdk_go//service/autoscaling:go_default_library",
        "@com_github_aws_aws_sdk_go//service/autoscaling/autoscalingiface:go_default_library",
        "@com_github_aws_aws_sdk_go//service/cloudformation:go_default_library",
        "@com_github_aws_aws_sdk_go//service/ec2:go_default_library",
        "@com_github_aws_aws_sdk_go//service/ec2/ec2iface:go_default_library",
        "@com_github_aws_aws_sdk_go//service/elb:go_default_library",
        "@com_github_aws_aws_sdk_go//service/iam:go_default_library",
        "@com_github_aws_aws_sdk_go//service/route53:go_default_library",
        "@com_github_aws_aws_sdk_go//service/route53/route53iface:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider/providers/aws/route53:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "aws_topology_test.go",
        "aws_utils_test.go",
    ],
    library = ":go_default_library",
    deps = [
        "//pkg/apis/kops:go_default_library",
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//service/ec2:go_default_library",
    ],
)
