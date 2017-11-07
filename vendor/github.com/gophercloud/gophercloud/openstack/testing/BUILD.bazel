load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["doc.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/testing",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "client_test.go",
        "endpoint_location_test.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/openstack/testing",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/identity/v2/tokens:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/identity/v3/tokens:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
    ],
)
