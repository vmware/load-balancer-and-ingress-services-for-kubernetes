# Restricting FQDN to single Namespace

## Overview

In Kubernetes environment, ingresses, deployed in multiple namespaces, can have same host(FQDN). In OpenShift, when `Route Admission Policy` is `InterNamespaceAllowed`, then routes from multiple namespaces can have same host(FQDN). For such deployment, AKO combines such routes/ingresses under one Virtual Service at AviController.

With AKO 1.13.1, AKO has introduced feature to restrict FQDN to a single namespace.

## Configuration

AKO has introduced knob `fqdnReusePolicy` in `L7Settings` section of `values.yaml`.

```yaml
L7Settings:
    .
    .
    .
    fqdnReusePolicy: "InterNamespaceAllowed"
```

`fqdnReusePolicy` can be assigned to one of the two values, `InterNamespaceAllowed` or `Strict`.
When the value is `InterNamespaceAllowed`, AKO accepts ingresses with same host/FQDN from all the namespaces. This is the `default` value.

When value is `Strict`, AKO restricts FQDN to a single namespace. FQDN will be associated with the namespace which claims it first. For example, if `ingress1` in `red` namespace is deployed with `foo.avi.internal`, then with `Strict` setting, `foo.avi.internal` will be associated with `red` namespace. Now `ingress2` in the `default` namespace is deployed with `foo.avi.internal`, then AKO will reject `ingress2` with the message `host already claimed`. VirtualService and corresponding AviController objects for `ingress2` will not be created.

In `Strict` setting, AKO does not associated one FQDN with another namespace automatically if all ingresses with given FQDN is deleted from the claimed namespace. For above example, if `ingress1` in `red` is deleted and there is no other ingress in `red` namespace associated with `foo.avi.internal`, AKO will not associate `foo.avi.internal` with `ingress2` of the `default` namespace. The user has to perform a create/update operation on ingresses, associated with `foo.avi.internal`, to claim the FQDN. The user can also reboot the AKO to associate `foo.avi.internal` with the `default` namespace.

For ingresses with multiple hosts(FQDNs), if one of the FQDNs is not accepted by AKO then whole ingress will not be accepted by AKO and configuration defined in that ingress will not be applied at AviController side.

AKO has above similar behaviour for OpenShift Routes under this knob.

**Note:**
1. Setting `fqdnReusePolicy` is applicable only in EVH deployment of AKO.
2. This setting is not applicable to GatewayAPI objects.
3. Change in the value of `fqdnReusePolicy` requires AKO reboot.

