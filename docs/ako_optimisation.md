# AKO Optimisation Recommendations

AKO watches events(CUD) of different kubernetes/openshift cluster objects to realise Avi controller side objects. AKO provides config level knobs that can help to filter K8/Openshift objects and help improving AKO performance.

This document discusses AKO `values.yaml`(`configmap`) level settings that will help in optimizing AKO performance.

## AKOSetttings.namespaceSelector.labelKey and AKOSetttings.namespaceSelector.labelValue

These two parameters act as a namespace filter. AKO syncs Ingresses/Routes, L4 services from namespaces having this namespace selector.

| **Parameter** | **Description** | **Default** |
| --------- | ----------- | ------- |
| `AKOSettings.namespaceSelector.labelKey` | Key used as a label based selection for the namespaces. | empty |
| `AKOSettings.namespaceSelector.labelValue` | Value used as a label based selection for the namespaces. | empty |

If either of the above values is left empty then AKO would sync objects from all namespaces with AVI controller.

For example, if user has specified values as labelKey:"app" and labelValue: "migrate" in values.yaml, then user has to label namespace with "app: migrate".

```
    apiVersion: v1
    kind: Namespace
    metadata:
      creationTimestamp: "2020-12-04T13:20:42Z"
      labels:
        app: migrate
      name: red
      resourceVersion: "14055620"
      selfLink: /api/v1/namespaces/red
      uid: a424bf13-2f4a-4005-a84d-f2fb65acfda0
    spec:
      finalizers:
      - kubernetes
    status:
      phase: Active
```

AKO will sync all objects from correctly labelled namespace/s.

If the label of 'red' namespace is changed from "app: migrate" (valid) to "app: migrate1" (invalid), then following objects of 'red' namespace will be deleted from AVI controller
- pools associated with, insecure ingresses/routes 
- SNI VSes associated with secure ingresses/routes 
- VSes associated with L4 objects
- EVH Vses associated with secure, insecure ingresses/routes.

AKO will sync back objects of a namespace with AVI controller if namespace label is changed from an invalid lable to a valid label.

***Note***: AKO reboot will be required if value of this knob is changed in AKO configmap.

## AKOSettings.blockedNamespaceList

The `blockedNamespaceList` lists the Kubernetes/Openshift namespaces blocked by AKO. AKO will not process any K8s/Openshift object update from these namespaces. Default value is `empty list`.

| **Parameter** | **Description** | **Default** |
| --------- | ----------- | ------- |
| `AKOSettings.blockedNamespaceList` | List of K8s/Openshift namespaces blocked by AKO | empty list |

For example, if user wants to block syncing objects from `kube-system`, `kube-public` namespaces, user can specify those namespaces as follows:

  AKOSettings:
    .
    .
    blockedNamespaceList:
      - kube-system
      - kube-public

***Note***: AKO reboot will be required if value of this knob is changed in AKO configmap.

## nodeSelectorLabels.key and nodeSelectorLabels.value

It might not be desirable to have all the nodes of a kubernetes/openshift cluster to participate in becoming server pool members, hence key/value is used as a label based selection on the nodes in kubernetes/openshift to filter out nodes. If key/value are not specified then all nodes are selected.
This setting is applicable in `NodePort` deployment only.

| **Parameter** | **Description** | **Default** |
| --------- | ----------- | ------- |
| `nodeSelectorLabels.key` | Key used as a label based selection for the nodes. | empty |
| `nodeSelectorLabels.value` | Value used as a label based selection for the nodes. | empty |

For example, if user has specified `nodeSelectorLabels.key` as a `nodeselected` and `nodeSelectorLabels.value` as a `yes`, then nodes which do have this label will be selected during pool server population.

```
    apiVersion: v1
    kind: Node
    metadata:
      annotations:
        node.alpha.kubernetes.io/ttl: "0"
        volumes.kubernetes.io/controller-managed-attach-detach: "true"
      labels:
        kubernetes.io/hostname: node2
        kubernetes.io/os: linux
        nodeselected: yes
      name: node2
    spec:
      .
      .
      .
```

AKO will select `node2` while populating pool servers.
***Note***: AKO reboot will be required if value of this knob is changed in AKO configmap.
