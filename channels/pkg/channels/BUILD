load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "addon.go",
        "addons.go",
        "apply.go",
        "channel_version.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//channels/pkg/api:go_default_library",
        "//upup/pkg/fi/utils:go_default_library",
        "//util/pkg/vfs:go_default_library",
        "@com_github_blang_semver//:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_apimachinery//pkg/types:go_default_library",
        "@io_k8s_apimachinery//pkg/util/validation/field:go_default_library",
        "@io_k8s_client_go//kubernetes:go_default_library",
        "@io_k8s_client_go//pkg/api/v1:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["addons_test.go"],
    library = ":go_default_library",
    deps = [
        "//channels/pkg/api:go_default_library",
        "@com_github_blang_semver//:go_default_library",
    ],
)
