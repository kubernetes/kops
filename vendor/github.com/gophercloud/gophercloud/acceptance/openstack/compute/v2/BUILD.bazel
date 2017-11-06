load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "compute.go",
        "pkg.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/acceptance/openstack/compute/v2",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/acceptance/clients:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/acceptance/tools:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/blockstorage/v1/volumes:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/bootfromvolume:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/defsecrules:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/networks:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/quotasets:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/schedulerhints:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/secgroups:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/tenantnetworks:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/flavors:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/servers:go_default_library",
        "//vendor/golang.org/x/crypto/ssh:go_default_library",
    ],
)
