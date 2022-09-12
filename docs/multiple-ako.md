# Multiple AKO instances in a cluster

This feature allows to run multiple instances of AKO per cluster.

## Overview

Prior to AKO 1.7.1, single instance of AKO was responsible to process kubernetes/openshift objects updates and synced corresponding objects in the Avi Controller.

With AKO 1.7.1, multiple AKO instances can be deployed in a cluster to create namespace based isolation.This will allow AKO to operate over a group of kubernetes namespaces, in order to handle objects from these namespaces only. To run multiple AKO, following two features will be used.

* <b>Namespace Sync feature</b>

    This feature is supported by AKO from 1.4.1.

    Namespace sync feature allows K8/Openshift objects from specific namespaces to be synced with AKO based Avi-Controller. For that, namespace has to be labelled with same key:value pair as that labelKey and labelValue mentioned in `values.yaml` file.

    Details about this feature is present at: [Namespace Sync in AKO](objects.md#namespace-sync-in-ako)

* <b>AKO installation in user provided namespace</b>

    This feature is supported from AKO 1.4.3 onwards.

    While installing AKO, using helm, flag `-n` or `--namespace` can be used to specify namespace in which AKO has to be installed. If this flag is not specified, AKO will be installed in `avi-system` namespace.

# Configuration

In multiple AKO deployment, one AKO instance works as `primary` AKO Instance. This AKO instance is responsible for vrf, static route configuration apart from syncing up K8/Openshift objects from set of namespaces.

Flag `primaryInstance` present in `values.yaml` denotes whether AKO instance is primary or not. This flag takes boolean `true`/`false` value. `true` indicates AKO instance is primary.

**Note**:
1. In multiple AKO deployment, only one AKO instance should be `primary`.
2. Each AKO should be deployed in a different namespace.

<b>Primary AKO installation</b>

```
helm install  ako/ako  --generate-name --version 1.7.1 -f /path/to/values.yaml  --set ControllerSettings.controllerHost=<controller IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --set AKOSettings.namespaceSelector.labelKey="app" --set AKOSettings.namespaceSelector.labelValue="migrate" --set AKOSettings.primaryInstance=true --namespace=avi-system

```

In above example, primary AKO instance is running in `avi-system` namespace with namespace sync filter `app: migrate`. So this AKO will sync up K8s/Openshift objects from namespaces who has labels `app: migrate`.

`helm install` command without `primaryInstance` parameter will deploy primary AKO instance.


<b>Non Primary AKO installation</b>

```
helm install  ako/ako  --generate-name --version 1.7.1 -f /path/to/values.yaml  --set ControllerSettings.controllerHost=<controller IP or Hostname> --set avicredentials.username=<avi-ctrl-username> --set avicredentials.password=<avi-ctrl-password> --set AKOSettings.namespaceSelector.labelKey="key" --set AKOSettings.namespaceSelector.labelValue="value2" --set AKOSettings.primaryInstance=false --namespace=blue

```

In above example, non-primary AKO instance is running in `blue` namespace with namespace sync filter `key: value2`. So this AKO will sync up K8s/Openshift objects from namespaces who has labels `key: value2`.

Few things that to be considered in multiple AKO instances deployment:
1. All AKO instances should interact with same AVI controller.
2. Each K8/openshift namespace should be handled by one AKO only.
3. All AKO should be deployed either in `ClusterIP` or `NodePort` or `NodePortLocal` mode.



# Avi Object naming convention

1. For non-primary AKO instance, naming convention for shared VS is: `Shared-VS-Name = <cluster-name>--<AKO-namespace>-Shared-L7-<Shard number>`. Here `<AKO-namespace>` is namespace in which AKO pod is deployed.

2. Non-primary AKO instance will create AVI objects with `username = ako-<cluster-name>-<AKO-namespace>`


