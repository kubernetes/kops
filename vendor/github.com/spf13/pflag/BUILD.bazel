load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "bool.go",
        "bool_slice.go",
        "count.go",
        "duration.go",
        "flag.go",
        "float32.go",
        "float64.go",
        "golangflag.go",
        "int.go",
        "int32.go",
        "int64.go",
        "int8.go",
        "int_slice.go",
        "ip.go",
        "ip_slice.go",
        "ipmask.go",
        "ipnet.go",
        "string.go",
        "string_array.go",
        "string_slice.go",
        "uint.go",
        "uint16.go",
        "uint32.go",
        "uint64.go",
        "uint8.go",
        "uint_slice.go",
    ],
    importpath = "github.com/spf13/pflag",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "bool_slice_test.go",
        "bool_test.go",
        "count_test.go",
        "export_test.go",
        "flag_test.go",
        "golangflag_test.go",
        "int_slice_test.go",
        "ip_slice_test.go",
        "ip_test.go",
        "ipnet_test.go",
        "string_array_test.go",
        "string_slice_test.go",
        "uint_slice_test.go",
    ],
    importpath = "github.com/spf13/pflag",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "github.com/spf13/pflag_test",
    deps = [":go_default_library"],
)
