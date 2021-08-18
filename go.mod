module github.com/cockroachdb/cockroach-operator

go 1.14

require (
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/banzaicloud/k8s-objectmatcher v1.3.2
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cockroachdb/errors v1.8.0
	github.com/dnaeon/go-vcr v1.0.1
	github.com/dustin/go-humanize v1.0.0
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.2.0
	github.com/google/go-cmp v0.5.4
	github.com/gosimple/slug v1.9.0
	github.com/jackc/pgx/v4 v4.9.0
	github.com/lib/pq v1.3.0
	github.com/octago/sflags v0.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.15.0
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/code-generator v0.20.2
	k8s.io/klog/v2 v2.4.0
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	sigs.k8s.io/controller-runtime v0.8.2
	sigs.k8s.io/controller-tools v0.5.0
	sigs.k8s.io/kubetest2 v0.0.0-20210720070532-ea531e01c240
	sigs.k8s.io/yaml v1.2.0
)

replace k8s.io/client-go v9.0.0+incompatible => k8s.io/client-go v0.20.2
