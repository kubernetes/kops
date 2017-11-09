load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["choose_version.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/utils",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/gophercloud/gophercloud:go_default_library"],
)
