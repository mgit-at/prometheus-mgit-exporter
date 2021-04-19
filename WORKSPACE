workspace(name = "at_mgit_prometheus_mgit_exporter")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "rules_mgit",
    sha256 = "f23905d2273994f4a8230e1cbfa585c100024d716f44d09eabb34ed603504758",
    strip_prefix = "rules_mgit-08348150e4a04506db22aec2c5c5e96d6eaf2242",
    type = "zip",
    urls = ["https://github.com/mgit-at/rules_mgit/archive/08348150e4a04506db22aec2c5c5e96d6eaf2242.zip"],
)

load("@rules_mgit//:deps.bzl", "rules_mgit_dependencies")

rules_mgit_dependencies()

load("@rules_mgit//:setup.bzl", "rules_mgit_setup")

rules_mgit_setup()

load("//:go_deps.bzl", "go_repositories")

# gazelle:repository_macro go_deps.bzl%go_repositories
go_repositories()
