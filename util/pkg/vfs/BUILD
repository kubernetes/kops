load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "context.go",
        "fs.go",
        "gsfs.go",
        "memfs.go",
        "s3context.go",
        "s3fs.go",
        "sshfs.go",
        "vfs.go",
        "vfssync.go",
        "writeoption.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//util/pkg/hashing:go_default_library",
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/awserr:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/credentials:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/endpoints:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/session:go_default_library",
        "@com_github_aws_aws_sdk_go//service/ec2:go_default_library",
        "@com_github_aws_aws_sdk_go//service/s3:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@com_github_pkg_sftp//:go_default_library",
        "@io_k8s_apimachinery//pkg/util/wait:go_default_library",
        "@org_golang_google_api//googleapi:go_default_library",
        "@org_golang_google_api//storage/v1:go_default_library",
        "@org_golang_x_crypto//ssh:go_default_library",
        "@org_golang_x_net//context:go_default_library",
        "@org_golang_x_oauth2//google:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["s3fs_test.go"],
    library = ":go_default_library",
)
