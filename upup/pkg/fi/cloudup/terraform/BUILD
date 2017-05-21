load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "hcl_printer.go",
        "lifecycle.go",
        "literal.go",
        "target.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/featureflag:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@com_github_hashicorp_hcl//hcl/ast:go_default_library",
        "@com_github_hashicorp_hcl//hcl/printer:go_default_library",
        "@com_github_hashicorp_hcl//json/parser:go_default_library",
    ],
)
