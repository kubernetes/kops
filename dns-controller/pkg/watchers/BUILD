package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = [
        "annotations.go",
        "ingresscontroller.go",
        "nodecontroller.go",
        "podcontroller.go",
        "servicecontroller.go",
    ],
    deps = [
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_kubernetes//pkg/api:go_default_library",
    ],
)
