# Change log:

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## AKO-Gateway-1.11.1

### Added:
 - AKO claims support for v1beta1 for HttpRoute, Gateway, GatewayClass.

## AKO-Gateway-1.12.1

### Changed:
 - If two Gateway listeners have same PORT and PROTOCOL pair, AKO GAteway will take Union of both the listeners.

### Added:
 - AKO now claims support for v1 for HttpRoute, Gateway, GatewayClass.

## AKO-Gateway-1.12.2

### Fixed:
 - Fix: AKO Gateway does not create a virtual service if Gateway has multiple listeners with the same host name.
 - Fix: AKO Gateway container crashes when it boots up in NPL mode.

## AKO-Gateway-1.13.1

### Changed:
 - AKO now accepts Gateway with some valid and some invalid listeners.
 - AKO now allows creation/updation of HTTPRoute and Gateway in any order([Details](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/gateway-api/gateway-api-v1.md#resource-creation)).
 - HTTPRoute deletion will not update Gateway listener status with "Virtual service deleted".
 - If Secret specified in the TLS section does not exist, AKO will invalidate the Gateway listener and will not generate any configuration corresponding to it. If an existing Secret is deleted, then Parent VS configurations corresponding to it will get deleted.

### Added:
 - Support for Service of type NodePortLocal.
 - Support for NSX-T cloud.
 - Support for wildcard in Gateway->listener->Hostname.
 - Support for wildcard prefixed Hostname in HTTPRoute.
 - Support for ListenerConditionAccepted, ListenerConditionResolvedRefs and ListenerConditionProgrammed in Gateway Listener status, GatewayConditionAccepted and GatewayConditionProgrammed in Gateway status and RouteConditionAccepted and RouteConditionResolvedRefs in HTTPRoute status.
 - Support for sending 404 Response code, if no path matches for a request.
 - Support for multiple listeners in a Gateway having same port and protocol and different hostname and name.
 - Support for graceful shutdown of backend servers.
 - Support for Endpointslices or Endpoints using `enableEndpointSlice` flag in configmap.
 - Events will now be raised if status update of Gateway or HTTPRoute fails.

### Fixed
 - Fix: AKO-Gateway does not create Virtual service if Gateway have multiple listeners with same host name but different port and single http route is attaching to it.


## AKO-Gateway-1.13.2

### Added
- Support Regular Expression in HTTPRoute Path, present in Matches section of HTTPRoute
- Support `urlRewrite` filter in HTTPRoute with `replaceFullPath` as path value.

### Fixed:
- Fix: AKO creates label on SE group when GatewayAPI is enabled.

## AKO-Gateway-1.14.1

### Changed:
- AKO will Set min pool up in vs to 1 so Gateway child VS will go down on zero pools up. 
- Support for `Endpoints` using `enableEndpointSlice` flag is removed.

### Added:
- AKO now claims support for v1.3 for HttpRoute, Gateway, GatewayClass.
- Suport for MTLS support for Istio with Gateway API in Nodeport and ClusterIP mode.
- Support for Named Route Rule.
- Support for Health Monitor CRD([Details](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/3d7b3e45d7361d0a035e7afc10572ef48c8fde45/docs/crds/healthmonitor.md)). 
- Support for Application profile CRD([Details](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/3d7b3e45d7361d0a035e7afc10572ef48c8fde45/docs/crds/applicationprofile.md)).
- Support for Persistence profile([Details](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/routebackendextension.md#session-persistence-configuration)).
- Support for RouteBackendExtension CRD([Details](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/3d7b3e45d7361d0a035e7afc10572ef48c8fde45/docs/crds/routebackendextension.md)).
- Support for L7Rule CRD([Details](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/l7rule.md#attaching-l7rule-to-httproute)). 
- Support for enabling AKO (with Gateway API) to leverage SSL to talk to backend servers([Details](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/crds/routebackendextension.md#backend-tlsssl-configuration)).
- Support to disable/enable the traffic enabled knob of VS for Gateway ParentVS.
- Support for VPC Mode when Gateway API is enabled.
- Regex support in path in GatewayAPI.
- Support for enabling WAF Protection for Gateways and HTTRoute.