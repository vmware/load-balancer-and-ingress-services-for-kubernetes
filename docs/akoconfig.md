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
  imageRepository: "ako:latest"
  imagePullPolicy: "IfNotPresent"
  akoSettings:
    logLevel: "WARN"
    fullSyncFrequency: "1800"
    apiServerPort: 8080
    deleteConfig: false
    disableStaticRouteSync: false
    clusterName: "k8s-cluster"
    cniPlugin: "antrea"
    namespaceSelector:
      labelKey: ""
      labelValue: ""

  networkSettings:
    subnetIP: "10.10.0.0"
    subnetPrefix: "16"
    vipNetworkList:
       - networkName: "vcd-ns"

  l7Settings:
    defaultIngController: true
    serviceType: "ClusterIP"
    shardVSSize: "LARGE" #enum
    passthroughShardSize: "SMALL"   #enum

  l4Settings:
    advancedL4: false
    defaultDomain: ""

  controllerSettings:
    serviceEngineGroupName: "Default-Group"
    controllerVersion: "20.1.2"
    cloudName: "Default-Cloud"
    controllerIP: "10.10.10.11"
    tenantsPerCluster: false
    tenantName: ""

  nodePortSelector: # only applicable if servicetype is nodePort
    key: ""
    value: ""

  resources:
    limits:
      cpu: "250m"
      memory: "300Mi"
    requests:
      cpu: "100m"
      memory: "200Mi"

  podSecurityContext: {}

  rbac:
    pspEnable: false

  service:
    type: "ClusterIP"
    port: 80

  mountPath: "/log"
  logFile: "avi.log"
  ```

  - `metadata.finalizers`: Used for garbage collection. This field must have this element: `ako.vmware.com/cleanup`. Whenever, the AKOConfig object is deleted, the ako operator takes care of removing all the AKO related artifacts because of this finalizer.
  - `metadata.name`: Name of the AKOConfig object. With `helm install`, the name of the default AKOConfig object is `avi-config`.
  - `metadata.namespace`: The namespace in which the AKOConfig object (and hence, the ako-operator) will be created. Only `avi-system` namespace is allowed for the ako-operator.
  - `spec.imageRepository`: The image repository for the ako-operator.
  - `spec.akoSettings`: Settings for the AKO Controller.
    * `logLevel`: Log level for the AKO controller. Supported enum values: `INFO`, `DEBUG`, `WARN`, `ERROR`.
    * `fullSyncFrequency`: Interval at which the AKO controller does a full sync of all the objects.
    * `apiServerPort`: The port at which the AKO API Server is available.
    * `deleteConfig`: Set to true if user wants to delete AKO created objects from Avi. Default value is `false`.
    * `disableStaticRouteSync`: Disables static route syncing if set to `true`. Default value is `false`.
    * `clusterName`: Unique identifier for the running AKO controller instance. The AKO controller identifies objects, which it created on Avi Controller using the `clusterName` param.
    * `cniPlugin`: CNI Plugin being used in kubernetes cluster. Specify one of: `calico`, `canal`, `flannel`.
    * `namespaceSelector.labelKey`: Set the key of a namespace's label, if the requirement is to sync k8s objects from that namespace.
    * `namespaceSelector.labelValue`: Set the value of a namespace's label, if the requirement is to sync k8s objects from that namespace.
  - `networkSettings`: Data network settings
    * `subnetIP`: Subnet IP of the data network. It is the network from which the VIP allocation (for the virtual services) takes place.
    * `subnetPrefix`: Subnet prefix for the data network specified in `networkSettings.subnetIP`.
    * `vipNetworkList`: List of Network Names for VIP network, multiple networks allowed only for AWS Cloud.
  - `l7Settings`: Settings for L7 Virtual Services
    * `defaultIngController`: Set to `true` if AKO controller is the default Ingress controller on the cluster.
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
    * `tenantsPerCluster`: If set to `true`, AKO controller will map each kubernetes cluster uniquely to a tenant in Avi.
    * `tenantName`: Name of the tenant where the AKO controller will create objects in AVI. Required only if `controllerSettings.tenantsPerCluster` is set to `true`.
  - `nodePortSelector`: Only applicable if `l7Settings.serviceType` is set to `NodePort`.
    * `key`
    * `value`
  - `resources`: Specify the resources for the AKO Controller's statefulset.
  - `rbac`: Enable a pod security policy for the AKO Controller.
    * `pspEnable`: Set to `true` to create a pod security policy for the AKO controller's statefulset.
  - `mountPath`: Mount path for the logs.
  - `logFile`: Log file name where the AKO controller will add it's logs.

  ## Editing the AKOConfig custom resource
  If we need any changes in the way the AKO controller was deployed, or if we want to tweak a knob in the above list, we can do that in the runtime. However, note that, only `spec.akoSettings.logLevel` and `spec.akoSettings.deleteConfig` can be changed without triggering a restart of the AKO controller. If any other knobs are changed, the ako-operator WILL trigger a restart of the AKO controller.