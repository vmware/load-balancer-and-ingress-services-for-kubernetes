## Install using *helm*

Step 1: Create the `avi-system` namespace:

```
kubectl create ns avi-system
```

> **Note**: Starting from AKO-1.4.3, AKO can run in namespaces other than `avi-system`. The namespace in which AKO is deployed is governed by the `--namespace` flag value provided during `helm install` (Step 4). There are no updates in any setup steps whatsoever. `avi-system` has been kept as is in the entire documentation, and should be replaced by the namespace provided during AKO installation.

> **Note**: Helm version 3.8 or above is required to proceed with Helm installation.


Step 2: Search the available charts for AKO

```
helm show chart oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.14.1

Pulled: projects.packages.broadcom.com/ako/helm-charts/ako:1.14.1
Digest: sha256:xyxyxxyxyx
apiVersion: v2
appVersion: 1.14.1
description: A helm chart for Avi Kubernetes Operator
name: ako
type: application
version: 1.14.1
```

Use the `values.yaml` from this chart to edit values related to Avi configuration. To get the values.yaml for a release, run the following command:

```
helm show values oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.14.1 > values.yaml

```

Values and their corresponding index can be found [here](#parameters).

Step 3: Install AKO

Starting from AKO-1.7.1, multiple AKO instances can be installed in a cluster.
> **Note**: <br>
    1. Only one AKO instance, out of multiple AKO instances, should be `Primary`. <br>
    2. Each AKO instance should be installed in a different namespace.

<b>Primary AKO installation</b>

Starting from AKO-1.14.1, the AKO Helm chart has a dependency chart `ako-crd-operator`. This can be enabled/disabled by setting `ako-crd-operator.enabled` in the AKO values.yaml.
```
helm install --generate-name oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.14.1 -f /path/to/values.yaml  --set ControllerSettings.controllerHost=<controller IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --set AKOSettings.primaryInstance=true --namespace=avi-system --dependency-update
```

Note: Set the dependent `ako-crd-operator` chart repository: `oci://projects.packages.broadcom.com/ako/helm-charts/ako-crd-operator` in AKO Chart.yaml before installation.  

<b>Secondary AKO installation</b>
```
helm install --generate-name oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.14.1 -f /path/to/values.yaml  --set ControllerSettings.controllerHost=<controller IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --set AKOSettings.primaryInstance=false --namespace=avi-secondary-system --set ako-crd-operator.enabled=false

```
Note: Since only one instance of ako-crd-operator can run in the cluster, for secondary AKO Installation, set ako-crd-operator.enabled to false.

Step 4: Check the installation

```
helm list -n avi-system

NAME          	NAMESPACE 	REVISION	UPDATED     STATUS  	CHART    	APP VERSION
ako-1691752136	avi-system	1       	2023-09-28	deployed	ako-1.14.1	1.14.1
```

## Uninstall using *helm*

Simply run:

*Step1:*

```
helm delete <ako-release-name> -n avi-system
```

> **Note**: The ako-release-name is obtained by running `helm list` as shown in the previous step.

*Step 2:*

```
kubectl delete ns avi-system
```

## Upgrade AKO using *helm*

Follow these steps if you are upgrading from an older AKO release.

*Step1*

Helm does not upgrade the CRDs during a release upgrade. Before you upgrade a release, run the following command to download and upgrade the CRDs:

```
helm template oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.14.1 --include-crds --output-dir <output_dir>
```

This will save the helm files to an output directory which will contain the CRDs corresponding to the AKO version.
To install the CRDs:

```
kubectl apply -f <output_dir>/ako/crds/
```

Note: This step is not necessary for CRDs included in the ako-crd-operator, as this will be the initial installation of the ako-crd-operator, and all associated CRDs will be newly installed during the upgrade process.

*Step2*

```
helm list -n avi-system

NAME          	NAMESPACE 	REVISION	UPDATED                             	    STATUS  	CHART    	APP VERSION
ako-1593523840	avi-system	1       	2024-08-04 13:44:31.609195757 +0000 UTC	    deployed	ako-1.12.2.1.22.2
```

*Step3*

Get the values.yaml for AKO version 1.14.1 and edit the values as per the requirement.

```
helm show values oci://projects.packages.broadcom.com/ako/helm-charts/ako --version 1.14.1 > values.yaml

```
*Step4*

Upgrade the Helm chart:

```
helm upgrade ako-1593523840  oci://projects.packages.broadcom.com/ako/helm-charts/ako -f /path/to/values.yaml --version 1.14.1 --set ControllerSettings.controllerHost=<IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --namespace=avi-system --dependency-update

```

Upgrade the secondary AKO:
```
helm upgrade  <secondary-ako-chart-name>  oci://projects.packages.broadcom.com/ako/helm-charts/ako -f /path/to/values.yaml --version 1.14.1 --set ControllerSettings.controllerHost=<IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --namespace=avi-secondary-system --set ako-crd-operator.enabled=false --dependency-update
```

***Note***
1. In a multiple AKO deployment scenario, all AKO instances should be on the same version.
2. In a given cluster, there can only be one ako-crd-operator instance running.

## Parameters

The following table lists the configurable parameters of the AKO chart and their default values. Please refer to this link for more details on [each parameter](../values.md).

| **Parameter** | **Description** | **Default** |
| --------- | ----------- | ------- |
| `ControllerSettings.controllerVersion` | Avi Controller version | Current Controller version |
| `ControllerSettings.controllerHost` | Specify Avi controller IP or Hostname | `nil` |
| `ControllerSettings.cloudName` | Name of the cloud managed in Avi | Default-Cloud |
| `ControllerSettings.tenantName` | Name of the tenant where all the AKO objects will be created in AVI. | admin |
| `ControllerSettings.primaryInstance` | Specify AKO instance is primary or not | true |
| `L7Settings.shardVSSize` | Shard VS size enum values: LARGE, MEDIUM, SMALL, DEDICATED | LARGE |
| `AKOSettings.fullSyncFrequency` | Full sync frequency | 1800 |
| `L7Settings.defaultIngController` | AKO is the default ingress controller | true |
| `ControllerSettings.serviceEngineGroupName` | Name of the Service Engine Group | Default-Group |
| `NetworkSettings.nodeNetworkList` | List of Networks (specified using either name or uuid) and corresponding CIDR mappings for the K8s nodes. | `Empty List` |
| `AKOSettings.clusterName` | Unique identifier for the running AKO instance. AKO identifies objects it created on Avi Controller using this param. | **required** |
| `NetworkSettings.subnetIP` | Subnet IP of the data network | **DEPRECATED** |
| `NetworkSettings.subnetPrefix` | Subnet Prefix of the data network | **DEPRECATED** |
| `NetworkSettings.vipNetworkList` | List of Network Names/ Network UUIDs and Subnet information for VIP network, multiple networks allowed only for AWS Cloud | **required** |
| `NetworkSettings.enableRHI` | Publish route information to BGP peers | false |
| `NetworkSettings.bgpPeerLabels` | Select BGP peers using bgpPeerLabels, for selective VsVip advertisement. | `Empty List` |
| `NetworkSettings.nsxtT1LR` | Unique ID (note: not display name) of the T1 Logical Router for Service Engine connectivity. Only applies to NSX-T cloud.| `Empty string` |
| `NetworkSettings.defaultDomain` | This flag has two use cases. It can be used to specify a default subdomain for L4 virtual services. It can also be used to generate the hostname for an OpenShift route if the OpenShift route uses a subdomain instead of a host field. | First domain name found in cloud's dnsprofile for L4 vs and `empty string` for an Openshift route |
| `FeatureGates.gatewayAPI` | FeatureGates is to enable or disable experimental features. GatewayAPI feature gate enables/disables processing of Kubernetes Gateway API CRDs. | false |
| `GatewayAPI.Image.repository` | Specify docker-registry that has the ako-gateway-api image | projects.registry.vmware.com/ako/ako-gateway-api |
| `GatewayAPI.Image.pullPolicy` | Specify when and how to pull the ako-gateway-api image | IfNotPresent |
| `L4Settings.defaultDomain` | Specify a default sub-domain for L4 LB services. This flag will be deprecated in a future release; use **NetworkSettings.defaultDomain** instead. If both NetworkSettings.defaultDomain and L4Settings.defaultDomain are set, then NetworkSettings.defaultDomain will be used.| First domain name found in cloud's dnsprofile |
| `L4Settings.autoFQDN`  | Specify the layer 4 FQDN format | default |  
| `L7Settings.noPGForSNI`  | Skip using Pool Groups for SNI children | false |  
| `L7Settings.fqdnReusePolicy` | Restrict FQDN to single namespace if set to `Strict`. enum: InterNamespaceAllowed, Strict | InterNamespaceAllowed |
| `AKOSettings.cniPlugin` | CNI Plugin being used in kubernetes cluster. Specify one of: calico, canal, flannel, openshift, antrea, ncp, ovn-kubernetes, cilium | **required** for calico, openshift, ovn-kubernetes, ncp setups. For Cilium CNI, set the string as **cilium** only when using Cluster Scope mode for IPAM and leave it empty if using Kubernetes Host Scope mode for IPAM. |
| `AKOSettings.enableEvents` | enableEvents can be changed dynamically from the configmap | true |
| `AKOSettings.logLevel` | logLevel enum values: INFO, DEBUG, WARN, ERROR. logLevel can be changed dynamically from the configmap | INFO |
| `AKOSettings.deleteConfig` | set to true if user wants to delete AKO created objects from Avi. deleteConfig can be changed dynamically from the configmap | false |
| `AKOSettings.disableStaticRouteSync` | Disables static route syncing if set to true | false |
| `AKOSettings.apiServerPort` | Internal port for AKO's API server for the liveness probe of the AKO pod | 8080 |
| `AKOSettings.layer7Only` | Operate AKO as a pure layer 7 ingress controller | false |
| `AKOSettings.blockedNamespaceList` | List of K8s/Openshift namespaces blocked by AKO | `Empty List` |
| `AKOSettings.istioEnabled` | set to true if user wants to deploy AKO in istio environment| false |
| `AKOSettings.ipFamily` | set to V6 if user wants to deploy AKO with V6 backend (vCenter cloud with calico CNI only) (tech preview)| V4 |
| `AKOSettings.useDefaultSecretsOnly` | Restricts the secret handling to default secrets present in the namespace where AKO is installed in Openshift clusters if set to true | false |
| `AKOSetttings.namespaceSelector` |  Key-value pair represent a label that is used by AKO to filter out namespace/s | empty |
| `avicredentials.username` | Avi controller username | empty |
| `avicredentials.password` | Avi controller password | empty |
| `avicredentials.authtoken` | Avi controller authentication token | empty |
| `avicredentials.certificateAuthorityData` | RootCA of the Avi controller, that AKO uses to verify the server certificate provided by the Avi Controller during the TLS handshake | empty |
| `image.repository` | Specify docker-registry that has the AKO image | avinetworks/ako |
| `image.pullSecrets` | Specify the pull secrets for the secure private container image registry that has the AKO image | `Empty List` |
| `nodePortSelector` | Key-Value pair used as a label based selection used by AKO to filter out K8s/OpenShift nodes while populating the pool members. Applicable in AKO NodePort mode | empty |
| `securityContext` |  Security configuration applied on container running in AKO POD | empty |
| `podSecurityContext` | Security configuration applied on AKO POD | empty |
| `ako-crd-operator.controllerManager.container.image.repository` | Specify docker-registry that has the ako-crd-operator image | packages.broadcom.com/ako/ako-crd-operator |
| `ako-crd-operator.controllerManager.container.image.tag` | Specify image tag to be used for ako-crd-operator image | latest |
| `ako-crd-operator.controllerManager.container.image.pullPolicy` | Specify when and how to pull the ako-crd-operator image | IfNotPresent |
| `ako-crd-operator.controllerManager.container.image.pullSecrets` | Specify the pull secrets for the secure private container image registry that has the ako-crd-operator image | `Empty List` |
| `ako-crd-operator.controllerManager.container.resources.limits.cpu` | Specify CPU limit for ako-crd-operator pod | 256m |
| `ako-crd-operator.controllerManager.container.resources.limits.memory` | Specify Memory limit for ako-crd-operator pod | 128Mi |
| `ako-crd-operator.controllerManager.container.resources.requests.cpu` | Specify CPU request for ako-crd-operator | 128m |
| `ako-crd-operator.controllerManager.container.resources.requests.memory` | Specify Memory request for ako-crd-operator | 64Mi |

> From AKO 1.5.1, fields `subnetIP` and `subnetPrefix` have been deprecated. See [Upgrade Notes](../upgrade/upgrade.md) for more details.

> `vipNetworkList` is a required field which is used for allocating  VirtualService IP by IPAM Provider module

> Each AKO instance mapped to a given Avi cloud should have a unique clusterName parameter. This would maintain the uniqueness of object naming across Kubernetes clusters.
