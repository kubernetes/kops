load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "0doc.go",
        "binc.go",
        "cbor.go",
        "decode.go",
        "decode_go.go",
        "decode_go14.go",
        "encode.go",
        "fast-path.generated.go",
        "gen.generated.go",
        "gen.go",
        "gen-helper.generated.go",
        "gen_15.go",
        "gen_16.go",
        "gen_17.go",
        "helper.go",
        "helper_internal.go",
        "helper_not_unsafe.go",
        "json.go",
        "msgpack.go",
        "noop.go",
        "prebuild.go",
        "rpc.go",
        "simple.go",
        "time.go",
    ],
    importpath = "github.com/ugorji/go/codec",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "cbor_test.go",
        "codec_test.go",
        "helper_test.go",
        "values_test.go",
    ],
    importpath = "github.com/ugorji/go/codec",
    library = ":go_default_library",
)
