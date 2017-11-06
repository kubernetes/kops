load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "fixtures.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external/testing",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["results_test.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external/testing",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/common:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/networks:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/networks/testing:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
    ],
)
