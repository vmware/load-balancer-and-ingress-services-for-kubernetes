# AKO Operator

## Overview

This is an operator that is used to deploy, manage, and remove an instance of the AKO controller on Openshift clusters. It takes the AKO installation/deployment configuration from a CRD called `AKOConfig` and creates an instance of the AKO controller and installs all the relevant objects specified below. 

1. AKO statefulset
2. Clusterrole and Clusterrolbinding
3. Configmap required for the AKO controller
and other artifacts.

## Run AKO Operator

### Pre-requisites

This is one of the ways to install the AKO controller. So, most of the pre-requisites that apply for installation of standalone AKO are also applicable for the AKO operator as well.

* <i>**Step 1**</i>: Configure an Avi Controller with a [vCenter cloud](https://avinetworks.com/docs/18.2/installing-avi-vantage-for-vmware-vcenter/) or any other preferred cloud. The Avi Controller should be versioned 18.2.10 / 20.1.2 or later.
* <i>**Step 2**</i>:
  * Make sure a PG network is part of the NS IPAM configured in the vCenter
* <i>**Step 3**</i>: If your POD CIDRs are not routable:
Data path flow is as described below:
![Alt text](data_path_flow.png?raw=true)
The markers in the drawing are described below:
    1. The client requests a specified hostname/path.
    2. The DNS VS returns an IP address corresponding to the hostname.
    3. The request is forwarded to the resolved IP address that corresponds to a Virtual IP hosted on an Avi Service Engine.
    The destination IP in the packet is set as the POD IP address on which the application runs.
    4. Service Engines use the static route information to reach the POD IP via the next-hop address of the host on which the pod is running.
    5. The pod responds and the request is sent back to the client.
  * Create a Service Engine Group dedicated to a Kubernetes cluster.
* <i>**Step 3.1**</i>: If your POD CIDRs are routable then you can skip step 2. Ensure that you skip static route syncing in this case using the `disableStaticRouteSync` flag in the `values.yaml` of your helm chart.
* <i>**Step 4:**</i> Openshift 4.10+.

### Install on Openshift cluster from OperatorHub using Openshift Container Platform Web Console

<i>**Step 1**</i>: Login to the Openshift Container Platform web console of your Openshift cluster.

<i>**Step 2**</i>: Navigate in the web console to the **Operators** → **OperatorHub** page.

<i>**Step 3**</i>: Find `AKO Operator` provided by VMware.

<i>**Step 4**</i>: Click `install` and select the 1.11.1 version. The operator will be installed in `avi-system` namespace. The namespace will be created if it doesn't exist.

<i>**Step 5**</i>: Verify installation by checking the pods in `avi-system` namespace.

**Note** To deploy an instance of the AKO controller, an `AKOConfig` object will have to be created. This, in turn, will prompt the AKO operator to deploy the AKO controller. Please see [this](../ako_operator.md#akoconfig-custom-resource) to know more about the `AKOConfig` object and how to manage the AKO controller using this object.

#### Parameters

The following table lists the configurable fields in the `AKOConfig` object and their default values.

| **Parameter** | **Description** | **Default** |
| --- | --- | --- |
| `replicaCount` | Specify the number of replicas for AKO StatefulSet | 1 |
| `imageRepository` | Specify docker-registry that has the ako image | projects.packages.broadcom.com/ako/ako:1.13.1 |
| `imagePullPolicy` | Specify when and how to pull the ako image | IfNotPresent |
| `imagePullSecrets` | ImagePullSecrets will add pull secrets to the statefulset for AKO. Required if using secure private container image registry for images. | `Empty List` |
| `AKOSettings.clusterName` | Unique identifier for the running AKO instance. AKO identifies objects it created on Avi Controller using this param. | **required** |
| `AKOSettings.fullSyncFrequency` | Full sync frequency | 1800 |
| `AKOSettings.cniPlugin` | CNI Plugin being used in Openshift cluster. Specify one of: openshift, ovn-kubernetes | **required** for openshift, ovn-kubernetes |
| `AKOSettings.enableEvents` | enableEvents can be changed dynamically from the configmap | true |
| `AKOSettings.logLevel` | logLevel enum values: INFO, DEBUG, WARN, ERROR. logLevel can be changed dynamically from the configmap | INFO |
| `AKOSettings.deleteConfig` | set to true if user wants to delete AKO created objects from Avi. deleteConfig can be changed dynamically from the configmap | false |
| `AKOSettings.disableStaticRouteSync` | Disables static route syncing if set to true | false |
| `AKOSettings.apiServerPort` | Internal port for AKO's API server for the liveness probe of the AKO pod | 8080 |
| `AKOSettings.layer7Only` | Operate AKO as a pure layer 7 ingress controller | false |
| `AKOSettings.blockedNamespaceList` | List of K8s/Openshift namespaces blocked by AKO | `Empty List` |
| `AKOSettings.istioEnabled` | set to true if user wants to deploy AKO in istio environment (tech preview)| false |
| `AKOSettings.ipFamily` | set to V6 if user wants to deploy AKO with V6 backend (vCenter cloud with calico CNI only) (tech preview)| V4 |
| `AKOSettings.enableEVH` | Enables the Enhanced Virtual Hosting Model in Avi Controller for the Virtual Services  | false |
| `AKOSettings.namespaceSelector` | namespaceSelector contains label key and value used for namespacemigration. same label has to be present on namespace/s which needs migration/sync to AKO  | false |
| `AKOSettings.servicesAPI` | servicesAPI enables AKO in services API mode. Currently implemented only for L4 | false |
| `AKOSettings.vipPerNamespace` | Enabling this flag would tell AKO to create Parent VS per Namespace in EVH mode  | false |
| `AKOSettings.useDefaultSecretsOnly` | If this flag is set to true, AKO will only handle default secrets from the namespace where AKO is installed. This flag is applicable only to Openshift clusters. | false |
| `ControllerSettings.controllerVersion` | Avi Controller version | 18.2.10 |
| `ControllerSettings.controllerIP` | Specify Avi controller IP or Hostname | `nil` |
| `ControllerSettings.cloudName` | Name of the cloud managed in Avi | Default-Cloud |
| `ControllerSettings.tenantName` | Name of the tenant where all the AKO objects will be created in AVI. | admin |
| `ControllerSettings.serviceEngineGroupName` | Name of the Service Engine Group | Default-Group |
| `L7Settings.shardVSSize` | Shard VS size enum values: LARGE, MEDIUM, SMALL | LARGE |
| `L7Settings.defaultIngController` | AKO is the default ingress controller | true |
| `L7Settings.serviceType` | enum NodePort|ClusterIP|NodePortLocal | ClusterIP |
| `L7Settings.passthroughShardSize` | Control the passthrough virtualservice numbers using this ENUM. ENUMs: LARGE, MEDIUM, SMALL | SMALL |
| `L7Settings.noPGForSNI`  | Skip using Pool Groups for SNI children | false |
| `L4Settings.defaultDomain` | Specify a default sub-domain for L4 LB services | First domainname found in cloud's dnsprofile |
| `L4Settings.autoFQDN`  | Specify the layer 4 FQDN format | default |
| `NetworkSettings.subnetIP` | Subnet IP of the data network | **DEPRECATED** |
| `NetworkSettings.subnetPrefix` | Subnet Prefix of the data network | **DEPRECATED** |
| `NetworkSettings.nodeNetworkList` | List of Network Names/UUIDs and corresponding CIDR mappings for the K8s nodes. | `Empty List` |
| `NetworkSettings.vipNetworkList` | List of Network Names/UUIDs and Subnet information for VIP network, multiple networks allowed only for AWS Cloud | **required** |
| `NetworkSettings.enableRHI` | Publish route information to BGP peers | false |
| `NetworkSettings.bgpPeerLabels` | Select BGP peers using bgpPeerLabels, for selective VsVip advertisement. | `Empty List` |
| `NetworkSettings.nsxtT1LR` | Specify the T1 router for data backend network, applicable only for NSX-T based deployments| `Empty string` |
| `FeatureGates.gatewayAPI` | FeatureGates is to enable or disable experimental features. GatewayAPI feature gate enables/disables processing of Kubernetes Gateway API CRDs. | false |
| `GatewayAPI.Image.repository` | Specify docker-registry that has the ako-gateway-api image | projects.packages.broadcom.com/ako/ako-gateway-api:1.13.1 |
| `GatewayAPI.Image.pullPolicy` | Specify when and how to pull the ako-gateway-api image | IfNotPresent |
| `logFile` | LogFile is the name of the file where ako container will dump its logs | avi.log |
| `akoGatewayLogFile` | AKOGatewayLogFile is the name of the file where ako-gateway-api container will dump its logs | avi-gw.log |
| `avicredentials.username` | Avi controller username | empty |
| `avicredentials.password` | Avi controller password | empty |
| `avicredentials.authtoken` | Avi controller authentication token | empty |


> Each AKO instance mapped to a given Avi cloud should have a unique clusterName parameter. This would maintain the uniqueness of object naming across Kubernetes clusters.

AKO Operator will also create the following list of CRDs to be used by AKO Controller when the `AKOConfig` object is created:

1. AKOConfig
2. HostRule
3. HTTPRule
4. L4Rule
5. SSORule

### Uninstall on Openshift cluster from OperatorHub using Openshift Container Platform Web Console

<i>**Step 1**</i>: Remove the aviconfig object, this should cleanup all the related artifacts for the AKO controller. See [Removing the AKO Controller](../ako_operator.md#removing-the-ako-controller) for more details.

<i>**Step 2**</i>: Login to the Openshift Container Platform web console of your Openshift cluster.

<i>**Step 3**</i>: Navigate in the web console to the **Operators** → **Installed Operators** page.

<i>**Step 4**</i>: Find `AKO Operator` provided by VMware.

<i>**Step 5**</i>: Click on the three vertical dots menu on the right and select `Uninstall Operator` option.

<i>**Step 6**</i>: Delete the `avi-system` namespace.

    kubectl delete ns avi-system

Or, if using the Openshift client, use
    
    oc delete ns avi-system