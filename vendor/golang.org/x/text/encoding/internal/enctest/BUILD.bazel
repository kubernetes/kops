load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["enctest.go"],
    importpath = "golang.org/x/text/encoding/internal/enctest",
    visibility = ["//vendor/golang.org/x/text/encoding:__subpackages__"],
    deps = [
        "//vendor/golang.org/x/text/encoding:go_default_library",
        "//vendor/golang.org/x/text/encoding/internal/identifier:go_default_library",
        "//vendor/golang.org/x/text/transform:go_default_library",
    ],
)
