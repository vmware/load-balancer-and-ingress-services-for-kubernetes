# Enhanced Virtual Hosting support in AKO

This feature supports EVH VS creation from the AKO. `AKOSettings.enableEVH` needs to be set to `true` to enable this feature. Currently, this feature is supported only for Kubernetes Ingress object.

## Overview

AKO currently creates an SNI child VS (Virtual Service) to a parent shared VS for the secure hostname. The SNI VS is used to bind the hostname to a sslkeycert object. The sslkeycert object is used to terminate the secure traffic on Avi's service engine. On the SNI VS, AKO creates httppolicyset rules to route the terminated (insecure) traffic to the appropriate pool object using the host/path specified in the rules section of this ingress object.

With EVH (Enhanced Virtual Hosting) support in AVI, virtual hosting on virtual service can be enabled irrespective of SNI. Also, the SNI can only handle HTTPS (HTTP over SSL) traffic whereas EVH children can handle both HTTP and HTTPS traffic. Unlike SNI which switches only TLS (Transport Layer Security) connections based on one-to-one mapping of children to FQDN (Fully Qualified Domain Name), EVH maps one FQDN to many children based on the resource path requested.

With EVH enabled host rule CRD's can be applied to insecure ingress as well. 

More details of EVH can be found here <https://avinetworks.com/docs/20.1/enhanced-virtual-hosting/>.

### Naming of AVI Objects with EVH enabled

##### Shared VS names

The shared VS names are derived based on a combination of fields to keep it unique per Kubernetes cluster. This is the only object in Avi that does not derive its name from any of the Kubernetes objects.

```
ShardVSName = clusterName + "--Shared-L7-EVH-" + <shardNum>
```

`clusterName` is the value specified in values.yaml during install. "Shared-L7" is a constant identifier for Shared VSes
`shardNum` is the number of the shared VS generated based on either hostname or namespace based shards.

##### EVH child VS names

```
vsName = clusterName + "--" + hostName
```

##### EVH pool names

The formula to derive the Child EVH virtual service's pools is as follows:

```
poolName = clusterName + "--" + namespace + "-" + host + "_" + path + "-" + ingName +  ServiceName
```

Here the `host` and `path` variables denote the secure hosts' hostname and path specified in the ingress object.

##### EVH poolgroup names

The formula to derive the Child EVH virtual service's pool group is as follows:

```
poolgroupname = clusterName + "--" + namespace + "-" + host + "_" + path + "-" + ingName
```
