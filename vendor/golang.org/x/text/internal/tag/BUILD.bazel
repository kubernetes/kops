load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["tag.go"],
    importpath = "golang.org/x/text/internal/tag",
    visibility = ["//vendor/golang.org/x/text:__subpackages__"],
)

go_test(
    name = "go_default_test",
    srcs = ["tag_test.go"],
    importpath = "golang.org/x/text/internal/tag",
    library = ":go_default_library",
)
