package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_binary",
)

go_binary(
    name = "dns-controller",
    srcs = [
        "main.go",
    ],
    deps = [
        "@com_github_golang_glog//:go_default_library",
        "@com_github_spf13_pflag//:go_default_library",
        "//dns-controller/pkg/dns:go_default_library",
        "//dns-controller/pkg/watchers:go_default_library",
    ],
)
