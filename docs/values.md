## Description of tunables of AKO

This document is intended to help the operator make the right choices while deploying AKO with the configurable settings.
The values.yaml in AKO affects a configmap that AKO's deployment reads to make adjustments as per user needs. Listed are detailed
explanation of various fields specified in the values.yaml. If the field is marked as "editable", it means that it can be edited without an AKO POD restart.

### AKOSettings.fullSyncFrequency

This field is used to set a frequency of consitency checks in AKO. Typically inconsistent states can arise if users make changes out
of band w.r.t AKO. For example, a pool is deleted by the user from the UI of the Avi Controller. The full sync frequency is used
to ensure that the models are re-conciled and the corresponding Avi objects are restored to the original state.

### AKOSettings.logLevel *(editable)*

This flag defines the logLevel for logging and can be set to one of `DEBUG`, `INFO`, `WARN`, `ERROR` (case sensitive).
The logLevel value specified here gets populated in the ConfigMap and can be edited at any time while AKO is running. AKO picks up the change in the param value and sets the logLevel at runtime, so AKO pod restart is not required.

### AKOSettings.deleteConfig *(editable)*

This flag is intended to be used for deletion of objects in AVI Controller. The default value is false.
If the value is set to true while booting up, AKO won't process any kubernetes object and stop regular operations.

While AKO is running, this value can be edited to "true" in AKO configmap to delete all abjects created by AKO in AVI.
After that, if the value is set to "false", AKO would resume processing kubernetes objects and recreate all the objects in AVI.

### AKOSettings.disableStaticRouteSync

This flag can be used in 2 scenarios:

* If your POD CIDRs are routable either through an internal implementation or by default.
* If you are working with multiple NICs on your kubernetes worker nodes and the default gateway is not from the same subnet as
your VRF's PG network.

### AKOSettings.clusterName

The `clusterName` field primarily identifies your running AKO instance. AKO internally uses this field to tag all the objects it creates on Avi Controller. All objects created by a particular AKO instance have a prefix of `<clusterName>--` in their names and also populates the `created_by` like so `ako-<clusterName>`.

Each AKO instance mapped to a given Avi cloud should have a unique `clusterName` parameter. This would maintain uniqueness of object naming across Kubernetes clusters.

### AKOSettings.apiServerPort

The `apiServerPort` field is used to run the API server within the AKO pod. The kubernetes API server uses the `/api/status` API to verify the health of the AKO pod on the pod:port where the port is defined by this field. This is configurable, because some enviroments might block usage of the default `8080` port. This field is purely used for AKO's internal API server and must not be confused with a kubernetes pod port.

### AKOSettings.cniPlugin

Use this flag only if you are using `calico`/`openshift` as a CNI and you are looking to a sync your static route configurations automatically.
Once enabled, for `calico` this flag is used to read the `blockaffinity` CRD to determine the POD CIDR to Node IP mappings. If you are
on an older version of calico where `blockaffinity` is not present, then leave this field as blank. For `openshift` hostsubnet CRD is used to to determine the POD CIDR to Node IP mappings.

AKO will then determine the static routes based on the Kubernetes Nodes object as done with other CNIs.

### AKOSettings.layer7Only

Use this flag if you want AKO to act as a pure layer 7 ingress controller. AKO needs to be rebooted for this flag change to take effect. If the configmap is edited while AKO is running, then the change will not take effect. If AKO was working for both L4-L7 prior to this change and then this flag is set to `true`, then AKO will delete the layer 4 LB virtual services from the Avi controller and keep only the Layer 7 virtualservices. If the flag is set to `false` the service of type Loadbalancers would be synced and Layer 4 virtualservices would be created.

### AKOSettings.enableEVH

Use this flag if you want to create Enhanced Virtual Hosting model for Virtual Service objects in AVI. It is disabled by default. Set the flag to `true` to enable the flag. This feature is currently supported for kubernetes clusters only.

Before enabling the flag in the existing deployment make sure to delete the config and enable the flag. This will ensure SNI based VS's are deleted before creating EVH VS's.

### AKOSetttings.namespaceSelector.labelKey and AKOSetttings.namespaceSelector.labelValue

AKO allows ingresses/routes from specific namespace/s to be synced to Avi controller. This key-value pair represent a label that is used by AKO to filter out namespace/s. If one of key/values specified empty, then ingresses/routes from all namespaces will be synched to Avi controller.

### AKOSetttings.servicesAPI

Use this flag to enable AKO to watch over Gateway API CRDs i.e. GatewayClasses and Gateways. AKO only supports Gateway APIs with Layer 4 Services. Setting this to `true` would enable users to configure GatewayClass and Gateway CRs to aggregate multiple Layer 4 Services and create one VirtualService per Gateway Object. 

### NetworkSettings.nodeNetworkList

The `nodeNetworkList` lists the Networks and Node CIDR's where the k8s Nodes are created. This is only used in the ClusterIP deployment of AKO and in vCenter cloud and only when disableStaticRouteSync is set to false.

If two Kubernetes clusters have overlapping POD CIDRs, the service engine needs to identify the right gateway for each of the overlapping CIDR groups. This is achieved by specifying the right placement network for the pools that helps the Service Engine place the pools appropriately.

### NetworkSettings.subnetIP and NetworkSettings.subnetPrefix

AKO supports dual arm deployment where the Virtual IP network can be on a different subnet than the actual Port Groups on which the kubernetes nodes are deployed.

These fields are used to specify the Virtual IP network details on which the user wants to place the Avi virtual services on.

### NetworkSettings.vipNetworkList

List of VIP Networks can be specified through vipNetworkList with key as networkName. Except AWS cloud, for all other cloud types, only one networkName is supported. For example in vipNetworkList:

    vipNetworkList:
      - networkName: net1

For all Public clouds, vipNetworkList must be have at least one networkName. For other cloud types too, it is suggested that networkName should be specified in vipNetworkList. With AVI IPAM, if networkName is not specified in vipNetworkList, an IP can be allocated from the IPAM of the cloud.

In AWS cloud, multiple networkNames are supported in vipNetworkList.


### NetworkSettings.enableRHI

This feature allows the Avi Service Engines to publish the VIP --> SE interface IP mapping to the upstream BGP peers. Using BGP, a virtual service enabled for RHI can be placed on up to 64 SEs within the SE group. Each SE uses RHI to advertise a /32 host route to the virtual service’s VIP address, and is able to accept the traffic. The upstream router uses ECMP to select a path to one of the SEs. Based on this update, the BGP peer connected to the Avi SE updates its route table to use the Avi SE as the next hop for reaching the VIP. The peer BGP router also advertises itself to its upstream BGP peers as a next hop for reaching the VIP.The BGP peer IP addresses, as well as the local Autonomous System (AS) number and a few other settings, are specified in a BGP profile on the Avi Controller.

This feature is available as a global setting in AKO which means if it's set to `true` then it would apply for all virtualservices created by AKO.

Since RHI is a Layer 4 construct, the settings applies to all the host FQDNs patched as pools/SNI virtualservices to the parent shared virtualservice.

#### NetworkSettings.bgpPeerLabels 

This feature allows configuring BGP Peer labels for BGP virtualservices. AKO configures the VSes with the appropriate peer labels, only when `enableRHI` is set to `true`, using the `NetworkSettings.enableRHI` field in `values.yaml`. If `enableRHI` is not set to `true`, AKO will consider the provided configuration as invalid and will reboot.

    bgpPeerLabels:
      - peer1
      - peer2

### L7Settings.shardVSSize

AKO uses a sharding logic for Layer 7 ingress objects. A sharded VS involves hosting multiple insecure or secure ingresses hosted by
one virtual IP or VIP. Having a shared virtual IP allows lesser IP usage since reserving IP addresses particularly in public clouds
incur greater cost.

We support a DEDICATED VIP feature as well per ingress hostname. This feature can be turned out by specifying DEDICATED against
the shardVSSize.

### L7Settings.noPGForSNI

Currently http caching is not available on PoolGroups from the Avi controller. AKO uses poolgroups for canary style deployments. If a user does not require canary deployments and they have an immediate requirement for HTTP caching then this flag can be helpful. Use of this flag is highly discouraged unless required, as it will be deprecated in future once Avi Pool Groups implement HTTP caching in the Avi Controller.

If this flag is set to `true` then AKO would program http policy set rules to switch between pools instead of poolgroups. This feature only applies to secure FQDNs.

### L7Settings.passthroughShardSize

This is applicable only in openshift environment.
AKO uses a sharding logic for passthrough routes, these are distinct from the shared Virtual Services used for Layer 7 ingress or route objects. For all passthrough routes, a set of shared Virtual Services are created. The number of such Virtual Services is controlled by this flag.

### L7Settings.defaultIngController

This field is related to the ingress class support in AKO specified via `kubernetes.io/ingress.class` annotation specified on an
ingress object.

* If AKO is set as the default ingress controller, then it will sync everything other than the ones on which the ingress class is specified and is not equals to “avi”.
* If Avi is not set as the default ingress controller then AKO will sync only those ingresses which have the ingress class set to “avi”.

If you do not use ingress classes, then keep this knob untouched and AKO will take care of syncing all your ingress objects to Avi.

### L4Settings.defaultDomain

If you have multiple sub-domains configured in your Avi cloud, use this knob to specify the default sub-domain.
This is used to generate the FQDN for the Service of type loadbalancer. If unspecified, the behavior works on a sorting logic.
The first sorted sub-domain in chosen, so we recommend using this parameter if you want to be in control of your DNS resolution for service of type LoadBalancer.

### L4Settings.autoFQDN

This knob is used to control how the layer 4 service of type Loadbalancer's FQDN is generated. AKO supports 3 options:

* default: In this case, the FQDN format is <svc-name>.<namespace>.<sub-domain> where the namespace refers to the Service's namespace. sub-domain is picked up from the IPAM DNS profile.

* flat: In this case, the FQDN format is <svc-name>-<namespace>.<sub-domain>

* disabled: In this case, FQDNs are not generated for service of type Loadbalancers.

### ControllerSettings.controllerVersion

This field is used to specify the Avi controller version. While AKO is backward compatible with most of the 18.2.x Avi controllers,
the tested and preferred controller version is `18.2.10`

### ControllerSettings.controllerHost

This field is usually not present in the `value.yaml` by default but can be provided with the `helm` install command to specify
the Avi Controller's IP address or Hostname. If you are using a containerized deployment of the controller, pls use a fully qualified controller
IP address/FQDN. For example, if the controller is hosted on 8443, then controllerHost should: `x.x.x.x:8443`

### ControllerSettings.cloudName

This field is used to specify the name of the IaaS cloud in Avi controller. For example, if you have the VCenter cloud named as "Demo"
then specify the `name` of the cloud name with this field. This helps AKO determine the IaaS cloud to create the service engines on.

### ControllerSettings.tenantsPerCluster

If this field is set to `true`, AKO will map each Kubernetes / OpenShift cluster uniquely to a tenant in AVI.
If enabled, then tenant should be created in AVI to map to a cluster and needs to be specified in `ControllerSettings.tenantName` field.

### ControllerSettings.tenantName

The `tenantName` field  is used to specify the name of the tenant where all the AKO objects will be created in AVI. This field is only required if `tenantsPerCluster` is set to `true`.
The tenant in AVI needs to be created by the AVI controller admin before the AKO bootup.

### ControllerSettings.cloudName

This field is used to specify the name of the IaaS cloud in Avi controller. For example, if you have the VCenter cloud named as "Demo"
then specify the `name` of the cloud name with this field. This helps AKO determine the IaaS cloud to create the service engines on.
<br>

#### AWS and Azure Cloud in NodePort mode of AKO

If the IaaS cloud is Azure then subnet name is specified in `networkName` within vipNetworkList. Azure IaaS cloud is supported only in `NodePort` mode of AKO.
If the IaaS cloud is AWS then subnet uuid is specified in `networkName` within vipNetworkList. AWS IaaS cloud is supported only in `NodePort` mode of AKO.
The `subnetIP` and `subnetPrefix` are not required for AWS and Azure Cloud.

### avicredentials.username and avicredentials.password

The username/password of the Avi controller is specified with this flag. The username/password are base64 encoded by helm and a corresponding secret
object is used to maintain the same. Editing this field requires a restart (delete/re-create) of the AKO pod.

### avicredentials.certificateAuthorityData

This field allows setting the rootCA of the Avi controller, that AKO uses to verify the server certificate provided by the Avi Controller during the TLS handshake. This also enables AKO to connect securely over SSL with the Avi Controller, which is not possible in case the field is not provided.
The field can be set as follows:

    certificateAuthorityData: |-
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----

### image.repository

If you are using a private container registry and you'd like to override the default dockerhub settings, then this field can be edited
with the private registry name.

### L7Settings.serviceType

This option specifies whether the AKO functions in ClusterIP mode or NodePort mode. By default it is set to `ClusterIP`. Allowed values are `ClusterIP`, `NodePort`. If CNI type for the cluster is `antrea`, then another serviceType named `NodePortLocal` is allowed.

### nodeSelectorLabels.key and nodeSelectorLabels.value

It might not be desirable to have all the nodes of a kubernetes cluster to participate in becoming server pool members, hence key/value is used as a label based selection on the nodes in kubernetes to participate in NodePort. If key/value are not specified then all nodes are selected.

### persistentVolumeClaim

By default, AKO prints all the logs to stdout. Instead, persistentVolumeClaim(PVC) can be used for publishing logs of AKO pod to a file in PVC. To use this, the user has to create a PVC (and a persistent volume, if required) and specify the name of the PVC as the value of persistentVolumeClaim.

### podSecurityContext

This can be used to set securityContext of AKO pod, if necessary. For example, in openshift environment, if a persistent storage with hostpath is used for logging, then securityContext must have privileged: true (Reference - https://docs.openshift.com/container-platform/4.4/storage/persistent\_storage/persistent-storage-hostpath.html)
