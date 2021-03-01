module github.com/cockroachdb/cockroach-operator

go 1.14

require (
	github.com/Azure/azure-storage-blob-go v0.12.0 // indirect
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/banzaicloud/k8s-objectmatcher v1.3.2
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cockroachdb/cockroach v0.0.0-20210204185734-162c5ac4968c
	github.com/cockroachdb/errors v1.8.0
	github.com/dustin/go-humanize v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.4.0
	github.com/jackc/pgx/v4 v4.9.0
	github.com/lib/pq v1.8.0
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.14.1
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9
	gopkg.in/yaml.v2 v2.3.0
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
	github.com/abourget/teamcity => github.com/cockroachdb/teamcity v0.0.0-20180905144921-8ca25c33eb11
	github.com/gogo/protobuf => github.com/cockroachdb/gogoproto v1.2.1-0.20210111172841-8b6737fea948
	github.com/grpc-ecosystem/grpc-gateway => github.com/cockroachdb/grpc-gateway v1.14.6-0.20200519165156-52697fc4a249
	github.com/olekukonko/tablewriter => github.com/cockroachdb/tablewriter v0.0.5-0.20200105123400-bd15540e8847
	go.etcd.io/etcd => github.com/cockroachdb/etcd v0.4.7-0.20200615211340-a17df30d5955
	gopkg.in/yaml.v2 => github.com/cockroachdb/yaml v0.0.0-20180705215940-0e2822948641
	k8s.io/client-go v9.0.0+incompatible => k8s.io/client-go v0.18.5
	sigs.k8s.io/controller-runtime v0.5.4 => sigs.k8s.io/controller-runtime v0.5.1-0.20200416234307-5377effd4043
	vitess.io/vitess => github.com/cockroachdb/vitess v2.2.0-rc.1.0.20180830030426-1740ce8b3188+incompatible

)
