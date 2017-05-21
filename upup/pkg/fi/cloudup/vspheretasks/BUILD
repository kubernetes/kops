load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "attachiso.go",
        "cloud_init.go",
        "virtualmachine.go",
        "vmpoweron.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/model:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/cloudup/vsphere:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@com_github_pborman_uuid//:go_default_library",
    ],
)
