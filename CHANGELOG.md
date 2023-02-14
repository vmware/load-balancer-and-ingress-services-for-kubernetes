# Change log:

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## AKO-0.9.1

### Changed:
 - SNI naming.  Requires full deletion of the SNI VS names in beta-1.
 - Retry for status updates for ingress.
 - Log levels. Transitioned some logs from INFO —> DEBUG.
 - Logging mechanism. Uses Uber-zap now.
 - Removed SDK logging.
 - Caching improvements. Fixes in race conditions.
 - Reduction of controller API calls during full sync.
 - Full sync fixes.
 - Enchanced Retry logic.
 - Removal of regular object cache syncs - only periodic refresh of cloud config parameters.
 
 
### Added:
 - Dynamic logging on the fly by editing the ConfigMap.
 - AKO API server - for liveness probe and basic controller debugging.
 - SNI VS sharing on the basis of hostnames. Same hostname will create only 1 VS across namespaces.
 - Option to disable full sync. Change fullSyncFrequency to 0.
 - Unused shared VS deletion on reboot of AKO.
 - Multiple sub-domain support with a specification of default sub-domain in `values.yaml` for service of type LB.
 
 ### Removed:
 - VRF context is now removed from `values.yaml` and instead is read from the network subnet.


## AKO-1.1.1

### Added:
 - HostRule/HTTPRule support for Kubernetes


## AKO-1.2.1

### Changed:
 - Liveness probe enhancements.
 - Stability fixes around pod restarts.
 - Retry layer improvements.
 - Cleanup fixes.
 - SDK bug fixes.
 - Logging improvements.

### Added:
 - Full OpenShift 4.x support for NodePort and ClusterIP
 - Per Cluster SE group support. Label based routing support.
 - NodePort Support for Kubernetes.
 - HostRule/HTTPRule support for Openshift.
 - Minimal public cloud support.

 ### Removed:
 - VRF context support deprecated.
 
 ## AKO-1.3.1

### Changed:
 - AKO support for IPAM without specification of networkName.
 - AKO support for controller credential change.

### Added:
 - AKO tenancy support. 
 - AKO operator feature.
 - AKO public cloud with ClusterIP support for GCP/Azure.
 - AKO support for GKE/AKS/EKS.
 - AKO selective namespace sync for Ingress/Route.
 - AKO support for static IP using LoadbalancerIP for L4.
 - OpenShift wildcard certificate support.
 - Global RHI support.
 - AKO support for avi controller object deletion updates via statefulset conditions.
 - AKO support for multiple new fields in HTTPRule/HostRule CRD.
 - Tolerance support for networking/v1 Ingress in k8s 1.19
 
## AKO-1.3.3
 
### Changed:
 - DNS IPAM configuration not required for L4.
 - Ingress class related fixes.
 - RHI knob related changes.
 
 
### Added:
 - Added auto-fqdn support

## AKO-1.3.4

### Added:
 - Option to use AKO as pure L7 ingress controller without L4 functionalities.
 - Option to enable/disable hostname addition for Services of type LB.

## AKO-1.4.1

### Added:
 - AviInfraSetting CRD for selecting specific Avi controller infra attributes.
 - Support for shared L4 VIP across multiple service of type loadbalancer. 
 - Selective namespace sync for L4 objects including GatewayAPI and Services of type LB.
 - Option to add global fqdn for a hostname via Host Rule.
 - Temporary support for HTTP Caching for secure ingresses/routes via Pool objects.
 - Option to use dedicated Virtual Service per Ingress hostname.
 - Support for Node Port Local with Antrea CNI.(Supported from Antrea 0.13 onwards)
 - Persistence profile in HTTPRule CRD.
 - Option to use a default secret for Ingresses via annotation.
 - AWS mult-vip support.
 - Enhanced Virtual Hosting support for Avi Enterprise License. (Tech preview)

### Changed:
 - `networkName` field in values.yaml is changed to `vipNetworkList`.
 - AKO qualification for Kubernetes 1.19, 1.20, 1.21.

### Removed:
 - namespace sharding is deprecated starting from this release.

## AKO-1.4.2

### Bugs fixed:

 - Fix: AKO removes LB status if annotations removal hits a snag
 - Fix: Failure in lb-service obtaining ip after expanding ipam range which is previously exhausted
 - Fix: EVH broken with SSL certs specified in HostRule
 - Fix: Multi-vip with AWS always assigns IP address from a single subnet
 - Fix: enable_rhi Error in ESSENTIALS license
 - Fix: AKO 1.4.1 Doesn't Watch Endpoints Object in NodePort mode
 - Fix: stale entries in httppolicysets cause AKO to panic
 - Fix: Unpredictable behavior in AKO for ingresses/routes with same FQDN and overlapping paths
 - Fix: Uncertain behavior of AKO for ingresses/routes with same FQDN but different paths and one of the path is "/"

## AKO-1.4.3

### Added:
 - Support for allowing AKO to get installed in user-provided namespace (other than avi-system).

### Bugs fixed:
 - Skip status updates on Service of type LoadBalancer during bootup when `layer7Only` flag is set to `true`.
 - Fix multi-host Ingress status updates during bootup.
 - Unblock AKO run if CRDs are not installed in cluster.
 - Fixed incorrect virtual service uuid annotation update for openshift secure routes with InsecureEdgeTermination set to Allow.

## AKO-1.5.1

### Added:
 - Add support for programming FQDN for L4 services via Gateway object when `servicesAPI` is set to `true`.
 - Multi-Protocol (TCP/UDP) support in gateway VS (shared VIP).
 - Make Service of type LoadBalancer work together with Gateways when using `serviceAPI` is set to `true`.
 - Public IP support for AKO on public clouds.
 - Support for passthrough hosts in Ingress.
 - Support for SAML based authentication for AKO using AuthToken as an alternative to usernme and password based authentication.
 - EVH support for Openshift.
 - NSX-T cloud support for VLAN and overlay based segments.
 - Support label based BGP peering for VSes.
 - Add markers to AVI objects.
 - Add length restriction on Avi object name upto 255 characters in SNI deployment.

### Changed:
 - Deprecate `subnetIP` and `subnetPrefix` in values.yaml, in favor of `cidr` field within `vipNetworkList`.
 - Update `spec.network` to include `networkName` and `cidr` information in AviInfraSetting CRD.
 - Encode Avi object names in EVH deployment.

### Known Issues
 - AKO update ingress status with VIP instead of public IP when public IP is enabled in public cloud deployments.

## AKO-1.5.2

### Added:
 - Bootup optimizations improves AKO’s boot time in scaled environments.
 - Support for HTTPRule CRD for Routes without paths.
 - Support for `Service` with multiple ports in EVH.

### Changed
 - In SNI deployment,  Host rule CRD with sslkeycertificate reference would get applied successfully to virtual service. However, upon deletion of this CRD, this virtual service would loose all cert configuration and get default cert. This issue is fixed.
 - In EVH deployment, existing certificate references on parent VS are overwritten if ingress (with host and secret) and Host rule CRD (with sslkeyandcertificate reference and different host) would get applied to same parent VS. Expected behaviour is to append all cert references belongs to different hosts that maps to same parent. This issue is fixed.

### Known Issues
 - Due to the use of Informers for Secrets, there is an adverse effect on bootup time in OpenShift based setups. AKO can further optimize bootup time on openshift setup by filtering out the Secrets on `avi-system` namespace. This feature will be added in 1.6.1
 - `ServiceType` of `NodePort` does not support multi-port `Services` with port number.


 ## AKO-1.6.1

 ### Added:
 - AKO now claims support for Kubernetes 1.22.
 - Support multi-port `Services` with port number for `ServiceType` of `NodePort` and `NodePortLocal`.
 - AKO introduces support for broadcasting kubernetes `Events` in order to enhance the observability and monitoring aspects.
 - Support custom port numbers for dedicated and shared virtual services through HostRule CRD.
 - Support for enabling/disabling non-significant logs through HostRule CRD.
 - Add autoFQDN to shared virtual services.
 - Support for Analytics Profile for virtual service through HostRule CRD.
 - Support for Static IP for Shared and Dedicated virtual service through HostRule CRD.
 - Tenant context support for SE group.
 - Support for multiple alias FQDNs for a host through HostRule CRD.
 - Support to configure `Pool Placement Setting` for child ingresses/routes/svclb through AviInfraSetting CRD.
 - Allow programming static routes by adding custom Pod CIDR to Node mapping via annotation.

### Changed:
- Update `annotation` instead of `status` field of AKO statefulset after avi object deletion through `deleteConfig` flag.
- AKO will create single HTTP Policyset object at Avi Controller side for all paths of same host.

### Known Issues:
 - AviInfraSetting CR can not be applied to passthrough ingress/route.
 - AKO is not updating the ingress status when annotation `passthrough.ako.vmware.com/enabled: "true"` is added to the ingress.
 - There are issues when shardVS size changed through AviInfra CR or values.yaml. Recommended workflow is to first delete existing config using `deleteConfig` flag as described [here](docs/faq.md#how-do-i-clean-up-all-my-configs) and then change `shardVS` size through AviInfraSetting.


## AKO-1.6.2

### Bugs fixed
 - Problem in creating VSVip for Passthrough routes.
 - Problem in correctly saving ipamType in AKO when no DNS providers are set in the Cloud.
 - Fixes around fqdnType Contains/Wildcard settings in HostRules.
 - Fix validations related to tcpSettings listener ports in HostRules.
 - Issue #611: AKO must not sync service fqdn via External DNS if autoFqdn is disabled.
 - Fix for attaching applicationProfile, datascripts, httpPolicies to Parent VS via HostRule.
 - Set non-significant log duration to infinite, when configuring analyticsPolicy via HostRule.
 - Fix for auth-token renewal after token expiration.

## AKO-1.6.4

### Bugs fixed
 - Problem in creating LoadBalancer Service with named ports.
 - Issue: FQDN aliases not getting added to all the HTTP policies.
 - Fixes improper dedicated VS creation of Service of type LB when Gateways and ServiceLB used at the same time.
 - Fixes an issue of an empty string fqdn programming in L4 VSVIP when autoFqdn is disabled and no subDomains are configured in the dnsProfile.
 - Fixes an issue of SEG label configuration during AviInfraSetting validation if static route sync is disabled.


## AKO-1.7.1

### Added
 - AKO now claims support for Kubernetes 1.23
 - [Multiple AKO instances](docs/multiple-ako.md) can be deployed in K8/Openshift cluster.
 - [Support for Shared VIP with Service of type LoadBalancer](docs/shared_vip.md) (Tech-preview)
 - Multiple certificate support for ingresses/routes through HostRule CRD.
 - Support for PKI profile reference, secrete reference through HostRule CRD.
 - Support for Openshift on Openstack
 - Optimization in nodeport mode using nodefilters.

### Changed
 - Control AKO Event broadcasting using ConfigMap `enableEvents` flag.
 - Allow AKO to continue clean up of avi objects when AKO boots up with `deleteConfig` flag set to true.
 - In EVH deployment, if AKO is processing two hosts, that belongs to same parent virtual service, AKO continues to process the next host even if the current host has errors except if the error code is:
    1. Between 500 to 509
    2. 408, indicating session timeout
    3. 403, Controller upgrade is in progress
    4. 401, invalid credentials
 - Set `Network Profile` to `System-TCP-Proxy` for L4 virtual services if Avi Controller has Enterprise License.

### Fixed
 - Fix: Donot program fqdn for L4 via external dns when autoFQDN is disabled.
 - Fix: Empty fqdn in L4 VSVIP when autoFqdn is disabled.
 - Fix: Dedicated VS creation of service type LB if Gateways and ServiceLB is used at same time.
 - Fix: HTTP rule is not getting applied on a route with empty path.
 - Fix: Ingress fails if client adds port to host header.
 - Fixes security vulnerability caused due to third party package import in AKO.
 - Fix: FQDN aliases not getting added to all the HTTP policies.
 - Fix: AKO is not updating the ingress status when annotation `passthrough.ako.vmware.com/enabled: "true"` is added to the ingress.
 - Fixes LoadBalancer service creation with named ports in NodePortLocal deployment.
 - Fix: Every SEGroup used in the AviInfraSetting is getting configured with the labels even when `disableStaticRouteSync` is set to `true`.
 - Fix: AKO pod keeps getting error "panic: runtime error: slice bounds out of range" then goes into CrashLoopBackOff state.


## AKO-1.7.2

### Added
 - Support for AviInfraSetting CRD for Shared Virtual Service of type LoadBalancer

### Fixed
 - Fix: HTTP Rule will be rejected if `pkiProfile` or `destinationCA` is not defined while defining `tls` section of rule.
 - Fix: L4 Pools, with new naming conventions, will not be attached to L4 VS if LoadBalancer kubernetes services, without annotation `ako.vmware.com/enable-shared-vip`, are migrated from older AKO version to AKO-1.7.1.
 - Fix: VRF context issue when AKO is deployed in NodePort mode for non-admin tenant.
 - Fix: Empty Ingress pool when named ports are used

### Known Issues:
 - `hostrule` with `sslKeyCertificate` of type `secret` will work only in AKO installed namespace in OpenShift clusters.


## AKO-1.8.1

### Added
 - AKO now claims support for Kubernetes 1.24.
 - Support for `AviInfraSetting` CRD in Passthrough Ingresses/Routes.
 - Support for ISTIO mTLS authentication. (Tech-preview)
 - Support for IPv6 networking. (Tech-preview)
 - Added a knob to block the processing of objects in system namespaces in OpenShift and K8s clusters.
 - Support for placement network for VIP in NSX-T VLAN deployments.

### Changed
 - For Insecure Ingress deployments, priority labels associated with pools will be case insensitive.

### Fixed
 - `hostrule` with `sslKeyCertificate` of type `secret` can now be configured in all namespaces in OpenShift clusters.
 - Fixed an issue of VS creation failing with 470 Ingress with a single path giving a 470 error.
 - `nodeNetworkList` is not mandatory for NSX-T Overlay deployments.

## AKO-1.8.2

### Changed
 - Logs, from Istio file watcher, are changed from Error level to Warn level.
 - Status of ingress/route will be populated with IPV6 address in IPV6 deployments instead of IPV4 ips except public IP is configured.

### Fixed
 - Fix: Error in passthrough VS datascript execution when avi-infrasetting is applied to the passthrough ingress/route.
 - Fix: Secure ingress is not working with Istio.
 - Validate enableSSL on listener Port only when listener settings are configured on the hostrule.
 - Fix: Security vulnerabilities in the Golang packages and the base image.


## AKO-1.9.1

### Added
 - AKO now claims support for Kubernetes 1.25, OCP 4.11.
 - AKO can be deployed in [HA mode](docs/ako_ha.md)
 - [Support for SCTP protocol on L4 virtual services](docs/sctp.md)
 - Support for IPV6 networking for Calico and Antrea CNI.
 - Bootup time optimization.

### Changed
 - [Annotation `ako.vmware.com/load-balancer-ip` support](docs/objects.md#service-of-type-loadbalancer-with-preferred-ip) to specify preferred IP for L4 services.

### Fixed
 - Fixed: In nodeport deployment, pool servers are populated with IPV4 IP addresses even if IPFamily is set as `V6`.
 - Fixed: AKO is not cleaning up ISTIO resources even if istio is disabled or `deleteConfig` is set to `true`.
 - Fixed: Virtual service is not getting updated if alternate cert is updated or deleted in hostrule CRD.
 - Fixed: AVI controller requires SSL profile to be set to enable SSL on pool for controller versions lower than 22.x.

### Known Issues
- When AKO is deployed in HA, 409 status code error messages will be seen in active AKO if failover happens during bulk object addition.
- When AKO is running in `dedicated mode`, virtual service and VIP is not deleted from AVI controller when ingress/route, with aviinfrasetting applied to ingress/route, is deleted.