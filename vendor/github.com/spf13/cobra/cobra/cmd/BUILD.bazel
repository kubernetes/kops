load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "add.go",
        "helpers.go",
        "init.go",
        "licenses.go",
        "root.go",
    ],
    importpath = "github.com/spf13/cobra/cobra/cmd",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/spf13/cobra:go_default_library",
        "//vendor/github.com/spf13/viper:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["helpers_test.go"],
    importpath = "github.com/spf13/cobra/cobra/cmd",
    library = ":go_default_library",
)
