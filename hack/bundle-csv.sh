docker run -v `pwd`:`pwd` -w `pwd` operator-courier/operator-courier:v2.1.7 operator-courier verify --ui_validate_io $PACKAGE_PATH#!/usr/bin/env bash
set -e

VERSION=$1
REGISTRY_IMAGE=$2
CREATED_TIME=`date +"%FT%H:%M:%SZ"`
DATETIME=$(date +"%FT%H%M%SZ")
OP=cockroach-operator

# TODO this only runs properly from the base directory, we need to make it run
# from this directory as well

rm -rf bundle || echo "nothing to clean"
mkdir -p bundle

cp -r "deploy/olm-catalog/$OP" "bundle"

PACKAGE_PATH="bundle/$OP"
CSV_PATH="bundle/$OP/${VERSION}"

yq w \
    -i $CSV_PATH/${OP}.v${VERSION}.clusterserviceversion.yaml \
    'metadata.annotations.containerImage' ${REGISTRY_IMAGE}

yq w \
    -i $CSV_PATH/${OP}.v${VERSION}.clusterserviceversion.yaml \
    'metadata.annotations.createdAt' ${CREATED_TIME}

yq w \
    -i $CSV_PATH/${OP}.v${VERSION}.clusterserviceversion.yaml \
    'spec.install.spec.deployments[0].spec.template.spec.containers[0].image' ${REGISTRY_IMAGE}

STABLE=$(yq r \
     $PACKAGE_PATH/${OP}.package.yaml \
    'channels.(name==stable).currentCSV' | sed 's/cockroach-operator\.v//')

BETA=$(yq r \
     $PACKAGE_PATH/${OP}.package.yaml \
     'channels.(name==beta).currentCSV' | sed 's/cockroach-operator\.v//')

docker run -v `pwd`:`pwd` -w `pwd` operator-courier/operator-courier:v2.1.7 operator-courier verify --ui_validate_io $PACKAGE_PATH

FILENAME="rhm-op-bundle-s${STABLE}-b${BETA}-d${DATETIME}.zip"

cd bundle/${OP} || echo "failed to cd"
zip -r ${FILENAME} . -x 'manifests/*' -x 'metadata/*'
mv ${FILENAME} ../

echo "created bundle: bundle/${FILENAME}"
