
#### Pre-requisites for running AKO in Openshift Cluster

Follow the steps 1 to 2, given in section [Pre-requisites](https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/blob/master/docs#pre-requisites). Additionally, the following points have to be noted for openshift environment.
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

To use this feature, a secret with name `router-certs-default` has to be created in the same namespace where AKO pod is running (avi-system). The secret must have tls.crt and tls.key fields in its data section. Additonally, alt.crt and alt.key can be fields can be populated to allow multiple default certificates when trying to configure both RSA and ECC signed certificates. Avi Controller allows a Virtual Service to be configured with two certificates at a time, one each of RSA and ECC. This enables Avi Controller to negotiate the optimal algorithm or cipher with the client. If the client supports ECC, in that case the ECC algorithm is preferred, and RSA is used as a fallback in cases where the clients do not support ECC.

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
  alt.key:
    -----BEGIN PRIVATE KEY-----
    [...]
    -----END PRIVATE KEY-----
  alt.crt:
    -----BEGIN CERTIFICATE-----
    [...]
    -----END CERTIFICATE-----
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

