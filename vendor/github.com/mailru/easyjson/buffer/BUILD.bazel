load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["pool.go"],
    importpath = "github.com/mailru/easyjson/buffer",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["pool_test.go"],
    importpath = "github.com/mailru/easyjson/buffer",
    library = ":go_default_library",
)
