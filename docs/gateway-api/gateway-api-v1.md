## Gateway API

AKO claims support for the v1 release of Gateway API. To enable this feature the field `GatewayAPI` under section `featureGates` in values.yaml must be set to **true**. This will spin up a new container within the AKO pod which will handle the Gateway API related objects.

In the current release, the following objects in the Gateway API are supported by AKO:
  1. GatewayClass (v1)
  2. Gateway (v1)
  3. HTTPRoute (v1)

**NOTE:** AKO Gateway API supports all the fields which are mentioned as **Support: Core** in the above objects for the current release(with a few exceptions. See limitations below). Other objects in the Gateway API and fields in the GatewayClass, Gateway and HTTPRoute will be supported in the future releases.

### Support Matrix

|                | GatewayClass | Gateway | HTTPRoute |   GRPCRoute   |   TLSRoute    |   TCPRoute    |   UDPRoute    | ReferenceGrant |
|:--------------:|:------------:|:-------:|:---------:|:-------------:|:-------------:|:-------------:|:-------------:|----------------|
| release-1.13.1 |      v1      |   v1    |    v1     | Not Supported | Not Supported | Not Supported | Not Supported | Not Supported  |

### Installation

The steps mentioned below must be followed to enable the GatewayAPI feature in AKO:

  1. Set the `GatewayAPI` under the section `featureGates` to **true** in the `values.yaml`
  2. Do a helm install or upgrade using the edited `values.yaml` to install or upgrade the AKO

A GatewayClass `avi-lb` with `controllerName` as `ako.vmware.com/avi-lb` will get installed as part of the installation. An Infrastructure Provider can ask the cluster operators to use this GatewayClass in their Gateway objects so that the AKO honours the objects created by them.

**NOTE:** The GatewayClass, Gateway, and Route CRD definitions must be installed on the cluster before enabling the GatewayAPI feature in AKO. The CRDs can be found [here](https://github.com/kubernetes-sigs/gateway-api/tree/main/config/crd/standard).

### Gateway API Objects

#### GatewayClass

GatewayClass aggregates a group of Gateway objects, similar to how IngressClass aggregates a group of Ingress objects. GatewayClasses formalizes types of load balancing implementations which can be different for different load-balancing vendors (Avi/Nginx/HAProxy etc.).

AKO identifies GatewayClasses that point to `ako.vmware.com/avi-lb` as the `.spec.controllerName` value, in the GatewayClass object. 

A sample GatewayClass object can look something like this:

  ```yaml
  apiVersion: gateway.networking.k8s.io/v1
  kind: GatewayClass
  metadata:
    name: avi-lb
  spec:
    controllerName: "ako.vmware.com/avi-lb"
  ```

It is important that the `.spec.controllerName` value specified MUST match `ako.vmware.com/avi-lb` for AKO to honor the GatewayClass and corresponding Gateway API objects.

**NOTE:** A GatewayClass named `avi-lb` will get installed as part of the helm install or upgrade when the user enables the Gateway API feature. 

#### Gateway

The Gateway object represents an instance of a service-traffic handling infrastructure by binding Listeners to a set of IP addresses. The AKO validates the Gateway object and checks if the listeners are valid and sets `ListenerConditionAccepted` to `true`. If one or more listener in a gateway is valid, `GatewayConditionAccepted` is set to `true`. It then checks whether all the references specified in the gateway listeners are valid and exist and then sets `ListenerConditionResolvedRefs` to `true` accordingly. It then translates the Gateway and its configuration to a Parent VS. The listeners in Gateway is mapped to service ports in Parent VS and secrets will be configured as Certificates in the AVI controller and will be referenced in the same Parent VS. AKO updates the status of Gateway with `GatewayConditionProgrammed` as `true` along with the VIP of the Parent VS once the VS creation is completed and sets `ListenerConditionProgrammed` to `true` if respective listener is programmed.

The parent VS created by AKO follows the naming convention `ako-gw-<cluster-name>--<namespace of the gateway>-<name of the gateway>-EVH`

A sample Gateway object is shown below:

  ```yaml
  apiVersion: gateway.networking.k8s.io/v1
  kind: Gateway
  metadata:
    name: my-gateway
  spec:
    gatewayClassName: avi-lb
    listeners:
    - name: foo-http
      protocol: HTTP
      port: 80
      hostname: *.example.com
    - name: bar-https
      protocol: HTTPS
      port: 443
      hostname: *.example.com
      tls:
        certificateRefs:
        - kind: Secret
          group: ""
          name: bar-example-com-cert
  ```

The above Gateway object would correspond to a single Layer-7 Virtual Service in the AVI controller, with two ports (80, 443) exposed and an sslKeyAndCertificate created based on the Secret **bar-example-com-cert**.

AKO only supports HTTP and HTTPS as protocol.

AKO only supports Secret kind for certificateRefs.

Users can also configure a user-preferred static IPv4 address in the Gateway Object using the `.spec.addresses` field as shown below. This would configure the Layer 7 virtual service with a static IP as mentioned in the Gateway Object.

  ```yaml
  spec:
    addresses:
    - type: IPAddress
      value: 10.1.1.10
  ```

**NOTE:** AKO claims support for a single address of type IPAddress. The Gateway must be re-created to update the address.

#### HTTPRoute

The HTTPRoute object provides a way to route HTTP requests. The AKO models a child VS based on this object. AKO supports match requests based on the hostname, path, and header specified. The filters to specify additional processing of the requests will be added as policy in the child VS by the AKO. The filters of type `RequestHeaderModifier`, `RequestRedirect` and `ResponseHeaderModifier` are supported in the current release.

A sample HTTPRoute object is shown below:

  ```yaml
  apiVersion: gateway.networking.k8s.io/v1
  kind: HTTPRoute
  metadata:
    name: my-http-app
  spec:
    parentRefs:
    - name: my-gateway
    hostnames:
    - "foo.example.com"
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /bar
      backendRefs:
      - name: my-service1
        port: 8080
    - filters:
      - type: RequestHeaderModifier
        requestHeaderModifier:
          add:
            - name: my-header
              value: foo
      matches:
      - headers:
        - type: Exact
          name: magic
          value: foo
        path:
          type: PathPrefix
          value: /foo
      backendRefs:
      - name: my-service2
        port: 8080
        weight: 1
      - name: my-service3
        port: 8081
        weight: 2
  ```

The above HTTPRoute object gets translated to two child VSes in the AVI controller. One child VS with match criteria as the path begins with `/bar` and a single Pool Group with a single pool and another child VS with match criteria as path begins with `/foo`, a single Pool Group with two pools, and an HTTP Request policy to add `my-header` to the HTTP request forwarded to the backends.

Hostnames are mandatory and can be prefixed with a wildcard.

AKO Gateway APIs does not support `filters` within `backendRefs`.

### Gateway API Objects to AVI Controller Objects Mapping

In AKO Gateway API Implementation, Gateway objects corresponds to following AVI Controller objects:

  1. `Gateway` translates to an `EVH Parent Virtual Service` with `port/protocol` from each `Listener` added as the `Service` within parent VS.
  2. `tls specification (certificate refs)` from each `Gateway Listener` will get added to the parent VS as `SSLKeyAndCertificateRef`. 
  3. Every `Secret` created corresponds to an `SSLKeyAndCertificate` object.
  4. `Addresses` in a Gateway specification gets added as static ip for `Vsvip` for parent VS.
  5. Each `hostname`, except `wild card hostnames`, specified in a Gateway listener is mapped to the `DNS` in the `Vsvip` of the parent VS.
  6. `hostname` specified in HTTPRoute gets mapped to `VHMatch` in the childVS.
  7. Every `Rule` in `HTTPRoute` corresponds to an `EVH Child Virtual Service`, with `Match` translated to `VH match` and `Filters` translated to `HTTPPolicySet` configuration. If no matches are specified for a particular HTTPRoute rule, a childVS with path as `/` in `VHmatch` will be created and attached to the parentVS corresponding to that rule.
  8. Each `backendRefs` specification (list of backends) in a `HTTPRoute Rule` will be added as a `Pool Group`.
  9. Each `backendRef` in a `HTTPRoute Rule` will be translated to a `Pool`.
  10. Every parentVS will have a default `HTTPPolicyset` attached to it which will return `404`, if no path matches a given HTTP request.     

### Naming Conventions

AKO Gateway Implementation follows following naming convention:

  1. ParentVS              `ako-gw-<cluster-name>--<gateway-namespace>-<gateway-name>-EVH`
  2. ChildVS               `ako-gw-<cluster-name>–-<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>>` 
  3. Pool                  `ako-gw-<cluster-name>--<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>-<backendRefs_namespace>-<backendRefs_name>-<backendRefs_port>>` 
  4. PoolGroup             `ako-gw-<cluster-name>–-<sha1 hash of <gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-<stringified FNV1a_32 hash of bytes(jsonified match)>>` 
  5. SSLKeyAndCertificate  `ako-gw-<cluster-name>--<sha1 hash of <gateway-namespace>-<gateway-name>-<secret-namespace>-<secret-name>>`
  6. DefaultHTTPPolicySet  `ako-gw-<cluster-name>--default-backend`

### Wildcard handling in Hostnames

In the current release, AKO Gateway will support following hostname combinations in Gateway listener and HTTPRoute:

  1. Wildcard prefixed hostname(like `*.avi.internal`) at Gateway listener and Non wildcard hostname(like `bar.avi.internal`) at HTTPRoute.
  2. Non-wildcard hotsname(like `bar.avi.internal`) at Gateway listener and wildcard prefixed hostname(like `*.avi.internal`) at HTTPRoute.
  3. Empty hotsname at Gateway listener(like `*`) and Non-wildcard and wildcard prefixed hostname(like `bar.avi.internal/*.avi.internal`) at HTTPRoute.

### HTTP Traffic Splitting

In the current release, AKO Gateway will support the Canary and Blue-Green traffic rollout. The configurations corresponding to this can be found [here](https://gateway-api.sigs.k8s.io/guides/traffic-splitting/)

### Status of Gateway API objects

AKO updates the status of all Gateway API objects with proper reasons. A typical status consists of a reason for the acceptance or rejection using which a user can debug the Gateway API object configuration.

A sample status of the Gateway is shown below:

  ```yaml
  Status:
    Addresses:
      Type:   IPAddress
      Value:  100.66.176.6
    Conditions:
      Last Transition Time:  2024-12-02T07:20:25Z
      Message:               Gateway configuration is valid
      Observed Generation:   1
      Reason:                Accepted
      Status:                True
      Type:                  Accepted
      Last Transition Time:  2024-12-02T07:21:20Z
      Message:               Virtual service configured/updated
      Observed Generation:   1
      Reason:                Programmed
      Status:                True
      Type:                  Programmed
    Listeners:
      Attached Routes:  0
      Conditions:
        Last Transition Time:  2024-12-02T07:20:25Z
        Message:               Listener is valid
        Observed Generation:   1
        Reason:                Accepted
        Status:                True
        Type:                  Accepted
        Last Transition Time:  2024-12-02T07:20:25Z
        Message:               All the references are valid
        Observed Generation:   1
        Reason:                ResolvedRefs
        Status:                True
        Type:                  ResolvedRefs
        Last Transition Time:  2024-12-02T07:21:20Z
        Message:               Virtual service configured/updated
        Observed Generation:   1
        Reason:                Programmed
        Status:                True
        Type:                  Programmed
      Name:                    https
      Supported Kinds:
        Group:  gateway.networking.k8s.io
        Kind:   HTTPRoute
  ```

A sample HTTPRoute status is shown below:

  ```yaml
  Status:
    Parents:
      Conditions:
        Last Transition Time:  2024-12-02T07:28:39Z
        Message:               Parent reference is valid
        Observed Generation:   1
        Reason:                Accepted
        Status:                True
        Type:                  Accepted
        Last Transition Time:  2024-12-02T07:28:39Z
        Message:
        Observed Generation:   1
        Reason:                ResolvedRefs
        Status:                True
        Type:                  ResolvedRefs
      Controller Name:         ako.vmware.com/avi-lb
      Parent Ref:
        Group:         gateway.networking.k8s.io
        Kind:          Gateway
        Name:          test-gateway
        Namespace:     test-httproute-ns
       Section Name:  https
  ```

### Conditions and Caveats

#### Gateway Limitations

AKO accepts the following Gateway configuration for this release:
  
  1. Gateway MUST contain at least one listener configuration in it.
  2. Gateway MUST NOT contain protocols other than HTTP or HTTPS.
  3. Gateway MUST NOT contain TLS modes other than `Terminate`.
  4. Two Gateways MUST NOT have listeners with same/overlapping hostname.
  5. `Gateway-> listeners-> allowedRoutes-> namespaces-> from` with value `selector` and thus `Gateway-> listeners-> allowedRoutes-> namespaces-> selector` is not supported.
  

#### HTTPRoute Limitations

AKO accepts the following HTTPRoute configuration for this release:

  1. HTTPRoute MUST contain at least one parent reference.
  2. HTTPRoute MUST NOT contain `*` as hostname.
  3. HTTPRoute MUST specify atleat one hostname.
  4. HTTPRoute MUST contain at least one hostname match with parent Gateway.
  5. Filters nested inside BackendRefs are not supported i.e. `HTTPRoute-> spec-> rules-> backendRefs-> filters` are not supported whereas `HTTPRoute-> spec-> rules-> filters` is supported.

#### Resource Creation

AKO Gateway API imposes a restriction on the order of GatewayClass and Gateway creation i.e. GatewayClass must be created before Gateway.

AKO allows creation/updation of HTTPRoute and Gateway in any order except for following cases which requires an explicit AKO reboot to work.
 1. When one gateway attached to multiple routes, having all valid listeners, transitions to gateway with some invalid listeners.
 2. When one gateway attached to multiple routes, having all valid listeners, transitions to gateway with all invalid listeners.
 3. When one gateway attached to multiple routes, having some valid listeners, transitions to gateway with all invalid listeners.
 4. When one gateway attached to one route, having some valid listeners, transitions to gateway with all invalid listeners.
 5. When one gateway attached to one route, having all valid listeners, transitions to gateway with all invalid listeners.

#### High Availability Support

AKO in active-stand-by mode does not support Gateway APIs. Avi continues to deliver highly available L4 and L7 LBs for Gateway APIs.

#### Configuring Static IP address

The AKO supports Gateway objects with a single IPv4 address. The user can configure their preferred static IPv4 address by specifying `spec.addresses` in the Gateway object. A sample configuration is shown below:

  ```yaml
  spec:
    addresses:
    - type: IPAddress
      value: 10.1.1.10
  ```

**NOTE:** The address of type IPAddress is only supported. The length of the addresses is also limited to a single address.

#### Kubernetes Service types

AKO Gateway API supports `ClusterIP`, `NodePort` and `NodePortLocal` Service types.

#### Container Network Interface (CNI) providers

AKO Gateway API claims support for `Calico` and `Antrea` as CNI provider.

#### Platforms/ Environments supported

AKO Gateway API is supported on `Kubernetes` platform.

#### Cloud Service Providers

AKO Gateway API claims support for `VMware vCenter/vSphere ESX` and `NSX-T` Cloud.