load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["metadata.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/internal:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/net/context/ctxhttp:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["metadata_test.go"],
    library = ":go_default_library",
)
