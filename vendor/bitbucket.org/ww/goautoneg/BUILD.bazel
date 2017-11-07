load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["autoneg.go"],
    importpath = "bitbucket.org/ww/goautoneg",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["autoneg_test.go"],
    importpath = "bitbucket.org/ww/goautoneg",
    library = ":go_default_library",
)
