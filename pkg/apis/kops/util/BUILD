load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "labels.go",
        "versions.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_blang_semver//:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_client_go//pkg/api/v1:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["versions_test.go"],
    library = ":go_default_library",
)
