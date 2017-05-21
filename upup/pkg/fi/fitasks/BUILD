load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cert_utils.go",
        "keypair.go",
        "keypair_fitask.go",
        "managedfile.go",
        "managedfile_fitask.go",
        "secret.go",
        "secret_fitask.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//upup/pkg/fi:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/util/validation/field:go_default_library",
    ],
)
