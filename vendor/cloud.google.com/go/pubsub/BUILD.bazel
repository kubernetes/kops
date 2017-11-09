load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "acker.go",
        "doc.go",
        "iterator.go",
        "keepalive.go",
        "message.go",
        "pubsub.go",
        "puller.go",
        "service.go",
        "subscription.go",
        "topic.go",
    ],
    importpath = "cloud.google.com/go/pubsub",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/googleapi:go_default_library",
        "//vendor/google.golang.org/api/iterator:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/pubsub/v1:go_default_library",
        "//vendor/google.golang.org/api/transport:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "acker_test.go",
        "endtoend_test.go",
        "integration_test.go",
        "iterator_test.go",
        "keepalive_test.go",
        "puller_test.go",
        "subscription_test.go",
        "topic_test.go",
        "utils_test.go",
    ],
    importpath = "cloud.google.com/go/pubsub",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "example_subscription_iterator_test.go",
        "example_test.go",
        "example_topic_iterator_test.go",
    ],
    importpath = "cloud.google.com/go/pubsub_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
