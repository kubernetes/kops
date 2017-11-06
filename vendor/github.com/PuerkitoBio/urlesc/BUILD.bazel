load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["urlesc.go"],
    importpath = "github.com/PuerkitoBio/urlesc",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["urlesc_test.go"],
    importpath = "github.com/PuerkitoBio/urlesc",
    library = ":go_default_library",
)
