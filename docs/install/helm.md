## Install using *helm*

Step 1: Create the `avi-system` namespace:

```
kubectl create ns avi-system
```

> **Note**: Starting AKO-1.4.3, AKO can run in namespaces other than `avi-system`. The namespace in which AKO is deployed, is governed by the `--namespace` flag value provided during `helm install` (Step 4). There are no updates in any setup steps whatsoever. `avi-system` has been kept as is in the entire documentation, and should be replaced by the namespace provided during AKO installation.

Step 2: Add this repository to your helm CLI


```
helm repo add ako https://projects.registry.vmware.com/chartrepo/ako

```
> **Note**: The helm charts are present in VMWare's public harbor repository

Step 3: Search the available charts for AKO

```
helm search repo

NAME                 	CHART VERSION	    APP VERSION	        DESCRIPTION
ako/ako              	1.8.1        	    1.8.1      	        A helm chart for Avi Kubernetes Operator

```

Use the `values.yaml` from this chart to edit values related to Avi configuration. To get the values.yaml for a release, run the following command

```
helm show values ako/ako --version 1.8.1 > values.yaml

```

Values and their corresponding index can be found [here](#parameters)

Step 4: Install AKO

```
helm install  ako/ako  --generate-name --version 1.8.1 -f /path/to/values.yaml --set ControllerSettings.controllerHost=<controller IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --namespace=avi-system
```

Step 5: Check the installation

```
helm list -n avi-system

NAME          	NAMESPACE 	
ako-1593523840	avi-system
```

## Uninstall using *helm*

Simply run:

*Step1:*

```
helm delete <ako-release-name> -n avi-system
```

> **Note**: the ako-release-name is obtained by doing helm list as shown in the previous step,

*Step 2:*

```
kubectl delete ns avi-system
```

## Upgrade AKO using *helm*

Follow these steps if you are upgrading from an older AKO release.

*Step1*

Helm does not upgrade the CRDs during a release upgrade. Before you upgrade a release, run the following command to download and upgrade the CRDs:

```
helm template ako/ako --version 1.8.1 --include-crds --output-dir <output_dir>
```

This will save the helm files to an output directory which will contain the CRDs corresponding to the AKO version.
To install the CRDs:

```
kubectl apply -f <output_dir>/ako/crds/
```

*Step2*

```
helm list -n avi-system

NAME          	NAMESPACE 	REVISION	UPDATED                             	    STATUS  	CHART    	APP VERSION
ako-1593523840	avi-system	1       	2020-09-16 13:44:31.609195757 +0000 UTC	    deployed	ako-1.8.1	1.8.1
```

*Step3*

Update the helm repo URL

```
helm repo add --force-update ako https://projects.registry.vmware.com/chartrepo/ako

"ako" has been added to your repositories

```
> **Note**: From AKO 1.3.3, we are migrating our charts repo to VMWare's harbor repository and hence a force update of the repo URL is required for a successful upgrade process from 1.3.3

*Step4*

Get the values.yaml for the latest AKO version

```
helm show values ako/ako --version 1.8.1 > values.yaml

```

Upgrade the helm chart

```
helm upgrade ako-1593523840 ako/ako -f /path/to/values.yaml --version 1.8.1 --set ControllerSettings.controllerHost=<IP or Hostname> --set avicredentials.password=<username> --set avicredentials.username=<username> --namespace=avi-system

```

## Parameters

The following table lists the configurable parameters of the AKO chart and their default values. Please refer to this link for more details on [each parameter](../values.md).

| **Parameter** | **Description** | **Default** |
| --------- | ----------- | ------- |
| `ControllerSettings.controllerVersion` | Avi Controller version | Current Controller version |
| `ControllerSettings.controllerHost` | Specify Avi controller IP or Hostname | `nil` |
| `ControllerSettings.cloudName` | Name of the cloud managed in Avi | Default-Cloud |
| `ControllerSettings.tenantName` | Name of the tenant where all the AKO objects will be created in AVI. | admin |
| `L7Settings.shardVSSize` | Shard VS size enum values: LARGE, MEDIUM, SMALL, DEDICATED | LARGE |
| `AKOSettings.fullSyncFrequency` | Full sync frequency | 1800 |
| `L7Settings.defaultIngController` | AKO is the default ingress controller | true |
| `ControllerSettings.serviceEngineGroupName` | Name of the Service Engine Group | Default-Group |
| `NetworkSettings.nodeNetworkList` | List of Networks and corresponding CIDR mappings for the K8s nodes. | `Empty List` |
| `AKOSettings.clusterName` | Unique identifier for the running AKO instance. AKO identifies objects it created on Avi Controller using this param. | **required** |
| `NetworkSettings.subnetIP` | Subnet IP of the data network | **DEPRECATED** |
| `NetworkSettings.subnetPrefix` | Subnet Prefix of the data network | **DEPRECATED** |
| `NetworkSettings.vipNetworkList` | List of Network Names and Subnet information for VIP network, multiple networks allowed only for AWS Cloud | **required** |
| `NetworkSettings.enableRHI` | Publish route information to BGP peers | false |
| `NetworkSettings.bgpPeerLabels` | Select BGP peers using bgpPeerLabels, for selective VsVip advertisement. | `Empty List` |
| `NetworkSettings.nsxtT1LR` | Specify the T1 router for data backend network, applicable only for NSX-T based deployments| `Empty string` |
| `L4Settings.defaultDomain` | Specify a default sub-domain for L4 LB services | First domainname found in cloud's dnsprofile |
| `L4Settings.autoFQDN`  | Specify the layer 4 FQDN format | default |  
| `L7Settings.noPGForSNI`  | Skip using Pool Groups for SNI children | false |  
| `L7Settings.l7ShardingScheme` | Sharding scheme enum values: hostname, namespace | hostname |
| `AKOSettings.cniPlugin` | CNI Plugin being used in kubernetes cluster. Specify one of: calico, canal, flannel, ncp | **required** for calico, ncp setups |
| `AKOSettings.enableEvents` | enableEvents can be changed dynamically from the configmap | true |
| `AKOSettings.logLevel` | logLevel enum values: INFO, DEBUG, WARN, ERROR. logLevel can be changed dynamically from the configmap | INFO |
| `AKOSettings.deleteConfig` | set to true if user wants to delete AKO created objects from Avi. deleteConfig can be changed dynamically from the configmap | false |
| `AKOSettings.disableStaticRouteSync` | Disables static route syncing if set to true | false |
| `AKOSettings.apiServerPort` | Internal port for AKO's API server for the liveness probe of the AKO pod | 8080 |
| `AKOSettings.layer7Only` | Operate AKO as a pure layer 7 ingress controller | false |
| `avicredentials.username` | Avi controller username | empty |
| `avicredentials.password` | Avi controller password | empty |
| `avicredentials.authtoken` | Avi controller authentication token | empty |
| `image.repository` | Specify docker-registry that has the AKO image | avinetworks/ako |

> From AKO 1.5.1, fields `subnetIP` and `subnetPrefix` have been deprecated. See [Upgrade Notes](../upgrade/upgrade.md) for more details.

> `vipNetworkList` is a required field which is used for allocating  VirtualService IP by IPAM Provider module

> Each AKO instance mapped to a given Avi cloud should have a unique clusterName parameter. This would maintain the uniqueness of object naming across Kubernetes clusters.
