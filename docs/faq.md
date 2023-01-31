## FAQ

This document answers some of the frequently asked questions w.r.t AKO.

#### How do I clean up all my configs?

The key deleteConfig in the data section of AKO configmap can be used to clean up the setup. Edit AKO configmap and set deleteConfig: "true" to delete ako created objects in Avi. After the flag is set in configmap, annotation `AviObjectDeletionStatus` is added in the AKO statefulset with the value as `Started`. 

```yaml
  annotations:
    AviObjectDeletionStatus: Started
```

After all relevant objects gets deleted from Avi, the value of the annotation is changed to `Done`.

```yaml
  annotations:
    AviObjectDeletionStatus: Done
```

To re-create the objects in Avi, the configmap has to be edited to set deleteConfig: "false".

#### How is the Shared VS lifecycle controlled?

AKO follows hostname based sharding to sync multiple ingresses with same hostname to a single virtual service. When an ingress object is created with multiple hostnames, AKO generates an md5 hash using the hostname and the Shard VS number. This uniquely maps an FQDN to a given Shared VS and avoids DNS conflicts. During initial clean bootup, if the Shared VS does not exist in Avi - AKO creates the same and then patches the ingress FQDN to it either in the form of a pool (for insecure routes) or in the form of an SNI child virtual service (in case of secure routes).

The Shared VSes aren't deleted if all the FQDNs mapped to it are removed from Kubernetes. However, if the user wants AKO to delete unused shared VSes - a pod restart is required that would evaluate the VS and delete it appropriately.

Even though the unused shared VSes are deleted, the shared VS VIPs are retained to regain the same FQDN to VS VIP mapping. To delete the retained shared VS VIPs, the user has to manually delete them from the controller UI or shell.

#### How are VSes sharded?

If you create ingress with an insecure host/path combination then AKO creates a corresponding Avi Pool object and patches the pool
on one of the existing shard virtual services. The shard VS has a datascript associated with it that reads the host/path of the incoming
request and appropriately selects a pool by matching it with the priority label specified for each pool member (corresponding to a host/path
combination).

For secure ingresses, an SNI virtual service is created which although is a dedicated virtual service, does not have any IP addresses
associated with it. The SNI virtual service is a child to a parent virtual service and is created based on the secret object specified
in the ingress file against the host/path that is meant to be accessed securely.

#### How do I decide the Shard VS size?

In the current AKO model, the Shard VS size is an enum. It allows 3 pre-fixed sets of values viz. `LARGE`, `MEDIUM` and `SMALL`. They
respectively correspond to 8, 4 and 1 virtual service. The decision of selecting one of these sizes for Shard VS is driven by the
size of the Kubernetes cluster's ingress requirements. Typically, it's advised to always go with the highest possible Shard VS number
that is - `LARGE` to account for future expansion.

The shard size can be set to `DEDICATED` to disable shard mode to create dedicated Virtual Services per hostname.

#### Can I change the Shard VS number?

To Shard to virtual services, AKO uses a sharding mechanism that is driven by the `hostname` of each rule within an ingress object. This ensures that a unique hostname is always sharded consistently to the same virtual service.

Since the sharding logic is determined by the number of Shard virtual services, changing the Shard VS number has the potential hazard
of messing up an existing cluster's already synced objects. Hence it's recommended that the Shard VS numbers are not changed once fixed.

#### How do I alter the Shard VS number?

Altering the shard VS number is considered as disruptive. This is because dynamic re-adjustment of shard numbers may re-balance
the ingress to VS mapping. Hence if you want to alter the shard VS number, first delete the older configmap and trigger a complete
cleanup of the VSes in the controller. Followed by an edit of the configmap and restart of AKO.

#### What happens if the number of DNS records exceed a Shard VS?

Currently, the number of A records allowed per virtual service is 1000. If a shard size of `SMALL` is selected and the number of A records via the Ingress objects exceed 1000, then a greater `shardSize` has to be configured via the `shardSize` knob. Alternatively one can create a separate IngressClass for a set of Ingress objects and specify a `shardSize` in the `AviInfraSettings` CRD which would allow AKO to place the A records scoped to the VS that is mapped to the IngressClass.

#### What is the use of static routes?

Static routes are created with cluster name as the label. While deploying AKO the admin or the operator decides a Service Engine Group for a given
Kubernetes cluster. The same labels are tagged on the routes of this AKO cluster. These routes are pushed to the Service Engine's created on the Service Engine Group.
The static routes map each POD CIDR with the Kubernetes node's IP address. However, for static routes to work, the Service Engines must
be L2 adjacent to your Kubernetes nodes.

#### What happens if I have the same SNI host across multiple namespaces?

The ingress API does not prohibit the user from creating the same SNI hostname across multiple namespaces. AKO will create 1 SNI virtual service and gather all paths associated with it across namespaces to create corresponding switching
rules. However, the user needs to denote each ingress with the TLS secret for a given hostname to qualify the host for the SNI virtual service.

Consider the below example:

    Ingress 1 (default namespace) --> SNI hostname --> foo.com path: /foo, Secret: foo

    Ingress 1 (foo namespace) --> SNI hostname --> foo.com path: /bar, Secret: foo

In the above case, only 1 SNI virtual service will be created with a sslkeyandcertificate as `foo`.

However if the following happens:

    Ingress 1 (default namespace) --> SNI hostname --> foo.com path: /foo, Secret: foo

    Ingress 1 (foo namespace) --> SNI hostname --> foo.com path: /bar, Secret: bar

Then the behaviour of the SNI virtual service would be indeterministic since the secrets for the same SNI are different. This is not supported.

#### What out of band operations can I do on the objects created by AKO?

AKO runs a refresh cycle that currently just refreshes the cloud object parameters. However, if some out of band operations are performed on objects created by AKO via directly interacting with the Avi APIs, AKO may not always be able to remediate
an error caused due to this.

AKO has the best effort, retry layer implementation that would try to detect a problem (For example an SNI VS deleted from the Avi UI), but it is not guaranteed to work for all such manual operations.

Upon reboot of AKO - a full reconciliation loop is run and most of the out-of-band changes are overwritten with AKO's view of the intended model. This does not happen in every full sync cycle.

#### What is the expected behaviour for the same host/path combination across different secure/insecure ingresses?

The ingress API allows users to add duplicate hostpaths bound to separate backend services. Something like this:

    Ingress1 (default namespace) --> foo.com path: /foo, Service: svc1

    Ingress2 (default namespace) --> foo.com path: /foo, Service: svc2

Also, ingress allows you to have a mix of secure and insecure hostpath bound to the same backend services like so:

    Ingress1 (default namespace) --> SNI hostname --> foo.com path: /foo, Secret: secret1

    Ingress2 (default namespace) --> foo.com path: /foo, Service: svc2

AKO does not explicitly handle these conditions and would continue syncing these objects on the Avi controller, but this may lead to traffic issues.
AKO does the best effort of detecting some of these conditions by printing them in logs. A sample log statement looks like this:

`key: Ingress/default/ingress2, msg: Duplicate entries found for hostpath default/ingress2: foo.com/foo in ingresses: ["default/ingress1"]`

#### What happens to static routes if the Kubernetes nodes are rebooted/shutdown?

AKO programs a static route for every node IP and the POD CIDR associated with it. Even though node state changes to `NotReady` in Kubernetes this configuration is stored in the node object and does not change when the node rebooted/shutdown.

Hence AKO will not remove the static routes until the Kubernetes node is completely removed from the cluster.

#### Can I point my ingress objects to a service of type Loadbalancer?

The short answer is No.
The ingress objects should point to the service of type clusterIP. Loadbalancer services either point to an ingress controller POD if one is using an in cluster ingress controller or they can directly point to application PODs that need layer 4 load-balancing.

If you have such a configuration where the ingress objects are pointing to services of the type load balancer, AKO's behaviour would be indeterministic.

#### What happens when AKO fails to connect to the AVI controller while booting up?

AKO would stop processing kubernetes objects and no update would be made to the AVI Controller. After the connection to AVI Controller is restored, AKO pod has to be rebooted. This can be done by deleting the exiting POD and ako deployment would bring up a new POD, which would start processing kubernetes objects after verifying connectivity to AVI Controller.  

#### What happens if we create ingress objects in Openshift environment ?

AKO will process both Ingress and Route objects in Openshift environments.

#### What are the virtual services for passthrough routes or ingresses?

A set of shared Virtual Services are created for passthrough hosts to listen on port 443 to handle secure traffic using L4 datascript. These virtual services have names of the format 'cluster-name'-`Shared-Passthrough`-'shard-number'. Number of shards can be configured using the flag `passthroughShardSize` while installation using helm.

#### What happens if insecureEdgeTerminationPolicy is set to `redirect` for a passthrough route?

 For passthrough routes, the supported values for insecureEdgeTerminationPolicy are None and Redirect. To handle insecure traffic for passthrough routes a set of shared Virtual Services are created with names of the format 'cluster-name'-`Shared-Passthrough`-'shard-number'-`insecure`. These Virtual Services listen on port 80. If for any passthrough route, the insecureEdgeTerminationPolicy is found to be 'Redirect', then an HTTP Policy is configured in the insecure passthrough shared VS to send appropriate response to an incoming insecure traffic.

#### How to debug 'Invalid input detected' errors?

AKO goes for a reboot and retries some of the invalid input errors. Below are some of the cases to look out for in the logs.

- If an invalid cloud name is given in `values.yaml` or if ipam_provider_ref is not set in the vCenter and No Access clouds.
- If the same Service Engine Group is used for multiple clusters for vCenter and No Access clouds in Cluster IP mode. This happens as AKO expects unique SE group per cluster if routes are configured by AKO for POD reachability. Look for the `Labels does not match with cluster name` message in the logs which points to two clusters using the same Service Engine Group.

#### How to fix when some of the pool servers in NodePort mode of AKO are down?

The default behaviour for AKO s to populate all the Node IP as pool server. If master node is not schedulable then, it will be marked down. `nodePortSelector` can be used to specify the `labels` for the node. In that case, all the node with that label will be picked for the pool server. If the master node is not schedulable then, the fix is to remove the `nodePortSelector` label for the master node.

#### Can we create a secure route with edge/reencrypt termination without key or certificate ?

For secure routes having termination type edge/reencrypt, key and certificate must be specified in the spec of the route. AKO would not handle routes of these types without key and certificate.

#### What happens if a route is created with multiple backends having same service name ?

AKO would reject those routes, as each backend should be unique ith it's own weight. Multiple backends having same service would make weight calculation indeterministic.

#### Is NodePortLocal feature available for all CNIs ?

No. The feature NodePortLocal can be used only with Antrea CNI and the feature must be enabled in Antrea feature gates.

#### Can we use kubernetes Service of type NodePort as backend of an Ingress in NodePortLocal mode ?

No. Users can only use service of type ClusterIP as backend of Ingresses in this mode.

#### Can we use serviceType NodePort or ClusterIP in AKO, if the CNI type is Antrea and NodePortLocal feature is enabled in Antrea ?

Yes. AKO would create AVI objects based on the relevant serviceType set in AKO.

#### What are the steps for ServiceType change?

The `serviceType` in AKO can be changed from `ClusterIP` to `NodePortLocal` or `NodePort`. The `serviceType` change is considered disruptive.
Hence before the `serviceType` change, all the existing AKO configuration must be deleted. This can be achieved as follows:

  - Set the `deleteConfig` flag to `true`.
  - Wait for AKO to delete all the relevant Avi configuration and update the deletion status in AKO's statefulset status.
  - Change the `serviceType`
  - Set the `deleteConfig` flag to `false`
  - Reboot AKO

For example, during the change of `serviceType` from `ClusterIP` to `NodePortLocal`, the `deleteConfig` flag will:

  - Delete the static routes.
  - Delete the SE group labels.

#### Can the serviceType in AKO be changed dynamically ?

No. After changing the serviceType, AKO has to be rebooted and all objects which are not required, would be deleted as part of the reboot process.

#### If serviceType is changed from NodePortLocal, would AKO remove NPL annotation from the Services automatically ?

No. After changing the serviceType, the users have to remove NPL annotation from the Services themselves.

#### How can I see the marker labels associated with an object on the Avi Controller?

Markers, associated with an Avi object, are visible on the Avi Vantage UI and the Avi shell.

1. Avi Vantage UI: 

Markers are visible on UI for avi controller version >= 20.1.6.
In order to view marker labels for an object, do the following:

- Edit an object.
- Navigate to the `Advanced` tab.
- The Role-Based Access Control (RBAC) shows the markers associated with the object.

![Alt-text](./images/Markers-on-UI.png)


2. Avi Shell

Markers can be viewed for an object with the command.

  `[avi-shell] show <object type> <object name>`

Sample command with output:

  `[avi-shell] show virtualservice kubernetes--Shared-L7-EVH-0`

    +------------------------------------+----------------------------------------------------------------------------------+
    | Field                              | Value                                                                            |
    +------------------------------------+----------------------------------------------------------------------------------+
    | uuid                               | virtualservice-bc4e964a-af79-4b8f-91b2-de0e7ee9388d                              |
    | name                               | kubernetes--Shared-L7-EVH-0                                                      |
    | enabled                            | True                                                                             |
    | vsvip_ref                          | kubernetes--Shared-L7-EVH-0                                                      |
    | use_vip_as_snat                    | False                                                                            |
    |        .                           | .                                                                                |
    |        .                           | .                                                                                |
    |        .                           | .                                                                                |
    |        .                           | .                                                                                |
    |        .                           | .                                                                                |
    | markers[1]                         |                                                                                  |
    |   key                              | clustername                                                                      |
    |   values[1]                        | kubernetes                                                                       |
    | allow_invalid_client_cert          | False                                                                            |
    | vh_type                            | VS_TYPE_VH_ENHANCED                                                              |
    +------------------------------------+----------------------------------------------------------------------------------+


#### What is differnce between `AKOSetttings.namespaceSelector` and `AKOSettings.blockedNamespaceList`? What are the scenarios to use one of it?

`AKOSetttings.namespaceSelector` allows AKO to process Ingress/Routes, L4 services from the namespaces that have given labels. For namespaces, that doesn't have given labels, AKO blocks processing Ingress/Routes and L4 services thus prevents Avi object creation. But AKO processes other objects like secrets, endpoints etc from wrongly/unlabelled namespaces.

`AKOSettings.blockedNamespaceList` is list of namespaces, specified upfront, from which AKO does not process any objects like ingress/route, L4 services, secrets, endpoints etc. So `AKOSettings.blockedNamespaceList` is broader way of blocking processing unwanted objects compared to `AKOSetttings.namespaceSelector`. With current implementation, `AKOSettings.blockedNamespaceList` does not support regex so user has to specify each and every namespace.

Few scenarios to consider `AKOSettings.blockedNamespaceList` are:
*  If list of namespaces, to be blocked, is known to user upfront. For example, system-namespaces where normally user does not deploy any app. Any new addition of namespace, to this list, will require AKO reboot.

* There are too many secret, pods, endpoint etc. objects in unwanted namespaces that are consuming AKO processing time.

Few scenarios to consider `AKOSetttings.namespaceSelector` are:
* In cluster, namespace churn is more frequent. So using namespace selector is better way of dealing with processing only wanted namespaces.

One of the way to use both settings `AKOSetttings.namespaceSelector` and `AKOSettings.blockedNamespaceList` effectively is: use `AKOSettings.blockedNamespaceList` for system generated namespaces (where there is no app deployment from user) and use `AKOSetttings.namespaceSelector` for user-defined namespaces.

#### What is the minimum stable Kubernetes version which supports SCTP protocol?

The minimum stable Kubernetes version which supports SCTP protocol is 1.20.

#### What is the minimum AVI Controller version which supports SCTP protocol?

The minimum AVI Controller version which supports SCTP protocol is 22.1.3.

#### What is the minimum AKO version which supports SCTP protocol?

The minimum AKO version which supports SCTP protocol is 1.9.1.

#### How to manually override the Active AKO when HA is enabled?

Trigger a deletion of the Active AKO pod, and the passive AKO pod automatically comes up as the new active AKO.

#### Can we scale beyond two instances of AKO in HA?

Currently, AKO beyond two has not been tested. Hence we don't claim the support.

#### What happens during a Split brain scenario in HA?

AKO detects this during the periodic refresh of the lease lock object and makes itself a passive AKO.
