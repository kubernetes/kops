load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["fixtures.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares/testing",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper/client:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["request_test.go"],
    importpath = "github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares/testing",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper/client:go_default_library",
    ],
)
