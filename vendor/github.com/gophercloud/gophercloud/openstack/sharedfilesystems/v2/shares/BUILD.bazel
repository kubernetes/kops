load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "requests.go",
        "results.go",
        "urls.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/gophercloud/gophercloud:go_default_library"],
)
