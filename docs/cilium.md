# Cilium Container Network Interface (CNI) plugin support on Kubernetes

This feature allows Cilium to be used as the CNI plugin with AKO on Kubernetes.

## Overview

Starting with AKO 1.10.1, the Cilium Container Network Interface (CNI) plugin is supported on Kubernetes. Cilium can be configured to use either Cluster Scope mode or Kubernetes Host Scope mode for IPAM, and AKO is capable of supporting both.  

To see the IPAM mode, check the `ipam` field in the `cilium-config` configmap in the `kube-system` namespace.  
For cluster scope mode, the ipam value is **cluster-pool**.

```yaml
  ipam: cluster-pool
```

While, for Kubernetes host scope mode, the ipam value is **kubernetes**.

```yaml
  ipam: kubernetes
```

## Configuration 

AKO needs to read the per-node PodCIDRs to be able to sync the static route configurations. With Cilium CNI, there are two modes to configure the per-node PodCIDRs.

### Cluster Scope IPAM mode

By **default**, Cilium uses the `Cluster Scope` mode for IPAM. To use Cilium in the cluster scope ipam mode with AKO, the `AKOSettings.cniPlugin` value in the AKO Helm chart **values.yaml** should be set to `cilium`. The sample **values.yaml** can be found at https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/helm/ako/values.yaml, and the description for the **AKOSettings.cniPlugin** field can be found at https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/values.md#akosettingscniplugin.  
In the cluster scope mode, the podCIDRs range are made available via the `CiliumNode (cilium.io/v2.CiliumNode)` CRD and AKO reads this CRD to determine the Pod CIDR to Node IP mappings. The CiliumNode CRD object is created with the same name as the node name (one per node) and specifies the podCIDRs range in the `spec.ipam.podCIDRs` field.

### Kubernetes Host Scope IPAM mode

In Kubernetes host scope mode, podCIDRs are allocated out of the PodCIDR range associated to each node by Kubernetes. This PodCIDR range is available in the Node `spec.podCIDRs` field. By default, when the `cniPlugin` flag is empty, AKO determines the Pod CIDR to Node IP mappings from Node `spec.podCIDRs` field and configures the static routes accordingly. Hence, the `cniPlugin` flag should be left empty for Kubernetes Host Scope IPAM mode.

## Conditions and Caveats

AKO supports the SCTP protocol for L4 services. Read https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/sctp.md for more information.  
Starting with version 1.13 Cilium CNI also provides basic SCTP support. However, SCTP support is not enabled by default and needs to be enabled in the Cilium configuration before the SCTP protocol can be used in port definitions. To enable SCTP support, the `enable-sctp` field in `cilium-config` configmap should be set to `"true"`.

```yaml
  enable-sctp: "true"
```

**NOTE**: The Kernel version on the Kubernetes nodes should be 5.2 or newer for Cilium to support SCTP. Otherwise, the Cilium daemon may crash with the following error.

```
level=fatal msg="failed to start: daemon creation failed: SCTP support needs kernel 5.2 or newer" subsys=daemon
```
