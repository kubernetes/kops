load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["doc.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/diskconfig/testing",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["requests_test.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/diskconfig/testing",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/diskconfig:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/servers:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
    ],
)
