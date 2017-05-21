package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = [
        "annotations.go",
        "ingresscontroller.go",
        "nodecontroller.go",
        "podcontroller.go",
        "servicecontroller.go",
    ],
    deps = [
        "//dns-controller/pkg/dns:go_default_library",
        "//dns-controller/pkg/util:go_default_library",
        "//pkg/apis/kops/util:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_apimachinery//pkg/watch:go_default_library",
        "@io_k8s_client_go//kubernetes:go_default_library",
        "@io_k8s_client_go//pkg/api/v1:go_default_library",
        "@io_k8s_client_go//pkg/apis/extensions/v1beta1:go_default_library",
        "@io_k8s_kubernetes//pkg/api:go_default_library",
        "@io_k8s_kubernetes//pkg/api/v1:go_default_library",
        "@io_k8s_kubernetes//pkg/client/clientset_generated/release_1_3/typed/core/v1:go_default_library",
        "@io_k8s_kubernetes//pkg/fields:go_default_library",
        "@io_k8s_kubernetes//pkg/labels:go_default_library",
        "@io_k8s_kubernetes//pkg/watch:go_default_library",
    ],
)
