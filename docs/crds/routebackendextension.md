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

##### Load Balancing Configuration

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
- **Description**: Defines health monitors for backend server health checks. Multiple health monitors can be specified, and a backend server is marked UP only when all health monitors return successful responses.

Each BackendHealthMonitor object has the following fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `kind` | string | Yes | Type of HealthMonitor object. Must be `AVIREF` (reference to health monitor on Avi Controller). CRD references are not supported. |
| `name` | string | Yes | Name of the HealthMonitor object. The health monitor must exist in the Avi Controller in the same tenant as the RouteBackendExtension namespace. |

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

The RouteBackendExtension CRD enforces the following validation rules:

1. **Consistent Hash Validation**: When `lbAlgorithm` is set to `LB_ALGORITHM_CONSISTENT_HASH`, the `lbAlgorithmHash` field must be specified.

2. **Custom Header Validation**: When `lbAlgorithmHash` is set to `LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER`, the `lbAlgorithmConsistentHashHdr` field must be provided.

3. **Domain Name Validation**: The `domainName` field in `backendTLS` can only be configured when `hostCheckEnabled` is set to `true`.

4. **Namespace Scope**: All referenced objects (PKI profiles) must be in the same namespace as the RouteBackendExtension object.

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

The RouteBackendExtension CRD provides status information:

- **controller**: Indicates which controller is managing the resource (typically "ako-crd-operator")
- **status**: Current status of the resource processing
- **error**: Error message if the resource processing failed


