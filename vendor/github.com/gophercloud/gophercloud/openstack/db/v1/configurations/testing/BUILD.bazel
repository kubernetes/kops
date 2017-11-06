load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "fixtures.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/openstack/db/v1/configurations/testing",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/gophercloud/gophercloud/openstack/db/v1/configurations:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["requests_test.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/db/v1/configurations/testing",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/gophercloud/gophercloud/openstack/db/v1/configurations:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/db/v1/instances:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/pagination:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper/client:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper/fixture:go_default_library",
    ],
)
