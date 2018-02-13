load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "position.go",
        "token.go",
    ],
    importpath = "github.com/hashicorp/hcl/json/token",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/hashicorp/hcl/hcl/token:go_default_library"],
)
