load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "equals.go",
        "reflect.go",
        "sanitize.go",
        "yaml.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_ghodss_yaml//:go_default_library",
        "@com_github_golang_glog//:go_default_library",
    ],
)
