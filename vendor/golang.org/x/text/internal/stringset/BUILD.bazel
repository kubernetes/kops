load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["set.go"],
    importpath = "golang.org/x/text/internal/stringset",
    visibility = ["//vendor/golang.org/x/text:__subpackages__"],
)

go_test(
    name = "go_default_test",
    srcs = ["set_test.go"],
    importpath = "golang.org/x/text/internal/stringset",
    library = ":go_default_library",
)
