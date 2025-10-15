### L4Rule 

L4Rule CRD can be used to modify the default properties of the L4 VS and the pools created from a service of Type LoadBalancer.
Service of type LoadBalancer has to be annotated with the name of the CRD to attach the CRD to the service. Although cross namespace usage is allowed between L4Rule and LB service, both should use the same tenant.

A sample L4Rule CRD looks like this:

```yaml
  apiVersion: ako.vmware.com/v1alpha2
  kind: L4Rule
  metadata:
    name: my-l4-rule
    namespace: green
  spec:
    analyticsProfileRef: Custom-Analytics-Profile
    analyticsPolicy:
      fullClientLogs:
        enabled: true
        duration: 0
        throttle: 30
    applicationProfileRef: Custom-L4-Application-Profile
    loadBalancerIP: "49.20.193.207"
    performanceLimits:
      maxConcurrentConnections: 105
      maxThroughput: 100
    networkProfileRef: Custom-Network-Profile
    networkSecurityPolicyRef: Custom-Network-Security-Policy
    securityPolicyRef: Custom-Security-Policy
    vsDatascriptRefs:
    - Custom-DS-01
    - Custom-DS-02
    backendProperties:
    - port: 80
      protocol: TCP
      enabled: true
      applicationPersistenceProfileRef: Custom-Application-Persistence-Profile
      healthMonitorRefs:
      - Custom-HM-01
      - Custom-HM-02
      healthMonitorCrdRefs:
      - my-health-monitor
      lbAlgorithm: LB_ALGORITHM_CONSISTENT_HASH
      lbAlgorithmHash: LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER
      lbAlgorithmConsistentHashHdr: "custom-string"
      sslProfileRef: Custom-SSL-Profile
      sslKeyAndCertificateRef: Custom-Key-And-Certificate
      pkiProfileRef: Custom-PKI-Profile
      analyticsPolicy:
        enableRealtimeMetrics: true
      minServersUp: 1
    listenerProperties:
    - port: 80
      protocol: TCP
      enableSsl: true
    sslKeyAndCertificateRefs:
    - "Custom-L4-SSL-Key-Cert"
    sslProfileRef: Custom-L4-SSL-Profile
    revokeVipRoute: true
```


### Specific usage of L4Rule CRD

L4Rule CRD can be created in a given namespace where the operator desires to have more control. 
The section below walks over the details and associated rules of using each field of the L4Rule CRD.

#### Attaching L4Rule to LoadBalancer type of Services

An L4Rule is applied to a virtual service and pool created from the LoadBalancer type of Services when the l4rule is attached to the service. An L4Rule can be attached by annotating the service with the name of the L4Rule CRD with `ako.vmware.com/l4rule` as the key and `namespace/l4RuleName of the l4rule crd` as the value. Here `namespace` in the `value` field is optional. If `namespace` is not provided, L4Rule, from same namespace as that of LB service, will be attached to LB service.

```yaml
  metadata:
    annotations:
      ako.vmware.com/l4rule: <namespace>/<name-of-the-l4-rule-crd>
```

Consider the following examples showing a service `my-service` of type LoadBalancer annotated with an L4Rule `my-l4-rule`.

***Example 1***

```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: my-service
    namespace: green
    annotations:
      ako.vmware.com/l4rule: my-l4-rule
  spec:
    selector:
      app.kubernetes.io/name: MyApp
    ports:
      - protocol: TCP
        port: 80
        targetPort: 9376
    clusterIP: 10.0.171.239
    type: LoadBalancer
```
In above example, AKO will attach `my-l4-rule` L4 CRD present in `green` namespace to LB service `my-service` present in `green` namespace.

***Example 2***

```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: my-service
    namespace: default
    annotations:
      ako.vmware.com/l4rule: green/my-l4-rule
  spec:
    selector:
      app.kubernetes.io/name: MyApp
    ports:
      - protocol: TCP
        port: 80
        targetPort: 9376
    clusterIP: 10.0.171.239
    type: LoadBalancer
```
In above example, AKO will attach `my-l4-rule` L4 CRD present in `green` namespace to LB service `my-service` present in `default` namespace.

#### Express custom analytics profiles

L4Rule CRD can be used to express analytics profile references. The analytics profile reference should have been created in the AVI Controller before the CRD creation.

```yaml
    analyticsProfile: Custom-Analytics-Profile
```

The analytics profiles can be used for various Network/Health Score analytics settings, log processing, etc.

#### Configure Analytics Policy

The L4Rule CRD can be used to configure analytics policies such as enable/disable non-significant logs, throttle the number of non-significant logs per second on each SE, and the duration for which the system should capture the logs.

```yaml
    analyticsPolicy:
      fullClientLogs:
        enabled: true
        duration: 0
        throttle: 30
```

The field `throttle` supports values from **0** to **65535** and will be in effect only when `enabled` is set to **true**. AKO defaults the value of `throttle` to **10**. Set the value of `throttle` to **Zero (0)** to deactivate throttling.

By default, the AKO sets the `duration` of logging the non-significant logs to **30 minutes**. The user has to configure the `duration` as **Zero (0)** to capture the logs indefinitely.

#### Express custom Application Profiles

L4Rule CRD can be used to express application profile references. The application profile can be used to enable PROXY Protocol, rate limit the connections from a client, etc. The application profile must be created in the AVI Controller before referring to it.

```yaml
    applicationProfile: Custom-L4-Application-Profile
 ```

**NOTE**: The application profile should be of type `L4` or `L4 SSL/TLS`. If SSL is enabled for any port in [listenerProperties](#configure-listener-properties) section then application profile should be of type `L4 SSL/TLS`. `L4 SSL/TLS` is supported starting AKO 1.11.1.

#### Express custom Load Balancer IP

The `loadBalancerIP` field can be used to provide a valid preferred IPv4 address for L4 virtual services. The preferred IP must be part of the IPAM configured for the Cloud, and must not overlap with any other IP addresses already in use. In case of any misconfigurations whatsoever, AKO would fail to configure the virtual service appropriately throwing an ERROR log for the same.

```yaml
    loadBalancerIP: "49.20.193.207"
```

**NOTE**: The L4Rule CRD is not aware of any misconfigurations during its creation process, and as a result, the L4Rule will still be marked as Accepted.

#### Configure Performance limits

The L4Rule CRD can be used to configure the performance limit settings such as maximum concurrent client connections allowed, and maximum throughput per second for all clients allowed through the client side.

```yaml
    performanceLimits:
      maxConcurrentConnections: 105
      maxThroughput: 100
```

The `maxConcurrentConnections` and `maxThroughput` supports values from **0** to **65535**.

#### Express custom Network Profile

The L4Rule CRD can be used to express a custom network profile. The network profile can be used to configure either TCP/UDP proxy settings or TCP/UDP fast path settings. The network profile must be created in the AVI Controller before referring to it.

```yaml
    networkProfileRef: Custom-Network-Profile
```

The AKO defaults the network profile to `System-TCP-Proxy`.

**NOTE**: The network profile settings are dependent on the license configured in the AVI controller. Please refer to the [document](https://avinetworks.com/docs/22.1/nsx-alb-license-editions/) before configuring the profile in the CRD. Also, SSL support has been added for L4 VS starting AKO 1.11.1. If SSL is enabled for any port in [listenerProperties](#configure-listener-properties) section then network profile should be of type TCP proxy, since only a single **TCP** port definition is allowed in the LoadBalancer service for L4 SSL.

#### Express custom Network Security Policy

The L4Rule CRD can be used to express a custom network security policy. The Network security policy can be configured with rules to allow/deny/rate limit connections from a single or group of IP addresses, etc.

```yaml
    networkSecurityPolicyRef: Custom-Network-Security-Policy
```

The Network Security Policy must be created in the AVI Controller before referring to it.

#### Express custom Security Policy

The L4Rule CRD can be used to express a custom Security Policy. Security Policy is applied to the traffic of the virtual service, and it is used to specify various configuration information used to perform Distributed Denial of Service (DDoS) attacks detection and mitigation.

```yaml
    securityPolicyRef: Custom-Security-Policy
```

The Security Policy must be created in the AVI Controller before referring to it.

#### Express custom datascripts

The L4Rule CRD can be used to express datascript references. The datascript references should have been created in the AVI Controller before the CRD creation.

```yaml
    vsDatascriptRefs:
    - Custom-DS-01
    - Custom-DS-02
```

The datascripts can be used to apply custom scripts to data traffic. The order of evaluation of the datascripts is in the same order they appear in the CRD definition.

#### Configure VIP Route Revocation
The L4Rule CRD can be used to configure VIP Route revocation. When `revokeVipRoute` is set to true, the VIP route is revoked when the virtual service is marked down (OPER_DOWN) by health monitor, and similarly, it is added back when the virtual service is OPER_UP.

```yaml
    revokeVipRoute: true
```

**NOTE**: `revokeVipRoute` is only supported for NSX-T clouds, otherwise the **L4Rule** will be *rejected*. `revokeVipRoute` is also not supported with `ako.vmware.com/enable-shared-vip` annotation. If such a combination is used, AKO will ignore the `revokeVipRoute` field.

### Configure Backend Properties

The `backendProperties` section in the L4Rule can be used to configure pool settings such as custom health monitors, application persistence profiles, LB algorithms, etc. The L4Rule CRD identifies the pools based on the port and protocol, and AKO applies the configuration to it. AKO logs a WARNING if the port and protocol don't match the service's port and protocol configurations.

A sample `backendProperties` looks like this:

```yaml
    backendProperties:
    - port: 80
      protocol: TCP
      enabled: true
      applicationPersistenceProfileRef: Custom-Application-Persistence-Profile
      healthMonitorRefs:
      - Custom-HM-01
      - Custom-HM-02
      healthMonitorCrdRefs:
      - my-health-monitor
      lbAlgorithm: LB_ALGORITHM_CONSISTENT_HASH
      lbAlgorithmHash: LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER
      lbAlgorithmConsistentHashHdr: "custom-string"
      sslProfileRef: Custom-SSL-Profile
      sslKeyAndCertificateRef: Custom-Key-And-Certificate
      pkiProfileRef: Custom-PKI-Profile
      analyticsPolicy:
        enableRealtimeMetrics: true
      minServersUp: 1
```

**NOTE**: The fields `port` and `protocol` are **mandatory** and AKO uses these fields to identify the pool. The `port` and `protocol` must equal the service's port and protocol.

AKO defaults the fields `enabled` to **true** and `lbAlgorithm` to **LB_ALGORITHM_LEAST_CONNECTIONS**.

#### Enable/Disable Pool

The field `enabled` in the L4Rule can be used to enable/disable pools attached to an L4 VS. By default, the value of the field is `true` and the user has to set the value to **false** to disable the pool.

```yaml
      enabled: true # or false
```

#### Express custom Application Persistence Profile

The L4Rule CRD can be used to express a custom Application Persistence Profile reference. The Application Persistence Profile reference should have been created in the AVI Controller before the CRD creation.

```yaml
      applicationPersistenceProfileRef: Custom-Application-Persistence-Profile
```

#### Express custom Health Monitors

L4Rule CRD can be used to express custom health monitor references. The health monitor reference should have been created in the AVI Controller before the CRD creation.

```yaml
      healthMonitorRefs:
      - Custom-HM-01
      - Custom-HM-02
```

The health monitors can be used to verify server health. A server (Kubernetes pods in this case) will be marked UP only when all the health monitors return successful responses. Health monitors provided here overwrite the default health monitor configuration set by AKO i.e. `System-TCP` for TCP traffic and `System-UDP` for UDP traffic based on the service configuration.

Alternatively, you can use the HealthMonitor CRD to define custom health monitoring configurations directly in Kubernetes:

```yaml
      healthMonitorCrdRefs:
      - my-health-monitor
      - my-backup-health-monitor
```

The `healthMonitorCrdRefs` field references HealthMonitor CRD objects that must be created in the same namespace as the L4Rule. The HealthMonitor CRD is managed by the AKO CRD Operator and supports TCP, HTTP, and PING health check types with fine-grained control over health check parameters. For more details on creating HealthMonitor CRDs, see the [HealthMonitor documentation](./healthmonitor.md).

**NOTE**: `healthMonitorCrdRefs` will not be used if `healthMonitorRefs` are specified.

#### Configure LB Algorithm

The L4Rule CRD can be used to select suitable LB algorithms to effectively distribute traffic across healthy servers. The LB algorithm may be used for distributing TCP and UDP connections across servers.

A sample LB algorithm configuration is shown below.

```yaml
      lbAlgorithm: LB_ALGORITHM_CONSISTENT_HASH
      lbAlgorithmHash: LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER
      lbAlgorithmConsistentHashHdr: "custom-string"
```

The `lbAlgorithm` allows a predefined set of values, and the user can choose the desired one. Presently the following values are supported for the field `lbAlgorithm`:

```yaml
  - LB_ALGORITHM_LEAST_CONNECTIONS
  - LB_ALGORITHM_ROUND_ROBIN
  - LB_ALGORITHM_FASTEST_RESPONSE
  - LB_ALGORITHM_CONSISTENT_HASH
  - LB_ALGORITHM_LEAST_LOAD
  - LB_ALGORITHM_FEWEST_SERVERS
  - LB_ALGORITHM_RANDOM
  - LB_ALGORITHM_FEWEST_TASKS
  - LB_ALGORITHM_NEAREST_SERVER
  - LB_ALGORITHM_CORE_AFFINITY
  - LB_ALGORITHM_TOPOLOGY
```

More information about the pool algorithms can be found [here](https://avinetworks.com/docs/latest/load-balancing-algorithms/).

The `lbAlgorithmHash` field is used only when the algorithm is chosen as `LB_ALGORITHM_CONSISTENT_HASH`. Otherwise, it's not applicable. The following values are supported for `lbAlgorithmHash`:

```yaml
  - LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS
  - LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT
  - LB_ALGORITHM_CONSISTENT_HASH_URI
  - LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER
  - LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING
  - LB_ALGORITHM_CONSISTENT_HASH_CALLID
```

The `lbAlgorithmConsistentHashHdr` field is used only when the `lbAlgorithmHash` is chosen as `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER`.

#### Express custom SSL Profile

The custom SSL profile can be used to configure the desired set of SSL versions and ciphers to accept SSL/TLS terminated connections.

```yaml
      sslProfileRef: Custom-SSL-Profile
```

The custom SSL profile should have been created in the AVI Controller before the CRD creation.

#### Express custom SSL Key And Certificate

The L4Rule CRD can be used to express a custom SSL key and certificate reference. The service engines present this certificate to the backend servers. The custom SSL key and certificate should have been created in the AVI Controller before the CRD creation.

```yaml
      sslKeyAndCertificateRef: Custom-Key-And-Certificate
```

#### Express custom PKI Profile

The L4Rule CRD can be used to express a custom PKI profile reference. Once configured, the AVI controller validates the SSL certificate present by a server against the custom PKI Profile configured in the CRD. The custom PKI Profile must be created in the AVI Controller before referring to it.

```yaml
      pkiProfileRef: Custom-PKI-Profile
```

#### Configure Analytics Policy

The L4Rule CRD can be used to configure the analytics settings for the pool. Set the `enableRealtimeMetrics` to **true**/**false** to enable/disable real-time metrics for server and pool metrics.

```yaml
      analyticsPolicy:
        enableRealtimeMetrics: true # or false
```

#### Configure Minimum Servers UP to make Pool UP

The L4Rule CRD can be used to configure the minimum number of servers in the UP state for marking the pool UP. 

```yaml
      minServersUp: 1
```

**NOTE**: The value given must be equal to or less than the number of health monitors attached to the pool. 

### Configure Listener Properties

The `listenerProperties` section in the L4Rule can be used to enable/disable SSL support for L4 virtual services. Each item in the `listenerProperties` array corresponds to a port definition in the LoadBalancer service along with the option to enable SSL termination in the service/listener settings created for that port as part of the AVI virtual service. When an L4Rule object is created with listener properties, AKO identifies the service/listener setting on the virtual service based on the port and protocol, and applies the SSL configuration to it. AKO logs a WARNING if the port and protocol don't match the service's port and protocol configurations. There are also some limitations and conditions for using listener properties. Please refer to the [Conditions and Caveats](#conditions-and-caveats) section for more details.

A sample `listenerProperties` looks like this:

```yaml
    listenerProperties
    - port: 80
      protocol: TCP
      enableSsl: true
```

**NOTE**: The fields `port` and `protocol` are **mandatory** and AKO uses these fields to identify the corresponding service/listener setting in the virtual service. The `port` and `protocol` must equal the service's port and protocol. Currently, only a single `TCP` port is allowed in the LoadBalancer service definition if SSL is required to be enabled. Hence, the same limitation also applies to `listenerProperties` which can also have only one matching **TCP** based port definition.

#### Enable/Disable SSL

The field `enableSsl` in the L4Rule can be used to enable SSL termination and offload for traffic from clients for an L4 VS. The `enableSsl` field is specified for a port and AKO configures the associated service/listener setting in the VS with the value. By default, the value of this field is **false** and the user has to set the value to **true** to enable SSL.

```yaml
      enableSsl: true # or false
```

#### Express custom SSL Profile for Virtual Service

The custom SSL profile can be used to configure the desired set of SSL versions and ciphers to accept SSL/TLS terminated connections for the virtual service.

```yaml
      sslProfileRef: Custom-SSL-Profile
```

The custom SSL profile should have been created in the AVI Controller before the CRD creation.

**NOTE**: The `sslProfileRef` should only be specified when SSL is enabled for a virtual service. The L4Rule will otherwise be rejected.

#### Express custom SSL Keys And Certificates for Virtual Service

The L4Rule CRD can be used to express custom SSL key and certificate references for a virtual service. These certificates will be presented to SSL/TLS terminated connections. The custom SSL keys and certificates should have been created in the AVI Controller before the CRD creation.

```yaml
      sslKeyAndCertificateRefs:
      - "Custom-SSL-Key-Cert"
```

**NOTE**: The `sslKeyAndCertificateRefs` should only be specified when SSL is enabled for a virtual service. The L4Rule will otherwise be rejected.

#### Status Messages

The status messages are used to give instantaneous feedback to the users about the reference objects specified in the L4Rule CRD.

Following are some of the sample status messages:

##### Accepted L4Rule object

    $ kubectl get l4rule
    NAME         STATUS     AGE
    my-l4-rule   Accepted   3d5s


An L4Rule is accepted only when all the reference objects specified inside it exist in the AVI Controller.

##### Rejected L4Rule object

    $ kubectl get l4rule
    NAME            STATUS     AGE
    my-l4-rule-alt  Rejected   2d23h
    
The detailed rejection reason can be obtained from the status:

```yaml
  status:
    error: applicationprofile "My-L4-Application" not found on controller
    status: Rejected
```

#### Conditions and Caveats

##### Sharing L4Rule with Load Balancer IP

The L4Rule CRD with load balancer IP can be shared among services only when the services contain the `ako.vmware.com/enable-shared-vip` annotation. However, L4Rule cannot be shared if SSL termination is required to be enabled for the services. So, if **enableSsl** is set to true for any port in `listenerProperties` section, then that L4Rule should only be applied to a single LoadBalancer service.

##### L4Rule deletion

If an L4Rule is deleted, the L4 VSes and Pools in the AVI controller will be configured with the default values.

##### L4Rule admission

An L4Rule CRD is only admitted if all the objects referenced in it, exist in the AVI Controller. If after admission the object references are
deleted out-of-band, then AKO does not re-validate the associated HostRule CRD objects. The user needs to manually edit or delete the object
for new changes to take effect.

##### Enabling SSL with L4Rule

There are some limitations when trying to enable SSL termintaion for an L4 virtual service with L4Rule.

1. Currently, only a single `TCP` port is allowed in the LoadBalancer service definition if SSL is required to be enabled. Hence, the same limitation also applies to `listenerProperties` which can also have only one matching **TCP** based port definition along with `enableSsl` field. This is because Avi only supports SSL termination with TCP protocol and also a VS of type L4 SSL can have only one backend pool configured.

2. If **enableSsl** is set to true for any port in `listenerProperties` section then `applicationProfile` should be of type `L4 SSL/TLS`. If application profile is not of type `L4 SSL/TLS`, then L4Rule will be rejected. If `applicationProfile` is not set, then it defaults to **System-L4-Application** in the CRD, but AKO intermally sets the application profile as **System-SSL-Application** which is the default value when SSL is enabled.

3. If **enableSsl** is set to true for any port in `listenerProperties` section then `networkProfileRef` should be of type TCP proxy, since only a single **TCP** port definition is allowed in the LoadBalancer service and listener properties.

4. The `sslProfileRef` and `sslKeyAndCertificateRefs` should be set for the VS only if SSL termination is enabled for any port in `listenerProperties` and application profile is of type `L4 SSL/TLS`, otherwise the L4Rule will be rejected. If **enableSsl** is set to true but `sslProfileRef` and `sslKeyAndCertificateRefs` are not set specified in the L4Rule, then these fields will be set with their default values in Avi.
