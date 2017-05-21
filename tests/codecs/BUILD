load("@io_bazel_rules_go//go:def.bzl", "go_test")

go_test(
    name = "go_default_test",
    srcs = ["componentconfig_test.go"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/install:go_default_library",
        "//pkg/apis/kops/v1alpha2:go_default_library",
        "//pkg/diff:go_default_library",
    ],
)
