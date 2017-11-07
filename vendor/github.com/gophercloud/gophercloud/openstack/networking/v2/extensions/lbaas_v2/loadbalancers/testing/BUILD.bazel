load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "fixtures.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers/testing",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/listeners:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/monitors:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/pools:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper/client:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["requests_test.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers/testing",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/common:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/lbaas_v2/loadbalancers:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/pagination:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
    ],
)
