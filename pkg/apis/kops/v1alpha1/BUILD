load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "bastion.go",
        "cluster.go",
        "componentconfig.go",
        "conversion.go",
        "defaults.go",
        "doc.go",
        "dockerconfig.go",
        "federation.go",
        "instancegroup.go",
        "networking.go",
        "register.go",
        "topology.go",
        "zz_generated.conversion.go",
        "zz_generated.defaults.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_apimachinery//pkg/conversion:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime/schema:go_default_library",
    ],
)
