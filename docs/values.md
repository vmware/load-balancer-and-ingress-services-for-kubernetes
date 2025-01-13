## Description of tunables of AKO

This document is intended to help the operator make the right choices while deploying AKO with the configurable settings.
The values.yaml in AKO affects a configmap that AKO's deployment reads to make adjustments as per user needs. Listed are detailed
explanation of various fields specified in the values.yaml. If the field is marked as "editable", it means that it can be edited without an AKO Pod restart.

### AKOSettings.fullSyncFrequency

This field is used to set a frequency of consitency checks in AKO. Typically inconsistent states can arise if users make changes out
of band w.r.t AKO. For example, a pool is deleted by the user from the UI of the Avi Controller. The full sync frequency is used
to ensure that the models are re-conciled and the corresponding Avi objects are restored to the original state.

### AKOSettings.enableEvents *(editable)*

This flag provides the ability to enable/disable Event broadcasting from AKO. The value specified here gets populated in the ConfigMap and can be edited at any time while AKO is running. AKO picks up the change in the param value and enables/disables Event broadcasting in the cluster at runtime, so AKO pod restart is not required.

### AKOSettings.logLevel *(editable)*

This flag defines the logLevel for logging and can be set to one of `DEBUG`, `INFO`, `WARN`, `ERROR` (case sensitive).
The logLevel value specified here gets populated in the ConfigMap and can be edited at any time while AKO is running. AKO picks up the change in the param value and sets the logLevel at runtime, so AKO pod restart is not required.

### AKOSettings.deleteConfig *(editable)*

This flag is intended to be used for deletion of objects in AVI Controller. The default value is false.
If the value is set to true while booting up, AKO will continue to delete all objects created by AKO in AVI. AKO won't process any kubernetes object and stop regular operations.

While AKO is running, this value can be edited to "true" in AKO configmap to delete all objects created by AKO in AVI.
After that, if the value is set to "false", AKO would resume processing kubernetes objects and recreate all the objects in AVI.

### AKOSettings.disableStaticRouteSync

This flag can be used in 2 scenarios:

* If your Pod CIDRs are routable either through an internal implementation or by default.
* If you are working with multiple NICs on your kubernetes worker nodes and the default gateway is not from the same subnet as
your VRF's PG network.

### AKOSettings.clusterName

The `clusterName` field primarily identifies your running AKO instance. AKO internally uses this field to tag all the objects it creates on Avi Controller. All objects created by a particular AKO instance have a prefix of `<clusterName>--` in their names and also populates the `created_by` like so `ako-<clusterName>`.

Each AKO instance mapped to a given Avi cloud should have a unique `clusterName` parameter. This would maintain uniqueness of object naming across Kubernetes clusters.

### AKOSettings.apiServerPort

The `apiServerPort` field is used to run the API server within the AKO pod. The kubernetes API server uses the `/api/status` API to verify the health of the AKO pod on the pod:port where the port is defined by this field. This is configurable, because some enviroments might block usage of the default `8080` port. This field is purely used for AKO's internal API server and must not be confused with a kubernetes pod port.

### AKOSettings.cniPlugin

Use this flag only if you are using `calico`/`openshift`/`ovn-kubernetes`/`cilium` as a CNI and you are looking to sync your static route configurations automatically.  
However, for `cilium` CNI, setting this flag is only required when using Cluster Scope mode for IPAM. With Cilium CNI, there are two ways to confugure the per-node PodCIDRs. In the **default** cluster scope mode, the podCIDRs range are made available via the `CiliumNode (cilium.io/v2.CiliumNode)` CRD and AKO reads this CRD to determine the Pod CIDR to Node IP mappings when the flag is set as `cilium`. In Kubernetes host scope mode, podCIDRs are allocated out of the PodCIDR range associated to each node by Kubernetes. Since AKO determines the Pod CIDR to Node IP mappings from Node Spec by default, the `cniPlugin` flag is not required to be set exclusively.

Once enabled, for `calico` this flag is used to read the `blockaffinity` CRD to determine the Pod CIDR to Node IP mappings. If you are
on an older version of calico where `blockaffinity` is not present, then leave this field as blank.  
For `openshift` hostsubnet CRD is used to to determine the Pod CIDR to Node IP mappings.  
For `ovn-kubernetes` the `k8s.ovn.org/node-subnets` annotation in the Node metadata is used to determine the Pod CIDR to Node IP mappings.

AKO will then determine the static routes based on the Kubernetes Nodes object as done with other CNIs.  
In case of `ncp` CNI, AKO automatically disables the configuration of static routes.

There are certain scenarios where AKO cannot determine the Pod CIDRs being used in the Kubernetes Nodes, for instance, when deploying calico using `etcd` as the datastore. In such cases AKO provides it's own interface to feed in Pod CIDR to Node mappings, using an annotation in the Node object. While keeping the `cniPlugin` value to be empty, add the following annotation in the Node object to provide Pod CIDRs being used in the Node. Note that for multiple Pod CIDRs that are being used in the Node, simply provide the entries as a comma separated string.

    annotations:
      ako.vmware.com/pod-cidrs: 192.168.1.0/24,192.169.1.0/24

### AKOSettings.layer7Only

Use this flag if you want AKO to act as a pure layer 7 ingress controller. AKO needs to be rebooted for this flag change to take effect. If the configmap is edited while AKO is running, then the change will not take effect. If AKO was working for both L4-L7 prior to this change and then this flag is set to `true`, then AKO will delete the layer 4 LB virtual services from the Avi controller and keep only the Layer 7 virtualservices. If the flag is set to `false` the service of type Loadbalancers would be synced and Layer 4 virtualservices would be created.

### AKOSettings.enableEVH

Use this flag if you want to create Enhanced Virtual Hosting model for Virtual Service objects in AVI. It is disabled by default. Set the flag to `true` to enable the flag. This feature is currently supported for kubernetes clusters only.

Before enabling the flag in the existing deployment make sure to delete the config and enable the flag. This will ensure SNI based VS's are deleted before creating EVH VS's.

### AKOSetttings.namespaceSelector.labelKey and AKOSetttings.namespaceSelector.labelValue

AKO allows ingresses/routes from specific namespace/s to be synced to Avi controller. This key-value pair represent a label that is used by AKO to filter out namespace/s. If one of key/values specified empty, then ingresses/routes from all namespaces will be synched to Avi controller.

### AKOSetttings.servicesAPI

Use this flag to enable AKO to watch over Gateway API CRDs i.e. GatewayClasses and Gateways. AKO only supports Gateway APIs with Layer 4 Services. Setting this to `true` would enable users to configure GatewayClass and Gateway CRs to aggregate multiple Layer 4 Services and create one VirtualService per Gateway Object. 

### AKOSetttings.primaryInstance

Multiple AKO instances can be deployed in a given cluster. This knob is used to specify current AKO instance is primary or not. Setting this to `true` would make current AKO as a primary instance. In a given cluster, there should be only one primary instance. Default value is `true`.

### AKOSettings.blockedNamespaceList

The `blockedNamespaceList` lists the Kubernetes/Openshift namespaces blocked by AKO. AKO will not process any K8s/Openshift object update from these namespaces. Default value is `empty list`.

    blockedNamespaceList:
      - kube-system
      - kube-public

### AKOSetttings.istioEnabled (Tech Preview)

AKO can be deployed in Istio environment. Setting this to `true` indicates to AKO that the environment is Istio. Default value is `false`.

### AKOSetttings.ipFamily (Tech Preview)

`V6` is currently supported only for `vCenter` cloud with `calico` CNI.

AKO can be deployed with ipFamily as `V4` or `V6`. When ipFamily is set to `V6`, AKO looks for `V6` IP for nodes from calico annotation and creates routes on controller. Only servers with `V6` IP will get added to Pools.

Default value is `V4`.

### AKOSettings.useDefaultSecretsOnly

This flag provides the ability to restrict the secret handling to default secrets present in the namespace where the AKO is installed. This flag is applicable only to Openshift clusters.
Default value is `false`.

### NetworkSettings.nodeNetworkList

The `nodeNetworkList` lists the Networks (specified using either `networkName` or `networkUUID`) and Node CIDR's where the k8s Nodes are created. This is only used in vCenter cloud and only when disableStaticRouteSync is set to false.

If two Kubernetes clusters have overlapping Pod CIDRs, the service engine needs to identify the right gateway for each of the overlapping CIDR groups. This is achieved by specifying the right placement network for the pools that helps the Service Engine place the pools appropriately.

### NetworkSettings.subnetIP and NetworkSettings.subnetPrefix

AKO 1.5.1 deprecates `subnetIP` and `subnetPrefix`. See [Upgrade Notes](./upgrade/upgrade.md) for more details.

### NetworkSettings.vipNetworkList

List of VIP Networks can be specified through vipNetworkList with key as `networkName` or `networkUUID`. Except AWS cloud, for all other cloud types, only one networkName is supported. For example in vipNetworkList:

    vipNetworkList:
      - networkName: net1

or

    vipNetworkList:
      - networkUUID: dvportgroup-4167-cloud-d4b24fc7-a435-408d-af9f-150229a6fea6f

In addition to the `networkName` or `networkUUID`, we can also provide CIDR information that allows us to specify the Virtual IP network details on which the user wants to place the Avi virtual services on.

    vipNetworkLists:
      - networkName: net1
        cidr: 10.1.1.0/24
        v6cidr: 2002::1234:abcd:ffff:c0a8:101/64

or

    vipNetworkLists:
      - networkUUID: dvportgroup-4167-cloud-d4b24fc7-a435-408d-af9f-150229a6fea6f
        cidr: 10.1.1.0/24
        v6cidr: 2002::1234:abcd:ffff:c0a8:101/64

`v6cidr` may only work for Enterprise license with AVI controller. We can provide either `cidr` or `v6cidr` or both.

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

#### NetworkSettings.nsxtT1LR 

This knob is used to specify the T1 logical router's name in the format of `/infra/tier-1s/<name-of-t1>`.
This T1 router with a logical segment must be pre-configured in the NSX-T cloud as a `data network segment`. AKO uses this information to populate the virtualservice's and pool's T1Lr attribute.

#### NetworkSettings.defaultDomain

The defaultDomain flag has two use cases.
For **L4** VSes, if multiple sub-domains are configured in the cloud, this flag can be used to set the default sub-domain to use for the VS. This is used to generate the FQDN for the Service of type loadbalancer. If unspecified, the behavior works on a sorting logic. The first sorted sub-domain in chosen, so we recommend using this parameter if you want to be in control of your DNS resolution for service of type LoadBalancer.  
This flag should be used instead of [L4Settings.defaultDomain](#L4SettingsdefaultDomain), as it will be deprecated in a future release.
If both `NetworkSettings.defaultDomain` and `L4Settings.defaultDomain` are set, then `NetworkSettings.defaultDomain` will be used.  
For **L7** VSes(created from OpenShift Routes), if `spec.subdomain` field is specified instead of `spec.host` field for an OpenShift route, then the default domain specified is appended to the `spec.subdomain` to form the FQDN for the VS. The **defaultDomain** should be configured as a sub-domain in your Avi cloud.

    defaultDomain: "avi.internal"

For example, if `spec.subdomain` for an OpenShift route is **my_route-my_namespace** and `defaultDomain` is specified as **avi.internal**, then FQDN for the L7 VS will be **my_route-my_namespace.avi.internal**.


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

AKO uses a sharding logic for passthrough hosts in routes or ingresses. These are distinct from the shared Virtual Services used for Layer 7 ingress or route objects. For all passthrough routes or ingresses, a set of shared Virtual Services are created. The number of such Virtual Services is controlled by this flag.

### L7Settings.defaultIngController

This field is related to the ingress class support in AKO specified via `kubernetes.io/ingress.class` annotation specified on an
ingress object.

* If AKO is set as the default ingress controller, then it will sync everything other than the ones on which the ingress class is specified and is not equals to “avi”.
* If Avi is not set as the default ingress controller then AKO will sync only those ingresses which have the ingress class set to “avi”.

If you do not use ingress classes, then keep this knob untouched and AKO will take care of syncing all your ingress objects to Avi.

### L7Settings.fqdnReusePolicy

This field is used to restrict or allow FQDN to be spanned across multiple namespaces.

* InterNamespaceAllowed: With this value, AKO will allow hostname/FQDN to be associate with Ingresses/Routes which are spanned across multiple namespaces.

* Strict: With this value, AKO will restrict hostname/FQDN to be associated with Ingresses/Routes, present in the same namespace.

### L4Settings.defaultDomain

If you have multiple sub-domains configured in your Avi cloud, use this knob to specify the default sub-domain.
This is used to generate the FQDN for the Service of type loadbalancer. If unspecified, the behavior works on a sorting logic.
The first sorted sub-domain in chosen, so we recommend using this parameter if you want to be in control of your DNS resolution for service of type LoadBalancer.

**Note:** This flag will be deprecated in a future release; use [NetworkSettings.defaultDomain](#NetworkSettingsdefaultDomain) instead. If both NetworkSettings.defaultDomain and L4Settings.defaultDomain are set, then NetworkSettings.defaultDomain will be used.

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

### ControllerSettings.tenantName

The `tenantName` field  is used to specify the name of the tenant where all the AKO objects will be created in AVI. The tenant in AVI needs to be created by the AVI controller admin before the AKO bootup.

### ControllerSettings.cloudName

This field is used to specify the name of the IaaS cloud in Avi controller. For example, if you have the VCenter cloud named as "Demo"
then specify the `name` of the cloud name with this field. This helps AKO determine the IaaS cloud to create the service engines on.

### ControllerSettings.vrfName

The `vrfName` field  is used to specify the name of the VRFContext where all the AKO objects will be created. The VRFContext in AVI needs to be created by the AVI controller admin before the AKO bootsup. This is applicable in VCenter cloud only.
<br>

#### AWS and Azure Cloud in NodePort mode of AKO

If the IaaS cloud is Azure then subnet name is specified in `networkName` within `vipNetworkList`. Azure IaaS cloud is supported only in `NodePort` mode of AKO.
If the IaaS cloud is AWS then subnet uuid is specified in `networkName` within `vipNetworkList`. AWS IaaS cloud is supported only in `NodePort` mode of AKO.
The `cidr` within `vipNetworkList` is not required for AWS and Azure Cloud.

### avicredentials.username and avicredentials.password

The username/password of the Avi controller is specified with this flag. The username/password are base64 encoded by helm and a corresponding secret
object is used to maintain the same. Editing this field requires a restart (delete/re-create) of the AKO pod.

### avicredentials.authtoken

The generated authtoken from the Avi controller can be specified with this flag as an alternative to password. The authtoken is also base64 encoded
and maintained by secret object. The token refresh is managed by AKO. In case of token refresh failure, a new token needs to be generated and updated
into the secret object.

A few ways to properly encode a token generated from controller to directly patch `avi-secret`
1. Shell

```
echo -n '<authtoken>' | base64
```

2. Python

```
import base64
authtoken = "<authtoken>"
print(base64.b64encode(authtoken.encode("ascii")))
```

**Note:** From release v1.12.1 onwards, AKO supports reading Avi Controller credentials including `certificateAuthorityData` from existing `avi-secret` from the namespace in which AKO is installed. If `username` and either `password` or `authtoken` are not specified, avi-secret will not be created as part of Helm installation. AKO will assume that avi-secret already exists in the namespace in which the AKO Helm release is installed and will reference it. 

### avicredentials.certificateAuthorityData

This field allows setting the rootCA of the Avi controller, that AKO uses to verify the server certificate provided by the Avi Controller during the TLS handshake. This also enables AKO to connect securely over SSL with the Avi Controller, which is not possible in case the field is not provided.
The field can be set as follows:

    certificateAuthorityData: |-
      -----BEGIN CERTIFICATE-----
      ...
      -----END CERTIFICATE-----

### replicaCount
This option specifies the number of replicas of the AKO pod.
**Note:** From release v1.9.1 onwards, two instances of AKO are supported.

One AKO runs in active mode, and the second in passive mode. The AKO, which is running in passive mode, will be ready to take over once the active AKO goes down.

### image.repository

If you are using a private container registry and you'd like to override the default dockerhub settings, then this field can be edited
with the private registry name.

### image.pullSecrets

If you are setting the [image.repository](#imagerepository) field to use a secure private container image registry for ako image, then you must specify the pull secrets in this field. The pull secrets are a list of Kubernetes Secret objects that are created from the login credentials of a secure private image registry. The container runtime uses the pull secrets to authenticate with the registry in order to pull the ako image. The image pull secrets must be created in the `avi-system` namespace before deploying AKO.

    pullSecrets:
    - name: regcred

### L7Settings.serviceType

This option specifies whether the AKO functions in ClusterIP mode or NodePort mode. By default it is set to `ClusterIP`. Allowed values are `ClusterIP`, `NodePort`. If CNI type for the cluster is `antrea`, then another serviceType named `NodePortLocal` is allowed.

### nodeSelectorLabels.key and nodeSelectorLabels.value

It might not be desirable to have all the nodes of a kubernetes cluster to participate in becoming server pool members, hence key/value is used as a label based selection on the nodes in kubernetes to participate in NodePort. If key/value are not specified then all nodes are selected.

### persistentVolumeClaim

By default, AKO prints all the logs to stdout. Instead, persistentVolumeClaim(PVC) can be used for publishing logs of AKO pod to a file in PVC. To use this, the user has to create a PVC (and a persistent volume, if required) and specify the name of the PVC as the value of persistentVolumeClaim.

### securityContext

SecurityContext holds security configuration that will be applied to the AKO pod. Some fields are present in both SecurityContext and PodSecurityContext. When both are set, the values in SecurityContext take precedence.(Reference - https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#securitycontext-v1-core)

### podSecurityContext

This can be used to set securityContext of AKO pod, if necessary. For example, in openshift environment, if a persistent storage with hostpath is used for logging, then securityContext must have privileged: true (Reference - https://docs.openshift.com/container-platform/4.11/storage/persistent_storage/persistent-storage-hostpath.html)


### featureGates.GatewayAPI

Use this flag if you want to enable Gateway API feature for AKO. It is disabled by default. Set the flag to `true` to enable the flag.

### GatewayAPI

Enable Gateway API in the featureGate to use this field.

### GatewayAPI.image.repository

If you are using a private container registry and you'd like to override the default dockerhub settings, then this field can be edited with the private registry name.

### featureGates.EnableEndpointSlice

Enable this flag to use EndpointSlices instead of Endpoints in AKO. This also supports graceful shutdown of servers. Enabled by default from AKO 1.13.1.
