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

## AKO-Gateway-1.13.1

### Changed:
 - AKO now accepts Gateway with some valid and some invalid listeners.
 - AKO now allows creation/updation of HTTPRoute and Gateway in any order except for following cases which requires an explicit ako reboot to work.
 1. When one gateway attached to multiple routes, having all valid listeners, transitions to gateway with some invalid listeners.
 2. When one gateway attached to multiple routes, having all valid listeners, transitions to gateway with all invalid listeners.
 3. When one gateway attached to multiple routes, having some valid listeners, transitions to gateway with all invalid listeners.
 4. When one gateway attached to one route, having some valid listeners, transitions to gateway with all invalid listeners.
 5. When one gateway attached to one route, having all valid listeners, transitions to gateway with all invalid listeners.
 - HTTPRoute deletion will not update Gateway listener status with "Virtual service deleted".
 - If Secret specified in the TLS section does not exist, AKO will invalidate the Gateway listener and will not generate any configuration corresponding to it. If an existing Secret is deleted, then Parent VS configurations corresponding to it will get deleted.
 - Hostname in Gateway listener is now an optional field.
 - HTTPRoute -> spec -> rules -> matches is now an optional field.

### Added:
 - Support for Service of type NodePortLocal.
 - Support for wildcard in Gateway->listener->Hostname.
 - Support for wildcard prefixed Hostname in HTTPRoute.
 - Support for ListenerConditionAccepted, ListenerConditionResolvedRefs, ListenerConditionProgrammed in Gateway Listener status.
 - Support for GatewayConditionAccepted and GatewayConditionProgrammed in Gateway status.
 - Support for RouteConditionAccepted, RouteConditionResolvedRefs in HTTPRoute status.
 - Support for sending 404 Response code, if no path matches for a request.
 - Support for multiple listeners in a Gateway having same port and protocol and different hostname and name.
 - Support for graceful shutdown of backend servers.
 - Support for Endpointslices or Endpoints using a toggle.
 - Events will now be raised if status update of Gateway or HTTPRoute fails.

### Fixed
 - Fix: AKO-Gateway does not create Virtual service if Gateway have multiple listeners with same host name but different port and single http route is attaching to it.

 