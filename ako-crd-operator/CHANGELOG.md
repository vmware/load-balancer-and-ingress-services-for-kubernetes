# Change log:

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## AKO-CRD-Operator-2.1.1

### Added:
 - Initial release of AKO CRD Operator for managing Avi-specific objects.
 - Support for [HealthMonitor CRD](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/healthmonitor.md) of type HTTP,TCP,Ping. 
 - Support for [ApplicationProfile CRD](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/applicationprofile.md).
 - Support for [PKIProfile CRD](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/pkiprofile.md).
 - Support for [RouteBackendExtension CRD](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/routebackendextension.md) for advanced backend configuration.
 - Support for referring [PKIProfile CRD](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/pkiprofile.md) in [RouteBackendExtension CRD](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/routebackendextension.md).
 - Multi-tenant support with namespace-scoped resources.

