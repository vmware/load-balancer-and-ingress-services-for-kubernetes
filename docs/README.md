# AKO: Avi Kubernetes Operator

### Run AKO

AKO runs as a POD inside the Kubernetes cluster.

#### Pre-requisites

To Run AKO you need the following pre-requisites:

* <i>**Step 1**</i>: Configure an Avi Controller with a [vCenter cloud](https://avinetworks.com/docs/18.2/installing-avi-vantage-for-vmware-vcenter/). The Avi Controller should be versioned 18.2.10 / 20.1.2 or later.
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
* <i>**Step 4:**</i> Kubernetes 1.16+.
* <i>**Step 5:**</i> `helm` cli pointing to your kubernetes cluster.

> **NOTE**: We only support `helm 3`

### Install using *helm*

For instructions on installing AKO using helm please use this [link](install/helm.md)

### AKO CRDs

Read more about AKO [CRDs](crds/overview.md)


### AKO in Openshift Cluster

AKO can be used in openshift cluster to configure Routes and Services of type Loadbalancer. For details about how to use AKO in an openshift cluster and features specific to openshift refer [here](openshift/openshift.md).

### AKO in NSX-T deployments

Starting release 1.5.1, AKO supports the NSX-T write access cloud for both NCP and non-NCP CNIs. In case of NCP CNI, the pods are assumed to be routable from the SE's backend data network segments. Due to this, AKO disables the static route configuration when the CNI is specified as `ncp` in the values.yaml. However, if non-ncp CNIs are used, AKO assumes that static routes can be configured on the the SEs to reach the pod networks. In order for this scenario to be valid, the SEs backend data network must be configured on the same logical segment on which the Kubernetes/OpenShift cluster is run. 

In addition to this, AKO supports both overlay as well as VLAN backed NSX-T cloud configurations. AKO automatically figures out if a cloud is configured with overlay segments or is used with VLAN networks. The VLAN backed NSX-T setup behaves the same as vCenter write access cloud, thus requiring no inputs from the user. However the overlay based NSX-T setups require the user to configure a logical segment as the backend data network and correspondingly configure the T1 router's info during bootup of AKO via a helm values parameter.

### Using NodePort mode

Service of type `NodePort` can be used to send traffic to the pods exposed through Service of type `NodePort`.

This feature supports Ingress/Route attached to Service of type `NodePort`. Service of type LoadBalancer is also supported, since kubernetes populates `NodePort` by default. AKO will function either in `NodePort` mode or in `ClusterIP` mode.

A new parameter serviceType has been introduced as config option in AKO's values.yaml. To use this feature, set the value of the parameter to **NodePort**.

| **Parameter** | **Description** | **Default** |
| --------- | ----------- | ------- |
| `configs.serviceType` | Type of Service to be used as backend for Routes/Ingresses | ClusterIP |
| `nodeSelectorLabels.key` | Key used as a label based selection for the nodes in NodePort mode. | empty |
| `nodeSelectorLabels.value` | Value used as a label based selection for the nodes in NodePort mode. | empty |

Kubernetes populates NodePort by default for service of type LoadBalancer. If config.serviceType is set to NodePort, AKO would use NodePort as backend for service of type Loadbalancer instead of using Endpoints, which is the default behaviour with config.serviceType set as ClusterIP.


### AKO in Public Clouds

Please refer to this [page](public_clouds.md) for details on support for ClusterIP mode for GCP and Azure IaaS cloud in Avi Controller.

### Tenancy in AKO

Please refer to this [page](ako_tenancy.md) for support in AKO to map each kubernetes / OpenShift cluster uniquely to a tenant in Avi.

### Networking/v1 Ingress Support

Please refer to this [page](ingress/ingress.md) for details on how AKO supports and implements networking/v1 Ingress and IngressClass.

### AKO objects

Please refer to this [page](objects.md) for details on how AKO interprets the Kubernetes objects and translates them to Avi objects.

### Cloud connector to AKO migration

Please refer to this [page](cc_to_ako.md) for details on how to migrate workloads from cloud connector based Avi controller to AKO based Avi controller.

### AKO Compatibility Guide
AKO version 11.12.22 support for Kubernetes, Openshift, Avi Controller is as below:

| **Orchestrator/ Controller** | **Versions Supported** |
| --------- | ----------- |
| `Kubernetes` | 1.25 - 1.30 |
| `Openshift` | 4.12 - 4.15 |
| `Avi Controller` | 22.1.3 - 30.2.1 |


### FAQ

For some frequently asked question refer [here](faq.md) 


