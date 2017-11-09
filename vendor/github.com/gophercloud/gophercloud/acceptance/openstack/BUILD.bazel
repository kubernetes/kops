load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["common.go"],
    importpath = "github.com/gophercloud/gophercloud/acceptance/openstack",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/gophercloud/gophercloud/openstack/common/extensions:go_default_library"],
)
