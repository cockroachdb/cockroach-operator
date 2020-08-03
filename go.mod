module github.com/cockroachdb/cockroach-operator

go 1.14

require (
	cloud.google.com/go v0.61.0 // indirect
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/banzaicloud/k8s-objectmatcher v1.4.0
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/google/go-cmp v0.5.1
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sys v0.0.0-20200727154430-2d971f7391a4 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver v0.18.6 // indirect
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451
	sigs.k8s.io/controller-runtime v0.6.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	k8s.io/utils => k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
)
