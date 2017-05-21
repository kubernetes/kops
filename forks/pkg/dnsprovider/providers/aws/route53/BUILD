package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = [
        "interface.go",
        "route53.go",
        "rrchangeset.go",
        "rrset.go",
        "rrsets.go",
        "zone.go",
        "zones.go",
    ],
    tags = ["automanaged"],
    deps = [
        "//forks/pkg/dnsprovider:go_default_library",
        "//forks/pkg/dnsprovider/providers/aws/route53/stubs:go_default_library",
        "//forks/pkg/dnsprovider/rrstype:go_default_library",
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/request:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/session:go_default_library",
        "@com_github_aws_aws_sdk_go//service/route53:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/util/uuid:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["route53_test.go"],
    library = ":go_default_library",
    tags = ["automanaged"],
    deps = [
        "//forks/pkg/dnsprovider:go_default_library",
        "//forks/pkg/dnsprovider/providers/aws/route53/stubs:go_default_library",
        "//forks/pkg/dnsprovider/rrstype:go_default_library",
        "//forks/pkg/dnsprovider/tests:go_default_library",
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//service/route53:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//federation/pkg/dnsprovider/providers/aws/route53/stubs:all-srcs",
    ],
    tags = ["automanaged"],
)
