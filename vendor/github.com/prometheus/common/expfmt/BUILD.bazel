load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "decode.go",
        "encode.go",
        "expfmt.go",
        "text_create.go",
        "text_parse.go",
    ],
    importpath = "github.com/prometheus/common/expfmt",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/matttproud/golang_protobuf_extensions/pbutil:go_default_library",
        "//vendor/github.com/prometheus/client_model/go:go_default_library",
        "//vendor/github.com/prometheus/common/internal/bitbucket.org/ww/goautoneg:go_default_library",
        "//vendor/github.com/prometheus/common/model:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "bench_test.go",
        "decode_test.go",
        "text_create_test.go",
        "text_parse_test.go",
    ],
    data = glob(["testdata/**"]),
    importpath = "github.com/prometheus/common/expfmt",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/matttproud/golang_protobuf_extensions/pbutil:go_default_library",
        "//vendor/github.com/prometheus/client_model/go:go_default_library",
        "//vendor/github.com/prometheus/common/model:go_default_library",
    ],
)
