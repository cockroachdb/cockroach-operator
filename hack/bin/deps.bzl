# Copyright 2019 The Jetstack cert-manager contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file", "http_archive")
load("@io_bazel_rules_docker//container:container.bzl", "container_pull")
load("@bazel_gazelle//:deps.bzl", "go_repository")

def install():
    install_misc()
    install_integration_test_dependencies()

    install_crdb()
    install_golangci_lint()
    install_kubectl()
    install_kind()
    install_kubetest2()
    install_kubetest2_aws()
    install_kubetest2_exe()
    install_kubetest2_gke()
    install_kubetest2_kind()
    install_kustomize()
    install_oc()
    install_operator_sdk()
    install_opm()
    install_openshift()

    # Install golang.org/x/build as kubernetes/repo-infra requires it for the
    # build-tar bazel target.
    go_repository(
        name = "org_golang_x_build",
        build_file_generation = "on",
        build_file_proto_mode = "disable",
        importpath = "golang.org/x/build",
        sum = "h1:hXVePvSFG7tPGX4Pwk1d10ePFfoTCc0QmISfpKOHsS8=",
        version = "v0.0.0-20190927031335-2835ba2e683f",
    )

def install_misc():
    http_file(
        name = "jq_linux",
        executable = 1,
        sha256 = "af986793a515d500ab2d35f8d2aecd656e764504b789b66d7e1a0b727a124c44",
        urls = ["https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64"],
    )

    http_file(
        name = "jq_osx",
        executable = 1,
        sha256 = "5c0a0a3ea600f302ee458b30317425dd9632d1ad8882259fcaf4e9b868b2b1ef",
        urls = ["https://github.com/stedolan/jq/releases/download/jq-1.6/jq-osx-amd64"],
    )

    http_file(
        #this tool is used on open shift generation to generate the csv
        name = "faq_linux",
        executable = 1,
        sha256 = "6c9234d0b2b024bf0e7c845fc092339b51b94e5addeee9612a7219cfd2a7b731",
        urls = ["https://github.com/jzelinskie/faq/releases/download/0.0.7/faq-linux-amd64"],
    )

    http_file(
        name = "faq_osx",
        executable = 1,
        sha256 = "869f4d8acaa1feb11ce76b2204c5476b8a04d9451216adde6b18e2ef2f978794",
        urls = ["https://github.com/jzelinskie/faq/releases/download/0.0.7/faq-darwin-amd64"],
    )

# Install dependencies used by the controller-runtime integration test framework
def install_integration_test_dependencies():
    http_file(
        name = "kube-apiserver_darwin_amd64",
        executable = 1,
        sha256 = "a874d479f183f9e4c19a5c69b44955fabd2e250b467d2d9f0641ae91a82ddbea",
        urls = ["https://storage.googleapis.com/cert-manager-testing-assets/kube-apiserver-1.17.3_darwin_amd64"],
    )

    http_file(
        name = "kube-apiserver_linux_amd64",
        executable = 1,
        sha256 = "b4505b838b27b170531afbdef5e7bfaacf83da665f21b0e3269d1775b0defb7a",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.17.3/bin/linux/amd64/kube-apiserver"],
    )

    http_archive(
        name = "com_coreos_etcd_darwin_amd64",
        sha256 = "27245adea2e0951913276d8c321d79b91caaf904ae3fdaab65194ab41c01db08",
        urls = ["https://github.com/etcd-io/etcd/releases/download/v3.4.16/etcd-v3.4.16-darwin-amd64.zip"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "etcd-v3.4.16-darwin-amd64/etcd",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

    http_archive(
        name = "com_coreos_etcd_linux_amd64",
        sha256 = "2e2d5b3572e077e7641193ed07b4929b0eaf0dc2f9463e9b677765528acafb89",
        urls = ["https://github.com/etcd-io/etcd/releases/download/v3.4.16/etcd-v3.4.16-linux-amd64.tar.gz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "etcd-v3.4.16-linux-amd64/etcd",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

# Define rules for different kubectl versions
def install_kubectl():
    http_file(
        name = "kubectl_1_18_darwin",
        executable = 1,
        sha256 = "5eda86058a3db112821761b32afce3fdd2f6963ab580b1780a638ac323864eba",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/darwin/amd64/kubectl"],
    )

    http_file(
        name = "kubectl_1_18_linux",
        executable = 1,
        sha256 = "bb16739fcad964c197752200ff89d89aad7b118cb1de5725dc53fe924c40e3f7",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectl"],
    )

# Define rules for different golangci-lint versions
def install_golangci_lint():
    http_archive(
        name = "golangci_lint_darwin",
        sha256 = "d4bd25b9814eeaa2134197dd2c7671bb791eae786d42010d9d788af20dee4bfa",
        urls = ["https://github.com/golangci/golangci-lint/releases/download/v1.42.0/golangci-lint-1.42.0-darwin-amd64.tar.gz"],
        build_file_content =
         """
filegroup(
     name = "file",
     srcs = [
        "golangci-lint-1.42.0-darwin-amd64/golangci-lint",
     ],
     visibility = ["//visibility:public"],
)
    """,
    )

    http_archive(
            name = "golangci_lint_m1",
            sha256 = "f649893bf2b1d24b2632b5e109884a15f3bf25cfdad46b34fb8fd13a016098fd",
            urls = ["https://github.com/golangci/golangci-lint/releases/download/v1.42.1/golangci-lint-1.42.1-darwin-arm64.tar.gz"],
            build_file_content =
             """
filegroup(
    name = "file",
    srcs = [
       "golangci-lint-1.42.1-darwin-arm64/golangci-lint",
    ],
    visibility = ["//visibility:public"],
)
    """,
    )

    http_archive(
        name = "golangci_lint_linux",
        sha256 = "6937f62f8e2329e94822dc11c10b871ace5557ae1fcc4ee2f9980cd6aecbc159",
        urls = ["https://github.com/golangci/golangci-lint/releases/download/v1.42.0/golangci-lint-1.42.0-linux-amd64.tar.gz"],
        build_file_content =
         """
filegroup(
     name = "file",
     srcs = [
        "golangci-lint-1.42.0-linux-amd64/golangci-lint",
     ],
     visibility = ["//visibility:public"],
)
    """,
    )

# Define rules for different oc versions
def install_oc():
    versions = {
        "oc_darwin": {
            "file": "openshift-client-mac.tar.gz",
            "sha": "d4fe964b9aa43a9287f0e9303eee509459964a5b27b9eb0feb4ea108d6ac42aa",
        },
        "oc_linux": {
            "file": "openshift-client-linux.tar.gz",
            "sha": "8906c39cc881eb8565e7cb756e9cd06917b439244e8bf9c4b0ee14d5a087588d",
        }
    }

    for k, v in versions.items():
        http_archive(
            name = k,
            sha256 = v["sha"],
            urls = ["https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable-4.9/{}".format(v["file"])],
            build_file_content =
             """
filegroup(
     name = "file",
     srcs = ["oc"],
     visibility = ["//visibility:public"],
)
    """,
        )

## Fetch kind images used during e2e tests
def install_kind():
    # install kind binary
    http_file(
        name = "kind_darwin",
        executable = 1,
        sha256 = "432bef555a70e9360b44661c759658265b9eaaf7f75f1beec4c4d1e6bbf97ce3",
        urls = ["https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-darwin-amd64"],
    )

    http_file(
            name = "kind_m1",
            executable = 1,
            sha256 = "4f019c578600c087908ac59dd0c4ce1791574f153a70608adb372d5abc58cd47",
            urls = ["https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-darwin-arm64"],
    )

    http_file(
        name = "kind_linux",
        executable = 1,
        sha256 = "949f81b3c30ca03a3d4effdecda04f100fa3edc07a28b19400f72ede7c5f0491",
        urls = ["https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-linux-amd64"],
    )


## Fetch kubetest2 binary used during e2e tests
def install_kubetest2():
    # install kubetest2 binary
    # TODO osx support
    http_file(
       name = "kubetest2_darwin",
       executable = 1,
       sha256 = "5b20aadd05eca47dead180a7c8296d75e81c184aabf182d4a41ef96597db543d",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/osx/kubetest2"],
    )

    http_file(
        name = "kubetest2_linux",
        executable = 1,
        sha256 = "7f0b05654fa43ca1c607db297b5f3a775f65eea90355bb6b10137a7fffff5e1a",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/kubetest2"],
    )

## Fetch kubetest2-kind binary used during e2e tests
def install_kubetest2_kind():
    # install kubetest2-kind binary
    # TODO osx support
    http_file(
       name = "kubetest2_kind_darwin",
       executable = 1,
       sha256 = "a68bad1b94fd5e432f0555d699d0ce0470d0bf16f1b087e857d55f16f5373385",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/osx/kubetest2-kind"],
    )

    http_file(
        name = "kubetest2_kind_linux",
        executable = 1,
        sha256 = "b13014d3e1464ce58e2bbbec94bead267936155d537f3232ec0a24727263a2a1",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/kubetest2-kind"],
    )

## Fetch kubetest2-gke binary used during e2e tests
def install_kubetest2_gke():
    # install kubetest2-gke binary
    # TODO osx support
    http_file(
       name = "kubetest2_gke_darwin",
       executable = 1,
       sha256 = "a1cbe02f61931dbe6c8d1662442f42cb538c81e4ec8cdd40f548f0e05cbd55a7",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/osx/kubetest2-gke"],
    )

    http_file(
        name = "kubetest2_gke_linux",
        executable = 1,
        sha256 = "9ac658234efc7f59968888662dd2d21908587789f6b812392ac5b6766b17c0b4",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/kubetest2-gke"],
    )
## Fetch kubetest2-tester-exe binary used during e2e tests
def install_kubetest2_exe():
    # install kubetest2-exe binary
    # TODO osx support
    http_file(
       name = "kubetest2_exe_darwin",
       executable = 1,
       sha256 = "818690cb55590440e163b18dd139c8a8714df9480f869bafe19eb344047cf37c",
       urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/osx/kubetest2-tester-exec"],
    )

    http_file(
        name = "kubetest2_exe_linux",
        executable = 1,
        sha256 = "4483f40f48b98e8a6aa41f58bfdf1f2787066a4e1ad1343e4281892aa1326736",
        urls = ["https://storage.googleapis.com/crdb-bazel-artifacts/linux/kubetest2-tester-exec"],
    )

## Fetch operator-sdk used on generating csv
def install_operator_sdk():
    versions = {
        "operator_sdk_darwin": {
            "file": "operator-sdk_darwin_amd64",
            "sha": "5fc30d04a31736449adb5c9b0b44e78ebeaa5cf968cc7afcbdf533135b72e31a",
        },
        "operator_sdk_linux": {
            "file": "operator-sdk_linux_amd64",
            "sha": "d2065f1f7a0d03643ad71e396776dac0ee809ef33195e0f542773b377bab1b2a",
        },
    }

    for k, v in versions.items():
      http_file(
         name = k,
         executable = 1,
         sha256 = v["sha"],
         urls = ["https://github.com/operator-framework/operator-sdk/releases/download/v1.15.0/{}".format(v["file"])],
      )

def install_kustomize():
    http_archive(
           name = "kustomize_darwin_arm",
           sha256 = "9556143d01feb9d9fa7706a6b0f60f74617c808f1c8c06130647e36a4e6a8746",
           urls = ["https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.4.0/kustomize_v4.4.0_darwin_arm64.tar.gz"],
           build_file_content = """
filegroup(
    name = "file",
    srcs = [
    "kustomize",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

    http_archive(
       name = "kustomize_darwin",
       sha256 = "77898f8b7c37e3ba0c555b4b7c6d0e3301127fa0de7ade6a36ed767ca1715643",
       urls = ["https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.3.0/kustomize_v4.3.0_darwin_amd64.tar.gz"],
       build_file_content = """
filegroup(
    name = "file",
    srcs = [
	"kustomize",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

    http_archive(
        name = "kustomize_linux",
        sha256 = "d34818d2b5d52c2688bce0e10f7965aea1a362611c4f1ddafd95c4d90cb63319",
        urls = ["https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.3.0/kustomize_v4.3.0_linux_amd64.tar.gz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
	"kustomize",
    ],
    visibility = ["//visibility:public"],
)
""",
    )

## Fetch opm used on generating csv
def install_opm():
    versions = {
        "opm_darwin": {
            "file": "opm-mac.tar.gz",
            "sha": "fe8a661e6b7a5096d4d9e223a62681fcd65860bf91a6dc7dc4d82ce07987c262",
         },
        "opm_linux": {
            "file": "opm-linux.tar.gz",
            "sha": "789e196feb4d1e943eca43e9011cc57dcdd5785946238dd85262f11489b3be3c",
        },
    }

    for k, v in versions.items():
      http_file(
         name = k,
         executable = 1,
         sha256 = v["sha"],
         urls = ["https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable-4.9/{}".format(v["file"])],
      )

## Fetch openshift-installer
def install_openshift():
    versions = {
        "openshift_darwin": {
            "file": "openshift-install-mac.tar.gz",
            "sha": "08171ba88f981c819cde2eb1fa1564b854ad534d9f75a28e75fc2a7de5f11ce9",
         },
        "openshift_linux": {
            "file": "openshift-install-linux.tar.gz",
            "sha": "e00f4da6b5041398a3fc7c53dcf6a6f0d4886b6be0ead0e1b1cb68a764483da0",
        },
    }

    for k, v in versions.items():
      http_archive(
          name = k,
          sha256 = v["sha"],
          urls = ["https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/stable-4.9/{}".format(v["file"])],
          build_file_content = """
filegroup(
    name = "file",
    srcs = ["openshift-install"],
    visibility = ["//visibility:public"],
)
"""
      )

## Fetch crdb used in our container
def install_crdb():
    http_archive(
       name = "crdb_darwin", # todo fix or remove
       sha256 = "bbe3a03c661555e8b083856c56c8a3b459f83064d1d552ed3467cbfb66e76db7",
       urls = ["https://binaries.cockroachdb.com/cockroach-v20.2.5.darwin-10.9-amd64.tgz"],
       build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "cockroach-v20.2.5.darwin-10.9-amd64/cockroach",
    ],
    visibility = ["//visibility:public"],
)
""",
   )

    http_archive(
        name = "crdb_linux",
        sha256 = "57f4b00c736d8511328d6f33997a3a66cb4ec7142cb126d872dade399a0922e6",
        urls = ["https://binaries.cockroachdb.com/cockroach-v20.2.5.linux-amd64.tgz"],
        build_file_content = """
filegroup(
    name = "file",
    srcs = [
        "cockroach-v20.2.5.linux-amd64/cockroach",
    ],
    visibility = ["//visibility:public"],
)
""",
   )

# fetch binaries for aws kubetest2
def install_kubetest2_aws():
    http_file(
        name = "aws-k8s-tester-linux",
        executable = 1,
        sha256 = "f6a95feef94ab9a96145fff812eeebae14f974edfead98046351f3098808df54",
        urls = ["https://github.com/aws/aws-k8s-tester/releases/download/v1.6.1/aws-k8s-tester-v1.6.1-linux-amd64"],
    )
    http_file(
            name = "aws-k8s-tester-m1",
            executable = 1,
            sha256 = "a0c4d6125c0dac4d5333560243975ecc2ef7712b71f5bd29e79c3f450ec7165e",
            urls = ["https://github.com/aws/aws-k8s-tester/releases/download/v1.6.5/aws-k8s-tester-v1.6.5-darwin-arm64"],
    )
    http_file(
        name = "aws-k8s-tester-darwin",
        executable = 1,
        sha256 = "a724288e8f2b87df89c711ed59fa2a09db5ad2b50a35cb3039fa610408b99b32",
        urls = ["https://github.com/aws/aws-k8s-tester/releases/download/v1.6.1/aws-k8s-tester-v1.6.1-darwin-amd64"],
    )
