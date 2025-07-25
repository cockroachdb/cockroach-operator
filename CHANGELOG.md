# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

# [Unreleased](https://github.com/cockroachdb/cockroach-operator/compare/v2.18.2...master)

# [v2.18.2](https://github.com/cockroachdb/cockroach-operator/compare/v2.18.1...v2.18.2)
* Added the priorityClassName configuration
* Make timestamp human readable from linux epoch in operator logs

# [v2.18.1](https://github.com/cockroachdb/cockroach-operator/compare/v2.18.0...v2.18.1)
* Added support for openshift 4.19

# [v2.18.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.17.0...v2.18.0)
* Added support for skipping innovative releases post 24.x
* Added support for QoS on init containers.

# [v2.17.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.16.1...v2.17.0)
* Added support for linux/arm64 images.
* Added support for TerminationGracePeriodSeconds in statefulset.
* Upgraded underlying dependencies for cockroach-operator.

# [v2.16.1](https://github.com/cockroachdb/cockroach-operator/compare/v2.16.0...v2.16.1)
* Increased memory limits for the version checker job

# [v2.16.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.15.1...v2.16.0)

# [v2.15.1](https://github.com/cockroachdb/cockroach-operator/compare/v2.15.0...v2.15.1)
* Bug fix, removing deprecated cluster setting `kv.snapshot_recovery.max_rate` from being checked during decommission.

# [v2.15.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.14.0...v2.15.0)

# [v2.14.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.13.0...v2.14.0)

# [v2.13.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.12.0...v2.13.0)

# [v2.12.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.11.0...v2.12.0)

# [v2.11.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.10.0...v2.11.0)

* Bug fix to return correct error message if value passed for `whenUnsatisfiable` under `topologySpreadConstraints` field is invalid.
* Add openshift release process for cockroach operator

# [v2.10.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.9.0...v2.10.0)

## Added

* Upgrade the underlying k8s dependencies from 1.20 to 1.21 to support operator installation on k8s 1.25+.

# [v2.9.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.8.0...v2.9.0)

## Fixed

* Install init container certs with 600 permissions
* Ensure operator can connect to DBs in all namespaces

# [v2.8.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.7.0...v2.8.0)

## Added

* `AutomountServiceAccountToken` field for cluster spec to allow mounting the default service account token.

## Fixed

* Delete the CancelLoop function, fixing a cluster status update bug
* Correctly detect failed version checker Pods
* retry cluster status updates, reducing test flakes

## Changed
* Update validation webhook to reject changes to cluster spec's AdditionalLabels field

# [v2.7.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.6.0...v2.7.0)

## Fixed

* Grant operator deletecollection permissions to fix fullcluster restart flow
* Grant operator list and update permissions on pvcs to fix pvc resize flow
* Bump TerminationGracePeriodSeconds from 1m to 5m
* Prefer user added --join flags over default when explicitly passed

# [v2.6.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.5.3...v2.6.0)

## Added

* Custom logging configuration can be used through the configmap when `spec.logConfigMap` is provided.

# [v2.5.3](https://github.com/cockroachdb/cockroach-operator/compare/v2.5.2...v2.5.3)

## Fixed

* Fix nil-pointer errors when `spec.Image` is not provided.
* Update gogo/protobuf to address CVE-2021-3121

# [v2.5.2](https://github.com/cockroachdb/cockroach-operator/compare/v2.5.1...v2.5.2)

## Fixed

* #863 - Add flag for `leader-election-id` to enable leader election support

## Changed

* Image digests now calculated when generating templates, rather than when creating bundles
* Use [bazelisk](https://github.com/bazelbuild/bazelisk) for CI workflows

# [v2.5.1](https://github.com/cockroachdb/cockroach-operator/compare/v2.5.0...v2.5.1)

## Fixed

* Bundle generation for updated OpenShift marketplace requirements
* Related images added to manager env for supporting cockroachDBVersion in the spec
* Fixed operator crash loop when cockroachDBVersion is used.
* Fix add custom annotations to the pod created by the job
* Fix issue when sidecar container is injected to job pod
* Add support for pod TopologySpreadConstraint and associated feature gate

## Changed

* Now validates if `pvc.Volumemode` set correctly to `Filesystem`

# [v2.5.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.4.0...v2.5.0)

## Fixed

* Mark nodes as decommissioned after draining the node

## Added

* Support for UI and SQL Ingress

# [v2.4.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.3.0...v2.4.0)

## Added

* Fixed resource requirements for vcheck container
* Ability to run the operator as a cluster scoped operator or a namespace scoped one

## Changed

* Deprecated legacy OpenShift packaging format in favor of new bundle format
* Removed unused beta channel
* Dynamically create instance specific service account, role, and role binding
* OpenShift deployment now allows the operator to run for all namespaces

## Fixed

* Boilerplate test after updating to Go 1.17
* Permissions for sts/scale subresource

## Deleted

* Versioned Dockerfiles for OpenShift

# [v2.3.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.2.1...v2.3.0)

## Added

* Support for nodeSelectors

## Changed

* Cleanup of initialization and structure of Actor and Director

## Fixed

* Error on startup in OpenShift related to webhook configuration
* Finalizer permissions to address `cannot set blockOwnerDeletion if an ownerReference refers to a resource you can’t set finalizers on` issue

# [v2.2.1](https://github.com/cockroachdb/cockroach-operator/compare/v2.2.0...v2.2.1)

## Changed

* Webhook CA certificate is stored in `cockroach-operator-webhook-ca` (autogenerated if missing)
* Webhook server certificates are now ephemeral and created at manager pod startup

## Deleted

* References to certificates/v1beta1 preventing the operator from working on K8s 1.22

# [v2.2.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.2.0-beta.2...v2.2.0)

## Added

* Added CHANGELOG.md to track changes across releases
* More information to operator logs (e.g. reconcilerID)
* Support for taints and tolerations
* Support for custom annotations
* Mutating and validating webhooks and associated TLS configuration
* Several enhancements for local development (e.g. `make dev/up`)

## Changed

* Refactor templates to leverage kustomize bases/overlays in //config
* Updated validation markers for the CRD (required, minimum, etc.)
* Introduced Director to consolidate the handles logic from the actors
* Standardized on cockroachdb/errors for wrapping errors
* Made enormous-sized tests run for less long

## Fixed

* Propagate affinity and tolerations for version checker job
* Ensure the controller watches owned job objects
* CertificateGenerated conditional correctly set during deploys
* No longer requeueing permanent errors
* PVCs no longer missing `AdditionalLabels`
* Some flakiness in e2e tests
* Version checker jobs no longer pile up when pods take >2m to come up

# [v2.2.0-beta.2](https://github.com/cockroachdb/cockroach-operator/compare/v2.2.0-beta.1...v2.2.0-beta.2)

## Added

* More information to operator logs (e.g. reconcilerID)

## Changed

* Refactor templates to leverage kustomize bases/overlays in //config

## Fixed

* Propagate affinity and tolerations for version checker job
* Ensure the controller watches owned job objects
* CertificateGenerated conditional correctly set during deploys

# [v2.2.0-beta.1](https://github.com/cockroachdb/cockroach-operator/compare/v2.1.0...v2.2.0-beta.1)

## Added

* Support for taints and tolerations
* Support for custom annotations
* Mutating and validating webhooks and associated TLS configuration
* Several enhancements for local development (e.g. `make dev/up`)

## Changed

* Updated validation markers for the CRD (required, minimum, etc.)
* Introduced Director to consolidate the handles logic from the actors
* Standardized on cockroachdb/errors for wrapping errors

## Fixed

* No longer requeueing permanent errors
* PVCs no longer missing `AdditionalLabels`
* Some flakiness in e2e tests
* Version checker jobs no longer pile up when pods take >2m to come up

# [v2.1.0](https://github.com/cockroachdb/cockroach-operator/compare/v2.0.1...v2.1.0)

## Added

* `AdditionalLabels` which are added to all managed resources
* Examples for new features (addition labels, affinity rules, etc)
* e2e tests for EKS, OpenShift, and downgrading

## Changed

* Skip the initContainer when running an insecure cluster
* Can now pass additional parameters to OpenShift
* Updating of crdbversions.yaml now automated via GitHub actions

## Deleted

* Dead code from multiple packages
* Dependency on ginko and gomega
