load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "config.go",
        "create_kubecfg.go",
        "kubecfg_builder.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/dns:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_client_go//rest:go_default_library",
        "@io_k8s_client_go//tools/clientcmd:go_default_library",
        "@io_k8s_client_go//tools/clientcmd/api:go_default_library",
    ],
)
