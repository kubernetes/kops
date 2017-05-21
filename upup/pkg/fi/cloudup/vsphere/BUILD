load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "vsphere_apitarget.go",
        "vsphere_cloud.go",
        "vsphere_utils.go",
        "vsphere_volume_metadata.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@com_github_pkg_errors//:go_default_library",
        "@com_github_vmware_govmomi//:go_default_library",
        "@com_github_vmware_govmomi//find:go_default_library",
        "@com_github_vmware_govmomi//object:go_default_library",
        "@com_github_vmware_govmomi//property:go_default_library",
        "@com_github_vmware_govmomi//vim25:go_default_library",
        "@com_github_vmware_govmomi//vim25/mo:go_default_library",
        "@com_github_vmware_govmomi//vim25/soap:go_default_library",
        "@com_github_vmware_govmomi//vim25/types:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider/providers/coredns:go_default_library",
    ],
)
