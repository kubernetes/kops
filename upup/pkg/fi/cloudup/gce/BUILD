load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "gce_apitarget.go",
        "gce_cloud.go",
        "gce_url.go",
        "labels.go",
        "op.go",
        "utils.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//upup/pkg/fi:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/util/wait:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider:go_default_library",
        "@io_k8s_kubernetes//federation/pkg/dnsprovider/providers/google/clouddns:go_default_library",
        "@org_golang_google_api//compute/v0.beta:go_default_library",
        "@org_golang_google_api//googleapi:go_default_library",
        "@org_golang_google_api//storage/v1:go_default_library",
        "@org_golang_x_net//context:go_default_library",
        "@org_golang_x_oauth2//google:go_default_library",
    ],
)
