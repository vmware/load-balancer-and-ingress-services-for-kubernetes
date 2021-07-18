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

#### Install using helm

  Step 1: Create the `avi-system` namespace:

    kubectl create ns avi-system

  Step 2: Add this repository to your helm CLI

    helm repo add ako https://avinetworks.github.io/avi-helm-charts/charts/stable/ako

Use the `values.yaml` from this repository to edit values related to Avi configuration. Values and their corresponding index can be found [here](#parameters).

  Step 3: Search the available charts for AKO Operator

    helm search repo

    NAME                          CHART VERSION APP VERSION DESCRIPTION
    ako/ako-operator              1.3.1         1.3.1       A helm chart for AKO Operator

 Step 4: Install AKO Operator

    helm install  ako/ako-operator  --generate-name --version 1.3.1 -f values.yaml  --set ControllerSettings.controllerHost=<controller IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --namespace=avi-system

  Step 5: Check the installation

    helm list -n avi-system

    NAME                       NAMESPACE
    ako-operator-2889212993     avi-system

**Note** that installing the AKO operator via `helm` will also add a `AKOConfig` object which in turn, will prompt the AKO operator to deploy the AKO controller. Please see [this](#AKOConfig-Custom-Resource) to know more about the `AKOConfig` object and how to manage the AKO controller using this object. List of CRDs added by the AKO operator installation:

1. AKOConfig
2. HostRule
3. HTTPRule

#### Uninstall using *helm*

To uninstall the AKO operator and the AKO controller, use the following steps:

*Step 1:* Remove the aviconfig object, this should cleanup all the related artifacts for the AKO controller.

    kubectl delete AKOConfig -n avi-system aviconfig

*Step2:* Remove the AKO operator's resources

    helm delete <ako-operator-release-name> -n avi-system

 Note: the `ako-operator-release-name` is obtained by doing helm list as shown in the previous step

*Step 3:* Delete the `avi-system` namespace.

    kubectl delete ns avi-system

## Parameters

The following table lists the configurable parameters of the AKO chart and their default values. Please refer to this link for more details on [each parameter](values.md).

| **Parameter** | **Description** | **Default** |
| --- | --- | --- |
| `operatorImage.repository` | Specify docker-registry that has the ako operator image | avinetworks/ako-operator |
| `operatorImage.pullPolicy` | Specify when and how to pull the ako-operator's image | avinetworks/ako-operator |
| `ControllerSettings.controllerVersion` | Avi Controller version | 18.2.10 |
| `ControllerSettings.controllerHost` | Specify Avi controller IP or Hostname | `nil` |
| `ControllerSettings.cloudName` | Name of the cloud managed in Avi | Default-Cloud |
| `L7Settings.shardVSSize` | Shard VS size enum values: LARGE, MEDIUM, SMALL | LARGE |
| `AKOSettings.fullSyncFrequency` | Full sync frequency | 1800 |
| `L7Settings.defaultIngController` | AKO is the default ingress controller | true |
| `ControllerSettings.serviceEngineGroupName` | Name of the Service Engine Group | Default-Group |
| `NetworkSettings.nodeNetworkList` | List of Networks and corresponding CIDR mappings for the K8s nodes. | `Empty List` |
| `AKOSettings.clusterName` | Unique identifier for the running AKO instance. AKO identifies objects it created on Avi Controller using this param. | **required** |
| `NetworkSettings.subnetIP` | Subnet IP of the data network | **required** |
| `NetworkSettings.subnetPrefix` | Subnet Prefix of the data network | **required** |
| `NetworkSettings.vipNetworkList` | List of Network Names for VIP network, multiple networks allowed only for AWS Cloud | **required** |
| `L4Settings.defaultDomain` | Specify a default sub-domain for L4 LB services | First domainname found in cloud's dnsprofile |
| `L7Settings.l7ShardingScheme` | Sharding scheme enum values: hostname, namespace | hostname |
| `AKOSettings.cniPlugin` | CNI Plugin being used in kubernetes cluster. Specify one of: calico, canal, flannel | **required** for calico setups |
| `AKOSettings.logLevel` | logLevel enum values: INFO, DEBUG, WARN, ERROR. logLevel can be changed dynamically from the configmap | INFO |
| `AKOSettings.deleteConfig` | set to true if user wants to delete AKO created objects from Avi. deleteConfig can be changed dynamically from the configmap | false |
| `AKOSettings.disableStaticRouteSync` | Disables static route syncing if set to true | false |
| `avicredentials.username` | Avi controller username | empty |
| `avicredentials.password` | Avi controller password | empty |
| `image.repository` | Specify docker-registry that has the AKO image | avinetworks/ako |

> `vipNetworkList`, `subnetIP` and `subnetPrefix` are required fields which are used for allocating VirtualService IP by IPAM Provider module

> Each AKO instance mapped to a given Avi cloud should have a unique clusterName parameter. This would maintain the uniqueness of object naming across Kubernetes clusters.

### AKOConfig Custom Resource

AKO Operator manages the AKO Controller. To deploy and manage the controller, it takes in a custom resource object called `AKOConfig`. Please go through the [description](AKOConfig.md#AKOConfig-Custom-Resource) to understand the different fields of this object.

#### Deploying the AKO Controller

If the AKO operator was installed using helm, a default `AKOConfig` object called `ako-config` is already added and hence, this step is not required for helm based installation.
**Note**: If the AKO operator was installed manually, then to install the AKO controller, add an `AKOConfig` object to the `avi-system` namespace.

    kubectl create -f ako-config.yaml -n avi-system

#### Tweaking/Manage the AKO Controller

If the user needs to change any properties of the AKO Controller, they can change the `AKOConfig` object and the changes will take effect once it is saved.

    kubectl edit akoconfig -n avi-system ako-config

**Note** that if the user edits the AKO controller's configmap/statefulset out-of-band, the changes will be overwritten by the AKO operator.

#### Removing the AKO Controller

To remove the AKO Controller, simply delete the `AKOConfig` object:

```
kubectl delete akoconfig -n avi-system ako-config
```
