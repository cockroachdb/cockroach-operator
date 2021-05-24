FROM scratch

LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=cockroachdb-certified
LABEL operators.operatorframework.io.bundle.channels.v1=beta,stable
LABEL operators.operatorframework.io.bundle.channel.default.v1=stable

LABEL com.redhat.openshift.versions="v4.6"
LABEL com.redhat.delivery.backport=false
LABEL com.redhat.delivery.operator.bundle=true

COPY 1.7.6/manifests /manifests/
COPY 1.7.6/metadata /metadata/
