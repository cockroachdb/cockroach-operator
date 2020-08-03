module github.com/cockroachdb/cockroach-operator

go 1.14

require (
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/banzaicloud/k8s-objectmatcher v1.3.2
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.4.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	go.uber.org/zap v1.14.1
	golang.org/x/crypto v0.0.0-20200311171314-f7b00557c8c4 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.18.5 // indirect
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/code-generator v0.18.5
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/controller-runtime v0.5.4
	sigs.k8s.io/controller-tools v0.2.9-0.20200414181213-645d44dca7c0
	sigs.k8s.io/kubetest2 v0.0.0-20200807173356-3d574132ed2e
	sigs.k8s.io/yaml v1.2.0
)

replace (
	k8s.io/client-go v9.0.0+incompatible => k8s.io/client-go v0.18.5
	sigs.k8s.io/controller-runtime v0.5.4 => sigs.k8s.io/controller-runtime v0.5.1-0.20200416234307-5377effd4043
)
