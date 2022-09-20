module github.com/cockroachdb/cockroach-operator

go 1.17

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/banzaicloud/k8s-objectmatcher v1.3.2
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cockroachdb/errors v1.8.0
	github.com/dnaeon/go-vcr v1.0.1
	github.com/dustin/go-humanize v1.0.0
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/google/go-cmp v0.5.8
	github.com/gosimple/slug v1.9.0
	github.com/jackc/pgx/v4 v4.9.0
	github.com/lithammer/shortuuid/v3 v3.0.7
	github.com/octago/sflags v0.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.21.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.25.1
	k8s.io/apimachinery v0.25.1
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/code-generator v0.25.0
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.70.1
	k8s.io/kube-openapi v0.0.0-20220803162953-67bda5d908f1
	sigs.k8s.io/controller-runtime v0.13.0
	sigs.k8s.io/controller-tools v0.5.0
	sigs.k8s.io/kubetest2 v0.0.0-20210720070532-ea531e01c240
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go v0.97.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/redact v1.0.6 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.8.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/gobuffalo/flect v0.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/goph/emperror v0.17.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.7.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.5 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.5.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/spf13/cobra v1.4.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858 // indirect
	golang.org/x/tools v0.1.12 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.25.0 // indirect
	k8s.io/component-base v0.25.0 // indirect
	k8s.io/gengo v0.0.0-20211129171323-c02415ce4185 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

// indirect dependencies that need to be updated

replace k8s.io/client-go v9.0.0+incompatible => k8s.io/client-go v0.25.1
