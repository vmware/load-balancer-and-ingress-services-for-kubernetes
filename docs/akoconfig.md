## AKOConfig Custom Resource

The `AKOConfig` custom resource is used to deploy and manage the AKO controller and is meant to be consumed by the AKO operator. The schema for the `AKOConfig` custom resource can be found [here](../ako-operator/bundle/manifests/ako.vmware.com_akoconfigs.yaml). This is what a sample AKOConfig looks like:
```yaml
apiVersion: ako.vmware.com/v1beta1
kind: AKOConfig
metadata:
  finalizers:
  - ako.vmware.com/cleanup
  name: ako-sample
  namespace: avi-system
spec:
  replicaCount: 1
  imageRepository: projects.packages.broadcom.com/ako/ako:1.13.2
  imagePullPolicy: "IfNotPresent"
  imagePullSecrets:
    - name: regcred
  akoSettings:
    enableEvents: true
    logLevel: "WARN"
    fullSyncFrequency: "1800"
    apiServerPort: 8080 
    deleteConfig: false
    disableStaticRouteSync: false
    clusterName: "my-cluster"
    cniPlugin: ""
    enableEVH: false
    layer7Only: false
    namespaceSelector:
      labelKey: ""
      labelValue: ""
    servicesAPI: false
    vipPerNamespace: false
    istioEnabled: false
    ipFamily: ""
    blockedNamespaceList: []
    useDefaultSecretsOnly: false
    vpcMode: false

  networkSettings:
    nodeNetworkList: []
    enableRHI: false
    nsxtT1LR: ""
    bgpPeerLabels: []
    vipNetworkList:
     - networkName: net1
       cidr: 100.1.1.0/24
       v6Cidr: 2002::1234:abcd:ffff:c0a8:101/64
    defaultDomain: ""

  l7Settings:
    defaultIngController: true
    noPGForSNI: false
    serviceType: ClusterIP
    shardVSSize: "LARGE"
    passthroughShardSize: "SMALL"
    fqdnReusePolicy: "InterNamespaceAllowed"

  l4Settings:
    defaultDomain: ""
    autoFQDN: "default"
    defaultLBController: true

  controllerSettings:
    serviceEngineGroupName: "Default-Group"
    controllerVersion: ""
    cloudName: "Default-Cloud"
    controllerIP: ""
    tenantName: "admin"
    vrfName: ""

  nodePortSelector:
    key: ""
    value: ""

  resources:
    limits:
      cpu: "350m"
      memory: "400Mi"
    requests:
      cpu: "200m"
      memory: "300Mi"

  pvc: ""
  mountPath: "/log"
  logFile: "avi.log"
  featureGates:
    gatewayAPI: true
    enablePrometheus: false
    enableEndpointSlice: true
  gatewayAPI:
    image:
      repository: "projects.packages.broadcom.com/ako/ako-gateway-api:1.13.2"
      pullPolicy: "IfNotPresent"
  akoGatewayLogFile: "avi-gw.log"
  ```

### The fields comprising the AKOConfig custom resource are as follows:

  - `metadata.finalizers`: Used for garbage collection. This field must have this element: `ako.vmware.com/cleanup`. Whenever, the AKOConfig object is deleted, the ako operator takes care of removing all the AKO related artifacts because of this finalizer.
  - `metadata.name`: Name of the AKOConfig object.
  - `metadata.namespace`: The namespace in which the AKOConfig object (and hence, the ako-operator) will be created. Only `avi-system` namespace is allowed for the ako-operator.
  - `spec.imageRepository`: The image repository for the ako controller.
  - `spec.imagePullPolicy`: The image pull policy for the AKO controller image. It defines when the image gets pulled.
  - `spec.imagePullSecrets`: ImagePullSecrets will add pull secrets to the statefulset for AKO. Required if using secure private container image registry for AKO image.
  - `spec.replicaCount`: The number of replicas for AKO StatefulSet.
  - `spec.akoSettings`: Settings for the AKO Controller.
    * `enableEvents`: Enables/disables Event broadcasting via AKO 
    * `logLevel`: Log level for the AKO controller. Supported enum values: `INFO`, `DEBUG`, `WARN`, `ERROR`.
    * `fullSyncFrequency`: Interval at which the AKO controller does a full sync of all the objects.
    * `apiServerPort`: The port at which the AKO API Server is available.
    * `deleteConfig`: Set to true if user wants to delete AKO created objects from Avi. Default value is `false`.
    * `disableStaticRouteSync`: Disables static route syncing if set to `true`. Default value is `false`.
    * `clusterName`: Unique identifier for the running AKO controller instance. The AKO controller identifies objects, which it created on Avi Controller using the `clusterName` param.
    * `cniPlugin`: The CNI plugin to be used in Openshift cluster. Specify one of: `openshift`, `ovn-kubernetes`.
    * `enableEVH`: This enables the Enhanced Virtual Hosting Model in Avi Controller for the Virtual Services
    * `layer7Only`: If this flag is switched on, then AKO will only do layer 7 loadbalancing.
    * `namespaceSelector.labelKey`: Set the key of a namespace's label, if the requirement is to sync k8s objects from that namespace.
    * `namespaceSelector.labelValue`: Set the value of a namespace's label, if the requirement is to sync k8s objects from that namespace.
    * `servicesAPI`: Flag that enables AKO in services API mode: https://kubernetes-sigs.github.io/service-apis/. Currently implemented only for L4. This flag uses the upstream GA APIs which are not backward compatible with the advancedL4 APIs which uses a fork and a version of v1alpha1pre1
    * `vipPerNamespace`: Enabling this flag would tell AKO to create Parent VS per Namespace in EVH mode
    * `istioEnabled`: This flag needs to be enabled when AKO is be to brought up in an Istio environment.
    * `ipFamily`: IPFamily specifies IP family to be used. This flag can take values `V4` or `V6` (default `V4`). This is for the backend pools to use ipv6 or ipv4. For frontside VS, use v6cidr
    * `blockedNamespaceList`: This is the list of system namespaces from which AKO will not listen any Kubernetes or Openshift object event.
    * `useDefaultSecretsOnly`: If this flag is set to true, AKO will only handle default secrets from the namespace where AKO is installed. This flag is applicable only to Openshift clusters.
    * `vpcMode`: VPCMode enables AKO to operate in VPC mode. This flag is only applicable to NSX-T.
  - `networkSettings`: Data network setting
    * `nodeNetworkList`: This list of Network Names/UUIDs and Cidrs are used in pool placement network for vcenter cloud. Either networkName or networkUUID should be specified. If duplicate networks are present for the network name, networkUUID should be used for appropriate network. Node Network details are not needed when in nodeport mode / static routes are disabled / non vcenter clouds.
    * `enableRHI`: This is a cluster wide setting for BGP peering.
    * `nsxtT1LR`: Unique ID (note: not display name) of the T1 Logical Router for Service Engine connectivity. Only applies to NSX-T cloud.
    * `bgpPeerLabels`: Select BGP peers using bgpPeerLabels, for selective VsVip advertisement.
    * `vipNetworkList`: List of Network Names/UUIDs and Subnet Information for VIP network, multiple networks allowed only for AWS Cloud. Either networkName or networkUUID should be specified. If duplicate networks are present for the network name, networkUUID should be used for appropriate network.
    * `defaultDomain`: The defaultDomain flag has two use cases. For L4 VSes, if multiple sub-domains are configured in the cloud, this flag can be used to set the default sub-domain to use for the VS. This flag should be used instead of L4Settings.defaultDomain, as it will be deprecated in a future release. If both NetworkSettings.defaultDomain and L4Settings.defaultDomain are set, then NetworkSettings.defaultDomain will be used. For L7 VSes(created from OpenShift Routes), if spec.subdomain field is specified instead of spec.host field for an OpenShift route, then the default domain specified is appended to the spec.subdomain to form the FQDN for the VS. The defaultDomain should be configured as a sub-domain in Avi cloud.
  - `l7Settings`: Settings for L7 Virtual Services
    * `defaultIngController`: Set to `true` if AKO controller is the default Ingress controller on the cluster.
    * `noPGForSNI`: Switching this knob to true, will get rid of poolgroups from SNI VSes. Do not use this flag, if you don't want http caching. This will be deprecated once the controller support caching on PGs.
    * `serviceType`: Type of services that we want to configure: Valid values: `ClusterIP` and `NodePort`.
    * `shardVSSize`: Use this to control the Avi Virtual service numbers. This applies to both secure/insecure VSes but does not apply for passthrough. Valud values: `LARGE`, `MEDIUM` and `SMALL`.
    * `passthroughShardSize`: Use this to control the passthrough virtualservice numbers. Valid values: `LARGE`, `MEDIUM` and `SMALL`.
    * `fqdnReusePolicy`: This flag can be used to control whether AKO allows cross-namespace usage of FQDNs. Valid values: `InterNamespaceAllowed` and `Strict`.
  - `l4Settings`: Settings for L4 Virtual Services
    * `autoFQDN`: Specify the layer 4 FQDN format | Valid values: `default`, `flat` and `disabled`. Defaults to `default`.
    * `defaultDomain`: If multiple sub-domains are configured in the cloud, use this knob to set the default sub-domain to use for L4 VSes. This flag will be deprecated in a future release; use networkSettings.defaultDomain instead. If both networkSettings.defaultDomain and l4Settings.defaultDomain are set, then networkSettings.defaultDomain will be used.
    * `defaultLBController`: DefaultLBController enables ako to check if it is the default LoadBalancer controller. Set to `true` if AKO controller is the default LoadBalancer controller on the cluster.
  - `controllerSettings`: Settings for the AVI Controller
    * `serviceEngineGroupName`: Name of the service engine group.
    * `controllerVersion`: The controller API version.
    * `cloudName`: The configured cloud name on the AVI controller.
    * `controllerIP`: The IP Address (URL) of the AVI Controller.
    * `tenantName`: Name of the tenant where the AVI controller will create objects in AVI.
    * `vrfName`: Name of the vrfContext present. All AKO created objects, static routes will be associated with this VRF Context.
  - `nodePortSelector`: Only applicable if `l7Settings.serviceType` is set to `NodePort`.
    * `key`
    * `value`
  - `resources`: Specify the resources for the AKO Controller's statefulset.
  - `pvc`: Persistent Volume Claim name which AKO controller will use to store its logs.
  - `mountPath`: Mount path for the logs.
  - `logFile`: Log file name where the AKO controller will add it's logs.
  - `featureGates`: FeatureGates is to enable or disable experimental features.
    * `enableEndpointSlice`: Enables/Disables processing of EndpointSlices instead of Endpoints. Defaults to `true`.
    * `gatewayAPI`: GatewayAPI enables/disables processing of Kubernetes Gateway API CRDs. Defaults to `false`.
    * `enablePrometheus`: EnablePrometheus enables/disables prometheus scraping for AKO container. Defaults to `false`.
  - `gatewayAPI`: GatewayAPI defines settings for AKO Gateway API container. These settings will only be used if **gatewayAPI** feature gate is enabled.
    * `image`: Image defines image related settings for AKO Gateway API container.
  - `akoGatewayLogFile`: AKOGatewayLogFile is the name of the file where ako-gateway-api container will dump its logs. This setting will only be used if **gatewayAPI** feature gate is enabled.

  ## Editing the AKOConfig custom resource
  If we need any changes in the way the AKO controller was deployed, or if we want to tweak a knob in the above list, we can do that in the runtime. However, note that, only `spec.akoSettings.logLevel` and `spec.akoSettings.deleteConfig` can be changed without triggering a restart of the AKO controller. If any other knobs are changed, the ako-operator WILL trigger a restart of the AKO controller.

  ## Upgrading the AKOConfig custom resource

  The **v1alpha1** version of the AKOConfig CRD is deprecated in AKO Operator 1.13.2. Please use the latest **v1beta1** version. On uprgade to AKO Operator 1.13.2 version, all the resources including AKOConfig CRD, AKO statefulset, configmap, serviceaccount, gatewayclass, clusterrole and clusterrolebinding, and ako managed CRD's will be automatically upgraded. In case alreday running an AKOConfig v1alpha1 version, please update the existing akoconfig object to use v1beta1 version once the AKO operator is upgraded to 1.13.2 version. The following fields have been removed from the AKOConfig CRD between AKO Operater versions 1.12.3 and 1.13.2 (or CRD versions v1alpha1 and v1beta1).

  - `spec.controllerSettings.tenantsPerCluster`
  - `spec.l7Settings.syncNamespace`
  - `spec.l4Settings.advancedL4`
  - `spec.rbac`

  These fields were already deprecated in previous versions of AKO Operator and should ideally not be in use. If v1alpha1 object is used with these fields, then the fields will be pruned before storing in etcd since v1beta1 is the stored version. Also, when the operator reads the AKOConfig object from the API server, it reads and processes it as a v1beta1 object. Therefore, from the operator's perspective, the removed fields are effectively ignored or "lost" as it processes the object.  
  For AKOConfig objects existing before operator upgrade, any subsequent write operation by the operator after upgrade(e.g., updating the status of the AKOConfig object), will use the v1beta1 version. When this updated object is sent back to the API server, only the fields known to the v1beta1 schema will be included. This action will overwrite the stored object in etcd, permanently removing any previously "unknown" fields that were part of the v1alpha1 manifest but not the v1beta1 schema.

  To ensure proper functionality and avoid unexpected behavior or data loss, you should:

  Get your existing AKOConfig:

  ```bash
  kubectl get akoconfig ako-sample -n avi-system -o yaml > akoconfig_v1alpha1_backup.yaml
  ```

  Manually update the API version and schema: Edit akoconfig_v1alpha1_backup.yaml.  
  Change apiVersion: ako.vmware.com/v1alpha1 to apiVersion: ako.vmware.com/v1beta1.  
  Remove any fields that are no longer supported in v1beta1 (refer to the above list for removed fields, though the operator will prune them anyway).
  Add any new fields introduced in AKOConfig CRD, with newer AKO Operator version (refer to [this](#the-fields-comprising-the-akoconfig-custom-resource-are-as-follows) for list of available fields for v1beta1 version).
  Apply the updated object:

  ```bash
  kubectl apply -f akoconfig_v1alpha1_backup.yaml
  ```

  This will update the existing AKOConfig object to the v1beta1 API version, ensuring it aligns with the operator's expectations.