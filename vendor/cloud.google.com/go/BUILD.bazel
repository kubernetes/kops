load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["cloud.go"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["license_test.go"],
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["authexample_test.go"],
    deps = [
        "//vendor/cloud.google.com/go/datastore:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
    ],
)
