AKO, from v1.4.1, claims support for Layer 4 Service integration with Gateway APIs v1alpha1. In order to enable the feature, and allow AKO to watch for Gateway API objects - GatewayClass and Gateway - the `servicesAPI` flag in the `values.yaml` must be set to `true`.

### Installation

AKO primarily uses GatewayClass and Gateway CRDs for it's Gateway API implementation and integration with Layer 4 Services. These GatewayClass and Gateway CRDs must be installed on the cluster running AKO. The CRDs can be installed on the cluster, post AKO release v1.4.1, the same way as any other AKO CRDs, via helm. More details around CRD installation can be found in the [installation guide](https://github.com/avinetworks/avi-helm-charts/docs/AKO/install/helm.md).

### Gateway APIs and Service objects
Starting v1.4.1, AKO allows users to expose Kubernetes/Opennshift Services, outside the cluster, using Gateway and GatewayClass constructs. AKO creates one Layer-4 Avi virtualservice per Gateway object, and configures the backend Services as distinct Avi Pools. In this case the type of Services, to be exposed via the Gateway object, is not limited to Service of Type `LoadBalancer`.

#### GatewayClass

GatewayClass aggregates a group of Gateway objects, similar to how IngressClass aggregates a group of Ingress objects. GatewayClasses formalize types of load balancing implementations which can be different for different load balancing vendors (Avi/Nginx/HAProxy etc.), or can point to different load balancing parameters for a single load balancing vendor (via the `parametersRef` key).

AKO identifies GatewayClasses that point to `ako.vmware.com/avi-lb` as the `.spec.controller` value, in the GatewayClass object. A sample GatewayClass object can look something like

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

It is important that the `.spec.controller` value specified MUST match `ako.vmware.com/avi-lb` for AKO to honour the GatewayClass and corresponding Gateway objects.

The `.spec.parametersRef` allows users to point to AKO's AviInfraSetting Custom Resource (cluster-scoped), to fine tune Avi specific load balancing parameters like the VIP network, Service Engine Group etc. More information on AviInfraSetting CRD can be found [here](https://github.com/avinetworks/avi-helm-charts/blob/master/docs/AKO/crds/avinfrasetting.md)


#### Gateway

The Gateway object provides a way to configure multiple Services as backends to the Gateway using label matching. The labels are specified as constant key-value pairs, the keys being `ako.vmware.com/gateway-namespace` and `ako.vmware.com/gateway-name`. The values corresponding to these keys must match the Gateway namespace and name respectively, in order for AKO to consider the Gateway valid. 
In case any one of the label keys are not provided as part of `matchLabels` OR the namespace/name provided in the label values do no match the actual Gateway namespace/name, AKO will consider the Gateway INVALID.

```
kind: Gateway
apiVersion: networking.x-k8s.io/v1alpha1
metadata:
  name: my-gateway
  namespace: blue
spec:
  gatewayClassName: avi-lb
  listeners:
  - protocol: TCP
    port: 80
    routes:
      selector:
        matchLabels:
          ako.vmware.com/gateway-namespace: blue
          ako.vmware.com/gateway-name: my-gateway
      group: v1
      kind: Service
  - protocol: TCP
    port: 8081
    routes:
      selector:
        matchLabels:
          ako.vmware.com/gateway-namespace: blue
          ako.vmware.com/gateway-name: my-gateway
      group: v1
      kind: Service
```

This Gateway object would correspond to a single Layer 4 virtualservice in Avi, with two TCP ports (80, 8081) exposed via the L4 virtualservice.

Users can also configure a user preferred static IPv4 address in the Gateway Object using the `.spec.addresses` field as shown below. This would configure the Layer 4 virtualservice with a static IP as provided in the Gateway Object.


```
spec:
  addresses:
  - type: IPAddress
    value: 10.10.10.11
```

AKO only supports assigning a single IPv4 address to the Layer 4 virtualservice. 

Avi does not allow users to update preferred virtual IPs bound to a particular virtualservice. Therefore in order to update the user preferred IP, it is required to re-create the Gateway object, failing which Avi/AKO throws an error. The following transition cases should be kept in mind, and for these, an explicit Gateway re-create with changed configuration is required.
 - updating IPAddress value, from `value: 10.10.10.11` to `value: 10.10.10.22`.
 - adding IPAddress entry after the Gateway is assigned an IP from Avi.
 - removing IPAddress entry after the Gateway is assigned an IP from Avi.

Recreating the Gateway object does the following:
 - deletes the Layer 4 virtualservice in Avi
 - frees up the applied virtual IP.
 - Re-creates the virtual service with, the intended configuration.


#### Service

Matching Gateways with backend Services via label selection, requires Services to have the same Labels as shown in the example below.

```
apiVersion: v1
kind: Service
metadata:
  name: avisvc-advlb
  namespace: blue
  labels:
    ako.vmware.com/gateway-name: my-gateway
    ako.vmware.com/gateway-namespace: blue
spec:
  type: LoadBalancer
  ports:
  - port: 8081
    name: eighty-eighty-one
    targetPort: 8080
    protocol: TCP
  selector:
    app: avi-server-one
---
apiVersion: v1
kind: Service
metadata:
  name: avisvc-advlb
  namespace: red
  labels:
    ako.vmware.com/gateway-name: my-gateway
    ako.vmware.com/gateway-namespace: blue
spec:
  type: LoadBalancer
  ports:
  - port: 80
    name: eighty-eighty
    targetPort: 8080
    protocol: TCP
  selector:
    app: avi-server-two
```

Each Service with the appropriate labels, corresponds to a single Avi Pool.
Note that the Service namespace is not required to be in the same namespace as that of the parent Gateway.
