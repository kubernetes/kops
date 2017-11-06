load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["transformer.go"],
    importpath = "k8s.io/apiserver/pkg/storage/value",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["transformer_test.go"],
    importpath = "k8s.io/apiserver/pkg/storage/value",
    library = ":go_default_library",
)
