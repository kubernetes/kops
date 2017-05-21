load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "api_loadbalancer.go",
        "autoscalinggroup.go",
        "context.go",
        "convenience.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/model:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/cloudup/awstasks:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/util/sets:go_default_library",
    ],
)
