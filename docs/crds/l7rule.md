### L7Rule 

L7Rule CRD can be used to modify the properties of the L7 VS which are not part of the HostRule CRD. L7Rule is applicable only when AKO is running in [EVH mode](../akoconfig.md#akoconfig-custom-resource). 


A sample L7Rule CRD looks like this:

```yaml
apiVersion: ako.vmware.com/v1alpha2
kind: L7Rule
metadata:
  name: my-l7-rule
  namespace: l7rule-ns
spec:
  allowInvalidClientCert: true
  closeClientConnOnConfigUpdate: false
  ignPoolNetReach: false
  removeListeningPortOnVsDown: false
  sslSessCacheAvgSize: 1024
  botPolicyRef: bot
  hostNameXlate: host.com
  minPoolsUp: 2
  performanceLimits:
          maxConcurrentConnections: 2000
          maxThroughput: 3000
  securityPolicyRef: secPolicy
  trafficCloneProfileRef: tcp
```

**NOTE**: The L7Rule CRD must be configured in the same namespace as HostRule.

### Specific usage of L7Rule CRD

L7Rule CRD can be created to set some of the default properties in a L7 VirtualService. 
The section below walks over the details and associated rules of using each field of the L7Rule CRD.

#### parameters
| **Parameter**                                    | **Description**                                                                                                          | **Default**                           |
| ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------ | ------------------------------------- |
| `allowInvalidClientCert`                      | Process request even if invalid client certificate is presented.                                                                                           | False                               |
| `closeClientConnOnConfigUpdate`            | Close client connection on vs config update.| False|
| `ignPoolNetReach`            | Ignore Pool servers network reachability constraints for Virtual Service placement.                                                                                      | False|
| `removeListeningPortOnVsDown`            | Remove listening port if VirtualService is down.                                                                                         | False |
| `sslSessCacheAvgSize`            | Expected number of SSL session cache entries (may be exceeded).Allowed values are 1024-16383.                                                                                         | 1024 |
| `botPolicyRef`            | Bot detection policy for the Virtual Service. It is a reference to an object of type BotDetectionPolicy.The BotDetectionPolicy reference used by VirtualService requires at least 552 MB `extra_shared_config_memory` configured in ServiceEngineGroup on Controller or else VS creation will fail.                                                                                    | Nil |
| `hostNameXlate`                         | Translate the HostName sent to the servers to this value.Translate the host name sent from servers back to the value used by the client. It is not applied on child vs                                                                                                   | Nil                                   |
| `minPoolsUp`                         | Minimum number of UP pools to mark VS up.                                                                                                   | 0                                    |
| `performanceLimits.maxConcurrentConnections`                         | The maximum number of concurrent client conections allowed to the Virtual Service. It is not applied on Child vs                               | Nil                                    | Nil
| `performanceLimits.maxThroughput`         | The maximum throughput per second for all clients allowed through the client side of the Virtual Service per SE. It is not applied on Child vs                                                                                    | Nil                               |
| `securityPolicyRef`         | Security policy applied on the traffic of the Virtual Service. This policy is used to perform security actions such as Distributed Denial of Service (DDoS) attack mitigation, etc. It is a reference to an object of type SecurityPolicy and is not applied on child vs.                                                                                       |   Nil                            |
| `trafficCloneProfileRef`          | Server network or list of servers for cloning traffic. It is a reference to an object of type TrafficCloneProfile.                                                                                     | Nil |



#### Attaching L7Rule to HostRule

An L7Rule can be specifed in HostRule spec. Respective L7Rule Properties will be applied to the VS created through corresponding Hostrule. An L7Rule can be attached in the Hostrule CRD spec with `l7Rule` as the key and `name of the l7rule crd` as the value.

```yaml
apiVersion: ako.vmware.com/v1beta1
kind: HostRule
metadata:
  name: my-host-rule
spec:
  virtualhost:
    fqdn: test-ingclass.avi.internal     
    fqdnType: Exact
    l7Rule: my-l7-rule
```



#### Status Messages

The status messages are used to give instantaneous feedback to the users about the reference objects specified in the L7Rule CRD.

Following are some of the sample status messages:

##### Accepted L7Rule object

    $ kubectl get l7rule
    NAME         STATUS     AGE
    my-l7-rule   Accepted   3d5s


An L7Rule is accepted only when all the reference objects specified inside it exist in the AVI Controller.

##### Rejected L7Rule object

    $ kubectl get l7rule
    NAME            STATUS     AGE
    my-l7-rule-alt  Rejected   2d23h
    
The detailed rejection reason can be obtained from the status:

```yaml
  status:
    error: botPolicyRef "My-L7-Application" not found on controller
    status: Rejected
```

#### Conditions and Caveats

##### L7Rule deletion

If an L7Rule is deleted, the corresponding fields in L7 VSes in the AVI controller will be configured with the default values.

##### HostRule deletion

If an HostRule referencing an L7Rule is deleted , the corresponding fields in L7 VSes in the AVI controller will be configured with the default values.

##### L7Rule admission

An L7Rule CRD is only admitted if all the objects referenced in it, exist in the AVI Controller. If after admission the object references are
deleted out-of-band, then AKO does not re-validate the associated HostRule CRD objects. The user needs to manually edit or delete the object
for new changes to take effect.
