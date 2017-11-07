load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["dns.go"],
    importpath = "github.com/gophercloud/gophercloud/acceptance/openstack/dns/v2",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/acceptance/tools:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/dns/v2/recordsets:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/dns/v2/zones:go_default_library",
    ],
)
