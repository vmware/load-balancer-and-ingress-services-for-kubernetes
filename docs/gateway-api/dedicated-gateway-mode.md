# Dedicated Gateway Mode

## Overview

Dedicated Gateway Mode is a special operating mode for Kubernetes Gateway API Gateways in AKO (Avi Kubernetes Operator) when there are no hostnames present in gateway or the HTTPRoute.

## Architecture

**Dedicated Gateway Mode:**
```
Gateway (Single Dedicated VS)
  └── HTTPRoute → HTTP Policy Set (attached directly to VS)
       ├── HTTP Request Rules (all route rules combined)
       ├── HTTP Response Rules
       ├── Pool Groups
       └── Pools
```

### Key Architectural Differences

1. **HTTPRoute Processing:** 
   - **Dedicated Mode**: All HTTPRoute rules are consolidated into a single HTTP Policy Set attached to the dedicated VS

2. **Naming Convention:** 
   - **Dedicated Mode**: Single VS: `ako-gw-<cluster>--<ns>-<gateway>-L7-dedicated-EVH`

## Configuration

### Enabling Dedicated Mode

Dedicated mode is enabled by adding an annotation to the Gateway resource:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: dedicated-gateway
  namespace: default
  annotations:
    ako.vmware.com/dedicated-gateway-mode: "true"
spec:
  gatewayClassName: avi-lb
  listeners:
  - name: http-listener
    protocol: HTTP
    port: 8080
    # Note: hostname must NOT be specified in dedicated mode
  - name: https-listener
    protocol: HTTPS
    port: 8443
    tls:
      certificateRefs:
      - kind: Secret
        name: my-tls-cert
```

## Limitations

### Gateway Limitations

1. **No Hostname in Listeners**
   - Listener-level hostname specification is **NOT supported**
   - Validation Error: `"Hostname is not supported in dedicated mode"`
   - **Reason**: Dedicated mode relies on HTTPRoute path-based routing only
   

2. **No Routes from All Namespaces**
   - `allowedRoutes.namespaces.from: All` is **NOT supported**
   - Validation Error: `"Routes from all namespaces is not supported in dedicated mode"`
   - **Reason**: Simplifies namespace isolation and security model   

### HTTPRoute Limitations

1. **Single Parent Reference Only**
   - Only **ONE** parent Gateway can be referenced per HTTPRoute
   - Validation Error: `"Dedicated Gateway Mode is enabled. Only one parent reference is allowed in HTTPRoute"`
   - **Reason**: One-to-one mapping between Gateway and HTTPRoute simplifies configuration
   
2. **No Hostnames in HTTPRoute**
   - HTTPRoute-level `hostnames` field is **NOT supported**
   - Validation Error: `"Dedicated Gateway Mode is enabled. Hostnames are not allowed in HTTPRoute"`
   - **Reason**: All routing decisions are based on path matching only

3. **One HTTPRoute Per Gateway**
   - Only **ONE** HTTPRoute can be attached to a dedicated Gateway at a time
   - Validation Error: `"Dedicated Gateway Mode is enabled. Only one route is allowed per listener in Gateway"`
   - **Reason**: Simplifies management and avoids routing conflicts
   - **Note**: The single HTTPRoute can contain multiple rules with different paths

   ```yaml
   apiVersion: gateway.networking.k8s.io/v1
   kind: HTTPRoute
   metadata:
     name: my-single-route
   spec:
     parentRefs:
     - name: dedicated-gateway
     rules:
     - matches:
       - path:
           type: PathPrefix
           value: /api
       backendRefs:
       - name: api-service
         port: 8080
     - matches:
       - path:
           type: PathPrefix
           value: /web
       backendRefs:
       - name: web-service
         port: 80
   ```

4. **Parent and Route Must Be in Same Namespace**
   - Cross-namespace parent references are **NOT supported**
   - Validation Error: `"Dedicated Gateway Mode is enabled. Parent Reference X is not in the same namespace as HTTPRoute Y"`
   - **Reason**: Simplifies RBAC and namespace isolation

## Supported HTTPRoute Features

Despite the limitations, dedicated mode supports all standard HTTPRoute routing and filtering capabilities:

### Path Matching
- **Exact** path matching
- **PathPrefix** matching
- **RegularExpression** matching

```yaml
rules:
- matches:
  - path:
      type: Exact
      value: /api/v1/users
  backendRefs:
  - name: users-service
    port: 8080
- matches:
  - path:
      type: PathPrefix
      value: /api
  backendRefs:
  - name: api-service
    port: 8080
- matches:
  - path:
      type: RegularExpression
      value: /api/v[0-9]+/.*
  backendRefs:
  - name: versioned-api
    port: 8080
```

### Header Matching
- **Exact** header matching
- **RegularExpression** header matching

```yaml
rules:
- matches:
  - headers:
    - type: Exact
      name: x-api-version
      value: v1
    - type: Exact
      name: authorization
      value: Bearer
    path:
      type: PathPrefix
      value: /api
```

### Request Filters
- **RequestHeaderModifier** - Add, set, remove headers
- **RequestRedirect** - HTTP redirects
- **URLRewrite** - Path and hostname rewriting

```yaml
rules:
- matches:
  - path:
      type: PathPrefix
      value: /api
  filters:
  - type: RequestHeaderModifier
    requestHeaderModifier:
      add:
      - name: X-Custom-Header
        value: custom-value
      set:
      - name: X-Environment
        value: production
      remove:
      - X-Debug
  - type: URLRewrite
    urlRewrite:
      path:
        type: ReplaceFullPath
        replaceFullPath: /v2/api
      hostname: api.example.com
  - type: RequestRedirect
    requestRedirect:
      hostname: new.example.com
      statusCode: 301
```

### Response Filters
- **ResponseHeaderModifier** - Add, set, remove response headers

```yaml
rules:
- matches:
  - path:
      type: PathPrefix
      value: /api
  filters:
  - type: ResponseHeaderModifier
    responseHeaderModifier:
      add:
      - name: X-Cache-Status
        value: HIT
      set:
      - name: X-Content-Type-Options
        value: nosniff
      remove:
      - Server
```

### Backend References
- Multiple backends with weight distribution
- Cross-namespace backend services (with proper RBAC)
- Service kind backends

```yaml
rules:
- matches:
  - path:
      type: PathPrefix
      value: /api
  backendRefs:
  - name: api-v1
    port: 8080
    weight: 70
  - name: api-v2
    port: 8080
    weight: 30
  - name: api-fallback
    namespace: fallback-ns  # Cross-namespace OK for backends
    port: 8080
    weight: 10
```

### Extension References

Dedicated mode supports all standard Gateway API extension references including:
- Custom CRD-based filters and backend references
- Health monitors via extensionRef
- Application profiles and other AKO-specific extensions

For detailed information on using extensionRef with HTTPRoute filters and backend references, see the [Gateway API v1 Extension Reference documentation](./gateway-api-v1.md).

## HTTP Policy Set Structure in Dedicated Mode

In dedicated mode, all HTTPRoute rules are consolidated into a single HTTP Policy Set:

### Rule Ordering and Priority

Rules are automatically sorted by AKO using the following priority:

1. **Path Match Type Priority** (highest to lowest):
   - Exact match
   - PathPrefix match
   - RegularExpression match

2. **Path Length** (for same match type):
   - Longer paths have higher priority
   - Example: `/api/v1/users` before `/api/v1` before `/api`

3. **Rule Order** (for same path):
   - Maintains the order defined in HTTPRoute

4. **Default Backend Rule** (lowest priority):
   - Automatically added if no root path (`/`) is present
   - Returns 404 for unmatched requests

```yaml
# HTTPRoute with multiple rules
rules:
- matches:  # Priority 1: Exact match
  - path:
      type: Exact
      value: /api/v1/users
- matches:  # Priority 2: Longer prefix
  - path:
      type: PathPrefix
      value: /api/v1
- matches:  # Priority 3: Shorter prefix
  - path:
      type: PathPrefix
      value: /api
- matches:  # Priority 4: Regex (lower than prefix)
  - path:
      type: RegularExpression
      value: /api/v[0-9]+/.*
# Implicit Default Rule (Priority 5): Catch-all 404
```

### Policy Set Naming

HTTP Policy Set name format:
```
<gateway-namespace>-<gateway-name>-<route-namespace>-<route-name>-httppolicyset
```

Example: `default-my-gateway-default-my-route-httppolicyset`



## Complete Example

```yaml
# Dedicated Gateway
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: production-gateway
  namespace: production
  annotations:
    ako.vmware.com/dedicated-gateway-mode: "true"
spec:
  gatewayClassName: avi-lb
  addresses:
  - type: IPAddress
    value: 10.10.10.100
  listeners:
  - name: http
    protocol: HTTP
    port: 80
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      certificateRefs:
      - kind: Secret
        name: api-tls-cert

---
# HTTPRoute with Multiple Rules
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: api-routes
  namespace: production
spec:
  parentRefs:
  - name: production-gateway
  rules:
  # Admin endpoint with header auth
  - matches:
    - path:
        type: Exact
        value: /admin
      headers:
      - type: Exact
        name: x-admin-token
        value: secret
    filters:
    - type: RequestHeaderModifier
      requestHeaderModifier:
        add:
        - name: X-Admin-Access
          value: "true"
    backendRefs:
    - name: admin-service
      port: 8080
 

## Troubleshooting

### Gateway Not Accepting Routes

**Symptom:** HTTPRoute shows `Accepted: False`

**Checks:**
1. Verify dedicated mode annotation on Gateway
2. Ensure HTTPRoute has no hostnames
3. Confirm only one parent reference
4. Check Gateway and HTTPRoute are in same namespace
5. Verify no other HTTPRoute is already attached

### Multiple HTTPRoutes Rejected

**Symptom:** Second HTTPRoute shows status `"Only one route is allowed per listener"`

**Solution:**
- Combine all routing rules into a single HTTPRoute
- Use multiple rules within one HTTPRoute instead of multiple HTTPRoutes

### Listener Not Accepted

**Symptom:** Gateway listener shows `Accepted: False`

**Common Causes:**
1. Hostname specified in listener (not allowed)
2. `allowedRoutes.namespaces.from: All` set (not allowed)
3. Invalid protocol (only HTTP/HTTPS supported)

### Virtual Service Not Created

**Checks:**
1. Verify GatewayClass controller is `ako.vmware.com/avi-lb`
2. Check Gateway status conditions
3. Ensure all required secrets exist
4. Review AKO controller logs for errors

## Best Practices

1. **Use Descriptive Rule Names:**
   ```yaml
   rules:
   - name: admin-endpoint
     matches: [...]
   - name: api-v2
     matches: [...]
   ```

2. **Order Rules by Specificity:**
   - Place more specific paths first
   - AKO will sort them, but explicit ordering helps readability

3. **Always Include a Catch-All:**
   ```yaml
   rules:
   - matches:
     - path:
         type: PathPrefix
         value: /
     backendRefs:
     - name: default-backend
       port: 80
   ```

4. **Use Path Prefixes for Flexibility:**
   ```yaml
   # Good: Allows /api/*, /api/v1/*, etc.
   - matches:
     - path:
         type: PathPrefix
         value: /api
   ```

5. **Leverage Header Matching for API Versioning:**
   ```yaml
   - matches:
     - path:
         type: PathPrefix
         value: /api
       headers:
       - type: Exact
         name: x-api-version
         value: v2
   ```

6. **Monitor Attachment Status:**
   - Regularly check `attachedRoutes` count on Gateway listeners
   - Should always be 0 or 1 in dedicated mode

7. **Document Mode Selection:**
   - Clearly document why dedicated mode was chosen
   - Include criteria for switching to regular mode if needed

## References

- [Gateway API v1 Specification](https://gateway-api.sigs.k8s.io/)
- [AKO Gateway API Support](./gateway-api-v1.md)
- [AviInfraSetting Configuration](../crds/aviinfrasetting.md)

