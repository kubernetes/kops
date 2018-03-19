load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["scanner.go"],
    importpath = "github.com/hashicorp/hcl/json/scanner",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/hashicorp/hcl/json/token:go_default_library"],
)
