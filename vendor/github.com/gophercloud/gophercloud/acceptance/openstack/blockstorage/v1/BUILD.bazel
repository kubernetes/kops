load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "blockstorage.go",
        "pkg.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/acceptance/openstack/blockstorage/v1",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/acceptance/tools:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/blockstorage/v1/snapshots:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/blockstorage/v1/volumes:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/blockstorage/v1/volumetypes:go_default_library",
    ],
)
