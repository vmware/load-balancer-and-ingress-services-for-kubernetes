### L4Rule 

L4Rule CRD can be used to modify the default properties of the L4 VS and the pools created from a service of Type LoadBalancer.
Service of type LoadBalancer has to be annotated with the name of the CRD to attach the CRD to the service.

**NOTE**: L4Rule CRD with Gateway API is not supported currently.

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

**NOTE**: The L4Rule CRD must be configured in the same namespace as the service of type LoadBalancer.

### Specific usage of L4Rule CRD

L4Rule CRD can be created in a given namespace where the operator desires to have more control. 
The section below walks over the details and associated rules of using each field of the L4Rule CRD.

#### Attaching L4Rule to LoadBalancer type of Services

An L4Rule is applied to a virtual service and pool created from the LoadBalancer type of Services when the l4rule is attached to the service. An L4Rule can be attached by annotating the service with the name of the L4Rule CRD with `ako.vmware.com/l4rule` as the key and `name of the l4rule crd` as the value.

```yaml
  metadata:
    annotation:
      ako.vmware.com/l4rule: <name-of-the-l4-rule-crd>
```

Consider the following example showing a service `my-service` of type LoadBalancer annotated with an L4Rule `my-l4-rule`.

```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: my-service
    annotation:
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

**NOTE**: The application profile should be of type `L4`. `L4 SSL/TLS` is not supported currently.

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

**NOTE**: The network profile settings are dependent on the license configured in the AVI controller. Please refer to the [document](https://avinetworks.com/docs/22.1/nsx-alb-license-editions/) before configuring the profile in the CRD.

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

The L4Rule CRD with load balancer IP can be shared among services only when the services contain the `ako.vmware.com/enable-shared-vip` annotation.

##### L4Rule deletion

If an L4Rule is deleted, the L4 VSes and Pools in the AVI controller will be configured with the default values.

##### L4Rule admission

An L4Rule CRD is only admitted if all the objects referenced in it, exist in the AVI Controller. If after admission the object references are
deleted out-of-band, then AKO does not re-validate the associated HostRule CRD objects. The user needs to manually edit or delete the object
for new changes to take effect.
