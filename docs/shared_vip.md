AKO provides support for sharing VIP among multiple kubernetes Services of type LoadBalancer deployed in the same namespace. Generally, with LoadBalancer Services, AKO creates dedicated L4 Virtual Services in the Avi Controller, but multiple LoadBalancer Services can also be combined to share a single VIP.

This can be achieved by providing an annotation to multiple LoadBalancer Services, where VIP sharing is intended.
The annotation to be applied is `ako.vmware.com/enable-shared-vip` with a string value as shown below:

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    ako.vmware.com/enable-shared-vip: "shared-vip-key-1"
  name: sharedvip-avisvc-lb1
  namespace: default
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: avi-server
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    ako.vmware.com/enable-shared-vip: "shared-vip-key-1"
  name: sharedvip-avisvc-lb2
  namespace: default
spec:
  type: LoadBalancer
  ports:
  - port: 80
    protocol: UDP
    targetPort: 8080
  selector:
    app: avi-server
```

The above Service spec would result in AKO creating a single L4 Virtual Service (with a single VIP) based on the annotation value, and the Port, Protocol, App Selector information will be used to configure Pools and Backend Servers for this Virtual Service.
After the successful creation of the corresponding Virtual Service and VIP, the Status of both the LoadBalancer Services will reflect the single VIP configured on the Avi controller.

```
NAME                      TYPE           CLUSTER-IP       EXTERNAL-IP     PORT(S)        AGE
sharedvip-avisvc-lb1      LoadBalancer   10.108.153.227   100.64.196.61   80:31658/TCP   6d23h
sharedvip-avisvc-lb2      LoadBalancer   10.102.147.29    100.64.196.61   80:31331/UDP   6d23h
```

In case their is a requirement to set a preferred static VIP through the `.spec.loadBalancerIP` field in the Service, it is required that all LB Services sharing the Annotation value must have the same preferred VIP provided in the spec. If two Services under the same Annotation value have different static VIP set, no Virtual Service will be configured. This is treated as a misconfiguration and will be logged in AKO accordingly.

An example of configuring multiple LB Services to share a preferred VIP is shown below.

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    ako.vmware.com/enable-shared-vip: "shared-vip-key-1"
  name: sharedvip-avisvc-lb1
  namespace: default
spec:
  type: LoadBalancer
  loadBalancerIP: 100.64.196.75
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: avi-server
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    ako.vmware.com/enable-shared-vip: "shared-vip-key-1"
  name: sharedvip-avisvc-lb2
  namespace: default
spec:
  type: LoadBalancer
  loadBalancerIP: 100.64.196.75
  ports:
  - port: 80
    protocol: UDP
    targetPort: 8080
  selector:
    app: avi-server
```

The expected status message should have the VIP matching the preferred static IP provided in the Service spec.
```
NAME                      TYPE           CLUSTER-IP       EXTERNAL-IP     PORT(S)        AGE
sharedvip-avisvc-lb1      LoadBalancer   10.108.153.227   100.64.196.75   80:31658/TCP   6d23h
sharedvip-avisvc-lb2      LoadBalancer   10.102.147.29    100.64.196.75   80:31331/UDP   6d23h
```

As `spec.loadBalancerIP` field has been deprecated from kubernetes version 1.24 onwards, AKO has introduced a new annotation `ako.vmware.com/load-balancer-ip` to L4 service that will be used to specify preferred IP. Example usage could look something like this:

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    ako.vmware.com/enable-shared-vip: "shared-vip-key-1"
    ako.vmware.com/load-balancer-ip: "100.64.196.75"
  name: sharedvip-avisvc-lb1
  namespace: default
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: avi-server
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    ako.vmware.com/enable-shared-vip: "shared-vip-key-1"
    ako.vmware.com/load-balancer-ip: "100.64.196.75"
  name: sharedvip-avisvc-lb2
  namespace: default
spec:
  type: LoadBalancer
  ports:
  - port: 80
    protocol: UDP
    targetPort: 8080
  selector:
    app: avi-server
```

There are a few things that must be considered while configuring the Services with the aforementioned annotation:
1. Make sure that LoadBalancer Services which are intended to share a VIP, must have the **same** annotation values. As shown in the example above, the annotation values `shared-vip-key-1` and `100.64.196.75` are same for both Services.
2. In order to avoid any errors while configuring the Virtual Service on the Avi controller, it is required that there is no conflicting Port-Protocol pairs in the LB Services that share the Annotation value. With the example shown above both Services are exposing a unique, non-conflicting Port-Protocol for the backend application i.e. 80/TCP and 80/UDP. Explicit checks will be added around this to ensure misconfigurations in future AKO releases.
3. The annotation must be provided only on Service of type LoadBalancers.

An L4Rule CRD can also be used to specify a preferred IP for the LoadBalancers. Please refer to [this](crds/l4rule.md#express-custom-load-balancer-ip) document for more information. However, L4Rule cannot be used for services with shared vip if **SSL** termination is required to be enabled for the services.

## AviInfrasetting Support

AviInfraSetting resources can be attached to LoadBalancer kubernetes services using annotation `aviinfrasetting.ako.vmware.com/name: <aviinfra-crd-name>`. Details can be found out [here](crds/avinfrasetting.md)

> ***Note***: Make sure that LoadBalancer services which are intended to share a VIP, must have **same** avinfrasetting annotation value.

## L4Rule CRD Support

An L4Rule CRD can be attached to the Services of type LoadBalancer that are intended to share the VIP using the annotation `ako.vmware.com/l4rule: <name-of-the-l4-rule-crd>`. More details of the L4Rule CRD can be found [here](crds/l4rule.md). However, there is an exception if SSL termination is required to be enabled for the services. So, if **enableSsl** is set to true for any port in `listenerProperties` section of the L4Rule, then that L4Rule should only be applied to a signle LoadBalancer service. This exception is because a VS of type L4 SSL can have only one backend pool configured.

> ***Note***: The Services of type LoadBalancer, which are intended to share a VIP, must have **same** L4Rule CRD annotation value.