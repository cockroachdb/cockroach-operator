# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

# [Unreleased](https://github.com/cockroachdb/cockroach-operator/compare/v2.2.1...master)

# [v2.2.1](https://github.com/cockroachdb/cockroach-operator/compare/v2.2.0...v2.2.1)

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
