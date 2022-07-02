# AKO Operator

## Overview

This is an operator which is used to deploy, manage and remove an instance of the AKO controller. This operator when deployed creates an instance of the AKO controller and installs all the relevant objects like:

1. AKO statefulset
2. Clusterrole and Clusterrolbinding
3. Configmap required for the AKO controller
and other artifacts.

## Run AKO Operator

### Pre-requisites

This is one of the ways to install the AKO controller. So, all the [pre-requisites](README.md#pre-requisites) that apply for installation of standalone AKO are also applicable for the AKO operator as well.

### Install using helm

<i>**Step 1**</i>: Create the avi-system namespace:

```
kubectl create ns avi-system
```

<i>**Step 2**</i>: Add this repository to your helm CLI:
```
helm repo add ako https://projects.registry.vmware.com/chartrepo/ako
```

<i>**Step 3**</i>: Search the available charts for the operator:
```
helm search repo

NAME            	CHART VERSION	APP VERSION	DESCRIPTION
ako/ako         	1.6.1        	1.6.1      	A helm chart for Avi Kubernetes Operator
ako/ako-operator	1.6.3        	1.6.3      	A Helm chart for Kubernetes AKO Operator
ako/amko        	1.6.1        	1.6.1      	A helm chart for Avi Kubernetes Operator
```

<i>**Step 4**</i>: Use the values.yaml from this chart to edit values related to Avi configuration. To get the values.yaml for a release, run the following command:
```
helm show values ako/ako-operator --version 1.6.3 > values.yaml
```

<i>**Step 5**</i>: Edit the <i>values.yaml</i> file and update the details according to your environment.

<i>**Step 6**</i>: Install the Operator:
```
helm install  ako/ako-operator  --generate-name --version 1.6.3 -f /path/to/values.yaml --namespace=avi-system
```

<i>**Step 7**</i>: Verify the installation:
```
helm list -n avi-system
``` 

**Note** that installing the AKO operator via `helm` will also add a `AKOConfig` object which in turn, will prompt the AKO operator to deploy the AKO controller. Please see [this](../akoconfig.md) to know more about the `AKOConfig` object and how to manage the AKO controller using this object. List of CRDs added by the AKO operator installation:

1. AKOConfig
2. HostRule
3. HTTPRule
4. AviInfraSetting

> For other methods of installation refer [here](../../ako-operator/README.md)

### Uninstall using *helm*

To uninstall the AKO operator and the AKO controller, use the following steps:

*Step 1:* Remove the aviconfig object, this should cleanup all the related artifacts for the AKO controller.

    kubectl delete AKOConfig -n avi-system aviconfig

*Step2:* Remove the AKO operator's resources

    helm delete <ako-operator-release-name> -n avi-system

> **Note**: the `ako-operator-release-name` is obtained by doing helm list as shown in the previous step

*Step 3:* Delete the `avi-system` namespace.

    kubectl delete ns avi-system

## Parameters

The following table lists the configurable parameters of the AKO chart and their default values. Please refer to this link for more details on [each parameter](../values.md).

| **Parameter** | **Description** | **Default** |
| --- | --- | --- |
| `operatorImage.repository` | Specify docker-registry that has the ako operator image | projects.registry.vmware.com/ako/ako-operator |
| `operatorImage.pullPolicy` | Specify when and how to pull the ako-operator's image | IfNotPresent |
| `akoImage.repository` | Specify docker-registry that has the ako image | projects.registry.vmware.com/ako/ako:1.6.1 |
| `akoImage.pullPolicy` | Specify when and how to pull the ako image | IfNotPresent |
| `AKOSettings.enableEvents` | enableEvents can be changed dynamically from the configmap | true |
| `AKOSettings.logLevel` | logLevel enum values: INFO, DEBUG, WARN, ERROR. logLevel can be changed dynamically from the configmap | INFO |
| `AKOSettings.fullSyncFrequency` | Full sync frequency | 1800 |
| `AKOSettings.apiServerPort` | Internal port for AKO's API server for the liveness probe of the AKO pod | 8080 |
| `AKOSettings.deleteConfig` | set to true if user wants to delete AKO created objects from Avi. deleteConfig can be changed dynamically from the configmap | false |
| `AKOSettings.disableStaticRouteSync` | Disables static route syncing if set to true | false |
| `AKOSettings.clusterName` | Unique identifier for the running AKO instance. AKO identifies objects it created on Avi Controller using this param. | **required** |
| `AKOSettings.cniPlugin` | CNI Plugin being used in kubernetes cluster. Specify one of: calico, canal, flannel, ncp | **required** for calico, openshift, ncp setups |
| `AKOSettings.layer7Only` | Operate AKO as a pure layer 7 ingress controller | false |
| `NetworkSettings.nodeNetworkList` | List of Networks and corresponding CIDR mappings for the K8s nodes. | `Empty List` |
| `NetworkSettings.enableRHI` | Publish route information to BGP peers | false |
| `NetworkSettings.nsxtT1LR` | Specify the T1 router for data backend network, applicable only for NSX-T based deployments| `Empty string` |
| `NetworkSettings.bgpPeerLabels` | Select BGP peers using bgpPeerLabels, for selective VsVip advertisement. | `Empty List` |
| `NetworkSettings.vipNetworkList` | List of Network Names and Subnet information for VIP network, multiple networks allowed only for AWS Cloud | **required** |
| `L7Settings.defaultIngController` | AKO is the default ingress controller | true |
| `L7Settings.serviceType` | enum NodePort|ClusterIP|NodePortLocal | ClusterIP |
| `L7Settings.shardVSSize` | Shard VS size enum values: LARGE, MEDIUM, SMALL, DEDICATED | LARGE |
| `L7Settings.passthroughShardSize` | Control the passthrough virtualservice numbers using this ENUM. ENUMs: LARGE, MEDIUM, SMALL | SMALL |
| `L7Settings.noPGForSNI`  | Skip using Pool Groups for SNI children | false |  
| `L4Settings.defaultDomain` | Specify a default sub-domain for L4 LB services | First domainname found in cloud's dnsprofile |
| `L4Settings.autoFQDN`  | Specify the layer 4 FQDN format | default | 
| `ControllerSettings.serviceEngineGroupName` | Name of the Service Engine Group | Default-Group | 
| `ControllerSettings.controllerVersion` | Avi Controller version | Current Controller version |
| `ControllerSettings.cloudName` | Name of the cloud managed in Avi | Default-Cloud |
| `ControllerSettings.controllerHost` | Specify Avi controller IP or Hostname | `nil` |
| `ControllerSettings.tenantsPerCluster` | If set to true, AKO will map each kubernetes cluster uniquely to a tenant in Avi | false |
| `ControllerSettings.tenantName` | Name of the tenant where all the AKO objects will be created in AVI. | admin |
| `avicredentials.username` | Avi controller username | empty |
| `avicredentials.password` | Avi controller password | empty |
| `avicredentials.authtoken` | Avi controller authentication token | empty |


> Each AKO instance mapped to a given Avi cloud should have a unique clusterName parameter. This would maintain the uniqueness of object naming across Kubernetes clusters.

