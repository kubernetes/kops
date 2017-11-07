load("@io_bazel_rules_go//go:def.bzl", "go_test")

go_test(
    name = "go_default_xtest",
    srcs = [
        "end2end_test.go",
        "servertester_test.go",
    ],
    data = glob(["testdata/**"]),
    importpath = "google.golang.org/grpc/test_test",
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/any:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/net/http2:go_default_library",
        "//vendor/golang.org/x/net/http2/hpack:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/rpc/status:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/credentials:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
        "//vendor/google.golang.org/grpc/health:go_default_library",
        "//vendor/google.golang.org/grpc/health/grpc_health_v1:go_default_library",
        "//vendor/google.golang.org/grpc/internal:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
        "//vendor/google.golang.org/grpc/peer:go_default_library",
        "//vendor/google.golang.org/grpc/status:go_default_library",
        "//vendor/google.golang.org/grpc/tap:go_default_library",
        "//vendor/google.golang.org/grpc/test/grpc_testing:go_default_library",
    ],
)
