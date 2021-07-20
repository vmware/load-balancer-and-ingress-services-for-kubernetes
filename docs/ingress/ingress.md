AKO, from v1.3.1 claims partial support for networking/v1 Ingress, released for general availability starting kubernetes 1.19. 


#### networking/v1 Ingress specific features supported in AKO

1. IngressClass
2. Default IngressClass

AKO automatically detects whether ingress-class api is enabled/available in the cluster it is operating in. If the ingress-class api is enabled, AKO switches to use the IngressClass objects, instead of the previously available alternative of using `kubernetes.io/ingress.class` annotations in Ingress objects. 

#### Avi IngressClass object
IngressClass corresponding to AKO as the ingress controller gets deployed as part of helm install/upgrade (AKO v1.3.1+). Helm autodetects the presence  of IngressClass api enabled on the cluster, and if it does, creates the IngressClass object. The IngressClass object should look something like this:

```
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: avi-lb
spec:
  controller: ako.vmware.com/avi-lb
  parameters:
    apiGroup: ako.vmware.com
    kind: IngressParameters
    name: external-lb
```

Although, it is OK to use any other names for the IngressClass, it is important that the `.spec.controller` value specified MUST match `ako.vmware.com/avi-lb`.

As part of the helm install/upgrade, if the `defaultIngController` is set to `true`, AKO's helm chart would apply the `ingressclass.kubernetes.io/is-default-class` as so:

```
metadata:
  name: avi-lb
  annotations:
    ingressclass.kubernetes.io/is-default-class: "true"
```

Setting the `ingressclass.kubernetes.io/is-default-class` to `"true"` enables AKO to implement all Ingresses, even if the `ingressClassName` is not explicitly specified/kept `None` in the Ingress objects.
The `ingressclass.kubernetes.io/is-default-class` annotation comes in handy when upgrading to an IngressClass enabled cluster. This is because while upgrading Ingresses from the ingress class annotation approach to the IngressClass object approach, the upgraded Ingresses would end having `ingressClassName` set to `None`.

### Ingress and Avi IngressClass
In order to provide a controller to implement a given ingress, in addition to creating the IngressClass object, the `ingressClassName` should be speecified, that matches the IngressClass name. The ingress would look like:

```
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: my-ingress
spec:
  ingressClassName: avi-lb
  rules:
    - host: myinsecurehost.avi.internal
      http:
        paths:
        - path: /foo
          backend:
            serviceName: service1
            servicePort: 80
```

Alternatively, if the `ingressClassName` is empty, AKO checks for `ingressclass.kubernetes.io/is-default-class` to be set to true on an IngressClass belonging to AKO (with `.spec.controller: ako.vmware.com/avi-lb`).

Removing an Avi IngressClass from the cluster would delete all Ingress associated objects from Avi, therefore it is suggested to handle IngressClass with caution.

### Default Secret for Ingress

This feature can be used when the user wants to apply a common key-cert for multiple Ingresses, e.g. a wild carded secret which can be used for all host names in the same subdomain.

To use this feature: 
- A Secret with name `router-certs-default` has to be created in the same namespace where AKO pod is running (avi-system). The Secret must have tls.crt and tls.key fields in its data section.
- The annotation "ako.vmware.com/enable-tls" has to be added in the desired Ingresses with its value set to "true"

An example of the default secret is given bellow:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: router-certs-default
  namespace: avi-system
type: kubernetes.io/tls
data:
  tls.key: 
    -----BEGIN PRIVATE KEY-----
    [...]
    -----END PRIVATE KEY-----
  tls.crt:
    -----BEGIN CERTIFICATE-----
    [...]
    -----END CERTIFICATE-----
```

Example of an Ingress using this default secret via annotation is given bellow:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress1
  annotations:
    ako.vmware.com/enable-tls: "true"
spec:
  ingressClassName: avi-lb
  rules:
  - host: "ingr1.avi.internal"
    http:
      paths:
      - path: /foo
        backend:
          service:
            name: avisvc1
            port:
              number: 80
```

It has to be noted that if any Host Rule specifies a AVI SSL Key Cert for the same host, then default Secret won't be used. Similarly if a Secret is specified in the TLS section of the Ingress Spec, then the default Secret won't be used.
