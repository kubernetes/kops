load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "pkg.go",
        "securityservices.go",
        "sharenetworks.go",
        "shares.go",
        "sharetypes.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/acceptance/openstack/sharedfilesystems/v2",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/acceptance/clients:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/acceptance/tools:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/securityservices:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharenetworks:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharetypes:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "availabilityzones_test.go",
        "securityservices_test.go",
        "sharenetworks_test.go",
        "shares_test.go",
        "sharetypes_test.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/acceptance/openstack/sharedfilesystems/v2",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/gophercloud/gophercloud/acceptance/clients:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/availabilityzones:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/securityservices:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharenetworks:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharetypes:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/pagination:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/testhelper:go_default_library",
    ],
)
