# This script installs the Operator in development.
# Bazel has go rules that require root access because of Big Sur (macOS).
# When running with sudo/root access, it opens a new shell, so the environment variables are not used with the existing Developer Install Instructions (https://github.com/cockroachdb/cockroach-operator/blob/master/DEVELOPER.md#developer-install-instructions).
# This script injects the required environment variables to allow the "make k8s/apply" command to succeed.
sudo \
CLUSTER_NAME=john-test \
GCP_ZONE=us-east1-c \
DEV_REGISTRY=us.gcr.io/cockroach-john-277713 \
GCP_PROJECT=cockroach-john-277713 \
make k8s/apply
