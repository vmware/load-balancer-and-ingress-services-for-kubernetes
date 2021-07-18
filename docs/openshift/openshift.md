
#### Pre-requisites for running AKO in Openshift Cluster

Follow the steps 1 to 2, given in section [Pre-requisites](https://github.com/avinetworks/avi-helm-charts/tree/master/docs/AKO#pre-requisites). Additionally, the following points have to be noted for openshift environment.
1. Make Sure Openshift version is >= 4.4
2. Openshift routes and services of type load balancer are supported in AKO
3. Ingresses, if created in the openshift cluster won't be handled by AKO.
4. cniPlugin should be set to **openshift**
5. Set `state_based_dns_registration` to false in AVI cloud configuration. Follow the instructions mentioned in https://avinetworks.com/docs/20.1/dns-record-additions-independent-of-vs-state/.

#### Features of Openshift Route supported in AKO
AKO supports the following Layer 7 functions of the OpenShift Route object:
1. Insecure Routes.
2. Insecure Routes with alternate backends.
3. Secure routes with edge termination policy.
4. Secure Routes with InsecureEdgeTerminationPolicy - Redirect or Allow.
5. Secure Routes of type passthrough
6. Secure Routes with re-encrypt functionality

### Default Secret for TLS Routes

This feature can be used when the user wants to apply a common key-cert for multiple routes, e.g. a wild carded secret which can be used for all host names in the same subdomain. 

By default AKO expects all routes with TLS termination to have key and cert specified in the route spec. But to handle such use cases, AKO supports TLS routes without key/cert specified in the Route spec.

In this case, the common key-cert can be specified in a secret that can be used for TLS routes that don't have key/cert specified in the route spec.

To use this feature, a secret with name `router-certs-default` has to be created in the same namespace where AKO pod is running (avi-system). The secret must have tls.crt and tls.key fields in its data section.

An example of the default secret is given bellow:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: router-certs-default
  namespace: avi-system
type: kubernetes.io/tls
data:
  tls.crt: 
    -----BEGIN CERTIFICATE-----
    [...]
    -----END CERTIFICATE-----
  tls.key:
    -----BEGIN PRIVATE KEY-----
    [...]
    -----END PRIVATE KEY-----
```

After creating the secret, we can add a secure route without without key or cert in the spec, for example:

```yaml
apiVersion: v1
kind: Route
metadata:
  name: secure-route-no-cert
spec:
  host: secure-no-cert.avi.internal
  to:
    kind: Service
    name: avisvc
  tls:
    termination: edge
```

AKO would use the default secret to fetch key and cert values for processing all such routes.

Regarding the default secret, following points have to be notes.
- For TLS routes with termination type reencrypt, the value of destinationCA has to be specified in the route spec itself.
- caCertificate can not be specified as part of the default secret.
- `router-certs-default` present in `openshift-ingress` namespace is not used by AKO. Users have to create `router-certs-default` in `avi-system` namespace.

