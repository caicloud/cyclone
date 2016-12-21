package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = [
        "cronjob_controller.go",
        "doc.go",
        "injection.go",
        "utils.go",
    ],
    tags = ["automanaged"],
    deps = [
        "//pkg/api:go_default_library",
        "//pkg/api/errors:go_default_library",
        "//pkg/api/v1:go_default_library",
        "//pkg/apis/batch/v2alpha1:go_default_library",
        "//pkg/apis/meta/v1:go_default_library",
        "//pkg/client/clientset_generated/clientset:go_default_library",
        "//pkg/client/clientset_generated/clientset/typed/core/v1:go_default_library",
        "//pkg/client/record:go_default_library",
        "//pkg/labels:go_default_library",
        "//pkg/runtime:go_default_library",
        "//pkg/runtime/schema:go_default_library",
        "//pkg/types:go_default_library",
        "//pkg/util/errors:go_default_library",
        "//pkg/util/metrics:go_default_library",
        "//pkg/util/runtime:go_default_library",
        "//pkg/util/wait:go_default_library",
        "//vendor:github.com/golang/glog",
        "//vendor:github.com/robfig/cron",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "cronjob_controller_test.go",
        "utils_test.go",
    ],
    library = "go_default_library",
    tags = ["automanaged"],
    deps = [
        "//pkg/api/v1:go_default_library",
        "//pkg/apis/batch/v2alpha1:go_default_library",
        "//pkg/apis/meta/v1:go_default_library",
        "//pkg/client/record:go_default_library",
        "//pkg/types:go_default_library",
    ],
)
