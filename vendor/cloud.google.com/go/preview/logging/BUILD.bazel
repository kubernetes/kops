load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "logging.go",
        "metrics.go",
        "resources.go",
        "sinks.go",
    ],
    importpath = "cloud.google.com/go/preview/logging",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/internal/bundler:go_default_library",
        "//vendor/cloud.google.com/go/logging/apiv2:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/struct:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/github.com/googleapis/gax-go:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/monitoredres:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/logging/type:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/logging/v2:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "logging_test.go",
        "metrics_test.go",
        "resources_test.go",
        "sinks_test.go",
        "unique_test.go",
    ],
    importpath = "cloud.google.com/go/preview/logging",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/bundler:go_default_library",
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/cloud.google.com/go/preview/logging/internal/testing:go_default_library",
        "//vendor/cloud.google.com/go/storage:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/struct:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/api/monitoredres:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/logging/type:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/logging/v2:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "example_entry_iterator_test.go",
        "example_metric_iterator_test.go",
        "example_paging_test.go",
        "example_resource_iterator_test.go",
        "example_sink_iterator_test.go",
        "examples_test.go",
    ],
    importpath = "cloud.google.com/go/preview/logging_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
    ],
)
