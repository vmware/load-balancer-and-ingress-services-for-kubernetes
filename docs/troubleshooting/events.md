## Kubernetes Events

AKO 1.6.1 introduces support for broadcasting kubernetes Events in order to enhance the observability and monitoring aspects of AKO as an ingress controller.

Kubernetes Events are stored objects that are generated by controllers in response to various user actions. These events are stored in the Kubernetes store for 1 hour by default, but can be changed while configuring the kube-apiserver.

Avi Kubernetes Operator broadcasts events in order to -
- to enhance debuggability.  
- use events for better error reporting, for support and engineering team to report issue after analyzing event timeline.
- provide granular debugging on Ingress/Routes/SvcLB/Gateways and show their relationship with Avi virtual services, that AKO creates.

AKO Event broadcasting can be controlled by setting the `enableEvents` flag in the ConfigMap appropriately. By default the Event broadcasting is enabled, but can be switched off by updating the ConfigMap, which comes into affect without rebooting AKO.

## Event Types

The events fired by AKO are segregated into two Event types. `Normal` and `Warning`.

### Normal

`Normal` events are expected responses to certain user actions, that confirm a successful workflow. These type of events do not require any further user input changes.

A few examples of a `Normal` events are:

```
20m         Normal    ValidatedUserInput   pod/ako-0   User input validation completed.
```

```
20m         Normal    StatusSync           pod/ako-0   Status syncing completed
```

```
33s         Normal   Synced               ingress/ingress1                                      Added virtualservice clusterName--Shared-L7-1 for bar.avi.com
```

### Warning

`Warning` events are deviations from the Normal workflow, and generally carry a Error message with it that tells more about what went wrong, and what needs to be fixed as part of it.

A few examples of a `Warning` events are:

```
23s         Warning   AKOShutdown          pod/ako-0   Invalid user input [No user input detected for vipNetworkList]
```

**Note**: Regardless of the `enableEvents` setting in the ConfigMap, Warning Events are always broadcasted via AKO.

## Event Categories

AKO broadcasted events can be categorized into 3 classes, and are described as follows:

### Pod events

AKO broadcasts Pod events referencing the AKO Pod. 
Pod events primarily consist of checkpoints that the AKO Pod goes through, starting from bootup to the time it is ready to sync objects to the Avi controller. It also covers `Warning` type Events in case of any user input errors, and other issues that prevent a successful AKO bootup. 



### Ingress/Route/ServiceLB/Gateway events

The second category of Events are referenced to objects corresponding to which AKO creates virtual services in the Avi controller, for instance Ingress, Openshift Routes, Services of type load balancer and Gateway objects. These objects directly correspond to one or more virtual services created in Avi via AKO, and also receive a VIP for the virtual service, which is updated in the status of the respective objects.
The events related to these objects are primarily `Normal` events, that tell the user when and which virtual service was created corresponding to the object. For instance:

Ingress
```
Events:
  Type    Reason  Age              From                     Message
  ----    ------  ----             ----                     -------
  Normal  Synced  6s               avi-kubernetes-operator  Added virtualservice ako-clusterName--Shared-L7-6 for foo.avi.com
  Normal  Synced  4s               avi-kubernetes-operator  Added virtualservice ako-clusterName--bar.avi.com for bar.avi.com
```

Ingress in EVH mode
```
Events:
  Type    Reason   Age                From                     Message
  ----    ------   ----               ----                     -------
  Normal  Synced   5s                 avi-kubernetes-operator  Added virtualservice ako-clusterName--ddd26961643229facf2b2d94d05e33519ed3fbfd for foo.avi.com
  Normal  Synced   5s                 avi-kubernetes-operator  Added virtualservice ako-clusterName--3aedd52095d8864d41be2264c181042b6fc58c28 for bar.avi.com
```

Service of Type LoadBalancer
```
Events:
  Type    Reason   Age   From                     Message
  ----    ------   ----  ----                     -------
  Normal  Type     68s   service-controller       ClusterIP -> LoadBalancer
  Normal  Synced   64s   avi-kubernetes-operator  Added virtualservice ako-clusterName--default-avisvc-https for avisvc-https
  Normal  Type     2s    service-controller       LoadBalancer -> ClusterIP
  Normal  Removed  1s    avi-kubernetes-operator  Removed virtualservice for avisvc-https
```

Apart from the virtual services being created/removed corresponding to these objects, other `Warning` events can tell certain misconfigurations in the object, for instance, when an multiple Ingresses contain duplicate host paths.

```
Events:
  Type     Reason             Age              From                     Message
  ----     ------             ----             ----                     -------
  Warning  DuplicateHostPath  8s               avi-kubernetes-operator  Duplicate entries found for hostpath default/ingress1: foo.avi.com/path4 in ingresses: ["default/ingress1","default/ingress2"]
```

### AKO CRD events

These are events that are referenced to AKO CRDs, specifically the HostRule/HTTPRule CRDs. Once a CRD is created, the configurations mentioned in the CR are applied to a VS or a Pool. The CRD events tell, to which specific VS/Pool, the HostRule/HTTPRule is applied. Example of a HostRule event is as follows:

```
Events:
  Type    Reason    Age   From                     Message
  ----    ------    ----  ----                     -------
  Normal  Attached  11s   avi-kubernetes-operator  Configuration applied to VirtualService ako-clusterName--3aedd52095d8864d41be2264c181042b6fc58c28
```


### Helpful Commands

This section covers details around where to find the Events, the commands that can be used, and how to filter AKO specific events.
All Events created by AKO have a `source` specified as `avi-kubernetes-operator`, and reference a single object, based on the Event category discussed above.

One can check for kubernetes events by simply using the following command: 

```
kubectl get events
```

Although this would show all the events generated by various other controllers in the cluster. Events can be filtered within a namespace, or by referenced object name etc. These filtering mechanisms are native to Kubernetes and are not AKO specific. For instance, in order to see Events generated for AKO Pod we can use the following command:

```
kubectl get events -n avi-system --field-selector involvedObject.name=ako-0
```

Similarly for Ingresses we can use

```
kubectl get events --field-selector involvedObject.name=ingress1
```

In addition to the `kubectl get events` command and the filters that come with it, we can also check the Events for a particular object using the `kubectl describe` command. The `describe` command, very neatly, aut-filters all the events corresponding to that object, and prints the output.

```
kubectl describe pod -n avi-system ako-0
```

```
kubectl describe ingress ingress1
```
