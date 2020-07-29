bazel test //api/... //pkg/... \
	--test_env=TEST_ASSET_ETCD=$(bazel info execution_root)/external/etcd_bin/file/etcd, \
	--test_env=TEST_ASSET_KUBECTL=$(bazel info execution_root)/external/kubectl_bin/file/kubectl \
	--test_env=TEST_ASSET_KUBE_APISERVER=$(bazel info execution_root)/external/kube_apiserver_bin/file/kube-apiserver
