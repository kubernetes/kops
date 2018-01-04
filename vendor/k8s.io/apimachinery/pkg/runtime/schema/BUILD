load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_test(
    name = "go_default_test",
    srcs = ["group_version_test.go"],
    embed = [":go_default_library"],
    importpath = "k8s.io/apimachinery/pkg/runtime/schema",
)

go_library(
    name = "go_default_library",
    srcs = [
        "group_version.go",
        "interfaces.go",
    ],
    embed = [":schema_go_proto"],
    importpath = "k8s.io/apimachinery/pkg/runtime/schema",
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
    name = "schema_proto",
    srcs = ["generated.proto"],
    deps = ["//k8s.io/apimachinery/pkg/util/intstr:intstr_proto"],
)

go_proto_library(
    name = "schema_go_proto",
    importpath = "k8s.io/apimachinery/pkg/runtime/schema",
    proto = ":schema_proto",
    deps = ["//vendor/k8s.io/apimachinery/pkg/util/intstr:go_default_library"],
)
