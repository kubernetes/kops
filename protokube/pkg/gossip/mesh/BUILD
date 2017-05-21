load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "glogger.go",
        "gossip.go",
        "mesh.pb.go",
        "peer.go",
        "state.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//protokube/pkg/gossip:go_default_library",
        "@com_github_gogo_protobuf//gogoproto:go_default_library",
        "@com_github_gogo_protobuf//proto:go_default_library",
        "@com_github_gogo_protobuf//sortkeys:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_weaveworks_mesh//:go_default_library",
    ],
)

filegroup(
    name = "go_default_library_protos",
    srcs = ["mesh.proto"],
    visibility = ["//visibility:public"],
)
