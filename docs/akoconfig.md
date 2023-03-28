## AKOConfig Custom Resource

The `AKOConfig` custom resource is used to deploy and manage the AKO controller and is meant to be consumed by the AKO operator. This is what a sample AKOConfig looks like:
```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: AKOConfig
metadata:
  finalizers:
  - ako.vmware.com/cleanup
  name: ako-sample
  namespace: avi-system
spec:
  replicaCount: 1
  imageRepository: projects.registry.vmware.com/ako/ako:1.9.1
  imagePullPolicy: "IfNotPresent"
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

  networkSettings:
    nodeNetworkList: []
    enableRHI: false
    nsxtT1LR: ""
    bgpPeerLabels: []
    vipNetworkList:
     - networkName: net1
       cidr: 100.1.1.0/24
       v6Cidr: 2002::1234:abcd:ffff:c0a8:101/64

  l7Settings:
    defaultIngController: true
    noPGForSNI: false
    serviceType: ClusterIP
    shardVSSize: "LARGE"
    passthroughShardSize: "SMALL"

  l4Settings:
    defaultDomain: ""
    autoFQDN: "default"

  controllerSettings:
    serviceEngineGroupName: "Default-Group"
    controllerVersion: ""
    cloudName: "Default-Cloud"
    controllerIP: ""
    tenantName: "admin"

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

  rbac:
    pspEnable: false

  pvc: ""
  mountPath: "/log"
  logFile: "avi.log"
  ```

  - `metadata.finalizers`: Used for garbage collection. This field must have this element: `ako.vmware.com/cleanup`. Whenever, the AKOConfig object is deleted, the ako operator takes care of removing all the AKO related artifacts because of this finalizer.
  - `metadata.name`: Name of the AKOConfig object. With `helm install`, the name of the default AKOConfig object is `ako-config`.
  - `metadata.namespace`: The namespace in which the AKOConfig object (and hence, the ako-operator) will be created. Only `avi-system` namespace is allowed for the ako-operator.
  - `spec.imageRepository`: The image repository for the ako-operator.
  - `spec.replicaCount`: The number of replicas for AKO StatefulSet
  - `spec.akoSettings`: Settings for the AKO Controller.
    * `enableEvents`: Enables/disables Event broadcasting via AKO 
    * `logLevel`: Log level for the AKO controller. Supported enum values: `INFO`, `DEBUG`, `WARN`, `ERROR`.
    * `fullSyncFrequency`: Interval at which the AKO controller does a full sync of all the objects.
    * `apiServerPort`: The port at which the AKO API Server is available.
    * `deleteConfig`: Set to true if user wants to delete AKO created objects from Avi. Default value is `false`.
    * `disableStaticRouteSync`: Disables static route syncing if set to `true`. Default value is `false`.
    * `clusterName`: Unique identifier for the running AKO controller instance. The AKO controller identifies objects, which it created on Avi Controller using the `clusterName` param.
    * `cniPlugin`: Set the string if your CNI is calico or openshift. Specify one of: `calico`, `canal`, `flannel`, `openshift`, `antrea`, `ncp`.
    * `enableEVH`: This enables the Enhanced Virtual Hosting Model in Avi Controller for the Virtual Services
    * `layer7Only`: If this flag is switched on, then AKO will only do layer 7 loadbalancing.
    * `namespaceSelector.labelKey`: Set the key of a namespace's label, if the requirement is to sync k8s objects from that namespace.
    * `namespaceSelector.labelValue`: Set the value of a namespace's label, if the requirement is to sync k8s objects from that namespace.
    * `servicesAPI`: Flag that enables AKO in services API mode: https://kubernetes-sigs.github.io/service-apis/. Currently implemented only for L4. This flag uses the upstream GA APIs which are not backward compatible with the advancedL4 APIs which uses a fork and a version of v1alpha1pre1
    * `vipPerNamespace`: Enabling this flag would tell AKO to create Parent VS per Namespace in EVH mode
    * `istioEnabled`: This flag needs to be enabled when AKO is be to brought up in an Istio environment.
    * `ipFamily`: IPFamily specifies IP family to be used. This flag can take values `V4` or `V6` (default `V4`). This is for the backend pools to use ipv6 or ipv4. For frontside VS, use v6cidr
    * `blockedNamespaceList`: This is the list of system namespaces from which AKO will not listen any Kubernetes or Openshift object event.
  - `networkSettings`: Data network setting
    * `nodeNetworkList`: This list of network and cidrs are used in pool placement network for vcenter cloud. Node Network details are not needed when in nodeport mode / static routes are disabled / non vcenter clouds.
    * `enableRHI`: This is a cluster wide setting for BGP peering.
    * `nsxtT1LR`: T1 Logical Segment mapping for backend network. Only applies to NSX-T cloud.
    * `bgpPeerLabels`: Select BGP peers using bgpPeerLabels, for selective VsVip advertisement.
    * `vipNetworkList`: List of Network Names and Subnet Information for VIP network, multiple networks allowed only for AWS Cloud.
  - `l7Settings`: Settings for L7 Virtual Services
    * `defaultIngController`: Set to `true` if AKO controller is the default Ingress controller on the cluster.
    * `noPGForSNI`: Switching this knob to true, will get rid of poolgroups from SNI VSes. Do not use this flag, if you don't want http caching. This will be deprecated once the controller support caching on PGs.
    * `serviceType`: Type of services that we want to configure: Valid values: `ClusterIP` and `NodePort`.
    * `shardVSSize`: Use this to control the Avi Virtual service numbers. This applies to both secure/insecure VSes but does not apply for passthrough. Valud values: `LARGE`, `MEDIUM` and `SMALL`.
    * `passthroughShardSize`: Use this to control the passthrough virtualservice numbers. Valid values: `LARGE`, `MEDIUM` and `SMALL`.
  - `l4Settings`: Settings for L4 Virtual Services
    * `advancedL4`: Knob to control the settings for the services API usage. Defaults to `false`.
    * `defaultDomain`: If multiple sub-domains are configured in the cloud, use this knob to set the default sub-domain to use for L4 Virtual Services.
  - `controllerSettings`: Settings for the AVI Controller
    * `serviceEngineGroupName`: Name of the service engine group.
    * `controllerVersion`: The controller API version.
    * `cloudName`: The configured cloud name on the AVI controller.
    * `controllerIP`: The IP Address (URL) of the AVI Controller.
    * `tenantName`: Name of the tenant where the AKO controller will create objects in AVI.
  - `nodePortSelector`: Only applicable if `l7Settings.serviceType` is set to `NodePort`.
    * `key`
    * `value`
  - `resources`: Specify the resources for the AKO Controller's statefulset.
  - `rbac`: Enable a pod security policy for the AKO Controller.
    * `pspEnable`: Set to `true` to create a pod security policy for the AKO controller's statefulset.
  - `pvc`: Persistent Volume Claim name which AKO controller will use to store its logs.
  - `mountPath`: Mount path for the logs.
  - `logFile`: Log file name where the AKO controller will add it's logs.

  ## Editing the AKOConfig custom resource
  If we need any changes in the way the AKO controller was deployed, or if we want to tweak a knob in the above list, we can do that in the runtime. However, note that, only `spec.akoSettings.logLevel` and `spec.akoSettings.deleteConfig` can be changed without triggering a restart of the AKO controller. If any other knobs are changed, the ako-operator WILL trigger a restart of the AKO controller.
