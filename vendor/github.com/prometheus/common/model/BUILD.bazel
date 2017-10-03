load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "alert.go",
        "fingerprinting.go",
        "fnv.go",
        "labels.go",
        "labelset.go",
        "metric.go",
        "model.go",
        "signature.go",
        "silence.go",
        "time.go",
        "value.go",
    ],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "alert_test.go",
        "labels_test.go",
        "metric_test.go",
        "signature_test.go",
        "silence_test.go",
        "time_test.go",
        "value_test.go",
    ],
    library = ":go_default_library",
)
