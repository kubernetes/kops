load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "apply.go",
        "apply_channel.go",
        "factory.go",
        "get.go",
        "get_addons.go",
        "root.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//channels/pkg/channels:go_default_library",
        "//util/pkg/tables:go_default_library",
        "@com_github_blang_semver//:go_default_library",
        "@com_github_spf13_cobra//:go_default_library",
        "@com_github_spf13_viper//:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_client_go//kubernetes:go_default_library",
        "@io_k8s_client_go//pkg/api/v1:go_default_library",
        "@io_k8s_client_go//tools/clientcmd:go_default_library",
    ],
)
