FROM scratch

LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=deploy/certified-metadata-bundle/cockroach-operator/
LABEL operators.operatorframework.io.bundle.channels.v1=beta,stable
LABEL operators.operatorframework.io.bundle.channel.default.v1=stable

COPY deploy/certified-metadata-bundle/cockroach-operator/1.1.0/manifests /manifests/
COPY deploy/certified-metadata-bundle/cockroach-operator/1.1.0/metadata /metadata/
