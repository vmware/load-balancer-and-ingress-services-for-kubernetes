### AviInfraSetting

AviInfraSetting provides a way to segregate Layer-4/Layer-7 VirtualServices to have properties based on different underlying infrastructure components,
like ServiceEngineGroup, intended VIP Network etc.

A sample AviInfraSetting CRD looks like this:

```
apiVersion: ako.vmware.com/v1alpha1
kind: AviInfraSetting
metadata:
  name: my-infra-setting
spec:
  seGroup:
    name: compact-se-group
  network:
    vipNetworks:
      - networkName: vip-network-10-10-10-0-24
        cidr: 10.10.10.0/24
    enableRhi: true
    bgpPeerLabels:
      - peer1
      - peer2
  l7Settings:
    shardSize: MEDIUM
```

### AviInfraSetting with Services/Ingress/Routes

AviInfraSetting is a Cluster scoped CRD and can be attached to the intended Services, Kubernetes Ingresses and Openshift Routes by ways described below.

#### Services
AviInfraSetting resources can be attached to Services using Gateway APIs, or simply by using annotations.

##### Using Gateway API

Gateway APIs provide interfaces to structure Kubernetes service networking. More information around Gateway API can be found [here](https://gateway-api.sigs.k8s.io/). AKO provides support for Gateway APIs via the `servicesAPI` flag in the `values.yaml`.
More details related to how AKO integrates with Gateway API is covered as part of the [gateway-api](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs/gateway-api/gateway-api.md) documentation.
The AviInfraSetting resource can be attached to a GatewayClass object, via the `.spec.parametersRef` as shown below

```
apiVersion: networking.x-k8s.io/v1alpha1
kind: GatewayClass
metadata:
  name: avi-gateway-class
spec:
  controller: ako.vmware.com/avi-lb
  parametersRef:
    group: ako.vmware.com
    kind: AviInfraSetting
    name: my-infrasetting
```

##### Using annotation

In case the `servicesAPI` flag is NOT set to `true`, and AKO is not watching over the Gateway APIs, Services of Type `LoadBalancer` can specify the AviInfraSetting using an annotation as shown below

```
  annotations:
    aviinfrasetting.ako.vmware.com/name: "my-infrasetting"
```

#### Ingress

AviInfraSettings can be applied to Ingress resources, using the IngressClass construct. IngressClass provides a way to configure controller specific load balancing parameters and applies these configurations to a set of Ingress objects. AKO supports listening to IngressClass resources in Kubernetes version 1.19+. The AviInfraSetting reference can be provided in the IngressClass as shown below

```
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: avi-ingress-class
spec:
  controller: ako.vmware.com/avi-lb
  parameters:
    apiGroup: ako.vmware.com
    kind: AviInfraSetting
    name: my-infrasetting
```

#### Openshift Routes

AviInrfaSetting integrates with Openshift Routes, via the annotation

```
  annotations:
    aviinfrasetting.ako.vmware.com/name: "my-infrasetting"
```


### AviInfraSetting CRD Usage

#### Configure ServiceEngineGroup 

AviInfraSetting CRD can be used to configure Service Engine Groups (SEGs) for virtualservices created corresponding to Services/Ingresses/Openshift Routes. The Service Engine Group should have been created and configureed in the Avi Controller prior to this CRD creation.

        seGroup:
          name: compact-se-group

AKO tries to configure labels on the SEG specified in the AviInfraSetting resources, which enables static route syncing on the member Service Engines. The labels are configured by AKO only when the SEGs do not have any other labels configured already. In case AKO finds the SEG configured with some different labels, the AviInfraSetting resource would be `Rejected`.
Note that the member Service Engines, once deployed, do not reflect any changes related to label additions/updates on the SEGs. Therefore, label based static route syncing will not work on already deployed Service Engines.
Please make sure that the SEGs have no member Service Engines deployed, before specifying the SEG in the AviInfraSetting resource, in order to correctly configure static routet syncing.

#### Configure VIP Networks

**Note**: AKO 1.5.1 updates the schema to provide VIP network information in AviInfraSetting CRD. See [Upgrade Notes](./upgrade/upgrade.md) for more details.

AviInfraSetting CRD can be used to configure VIP networks for virtualservices created corresponding to Services/Ingresses/Openshift Routes. The Networks must be present in the Avi Controller prior to this CRD creation.

        network:
          vipNetworks:
            - networkName: vip-network-10-10-10-0-24
              cidr: 10.10.10.0/24

Note that multiple networks names can be added to the CRD (only in case of AWS cloud). The Avi virtualservices will acquire a VIP from each of these specified networks. Failure in allocating even a single vip (for example, in case of IP exhaustion) **will** result in complete failure of entire request. *This is same as vip allocation failures in single vip.*

#### Enable/Disable Route Health Injection

AviInfraSetting CRD can be used to enable/disable Route Health Injection (RHI) on the virtualservices created by AKO. 

        network:
          enableRhi: true

This overrides the global `enableRHI` flag for the virtualservices corresponding to the AviInfraSetting.

#### Enable/Disable Public IP

AviInfraSetting CRD can be used to enable/disable Public IP on the virtualservices created by AKO. 

        network:
          enablePublicIP: true

Note that enabling Public IP is only supported for **public clouds**.

#### Configure BGP Peer Labels for BGP VSes 

AviInfraSetting CRD can be used to configure BGP Peer labels for BGP virtualservices. AKO configures the VSes with the appropriate peer labels, only when `enableRHI` is enabled, either during AKO installation via helm chart's `values.yaml` or via the AviInfraSetting CRD itself. If not set to `true`, the AviInfraSetting resource will be marked `Rejected`, 

        bgpPeerLabels:
          - peer1
          - peer2

#### Use dedicated vip for Ingress

AviInfraSetting CRD can be used to allocate a dedicated vip per Ingress FQDN.

        l7Settings:
          shardSize: DEDICATED

For the subset of ingresses, that refer to an ingress class which in turn refers to an AviInfraSetting CRD setting that has shardSize as DEDICATED, will get vip per Ingress FQDN.

