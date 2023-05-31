# OVN-Kubernetes Container Network Interface (CNI) plugin support on Openshift

This feature allows OVN-Kubernetes to be used as the CNI plugin with AKO on Openshift.

## Overview

Starting with AKO 1.10.1, the OVN-Kubernetes Container Network Interface (CNI) plugin is supported in Openshift. Prior to 1.10.1, only Openshift SDN was supported as the CNI plugin on Openshift.

## Configuration 

In order to support OVN-Kubernetes as the CNI plugin with AKO, the **AKOSettings.cniPlugin** value in the AKO Helm chart **values.yaml** should be set to `ovn-kubernetes`. The sample **values.yaml** can be found at https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/helm/ako/values.yaml, and the description for the **AKOSettings.cniPlugin** field can be found at https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/values.md#akosettingscniplugin.

AKO needs to read the pod CIDR subnets configured for Kubernetes or Openshift nodes to create static routes in the Avi controller for the pool backend servers (pods) to be reachable from the Service Engine. The OVN-Kubernetes CNI configures the Pod CIDR subnet on each node as part of the `k8s.ovn.org/node-subnets` annotation. AKO reads the **default** pod CIDR subnet value from this annotation for each node and configures the required static routes on the Avi controller. A sample annotation value is shown below.

```yaml
  k8s.ovn.org/node-subnets: '{"default":"10.128.0.0/23"}'
```

**NOTE**: AKO only supports a single pod CIDR subnet per node configured as default in the `k8s.ovn.org/node-subnets` annotation.

## Workarounds and Fixes

### Pools are down when running AKO with service type as ClusterIP 
For OVN-Kubernetes CNI, there are some Openshift setup installations where, by default, the routing gateway (OVS) performs Source NAT for traffic from PODs leaving the nodes. This Source NAT results in the pool servers, i.e., pods, being marked down in Avi Controller as it leads to failure in the health monitoring performed by Avi Controller. This issue is seen only when AKO runs with the service type configured as **ClusterIP** (ClusterIP mode). No such issue is seen in the **NodePort** mode, and pools come up normally. To learn more about `serviceType` configuration, please see https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/values.md#l7settingsservicetype.

In order to use ClusterIP mode, the Source NAT has to be disabled. However, disabling SNAT will break the ability of pods to route externally with the Node's IP Address, there by leading to failure in NodePort mode. NodePort mode should be leveraged if disabling SNAT is not desired. The changes below will disable the SNAT functionality for the namespaces that require Ingress/Route support.

1. Create a ConfigMap to set **disable-snat-multiple-gws** for cluster **network.operator**. Create a file named cm_gateway-mode-config.yaml with the following content.

```yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: gateway-mode-config
    namespace: openshift-network-operator
  data:
    disable-snat-multiple-gws: "true"
    mode: "shared"
  immutable: true
```

2. Create ConfigMap with `oc apply -f cm_gateway-mode-config.yaml`

3. Add `k8s.ovn.org/routing-external-gws` annotation to namespaces that require Ingress/Route support.
- Edit any namespace with `oc edit namespace <name-of-namespace>` and add the `k8s.ovn.org/routing-external-gws` annotation as shown below:

```yaml 
  apiVersion: v1
  kind: Namespace
  metadata:
    annotations:
      k8s.ovn.org/routing-external-gws: <ip-of-node-gateway>
```