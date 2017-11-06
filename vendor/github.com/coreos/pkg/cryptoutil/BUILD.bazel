load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["aes.go"],
    importpath = "github.com/coreos/pkg/cryptoutil",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["aes_test.go"],
    importpath = "github.com/coreos/pkg/cryptoutil",
    library = ":go_default_library",
)
