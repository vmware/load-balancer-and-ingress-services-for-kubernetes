# Enhanced Virtual Hosting support in AKO

This feature supports EVH VS creation from the AKO. `AKOSettings.enableEVH` needs to be set to `true` to enable this feature. This feature is supported for Kubernetes Ingress object and Openshift Route object.

## Overview

AKO currently creates an SNI child VS (Virtual Service) to a parent shared VS for the secure hostname. The SNI VS is used to bind the hostname to a sslkeycert object. The sslkeycert object is used to terminate the secure traffic on Avi's service engine. On the SNI VS, AKO creates httppolicyset rules to route the terminated (insecure) traffic to the appropriate pool object using the host/path specified in the rules section of this ingress object.

With EVH (Enhanced Virtual Hosting) support in AVI, virtual hosting on virtual service can be enabled irrespective of SNI. Also, the SNI can only handle HTTPS (HTTP over SSL) traffic whereas EVH children can handle both HTTP and HTTPS traffic. For each unique host in kubernetes/openshift cluster, unique EVH child will be created in secure, insecure deployments. L4 and Passthrough deployments will be similar to that of SNI way of deployment.

With EVH enabled host rule CRD's can be applied to insecure ingress as well. 

More details of EVH can be found here <https://avinetworks.com/docs/20.1/enhanced-virtual-hosting/>.

### Naming of AVI Objects with EVH enabled

In AVI controller, attributes of AVI objects have size restriction. Name of each AVI object has certain length restriction. AKO generates name of each AVI object by concatenating different fields of ingress/routes/services. Sometime this results in overflow of size restriction imposed by AVI Controller.

To avoid this issue, name of object will be encoded for objects created by AKO in EVH deployments. Name of shared object such as Shared VirtualService, VIP will not be encoded. Also name of object in Advance L4 deployment will not be encoded.

##### Shared VS names

The shared VS names are derived based on a combination of fields to keep it unique per Kubernetes cluster/ Openshift cluster. This is the only object in Avi that does not derive its name from any of the Kubernetes/Openshift objects.

```
ShardVSName = clusterName + "--Shared-L7-EVH-" + <shardNum>
```

`clusterName` is the value specified in values.yaml during install. "Shared-L7-EVH" is a constant identifier for Shared VSes
`shardNum` is the number of the shared VS generated based on hostname based shards.

##### EVH child VS names

```
vsName = clusterName + "--" + encoded-value
```

##### EVH pool names

```
poolName = clusterName + "--" + encoded-value
```

##### EVH poolgroup names

```
poolgroupname = clusterName + "--" + encoded-value
```
