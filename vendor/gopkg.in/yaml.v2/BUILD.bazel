load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "apic.go",
        "decode.go",
        "emitterc.go",
        "encode.go",
        "parserc.go",
        "readerc.go",
        "resolve.go",
        "scannerc.go",
        "sorter.go",
        "writerc.go",
        "yaml.go",
        "yamlh.go",
        "yamlprivateh.go",
    ],
    importpath = "gopkg.in/yaml.v2",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "decode_test.go",
        "encode_test.go",
        "suite_test.go",
    ],
    importpath = "gopkg.in/yaml.v2_test",
    deps = [
        ":go_default_library",
        "//vendor/gopkg.in/check.v1:go_default_library",
    ],
)
