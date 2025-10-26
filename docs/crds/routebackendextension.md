### RouteBackendExtension

RouteBackendExtension CRD is used to configure backend-specific properties for routes in Gateway API implementations. This CRD allows users to define load balancing algorithms, session persistence, health monitoring, and TLS/SSL settings for backend servers.

**NOTE**: RouteBackendExtension CRD is specifically designed for use with Gateway API and is not supported with traditional Ingress resources.

A sample RouteBackendExtension CRD looks like this:

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: RouteBackendExtension
metadata:
  name: my-route-backend-extension
  namespace: default
spec:
  # Load balancing configuration
  lbAlgorithm: LB_ALGORITHM_CONSISTENT_HASH
  lbAlgorithmHash: LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER
  lbAlgorithmConsistentHashHdr: "X-Forwarded-For"
  
  # Session persistence configuration
  persistenceProfile: System-Persistence-Client-IP
  
  # Health monitoring configuration
  healthMonitor:
  - kind: "AVIREF"
    name: "my-health-monitor"
  
  # Backend TLS/SSL configuration
  backendTLS:
    pkiProfile:
      kind: "CRD"
      name: "my-pki-profile"
    hostCheckEnabled: true
    domainName:
    - "backend.example.com"
    - "api.example.com"
```

#### Spec Fields

The RouteBackendExtension CRD supports the following configuration options:

##### Field Reference Summary

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `lbAlgorithm` | string | No | `LB_ALGORITHM_LEAST_CONNECTIONS` | Load balancing algorithm for traffic distribution |
| `lbAlgorithmHash` | string | Conditional* | - | Hash criteria for consistent hashing (required when using consistent hash) |
| `lbAlgorithmConsistentHashHdr` | string | Conditional** | - | HTTP header name for custom header hashing |
| `persistenceProfile` | string | No | - | Session persistence mechanism |
| `healthMonitor` | array | No | - | Health monitors for backend server health checks |
| `backendTLS` | object | No | - | TLS/SSL configuration for backend communication |
| `backendTLS.pkiProfile` | object | No | - | PKI profile reference for TLS certificates |
| `backendTLS.hostCheckEnabled` | boolean | No | `false` | Enable hostname verification during TLS handshake |
| `backendTLS.domainName` | array | Conditional*** | - | Domain names for server certificate verification |

\* Required when `lbAlgorithm` is `LB_ALGORITHM_CONSISTENT_HASH`
\** Required when `lbAlgorithmHash` is `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER`
\*** Can only be set when `hostCheckEnabled` is `true`

##### Load Balancing Configuration

**lbAlgorithm** (optional)
- **Type**: string
- **Default**: `LB_ALGORITHM_LEAST_CONNECTIONS`
- **Description**: Defines the load balancing algorithm used to distribute traffic across backend servers. The algorithm determines how incoming requests are routed to available backend servers.
- **Supported Values**:
  - `LB_ALGORITHM_LEAST_CONNECTIONS`: Routes requests to the server with the fewest active connections
  - `LB_ALGORITHM_ROUND_ROBIN`: Distributes requests sequentially across all available servers
  - `LB_ALGORITHM_FASTEST_RESPONSE`: Routes requests to the server with the fastest response time
  - `LB_ALGORITHM_CONSISTENT_HASH`: Uses consistent hashing to ensure the same client is always routed to the same server (requires `lbAlgorithmHash`)
  - `LB_ALGORITHM_LEAST_LOAD`: Routes requests to the server with the lowest current load
  - `LB_ALGORITHM_FEWEST_SERVERS`: Routes requests to minimize the number of servers used
  - `LB_ALGORITHM_RANDOM`: Randomly distributes requests across available servers
  - `LB_ALGORITHM_FEWEST_TASKS`: Routes requests to the server with the fewest pending tasks
  - `LB_ALGORITHM_NEAREST_SERVER`: Routes requests to the geographically nearest server
  - `LB_ALGORITHM_CORE_AFFINITY`: Routes requests based on CPU core affinity
  - `LB_ALGORITHM_TOPOLOGY`: Routes requests based on network topology considerations

**lbAlgorithmHash** (conditional)
- **Type**: string
- **Description**: Specifies the criteria used as a key for consistent hashing when `lbAlgorithm` is set to `LB_ALGORITHM_CONSISTENT_HASH`. This field is required when using consistent hash load balancing and determines what client information is used to generate the hash.
- **Required When**: `lbAlgorithm` is `LB_ALGORITHM_CONSISTENT_HASH`
- **Supported Values**:
  - `LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS`: Uses the client's source IP address for hashing
  - `LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT`: Uses both source IP address and port for hashing
  - `LB_ALGORITHM_CONSISTENT_HASH_URI`: Uses the request URI for hashing
  - `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER`: Uses a custom HTTP header for hashing (requires `lbAlgorithmConsistentHashHdr`)
  - `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING`: Uses a custom string for hashing
  - `LB_ALGORITHM_CONSISTENT_HASH_CALLID`: Uses the SIP Call-ID header for hashing (SIP applications)

**lbAlgorithmConsistentHashHdr** (conditional)
- **Type**: string
- **Description**: Specifies the name of the HTTP header to use as the hash key when using custom header-based consistent hashing. This field is required when `lbAlgorithmHash` is set to `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER`.
- **Required When**: `lbAlgorithmHash` is `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER`
- **Example Values**: `"X-User-ID"`, `"X-Forwarded-For"`, `"Authorization"`

##### Session Persistence Configuration

**persistenceProfile** (optional)
- **Type**: string
- **Description**: Defines the session persistence mechanism to ensure client requests are consistently routed to the same backend server. This will apply the corresponding System-Default Persistence Profile present in Avi Load Balancer to the pool. Check for more details [here](https://techdocs.broadcom.com/us/en/vmware-security-load-balancing/avi-load-balancer/avi-load-balancer/31-1/vmware-avi-load-balancer-configuration-guide/load-balancing-overview/persistence.html).
- **Supported Values**:
  - `System-Persistence-Client-IP`: IP-based session persistence
  - `System-Persistence-Http-Cookie`: HTTP cookie-based persistence
  - `System-Persistence-TLS`: TLS session ID-based persistence
  - `System-Persistence-App-Cookie`: Application cookie-based persistence

##### Health Monitoring Configuration

**healthMonitor** (optional)
- **Type**: array of BackendHealthMonitor objects
- **Description**: Defines health monitors for backend server health checks. Multiple health monitors can be specified, and a backend server is marked UP only when all health monitors return successful responses. Health monitors provide automated health checking to ensure traffic is only sent to healthy backend servers.

Each BackendHealthMonitor object has the following fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `kind` | string | Yes | Type of HealthMonitor object. Must be `"AVIREF"` (reference to health monitor on Avi Controller). This indicates that the health monitor is a reference to an existing health monitor configured in the Avi Load Balancer Controller. |
| `name` | string | Yes | Name of the HealthMonitor object. The health monitor must exist in the Avi Controller in the same tenant as the RouteBackendExtension namespace. This should match the exact name of the health monitor as configured in Avi. |

**Example:**
```yaml
healthMonitor:
  - kind: "AVIREF"
    name: "http-health-check"
  - kind: "AVIREF"
    name: "tcp-health-check"
```

**Important Notes:**
- Only `AVIREF` kind is supported for health monitors in RouteBackendExtension
- Health monitors must be pre-created in the Avi Controller
- Multiple health monitors can be specified; servers must pass all checks to be marked UP
- Health monitors referenced must exist in the same tenant as the RouteBackendExtension namespace

##### Backend TLS/SSL Configuration

The BackendTLS section contains a set of properties that are crucial for securing communication between gateways and backend services, ensuring trust validation in modern cloud-native architectures. The sample BackendTLS configuration is given below:

```yaml
# BackendTLS defines the TLS/SSL configuration for secure communication with backend servers
backendTLS:
  # PKI Profile for certificate validation
  pkiProfile:
    kind: "CRD"  # Must be "CRD" for PKIProfile CRD references
    name: "backend-ca-profile"
  
  # Enable hostname verification during TLS handshake
  hostCheckEnabled: true
  
  # Domain names for backend certificate validation
  # Note: domainName can only be configured when hostCheckEnabled is set to true
  domainName:
    - "backend.example.com"
    - "api.example.com"
```

**Field Descriptions:**

- **pkiProfile**: References a PKIProfile CRD object that contains trusted CA certificates for validating backend server certificates. The `kind` must be set to "CRD" and `name` should reference an existing PKIProfile in the same namespace. For more information on this field, refer to the [PKIProfile documentation](pkiprofile.md).

- **hostCheckEnabled**: Boolean flag that enables hostname verification during the TLS handshake. When set to `true`, the system validates that the backend server's certificate matches the expected hostname.

- **domainName**: List of domain names used for certificate subject validation when `hostCheckEnabled` is `true`. The backend server's certificate must match one of the specified domain names for the connection to be considered valid.

**Default SSL Profile**

When BackendTLS features are configured in RouteBackendExtension, AKO automatically attaches the **System-Standard** SSL Profile to the backend pool configuration. This profile enables SSL/TLS re-encryption for traffic to backend servers, provides modern and secure cipher suites for backend communication, supports TLS 1.2 and TLS 1.3 for secure connections, and works in conjunction with PKIProfile for certificate verification. The System-Standard SSLProfile includes the following security features:

| Feature | Configuration | Description |
|---------|---------------|-------------|
| **SSL Versions** | TLS 1.2, TLS 1.3 | Modern TLS versions for secure communication |
| **Cipher Suites** | ECDHE-RSA-AES256-GCM-SHA384, ECDHE-RSA-AES128-GCM-SHA256, etc. | Strong encryption algorithms |
| **Perfect Forward Secrecy** | Enabled | Ensures session keys are not compromised if private key is compromised |
| **SSL Session Reuse** | Enabled | Improves performance by reusing SSL sessions |
| **SSL Session Timeout** | 86400 seconds (24 hours) | Balances security and performance |

For detailed information about SSL/TLS Profile configuration, cipher suites, and SSL versions, refer to the [VMware Avi Load Balancer SSL/TLS Profile documentation](https://techdocs.broadcom.com/us/en/vmware-security-load-balancing/avi-load-balancer/avi-load-balancer/30-2/security/ssl-tls-profile.html).


**Example Pool Configuration:**
When BackendTLS is configured, the resulting Avi pool will have:

```json
{
  "pool": {
    "name": "cluster-namespace-service-port",
    "ssl_profile_ref": "/api/sslprofile?name=system-standard",
    "pki_profile_ref": "/api/pkiprofile?name=cluster-namespace-pkiprofile-name",
    "ssl_key_and_certificate_ref": null,
    "host_check_enabled": true,
    "server_name": "",
    "sni_enabled": false,
    "rewrite_host_header_to_server_name": false,
    "domain_name": ["backend.example.com", "api.example.com"]
  }
}
```

**Note**: SSLProfile is automatically added by default and is not configurable.

#### Usage Examples

##### Basic Load Balancing Configuration

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: RouteBackendExtension
metadata:
  name: basic-lb-config
  namespace: production
spec:
  lbAlgorithm: LB_ALGORITHM_ROUND_ROBIN
```

##### Consistent Hash with Custom Header

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: RouteBackendExtension
metadata:
  name: consistent-hash-config
  namespace: production
spec:
  lbAlgorithm: LB_ALGORITHM_CONSISTENT_HASH
  lbAlgorithmHash: LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER
  lbAlgorithmConsistentHashHdr: "X-User-ID"
  persistenceProfile: System-Persistence-Client-IP
```

##### Health Monitoring Configuration

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: RouteBackendExtension
metadata:
  name: health-monitor-config
  namespace: production
spec:
  lbAlgorithm: LB_ALGORITHM_LEAST_CONNECTIONS
  healthMonitor:
  - kind: "AVIREF"
    name: "http-health-check"
  - kind: "AVIREF"
    name: "tcp-health-check"
```

##### Backend TLS Configuration

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: RouteBackendExtension
metadata:
  name: backend-tls-config
  namespace: production
spec:
  lbAlgorithm: LB_ALGORITHM_LEAST_CONNECTIONS
  backendTLS:
    pkiProfile:
      kind: "CRD"
      name: "backend-pki-profile"
    hostCheckEnabled: true
    domainName:
    - "api.backend.com"
    - "service.backend.com"
```

##### Complete Configuration Example

```yaml
apiVersion: ako.vmware.com/v1alpha1
kind: RouteBackendExtension
metadata:
  name: complete-backend-config
  namespace: production
spec:
  # Load balancing with consistent hashing
  lbAlgorithm: LB_ALGORITHM_CONSISTENT_HASH
  lbAlgorithmHash: LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS
  
  # Session persistence
  persistenceProfile: System-Persistence-Client-IP
  
  # Health monitoring
  healthMonitor:
  - kind: "AVIREF"
    name: "production-health-monitor"
  
  # Secure backend communication
  backendTLS:
    pkiProfile:
      kind: "CRD"
      name: "production-pki-profile"
    hostCheckEnabled: true
    domainName:
    - "backend.production.com"
```

#### Validation Rules and Constraints

The RouteBackendExtension CRD enforces the following validation rules to ensure proper configuration:

1. **Consistent Hash Algorithm Validation**:
   - When `lbAlgorithm` is set to `LB_ALGORITHM_CONSISTENT_HASH`, the `lbAlgorithmHash` field must be specified
   - Conversely, `lbAlgorithmHash` can only be set when `lbAlgorithm` is `LB_ALGORITHM_CONSISTENT_HASH`
   - **Error Message**: "lbAlgorithmHash must be set if and only if lbAlgorithm is LB_ALGORITHM_CONSISTENT_HASH"

2. **Custom Header Hash Validation**:
   - When `lbAlgorithmHash` is set to `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER`, the `lbAlgorithmConsistentHashHdr` field must be provided
   - The `lbAlgorithmConsistentHashHdr` field can only be set when using custom header hashing
   - **Error Message**: "lbAlgorithmConsistentHashHdr must be set if and only if lbAlgorithmHash is LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER"

3. **Backend TLS Domain Name Validation**:
   - The `domainName` field in `backendTLS` can only be configured when `hostCheckEnabled` is set to `true`
   - This ensures domain name verification is only attempted when hostname checking is enabled
   - **Error Message**: "domainName can only be configured when hostCheckEnabled is set to true"

4. **Namespace Scope Requirements**:
   - All referenced objects (PKI profiles, health monitors) must be in the same namespace as the RouteBackendExtension object
   - This ensures proper RBAC and resource isolation

5. **Health Monitor Constraints**:
   - Health monitor `kind` must be `"AVIREF"` (only Avi Controller references are supported)
   - Both `kind` and `name` fields are required for each health monitor entry

6. **PKI Profile Constraints**:
   - PKI profile `kind` must be `"CRD"` (only Custom Resource Definition references are supported)
   - Both `kind` and `name` fields are required when specifying a PKI profile

#### Gateway API Integration

RouteBackendExtension is designed to work with Gateway API resources. To use this CRD with Gateway API:

1. **HTTPRoute Integration**: Reference the RouteBackendExtension in HTTPRoute resources using `ExtensionRef` filters in the `backendRefs` section.

2. **Namespace Alignment**: Ensure the RouteBackendExtension is created in the same namespace as the HTTPRoute that references it.

3. **Backend Service Mapping**: The configuration applies to the specific backend services that reference the RouteBackendExtension through filters.

##### Example: HTTPRoute with RouteBackendExtension

Here's an example showing how to use RouteBackendExtension named complete-backend-config with HTTPRoute:

**HTTPRoute with RouteBackendExtension reference**

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: api-route
  namespace: production
spec:
  parentRefs:
  - name: production-gateway
    namespace: gateway-system
  hostnames:
  - "api.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: "/v1/users"
    backendRefs:
    - name: user-service
      port: 8080
      weight: 100
      filters:
      - type: ExtensionRef
        extensionRef:
          group: ako.vmware.com
          kind: RouteBackendExtension
          name: complete-backend-config
  - matches:
    - path:
        type: PathPrefix
        value: "/v1/orders"
    backendRefs:
    - name: order-service
      port: 8080
      weight: 100
      filters:
      - type: ExtensionRef
        extensionRef:
          group: ako.vmware.com
          kind: RouteBackendExtension
          name: complete-backend-config
```



#### Status Fields

The RouteBackendExtension CRD provides status information to track the state of the resource:

**status.controller** (read-only)
- **Type**: string
- **Description**: Indicates which controller is managing the resource. This field is automatically populated by the AKO CRD operator.
- **Typical Value**: `"ako-crd-operator"`

**status.status** (read-only)
- **Type**: string
- **Description**: Current status of the resource processing. Indicates whether the RouteBackendExtension has been successfully processed and applied.
- **Possible Values**:
  - `"Accepted"`: The configuration has been successfully validated and applied
  - `"Rejected"`: The configuration contains errors and could not be applied

**status.error** (read-only)
- **Type**: string
- **Description**: Error message providing details about why the CRD object was transitioned to the "Rejected" state. This field is only populated when there are validation or processing errors.
- **Usage**: Check this field when `status.status` is "Rejected" to understand what needs to be corrected

#### kubectl Output

When listing RouteBackendExtension resources using `kubectl get routebackendextensions` or `kubectl get rbe` (short name), the following columns are displayed:

- **NAME**: The name of the RouteBackendExtension resource
- **STATUS**: Current processing status (Accepted/Rejected)
- **AGE**: Time since the resource was created

Example output:
```
$ kubectl get rbe
NAME                     STATUS     AGE
my-route-backend-ext     Accepted   5m30s
failed-config           Rejected   2m15s
```

#### Troubleshooting

##### Common Issues

**1. PKIProfile Not Found**
*Symptom*: RouteBackendExtension shows error about missing PKIProfile

*Solution*: 
- Verify PKIProfile exists in the same namespace
- Check PKIProfile name spelling in RouteBackendExtension
- Ensure PKIProfile has valid CA certificates

```bash
kubectl get pkiprofiles -n <namespace>
kubectl describe pkiprofile <name> -n <namespace>
```

**2. Domain Name Validation Error**
*Symptom*: Validation error about domainName configuration

*Solution*: 
- Ensure `hostCheckEnabled` is set to `true` when using `domainName`
- Verify domain names match your backend service certificates

**3. AKO CRD Operator Not Running**
*Symptom*: CRDs not being processed

*Solution*:
- Check AKO CRD Operator pod status
- Review ako-crd-operator logs for errors
- Verify RBAC permissions

```bash
kubectl get pods -n avi-system
kubectl logs -f deployment/ako-crd-operator -n avi-system
```

**4. TLS Handshake Failures**
*Symptom*: Backend connection failures with TLS errors

*Solution*:
- Verify CA certificates in PKIProfile match backend certificates
- Check domain names in RouteBackendExtension match backend certificate SANs
- Ensure backend services are correctly configured for TLS

##### Debug Commands

```bash
# Check CRD status
kubectl get pkiprofiles,routebackendextensions -A

# Describe resources for detailed information
kubectl describe pkiprofile <name> -n <namespace>
kubectl describe routebackendextension <name> -n <namespace>

# Check AKO CRD Operator logs
kubectl logs deployment/ako-crd-operator -n avi-system -f

# Verify Avi Controller configuration
# (Access Avi Controller UI to verify Pool and SSL configurations)
```

