# Change log:

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## AKO-CRD-Operator-2.1.1

### Added:
 - Initial release of AKO CRD Operator for managing Avi Controller specific objects.
 - Support for [HealthMonitor CRD](../docs/crds/healthmonitor.md) of type HTTP, TCP, Ping. 
 - Support for [ApplicationProfile CRD](../docs/crds/applicationprofile.md).
 - Support for [PKIProfile CRD](../docs/crds/pkiprofile.md).
 - Support for [RouteBackendExtension CRD](../docs/crds/routebackendextension.md) for advanced backend configuration.
 - Support for referring [PKIProfile CRD](../docs/crds/pkiprofile.md) in [RouteBackendExtension CRD](../docs/crds/routebackendextension.md).
 - Multi-tenant support with namespace-scoped resources.

