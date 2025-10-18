### L7Rule 

L7Rule CRD can be used to modify the properties of the L7 VS which are not part of the HostRule CRD. L7Rule is applicable only when AKO is running in [EVH mode](../ako_evh.md).

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
  analyticsProfile:
    kind: AviRef
    name: my-analytics-profile
  applicationProfile:
    kind: AviRef
    name: my-application-profile
  wafPolicy:
    kind: AviRef
    name: my-waf-policy
  icapProfile:
    kind: AviRef
    name: my-icap-profile
  errorPageProfile:
    kind: AviRef
    name: my-error-page-profile
  analyticsPolicy:
    fullClientLogs:
      enabled: false
      throttle: HIGH
      duration: 0
    logAllHeaders: false
  httpPolicy:
    overwrite: true
    policySets:
      - policy-set-1
      - policy-set-2
```

**NOTE**: The L7Rule CRD must be configured in the same namespace as HostRule.

### Comprehensive Examples

#### Basic L7Rule Example
```yaml
apiVersion: ako.vmware.com/v1alpha2
kind: L7Rule
metadata:
  name: basic-l7-rule
  namespace: default
spec:
  allowInvalidClientCert: false
  closeClientConnOnConfigUpdate: true
  ignPoolNetReach: false
  removeListeningPortOnVsDown: true
  sslSessCacheAvgSize: 2048
  minPoolsUp: 1
  performanceLimits:
    maxConcurrentConnections: 1000
    maxThroughput: 2000
```

#### Advanced L7Rule with All Profiles
```yaml
apiVersion: ako.vmware.com/v1alpha2
kind: L7Rule
metadata:
  name: advanced-l7-rule
  namespace: production
spec:
  # Basic settings
  allowInvalidClientCert: true
  closeClientConnOnConfigUpdate: false
  ignPoolNetReach: true
  removeListeningPortOnVsDown: false
  sslSessCacheAvgSize: 4096
  minPoolsUp: 2

  # Performance limits
  performanceLimits:
    maxConcurrentConnections: 5000
    maxThroughput: 10000

  # Policy references
  botPolicyRef: "production-bot-policy"
  securityPolicyRef: "ddos-protection-policy"
  trafficCloneProfileRef: "traffic-mirror-profile"
  hostNameXlate: "internal-service.company.com"

  # Profile references
  analyticsProfile:
    kind: AviRef
    name: "web-analytics-profile"

  applicationProfile:
    kind: AviRef
    name: "http-application-profile"

  wafPolicy:
    kind: AviRef
    name: "owasp-waf-policy"

  icapProfile:
    kind: AviRef
    name: "antivirus-scan-profile"

  errorPageProfile:
    kind: AviRef
    name: "custom-error-pages"

  # Analytics policy
  analyticsPolicy:
    fullClientLogs:
      enabled: true
      throttle: MEDIUM
      duration: 3600
    logAllHeaders: true

  # HTTP policy
  httpPolicy:
    overwrite: true
    policySets:
      - "rate-limiting-policy"
      - "header-rewrite-policy"
      - "redirect-policy"
```

#### Minimal L7Rule Example
```yaml
apiVersion: ako.vmware.com/v1alpha2
kind: L7Rule
metadata:
  name: minimal-l7-rule
  namespace: default
spec:
  minPoolsUp: 1
  performanceLimits:
    maxConcurrentConnections: 500
```

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
| `sslSessCacheAvgSize`            | Expected number of SSL session cache entries (may be exceeded). Allowed values are 1024-16383.                                                                                         | 1024 |
| `botPolicyRef`            | Bot detection policy for the Virtual Service. It is a reference to an object of type BotDetectionPolicy. The BotDetectionPolicy reference used by VirtualService requires at least 552 MB `extra_shared_config_memory` configured in ServiceEngineGroup on Controller or else VS creation will fail.                                                                                    | Nil |
| `hostNameXlate`                         | Translate the HostName sent to the servers to this value. Translate the host name sent from servers back to the value used by the client. It is not applied on child vs                                                                                                   | Nil                                   |
| `minPoolsUp`                         | Minimum number of UP pools to mark VS up. Allowed values are 0-65535.                                                                                                   | 0                                    |
| `performanceLimits.maxConcurrentConnections`                         | The maximum number of concurrent client connections allowed to the Virtual Service. It is not applied on Child vs. Allowed values are 0-65535.                               | Nil                                    |
| `performanceLimits.maxThroughput`         | The maximum throughput per second for all clients allowed through the client side of the Virtual Service per SE. It is not applied on Child vs. Allowed values are 0-65535.                                                                                    | Nil                               |
| `securityPolicyRef`         | Security policy applied on the traffic of the Virtual Service. This policy is used to perform security actions such as Distributed Denial of Service (DDoS) attack mitigation, etc. It is a reference to an object of type SecurityPolicy and is not applied on child vs.                                                                                       |   Nil                            |
| `trafficCloneProfileRef`          | Server network or list of servers for cloning traffic. It is a reference to an object of type TrafficCloneProfile.                                                                                     | Nil |
| `analyticsProfile`                         | AnalyticsProfile allows to set the threshold for client experience depending upon application type. Contains `kind` (default: AviRef) and `name` fields.                                                                 | Nil                                   |
| `applicationProfile`                         | Application profile determines the behaviour of virtual services, based on application type. Contains `kind` (default: AviRef) and `name` fields.                                                                                                    | Nil                                   |
| `wafPolicy`                         | Defines specific set of protections for the application. Contains `kind` (default: AviRef) and `name` fields.                                                                                                   | Nil                                   |
| `icapProfile`                         | ICAP profile can be used for transporting HTTP traffic to 3rd party services for processes such as content sanitization and antivirus scanning. Contains `kind` (default: AviRef) and `name` fields.                                                                                                   | Nil                                   |
| `errorPageProfile`                         | ErrorPage profile is used to send custom error page to the client on specific error condition. Contains `kind` (default: AviRef) and `name` fields.                                                                                                   | Nil                                   |
| `analyticsPolicy.fullClientLogs.enabled`                         | Enable full client logs.                                                                                                   | False                                   |
| `analyticsPolicy.fullClientLogs.throttle`                         | Throttle level for full client logs. Allowed values: LOW, MEDIUM, HIGH, DISABLED.                                                                                                   | HIGH                                   |
| `analyticsPolicy.fullClientLogs.duration`                         | Duration for full client logs.                                                                                                   | 0                                   |
| `analyticsPolicy.logAllHeaders`                         | Log all headers in analytics.                                                                                                   | False                                   |
| `httpPolicy.overwrite`                         | Overwrite existing HTTP policies.                                                                                                   | Nil                                   |
| `httpPolicy.policySets`                         | Array of policy set names to apply.                                                                                                   | Nil                                   |

#### Profile Reference Types

All profile references (`analyticsProfile`, `applicationProfile`, `wafPolicy`, `icapProfile`, `errorPageProfile`) follow the same structure:

```yaml
profileName:
  kind: AviRef    # Default value, defines where the object is created
  name: "profile-name"    # Name of the profile object in AVI Controller
```

**Profile Types:**

- **analyticsProfile**: Controls client experience analytics and monitoring thresholds
- **applicationProfile**: Determines virtual service behavior based on application type (HTTP, HTTPS, etc.)
- **wafPolicy**: Web Application Firewall policies for application protection
- **icapProfile**: ICAP (Internet Content Adaptation Protocol) for content filtering and antivirus scanning
- **errorPageProfile**: Custom error pages for specific error conditions

#### Analytics Policy Configuration

The `analyticsPolicy` section allows fine-grained control over logging and analytics:

```yaml
analyticsPolicy:
  fullClientLogs:
    enabled: true          # Enable/disable full client logging
    throttle: MEDIUM       # Throttle level: LOW, MEDIUM, HIGH, DISABLED
    duration: 3600         # Duration in seconds for logging
  logAllHeaders: true      # Log all HTTP headers
```

#### HTTP Policy Configuration

The `httpPolicy` section controls HTTP policy application:

```yaml
httpPolicy:
  overwrite: true          # Overwrite existing HTTP policies
  policySets:              # Array of policy set names
    - "rate-limiting-policy"
    - "header-rewrite-policy"
    - "redirect-policy"
```

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

#### Attaching L7Rule to HTTPRoute

An L7Rule can be specified in `ExtensionRef` filter of HTTPRoute rule to customise Child VS specific properties as follows.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: hostname1-route-01
spec:
  parentRefs:
  - name: gw-api-01
  hostnames:
  - "hostname1.avi.internal"
  rules:
  - matches:
    - headers:
      -  type: Exact
         name: Host
         value: hostname1.avi.internal:8080
    backendRefs:
    - name: avisvc1
      port: 8080
    filters:
    - type: RequestHeaderModifier
      requestHeaderModifier:
        add:
        -  name: Host
           value: "hostname1.avi.internal"
    - type: ExtensionRef
      extensionRef:
        group: ako.vmware.com
        kind:  L7Rule
        name:  l7rule-ns
```

**Note**
- L7Rule should be created in the same namespace as that of HTTPRoute.

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

#### Validation and Constraints

**Numeric Constraints:**
- `sslSessCacheAvgSize`: Must be between 1024 and 16383
- `minPoolsUp`: Must be between 0 and 65535
- `performanceLimits.maxConcurrentConnections`: Must be between 0 and 65535
- `performanceLimits.maxThroughput`: Must be between 0 and 65535

**Enum Values:**
- `analyticsPolicy.fullClientLogs.throttle`: Must be one of `LOW`, `MEDIUM`, `HIGH`, `DISABLED`
- `analyticsProfile.kind`: Must be `AviRef` (default)
- `applicationProfile.kind`: Must be `AviRef` (default)
- `wafPolicy.kind`: Must be `AviRef` (default)
- `icapProfile.kind`: Must be `AviRef` (default)
- `errorPageProfile.kind`: Must be `AviRef` (default)

**Required Fields:**
- All profile references require both `kind` and `name` fields
- `kind` field only supports `AviRef` value

#### Conditions and Caveats

##### L7Rule deletion

If an L7Rule is deleted, the corresponding fields in L7 VSes in the AVI controller will be configured with the default values.

##### HostRule deletion

If an HostRule referencing an L7Rule is deleted , the corresponding fields in L7 VSes in the AVI controller will be configured with the default values.

##### L7Rule admission

An L7Rule CRD is only admitted if all the objects referenced in it, exist in the AVI Controller. If after admission the object references are
deleted out-of-band, then AKO does not re-validate the associated HostRule CRD objects. The user needs to manually edit or delete the object
for new changes to take effect.

##### L7Rule fields ignored in GatewayAPI context

As L7Rule is applied on HTTPRoute (child virtual service in AviController), following fields are ignored by AKO Gateway API container
- HostNameXlate
- performanceLimits
- SecurityPolicyRef

#### Field Priority: HostRule vs L7Rule

When both HostRule and L7Rule are configured for the same virtual service, some fields may be defined in both resources. In such cases, **HostRule takes precedence** over L7Rule.

**How it works:**
- If a field is specified in both HostRule and L7Rule, the HostRule value will be used
- If a field is only specified in L7Rule (and not in HostRule), the L7Rule value will be used
- This ensures that HostRule configurations can override L7Rule settings when needed

**Fields where HostRule has priority over L7Rule:**
- `analyticsPolicy` - Analytics and logging configuration
- `analyticsProfile` - Analytics profile reference
- `errorPageProfile` - Custom error page configuration
- `applicationProfile` - Application behavior profile
- `icapProfile` - ICAP content filtering profile
- `wafPolicy` - Web Application Firewall policy
- `httpPolicy` - HTTP policy sets

**Example scenario:**
```yaml
# HostRule configuration
apiVersion: ako.vmware.com/v1beta1
kind: HostRule
metadata:
  name: hostrule-1
spec:
  virtualhost:
    fqdn: example.com
    analyticsProfile:
      name: "hostrule-analytics"  # This will be used
    l7Rule: my-l7-rule

---
# L7Rule configuration
apiVersion: ako.vmware.com/v1alpha2
kind: L7Rule
metadata:
  name: L7Rule-1
spec:
  analyticsProfile:
    name: "l7rule-analytics"  # This will be ignored
```

In this example, the `analyticsProfile` from HostRule will be applied, and the one from L7Rule will be ignored.
