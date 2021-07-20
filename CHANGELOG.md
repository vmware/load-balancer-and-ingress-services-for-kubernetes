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
 
