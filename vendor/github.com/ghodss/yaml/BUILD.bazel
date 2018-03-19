load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "fields.go",
        "yaml.go",
    ],
    importpath = "github.com/ghodss/yaml",
    visibility = ["//visibility:public"],
    deps = ["//vendor/gopkg.in/yaml.v2:go_default_library"],
)
