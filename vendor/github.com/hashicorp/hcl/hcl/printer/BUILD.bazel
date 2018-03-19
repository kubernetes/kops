load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "nodes.go",
        "printer.go",
    ],
    importpath = "github.com/hashicorp/hcl/hcl/printer",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/hashicorp/hcl/hcl/ast:go_default_library",
        "//vendor/github.com/hashicorp/hcl/hcl/parser:go_default_library",
        "//vendor/github.com/hashicorp/hcl/hcl/token:go_default_library",
    ],
)
