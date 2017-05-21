load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_test")

go_binary(
    name = "channel",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["integration_test.go"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "@com_github_blang_semver//:go_default_library",
    ],
)
