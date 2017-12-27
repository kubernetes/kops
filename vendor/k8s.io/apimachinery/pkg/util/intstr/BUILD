load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_test(
    name = "go_default_test",
    srcs = ["intstr_test.go"],
    embed = [":go_default_library"],
    importpath = "k8s.io/apimachinery/pkg/util/intstr",
    deps = ["//vendor/github.com/ghodss/yaml:go_default_library"],
)

go_library(
    name = "go_default_library",
    srcs = ["intstr.go"],
    embed = [":intstr_go_proto"],
    importpath = "k8s.io/apimachinery/pkg/util/intstr",
    deps = [
        "//vendor/github.com/go-openapi/spec:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/google/gofuzz:go_default_library",
        "//vendor/k8s.io/kube-openapi/pkg/common:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)

proto_library(
    name = "intstr_proto",
    srcs = ["generated.proto"],
)

go_proto_library(
    name = "intstr_go_proto",
    importpath = "k8s.io/apimachinery/pkg/util/intstr",
    proto = ":intstr_proto",
)
