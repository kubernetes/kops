load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["transform.go"],
    importpath = "golang.org/x/text/transform",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["transform_test.go"],
    importpath = "golang.org/x/text/transform",
    library = ":go_default_library",
    deps = ["//vendor/golang.org/x/text/internal/testtext:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["examples_test.go"],
    importpath = "golang.org/x/text/transform_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/text/unicode/norm:go_default_library",
    ],
)
