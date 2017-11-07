load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "file_flag.go",
        "main.go",
    ],
    importpath = "github.com/evanphx/json-patch/cmd/json-patch",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/evanphx/json-patch:go_default_library",
        "//vendor/github.com/jessevdk/go-flags:go_default_library",
    ],
)

go_binary(
    name = "json-patch",
    importpath = "github.com/evanphx/json-patch/cmd/json-patch",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
