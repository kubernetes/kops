load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["integer.go"],
    importpath = "k8s.io/client-go/util/integer",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["integer_test.go"],
    importpath = "k8s.io/client-go/util/integer",
    library = ":go_default_library",
)
