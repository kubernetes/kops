load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "any.go",
        "doc.go",
        "duration.go",
        "timestamp.go",
    ],
    importpath = "github.com/golang/protobuf/ptypes",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/any:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/duration:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "any_test.go",
        "duration_test.go",
        "timestamp_test.go",
    ],
    importpath = "github.com/golang/protobuf/ptypes",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/descriptor:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/any:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/duration:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
    ],
)
