load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "codesize.go",
        "flag.go",
        "gc.go",
        "go1_6.go",
        "go1_7.go",
        "text.go",
    ],
    importpath = "golang.org/x/text/internal/testtext",
    visibility = ["//vendor/golang.org/x/text:__subpackages__"],
    deps = ["//vendor/golang.org/x/text/internal/gen:go_default_library"],
)
