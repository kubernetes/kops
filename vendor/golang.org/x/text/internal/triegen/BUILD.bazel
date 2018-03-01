load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "compact.go",
        "print.go",
        "triegen.go",
    ],
    importpath = "golang.org/x/text/internal/triegen",
    visibility = ["//vendor/golang.org/x/text:__subpackages__"],
)
