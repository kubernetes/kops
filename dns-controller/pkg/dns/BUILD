package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = [
        "dnscontext.go",
        "dnscontroller.go",
        "record.go",
        "zonespec.go",
    ],
    deps = [
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider/rrstype:go_default_library",
        "//dns-controller/pkg/util:go_default_library",
    ],
)
