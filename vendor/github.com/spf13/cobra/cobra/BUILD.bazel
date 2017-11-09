load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/spf13/cobra/cobra",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/spf13/cobra/cobra/cmd:go_default_library"],
)

go_binary(
    name = "cobra",
    importpath = "github.com/spf13/cobra/cobra",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
