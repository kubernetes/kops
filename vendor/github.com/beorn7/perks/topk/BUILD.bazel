load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["topk.go"],
    importpath = "github.com/beorn7/perks/topk",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["topk_test.go"],
    importpath = "github.com/beorn7/perks/topk",
    library = ":go_default_library",
)
