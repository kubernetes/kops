load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "collector.go",
        "counter.go",
        "desc.go",
        "doc.go",
        "expvar_collector.go",
        "fnv.go",
        "gauge.go",
        "go_collector.go",
        "histogram.go",
        "http.go",
        "metric.go",
        "observer.go",
        "process_collector.go",
        "registry.go",
        "summary.go",
        "timer.go",
        "untyped.go",
        "value.go",
        "vec.go",
    ],
    importpath = "github.com/prometheus/client_golang/prometheus",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/beorn7/perks/quantile:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/prometheus/client_model/go:go_default_library",
        "//vendor/github.com/prometheus/common/expfmt:go_default_library",
        "//vendor/github.com/prometheus/common/model:go_default_library",
        "//vendor/github.com/prometheus/procfs:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "benchmark_test.go",
        "counter_test.go",
        "gauge_test.go",
        "go_collector_test.go",
        "histogram_test.go",
        "http_test.go",
        "metric_test.go",
        "process_collector_test.go",
        "summary_test.go",
        "timer_test.go",
        "vec_test.go",
    ],
    importpath = "github.com/prometheus/client_golang/prometheus",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/prometheus/client_model/go:go_default_library",
        "//vendor/github.com/prometheus/common/expfmt:go_default_library",
        "//vendor/github.com/prometheus/procfs:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "example_clustermanager_test.go",
        "example_timer_complex_test.go",
        "example_timer_gauge_test.go",
        "example_timer_test.go",
        "examples_test.go",
        "expvar_collector_test.go",
        "registry_test.go",
    ],
    importpath = "github.com/prometheus/client_golang/prometheus_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus/promhttp:go_default_library",
        "//vendor/github.com/prometheus/client_model/go:go_default_library",
        "//vendor/github.com/prometheus/common/expfmt:go_default_library",
    ],
)
